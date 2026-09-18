# Protodoc Format Contracts

Status: APPROVED (Eyvar, 2026-09-07) | Spec ID: 001-protodoc-format-core | Phase: 3 (plan) | Date: 2026-09-07

This directory is the normative wire-level and interface contract for the Protodoc format, derived from
the architecture decision "Protodoc Ledger (PDL), Repair Set 2" (`../plan.md`) and its sibling artefacts
`../data-model.md` and `../research.md`. Every field in every file here carries an exact type and an
exact limit, per CON-009. Nothing here reopens a requirement in `../spec.md` or a resolved decision in
`../clarify.md`; where this pass found the architecture materials underspecified, self-contradictory, or
silent on a question their own preimage formulas depend on, the gap is named and a plan-level resolution
is proposed, flagged as such, never silently filled in and presented as settled.

## 1. Files in this directory

| File | Formalism | Scope |
|---|---|---|
| `container.abnf` | ABNF | The fixed 1,048,576-octet prefix (header, commit ring, frontmatter, segment table) and the append-only ledger's physical segment/frame/frame-directory framing. PDL-VARINT and the generic PDL-TLV field grammar are defined here and repeated byte-identically in the other two grammar files. |
| `document.abnf` | ABNF | The content-model record shapes carried inside CONTENT, RESOURCE and HISTORY segments: text and identity, anchors, annotations, tables, notes, cross-references, the extension envelope, embedded resources (rasters, font subsets, presentation artefacts, registry excerpts, the unit-index leaf) and history operations. |
| `integrity.abnf` | ABNF | Everything carried inside an ATTEST segment plus the two digest trees that reach into CONTENT/RESOURCE: T_S, T_C, structure_digest, signed_object, the coverage descriptor, the SIGNATURE record, the closed 7-step Ed25519 verification procedure, redaction and erasure, the RescindResign longevity record, and long-term-validation evidence. |
| `cli.md` | Prose + JSON schema fragments | The 11 TR-012 command-line verbs as an interface contract: arguments, exit codes, stdout shape, and which requirement each verb satisfies. |
| `README.md` | This file | Index, shared conventions, consolidated registries, corrections, originated fields, and known gaps. |

### 1.1 Why ABNF, and not protobuf, CBOR, or ASN.1

The task instruction for this contracts pass is explicit: write the formalism the final architecture
actually chose, and do not default to `.proto` merely because `protoc` 35.1 happens to be installed on
this machine. The architecture chose neither protocol buffers, deterministic CBOR nor ASN.1 DER for the
wire encoding: DP-002 (`../research.md` S3) rejects protocol buffers on 11 independent, documented
grounds (no canonical serialisation guarantee, non-minimal varints legal, packed/unpacked ambiguity,
unspecified field order, non-deterministic map order, float types, library-default ceilings, a
library-version-dependent unknown-field guarantee, nested length-prefixed containment that makes an edit
rewrite every ancestor's length, allocate-before-validate parsing, and a shared `.proto` schema defeating
CP-003's no-shared-source stability gate), and rejects deterministic CBOR and ASN.1 DER as runner-ups
with their own residual freedoms or structural costs. What the architecture chose instead is a **bespoke,
closed grammar**: raw fixed-layout structs with zero self-description for the physically-patched prefix
regions (Header, CommitRing, SegmentTable), and PDL-TLV, a from-scratch tag-length-value grammar with
every free parameter closed (1-octet per-record-type-scoped tags, PDL-VARINT adopted verbatim from
Bitcoin Core's CompactSize varint, a fixed byte-lexicographic sorted-vector comparator, exactly one
repeated-value form, no floats, no maps), for everything sealed once and never patched in place. ABNF
(RFC 5234) is the formalism this contracts pass uses to state that bespoke grammar precisely, per the
task's own instruction to use ABNF specifically when the chosen encoding is a bespoke grammar. `buf`
being absent from this machine is, per `../research.md` S7, moot for the identical reason: there is no
`.proto` schema anywhere in this format for either tool to operate on.

ABNF's own core rules (`CRLF`, `SP`, `HTAB`, and the rest of RFC 5234's text-protocol vocabulary) carry no
meaning here and are never used; the only core rule actually consumed is `OCTET` (`%x00-FF`), because
this is a binary grammar described in ABNF, not a text protocol. Every constraint ABNF's own notation
cannot express (exact integer endianness, cross-field bounds-safe arithmetic, an algorithm's step order,
a domain-tag registry, which numeric widths were chosen and why) is stated in a comment tagged
`NORMATIVE` immediately next to the field or rule it binds, per the task's own instruction; a
`NORMATIVE` comment is binding text, not explanatory prose, and CI (S7 below) is expected to treat it
that way.

