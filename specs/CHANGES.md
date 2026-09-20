# Protodoc spec — governance & release-gate change log

This file records structured, dated governance decisions that the frozen `spec.md` cannot carry
(spec.md is frozen; these are `plan.md`/release-gate decisions that reference it). Each entry names its
decision-maker honestly. Entries requiring **Eyvar García's** explicit ruling are recorded as **OPEN**
until that ruling genuinely occurs — no decision-maker is fabricated.

## Status of the M19 release-gate governance items

The following M19 tasks each require an action that CANNOT be performed by the implementing agent and
have therefore NOT been marked complete. They are recorded here honestly as OPEN, with what is blocked
and what would unblock each. This mirrors the project's standing precedent (T-0267 recorded a design
ruling OPEN rather than fabricate an Eyvar sign-off; T-0299 recorded the two-implementation trial OPEN;
T-0306 recorded an author-conducted review without fabricating an external audit).

### T-0347 — NFR-027 rendering schedule-risk ruling (OPEN, awaiting Eyvar García)

- **Status:** OPEN. plan.md Section 9 Conflict 4 flags NFR-027's 30-day rendering budget as at-risk
  (revised estimate 28–49 days). Resolving it requires **Eyvar García's** explicit ruling (accept the
  revised bound / descope / accept the miss).
- **Blocked on:** an explicit decision by Eyvar García. The task's test asserts
  `decision-maker == 'Eyvar García'`; that name must not be written on a decision the author invented.
- **Unblocks when:** Eyvar records the ruling (date, decision-maker, decision, rationale) here and in
  plan.md Section 9.

### T-0349 — NFR-028 / CQ-012 two-implementation funding decision (RESOLVED)

- **Date:** 2026-09-18 (path-forward revised 2026-09-19)
- **Decision-maker:** Eyvar García
- **Path-forward:** open volunteer call, unfunded. No cash bounty; instead, a public GitHub issue and
  outreach (open-source/PL/crypto communities, CS department capstone-project channels) inviting a
  volunteer to build the second independent implementation of container/validate/canon, per the T-0348
  acceptance protocol. Compensation is non-monetary: named credit as CP-003's second implementer,
  permanently recorded in the project's governance record.
- **Fallback-if-unfunded:** if no volunteer emerges within a reasonable window, T-0351 remains deferred
  indefinitely; T-0358's v1-stable capstone records CP-003/NFR-028 as an open blocking gate rather than
  proceeding without it satisfied.
- Mirrored in plan.md Section 8's risk row for this item.

### plan.md Conflict 5 — TR-010 storage-backend abstraction adopted (RESOLVED)

- **Date:** 2026-09-19
- **Decision-maker:** Eyvar García
- **Decision:** adopted the already-implemented `ledger.ConditionalWriter` interface
  (`pkg/ledger/conditionalwriter.go`, T-0044/T-0045) as `plan.md`'s architecture-level answer to TR-010.
  The frozen plan as originally written described no storage-backend abstraction; M02 implemented one
  anyway and M18's CLI (T-0341) built on top of it before the plan document was updated to match. This
  decision closes that gap by formally recording the mechanism in `plan.md` Section 9 Conflict 5.
- **Rationale:** the code was already correct, tested, and in production use (`TestTR_010_
  ConditionalWriterInterfaceContract`, `TestTR_010_ConditionalWriteRefusalNamesCurrentHolder`,
  `TestTR_010_ConditionalWriteRefusesOnMismatch`); this is a documentation/architecture-record fix, not
  a design or wire-format change.
- Mirrored in `plan.md` Section 9 Conflict 5 and `docs/ledger-conditional-write-design-note.md`.

### T-0356 — CON-026 licence/governance gate approval (RESOLVED)

- **Date:** 2026-09-19
- **Decision-maker:** Eyvar García
- **Decision:** approved
- **Referenced-files:** LICENSE, LICENSE-CODE, GOVERNANCE.md
- All four of CON-026's required elements approved as drafted: the Licence Grant (LICENSE for the
  specification, LICENSE-CODE/Apache 2.0 for the reference implementation), Eyvar García as Named
  Steward, the Succession Process (180-day unresponsiveness trigger, 30-day comment period, majority
  vote if contested), and the Deprecation Window Policy (3-year minimum). `GOVERNANCE.md`'s status
  updated to APPROVED accordingly.

