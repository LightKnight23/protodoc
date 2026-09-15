# Protodoc CLI Contract

Status: DRAFT (awaiting approval by Eyvar) | Spec ID: 001-protodoc-format-core | Phase: 3 (plan) | Date: 2026-09-07

This file is the normative interface contract for the Protodoc command-line tool (TR-012). It is not
a usage guide. Every verb, every flag that changes observable behaviour, every exit code and every
stdout field named here is a conformance surface: release gate G-CLI (TR-012's own verify clause)
requires each verb to be exercised by at least one conformance case, and CP-011 requires every
normative statement here to carry a stable identifier mapped to an executable case.

TR-012's exact text: "The Protodoc command-line tool SHALL provide validate, inspect, extract, verify,
diff, merge, project, redact, publish, sign and migrate operations at the first stable release." That
is 11 verbs, in that order, and this file defines exactly those 11 and no others. DP-017's
RESCIND-AND-RESIGN operation is exposed as a flag on `migrate` (S11), not a 12th verb, so the count
stated in TR-012 stays exactly 11; this is a deliberate design choice recorded in plan.md DP-012/DP-017,
not an oversight.

## 0. Conventions that apply to every verb

**Invocation shape.** `protodoc <verb> <file> [<file2>] [flags...]`. A verb taking two input documents
(`diff`, `merge`) names both positionally before any flag; every other verb takes exactly one input
document positionally. An output-producing verb takes its destination as `--out <path>` (never a second
positional argument), so a reader of an invocation can always tell input from output without knowing the
verb.

**stdout shape.** Every verb emits exactly one JSON object to stdout per invocation, and nothing else on
stdout: no banner, no progress line, no trailing newline-delimited log. A tool MUST NOT begin writing
stdout until its own exit code is already determined internally (no streaming a partial JSON object and
then failing partway through emitting it); this is CP-006's verification-precedes-decoding principle
applied to the tool's own output discipline, not only to document decoding. Diagnostics, progress and
human-readable narration go to stderr only. `--format=text` (accepted by every verb) additionally emits
a human-readable rendering of the same information to stderr for interactive use (WCAG 2.2 AA applies to
this text where it is displayed in a GUI wrapper); the JSON on stdout is unaffected by `--format` and is
always present, because scripts and CI pipelines depend on one stable shape existing unconditionally
(Usability: errors actionable not cryptic, ISO/IEC 25010).

**Every result object carries at least:**

```json
{
  "verb": "validate",
  "status": "OK",
  "exit_code": 0,
  "findings": []
}
```

`status` is one of the 8 names in S1's table, always. `findings` is a JSON array, always present (empty
on success), of objects shaped `{"rule_id": "FR-104", "offset": 4194432, "unit_id": "9f2a...", "message": "..."}`
with `offset` and `unit_id` present only when the finding names one (per FR-102's "report the octet
offset, the unit identifier and the stable rule identifier"). A verb-specific payload is added under
verb-specific top-level keys (documented per verb below); `verb`, `status`, `exit_code` and `findings`
never change shape or meaning across verbs.

**Verification precedes decoding, at the tool-invocation level (CP-006).** Every verb that reads a
document runs validate.md's structural checks (S7 below) as an implicit precondition before doing its
own verb-specific work, regardless of which verb was actually invoked. A verb never partially completes
its own operation on a document that fails an earlier-numbered structural check; `extract`, `project`,
`diff`, `merge`, `redact`, `publish`, `sign` and `migrate` all report `INVALID` (or a more specific code
per S1) rather than a partial or best-effort result, per non-negotiable #8 (no heuristic recovery).

## 1. Exit codes (NORMATIVE)

Exactly 8 values. A single process exit code is necessarily a compression of a document's full status
into one integer; the complete, uncollapsed set of findings is always in stdout's `findings` array
regardless of which single code the process exits with (S1.1 states the precedence when more than one
category is simultaneously true).

| Code | Name | Meaning | Is this a validity verdict? |
|---|---|---|---|
| 0 | `OK` | The operation completed and, where the verb has a positive/negative distinction (`verify`), every checked item reports positively. | n/a |
| 1 | `INVALID` | Structural validation failed: truncation, unparseable structure, a digest mismatch, two disagreeing inventories, a dangling or ambiguous reference, a reference-graph cycle, a duplicate identifier, non-NFC text, a malformed coverage descriptor, a non-minimal or out-of-order encoding, or any other FR-102 through FR-110 class failure. | Yes |
| 2 | `UNSUPPORTED` | The document declares a `format-major` (container.abnf S2) higher than this tool implements. Per FR-123, the tool declines naming the required version and applies **zero** construct dispositions; this is deliberately distinct from `INVALID`, because the document may be perfectly well-formed under a later version this tool simply does not know. | No (a decline, not a verdict on content) |
| 3 | `UNVERIFIED` | The document is structurally valid, but a requested cryptographic check produced a negative result: a signature fails EdDSA-Protodoc-1 (integrity.abnf S6), a time attestation's interval is unsupported (FR-072), revocation predates signing (FR-073), or an undeclared omission is detected (FR-076). | Yes, for the verified operation specifically |
| 4 | `UNAVAILABLE` | A signature covers a state the current file can no longer reconstruct (FR-062). Textually and structurally distinct from both `OK` and `UNVERIFIED` on the spec's own words ("report that signature as covering an unavailable state rather than as failed or as valid"); collapsing this into `UNVERIFIED` would misreport a lawful history trim as a security failure. | No (neither valid nor invalid; a third state) |
| 5 | `OVER_BUDGET` | The document is otherwise structurally valid but a stated resource ceiling (CON-009, CON-010, CON-011) is exceeded for THIS reader's own budget; reported as `rule_id: "PD-BUDGET-xxx"` findings. CON-011 requires this be "distinct from every validity verdict"; per data-model.md S7's validation-order note, an over-budget document that is ALSO structurally invalid reports both findings, and the exit code follows S1.1's precedence rather than suppressing either finding from `findings`. | No, by CON-011's own text |
| 6 | `REFUSED` | The input is structurally valid, but the requested WRITE operation is declined by an explicit policy rule on an otherwise-fine input: `publish` refusing full compaction under NFR-004 while a signature is present, `migrate` halting before any output is written because phase 1 found an unrepresentable construct (FR-121), `merge` refusing a genuine identifier collision (FR-024), `sign` refusing a 65th signature (MAX_SIGNATURES). Distinct from `INVALID` because the document itself is fine; distinct from `OVER_BUDGET` because this is a policy/safety rule, not a resource ceiling. | No |
| 7 | `USAGE` | The invocation itself is malformed: a missing required flag, an unreadable path, an unrecognised verb, conflicting flags. Never a statement about any document's content. | No |

### 1.1 Precedence when more than one status is simultaneously true

Ranked highest to lowest; the exit code reports the HIGHEST-ranked status that applies, and `findings`
always carries every applicable finding regardless of rank:

`USAGE` > `INVALID` > `UNSUPPORTED` > `OVER_BUDGET` > `UNAVAILABLE` > `UNVERIFIED` > `REFUSED` > `OK`

This mirrors data-model.md S7's validation-check ordering (earlier-numbered structural checks precede
later ones) with one named exception carried over unchanged from that ordering: an `OVER_BUDGET` finding
is reported ALONGSIDE any validity verdict from an earlier check, per CON-011, rather than only ever
being reported instead of one; ranking `OVER_BUDGET` above `UNAVAILABLE`/`UNVERIFIED`/`REFUSED` here (but
below `INVALID`) reflects that a budget refusal is itself decidable before any cryptographic or
policy-level operation is attempted, not that it silently replaces an `INVALID` verdict that is also true.