## 2. Shared conventions across all three grammar files

Section 0 of `container.abnf`, `document.abnf` and `integrity.abnf` repeats the identical primitive
rules (`u8`, `u16`, `u32`, `u48`, `u64`, `digest256`, `unit-id`, `state-id`, `ext-token`, `zero16`, the
`varint` family, the generic PDL-TLV `field`/`field-tag`/`field-len`/`field-value`/`record` grammar, and
the `plain-seq-of-X`/`sorted-vec-of-X` sequence conventions) byte-for-byte, so each file parses standalone
without needing to `#include` another. `document.abnf` additionally defines `i64` (signed, for geometric
values that may be negative) and `nfc-string` (an NFC-validated UTF-8 byte string), since only that file
needs them.

**All fixed-width integers in this format are big-endian.** `container.abnf` states this explicitly for
`u8`/`u16`/`u32`/`u48`/`u64` and for PDL-VARINT's own multi-octet extension forms; this contracts pass
extends the SAME convention uniformly to every fixed-width field defined in `document.abnf` and
`integrity.abnf`, since the architecture materials state big-endian explicitly only for PDL-VARINT's
wider forms and leave the PREFIX's own raw fixed-width fields' byte order unstated. One byte order for
the whole format, chosen to match the one explicit data point given, is a smaller, safer surface than
leaving prefix byte order to be inferred per-implementation; getting this wrong would silently break
every cross-implementation digest comparison in the format (NFR-001/NFR-003), so it is called out here
rather than left implicit in three separate files.

**PDL-VARINT sorts correctly by its own encoded bytes.** Because marker octets 253/254/255 are
numerically greater than every literal single-octet value (0-252), and each wider form is fixed-width
big-endian, unsigned byte-lexicographic comparison of two complete PDL-VARINT encodings always agrees
with numeric comparison of the values they encode (stated once, in `integrity.abnf` S0, where the
`sorted-vec-of(segment-range)` construction in the coverage descriptor depends on it). This does not hold
for every variable-length integer scheme (LEB128 does not have this property), so it is proven rather
than assumed.

**Discriminant, domain-tag, and reserved-range non-overlap.** Every closed 1-octet enum in this format
(the record-type discriminant, the digest-tree domain tag, `slot-segment-type`, `cd-mode`, `ae-kind`,
`ae-format`, `sig-intent`, `ann-kind-value`, `anchor-boundary`) rejects an out-of-range value rather than
skipping it, for the identical reason stated once in `container.abnf` S5.1 and repeated at each site: the
value governs a property (tree treatment, coverage partitioning, merge disposition) that must stay closed
to writer extension, never opened via a minor capability generation. Section 3 below consolidates every
such registry across all three files into one place, so a future addition can be checked for collision
against the full set, not just the one file it is being added to.

## 3. Consolidated registries (cross-file view)

### 3.1 Record-type discriminant (1 octet, tag=0 in every record; container.abnf S6.2, document.abnf S1, integrity.abnf S1)