### T-0345 — NFR-026 external extract+validate trial (OPEN, recruiting)

- **Status:** OPEN. Requires an **outside implementer** to build a passing extracting-and-validating
  reader from the spec text alone within 5 working days, with actual elapsed days recorded. No such
  external trial has been run; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#2](https://github.com/LightKnight23/protodoc/issues/2).
- **Unblocks when:** an outside implementer runs the trial and a dated report records elapsed days and
  the pass/fail result against the extract+validate conformance corpus.

### T-0365 — NFR-026 extracting-and-validating trial-miss ruling (OPEN or N/A, awaiting inputs)

- **Status:** OPEN pending the T-0345 trial outcome. If the T-0345 extracting-and-validating reader trial
  passed within its 5-working-day budget, this task is NOT APPLICABLE (to be logged as such once the
  trial result is confirmed). If it missed budget or failed the corpus, an explicit ruling by
  **Eyvar García** is required and is recorded OPEN here.
- **Blocked on:** the confirmed T-0345 result, then (if a miss) Eyvar's ruling.

### T-0346 — NFR-027 external rendering trial (OPEN, recruiting)

- **Status:** OPEN. Requires an **outside implementer** to build a passing rendering reader from the
  spec text alone within 30 working days, with actual elapsed days recorded. No such external trial has
  been run; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#3](https://github.com/LightKnight23/protodoc/issues/3).
- **Unblocks when:** an outside implementer runs the trial and a dated report records elapsed days and
  the pass/fail result against the rendering conformance corpus.

### T-0351 — NFR-028 second-implementation differential trial (OPEN, recruiting)

- **Status:** OPEN. Requires a genuinely **second, independently-authored implementation** of
  container/validate/canon, run through the T-0350 differential harness against the full corpus. No
  second implementation exists; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#1](https://github.com/LightKnight23/protodoc/issues/1), per the T-0349
  unfunded-open-volunteer-call decision.
- **Unblocks when:** a volunteer is found, the implementation is built, and it passes (or has every
  mismatch triaged) through the T-0350 harness.

### T-0371 — CP-014 IANA media-type / PRONOM format registration (IN PROGRESS)

- **Status:** IN PROGRESS. Repository made public at
  https://github.com/LightKnight23/protodoc (default branch `001-protodoc-format-core`), with GitHub
  Pages enabled and verified serving `spec.md` at a stable public URL, so both filings below have a
  real "published specification" link to cite.
- **IANA media-type registration:** submitted 2026-09-19 by Eyvar García to media-types@iana.org
  (subtype `application/vnd.protodoc`, vendor tree per RFC 6838 Section 5.3.4). Awaiting IANA Designated
  Expert review; no tracking ID assigned yet by IANA as of this entry.
- **PRONOM format registration:** submitted 2026-09-19 by Eyvar García to PRONOM@nationalarchives.gov.uk,
  via the official PRONOM Submission template (docx), with a real, decoder-valid, minimal `protodoc-sample.pdl`
  attached for signature testing. Signature independently verified beforehand with The National Archives'
  own DROID tooling: `sigtool` byte-match (1/1 hit on the sample, 0/1 on a control file) and a full DROID
  identification run using a locally patched copy of the real DROID_SignatureFile_V119.xml (correctly
  identified the sample, correctly left the control file unidentified as Protodoc). Awaiting PRONOM team
  review; no PUID assigned yet as of this entry.
- **Unblocks when:** IANA approves and publishes the registration (tracking: the published registry
  entry) AND PRONOM assigns a PUID. Both dates/reference IDs get recorded here, and
  TestFR_125_MediaTypeRegistrationOnRecord is written and made to pass, before this task closes.

## DEFECT-2026-09-19 — CLI verbs never wired to real decode chain (found post-M18)

**Severity:** high (correctness). **Found:** 2026-09-19, external review. **Status:** RESOLVED 2026-09-19
(all 10 stub verbs wired, T-0373..T-0382, verified against real commits and a rebuilt binary that
correctly rejects a nonexistent/garbage/truncated file and accepts a genuine sample).

M18 (CLI Surface) was marked 17/17 with all tests green, but the compiled `protodoc` binary does not
process real files for 10 of its 11 verbs. Reproduction (current tree, before fix):

```
$ go build -o /tmp/protodoc ./cmd/protodoc
$ /tmp/protodoc validate /tmp/this-file-does-not-exist-at-all.pdl
{"checks":0,"exit_code":0,"findings":[],"status":"OK","verb":"validate"}
```

