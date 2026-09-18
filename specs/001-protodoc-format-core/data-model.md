# Protodoc: Data Model

Status: APPROVED — Eyvar, 2026-09-07 | Spec ID: 001-protodoc-format-core | Phase: 3 (plan) | Date: 2026-09-05

This document is the data-model artefact for plan phase 3, derived from the architecture decision "Protodoc Ledger, Repair Set 2" (`plan.md`). It states every persisted entity, its exact field types, its constraints, and the indexes and derived structures a reader builds over it. Every field is traceable to a requirement, a clarification decision (CQ-xxx), or a constitution principle (CP-xxx). No requirement is reopened here; where a requirement's literal text conflicts with the design, the conflict is named in `plan.md` §spec_conflicts and referenced, not silently resolved.

## 1. Layer map

Every entity belongs to exactly one of six layers. An entity never spans layers; a field that references another layer's entity does so by ordinal, digest, or opaque token only, never by embedding.

| Layer | Responsibility | Entities |
|---|---|---|
| Container (prefix) | Content-independent identity, bounded-prefix determination, crash-atomic commit | Header, CommitRingRecord, Frontmatter, SegmentTableSlot |
| Storage (ledger) | Append-only sealed segment bytes and their internal frame structure | Segment, PDL-VARINT (encoding primitive) |
| Identity | Character-granular durable identity, run compression, anchors | TextBlock, Run, Annotation/Range, Table, Note, CrossReference |
| Integrity | Structural digests, content commitment, redaction, signatures, longevity | T_C node, T_S node, CoverageDescriptor, Signature, RedactionCommitment, PresentationArtefact, FontRecord, RescindResignRecord, RegistryExcerpt |
| History | Operation retention, erasure | HistorySegment, ErasureRecord |
| Extensibility | Forward-compatible unknown-construct carriage with no reliance on encoding-level skip behaviour | ExtensionEnvelope |
| Derived (rebuilt on open, non-normative) | Reader-side acceleration structures not part of any signature | PageDirectory entry, UnitIndex (root + leaf) |

## 2. Entities

### 2.1 Header

Purpose: content-independent format identity and version/capability gate, readable from the first 512 octets with zero decode (TR-006, TR-007, TR-008).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| magic | `[8]byte` | yes | fixed constant | file-type sniff |
| format_major | `uint16` | yes | — | frozen at file creation |
| format_minor | `uint16` | yes | — | monotonic within major |
| document_class | `uint16` | yes | closed enum | CON-005 |
| capability_written | `uint16` | yes | — | bitset of generations used |
| capability_required | `uint16` | yes | `capability_required <= capability_written` else reject naming both (FR-010) | |
| durable_claim | `uint8` | yes | `0` or `1` | gates mandatory RegistryExcerpt (FR-011) |
| history_mode | `uint8` | yes | one of `{NO_HISTORY, RETAINED_FROM_POINT, COMPLETE_HISTORY}` | immutable at creation (CON-022) |
| unicode_version_id | `uint16` | yes | registry-defined | CQ-006 |
| shaping_profile_id | `uint16` | yes | registry-defined | CQ-006 |
| prefix_layout_id | `uint64` | yes | — | selects prefix region sizes for a future major version |
| reserved | `[…]byte` | yes | MBZ (must-be-zero) | to octet 480 |
| header_digest | `[32]byte` (SHA-256) | yes | `SHA-256(header_bytes[0,480))` | |

Invariants:
1. Octets `[0,32)` are frozen forever from format_major=1 onward; no future version reinterprets them (CP-008, DP-012).
2. Zero document-text, user-metadata, or content-derived values appear anywhere in Header.
3. `capability_required > capability_written` is a structural reject naming both values (FR-010).
4. A `format_major` greater than the reader's supported major is a decline naming the version, zero dispositions applied (FR-123).
5. `header_digest` participates in `structure_digest`'s preimage only when bit `HEADER` is set in a signature's `covered_prefix_regions_bitmask`.

Identity: singleton at file offset 0; no instance identity.

Limits: exactly 512 octets, fixed (`prefix_layout_id`-selected alternate sizes are a future-major-version construct, not a v1 variable).

### 2.2 CommitRingRecord

Purpose: crash-atomic, self-digesting commit slot; the sole source of "current state" (HC-027, FR-117).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| magic | `[4]byte` | yes | fixed | |
| sequence | `uint64` | yes | monotonic per file | |
| ledger_length | `uint64` | yes | `<= file length` | anti-truncation/rollback |
| ledger_root | `[32]byte` | yes | T_S root at commit time | |
| segment_count | `uint16` | yes | `<= 16384` | corrected from `uint32`; container.abnf S3's ring-field arithmetic (14 fields summing to exactly 284 octets) only balances at `u16` |
| state_id | `[32]byte` | yes | CSPRNG or content-derived per DP-003/DP-015 usage; container.abnf `state-id-field`, equals the T_C root per FR-003 | also the R2 tiebreak key |
| frontmatter_digest | `[32]byte` | yes | | |
| segment_table_digest | `[32]byte` | yes | | |
| integrity_block_digest | `[32]byte` | yes | container.abnf `t-c-root`; the T_C root as carried in the ring slot for signature binding (DP-013), a distinct wire position from `state_id` even though both equal the T_C root value | |
| parent_state_id | `[32]byte` | yes | `0` for the root state | |
| retention_point | `uint16` | yes | ordinal into HISTORY segments | corrected from `uint32`, same arithmetic reason; CQ-008 |
| compaction_generation | `uint32` | yes | monotonic, incremented by full or partial compaction | |
| index_route | `[16]uint16` | yes | container.abnf `index-route`, 32 octets | 16-bucket routing table into `SegmentTableSlot` ordinals for `UnitIndex`; missing from an earlier draft of this table, already implemented by T-0042 |
| structure_digest | `[32]byte` | yes | see §Integrity layer, `plan.md` DP-013 | |
| record_digest | `[32]byte` | yes | `SHA-256(record_bytes[0,480))` | |

Invariants:
1. Slot count is exactly 7, occupying file offsets `[512,4096)`.
2. Opening the file selects the slot with the highest `sequence` whose `record_digest` verifies AND whose `ledger_length <= actual file length`.
3. Two slots sharing the highest valid `sequence` is a structural reject naming both slot indices (PD-RING-001).
4. The 6 non-winning slots are permanently, structurally excluded from every signature's `covered_segment_ranges` and from `covered_prefix_regions_bitmask` — they are never `RING_WINNER`.
5. `ledger_length` is included directly in `structure_digest`'s preimage independent of which ring record is later inspected, so truncation to an earlier valid boundary is caught two ways (FR-117).

Identity: `(ring slot index 0..6, sequence)`.

Limits: 7 slots x 512 octets = 3,584 octets exactly, occupying `[512,4096)`.

### 2.3 Frontmatter

Purpose: bounded preview raster and document metadata readable without touching the ledger (FR-051, FR-053, FR-054, HC-009).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| preview_raster | `[]byte` | no | `<= 131072` octets, PLP-1 or restricted-PNG | non-normative (HC-009) |
| preview_source_snapshot | `[]byte` | no | `<= 4096` octets | digest of render input at preview time, for staleness detection |
| document_metadata.title | UTF-8 `string` | no | NFC (CQ-009) | participates in `structure_digest` |
| document_metadata.direction | `uint8` | mandatory | closed value set `{0 = LTR, 1 = RTL}` | top-level document base writing direction (FR-032); provisional per T-0267 ruling (clarify-002.md, OPEN) |
| document_metadata.page_count | `uint32` | no | | participates |
| document_metadata.page_dimensions | `[2]int64` (914400-per-inch base units) | no | integer only, no float (CON-012) | participates |
| document_metadata.language | BCP-47 `string` | no | NFC | participates |
| document_metadata.colour_profile_id | `uint16` | no | CON-014 | participates |
| retired_token_list | `[]uint16` | no | `<= 16384` octets encoded | permanently-retired extension tokens (DP-011) |