| Range | Owner | Status |
|---|---|---|
| `0x00` | -- | reserved, never valid (catches an all-zero/corrupt frame) |
| `0x01` - `0x3F` | `document.abnf` | CONTENT/RESOURCE/HISTORY record kinds; `0x01`-`0x0D` assigned (S1's table), `0x0E`-`0x3F` reserved |
| `0x40` - `0x7F` | `integrity.abnf` | ATTEST record kinds; `0x40`-`0x42` assigned, `0x43`-`0x7F` reserved |
| `0x80` - `0xFE` | -- | reserved for a future MAJOR version; a v1 reader rejects, never skips |
| `0xFF` | -- | reserved, never assigned (symmetric bookend to `0x00`) |

### 3.2 Digest-tree domain tag (1 octet, first octet of every hash preimage; integrity.abnf S2)

| Tag | Meaning | Given upstream or plan-assigned |
|---|---|---|
| `0x00` | `ABSENT_CHILD_DIGEST` filler | given (architecture text) |
| `0x01` | T_S internal node | given |
| `0x02` | T_C redactable leaf | given |
| `0x03`, `0x05`, `0x06` | reserved | -- |
| `0x04` | `signed_object` | given |
| `0x07` | T_C non-redactable leaf | given |
| `0x08` | T_C internal node | given |
| `0x09` | `severance_commitment` (ErasureRecord) | **plan-assigned this pass** (S5.1 below) |
| `0x0A` | `structure_digest` | given |
| `0x0B` - `0xFF` | reserved | -- |

### 3.3 `param_set_id` (u16; integrity.abnf S10.2), `slot-segment-type` (u8; container.abnf S5.1), and prefix-region bitmask bits (integrity.abnf S4)

Restated in full at their own definition sites; not repeated here to avoid a second, driftable copy of a
table that already carries CI-checkable values at its one authoritative location (S7.3 below names the
check that keeps this README's cross-references honest instead).

## 4. Corrections made to inherited artefacts

This contracts pass found and fixed one arithmetic defect already committed to disk before this pass
began, in `container.abnf`'s CommitRing section, and adopted (rather than re-litigated) one prior
resolution of a genuine layout tension between the architecture text and the fixed 1,048,576-octet total.

### 4.1 CommitRingRecord field-width arithmetic (fixed in `container.abnf` S3)

The version of `container.abnf` this pass inherited claimed `ring-fields` totals 316 octets and
`ring-reserved` is 164 octets. Summing the 14 named fields at their own individually-stated widths
(`ring-magic(4) + sequence(8) + ledger-length(8) + ledger-root(32) + segment-count(2) + state-id(32) +
frontmatter-digest(32) + segment-table-digest(32) + t-c-root(32) + parent-state-id(32) +
retention-point(2) + compaction-generation(4) + index-route(32) + structure-digest(32)`) gives 284, not
316; the two stated numbers were never consistent with each other, and 316 was never consistent with the
outer 512-octet slot total (`ring-fields + ring-reserved + record-digest = 512`) either way. This pass
corrects both totals to 284 and 196 respectively, and the reserved region's offset range from `[316,480)`
to `[284,480)`. No field's width and no field's presence changes; this is a pure arithmetic correction,
verified by direct summation (`python3 -c` recomputation, not re-eyeballed), not a design change.

### 4.2 The "IntegrityBlock+UnitIndexRoot" trailing-region tension (resolved layout in `container.abnf` S1, retained)

The architecture text (and the inherited `container.abnf`, and `../data-model.md` S2.4's own limits line)
state the SegmentTable occupies `[262144, 1048960)`. That range is 786,816 octets, 384 octets more than
`MAX_SEGMENTS(16384) x SLOT_WIDTH(48) = 786,432`, and its stated end (1,048,960) exceeds the format's own,
far-more-heavily-repeated total prefix size of 1,048,576 by exactly 384 octets, an impossibility (a
region cannot end past the file's own stated total). The architecture text separately describes an
`IntegrityBlock+UnitIndexRoot` region running from wherever SegmentTable ends "to 1,048,576," which,
given the arithmetic above, would need to occupy zero or negative space. `container.abnf`'s own S1
resolves this (inherited by this pass, not re-argued): the SegmentTable's independently-confirmed exact
size (`786,432`, corroborated by `MAX_SEGMENTS=16384`, `T_S`'s arity-16/depth-4 sizing, and every
`numeric_budget_compliance` row that cites it) already consumes the entire prefix remainder after
Header+CommitRing+Frontmatter (`512 + 3584 + 258048 + 786432 = 1,048,576` exactly), leaving no room for a
fourth physical region at all. "IntegrityBlock" and "UnitIndexRoot" are therefore carried as FIELDS of
the winning `CommitRingRecord` (`t-c-root` and `index-route`) rather than a disjoint byte range: still
determinable from the leading 1,048,576 octets exactly as TR-006/007/008 require, just co-located with
the ring instead of trailing the segment table. This pass's own `../data-model.md` S4 independently
reaches the same conclusion by a different route (T_S's root "cached in `CommitRingRecord.ledger_root`,"
UnitIndex "derived, non-normative... rebuilt (or verified) on open"), which this contracts pass treats as
corroborating rather than redundant. No requirement is reopened by this reading; it corrects an internal
arithmetic inconsistency in the architecture's own physical-shape prose using that same prose's own more
heavily corroborated numbers. Every "[262144, 1048576)" citation across all three grammar files is
consistent with this resolution.

### 4.3 A shared field name, disambiguated: `structure-digest` in the CommitRing versus in a Signature

`container.abnf` S3's `CommitRingRecord` carries its own field named `structure-digest`, and
`integrity.abnf` S3 defines `structure_digest` as an input to a SPECIFIC signature's `signed_object`,
computed from THAT signature's own `covered_prefix_regions_bitmask` and `covered_segment_ranges`. Two
different signatures on one document can legitimately have two different `structure_digest` values (a
`TOTAL`-mode one and a `SUBSET`-mode one, say), so "the" `structure_digest` is inherently
signature-relative, not file- or commit-relative. This pass makes explicit what the ring's own field
actually is, so the shared name is never mistaken for a circular definition: `CommitRingRecord.structure-
digest` is a cached, DEFAULT-COVERAGE computation of the identical S3.1 algorithm (every bitmask bit set,
every currently-populated non-ATTEST ordinal covered) made at commit time, never itself trusted as
authoritative, exactly like `ledger-root` and `t-c-root` in the same record. It equals a `TOTAL`-mode
signature's own `structure_digest` by construction, and is unrelated to a `SUBSET`-mode one. Stated once
at each field's own definition site (`container.abnf` S3, `integrity.abnf` S3.1 item 3) rather than only
here.

## 5. Fields, values and design choices this phase originated

Per the task's own framing, a contracts phase is where exact wire values get pinned down, not only where
already-decided ones get transcribed. The architecture materials name concepts (a discriminant exists, a
severance domain tag exists, a signing-intent enum exists) far more often than they assign the concept's
actual numeric wire value; this pass had to assign one wherever a real, working grammar needed a concrete
number and none was given. Every item below is either (a) an arbitrary-but-consistent labelling choice
with no risk of being "wrong" beyond internal self-consistency, which this pass assigns with ordinary
confidence, or (b) flagged as needing an authoritative external source before implementation, when the
risk of a wrong value is real (S5.2).

### 5.1 Low-risk plan-level assignments (arbitrary-but-consistent; any distinct, closed set would serve equally)

- File magic (`container.abnf` S2): `0x50 0x44 0x4C 0x31 0x00 0x00 0x00 0x00` (ASCII "PDL1" + 4 reserved octets), inherited from the pre-existing `container.abnf`, retained.
- Ring-slot magic (`0x50 0x44 0x52 0x31`, "PDR1"), segment header magic (`0x50 0x44 0x53 0x31`, "PDS1"): inherited, retained.
- The record-type discriminant registry's full numeric assignment (S3.1 above): every value 0x01-0x0D and 0x40-0x42 is a fresh assignment; only the RANGE SPLIT (document.abnf owns 0x01-0x3F, integrity.abnf owns 0x40-0x7F) was implied, not chosen, by the inherited container.abnf's own cross-reference to "document.abnf S1 and integrity.abnf S1."
- Domain tag `0x09` for `severance_commitment` (integrity.abnf S2): the architecture text names "a severance-domain-tag" conceptually, never a number; this pass assigns `0x09`, the first free slot between the T_C internal-node tag (`0x08`, given) and `structure_digest`'s tag (`0x0A`, given).
- `cd-mode` (`0x00 TOTAL`/`0x01 SUBSET`), `ae-kind` (`0x00`/`0x01`/`0x02`), `ae-format` (`0x00`-`0x03`), `sig-intent` (4 values, integrity.abnf S10.3), `ann-kind-value` (4 values, document.abnf S10.1), `note-placement` (2 values), `xref-kind` (3 values), `anchor-side` (2 values): all closed enums named by data-model.md or the architecture text at the concept level only; this pass assigns the numeric values and reserves the remainder of each 1-octet space explicitly.
- `param_set_id = 1` for the v1 {Ed25519-EdDSA-Protodoc-1, SHA-256} allowlist entry: the architecture names the pair; this pass assigns it a wire id and reserves `0x0002`-`0xFFFF` for a future major-version entry reachable only via RESCIND-AND-RESIGN.
- The 64-bit-vs-16-bit width choice for segment ordinals, `segment-count` and `retention-point` (`u16`, matching `container.abnf`'s inherited choice; `../data-model.md` types these `uint32`, a now-stale width in that sibling artefact this pass does not have authority to edit, flagged here rather than silently reconciled).
- `ExtensionEnvelope`'s `position_key` and `signature`'s `sig-presentation-ref` (document.abnf S6, integrity.abnf S5): the anchor-point structure is reused for the former (a position-claim needs some totally-orderable shape, and reusing the one already defined for Annotation/Note avoids inventing a third); the latter is a genuinely new field, since `signed_object`'s own formula (S3.2) requires a `presentation_artefact_digest` input, but no field anywhere in the given materials states which `PresentationArtefact` a `Signature` names to derive that digest from.
- The `Table`, `Note` and `CrossReference` record shapes in their entirety (document.abnf S4, S5): the architecture and both sibling plan artefacts name these constructs only at the merge-classification or task-description level ("table row/table-column insertion is a position-claim"; "tables, notes, references" as this file's scope per the task instruction); no record shape existed anywhere upstream. The shapes given are the minimum needed to make FR-108's reference resolution and the R1 classification concrete, built only from primitives already established elsewhere.
- The `ATTESTATION_EVIDENCE` record and its unification of credential chains, revocation evidence and time attestations under one record kind with a kind selector (integrity.abnf S9): `../data-model.md`'s `Signature` entity types these as opaque `[]byte` refs with no wire shape; this pass gives them one, over the external, open IETF standards FR-070/071 already imply (X.509 v3 chains per RFC 5280, OCSP per RFC 6960 or CRLs per RFC 5280, RFC 3161 TimeStampTokens), reasoning that referencing an open standard by RFC number is interoperability (ISO/IEC 25010), not the CON-006/CP-009 "named application" violation CQ-006's shaping-oracle exception is separately, explicitly weighed against.

### 5.2 Flagged, not fabricated: the EdDSA-Protodoc-1 small-order point table (integrity.abnf S6)

The 7-step Ed25519 verification procedure's Steps 2 and 4 require comparing a decoded point against a
fixed table of the curve's 8 small-order points. This contracts pass states the exact mathematical
characterisation (points whose order divides 8 on edwards25519, the full torsion subgroup, exactly 8 of
them) and the exact non-cryptographic checks around them (the canonical-encoding comparison against
`p = 2^255 - 19`, the `S >= L` check with `L` given exactly), but does **not** reproduce the 8 points' own
32-octet compressed hex encodings from memory. A transcription error in a normative security document is
worse than a named-but-not-inlined table: too permissive a wrong value silently reopens the small-order
acceptance vulnerability the whole procedure exists to close, and too strict a wrong value rejects valid
signatures. Implementers MUST populate this table from RFC 8032 section 5.1.3 together with Chalkias,
Garillot and Nikolaenko, "Taming the many EdDSAs" (2020) Table 1, the same paper `../research.md` S5
already cites as the source of the documented cross-implementation divergence this procedure closes.
`p` and `L` themselves ARE stated as exact literals, both being simple, universally-quoted integers with
no comparable transcription risk and both already given verbatim in the architecture text this pass was
handed.

### 5.3 PLP-1's own bitstream syntax is out of this contract's scope

`document.abnf` S7.1 fixes the RASTER_IMAGE frame (a codec selector, dimensions, an opaque bitstream
field); it does not enumerate PLP-1's 8 fixed quantization matrices or its fixed canonical-Huffman code
tables, because no numeric coefficients or code tables for either were given anywhere in the materials
available to this pass, only the concepts ("8 fixed quantization matrices," "fixed canonical-Huffman
entropy tables"). Fabricating roughly 512 quantization coefficients and a Huffman table set here, with no
cited derivation and no way to check them against CP-003's two-implementation gate before a second
implementation exists to compare against, would itself be an unreviewed act of invention disproportionate
to a wire-framing contract. A dedicated PLP-1 codec specification, with exhaustive decode conformance
vectors, is recommended as a first-implementation activity before phase 6, per `../research.md` S8 item
8's own recommendation.

## 6. Known gaps carried forward, not hidden

- **T_C's traversal order has no defined "document logical structure" to be fixed by** (`integrity.abnf`
  S2.2.1). Research.md requires T_C's subtree ordinal order be "fixed by the document's logical
  structure... never storage ordinal and never by digest value," but no document-root or top-level
  reading-order record exists anywhere upstream. This pass adopts ascending unit-id sort as an interim,
  fully deterministic rule (satisfying the negative constraints exactly) while explicitly not satisfying
  the positive "reflects reading order" aspiration, and recommends a `ROOT_SEQUENCE`-shaped addition to
  `document.abnf` before phase 6.
- **Cross-reference edges are not one of FR-109's 5 named cycle-detection edge kinds** (`document.abnf`
  S5, `../data-model.md` S7 step 8's own list: structural moves, extension-envelope fallback references,
  annotation anchor references, run split/merge lineage, RescindResignRecord chains). FR-109's own literal
  text names "cross-references" as one of its 4 abstract graph categories, but the already-committed
  concrete 5-item list in `data-model.md` does not include a `CROSS_REFERENCE` edge. This pass does not
  unilaterally expand that already-committed list to 6 on its own authority; it records the gap instead.
  FR-108's own per-reference resolution check still applies regardless of cycle participation.
- **FR-061 versus FR-075** (`integrity.abnf` S7.2): `ErasureRecord`'s wire form uses a salted commitment,
  not the bare unsalted digest FR-061's literal text specifies, for the reason `../research.md` S5 and
  S8 item 1 already give (a bare digest reopens FR-075's own privacy hole for trimmed history). Carried
  forward unchanged from the inherited materials; not re-argued here, only re-stated at the field it binds.
- **CON-006 versus CQ-006 option B** (`document.abnf` S7.4.1's `fr-features` comment): the pinned,
  versioned external shaping algorithm CQ-006 already approved sits close to CON-006's boundary against
  product-behaviour references; `../research.md` S8 item 2 records the required governance exception
  (EX-001), which remains outside this contracts pass's authority to grant.
- **NFR-030's literal 4x-input-memory bound and FR-055's read-count-versus-octet-ceiling reading**: both
  readings this pass's sibling artefacts adopt (`../data-model.md` S7's closing notes) are assumed
  unchanged by `cli.md`'s `OVER_BUDGET`/read-cost design; neither is re-litigated here.

## 7. How this is validated in CI

ABNF describes a binary grammar for human and cross-implementation review; it is not itself compiled or
executed by either reference implementation, so validating conformance to it needs a companion, executable
harness rather than an ABNF parser generator. The following is the intended CI shape for phase 6 onward:

1. **Constant-table generation and equality.** Every numeric ceiling, enum value, discriminant, and domain
   tag named in these four files is transcribed once into checked-in Go constants (mirroring CON-009/010/
   011's `LIM` table convention already established for `../data-model.md` S5). A generated-from-source-text
   check (a small script diffing the `.abnf`/`.md` files' own stated numbers against the Go constants) fails
   the build if the two ever drift, closing the exact class of defect S4.1 above found and fixed by hand.
2. **Golden conformance corpus, one file per record type and per prefix region.** Each record type defined
   in `document.abnf` and `integrity.abnf`, and each prefix region defined in `container.abnf`, gets an
   at-limit and an over-limit fixture (CON-010), hand- or property-test-generated to match this grammar
   exactly; both reference implementations (CP-003, CQ-012) must produce identical accept/reject verdicts
   and, where applicable, identical canonical octets on every fixture. The **migration** golden corpus
   (`pkg/migrate/testdata/migration/`, FR-119/CP-003/NFR-028) is the migrate-specific instance of this
   set: `(source, target-major)` pairs with expected canonical migrated octets, reusable by an
   independent implementation; it feeds M19's two-implementation gate but does not itself run it.
3. **Negative corpus per NORMATIVE reject rule.** Every `rule_id` named in these files (`PD-VARINT-001`,
   `PD-TLV-001`, `PD-SORT-001`, `PD-RING-001`, `PD-DISC-001`, `PD-COVER-001` through `004`, `PD-EXT-002`,
   `PD-BOUNDARY-001`, `PD-NFC-001`/`002`, `PD-INDEX-001`, `PD-SEGTYPE-001`, `PD-CAPPAIR-001`,
   `PD-DURABLE-001`, `PD-MODE-001`, `PD-HDRZERO-001`, `PD-LANG-001`,
   `PD-BIDI-001`, `PD-A11Y-001` through `005`, `PD-TBL-001a`/`001b`, `PD-2D-001`, `PD-INFER-001` (the M15
   accessibility/semantic rules, T-0270..T-0293 — added here so the 16-of-18 spec-named rule ids that were
   previously never carried downstream now each have a registered conformance case, per CP-011/NFR-029))
   gets at least one hostile fixture proving the rule
   fires, contributing to CP-011's 200-hostile-case negative corpus floor and CP-002's per-role statement
   budget (each mapped `rule_id` is one countable normative statement). This list is generated, not
   hand-maintained: CI greps every `.abnf`/`.md` file in this directory for the `PD-[A-Z]+-[0-9]+` pattern
   and fails if any match here has zero corpus fixtures, or if a file introduces a new one this list (and
   therefore this corpus) has not yet grown to cover.
4. **Field-tag ascending-order and cross-file discriminant/domain-tag non-overlap, checked mechanically.**
   A small static check parses each `.abnf` file's own `field-tag`/`tag=` comments and asserts strict
   ascending order per record type (the property this pass hand-verified for all 16 record types across
   `document.abnf` and `integrity.abnf` while writing them) and asserts the S3.1/S3.2 registries above have
   no two files claiming the same numeric value, so a future addition to either file is caught at review
   time rather than at a two-implementation divergence months later.
5. **The EdDSA-Protodoc-1 constant table (S5.2).** CI asserts the small-order point table's own SHA-256
   digest against a value computed once from the cited RFC 8032 / "Taming the many EdDSAs" source and
   pinned in the conformance corpus, so an accidental future edit is caught without requiring anyone to
   re-derive or re-eyeball 8 curve points by hand.
6. **cli.md's exit-code and stdout-shape contract**, exercised directly: release gate `G-CLI` (TR-012's own
   verify clause) requires each of the 11 verbs be exercised by at least one conformance case per the
   traceability table in `cli.md` S13, asserting the documented exit code, `status` name and JSON payload
   shape for at least one success and one failure case per verb.

## Extension-token registry governance (CON-021)

The extension-token space's owner-id partition (registered / owner-scoped / permanently-retired) is
defined in `container.abnf` S5.2. The **governance** of the registered tier -- the published maximum
review turnaround (15 business days) and the per-request/per-release mechanism that tracks observed
turnaround against it (metric `G-REG`) -- is a process obligation, not a wire mechanism, and lives in
[`docs/token-registry-governance.md`](../../../docs/token-registry-governance.md) (CON-021). Owner-scoped
tokens are self-issued and out of scope for the review SLA; the reserved and permanently-retired
namespaces are never issued.