A file that does not exist reports OK. **Root cause:** every verb except `inspect` routes through a
package-level "injectable backend" variable (`ValidateStepsFor`, `ProjectStateFor`, `ExtractRun`,
`VerifyRun`, `DiffRun`, `MergeRun`, `RedactRun`, `PublishRun`, `SignRun`, `MigrateRun`) that DEFAULTS to
a no-op stub. The stubs are legitimate test-doubles, but the production wiring that replaces them with
the real M01–M17 decode chain was never written, and `cmd/protodoc/main.go` imports only `pkg/cli`.
`go test ./...` stayed green because the verb tests exercise the stubs directly, never real file I/O.
Only `pkg/cli/inspect.go` genuinely opens the file and decodes real bytes.

This defect is NOT covered by any T-NNN task in the original spine (all 17 M18 tasks tested the stub
seam, not the production seam). It is fixed under new numbered tasks **T-0373..T-0382** (one per stub
verb) added to `tasks.md`'s M18 section, one verb per commit, each wiring the stub to the real decode
chain (following `inspect.go`'s pattern) plus a real test proving it rejects a garbage/truncated/absent
file and accepts a genuine sample. `validate` is done first because CP-006 requires every other verb to
run validation as a precondition. No heuristic recovery, no partial output past the first structural
failure (CP-006/FR-103).

### GAP-VERIFY-CONTENT-REBUILD (CLOSED 2026-09-19) — whole-document ContentRecord decode path shipped

While wiring `verify` (T-0374) honestly, a genuine missing capability surfaced: there was no function
that reads a whole document's CONTENT segment bodies from disk and decodes each frame into an
`integrity.ContentRecord` (unit-id + canonical frame bytes). `integrity.TCRoot(records)` could recompute
the content-commitment tree, but nothing assembled `records` from a real file.

**CLOSED.** `extract.LoadContentRecords(r)` + `extract.DecodeContentFrameUnitID` now read every CONTENT
segment body and decode each content-model frame into its AUTHORED unit-id (the shared tag=1 field,
document.abnf S2-S5) plus its canonical frame octets, via the bounded streaming Walk + `pdlfmt` TLV
decode (no font/image/crypto/merge facility, CP-006; a malformed frame is a hard error, FR-103).

Consequences, now shipped:
- `verify` REBUILDS the content-commitment tree from the real decoded ContentRecords
  (`integrity.TCRoot`) and compares it to the winning commit-ring record's recorded `T_C_root` to decide
  reconstructability; a document whose real content does not hash to its recorded root no longer verifies
  as covered. A CLI test proves the recomputed tree over loaded records equals the genuine tree.
- `project` and `redact` now key constructs by the AUTHORED unit-id decoded from each frame, not a
  slot-digest surrogate.

Remaining honest scope (narrower, its own future concern, NOT this gap): full EdDSA byte-verification of
`sig-value` against the recomputed signed_object needs the signer's public key from the credential chain
(the M09/M10 offline-evidence path). The CLI verify surface reports the state-coverage verdict derived
from the recomputed tree and defers that credential-chain crypto to the offline-evidence path; it does
not fabricate a cryptographic pass.

## DEFECT-2026-09-19b — CLI write verbs report OK without writing output; required flags unenforced; two contract-shape mismatches

**Severity:** high (correctness). **Found:** 2026-09-19, via an extensive black-box smoke test
(42 cases across all 11 verbs and both happy-path and failure scenarios) run independently after
DEFECT-2026-09-19's fix landed, to verify the fix rather than trust it. 31/42 passed; the 11 failures
are four distinct, real bugs, none of them touching DEFECT-2026-09-19's fix (which is genuinely solid:
read/decode/reject-garbage all work correctly now). **Status: RESOLVED 2026-09-19**, fixed under
T-0383..T-0390 (one task per bug, one commit per task, verified via the full test suite plus manual
end-to-end runs re-testing every reproduction below against the fixed binary).

1. **No write verb actually writes its output file.** `project --to`, `merge --out`, `redact --out`,
   `publish --out`, `sign --out`, `migrate --out` all report `"status":"OK"` with a plausible-looking
   payload, but the destination path is never created on disk in any of the 6 cases. Reproduction:
   ```
   protodoc publish valid-sample.pdl --out /tmp/published.pdl
   {"custody_preserved":true,"exit_code":0,"findings":[],"residue_octets":0,"status":"OK","verb":"publish"}
   $ ls /tmp/published.pdl
   ls: /tmp/published.pdl: No such file or directory
   ```