Invariants:
1. Region occupies `[4096,262144)`; total content `<= 258,048` octets after the fixed region header.
2. Only `document_metadata`'s named fields participate in `structure_digest`'s `frontmatter_metadata_bytes`; `preview_raster` and `preview_source_snapshot` never do (HC-009: preview is advisory, not authoritative).
3. `title` and `language` are NFC-normalized at write time; a non-NFC input is rejected, never silently renormalized (CON-003).

Identity: singleton region; no instance identity.

Limits: 258,048 octets total region; `preview_raster <= 131072`; `preview_source_snapshot <= 4096`; `document_metadata` fields `<= 16384` octets encoded; `retired_token_list <= 16384` octets.

### 2.4 SegmentTableSlot

Purpose: the sole inventory of ledger segments (TR-006, no second digest-keyed index per FR-102-105).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| segment_type | `uint8` | yes | one of `{CONTENT, RESOURCE, HISTORY, ATTEST}` | closed enum |
| flags | `uint8` | yes | bit0 = coverage-hint only, never authoritative | authoritative coverage is `CoverageDescriptor` |
| offset | `uint48` | yes | absolute file offset | `>= 1048576` |
| length | `uint48` | yes | `> 0` | |
| frame_count | `uint16` | yes | `<= 8192` | |
| segment_digest | `[32]byte` | yes | `SHA-256` of the sealed segment's frame bytes | recomputed fresh at verify time, never trusted from this slot alone |

Invariants:
1. Slot ordinal is a monotonic storage ordinal assigned at append time, independent of name, digest, or content (DP-005).
2. An unused slot is all-zero — a normative constant, not uninitialised padding.
3. `frame_count*48 + 32 <= length - 64` is checked with bounds-safe (overflow-checked) arithmetic before the frame-directory offset is computed.
4. A PARTIAL COMPACTION may reassign an **uncovered** ordinal's `(offset, length, segment_digest)` to a new physical location; a **covered** ordinal's triple is never touched by any operation once a signature names it covered (DP-016).
5. `segment_type = ATTEST` marks a segment categorically excluded from every signature's `covered_segment_ranges`, including its own.

Identity: slot ordinal (0..16383), the sole addressing key for physical storage.

Limits: `MAX_SEGMENTS = 16384` slots x 48 octets = 786,432 octets, occupying `[262144,1048576)` (corrected from `1048960`, a typo: `262144 + 786432 = 1048576`, matching container.abnf's reconciled prefix total).

### 2.5 Segment

Purpose: an immutable, single-typed unit of ledger content (CQ-007).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| header | `[64]byte` | yes | fixed layout: type echo, frame_count echo, reserved | |
| frames | PDL-TLV encoded | yes | schema-specific per `segment_type` | see §PDL-VARINT |
| frame_directory | `[]DirEntry` | yes | tail-placed, `frame_count x 48 + 32` octets | 32 = directory digest |

Invariants:
1. Single-typed: exactly one of `CONTENT`, `RESOURCE`, `HISTORY`, `ATTEST` (CQ-007).
2. Immutable once sealed; never rewritten in place. An uncovered segment may be relocated wholesale (identical bytes, new offset) by partial compaction; it is never edited in place.
3. `ATTEST`-typed segments are categorically excluded from every `CoverageDescriptor`, including their own and every co-existing signature's (closes the FR-002 self-coverage circularity by construction).
4. Content is addressed by the digest of its own octets (`segment_digest`), paired with a canonical storage order (storage ordinal) that never depends on that digest (CQ-007).

Identity: storage ordinal, equal to its `SegmentTableSlot` index.

Limits: `MAX_FRAMES_PER_SEGMENT = 8192`; `MAX_DECODED_UNIT = 268435456` octets; `MAX_TEXT_UNIT_OCTETS = 65536` for a CONTENT text-block frame.

### 2.6 PDL-VARINT (encoding primitive, not a stored entity)

Purpose: minimal-length, self-terminating integer encoding for every length and non-fixed-width integer inside PDL-TLV frame content (DP-002, Bitcoin CompactSize form).

Invariants:
1. `0 <= n <= 252`: 1 octet, value = n.
2. `253 <= n <= 65535`: marker octet `253` + 2 big-endian octets; reject if `n <= 252` (non-minimal).
3. `65536 <= n <= 2^32-1`: marker octet `254` + 4 big-endian octets; reject if `n <= 65535`.
4. `2^32 <= n <= 2^64-1`: marker octet `255` + 8 big-endian octets; reject if `n <= 2^32-1`.
5. Never applied to the fixed-layout prefix (Header, CommitRingRecord, SegmentTableSlot), which uses raw fixed-width fields exclusively — this is a hard layer boundary, not a style preference (zero-parse requirement, HC-006/HC-007).

Limits: max representable value `2^64 - 1`.

### 2.7 TextBlock (addressable text segment)

Purpose: NFC-normalized, run-composed unit of identity-bearing text (CQ-001, CQ-002, CQ-009).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| block_id | `[16]byte` | yes | unique within document | |
| direction | `uint8` | mandatory | closed value set `{0 = LTR, 1 = RTL}` | base writing direction (FR-032); provisional per T-0267 ruling (clarify-002.md, OPEN) |
| language_ref | `uint16` | no | | |
| runs | `[]Run` | yes | `>= 1` | ordered by base_ordinal within the block |

Invariants:
1. Text is NFC per block; no renormalisation across block boundaries (CQ-009).
2. PD-NFC-001: a run's stored text always starts at an NFC-legal boundary (no combining mark opens a run except where the previous run's tail already isolates it).
3. RED-ALIGN: every redactable subtree boundary lands on a block boundary — a redaction can never split a TextBlock.
4. A comment or annotation's own text body is itself an ordinary TextBlock, so concurrent edits to it resolve via R1 SEQUENCE-ORDER exactly like document body text.

Identity: `block_id`; characters are addressed by `(run_id, base_ordinal+i)`, never by block-relative count (CQ-001).

Limits: `<= 65536` octets per block; overflow is handled by an explicit continuation link to a successor block, never silent truncation.

### 2.8 Run

Purpose: the unit of persisted character identity, compressing to run granularity (CQ-001, CQ-002, HC-010, HC-011).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| run_id | `[16]byte` | yes | 128-bit CSPRNG, minted once per contiguous typing burst | permitted non-deterministic site #1 (CQ-004) |
| base_ordinal | `uint32` | yes | reader-computed at read time is also legal; persisted value is authoritative | IDENTITY-COMPONENT (A-FIELD-ROLE) |
| char_count | `uint32` | yes | `> 0` | |
| props_ref | `uint8` | no | | |
| lang_ref | `uint8` | no | | |
| text | UTF-8 `string` | yes | NFC | |

Invariants:
1. A character's identity is the pair `(run_id, base_ordinal + offset)`, never a document-relative counted position (CQ-001, CQ-002).
2. Split preserves `run_id`, shifts only `base_ordinal`; merge is a pure syntactic predicate over adjacent runs with matching `run_id` lineage.
3. `run_id` is minted once per contiguous typing burst, never derived from actor, device, clock, or session (CQ-004).
4. A-FIELD-ROLE tags both `run_id` and `base_ordinal` as `IDENTITY-COMPONENT`: comparable and look-up-able, but never usable for positional arithmetic from a sequence start. This is the mechanical closure of the run-internal-index-vs-counted-position ambiguity.