### 1.2 Departure from POSIX convention, stated explicitly

`diff` does NOT use exit code 1 to mean "differences were found," unlike POSIX `diff`. Whether two
documents differ is reported in `diff`'s own stdout payload (`"identical": false`), never via the exit
code, so that exit code 1 means the SAME thing (`INVALID`) across all 11 verbs with no per-verb special
case. This uniformity is deliberate: a caller scripting against several Protodoc verbs needs one exit-
code table, not one exception for the one verb that happens to share a name with a POSIX utility.

## 2. `validate <file>`

Runs the full structural validation pipeline (data-model.md S7 steps 1-9 and 11-13; step 10, signature
verification, is `verify`'s job and does not run here) with zero cryptographic checks and zero content
extraction. Satisfies FR-102, FR-103, FR-104, FR-105, FR-106, FR-107, FR-108, FR-109, FR-110, FR-117,
CON-003, CON-009, CON-010, CON-011, and the non-cryptographic portion of TR-006/007/008.

**Flags:** none beyond S0's common set.

**Exit codes used:** 0, 1, 2, 5, 7. (Never 3, 4 or 6: this verb performs no cryptographic check and no
write, so `UNVERIFIED`, `UNAVAILABLE` and `REFUSED` cannot arise from it.)

**stdout payload (`"checks"` key):**

```json
{
  "verb": "validate", "status": "OK", "exit_code": 0, "findings": [],
  "checks": {
    "header_and_capability": "OK",
    "commit_ring_winner": "OK",
    "truncation": "OK",
    "bounded_prefix_structure": "OK",
    "storage_integrity_tree": "OK",
    "segment_type_and_coverage_wellformed": "OK",
    "cycle_detection": "OK",
    "structural_ceilings": "OK",
    "nfc_and_identity": "OK",
    "registry_excerpt_completeness": "OK",
    "extension_envelope_reachability": "OK"
  },
  "ceilings": [{"name": "MAX_SEGMENTS", "limit": 16384, "observed": 41, "status": "OK"}]
}
```

Each `checks` entry corresponds 1:1 to a numbered step in data-model.md S7 and reports `"OK"` or the
`rule_id` of the first failure at that step. Checks after the first failing one still run and still
report here, because `validate`'s job is diagnostic completeness, not a security decision; this is
unlike the Ed25519 verification procedure itself (integrity.abnf S6), whose 7 steps short-circuit at the
first failing one specifically so no later step's cost or behaviour is ever observable on a rejected
input, a property `verify` (S5 below) inherits when it runs that procedure.

## 3. `inspect <file>`

Reads ONLY the leading 1,048,576 octets (the fixed prefix, container.abnf S1) and reports what TR-006,
TR-007 and TR-008 require be determinable from it with zero decode: per-unit type, length and digest,
a non-cryptographic coverage HINT, and the inertness determination (no executable construct exists).
Never reads a single octet of the ledger past the prefix. Satisfies TR-006, TR-007, TR-008, HC-006,
HC-007, HC-015.

**Flags:** none beyond S0's common set.

**Exit codes used:** 0, 1, 2, 7. (`inspect` never touches segment content past the prefix, ledger
integrity, signatures, or ceilings that need more than the prefix to evaluate, so 3, 4, 5 and 6 cannot
arise; a prefix-only structural defect, such as a bad magic, a commit-ring tie, PD-RING-001, or a
segment-table bounds-unsafe arithmetic failure, is `INVALID`.)

**stdout payload (`"prefix"` key):** header fields, the winning ring slot's fields, a frontmatter summary
(title/page_count/language/colour_profile_id; never the preview raster's decoded pixels), and the full
segment inventory with each entry's `{ordinal, type, offset, length, digest, coverage_hint}`. NORMATIVE:
`coverage_hint` is explicitly labelled non-authoritative in this payload (it mirrors
`SegmentTableSlot.flags` bit0 and Frontmatter's `fm-coverage-summary`, container.abnf S4, S5); `inspect`
never reports a positive verification indicator, because verification is never a bounded-prefix
operation (this is the same distinction container.abnf S4 draws for `fm-coverage-summary`).

## 4. `extract <file> [--to <path>]`

Reference extraction: reproduces the writer's own input text exactly, in structural reading order, with
no font, shaping, layout or graphics dependency (TR-011). Satisfies TR-011, NFR-012, NFR-013, NFR-014,
FR-035, FR-041, FR-042, FR-047, FR-048.

**Flags:** `--to <path>` (default: stdout's `"text"` field carries the content instead of a file being
written); `--locators` (adds `{block_id, run_id, base_ordinal}` per emitted span instead of a flat string).

**Exit codes used:** 0, 1, 2, 5, 7.

**stdout payload:** `{"text": "..."}` by default, or `{"out": "<path>"}` when `--to` is given, or
`{"spans": [{"block_id": "...", "text": "..."}]}` with `--locators`.

## 5. `verify <file> [--signature <id>]`

Runs EdDSA-Protodoc-1 (integrity.abnf S6) and the LTV chain (FR-070 through FR-073) for every SIGNATURE
frame in the document, or only the one named by `--signature`, plus the redaction-omission
classification (FR-076 through FR-078) for each. Satisfies FR-062 through FR-078, FR-104, FR-105,
CON-011.

**Flags:** `--signature <id>` (restrict to one signature's own `sig-signed-object` hex value); `--offline`
(force the LTV check to run exactly as it will thirty years hence: no network, a fixed local trust-anchor
list; this is the DEFAULT and only mode `verify` ever runs in, since network access in any reader path is
permanently out of scope, CP-005/CP-010 stack constraints, so `--offline` exists only to make the
guarantee explicit in scripts, not to toggle behaviour).

**Exit codes used:** 0, 1, 2, 3, 4, 5, 7. Never 6 (`verify` never writes).

**Exit-code selection when a document carries more than one signature:** per S1.1's ranking restricted
to this verb's own possible statuses (`INVALID` > `UNSUPPORTED` > `OVER_BUDGET` > `UNAVAILABLE` >
`UNVERIFIED` > `OK`), the process exit code reports the worst signature's status; `findings` and the
per-signature `signatures` array always report every signature's own individual verdict, never only the
worst one.

**stdout payload (`"signatures"` key):**

```json
{
  "verb": "verify", "status": "UNVERIFIED", "exit_code": 3, "findings": [
    {"rule_id": "FR-073", "unit_id": "6e11...", "message": "revocation predates signing instant"}
  ],
  "signatures": [
    {
      "id": "6e11...", "param_set_id": 1, "verdict": "unverified",
      "coverage": {"mode": "SUBSET", "covered_ranges": [[0, 12], [15, 20]]},
      "redaction": {"declared_omissions": [], "undeclared_omission_detected": false},
      "ltv": {"signing_instant_supported": true, "revocation_before_signing": true}
    }
  ]
}
```

`verdict` is one of exactly 3 values, matching S1's status names restricted to this context: `"valid"`,
`"unverified"`, `"unavailable_state"`. Non-negotiable #4 applies here directly: `verify` never displays a
signer identity alongside anything other than `"valid"`.

## 6. `diff <fileA> <fileB> [--format=json|text]`

Enumerates every content unit that differs between two document states (a generalisation of FR-068's
signed-state-vs-current-state comparison to any two documents or states of the same lineage). Satisfies
FR-068.

**Flags:** none beyond S0's common set (`--format` is already common to every verb, listed here because
`diff`'s human-readable mode is the one most often used interactively).

**Exit codes used:** 0, 1, 2, 5, 7. See S1.2: exit code 1 means `INVALID` (a malformed input), never
"differences were found."

**stdout payload:** `{"identical": false, "added": ["<unit_id>", ...], "removed": [...], "changed": [...]}`.
`identical: true` with an empty `added`/`removed`/`changed` set is a normal, successful `OK` result, not
a special case.

## 7. `merge <base> <a> <b> --out <path>`

Applies the exhaustive R1/R2/R3 concurrent-operation classification (document.abnf S9) to reconcile two
divergent descendants of one common base state. Satisfies FR-092, FR-093, FR-094, FR-095, FR-096,
FR-116, FR-024, HC-026.

**Flags:** `--out <path>` (required).

**Exit codes used:** 0, 1, 2, 4, 5, 7. (Never 3 or 6: merge performs no signature verification and no
policy write-refusal beyond the one collision case, which is `REFUSED` not a sixth category of its own.)

`REFUSED` here specifically means FR-024's case: two lineages independently minted the identical `run_id`
(an actual CSPRNG collision, astronomically unlikely per FR-023's own bound, but a real, tested,
required-refusal case, not an ordinary concurrent edit): this is refused rather than merged, because no
classification rule in S9 covers "the same identity means two different things."

**stdout payload:** `{"out": "<path>", "merged_state_id": "...", "operations_applied": {"disjoint_commute": 412, "r1_sequence_order": 38, "r2_tiebreak": 5, "r3_delete_dominates": 2}}`.

## 8. `project <file> --to <path> [--format=text|html]`

TR-004/TR-005's deterministic, one-directional projection of `canon(S)` to a non-normative text or
markup representation. Satisfies TR-004, TR-005, CP-008 ("tool-generated projections sit outside the
conformance surface and are never accepted as document input").

**Flags:** `--to <path>` (required); `--format` selects the projection's own shape (both are
non-normative outputs; the CHOICE of format has no bearing on this contract's normative surface, which
is only that the projection is deterministic and never re-ingestable).

**Exit codes used:** 0, 1, 2, 5, 7.

**stdout payload:** `{"out": "<path>", "format": "text", "reingestable": false}`. NORMATIVE: `reingestable`
is always `false`, present so a script can assert this contract's own guarantee rather than assuming it.

## 9. `redact <file> --subtree <unit-id> [--subtree <unit-id>...] --out <path>`

Executes a declared-redactable-subtree omission and emits a published copy via a fresh canonical
re-emission touching only reachable content (FR-078). Satisfies FR-074, FR-075, FR-076, FR-077, FR-078,
FR-080, FR-081, CQ-005.

**Flags:** `--subtree <unit-id>` (repeatable, required at least once); `--out <path>` (required).

**Exit codes used:** 0, 1, 2, 4, 5, 7.

`REFUSED` here covers: a named subtree was not designated redactable at signing time (no
`t-c-leaf-redactable` tag at that position, integrity.abnf S2.2); a named subtree's boundary is not
RED-ALIGNed to a document-model record boundary (integrity.abnf S7.1); the document carries no signature
at all (there is nothing to redact-and-still-attest; redacting an unsigned document is just `publish`
with content removed, not this verb's contract); or a named subtree straddles a boundary that would
leave a present SUBSET-mode signature's `covered_segment_ranges` pointing at a mix of removed and
retained ordinals within one previously-atomic range entry.

**stdout payload:** `{"out": "<path>", "redacted_subtrees": ["<unit_id>", ...], "surviving_signatures": ["<sig_id>", ...], "orphaned_annotations": ["<unit_id>", ...]}`.

## 10. `publish <file> --out <path> [--partial]`

Full compaction / canonicalisation to a fresh file (`place(BOTTOM, S) = L*(S)`, container.abnf S8),
refused whenever any signature is present, UNLESS `--partial` invokes DP-016's PARTIAL COMPACTION
instead, which relocates or reclaims only segments outside every present signature's
`covered_segment_ranges`. Satisfies NFR-004, CQ-003, DP-016.

**Flags:** `--out <path>` (required); `--partial` (invoke partial rather than full compaction).

**Exit codes used:** 0, 1, 2, 4, 5, 7.

`REFUSED` here covers: full compaction (`--partial` absent) attempted while any signature is present
(NFR-004, unconditional); `--partial` given but every present signature is TOTAL-mode, so there is
nothing uncovered to reclaim (research.md S8 item 7's disclosed residual: `publish --partial` on such a
document reports `REFUSED` naming the observed `MAX_SEGMENTS` ceiling and the current segment count,
distinct from silently doing nothing).

**stdout payload:** `{"out": "<path>", "mode": "full", "segments_reclaimed": 0, "octets_reclaimed": 0, "state_id_unchanged": true}`.
NORMATIVE: `state_id_unchanged` MUST be `true` for both `full` and `partial` mode on a successful run
(compaction never changes a document's logical content, therefore never changes its T_C_root); a `false`
value here would itself indicate a defect in the compaction implementation, not a legitimate outcome.

## 11. `sign <file> --key <ref> --coverage total|subset [--subset-range <start>:<end> ...] --intent <value> --out <path>`

Creates a new SIGNATURE frame (integrity.abnf S5) plus its required `ATTESTATION_EVIDENCE` references
(integrity.abnf S9: credential chain and time attestation are mandatory, revocation evidence is
mandatory when one was current at signing time) and its `PRESENTATION_ARTEFACT` reference
(document.abnf S7.4). Satisfies FR-063, FR-064, FR-065, FR-066, FR-067, FR-068, FR-069, FR-070, FR-071,
FR-089, CQ-004, CQ-005.

**Flags:** `--key <ref>` (required, the signing key material; the tool never accepts a raw private key
on the command line as plaintext, per this contract's own credential-handling posture: `<ref>` names a
key held by the platform's own key store or HSM, never an inline value); `--coverage total|subset`
(required); `--subset-range <start>:<end>` (repeatable, required at least once when `--coverage subset`
is given, forbidden with `--coverage total`; each names a half-open segment-ordinal range to include in
`covered_segment_ranges`, integrity.abnf S4); `--intent <value>` (required, one of integrity.abnf S10.3's
4 closed values); `--out <path>` (required).

**Exit codes used:** 0, 1, 2, 4, 5, 7. (Never 3 or 6 in the everyday sense: `sign` does not itself verify
an existing signature, so `UNVERIFIED` does not arise from its own action; `UNAVAILABLE` likewise applies
only to verifying a signature over a state that no longer exists, not to creating a new one.)

`REFUSED` here covers: `MAX_SIGNATURES = 64` already reached (integrity.abnf S5); a `--subset-range`
naming an ordinal whose current `SegmentTableSlot.slot-segment-type = ATTEST` (PD-COVER-004, never
nameable in any coverage descriptor); required LTV evidence (a reachable, currently-valid credential
chain, a time attestation) is not obtainable at signing time (FR-070's carriage requirement cannot be
satisfied after the fact, so `sign` refuses rather than emit a signature with a missing mandatory
reference).

**stdout payload:** `{"out": "<path>", "signature_id": "<unit_id>", "param_set_id": 1, "signed_object": "<hex>", "coverage": {"mode": "SUBSET", "covered_ranges": [[0, 12]]}}`.

## 12. `migrate <file> --to-major <N> --out <path> [--rescind-and-resign --new-key <ref> --new-param-set <id>]`

Refusal-first, two-phase major-version migration (DP-012): phase 1 checks every construct is
representable under major version `N` and halts, writing nothing, naming the first unrepresentable
construct if any exists; phase 2, only if phase 1 found nothing, emits `L*(migrate(S))` as a fresh file
with every content-unit identifier passed through unchanged. `--rescind-and-resign` additionally invokes
DP-017 (integrity.abnf S8), adding a new signature under a new allowlist scheme over the migrated content
while the prior signature and its metadata remain retained, unmodified, in the preserved pre-migration
state. Satisfies FR-119, FR-120, FR-121, FR-122, FR-123, CP-008, DP-017.

**Flags:** `--to-major <N>` (required); `--out <path>` (required); `--rescind-and-resign` (optional,
requires `--new-key` and `--new-param-set`); `--new-key <ref>`; `--new-param-set <id>` (integrity.abnf
S10.2).

**Exit codes used:** 0, 1, 2, 4, 5, 7.

`REFUSED` here covers: phase 1 finds an unrepresentable construct (the migration halts before any output
octet is written, per FR-121, naming the construct); `--to-major <N>` names a version not greater than
the file's current `format-major` (a migration is always forward); `--rescind-and-resign` given without
`--to-major` naming an actual major-version increase (DP-017 restricts the operation to a migration
boundary specifically, never an ordinary edit).

**stdout payload:** `{"out": "<path>", "from_major": 1, "to_major": 2, "identifiers_preserved": true, "rescind_resign": null}`,
or with the flag: `"rescind_resign": {"prior_signature_id": "...", "new_param_set_id": 2, "new_signature_id": "..."}`.

## 13. Traceability summary

| Verb | Primary requirements satisfied |
|---|---|
| `validate` | FR-102-110, FR-117, CON-003, CON-009-011, TR-006/007/008 (structural portion) |
| `inspect` | TR-006, TR-007, TR-008, HC-006, HC-007, HC-015 |
| `extract` | TR-011, NFR-012, NFR-013, NFR-014, FR-035, FR-041, FR-042, FR-047, FR-048 |
| `verify` | FR-062-078, FR-104, FR-105, CON-011 |
| `diff` | FR-068 |
| `merge` | FR-092-096, FR-116, FR-024, HC-026 |
| `project` | TR-004, TR-005, CP-008 |
| `redact` | FR-074-078, FR-080, FR-081, CQ-005 |
| `publish` | NFR-004, CQ-003, DP-016 |
| `sign` | FR-063-071, FR-089, CQ-004, CQ-005 |
| `migrate` | FR-119-123, CP-008, DP-017 |

This table is the seed of `analysis.md`'s phase-5 traceability matrix (requirement -> task -> test ->
file); it is not itself that matrix, and every row above still needs a `T-NNN` task identifier and a
named test before phase 6 (Implement) may begin, per CP-001.

## 14. Requirement-to-verb traceability table

S13 above maps each verb to the requirements it satisfies; this section inverts that mapping to one row
per requirement identifier, restricted to the FR/NFR/CON/TR ids S13 actually names (CP-*, CQ-*, DP-* and
HC-* ids cited in S13's prose are constitution/clarification/plan/hard-constraint identifiers, not the
FR/NFR/CON/TR scheme this table tracks, and are out of scope here). `pkg/cli/traceability.go` is a
verbatim copy of this table's two columns, and `TestTR_012_TraceabilityTableMatchesDispatch`
(`pkg/cli/traceability_test.go`) re-reads this table off disk at test time and fails the build the moment
it disagrees with that copy, or names a verb the dispatch registry (`pkg/cli.Names()`, T-0325) does not
have registered, or a registered verb is missing from every row's Verb(s) column.

| Requirement ID | Verb(s) |
|---|---|
| FR-024 | merge |
| FR-035 | extract |
| FR-041 | extract |
| FR-042 | extract |
| FR-047 | extract |
| FR-048 | extract |
| FR-062 | verify |
| FR-063 | sign, verify |
| FR-064 | sign, verify |
| FR-065 | sign, verify |
| FR-066 | sign, verify |
| FR-067 | sign, verify |
| FR-068 | diff, sign, verify |
| FR-069 | sign, verify |
| FR-070 | sign, verify |
| FR-071 | sign, verify |
| FR-072 | verify |
| FR-073 | verify |
| FR-074 | redact, verify |
| FR-075 | redact, verify |
| FR-076 | redact, verify |
| FR-077 | redact, verify |
| FR-078 | redact, verify |
| FR-080 | redact |
| FR-081 | redact |
| FR-089 | sign |
| FR-092 | merge |
| FR-093 | merge |
| FR-094 | merge |
| FR-095 | merge |
| FR-096 | merge |
| FR-102 | validate |
| FR-103 | validate |
| FR-104 | validate, verify |
| FR-105 | validate, verify |
| FR-106 | validate |
| FR-107 | validate |
| FR-108 | validate |
| FR-109 | validate |
| FR-110 | validate |
| FR-116 | merge |
| FR-117 | validate |
| FR-119 | migrate |
| FR-120 | migrate |
| FR-121 | migrate |
| FR-122 | migrate |
| FR-123 | migrate |
| NFR-004 | publish |
| NFR-012 | extract |
| NFR-013 | extract |
| NFR-014 | extract |
| CON-003 | validate |
| CON-009 | validate |
| CON-010 | validate |
| CON-011 | validate, verify |
| TR-004 | project |
| TR-005 | project |
| TR-006 | inspect, validate |
| TR-007 | inspect, validate |
| TR-008 | inspect, validate |
| TR-011 | extract |

Every one of TR-012's 11 verbs (`validate`, `inspect`, `extract`, `verify`, `diff`, `merge`, `project`,
`redact`, `publish`, `sign`, `migrate`) appears in the Verb(s) column of at least one row above; the CI
check named at the top of this section asserts this mechanically rather than by inspection.