2. **Required flags are not enforced.** `project` (`--to`), `redact` (`--subtree`, `--out`), `publish`
   (`--out`), `sign` (`--key`, `--coverage`), `migrate` (`--to-major`) all silently "succeed" (`OK`,
   exit 0) when a flag `cli.md` documents as required is omitted entirely, instead of `USAGE` (exit 7).
3. **`merge` returns `REFUSED` (exit 6) for a missing input file.** `cli.md` S7's own text restricts
   `merge`'s `REFUSED` status to exactly one case (FR-024's identifier-collision refusal) and lists
   `merge`'s exit codes used as "0, 1, 2, 4, 5, 7" — 6 is not among them. A missing/unreadable base
   file returned `{"status":"REFUSED","exit_code":6,...}` instead of `INVALID` or `USAGE`.
4. **`diff`'s stdout payload does not match `cli.md` S6's documented shape.** Contract requires
   `{"identical": bool, "added": [...], "removed": [...], "changed": [...]}`; the real payload is
   `{"change_count":0,"changed_constructs":null}` — no `identical` field at all, so a caller scripting
   against the documented contract cannot read the result.
5. **`inspect` and every other verb disagree on how to report a nonexistent file**, and `cli.md`'s own
   text settles which is right: S1's `USAGE` definition explicitly lists "an unreadable path" as a
   `USAGE` case (exit 7), which is what `inspect` already correctly returns. The 10 verbs DEFECT-2026-09-19
   just fixed instead return `INVALID` (exit 1) for a nonexistent file — correct for a file that exists
   but is structurally malformed (confirmed still correct for garbage/truncated-but-present files in the
   same test run), wrong for a file that isn't there at all.

Filed as tasks **T-0383..T-0390** in `tasks.md`'s M18 section (one task per bug, T-0390 covering the
cross-verb USAGE-vs-INVALID unification for item 5). The 42-case smoke-test script and its full log are
not committed (ephemeral verification artifacts); the reproduction above is sufficient to re-derive them.