Identity: `run_id`.

Limits: collision bound `P < 2^-60` below `2^34` mints per lineage (approximately `1.5 x 10^-23` at the `10^8` mints / 1,000-replica benchmark scale, FR-023).

### 2.9 Annotation / Range

Purpose: durable, anchor-based reference into text that survives concurrent edits without a counted position (CQ-001, CQ-002, FR-025 through FR-030).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| start.run_id | `[16]byte` | yes | | |
| start.birth_ordinal | `uint32` | yes | IDENTITY-COMPONENT | |
| start.side_bit | `bool` (1 bit) | yes | | |
| start.boundary_behaviour | `uint8` | yes | one of exactly 4: `{inside, outside, inside-if-inserted-before, inside-if-inserted-after}` | |
| end.* | same shape as start.* | yes | | |
| orphan.author_ref | `uint16` | no | populated only at orphaning time | |
| orphan.quoted_text | UTF-8 `string` | no | NFC | |
| orphan.prev_surviving_id | `[16]byte` | no | | |
| orphan.next_surviving_id | `[16]byte` | no | | |

Invariants:
1. Zero persisted counted positions anywhere in this entity (CQ-001, CQ-002).
2. `boundary_behaviour` is one of exactly 4 closed values, declared independently per endpoint, preserved unmodified across save/load/merge (FR-026, FR-027).
3. Orphan carriage fields are populated at the moment the anchored range is deleted, not recomputed at read time.
4. Concurrent creation of two distinct annotations is DISJOINT-COMMUTE by default. Concurrent edits to one annotation's non-text fields resolve via R2 TOTAL-ORDER-TIEBREAK; edits to its text body (a TextBlock) resolve via R1 SEQUENCE-ORDER; a concurrent delete of the anchored range resolves via R3 DELETE-DOMINATES with disposition = orphan-carry (DP-004, DP-015).

Identity: the annotation's own opaque unit identity (not derived from its anchors).

Limits: `MAX_REFERENCES = 4194304`.

### 2.10 T_C node (content commitment / redaction tree)

Purpose: the primary signed structural commitment, redaction-compatible via per-subtree salting (CQ-005, CQ-007).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| leaf.tag | `uint8` | yes | `0x02` (redactable) or `0x07` (non-redactable) | domain separation |
| leaf.salt | `[32]byte` | conditional | present iff `tag=0x02`; CSPRNG | permitted non-deterministic site #2 |
| leaf.canon | `[]byte` | yes | canonical encoding of the committed subtree | |
| internal.tag | `uint8` | yes | `0x08` | |
| internal.children | `[16][32]byte` | yes | always exactly 16 slots | absent children filled with `ABSENT_CHILD_DIGEST` |

Invariants:
1. `ABSENT_CHILD_DIGEST = SHA-256(0x00)`, a reserved domain tag no real digest preimage ever starts with (real preimages start `0x01`/`0x02`/`0x07`/`0x08`).
2. Every internal node hashes `tag || 16 x 32-octet-slot` = exactly 513 octets, regardless of actual child count — closes the collision-filler gap.
3. `T_C_root` is the node at subtree ordinal 0 in canonical depth-first order over content subtrees; this order is never the storage ordinal.
4. Salt lives inside the subtree it commits and is destroyed with it on redaction; the retained value after redaction is the 32-octet commitment alone.
5. `T_C_root` is signed directly inside `signed_object`; it is not routed through `T_S`.

Identity: subtree ordinal, canonical depth-first order (never storage ordinal).

Limits: depth bounded by `MAX_CONTENT_UNITS = 1048576` at arity 16 (depth 5).

### 2.11 T_S node (storage integrity tree)

Purpose: general, signature-independent self-consistency check over the physical segment table (used by unsigned documents and bounded-prefix consumers).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| leaf | `[32]byte` | yes | `= SegmentTableSlot.segment_digest` | |
| internal.tag | `uint8` | yes | `0x01` | |
| internal.children | `[16][32]byte` | yes | always exactly 16, absent = `ABSENT_CHILD_DIGEST` | |

Invariants:
1. Tree position equals segment ordinal (storage order), unlike T_C.
2. `T_S`'s root is never embedded in `signed_object`; it is free to change under partial compaction without invalidating any signature.

Identity: tree position = segment ordinal.

Limits: arity 16, depth 4 over 16,384 fixed slots (`16^4 = 65536` capacity, headroom over `MAX_SEGMENTS`).

### 2.12 CoverageDescriptor

Purpose: the exact, mechanically re-verifiable set of prefix regions and segment ordinals a given signature covers (closes FR-002 self-coverage circularity and the forged-inventory finding).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| mode | `uint8` | yes | `TOTAL` or `SUBSET` | |
| covered_segment_ranges | `[]PDL-VARINT pair` | yes | sorted, half-open `(start_ordinal, end_ordinal)`, merge-adjacent canonicalised | exactly one valid encoding per coverage set |
| uncovered_segment_ranges | `[]PDL-VARINT pair` | yes | same shape and canonicalisation rule | |
| covered_prefix_regions_bitmask | `uint8` (1 octet) | yes | bits for `{HEADER, RING_WINNER, FRONTMATTER, SEGMENT_TABLE, INTEGRITY_BLOCK}` | |

Invariants:
1. A range naming any `ATTEST`-typed segment ordinal, in either `covered_` or `uncovered_segment_ranges`, is rejected before verification proceeds — ATTEST segments are never nameable at all.
2. A range appearing in both `covered_` and `uncovered_segment_ranges`, or a gap in `[0, segment_count)` covered by neither, is rejected before verification proceeds.
3. A verifier recomputes `structure_digest` fresh from the current file's actual octets restricted to the claimed covered set and compares it to the value inside `signed_object`; a stored coverage claim is never trusted on its own.
4. Non-minimal or unmerged-adjacent range encodings are rejected, not accepted as an alternate valid form (closing the interop-divergence finding on this wire format).

Identity: embedded in exactly one `Signature`; not independently addressable.

Limits: bounded by `MAX_SEGMENTS = 16384`; typical edit delta is O(10) range entries, worst case a few KB, within the 262,144-octet delta budget (FR-057).

### 2.13 Signature

Purpose: deterministic, redaction-compatible, coverage-explicit authentication of a document state (CQ-004, CQ-005).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| param_set_id | `uint16` | yes | v1 closed allowlist: `{Ed25519-EdDSA-Protodoc-1, SHA-256}` | |
| signed_object | `[32]byte` | yes | `H(0x04 \|\| T_C_root \|\| structure_digest \|\| presentation_artefact_digest \|\| coverage_descriptor_digest)` | |
| signature_value | `[64]byte` | yes | Ed25519 signature | permitted non-deterministic site #3 |
| coverage_descriptor | `CoverageDescriptor` | yes | | |
| credential_chain_ref | `[]byte` | yes | | |
| revocation_evidence_ref | `[]byte` | no | | |
| time_attestation_ref | `[]byte` | no | | permitted non-deterministic site #4 |
| signing_intent | `uint8` (enum) | yes | | |

Invariants:
1. `signature_value` octets are `ATTEST`-typed and categorically excluded from every `CoverageDescriptor` in the document, its own included (closes FR-002 by construction).
2. Verification runs the closed 7-step EdDSA-Protodoc-1 procedure: (1)-(2) reject non-canonical/small-order `A` via integer/byte comparison against `p = 2^255-19` and a fixed 8-element low-order-point table; (3)-(4) identical checks on `R`; (5) reject `S >= L` (`L = 2^252 + 27742317777372353535851937790883648493`); (6) `k = SHA-512(R||A||msg) mod L`; (7) delegate the curve-arithmetic equation to unmodified stdlib `crypto/ed25519.Verify`.
3. The verifier always independently recomputes the acted-upon-set complement against the declared coverage, in both `TOTAL` and `SUBSET` modes — never signer-trusted.
4. Identity of a signature is the covered state, named by `(T_C_root, structure_digest)` at signing time; two signatures over the same document state are distinguishable only by `param_set_id` and `signature_value`.

Identity: `(T_C_root, structure_digest)` at signing time.

Limits: `MAX_SIGNATURES = 64` per document.

### 2.14 RedactionCommitment

See T_C leaf-redactable (§2.10); listed separately for its lifecycle role.

Invariants:
1. Removal takes octets and salt together; the retained value is the 32-octet commitment alone.
2. Post-removal search cost is `2^256`, exceeding the `2^80` floor (FR-075) by 176 bits of margin.

Identity: the subtree it commits (a T_C leaf position).

Limits: one salt per designated subtree, minted at signing time.

### 2.15 PresentationArtefact

Purpose: the exact rendering geometry and font identity a signature attests to, decoupled from the live document's current presentation state.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| profile_version | `uint16` | yes | | |
| page_geometry | `[]int64` (914400-per-inch base units) | yes | integer only | |
| font_identity | `[]FontRecord` (7 fields each) | yes | see FR-089 | |

Invariants:
1. Bound directly into `signed_object`; exempt from staleness refusal relative to the document's current live state.
2. Any other presentation than the one this artefact names is reported not-attested-by-this-signature, never silently accepted.

Identity: content digest of the artefact's own octets.

Limits: font_identity count bounded by `MAX_REFERENCES = 4194304`.

### 2.16 HistorySegment (operation batch)

Purpose: RLE-batched operation log serving all three `history_mode` values with one representation (CQ-008, CON-022).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| op_kind | `uint8` | yes | tagged `SEQUENCE-POSITION-CLAIM` \| `VALUE-CLAIM` \| `DELETE` | DP-015 classification |
| target_ref | opaque token | yes | | |
| causal_predecessor_state_id | `[32]byte` | yes | | |
| operation_total_order_key | `[32]byte` (= state_id) | yes | compared as unsigned big-endian octets | R2 tiebreak key |

Invariants:
1. Canonical emission order follows the DP-015 classification: R1 Fugue-order for `SEQUENCE-POSITION-CLAIM`, R2 `state_id`-tiebreak for `VALUE-CLAIM`, R3 delete-dominates whenever a `DELETE` shares a target with another op — never local segment ordinal.
2. `retention_point` (in the winning CommitRingRecord) gates reconstructability.
3. For `history_mode = NO_HISTORY` documents, HISTORY segments are retyped to `ErasureRecord` immediately after each commit.

Identity: `(segment ordinal, RLE-batch position)`.

Limits: bounded by `MAX_SEGMENTS` and per-batch RLE compression; no independent ceiling.

### 2.17 ErasureRecord