**Resolution summary:**
1. `project`/`merge`/`redact`/`publish`/`sign`/`migrate` now actually write their `--out`/`--to` file
   (T-0383, T-0385, T-0386, T-0387, T-0388; `merge`'s real merged-content write and full SIGNATURE-segment
   splicing for `sign` remain honestly-scoped narrower gaps, documented in their own commits/code comments
   -- writing fabricated output there would be worse than not writing it, per this project's honesty rule).
2. Required flags are now enforced (`USAGE` if missing) on `project`, `redact`, `publish`, `sign`, `migrate`
   (same tasks as above).
3. `merge` now reports `INVALID` for a missing/unreadable input, never `REFUSED` (T-0384).
4. `diff`'s stdout now carries the documented `{identical, added, removed, changed}` fields (T-0389),
   alongside the pre-existing fields for backward compatibility.
5. Every verb now reports `USAGE` for a nonexistent/unopenable file, matching `inspect`'s original
   correct behaviour, via a shared `ErrFileUnreadable` sentinel (T-0390). A file that exists but is
   structurally malformed is unaffected and still correctly reports `INVALID`.

Independently verified (not self-reported): full `go build`/`go vet`/`go test ./...` green, plus manual
end-to-end re-runs of every reproduction command above against the fixed binary.

## DEFECT-2026-09-19c — merge's clean case and sign's --out never wrote real output (a narrower, honestly-scoped follow-on to DEFECT-2026-09-19b)

**Severity:** medium (functionality gap, not a false-success report -- both cases were already disclosed
plainly, never silently claimed complete). **Found:** 2026-09-19, while scoping how to close the two
honest gaps DEFECT-2026-09-19b's resolution deliberately left open. **Status:** RESOLVED 2026-09-19
(merge: T-0391 then T-0393; sign: T-0392, including a follow-up ring-reissue bug found and fixed during
independent re-verification -- see below). No disclosed gap remains open in this defect.

**merge (RESOLVED, T-0391):** the previous fix only corrected `merge`'s exit code (T-0384); it still
classified conflicts from ordinal-keyed SegmentTable digests (a storage-position heuristic, not real
content identity) and never wrote a merged file at all, nor enforced `--out`. Replaced with a genuine
per-construct three-way merge (`pkg/merge.ThreeWayMerge`) over content decoded via
`extract.LoadContentRecords` and keyed by the AUTHORED unit-id -- the same identity `project`/`redact`/
`publish` already use post GAP-VERIFY-CONTENT-REBUILD. A real divergent edit now reports a genuine
CONFLICT naming both real values (verified with a test that constructs two real conflicting documents);
a clean merge is re-canonicalized and written to the now-enforced required `--out` (verified by reading
the written file back and confirming it contains the real changed content, not a stub or a copy).
Disclosed remaining gap at the time: this wired `ThreeWayMerge` but not the full `pkg/merge.Orchestrate`
precondition set -- CON-024 (retention-point crossing) and FR-096 (erased-unit replay) still needed real
History/Erasure segment decoding, which no CLI verb performed yet. Closed below by T-0393.

**merge CON-024/FR-096 (RESOLVED 2026-09-19, T-0393).** Added `pkg/cli/historyreader.go`
(`DiscoverErasureRecords`), a HISTORY-segment reader mirroring `attestreader.go`'s ATTEST-segment
convention: each HISTORY segment carries one `ERASURE_RECORD` (FR-061), decoded via the package's own
`history.DecodeErasureRecord`. `loadMergeState` now also decodes the winning commit-ring record's real
`RetentionPoint` and each authored unit's real CONTENT-segment storage ordinal. `realMergeRun` builds the
base's `history.Declaration` and an `merge.ErasureIndex` from its real erasure records, then for each
incoming branch: every genuinely changed construct (via `merge.DiffConstructs`, never a storage-position
heuristic) is checked with the already-existing `merge.CheckReplayOfErasedUnit` (digest match against a
real erased identity -> REFUSED naming FR-096) and its real CONTENT ordinal is checked with the
already-existing `merge.CheckRetentionPoint` (ordinal predates the declared retention point -> REFUSED
naming CON-024). Both primitives already existed in `pkg/merge` (built by earlier M13 tasks) and were
fully unit-tested there; T-0393's work was building the real HISTORY-segment decode path the CLI needed to
feed them, not the guard logic itself. Verified with `TestCON_024_MergeRefusesAcrossRetentionPoint` (a
change at a CONTENT ordinal predating a real declared retention point is refused; a change at/after it
merges normally) and `TestFR_096_MergeRefusesReplayOfErasedUnit` (a branch reproducing a real erased
unit's exact recorded digest is refused; the same identity with genuinely different content is a
permitted new authoring act). No known gap remains in `merge`'s precondition coverage.

**sign (OPEN, investigated, handed off as T-0392):** building a real signed-document write turned out to
need more foundational plumbing than merge did, discovered by direct investigation of `pkg/integrity`:

1. `SIGNATURE`/`ATTESTATION_EVIDENCE` records use PDL-TLV's discriminant as an actual **tag=0 field**
   inside the record (`pdlfmt.DecodeRecord(src, known, -1)` decodes it directly); content-model frames
   use a **raw leading discriminant byte** with the TLV body starting after it
   (`extract.DecodeContentFrameUnitID`: `body := frame[1:]`). These are two different framing
   conventions in this same codebase; no ATTEST-segment reader analogous to
   `extract.LoadContentRecords` exists yet, and building one incorrectly risks silently misparsing
   exactly the record type most worth getting right (a cryptographic signature).
2. `SignatureRecord.Encode()` already enforces (correctly) that `sig-cred-chain-ref` and
   `sig-time-attestation-ref` must never be `zero16` -- real evidence must already exist in the document
   as `ATTESTATION_EVIDENCE` records (`ae-id`, tag=1, a genuine minted unit-id) for `sign` to reference.
   `cli.md`'s own S11 already specifies the right behavior for the no-evidence case: `sign` REFUSES
   rather than fabricating a reference -- this is not a new design decision, just an unbuilt one.
3. No verb builds a new segment and appends it to the segment table + reissues a commit-ring record
   yet. Even `migrate`'s "clean" path (T-0388) only carries the existing prefix forward unchanged; a
   real `sign` needs genuinely new segment-table-growth plumbing no CLI verb has built before.

This is closer in scope to several of the original M09 implementation tasks than to a bug fix, and
touches the format's actual cryptographic signing path, so it is handed off with this concrete
investigation rather than rushed. Filed as **T-0392**.

**sign (RESOLVED 2026-09-19, T-0392).** All three pieces built:

1. **ATTEST-segment record reader** (`pkg/cli/attestreader.go`): decodes ATTEST records with the correct
   framing (discriminant is a real tag=0 field via `pdlfmt.DecodeRecord(body, nil, 0)`), distinct from
   the raw-leading-byte content-frame framing. `DiscoverEvidence` scans every ATTEST segment for
   `ATTESTATION_EVIDENCE` records and indexes their `ae-id` by `AeKind`.
2. **Refusal path** (cli.md S11): when a credential-chain or time-attestation evidence record is not
   present, `sign` REFUSES (`FR-070` finding, no `--out` written) — no evidence is fabricated.
3. **Embed path**: when both required evidence records are present, `sign` builds a real
   `SignatureRecord` (total coverage, referencing the discovered `ae-id`s and the real deterministic
   EdDSA-Protodoc-1 signature), encodes it, frames it in a ledger ATTEST segment, appends it as a NEW
   segment-table slot, recomputes the segment-table digest and `structure_digest`, reissues the winning
   commit-ring record (sequence+1, segment_count+1, new ledger_length, parent = prior state), and writes
   the complete re-serialized signed document to `--out`. `T_C_root` is carried forward unchanged
   (signing never alters content). The `"signed_document_complete": false` caveat and its finding are
   removed; the response now reports `"signed_document_complete": true`. Verified by
   `TestTR_012_SignVerbEmbedsRealSignature`: the no-evidence document is REFUSED with no output; the
   evidence-bearing document yields a file larger than the input that decodes with one more ATTEST
   segment (the spliced SIGNATURE, which itself decodes and references the real evidence), and signing
   twice is byte-identical (NFR-006).

**Follow-up bug found and fixed during independent re-verification (2026-09-19, same day):** the commit-
ring reissue wrote the same new record into all 7 ring slots, giving every slot an identical, tied
sequence number -- `PD-RING-001` correctly rejects this as a structural ambiguity, so the "signed"
output failed `validate`/`inspect` outright despite `sign` itself reporting `OK`. Caught by generating a
real evidence-bearing fixture, signing it with the actual compiled binary, and round-tripping the result
through `validate` -- not by inspecting the diff. Fixed to write the reissued record into exactly the
one ring slot the round-robin convention names (`sequence % 7`, per `pkg/ledger/writecost.go`'s own
documented scheme), leaving the other 6 slots' prior records untouched. `TestTR_012_SignVerbEmbedsRealSignature`
now also round-trips its embed-path output through `runValidate` to catch this class of regression
directly, since the original test decoded the SIGNATURE record in isolation without ever confirming the
whole document it lived in was still valid.

## DEFECT-2026-09-20 — verify never stripped an ATTEST segment's ledger header, so every real ATTEST record (SIGNATURE included) failed to decode

**Severity:** medium (a real signature was always mis-reported "unverified"; not a false-success report —
UNVERIFIED is the honest, conservative status, just for the wrong reason). **Found:** 2026-09-20, while
investigating a plainly-flagged pre-existing limitation ("verify doesn't distinguish SIGNATURE from
ATTESTATION_EVIDENCE records") and reproducing it against a real signed document rather than trusting the
description at face value. **Status:** RESOLVED 2026-09-20, T-0394.

Manual reproduction against a real `sign`-produced document showed all 4 ATTEST segments (3
`ATTESTATION_EVIDENCE` + 1 real `SIGNATURE`) reporting `"verdict":"unverified"` — including the genuine
signature, not just the unrelated evidence records the original description named. Root cause:
`verify_real.go` read each ATTEST segment's raw octets and passed them straight to
`integrity.DecodeSignatureRecord` without stripping the leading 64-octet ledger segment header
(`ledger.EncodeSegmentHeader`'s "PDS1"-magic framing, the same framing `sign` itself writes and
`attestreader.go` already strips correctly for evidence discovery) — so decode failed on every ATTEST
segment regardless of kind, real signature included. Separately, `verify` was also reporting one
`signatures` array entry per ATTEST segment rather than per SIGNATURE frame, contradicting cli.md S5's
own contract ("one entry per SIGNATURE frame... `ATTESTATION_EVIDENCE` is not itself a signature").

**Fix:** `verify_real.go` now reuses `attestreader.go`'s own `attestSegmentBody` (header strip) and
`readAttestRecordDiscriminant` (kind dispatch), decoding only records whose discriminant is `0x40`
(SIGNATURE) and silently skipping `0x42` (ATTESTATION_EVIDENCE) segments entirely — no fabricated verdict
for a record type verify was never asked to verify. Verified with
`TestTR_012_VerifyDistinguishesSignatureFromEvidence`: a real signed document carrying 3 evidence records
and 1 signature now reports exactly one `signatures` entry, and that entry is no longer `"unverified"`
from a decode failure.

**New, separate, honestly-disclosed limitation surfaced (NOT fixed here, out of this defect's scope):**
the real signature now decodes and reaches `integrity.SignedObjectForSignature`, which currently always
returns `ErrPresentationRefUnresolved` for `sign`'s output — `sign` leaves `sig-presentation-ref` at
`zero16` for total-coverage signatures (`signmigrate_real.go`'s own comment: "no presentation artefact
bound in this minimal signed document"), but `SignedObjectForSignature`'s docstring states a zero16
presentation ref is *unconditionally* unresolved, with no total-coverage exception. The practical effect:
a `sign`-produced signature currently verifies as `"covering_unavailable_state"` (`UNAVAILABLE`, exit 4),
never `"valid"`, even when the state is fully reconstructable. This is a real gap in either `sign` (it may
need to mint a real, empty/no-op PRESENTATION_ARTEFACT segment for total coverage) or in
`SignedObjectForSignature`'s contract (it may need a documented total-coverage exception) — a design
question for `sign`/`integrity`, not a `verify`-side bug, and not something to guess at silently.

## DEFECT-2026-09-20b — diff never populated/compared a real content digest, so same-count differing documents always reported identical

**Severity:** medium (a real, silent false-negative: `diff` could report `"identical": true` for two
documents whose actual content genuinely differs). **Found:** 2026-09-20, running a from-scratch
end-to-end CLI battery against the real compiled binary and real fixture files, specifically covering a
case none of the existing tests exercised (same CONTENT segment count, different real payload). **Status:**
RESOLVED 2026-09-20, T-0395.

`diff_real.go` classified a changed construct by comparing `SegmentTableSlot.Digest` between the two
documents' segment tables. Nothing in any writer path in this codebase ever populates that field with a
real content digest for CONTENT segments — every fixture across the test suite, and every document
`merge`/`sign`/`migrate` write, leaves it zero. Two documents with the same segment count therefore always
compared zero-equals-zero and were reported identical, regardless of what their real content actually
was. The only existing test (`TestTR_012_DiffVerbReadsRealFiles`) never caught this because its one
"differs" case used a different segment *count* (2 vs 3), which the real bug still detected correctly
(missing/extra ordinal) — it just never exercised "same count, different content," the common real case.
This also contradicted `diff.go`'s own documented contract: "reports construct-level changes... not
storage units" — comparing by storage ordinal/slot-digest was never construct-level to begin with.

**Fix:** `diff_real.go` now decodes real CONTENT via `extract.LoadContentRecords`, keyed by the AUTHORED
unit-id — the same identity `merge`/`project`/`redact`/`publish`/`verify` already use post
GAP-VERIFY-CONTENT-REBUILD — and compares real frame bytes directly (SHA-256 previews in the change
description, not the comparison itself). The existing test's fixture builder (`writeDocWithContentSegments`,
raw non-decodable filler bytes, fine for `extract`'s byte-count assertions but never a valid content-model
frame) could not exercise this at all; replaced with `writeDocWithDecodableContent` building genuine
decodable frames, and the test now explicitly covers "same construct count, real content differs."
Verified with the actual compiled binary via a from-scratch end-to-end battery (37/37 real-file cases
passing across all 11 verbs) as well as the unit test.

## FINDING-2026-09-20 — validate never wires the storage_integrity_tree check (FR-104/FR-105), RESOLVED

**Severity:** potentially high — real gap, initially flagged rather than rushed, given its blast radius
(see below). **Found:** 2026-09-20, investigating why the DEFECT-2026-09-20b diff bug's root cause (a
CONTENT segment's `SegmentTableSlot.Digest` is never populated with a real content digest anywhere) was
never itself caught by `validate`. **Status:** RESOLVED 2026-09-20, T-0396 (handed to Kiro; see below).

`pkg/validate/storageintegrity.go` already implements a real, tested `CheckStorageIntegrityTree`
function: it recomputes T_S (the storage-integrity Merkle tree) fresh from the current SegmentTable and
compares it against the winning commit-ring record's `ledger_root`, per its own doc comment citing
FR-104/FR-105 and naming `storage_integrity_tree` as "cli.md's `validate` stdout schema['s] required check
key." **`pkg/cli/validate_real.go`'s `realValidateStepsFor` never calls it.** The real `validate` verb
today runs exactly 4 checks (header decode, capability arithmetic, ring-winner selection,
frontmatter+segment-table decode) and none of them recompute or compare any segment's real content
digest against its declared `slot-digest`. A document whose CONTENT segment bytes do not match their
declared digest — including, but not limited to, every zero-digest fixture DEFECT-2026-09-20b's writeup
describes — currently passes `validate` as `OK`.

**Why this is reported rather than fixed in the same pass as DEFECT-2026-09-20b:** wiring this check for
real would very likely flip a large, currently-unknown number of existing fixtures across `pkg/cli`'s own
test suite (and possibly other packages') from passing `validate` to failing it, since essentially none of
them populate a real per-segment digest today — this needs its own audit of the fixture-generation
convention codebase-wide, not a same-turn patch alongside two unrelated CLI defects. Recorded here plainly
per this project's honesty rule, rather than silently left un-mentioned or rushed.

**RESOLVED 2026-09-20 (T-0396).** `realValidateStepsFor` now adds a 5th step at `StepTSRecompute` (step 6)
that decodes the ring winner + segment table, builds a real `SegmentDigester` reading each live segment's
octets from the file and SHA-256'ing them (the same sha256-over-complete-octets convention `slot-digest`
uses), calls `CheckStorageIntegrityTree`, and maps a failing result to the `storage_integrity_tree`
finding — recomputing T_S fresh over the full decoded table and comparing it to the winner's `ledger_root`
(never trusting the stored root). The predicted fixture blast radius was real and was fixed properly, not
dodged: a single shared assembler (`assembleDoc`/`writeDoc` in `fixtures_test.go`) now seals every fixture
with real per-segment SHA-256 digests AND the correct `ledger_root = T_S(full table)`, and all 11
fixture-builders across `pkg/cli` were migrated to it. `realSignRun` (T-0392) was corrected to recompute
`ledger_root` over the new segment table after splicing its SIGNATURE segment, so a genuinely-signed
document still validates. `TestTR_012_ValidateDetectsRealDigestMismatch` proves a correct document passes
and a corrupted slot-digest is INVALID with the `storage_integrity_tree` finding. Both a critical T_S bug
(recompute over the full `MaxSegments` table, not the live-only slice — unused slots' zero digest is
distinct from an absent leaf) surfaced and were fixed. `go build/vet/test ./...` clean, the check never
skipped or weakened.

**Data-file follow-up (same day, `7aa79b6`):** the previously-untracked `protodoc-sample.pdl` (Eyvar's
PRONOM/IANA submission artifact, prefix-only, built before this check existed) failed `validate` once the
check went live, because its stored `ledger_root` was zero rather than a real `T_S`. Rather than leaving it
broken, Kiro added `cmd/protodoc-gensample` (a reproducible generator) and regenerated the file in a
follow-up commit, sealing a real `ledger_root`/digests while keeping the first 480 octets — including the
offset-0 DROID/PRONOM signature bytes the submission's identification test relies on — byte-identical to
the original (independently confirmed: only bytes at offset ≥523 changed). **Caveat, not a defect:** the
exact bytes already emailed to `PRONOM@nationalarchives.gov.uk` on 2026-09-19 (T-0371) predate this
regeneration, so the file now committed to the repo differs from that specific email attachment. The
signature-matching region is unaffected either way, but if PRONOM's review process ever cross-references
the public repo against the mailed attachment, the two are no longer byte-identical. No action taken on
this by Kiro or Claude; flagged for Eyvar's awareness only.

## Honesty note

T-0349, T-0356, and plan.md's Conflict 5 above name Eyvar García as decision-maker because those
decisions were actually made by him, on the dates given. Every remaining OPEN entry names no
decision-maker, because none of those decisions has actually been made yet. Recording them as OPEN is
the correct, honest state; the corresponding tests will remain red-by-design until the real
ruling/trial occurs, which is the release gate doing its job rather than a defect to paper over.