Purpose: content-free record of a severed unit or state (CQ-008), using the same salted-commitment category as RedactionCommitment (flagged deviation from FR-061's literal bare-digest text — see §7 and `plan.md` spec_conflicts).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| identity | `[16]byte` | yes | the severed unit's or state's original identity | |
| digest | `[32]byte` | yes | | |
| severance_commitment | `[32]byte` | yes | `H(severance-domain-tag \|\| severance_salt \|\| state_digest)` | flagged deviation, see §7 |
| severance_salt | `[32]byte` | yes, destroyed at trim | CSPRNG | |

Invariants:
1. Carries no content octets.
2. Replay of an erased unit is refused naming its identifier, never silently ignored.

Identity: the severed unit's or state's original identity.

Limits: one record per severed unit/state.

### 2.18 RescindResignRecord

Purpose: two-scheme custody chain across a future signature-scheme break, available only at a major-version migration boundary (DP-017).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| prior_signature_ref | opaque ref | yes | points to the retained pre-migration `Signature` | |
| new_param_set_id | `uint16` | yes | | |
| new_signed_object | `[32]byte` | yes | | |
| new_signature_value | `[64]byte` | yes | | |
| migration_state_id | `[32]byte` | yes | | |

Invariants:
1. Only creatable during a major-version migration.
2. The prior `Signature` and its full metadata remain retained, unmodified, in the pre-migration state.
3. `ATTEST`-typed, excluded from every `CoverageDescriptor` exactly like an ordinary `Signature`.
4. A verifier presented with both reports an explicit two-entry custody chain, never silently preferring one.

Identity: `(migration_state_id, new_param_set_id)`.

Limits: one per `(migration event, new scheme)` pair.

### 2.19 RegistryExcerpt

Purpose: self-hosted external-registry entries so a durable-profile document's own IDs stay meaningful without the central registry (DP-017).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| entries[].external_id_kind | `uint8` (enum) | yes | `extension-token` \| `shaping_profile_id` \| `unicode_version_id` | |
| entries[].external_id_value | `uint16` | yes | | |
| entries[].registered_meaning_or_spec_pointer | `[]byte` | yes | | |
| entries[].registry_snapshot_digest | `[32]byte` | yes | | |

Invariants:
1. MANDATORY for any document declaring `durable_claim = 1` (FR-011), for every external ID it actually references — no whole-registry mirroring.
2. OPTIONAL for `durable_claim = 0` documents.
3. Absence on a durable-profile document referencing an ID with no local excerpt is a structural validation failure, not silently ignored.

Identity: a `RESOURCE`-typed segment, content-addressed like any other resource.

Limits: `MAX_REGISTRY_EXCERPT_OCTETS = 1048576`.

### 2.20 ExtensionEnvelope (frame within a CONTENT or RESOURCE segment)

Purpose: the single universal carrier for any construct not defined by the core specification, so a previous-generation reader can always determine a future construct's extent, disposition and fallback without understanding its payload (FR-012 through FR-018, FR-107, CON-005, CON-020, HC-013).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| token | opaque registry identifier | yes | partitioned `registered` \| `owner-scoped` \| `permanently-retired` (DP-011); no experimental/unregistered prefix exists | a retired token is never reissued |
| length | PDL-VARINT | yes | total octet length of `payload` | |
| disposition | `uint8` (enum) | yes | exactly one of `{ignore, degrade, refuse}` | absent disposition is a structural reject naming the token (FR-013, FR-107) |
| fallback_ref | opaque ref | conditional | mandatory when `disposition in {ignore, degrade}`; expressible entirely in the core feature set; yields >=1 extractable text unit or >=1 non-decorative mark | audited for reachability (Anvil graft) |
| payload_digest | `[32]byte` (SHA-256) | yes | digest over `payload` octets | payload itself is opaque to a reader that does not implement the token |
| position_key | opaque, comparable | yes | orders the envelope among its concurrent siblings | canonical order among unrecognised constructs is `(token, position_key)` (DP-011) |

Invariants:
1. Every construct outside the core specification is carried in exactly one `ExtensionEnvelope`; a non-enveloped extension construct is rejected, not tolerated (FR-012).
2. `disposition` is one of exactly 3 closed values; a missing disposition is a structural reject naming the token, never a reader-chosen default (FR-013, FR-107).
3. `ignore` or `degrade` without a well-formed `fallback_ref` is rejected (validator rule PD-EXT-002 in `spec.md`); `refuse` requires no fallback.
4. `fallback_ref` reachability is audited: the fallback must have no other live path, closing the fallback/primary exclusion pair to a genuine, checkable relationship rather than a declared-but-unverified role.
5. EXT-DUP resolves which of `{payload, fallback}` is used as a pure function of `(disposition, reader generation)`; the same input always yields the same choice across implementations.
6. Insertion of an `ExtensionEnvelope` at a position is itself a position-claim: concurrent insertions resolve via R1 SEQUENCE-ORDER (DP-015), identically to a text or table-row insertion.
7. A permanently-retired token is recorded in `Frontmatter.retired_token_list` and never reissued to a new construct.

Identity: `(token, position_key)`.

Limits: `disposition` exactly 3 values; token-space partition is closed (3 categories), per CON-005/DP-011.

### 2.21 FontRecord

Purpose: the closed, exactly-7-field description of an embedded font that a `PresentationArtefact` may reference, carrying no executable font instruction stream (FR-089, CQ-006).

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| name | UTF-8 `string` | yes | NFC | |
| version | opaque version token | yes | | |
| digest | `[32]byte` (SHA-256) | yes | binds the subset-embedded font RESOURCE segment | content-addressed per CQ-007 |
| variation_axis_coordinates | `[]int` | yes | numeric axes only | no named-instance-by-label lookup |
| required_code_point_set | `[]rune` (or equivalent set encoding) | yes | | |
| required_layout_feature_set | `[]uint16` (feature tag set) | yes | | |
| embedding_permission_values | `uint8` (enum/bitset) | yes | | |

Invariants:
1. Exactly 7 fields exist, no more and no fewer; the record shape is closed (FR-089).
2. No font instruction stream (hinting bytecode or otherwise) is ever recorded or executed; only these 7 values are normative (CQ-006).
3. `digest` is the sole link to the actual subset-embedded font bytes, stored as an ordinary content-addressed `RESOURCE` segment; the `FontRecord` itself carries no font program bytes.

Identity: `digest` (content-addressed).

Limits: appears as an element of `PresentationArtefact.font_identity`, itself bounded by `MAX_REFERENCES = 4194304`.

### 2.22 PageDirectory entry (derived, non-normative)

Purpose: fast page-break lookup, rank-augmented over content identity, never absolute page ordinal.

Confirmed by ruling (CQ-016, clarify.md, 2026-09-15): PageDirectory intentionally has no container.abnf wire
discriminant, unlike its sibling UnitIndex/UnitIndexLeaf (`0x0A`). It is genuinely lazy-rebuilt on first
page-oriented access and bounded by `MAX_PAGES`, and plays no role in NFR-012's extraction-budget obligation
the way UnitIndex does — it does not need a persisted, randomly-addressable wire structure. This is not an
oversight; do not assign it a discriminant.

Invariants:
1. Digest-bound to its inputs; refused when stale.
2. Insertion writes only the touched path.

Identity: content-identity of the page-break unit.

Limits: `MAX_PAGES = 131072`.

### 2.23 UnitIndex (root + leaf, derived, non-normative)

Purpose: content-unit-identity to absolute-offset lookup, rebuilt (or verified) on open.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| identity | `[16]byte` | yes | opaque content-unit token | |
| absolute_offset | `uint48` | yes | | legal only because append-only guarantees extents never move once assigned an ordinal |
| length | `uint32` | yes | | |
| kind | `uint16` | yes | | |
| reserved | `[4]byte` | yes | MBZ | |
| digest | `[32]byte` | yes | | |

Invariants:
1. Storing the absolute file offset is legal only under the append-only guarantee; a partial compaction's reassignment of an uncovered ordinal is the one controlled, explicit exception, never an implicit move.

Identity: content-unit identity (16-octet opaque token).

Limits: capacity `= MAX_CONTENT_UNITS = 1048576`.

### 2.24 Table

Purpose: a grid of cells tiling exactly once with no overlap and no gap (FR-082). Added at phase 5 (analyze):
document.abnf S4 already assigns this record discriminant `0x03` and its field shape (`tbl-discriminant`,
`tbl-id`, `tbl-rows`, `tbl-columns`, `tbl-cells`), but no entity table existed here for it before this fix.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| tbl_id | unit-id | yes | | content-unit identity of the table itself |
| direction | `uint8` | mandatory | closed value set `{0 = LTR, 1 = RTL}` | base writing direction (FR-032); provisional per T-0267 ruling (clarify-002.md, OPEN) |
| tbl_rows | `[]unit-id` | yes | plain sequence, current document order | row ids |
| tbl_columns | `[]unit-id` | yes | plain sequence, symmetric to `tbl_rows` | column ids |
| tbl_cells | `[]cell-entry` | yes | each entry names `(cell_row, cell_col, content)` | `cell_row` and `cell_col` must each resolve into `tbl_rows`/`tbl_columns` respectively |

Invariants:
1. `tbl_cells` tiles the `tbl_rows` x `tbl_columns` grid exactly once: no two entries name the same `(cell_row, cell_col)` pair, and no `(row, col)` pair in range is left unnamed (FR-082).
2. A `cell_entry` naming a row or column id absent from `tbl_rows`/`tbl_columns` is a structural reject.

Identity: `tbl_id` (content-unit identity).

Limits: `tbl_rows`/`tbl_columns` counts and `tbl_cells` count are each bounded by `MAX_REFERENCES = 4194304`.

### 2.25 Note

Purpose: a footnote or endnote, anchored to a point in the main text, carrying its own text block (FR-036-style
structural placement). Added at phase 5 (analyze): document.abnf S5 already assigns discriminant `0x04` and the
field shape (`note-discriminant`, `note-id`, `note-anchor`, `note-body-block`, `note-placement`), but no entity
table existed here for it before this fix.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| note_id | unit-id | yes | | content-unit identity of the note |
| note_anchor | anchor-point (§3.1) | yes | | the point in the main text the note attaches to |
| note_body_block | unit-id | yes | must resolve to a TextBlock (§2.7) | the note's own text |
| note_placement | `uint8` (enum) | yes | `0x00` footnote, `0x01` endnote; `0x02-0xFF` reserved, rejected | |

Invariants:
1. `note_body_block` not resolving to a `TextBlock` is a structural reject (FR-108's reference-resolution rule).
2. `note_placement` outside the closed two-value enum is a structural reject.

Identity: `note_id` (content-unit identity).

Limits: none beyond `MAX_CONTENT_UNITS` and `MAX_REFERENCES`.

### 2.26 CrossReference

Purpose: a reference to another content unit rendered via a named presentation function rather than frozen
literal text (FR-084, FR-085). Added at phase 5 (analyze): document.abnf S5 already assigns discriminant `0x05`
and the field shape (`xref-discriminant`, `xref-id`, `xref-target`, `xref-kind`, `xref-anchor`), but no entity
table existed here for it before this fix.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| xref_id | unit-id | yes | | content-unit identity of the cross-reference itself |
| xref_target | unit-id | yes | must resolve to exactly one unit present in the document (FR-108) | the referenced unit |
| xref_kind | `uint8` (enum) | yes | `0x00` internal hyperlink, others per document.abnf | selects the presentation function |
| xref_anchor | anchor-point (§3.1) | yes | | the reference's own position in the main text |

Invariants:
1. `xref_target` not resolving to exactly one present unit is a structural reject (FR-108).
2. `xref_target` is never a literal frozen string; staleness is determined without computing layout (FR-085).
3. A `cross-reference` edge participates in FR-109's reference-graph cycle check as one of the 5 named edge kinds (integrity.abnf S8); `xref_id -> xref_target` forming a cycle with other named edges is a structural reject.

Identity: `xref_id` (content-unit identity).

Limits: none beyond `MAX_CONTENT_UNITS` and `MAX_REFERENCES`.

### 2.27 ROOT_SEQUENCE (authored reading-order record)

Purpose: the document's AUTHORITATIVE logical top-to-bottom reading order over its content units (FR-036).
Added per the T-0267 ruling (clarify-002.md, OPEN — provisional). It resolves the self-disclosed absences that
`integrity.abnf` S2.2.1 flagged for T_C's own subtree traversal and the `plan.md` Section 3 extraction
storage-order / reading-order inconsistency: reading order is an authored property, never inferred from
coordinates or storage/append order.

| Field | Type | Required | Constraint | Notes |
|---|---|---|---|---|
| rs_id | unit-id | yes | singleton per document | the ROOT_SEQUENCE record's own identity |
| rs_order | `[]unit-id` | yes | one entry per content unit; no duplicates, no omissions | the authored reading order, in top-to-bottom order |

Invariants:
1. `rs_order` lists EVERY content unit in the document EXACTLY ONCE — no duplicates, no omissions (validator rule
   PD-A11Y-005; T-0274). A unit absent from `rs_order`, or listed twice, is a structural reject naming the unit id.
2. `rs_order` is the authoritative reading order and is INDEPENDENT of storage/append order: a unit inserted
   logically mid-document (and therefore appended at the end of storage) still appears at its authored position
   in `rs_order`.
3. T_C subtree traversal (integrity.abnf S2.2.1) and the `extract` walk both key on `rs_order`, not storage order
   or a CSPRNG placeholder (T-0275, T-0276).

Identity: `rs_id` (singleton content-unit identity per document).

Limits: `rs_order` count is bounded by `MAX_CONTENT_UNITS`.

### 2.28 Accessibility role map (FR-038)

Purpose: a CLOSED set of accessibility roles and a TOTAL, SINGLE-VALUED map from every construct kind
defined in `document.abnf`'s frame-discriminant registry (0x01–0x0E) onto exactly one role. Assistive
technology consumes the role, not the raw discriminant. Added per the T-0267 ruling (clarify-002.md,
OPEN — provisional). Validator rule PD-A11Y-002 (T-0280) fails closed on any construct instance that does
not resolve to exactly one role from this table.

Closed role enum: `PROSE`, `ANNOTATION`, `TABLE`, `NOTE`, `REFERENCE`, `GRAPHIC`, `STRUCTURAL_NAVIGATION`,
`NON_SEMANTIC`.

| Construct kind (discriminant) | Accessibility role |
|---|---|
| TEXT_BLOCK (0x01) | PROSE |
| ANNOTATION (0x02) | ANNOTATION |
| TABLE (0x03) | TABLE |
| NOTE (0x04) | NOTE |
| CROSS_REFERENCE (0x05) | REFERENCE |
| EXT_ENVELOPE (0x06) | NON_SEMANTIC |
| RASTER_IMAGE (0x07) | GRAPHIC |
| FONT_SUBSET (0x08) | NON_SEMANTIC |
| REGISTRY_EXCERPT (0x09) | NON_SEMANTIC |
| UNIT_INDEX_LEAF (0x0A) | NON_SEMANTIC |
| HISTORY_OP_BATCH (0x0B) | NON_SEMANTIC |
| PRESENTATION_ARTEFACT (0x0C) | NON_SEMANTIC |
| ERASURE_RECORD (0x0D) | NON_SEMANTIC |
| ROOT_SEQUENCE (0x0E) | STRUCTURAL_NAVIGATION |

Invariants:
1. The map is TOTAL over the assigned construct kinds: every discriminant 0x01–0x0E maps to exactly one
   role. A construct kind absent from this table (a reserved or future discriminant) has NO role and fails
   closed under PD-A11Y-002.
2. The map is SINGLE-VALUED: no construct kind maps to more than one role.
3. The role enum is CLOSED for v1: a new role requires a spec amendment, not a document-level extension.

## 3. Relationships

```
Header ──(frozen bytes)── CommitRingRecord (7 slots, winner selected)
                                │
                                ├─ ledger_root ──> T_S (over SegmentTableSlot[])
                                ├─ structure_digest ──> covers ──> Header, Frontmatter,
                                │                                   SegmentTableSlot,
                                │                                   IntegrityBlock (by
                                │                                   covered_prefix_regions_bitmask)
                                └─ retention_point ──> HistorySegment (RLE batch)

SegmentTableSlot[ordinal] ──1:1── Segment (CONTENT | RESOURCE | HISTORY | ATTEST)

Segment(CONTENT) ──contains── TextBlock ──contains── Run[] ──addresses── Annotation/Range (via run_id, birth_ordinal)

T_C node (leaf) ──canon(subtree)── Segment(CONTENT | RESOURCE) subtree
T_C_root ──signed by── Signature.signed_object
Signature ──names── CoverageDescriptor ──ranges over── SegmentTableSlot ordinals (excludes all ATTEST)
Signature ──1:1── PresentationArtefact (digest-bound)
RescindResignRecord ──prior_signature_ref──> Signature (retained pre-migration)
RegistryExcerpt ──entries reference──> Header.{unicode_version_id, shaping_profile_id}, extension tokens
Signature ──PresentationArtefact.font_identity[]──> FontRecord ──digest──> Segment(RESOURCE) (subset font bytes)

Segment(CONTENT | RESOURCE) ──contains── ExtensionEnvelope (token, position_key) ──fallback_ref── core-feature-set construct
ExtensionEnvelope.token ──partitioned by── registered | owner-scoped | permanently-retired (retired tokens listed in Frontmatter.retired_token_list)

HistorySegment ──retypes to (NO_HISTORY mode)──> ErasureRecord

PageDirectory / UnitIndex (derived) ──rebuilt from── SegmentTableSlot[] + Segment content
                                                    (never authoritative; digest-checked against T_S/T_C)
```

| Relationship | Cardinality | Referential rule |
|---|---|---|
| CommitRingRecord -> SegmentTableSlot[] | 1 : 0..16384 | winner's `segment_count` bounds the live slot range; non-winners are inert |
| SegmentTableSlot -> Segment | 1 : 1 | slot ordinal is the segment's sole identity; deletion is never in-place, only via compaction of an uncovered ordinal |
| Segment(CONTENT) -> TextBlock | 1 : 0..n | a segment may batch multiple text blocks as sibling frames |
| TextBlock -> Run | 1 : 1..n | ordered by `base_ordinal`; split/merge preserve `run_id` lineage |
| Run -> Annotation/Range endpoint | 1 : 0..n | an annotation endpoint references a run by `(run_id, birth_ordinal)`, never by index |
| T_C leaf -> Segment subtree | 1 : 1 | canonical depth-first order, independent of storage ordinal |
| Signature -> CoverageDescriptor | 1 : 1 | embedded, not independently addressable |
| Signature -> Segment(ATTEST) | 0 : 1 (self) | a signature's own bytes are always `ATTEST`-typed and always excluded from its own and every other signature's coverage |
| RescindResignRecord -> Signature | 1 : 1 (prior) | the prior signature is retained unmodified; never overwritten |
| RegistryExcerpt -> Header fields | 1 : n | mandatory presence gated by `durable_claim = 1`; referential completeness checked at validation (§7) |
| PresentationArtefact -> FontRecord | 1 : 0..n | `font_identity` array; each element digest-bound to a RESOURCE segment holding the subset-embedded font bytes |
| Segment(CONTENT\|RESOURCE) -> ExtensionEnvelope | 1 : 0..n | a segment may carry zero or more envelope frames alongside its typed frames |
| ExtensionEnvelope -> core-feature-set fallback | 0..1 : 1 | mandatory when `disposition in {ignore, degrade}`; reachability audited, never merely declared |
| HistorySegment -> ErasureRecord | 1 : 1 retype (NO_HISTORY only) | irreversible; content octets are destroyed at the retype |

## 4. Indexes and derived structures

| Structure | Stored or derived | Rebuilt on open | Cost |
|---|---|---|---|
| SegmentTableSlot array | stored (authoritative) | no — read directly | O(1), fixed 786,432-octet region read |
| T_S (storage integrity tree) | derived from SegmentTableSlot digests; a cached root is stored in CommitRingRecord.ledger_root | recomputed to verify, not to use — the cached root is trusted only after recomputation matches | O(segment_count), arity-16 tree, depth 4 |
| T_C (content commitment tree) | derived from content subtree canon(); root cached in Signature.signed_object's preimage | recomputed at signature verification time; never trusted from a stored slot | O(content_unit_count), arity-16, depth <= 5 |
| structure_digest | derived, recomputed FRESH from current file octets restricted to the claimed covered set at every verification | always recomputed, never read from a stored slot | O(covered ordinal count), bounded by MAX_SEGMENTS |
| UnitIndex (root + leaf) | stored as an accelerator; content-addressable truth is T_C, not the index | verified against T_C on open; rebuilt if absent or stale | O(MAX_CONTENT_UNITS) worst case, amortized O(log n) per lookup after build |
| PageDirectory | derived, non-normative | rebuilt lazily on first page-oriented access; digest-checked against inputs | O(touched path) per insertion, O(MAX_PAGES) worst-case full rebuild |
| CoverageDescriptor range sets | stored inside Signature | canonical form checked (merge-adjacent, sorted) at every read, never assumed valid | O(range count), bounded by MAX_SEGMENTS |

## 5. Ceilings

Per CP-007, every structural limit is an exact decimal integer with a requirement identifier. This is the complete v1 ceiling table.

| Ceiling | Value (exact decimal) | Requirement ID |
|---|---|---|
| File prefix total size | 1048576 | TR-006, TR-007, TR-008 |
| Header region size | 512 | FR-006, FR-007, FR-008, FR-009, FR-011 |
| Header frozen-forever window | 32 (octets `[0,32)`) | CP-008, FR-123 |
| CommitRing slot count | 7 | HC-027, FR-117 |
| CommitRing slot size | 512 | HC-027 |
| CommitRing region size | 3584 (7 x 512) | HC-027 |
| Frontmatter region size | 258048 | FR-051, FR-053, FR-054 |
| Frontmatter preview_raster max | 131072 | FR-051 |
| Frontmatter preview_source_snapshot max | 4096 | FR-053 |
| Frontmatter document_metadata max | 16384 | FR-054 |
| Frontmatter retired_token_list max | 16384 | DP-011 |
| SegmentTable slot count (MAX_SEGMENTS) | 16384 | TR-006, NFR-004 |
| SegmentTable slot size | 48 | TR-006 |
| SegmentTable region size | 786432 (16384 x 48) | TR-006 |
| MAX_FRAMES_PER_SEGMENT | 8192 | FR-102 |
| MAX_DECODED_UNIT | 268435456 | NFR-030 |
| MAX_TEXT_UNIT_OCTETS | 65536 | DP-009 |
| PDL-VARINT 1-octet range | 0 to 252 | DP-002 |
| PDL-VARINT 2-octet-extension range | 253 to 65535 | DP-002 |
| PDL-VARINT 4-octet-extension range | 65536 to 4294967295 | DP-002 |
| PDL-VARINT 8-octet-extension range | 4294967296 to 18446744073709551615 | DP-002 |
| PDL-TLV field tag width | 1 octet, max 256 fields per record type | DP-002 |
| Run collision bound scale | below 2^34 mints per lineage at P < 2^-60 | FR-023 |
| MAX_REFERENCES (annotations/font records) | 4194304 | FR-028, FR-089 |
| Annotation boundary_behaviour values | 4 (exactly) | FR-026, FR-027 |
| T_C / T_S tree arity | 16 (always, absent children filled) | DP-006 |
| T_C max depth | 5 (16^5 >= 1048576) | MAX_CONTENT_UNITS |
| T_S depth | 4 (16^4 = 65536 capacity) | headroom over MAX_SEGMENTS |
| MAX_CONTENT_UNITS | 1048576 | T_C sizing |
| Redaction search-cost floor | 2^80 | FR-075 |
| Redaction actual search cost | 2^256 | FR-075 (margin) |
| MAX_SIGNATURES | 64 | FR-063 |
| de Casteljau flattening tolerance | 762 base units (914400/300/4) | CON-012, HC-020 |
| de Casteljau max recursion depth | 16 | HC-020 |
| PLP-1 block size | 8 x 8 | CQ-010 |
| PLP-1 fixed quantization matrix count | 8 | CQ-010 |
| PLP-1 decode line-count estimate | 480 | CP-003 feasibility |
| Font recorded values | 7 (exactly) | FR-089 |
| MAX_PAGES | 131072 | PageDirectory |
| MAX_REGISTRY_EXCERPT_OCTETS | 1048576 | DP-017 |
| Extraction role statement ceiling | 150 | CP-002 (current estimate ~65) |
| Validating-and-verifying role statement ceiling (cumulative) | 400 | CP-002 (current estimate ~220) |
| Rendering role statement ceiling (cumulative) | 500 | CP-002 (current estimate ~321) |
| Extraction reference-implementation line budget | 1000 (source lines, no font/graphics dependency) | CP-010 |
| Extraction reference-implementation current estimate | 770 (source lines) | CP-010 |
| NFR-008 write amplification per K-octet edit | 8K + 262144 octets, tested at 1 MB / 50 MB / 500 MB | NFR-008 |
| FR-057 combined index+integrity delta per edit | 262144 octets | FR-057 |
| FR-055 dependent reads to locate and read a unit | 3 (inclusive of the unit read itself) | FR-055 |
| NFR-012 bounded read fraction | 15 percent of file | NFR-012 |
| NFR-013 extraction CPU budget | 20 seconds | NFR-013 |
| NFR-014 extraction peak memory | 33554432 octets | NFR-014 |
| NFR-015 cold-cache preview budget | 300 milliseconds | NFR-015 |
| NFR-016 preview peak memory | 67108864 octets | NFR-016 |
| NFR-017 open-and-render peak memory | 209715200 octets | NFR-017 |
| NFR-018 per-page-plus-resources read budget | 8388608 octets | NFR-018 |
| CP-012 fuzzing memory-oracle multiplier | 4x input octet length (floor-exempted below 1048577 octets, see §7) | NFR-030 |
| CP-012 disclosure clock | 90 days, report to fix or advisory | CP-012 |
| NFR-026 extracting-and-validating implementer budget | 5 working days | NFR-026 |
| NFR-027 rendering implementer budget | 30 working days (current estimate 28-49 days, at risk — see §7) | NFR-027 |
| Negative corpus minimum | 200 hostile cases | CP-011 |
| NFR-032 complete-history size overhead | 2.0x the size of the same visible content saved with no history | NFR-032 |
| NFR-033 open/render time and memory bound | proportional to current content only; two documents with identical visible content and differing history bound within 1.5x of each other at 10x operation-count difference | NFR-033 |

## 6. Ordering rules

Every place where two implementations could otherwise diverge on order is closed here with an exact total order, per CP-003.

1. **Segment storage order.** The append order of `SegmentTableSlot` ordinals is the order operations occurred in at write time. It is monotonic and is never re-derived from content. It is never the order used for T_C traversal (rule 4) or HISTORY emission (rule 6).
2. **CommitRing winner selection.** Highest `sequence` whose `record_digest` verifies AND `ledger_length <= file length` wins. A tie at the highest valid `sequence` across two or more slots is a structural reject naming both slot indices, never resolved by slot position or wall-clock.
3. **PDL-TLV sorted vectors.** Compared by unsigned byte-lexicographic order over the full encoded octets (`bytes.Compare` semantics). Key/value pairs are sorted by key octets only, ties broken by declaration order being illegal (keys are unique by construction).
4. **T_C traversal order (subtree ordinal).** Canonical depth-first order over content subtrees, fixed by the document's logical structure (paragraph/table/resource nesting), never by storage ordinal and never by digest value.
5. **CoverageDescriptor range canonicalisation.** `covered_segment_ranges` and `uncovered_segment_ranges` are each a sorted vector of half-open `(start_ordinal, end_ordinal)` pairs, sorted ascending by `start_ordinal`, with mandatory merge-adjacent canonicalisation: two ranges `(a,b)` and `(b,c)` are illegal as separate entries and must appear as `(a,c)`. Exactly one valid encoding exists per coverage set; any other encoding is rejected, not accepted as an alternate form.
6. **HistorySegment operation emission order.** Governed by the DP-015 4-way classification, not storage ordinal: (0) DISJOINT-COMMUTE operations may appear in either order since they commute; (1) R1 SEQUENCE-ORDER operations (any position-claim: text/table-row/table-column/extension-envelope/annotation insertion) use Fugue's proven non-interleaving total order; (2) R2 TOTAL-ORDER-TIEBREAK operations (any two value-claims on the same field of the same unit) order by `state_id` compared as unsigned big-endian octets, smaller first; (3) R3 DELETE-DOMINATES orders the delete before the dominated operation, whose disposition is read from a fixed per-kind table (no-op or orphan-carry), never invented per instance.
7. **T_S / T_C internal node child order.** Fixed at exactly 16 slots per node, indexed by the child's position in the tree (never sorted by digest value); absent children are `ABSENT_CHILD_DIGEST` at their fixed index, not omitted or compacted to the front.
8. **Ed25519 verification step order.** The 7 steps of EdDSA-Protodoc-1 execute in the stated order: canonical/small-order checks on `A`, then on `R`, then the `S < L` check, then challenge-hash computation, then delegation to `crypto/ed25519.Verify`. A step failing short-circuits before later steps run, so no later step's cost or behavior is observable on a rejected input.
9. **Run split/merge base_ordinal.** A split assigns the successor run a `base_ordinal` strictly greater than the predecessor's last covered ordinal; a merge is legal only between a run and its syntactic successor (contiguous `base_ordinal`, matching `run_id` lineage), never between arbitrary runs.

## 7. Validation

A validator applies checks in the following order. Earlier checks strictly precede later ones; a document failing an earlier check is rejected on that check's verdict and later checks do not run (CP-006: verification precedes decoding).

1. **Magic and frozen header window.** Read `Header[0,32)`. Reject unrecognized magic. Reject `format_major` greater than supported (FR-123, decline naming the version, zero dispositions applied) before evaluating any other field.
2. **Capability arithmetic.** Reject `capability_required > capability_written`, naming both values (FR-010).
3. **CommitRing winner selection.** Select the winning slot per §6 rule 2. A tie at the highest valid `sequence` is a distinct rejection (PD-RING-001), independent of and prior to any content check.
4. **Truncation/rollback check.** Reject if the winning record's `ledger_length` exceeds the actual file length.
5. **Bounded-prefix structural checks (TR-006/007/008).** Validate `SegmentTableSlot` array bounds-safe arithmetic (`frame_count*48+32 <= length-64`) for every slot without decoding any frame. A malformed slot is rejected at the offset/length pair that fails, before any segment is opened.
6. **T_S recomputation.** Recompute `T_S` over live `SegmentTableSlot.segment_digest` values; compare to `CommitRingRecord.ledger_root`. Divergence is a distinct verdict from a missing/invalid signature.
7. **Segment-type and coverage well-formedness.** For every `SegmentTableSlot`, confirm `segment_type` is one of the 4 closed values. For every `Signature`, confirm its `CoverageDescriptor` is well-formed per §6 rule 5 and names no `ATTEST` ordinal; malformed coverage is rejected before signature verification runs.
8. **Cycle detection (5 edge kinds).** Run FR-109's cycle detection over all 5 named edge kinds (structural moves, extension-envelope fallback references, annotation anchor references, run split/merge lineage, RescindResignRecord chains) before any structural ceiling (§5) is checked, so a cyclic input cannot be misreported as merely over-ceiling.
9. **Structural ceilings.** Check every value in §5's table against its exact integer. A ceiling violation produces a distinct, non-validity status (PD-BUDGET-xxx) when the document is otherwise structurally valid but exceeds a resource budget (CP-007); it is never conflated with an invalid-document verdict.
10. **T_C recomputation and signature verification.** For each `Signature`, recompute `T_C_root` over the covered subtree set, recompute `structure_digest` fresh per its DP-013 definition, recompute `signed_object`, and run the 7-step EdDSA-Protodoc-1 procedure (§6 rule 8). A signature over an unavailable pre-migration state (NO_HISTORY mode, post-migration) reports a distinct "covering an unavailable state" verdict, never "invalid" and never "valid."
11. **NFC and identity checks.** Reject any TextBlock/Run/metadata field failing NFC quick-check (CON-003), and reject any use of an `IDENTITY-COMPONENT`-tagged field for positional arithmetic detected by static audit (A-FIELD-ROLE).
12. **RegistryExcerpt completeness.** For a document with `durable_claim = 1`, confirm a `RegistryExcerpt` entry exists for every externally-registered ID the document references; absence is a structural validation failure, distinct from every earlier verdict category.
13. **Extension-envelope reachability and disposition.** Confirm every `EXT_ENVELOPE` fallback reference is reachable and that disposition selection is a pure function of `(disposition, reader generation)`; an unregistered tag is rejected, never skipped.

**Error precedence when two rules both reject.** The earliest-numbered check in this list that fails determines the reported verdict; later checks are not run and their potential failures are not reported. This is a total order over check categories, chosen so that a caller can always determine, from the verdict alone, which check category failed first, without ambiguity between two simultaneously-true rejection conditions. The one documented exception is step 9 (structural ceilings), whose PD-BUDGET-xxx status is explicitly defined by CP-007 as non-validity and is therefore reported alongside, not instead of, any validity verdict produced by steps 1-8 if both are independently true — a validator MUST report both codes in that case, never suppress one.

**Data-quality traceability (ISO/IEC 25012).** `created_at`/`updated_at`-equivalent traceability in Protodoc is carried structurally, not as user-editable timestamp fields: `CommitRingRecord.sequence` plus `state_id`/`parent_state_id` gives a verifiable causal chain for every commit, and `Signature.time_attestation_ref` gives an attested point in time for every signed state, both independent of any client clock. Completeness (`NOT NULL`-equivalent) is enforced structurally by fixed-width fields with no optional-but-required case: every field marked "yes" in the Required column above is present in every valid document by construction, checked at validation step 5/9, never left to application-level convention. Precision is stated per field: all monetary/geometric values use integer base units (914400 per inch, CON-012), never floating point, per DP-002's "no floats, ever" rule.

**Known deviations, flagged not hidden (traceable to `plan.md` §spec_conflicts):**
- `ErasureRecord.severance_commitment` uses a salted commitment, not the bare unsalted digest FR-061's literal text specifies, because the literal text reopens the exact FR-075 privacy hole for trimmed history. This validator implements the salted form; the requirement text needs a clarifying amendment before this is anything other than a flagged deviation.
- The CP-012/NFR-030 fuzzing memory oracle in step 9 uses a heap-for-document-data reading with a floor exemption below 1,048,577 octets, since the literal 4x-of-input-length bound is unsatisfiable by any real process on small inputs. This reading is applied consistently everywhere NFR-030 is checked.
