# Protodoc: Implementation Plan

Status: APPROVED, Eyvar 2026-09-07 | Spec ID: 001-protodoc-format-core | Phase: 3 (plan) | Date: 2026-09-05

This plan states HOW the 197 approved requirements (125 FR, 34 NFR, 26 CON, 12 TR) get built. It does not reopen `spec.md` or `clarify.md`. Where the chosen design cannot satisfy a requirement's literal text, that is recorded as a conflict in Section 9, not silently designed around. The architecture named here is **Protodoc Ledger (PDL), Repair Set 2**, selected from five independently-drafted candidates (PDL, Slab-DAG, PDCS, Anvil, Stratum) scored against three adversarial lenses and then refuted along three more angles; see `research.md` for the full scoring and refutation record. This document states the decisions that record leaves standing, the exact mechanisms, and the numbers.

Sibling artifacts already drafted in this phase: `data-model.md` (entity field tables, ceilings, ordering rules, the 13-step validation pipeline) and `contracts/container.abnf` (the byte-exact prefix and ledger grammar). This plan is consistent with both; where a number here is stated at lower precision than those files, the file is authoritative for the byte offset and this plan is authoritative for the reasoning behind it.

---

## 1. Architecture overview

### Thesis

A Protodoc file is never itself the canonical form. Three distinct objects exist, and no tool may collapse them into one:

1. **The file at rest**: a fixed 1,048,576-octet prefix followed by an append-only ledger of immutable, single-typed, sealed segments. Never rewritten in place except by an explicit, signature-aware compaction.
2. **The canonical octet sequence C(S)**: a deterministic depth-first traversal of document state S, streamed on demand for hashing, signing, or comparison. Never materialized as a file. This is the object CQ-003 makes signable.
3. **The placement function `place()`**: the non-discretionary rule that turns an edit into physical writes. `place(prior, no-op) = prior` (zero writes). `place(prior, edit)` = append exactly one sealed segment plus patch a bounded set of fixed-width prefix slots. `place(BOTTOM, S) = L*(S)`, a full canonical re-emission, used only by compaction, publish, and migration, and refused whenever any signature is present (NFR-004).

The three nested prefix budgets required by HC-006 and HC-007 (512 / 262,144 / 1,048,576 octets) are not just satisfied, they are exact region boundaries with zero slack: 512 + 3,584 + 258,048 + 786,432 = 1,048,576. There is no separate trailing "integrity block" region: the T_C root and the unit-index routing table are carried as fields inside the winning CommitRingRecord itself, because the SegmentTable already consumes the entire remainder of the prefix after Header, CommitRing, and Frontmatter. This is stated explicitly here because the original architecture summary's prose names an "IntegrityBlock+UnitIndexRoot" region as if it were a fifth disjoint byte range; the arithmetic above shows no such range fits, and `contracts/container.abnf` already resolves this by co-locating those two values inside each 512-octet ring slot. No requirement is reopened by this; it corrects an internal inconsistency in the architecture's own prose using the same document's more heavily corroborated numbers (SegmentTable size is independently corroborated by the T_S arity-16/depth-4 sizing and by the FR-057/NFR-008 write-budget arithmetic in Section 5).

Everything else follows from keeping those three objects separate: octet-stable saves and bounded per-edit writes are jointly satisfiable (CQ-003 option A) because the file at rest never has to equal the canonical form; content addressing and stable storage ordinals coexist (CQ-007 option B) because storage order is a monotonic counter, never a digest; and signature coverage can be total by construction over content (T_C_root signed directly) without also having to enumerate every physical prefix octet, because `structure_digest` is a separately-defined, freshly-recomputed projection over exactly the covered set (Decision DP-013).

### Layered figure

```
 ┌────────────────────────────────────────────────────────────────────┐
 │ FIXED PREFIX -- exactly 1,048,576 octets, zero self-description,   │
 │ zero decode required (TR-006 / TR-007 / TR-008)                    │
 │                                                                     │
 │  [0,        512)  HEADER            frozen [0,32) forever          │
 │  [512,     4096)  COMMIT RING       7 x 512B self-digesting slots  │
 │                                     (T_C root + index-route live   │
 │                                      inside the winning slot)      │
 │  [4096,  262144)  FRONTMATTER       preview + metadata, PDL-TLV    │
 │  [262144,1048576) SEGMENT TABLE     16,384 x 48B ordinal slots     │
 │                                     (the sole stored-unit          │
 │                                      inventory, CQ-007)            │
 └───────────────────────────────┬────────────────────────────────────┘
                                  │ append only, sealed, single-typed
                                  ▼
 ┌────────────────────────────────────────────────────────────────────┐
 │ LEDGER -- offsets >= 1,048,576, *segment                            │
 │ CONTENT | RESOURCE | HISTORY | ATTEST, never rewritten in place    │
 │ except by explicit compaction (full or partial)                    │
 └────────────────────────────────────────────────────────────────────┘

 Cross-cutting layers (not physical regions, mechanisms over the above):

   Identity & Anchor  ── run_id + base_ordinal, zero persisted counted positions
   Integrity & Signature ── T_S (storage), T_C (content/redaction), EdDSA-Protodoc-1
   History & Erasure  ── HistorySegment (RLE ops) + retention_point + ErasureRecord
   Extensibility       ── EXT_ENVELOPE: token, length, disposition, fallback, digest
   Rendering & Resource── PLP-1 codec, exact-rational rasterizer, Knuth-Plass reflow
   Concurrent-Edit/Merge── DISJOINT-COMMUTE / R1 SEQUENCE-ORDER / R2 TIEBREAK / R3 DELETE
   Evolution/Migration ── frozen header window, refusal-first two-phase migration
```

### Layer table

| Layer | Responsibility | Mechanism | Requirements served |
|---|---|---|---|
| Prefix / Container | Content-independent identity, bounded-prefix determination with zero decode, crash-atomic commit | Fixed 1,048,576-octet prefix as above; 7-slot self-digesting commit ring, equal-sequence ties rejected (PD-RING-001); `structure_digest` names its own prefix-region coverage via a bitmask over {HEADER, RING_WINNER, FRONTMATTER, SEGMENT_TABLE} and includes `ledger_length(8)` to block truncation replay | FR-006..011, FR-051, FR-053, FR-054, TR-006..008, FR-117, HC-006, HC-007, HC-015, HC-027 |
| Ledger / Storage | Append-only sealed segments, bounded per-edit write cost, a placement function with zero writer discretion | `place()` as above; text frames capped at 65,536 octets; partial compaction reclaims only ordinals outside every present signature's covered range | HC-001..005, NFR-001..004, NFR-008..010, FR-056..058, CQ-003, CQ-007 |
| Identity & Anchor | Character-granular durable identity that compresses to run granularity, zero persisted counted positions | `run_id` = 128-bit `crypto/rand` token per contiguous typing burst; character identity = `(run_id, base_ordinal+i)`; a closed 7-tag field-role vocabulary (IDENTITY-COMPONENT, PHYSICAL-OFFSET-OR-LENGTH, ORDINAL, DIGEST, ENUM, SCALAR-VALUE, COUNT) with no POSITION role available to declare | HC-010, HC-011, CQ-001, CQ-002, FR-019..030, CON-001 |
| Integrity & Signature | One storage-integrity tree, one content-commitment/redaction tree, deterministic redaction-compatible signatures with mechanically classifiable coverage | T_S (storage, arity-16/depth-4) and T_C (content, domain tags 0x02/0x07/0x08) both fixed at exactly 16 child slots with `ABSENT_CHILD_DIGEST = SHA-256(0x00)` filling absent children; `signed_object = H(0x04\|\|T_C_root\|\|structure_digest\|\|presentation_artefact_digest\|\|coverage_descriptor_digest)`; Ed25519 verified by the closed 7-step EdDSA-Protodoc-1 procedure | CQ-004, CQ-005, HC-014..017, FR-063..077, FR-115, CON-015, CON-016 |
| History & Erasure | One representation serving all three CON-022 history modes, a lawful in-place trim, content-free erasure records | Typed HISTORY segments of RLE-batched operations ordered by the DP-015 classification; `retention_point` in the commit ring gates reconstructability; NO_HISTORY mode retypes HISTORY to ErasureRecord immediately after each commit | CQ-008, HC-018, HC-019, FR-059..062, CON-022..025 |
| Extensibility | Forward compatibility with no reliance on the wire encoding's own skip behavior | First-class EXT_ENVELOPE frame: token, PDL-VARINT length, disposition (exactly 3 values), fallback_ref, payload digest; unregistered tags rejected, never skipped | CP-013, CON-005, CON-020, HC-013, FR-012..018, FR-107 |
| Rendering & Resource | Deterministic, font-less, network-less rendering with resources segregated from text | RESOURCE segments always after CONTENT segments; PLP-1 lossy codec (fixed 8x8 blocks, ~480 decode lines); one restricted-PNG lossless profile; exact-rational active-edge rasterizer; Knuth-Plass integer-demerits reflow | CQ-006, CQ-010, HC-020..023, HC-029, FR-089..101, FR-111..114, CON-012..014 |
| Concurrent-Edit / Merge | Exactly one outcome per pair of concurrent operation kinds, via an exhaustive classification rather than a per-pair table | Every operation is a position-claim, a value-claim, or a delete: (0) DISJOINT-COMMUTE, (1) R1 SEQUENCE-ORDER (Fugue's proven order), (2) R2 TOTAL-ORDER-TIEBREAK (state_id, unsigned big-endian compare), (3) R3 DELETE-DOMINATES (fixed per-kind disposition table) | FR-092..096, FR-116, TR-003, TR-010, HC-026 |
| Evolution / Migration | Decade-scale forward compatibility with a permanently frozen version-detection window | Header octets `[0,32)` frozen forever; refusal-first two-phase migration emitting a fresh file; RESCIND-AND-RESIGN available only at a major-version boundary | CP-008, HC-028, FR-119..123, CON-020 |

---

## 2. Decisions

Each decision point below states what was decided, why, and what was rejected. Decision IDs (DP-001..DP-017) match `research.md`'s numbering so the two documents cross-reference without renumbering.

### DP-001: Container layout

**Decision.** Fixed 512-octet header, 7-slot self-digesting commit ring, bounded preview-plus-metadata region, front-placed fixed-capacity segment table (16,384 slots), append-only extent ledger. Unchanged from the architecture's first repair pass.

**Rationale.** No refutation angle in this repair round attacked the container layout itself; every kill shot targeted an encoding or algorithm layered on top of it (Ed25519's procedure, the lossy codec, the merge classification, the tree filler value, the coverage-descriptor wire form). A skeleton that survives three adversarial rounds without needing to be re-derived is evidence it was sound going in.

**Rejected.** Zip-of-parts (fails FR-006's content-independent window because the first octets are a content-derived local file header; fails FR-055's 3-read bound because the manifest is a backward-scanned, variable-length-comment-terminated end-of-central-directory record; fails FR-056/NFR-009 because saving rewrites the central directory and shifts every later local-header offset; fails CON-008 because entry names are path-like with case-folding ambiguity; fails NFR-003 because per-writer extra fields, timestamps, and DEFLATE level are unspecified; fails HC-015 because data-descriptor slack and prepended octets are permitted and unclassifiable). A page-oriented file with a free list (reintroduces reclaimable holes carrying prior octets, which HC-019 forbids as redaction residue). A trailing typed-chunk index (forbidden three times over by TR-006/007/008's requirement that the full inventory sit in the leading 1,048,576 octets). A single flat length-prefixed message tree, i.e. protobuf-as-container (a leaf edit rewrites every enclosing length prefix and shifts the whole contiguous encoding, a direct HC-004 kill).

### DP-002: Structural byte encoding

**Decision.** Two layers, split by mutation frequency. The PREFIX (header, commit ring, segment-table slots) is raw fixed-layout structs at fixed byte offsets: zero tags, zero self-description, chosen because HC-006/HC-007 need zero-parse, content-independent access. Everything written once and sealed (segment frames, Frontmatter's record, every content-model and integrity record) uses **PDL-TLV**, a closed grammar: 1-octet field tags scoped per record-type schema (maximum 256 fields per type, a permanent ceiling per CON-009), fields strictly ascending or the record is rejected; **PDL-VARINT** (Bitcoin-CompactSize form) for every length and non-fixed-width integer: 1 octet for 0-252, marker octet 253/254/255 plus 2/4/8 big-endian trailing octets for wider values, non-minimal encodings rejected; sorted vectors compared by unsigned byte-lexicographic order over the complete encoded octets (`bytes.Compare` semantics), key/value pairs sorted by key octets only; exactly one repeated-value form (a count-prefixed sequence, never packed-versus-unpacked); no floats anywhere; no maps, ever (every keyed association is a sorted vector).

**Rationale.** A prior interop-divergence review showed that naming "six closure properties" without a byte grammar still leaves tag width, length-of-length, and the sort comparator free, each a demonstrated real-world divergence source. Naming each parameter exactly closes the grammar with zero implementer choice remaining. Splitting prefix from ledger encoding is what lets the prefix stay zero-parse (a `SegmentTableSlot` is read by fixed arithmetic, never by walking tags) while the ledger stays self-describing enough to carry EXT_ENVELOPE's forward-compatibility contract.

**Rejected.** Protocol buffers (the origin sketch's proposal): proto3's own maintainers document deterministic serialization as stable within one binary only, never canonical across languages; non-minimal varints are legal and must be accepted; packed and unpacked repeated fields are both legal for one logical field; field order is unspecified; nested submessages are length-prefixed containment, so a leaf edit rewrites every ancestor length prefix (a direct HC-004 kill); a shared `.proto` compiled by both conformance implementations is shared source code, defeating the CP-003 two-implementation gate it would be adopted to satisfy. Deterministic CBOR (RFC 8949 core profile): leaves float encoding, indefinite-length items, and the tag registry open, and stays self-describing (a text string and a tagged byte string can express one logical value two ways), which is exactly the second-representation surface CON-005 forbids. ASN.1 DER: genuinely canonical, but needs a schema compiler whose output CP-010 requires be reviewable in-repo, and its variable-width length encoding defeats fixed-offset in-place patching. FlatBuffers / Cap'n Proto: vtable sharing and unspecified padding admit multiple layouts for one logical value, and both put offsets inside the data, which is hazardous under HC-010's ban on persisted counted positions. Applying PDL-TLV's own tag machinery to the prefix itself: rejected because it would reintroduce a parse step exactly where HC-006/HC-007 require zero-parse fixed-offset access.

### DP-003: Identity token construction

**Decision.** Fixed-width 128-bit CSPRNG token minted per contiguous typing burst (`run_id`, via `crypto/rand`); character identity = `(run_id, base_ordinal+i)`. Unchanged from the prior pass.

**Rationale.** Unaffected by this repair round's findings. The run-internal-ordinal-versus-counted-position ambiguity it depends on is closed by the 7-tag A-FIELD-ROLE vocabulary in the Identity layer: `run_id` and `base_ordinal` are both tagged IDENTITY-COMPONENT, usable for comparison and lookup but never for positional arithmetic measured from a sequence start. Collision bound: P < 2^-60 below 2^34 mints per lineage, approximately 1.5e-23 at FR-023's stated 10^8-mints/1,000-replica scale.

**Rejected.** Per-character explicit 128-bit tokens (16x the text at Latin-content scale, blowing NFR-032's 2.0x and NFR-033's 1.5x ratios by an order of magnitude, with no run-merging path to recover the cost). Hash-derived identity (digest of nonce, parent identity, ordinal): the ordinal component's interpretation would depend on the parent's current children, which is exactly the counted-position construction TN-003's discriminating rule excludes. Actor-identifier-plus-counter (the conventional CRDT construction): FR-023 bans actor-derived values outright, and FR-080's identity strip on publish would delete the actor component and detach every anchor built on it. Dense-space position identifiers (Logoot/LSEQ-style): these are still ordered, counted positions in a shared coordinate space, banned by CON-001 regardless of the base being dense rather than integer, and published results show interleaving of concurrently-typed passages even in refined dense-identifier families, failing FR-093.

### DP-004: Annotation endpoint representation

**Decision.** `(run_id, birth_ordinal, side_bit, boundary_behaviour)` per endpoint; orphan carriage (author, quoted text, nearest surviving preceding and following identifiers) stored on the annotation at the moment of orphaning, not recomputed at read time. Unchanged from the prior pass.

**Rationale.** Recomputing orphan carriage at read time is unavailable once the neighboring units are themselves later deleted, so the record must be captured once, at orphaning, and carried forward. This is now the explicit disposition target for R3 DELETE-DOMINATES (Section 2, DP-015) whenever an annotation's anchor range is concurrently deleted.

**Rejected.** A separate within-run-index form (adds a second identity representation for the same character, which the A-FIELD-ROLE audit and CON-005 both disfavor). First-class interstitial (gap) identities as the anchor target (adds a new identity class solely to express "between two characters," when a side bit on an existing character identity expresses the same thing more cheaply). Read-time-recomputed orphan carriage (unavailable once neighbors are gone, as above).

### DP-005: Canonical form and equivalence class

**Decision.** Canonical form = a deterministic depth-first serialization of state, streamed without ever being materialized as a file. The at-rest equivalence class is any file whose replay yields that state; writer placement is `place(prior-or-BOTTOM, edit-or-state)` with zero discretion (Section 1).

**Rationale.** This is the direct implementation of CQ-003 option A. Unaffected by this repair round, which targeted encodings and algorithms downstream of the definition, not the definition itself.

**Rejected.** Canonical form as the concatenation of canonical unit octets with the index recomputed (reintroduces a full-index rewrite per edit). Canonical form as a normal form over the append log with tombstones elided (conflates "canonical" with "compacted," which breaks the no-op-save-is-zero-writes property). An equivalence class defined as bare octet equality (collapses back to CQ-003 option B, rejected in `clarify.md` because compaction becomes a whole-file rewrite that must run before every publish or signature).

### DP-006: Digest and tree structure

**Decision.** Two structures, not one. T_S: storage-integrity tree, arity-16, depth-4, over the 16,384-slot segment array; signature-independent, used by unsigned documents and by bounded-prefix consumers. T_C: content-commitment tree, content-subtree-aligned, per-subtree salted, domain tags 0x02 (redactable leaf) / 0x07 (non-redactable leaf) / 0x08 (internal node). Every internal node in either tree always hashes exactly 16 child-digest slots; absent children are filled with `ABSENT_CHILD_DIGEST = SHA-256(0x00)`, using reserved domain tag 0x00, a preimage no real digest ever starts with (real preimages start 0x01, 0x02, 0x07, or 0x08). T_S is fully decoupled from `signed_object`: T_C_root is signed directly, so T_S's root is free to change under partial compaction without touching any signature.

**Rationale.** A fixed arity without a stated filler value lets two implementations hash different preimages for the identical logical tree whenever a node has fewer than 16 real children, which is the common case, breaking every downstream digest. A single reserved constant closes this at zero marginal cost. Decoupling T_S from `signed_object` is what makes the coverage-descriptor's placement-insensitivity (DP-013) sound: compacting uncovered segments no longer needs to preserve T_S's root, because nothing signed depends on it.

**Rejected.** Hashing only the actual children with a count field (leaves "is the count itself inside the hashed preimage" ambiguous, its own divergence source). Zero-length omission of absent children (produces a shorter, non-uniform preimage per node, a second divergence source). A single tree serving both storage integrity and content commitment (forces storage placement to follow the content tree's shape, so a content-tree restructuring would move physical extents, breaking FR-058's content-independent placement).

### DP-007: Signature scheme and verification procedure

**Decision.** Ed25519 (RFC 8032) remains the sole v1 allowlist entry, verified by a closed 7-step wrapper procedure, **EdDSA-Protodoc-1**:

1. Decode the public key `A`.
2. Reject `A` if it is a non-canonical point encoding or lies in the fixed 8-element low-order-point table (curve25519's order-8 torsion subgroup).
3. Decode the signature component `R`.
4. Reject `R` under the identical canonical-encoding and small-order checks used for `A`.
5. Reject the scalar `S` if `S >= L`, where `L = 2^252 + 27742317777372353535851937790883648493`.
6. Compute the challenge `k = SHA-512(R || A || msg) mod L` (`crypto/sha512`, stdlib).
7. Delegate the actual curve-arithmetic verification equation to unmodified `crypto/ed25519.Verify` (stdlib), which is now only ever invoked on inputs where Go's internal behavior cannot diverge from strict RFC 8032, because every input reaching it has already passed the canonical and small-order filters above.

Steps 1-5 are plain integer or byte comparisons against fixed constants (`math/big` or direct byte compare), costing roughly 60-100 new lines, entirely stdlib.

**Rationale.** Verified directly against the installed go1.25.1 toolchain: `go doc crypto/ed25519` shows `VerifyWithOptions`'s `Options` struct exposes only `{Hash, Context}`, no lever for choosing cofactored versus cofactorless verification, no lever to accept or reject non-canonical encodings, no lever for small-order key handling. A prior label of "ZIP-215-style cofactored, but strictly rejecting non-canonical and small-order inputs" is self-contradictory: ZIP-215's entire point is to accept some of exactly what "strictly rejecting" would reject. Pre-filtering the documented divergence points with cheap, non-curve-arithmetic checks, then delegating the uncontested remainder to stdlib, satisfies CP-010 (stdlib-only in core paths) exactly and gives two independent implementations byte-identical behavior on every input, because both filter on the same fixed constants before either touches curve arithmetic.

**Rejected.** ZIP-215-style cofactored acceptance of non-canonical or small-order inputs (self-contradicts the format's own mandatory-rejection language, and is the wrong posture for document integrity, which wants strict rejection, not consensus-permissiveness). Hand-rolled full curve arithmetic in reviewable Go (closes the same gap at roughly 400-600 lines against 60-100, with no correctness benefit once the pre-filter already removes every documented divergence source). RFC 6979 deterministic ECDSA over P-256 or P-384 (deterministic, but `crypto/ecdsa` in the Go standard library signs only via a hedged nonce mixing the `rand` reader; RFC 6979 would need roughly 300-450 lines of new HMAC-DRBG and scalar arithmetic inside the core verify path for zero benefit over Ed25519, which is deterministic by construction and already present). Deterministic RSA (RSASSA-PKCS1-v1_5): deterministic and in the standard library, but signature and key size cost against the format's hard prefix budgets is unjustified next to Ed25519's 64-octet signatures.

### DP-008: History and operation-log storage

**Decision.** Typed HISTORY segments of RLE-batched operations plus a `retention_point` ordinal in the commit ring serve all three CON-022 modes from one representation. Operation ordering is exactly the DP-015 R1/R2/R3 classification, not a bespoke history-specific order. NO_HISTORY mode's interaction with post-migration signature verifiability is disclosed explicitly rather than silently assumed uniform (see Section 7 and Section 9).

**Rationale.** CON-022's own stated rationale (complete retention, full reconstruction, and in-place erasure are pairwise but not jointly satisfiable) already anticipates and justifies the NO_HISTORY degradation case; composing the already-approved FR-062 (unavailable-state verdict) with FR-122 (a signature reports covering the pre-migration state) gives the correct verdict without inventing a new mechanism.

**Rejected.** A state-delta chain (a trim severs every later state's reconstruction path). Full snapshots with structural sharing by content address (workable in principle, but doesn't share a representation cleanly with the trim/no-history modes without a second construct, which CON-005 forbids). A hash-chained op log as the sole structure (trimming the head destroys the tail's verifiability, and actor-keyed version vectors are independently banned by FR-023).

### DP-009: Addressable text segment

**Decision.** Unchanged: an addressable text segment is one identity-run-bearing text block, hard-ceilinged at 65,536 octets (`MAX_TEXT_UNIT_OCTETS`), with an explicit continuation link for logically longer text. This is simultaneously the NFC scope (CQ-009), the extraction-locator scope (FR-042), the run-identity scope (CQ-002), and the redaction-alignment scope (RED-ALIGN: every redactable subtree boundary lands on a block boundary).

**Rationale.** Unaffected by this repair round's findings; the choice was already made to give every one of these four scopes a single, shared denominator rather than four separately-tuned boundaries, which is what CON-005's one-representation principle wants.

**Rejected.** No new alternative was reconsidered in this pass; see `research.md` DP-009 for the original candidate set (paragraph-level container, authored run, whole-block-including-inline-objects).

### DP-010: Rasterization, codecs, and reflow

**Decision, in four parts.**

- **Lossless raster**: one restricted-PNG profile, unchanged.
- **Lossy raster**: **PLP-1** (Protodoc Lossy Profile 1), a from-scratch, exactly-specified, roughly 480-decode-line codec. Fixed 8x8 blocks with edge-replicate padding to a multiple of 8; one of 8 normative fixed quantization matrices (writer chooses which, decoder needs no signaling ambiguity because the choice is carried as a field, not inferred); fixed canonical-Huffman entropy tables (no per-image optimization); DC-DPCM prediction across blocks; an 8x8 integer IDCT specified as an exact integer-matrix-plus-fixed-point-shift-schedule; fixed-point integer YCbCr-to-RGB using BT.601 coefficients as scaled integers with a stated rounding rule; no chroma subsampling in v1 (4:4:4 only); no loop filter, no scalability, no film grain; intra, single-image only. Only decode need be bit-exact (CQ-010's actual requirement); the encoder is writer-discretionary.
- **Rasterizer**: active-edge-list scanline rasterization with exact-rational (`math/big.Rat`, stdlib) signed-area-per-cell coverage accumulation, in the FreeType `smooth` / font-rs lineage, substituting exact rationals for that algorithm's fixed-point approximation. Curves flattened by recursive de Casteljau subdivision to a tolerance of exactly 762 base units (`914400 / 300 / 4`, an exact integer division: 1/4 device pixel at 300 dpi per CON-012), maximum recursion depth 16. Coverage-to-sample by exact-rational-to-integer round-half-to-even. Sample-grid origin at pixel-cell (0,0), no half-pixel offset. Image resampling by an exact-rational area-weighted box filter in both directions.
- **Reflow**: Knuth-Plass optimal-fit line breaking, dynamic programming over candidate breakpoints with an integer, not floating-point, demerits formula, consuming the same in-document break/hyphenation table that fixed pagination already requires; hyphenation points are inserted as additional integer-penalty candidates before the DP runs.

**Rationale.** AV1-intra (the prior pass's choice) is fatal under CP-003: an independently-estimated 12,000-20,000-line bit-exact decoder forces a choice between two implementations sharing a real decoder (defeating source-independence) or two research-grade from-scratch efforts with no credible schedule. A small, fully self-specified codec is the only way to keep CQ-010's "exactly one lossy codec, decode bit-exact by specification" and CP-003's source-independence requirement simultaneously satisfiable. The rasterizer and reflow algorithm were previously named only by technique label with no stated coverage computation, tolerance, filter, or demerits formula; HC-020 and FR-100 require these stated normatively, so each is now specified to the operation, borrowing well-precedented published techniques restated in exact or integer arithmetic rather than inventing a new algorithm (which would itself be an unreviewed research project, exactly what this repair pass exists to eliminate elsewhere).

**Rejected.** AV1-intra-only bare-OBU-sequence: itself now a rejected alternative, killed by the CP-003 source-independence argument above. Baseline JPEG (its IDCT was specified only to an IEEE 1180 statistical tolerance, now withdrawn, so implementations legitimately drift, failing NFR-019's zero-tolerance raster equality). VP8/WebP and JPEG XL (VP8's specification is normative by reference to reference source code, a CP-009/CON-006 violation on its face; JPEG XL specifies parts of decode in floating point). Full-face font embedding (unbounded size, unbounded attack surface from font-instruction interpretation, which CP-005 forbids outright). An invented novel rasterization or line-breaking algorithm (rejected in favor of adopting proven techniques restated exactly, for the reason stated above).

### DP-011: Extension envelope encoding

**Decision.** Unchanged: one universal EXT_ENVELOPE frame kind; canonical order among unrecognized constructs is `(token, position_key)`; the token space is partitioned into registered, owner-scoped, and permanently-retired ranges.

**Rationale.** Unaffected by this repair round's findings.

**Rejected.** See `research.md` DP-011 for the original candidate set (a reserved field-number range inside the structural encoding, hierarchical owner-prefixed tokens, ordering by storage ordinal instead of a declared position key).

### DP-012: Major-version migration mechanism

**Decision.** Unchanged core mechanism: refusal-first two-phase migration emitting `L*(migrate(state))` as a fresh file; header octets `[0,32)` frozen forever. New in this pass: this is the sole point at which RESCIND-AND-RESIGN (DP-017) may occur.

**Rationale.** Migration is already the one lifecycle event with the properties (deterministic, fresh-file, identity-preserving, pre-migration-state-retaining) that a signature-scheme-break recovery needs. Adding RESCIND-AND-RESIGN here reuses those properties instead of inventing a second migration-like mechanism with its own atomicity and retention story.

**Rejected.** Allowing RESCIND-AND-RESIGN at an ordinary edit (would need its own atomicity and retention guarantees instead of reusing migration's already-proven ones). See `research.md` DP-012 for the original candidate set (side-by-side pre/post-migration state retention as two peer states, retired-token retirement recorded only in the registry rather than in-format).

### DP-013: Reconciling the mutable prefix with append-only writing and single-state crash atomicity

**Decision.** The 7-slot self-digesting commit ring is unchanged. What is redefined in this pass is `structure_digest`, the structural half of `signed_object`:

```
structure_digest = H(0x0A
  || header_bytes[0,480)
  || ledger_length(8)
  || frontmatter_metadata_bytes
  || covered_prefix_regions_bitmask(1)   -- over {HEADER, RING_WINNER, FRONTMATTER, SEGMENT_TABLE}
  || sorted-vector-of(covered_ordinal, type, length, digest)
       -- for every ordinal the coverage_descriptor claims covered,
       -- digests recomputed FRESH from the file's actual current octets
       -- at verify time, NEVER read from a stored slot)
```

**Rationale.** The original "segment-table shape" preimage was never pinned to an exact byte range, and read as the raw table bytes it would make `structure_digest` sensitive to every uncovered segment's physical offset, directly blocking any compaction under a present signature (the segment-ceiling bricking risk in Section 8) and leaving an interop-divergence gap over what "shape" means. Redefining it as a covered-ordinals-only, freshly-recomputed projection closes both problems: it is an unambiguous, concrete byte recipe, and it is placement-insensitive, which is exactly what makes DP-016's partial compaction safe. Requiring fresh recomputation rather than trusting a stored digest is what makes a forged inventory detectable (TR-009): a tampered slot and a tampered stored digest can agree with each other, but they cannot also agree with the segment's actual current octets.

**Rejected.** Raw segment-table bytes as the preimage (ambiguous and placement-sensitive). Hashing stored digests rather than recomputing fresh at verify time (lets a tampered stored digest and a tampered slot agree with each other without the verifier ever touching real segment octets, weakening the guarantee to no better than trusting T_S). A front slot carrying only a pointer to a back-placed current inventory (abandons TR-006/007/008's zero-decode determination and costs an extra dependent read against FR-055). A single non-double-buffered superblock rewritten in place with a torn-write-detecting digest (still needs the redo-record-and-replay machinery a 7-slot ring avoids entirely by making the loser simply not win).

### DP-014: Validation pipeline ordering

**Decision.** Unchanged: two-pass validation per extent (declared-length arithmetic against actual available octets, then materialization into a caller-supplied bounded arena), with FR-109's 5-edge-kind cycle detection running before any structural ceiling is evaluated. The full 13-step order is in `data-model.md` Section 7 and repeated in Section 3 of this plan.

**Rationale.** Unaffected by this repair round's findings.

**Rejected.** Declared-length checking at the framing layer only, with ceilings supplied later as a separate pass (risks a cycle resolving to "present, in-bounds" units and generating unbounded work before the ceiling check ever runs, which is exactly the attack FR-109 exists to close).

### DP-015: Concurrent-operation merge classification

**Decision.** An exhaustive 4-way classification, not a per-operation-kind-pair table:

- **(0) DISJOINT-COMMUTE**: non-overlapping targets. Apply both, in either order.
- **(1) R1 SEQUENCE-ORDER**: any operation claiming a position in a sequence, text insertion, table-row insertion, table-column insertion, extension-envelope insertion, annotation creation. Resolved by Fugue's actually-proven non-interleaving total order, because each of these literally is a sequence insertion, not an analogy to one.
- **(2) R2 TOTAL-ORDER-TIEBREAK**: any two operations setting the same single-valued field of the same unit, property set, rename, a non-cycle concurrent move, resize. Resolved by comparing `state_id` as unsigned big-endian octets, smaller wins, the same rule already used for move-cycles, now applied uniformly.
- **(3) R3 DELETE-DOMINATES**: a delete concurrent with anything else targeting the same unit. The delete applies; the other operation's disposition is read from a fixed per-kind table (no-op, or DP-004's orphan-carry for an annotation anchored into a deleted range).

Every operation kind in the data model is, by construction, exactly one of a position-claim, a value-claim, or a delete, so this 3-primitive-plus-default partition is exhaustive over kind, not a table that must be checked for completeness pair by pair.

**Rationale.** Real Fugue's non-interleaving proof is established specifically for text-sequence insertion; "generalized to every operation kind" was previously asserted, not derived, and HC-026 demands zero implementation-defined cells across every pair among a dozen-plus operation kinds, tens of pairs. The fix is not a bigger table, it is a coarser, exhaustive partition of what an operation IS, so classifying by kind closes every cell without enumerating pairs, and each of the three non-trivial primitives already has a justified resolution rule: R1 is proven for its actual case (sequence insertion), R2 and R3 are simple, total, and already used elsewhere in the design.

**Rejected.** Literal per-pair enumeration across a dozen-plus operation kinds (unnecessary once the 3-way kind classification is exhaustive, and itself an error-prone documentation burden to keep complete). Adopting a different published multi-type CRDT wholesale (none of the reviewed literature proves non-interleaving for this format's specific mixed kind-set, structural moves, table axes, extension envelopes, so adopting one would only relocate the same "proof does not cover this case" gap).

### DP-016: Partial compaction

**Decision.** A new, explicit operation, permitted even while a signature is present, restricted to segments and ordinals named in NO signature's `covered_segment_ranges`. It may relocate, coalesce, and reclaim uncovered slots; covered segments' ordinals, offsets, lengths, and digests are untouched, byte-identical. Like full compaction, it is explicit-only, never invoked by an ordinary save.

**Rationale.** NFR-004's unconditional "refuse compaction whenever any signature is present" rule, combined with a fixed 16,384-slot segment table and continuous autosave, was shown to brick an actively-edited signed document in roughly 58 days regardless of how much of the document a signature actually covers. Since DP-013's redefined `structure_digest` is placement-insensitive for uncovered segments, reclaiming them cannot invalidate any signature, so the refusal only ever needed to be as broad as "do not touch covered octets," not "do not compact at all."

**Rejected.** Scaling `MAX_SEGMENTS` with expected document lifetime (converts a mechanism gap into a guessing game about lifetime with no principled number). Leaving NFR-004's blanket refusal unchanged and accepting the bricking as a pure operational-mitigation residual risk (used instead, honestly, as the residual framing for the one case, TOTAL-mode coverage, where reclamation genuinely cannot help; see Section 8).

### DP-017: Longevity mechanisms (RESCIND-AND-RESIGN and Registry Excerpt)

**Decision, two additive mechanisms.**

- **RESCIND-AND-RESIGN**: available only at a major-version migration boundary. Adds a signature under a NEW allowlist scheme over the migrated `T_C_root` while the OLD signature and its metadata remain retained, unmodified, in the already-mandatory pre-migration state. Gives an auditable two-scheme custody chain if Ed25519/Curve25519 itself, not just a hash, is ever broken.
- **Registry Excerpt**: mandatory for any document declaring the durable-profile claim (FR-011). Self-hosts the specific extension-token, `shaping_profile_id`, and `unicode_version_id` registry entries the document actually references, so the document's own IDs remain meaningful even if the central registry later disappears. Non-durable-profile documents remain dependent on the external registry, an accepted, scoped tradeoff.

**Rationale.** CON-016's existing re-protection layer wraps the ORIGINAL signature with a stronger hash and a fresh time attestation, but never touches the signing act itself, so it provides zero protection once the underlying scheme, not just a hash, breaks, a real, foreseeable event inside the format's stated decade-plus horizon. Separately, the format's self-certifying-header philosophy (HC-006, HC-015: never trust an external authority to interpret the file) was being undermined by making the meaning of in-header IDs depend on an external registry with no stated succession plan; embedding only the specific entries a given document actually uses is a bounded, self-contained fix consistent with that philosophy.

**Rejected.** Reserving unpopulated allowlist slots for a hypothetical post-quantum scheme with no defined re-signing operation (insufficient alone, since CQ-004/HC-016 already let a future minor version add allowlist entries; what was missing was the OPERATION connecting an old signature to a new one). Publishing the full external registry inside every document (unbounded and growing; embedding only the referenced entries keeps the mechanism small and content-addressed like everything else in the format). Mandating Registry Excerpts for every document regardless of profile (unnecessary overhead for ephemeral, non-durable documents).

---

## 3. Data flow

Every operation below states what is touched, in what order, and its cost. "Read" costs assume a document up to 1,073,741,824 octets, the ceiling most NFRs cite.

### Open

1. Read the fixed prefix, offsets `[0, 1048576)`. One sequential read of exactly 1,048,576 octets, satisfying TR-006/007/008 with no scan.
2. Verify `header_digest` over `header_bytes[0,480)`. Reject unrecognized magic; reject `format_major` greater than supported before evaluating any other field (FR-123, no dispositions applied).
3. Reject `capability_required > capability_written` naming both values (FR-010).
4. Select the CommitRing winner: highest `sequence` whose `record_digest` verifies AND `ledger_length <= actual file length`. A tie at the highest valid sequence across two or more slots is a structural reject naming both slot indices (PD-RING-001), not resolved by slot position.
5. Reject if the winner's `ledger_length` exceeds the actual file length (truncation/rollback check).
6. Bounds-safe arithmetic check of every live `SegmentTableSlot` (`frame_count*48+32 <= slot_length-64`) without decoding any frame.
7. Recompute T_S over live `SegmentTableSlot.segment_digest` values; compare to the winner's `ledger_root`.

Cost: 1 read of 1,048,576 octets, O(segment_count) hashing bounded by `MAX_SEGMENTS = 16384`, zero ledger content touched. This is the full extent of "opening" a document for a bounded-prefix consumer (TR-006/007/008's audience); an editor session additionally loads content lazily per the Read operation below as the user navigates.

### Read (random single-unit access)

1. Read 1 is the already-resident prefix: the winning ring slot's `index-route` gives, for a target `unit_id`, the `SegmentTableSlot` ordinal of the UnitIndexLeaf segment covering that unit's bucket (`unit_id[0] >> 4`).
2. Read 2 is that UnitIndexLeaf segment (RESOURCE-typed, one frame, a sorted vector of `(unit_id, absolute_offset, length, kind, digest)` tuples); binary search by `unit_id`.
3. Read 3 is the unit itself at the leaf's recorded absolute offset and length; its digest is checked against the leaf's recorded value before any decode.

Total: 3 dependent reads (FR-055), no more than `max(1048576, 1% of file size)` octets beyond the unit itself.

### Extraction (full-document text view)

Walks CONTENT-typed segments only, in storage order, decoding TextBlock and Run frames via a PDL-TLV field walk. Never opens a RESOURCE, HISTORY, or ATTEST segment. Emits text in reading order, one structural locator `(unit identity, scalar position)` per text unit, before any later unit's octets are read (streaming, abandonable at any unit). No index, no integrity structure, no crypto allowlist needed; this is what keeps the extraction role at TR-011's under-1000-line, no-font-no-graphics-no-crypto budget and NFR-012's 15%-of-file read bound.

### Edit

- No-op: `place(prior, no-op) = prior`. Zero writes (NFR-003).
- Content edit changing K octets: append exactly one new sealed CONTENT segment (the touched TextBlock frame, plus any newly split or merged Run records, capped at 65,536 octets per DP-009), then patch (a) the touched `SegmentTableSlot` entries, 48 octets each, (b) the next CommitRingRecord slot (round-robin by `sequence mod 7`, 512 octets), (c) `index-route` and the UnitIndexLeaf segment if unit-index membership changed. Every previously sealed segment is byte-identical before and after (FR-056).

Cost: at most `8K + 262144` octets total write for a K-octet logical edit (NFR-008); combined index-plus-integrity delta at most 262,144 octets (FR-057).

### Save

The atomic commit of an edit: write the new CommitRingRecord into its ring slot. Atomicity of the underlying write is not required, only detectability of a torn write: a reader re-checks `record_digest` and `ledger_length` on open, so a crash mid-write simply leaves the previous winner in place (HC-027, FR-117). No sidecar journal, no temp file, no rename.

### Compact (full)

`L*(state)`: explicit-only, whole-document re-emission via canonical depth-first traversal of T_C, recomputing every prefix field from scratch. Refused outright whenever any signature is present (NFR-004). Cost scales with total document size, which is why it is explicit and rare, never invoked by an ordinary save.

### Compact (partial, new in this pass)

Equivalent to `place(prior, no-op)` for every ordinal named in ANY present signature's `covered_segment_ranges`; ordinals named in NO signature's coverage may be relocated, coalesced, or reclaimed. Explicit-only; permitted under a SUBSET-mode signature (never useful under TOTAL-mode, since nothing is uncovered). Cost proportional to the uncovered churn being reclaimed, not to the whole document.

### Sign

1. Writer selects coverage mode, TOTAL or SUBSET, naming `covered_segment_ranges` and `uncovered_segment_ranges` as sorted, merge-adjacent-canonicalized half-open ordinal-pair vectors.
2. Compute T_C over the covered content subtrees in canonical depth-first order (leaf tags 0x02 salted-redactable or 0x07 non-redactable; internal tag 0x08; always 16 child slots, absent filled with `ABSENT_CHILD_DIGEST`).
3. Compute `structure_digest` per DP-013's recipe, over the current, actual octets of exactly the covered ordinals.
4. Compute `presentation_artefact_digest` over the pinned PresentationArtefact (profile version, page geometry, and the 7 recorded font values per referenced font).
5. Compute `coverage_descriptor_digest` over the CoverageDescriptor's canonical wire form.
6. `signed_object = H(0x04 || T_C_root || structure_digest || presentation_artefact_digest || coverage_descriptor_digest)`.
7. Sign `signed_object` with Ed25519 (deterministic; Go's `Sign` API takes no randomness parameter) to produce a 64-octet `signature_value`.
8. Append a new ATTEST-typed segment holding the signature record. ATTEST segments are categorically excluded from every signature's own and every other co-existing signature's coverable set, closing the FR-002 self-coverage circularity by construction.

Cost: O(content_unit_count) for the T_C tree-shape walk (arity-16, depth at most 5), plus O(covered content octets) to canonicalize covered subtrees; uncovered subtrees contribute only their previously-computed leaf digests, not their content octets.

### Verify

Runs `data-model.md`'s 13-step validation pipeline (repeated at the level this plan needs): header and version checks, capability arithmetic, ring winner selection, truncation check, bounded-prefix structural checks, T_S recomputation, segment-type and coverage well-formedness, 5-edge-kind cycle detection (structural moves, extension-envelope fallback references, annotation anchors, run split/merge lineage, RescindResignRecord chains) before any structural ceiling, structural ceilings, T_C recomputation and the 7-step EdDSA-Protodoc-1 procedure, NFC and identity-role checks, RegistryExcerpt completeness for durable-profile documents, extension-envelope reachability and disposition resolution. `T_C_root`, `structure_digest`, `presentation_artefact_digest`, and `coverage_descriptor_digest` are always recomputed fresh from the current file; none are ever trusted from a stored or cached slot (the Frontmatter `fm-coverage-summary` field is an explicitly non-authoritative cache). Verdicts: valid, attested-with-declared-omissions, unverified, covering-an-unavailable-state. Cost: O(covered content) plus O(segment_count) for T_S, both bounded by the ceiling table.

### Redact

1. Identify the designated redactable subtree(s) (T_C leaf tag 0x02, with a 32-octet salt stored inside the subtree it commits).
2. Remove the subtree's content octets and its salt together; removing one without the other leaves the commitment openable.
3. The retained T_C leaf is now the bare 32-octet commitment alone. Search cost to recover the removed content is 2^256, far above FR-075's 2^80 floor.
4. Verification reconstructs `T_C_root` using the retained commitment in place of `canon(subtree)` for every declared-omitted subtree, and re-canonicalizes actual octets for everything else; if the reconstruction matches the originally-signed root, the verdict is attested-with-declared-omissions, enumerating every omission. The original signature is never re-signed.

Cost: proportional to the content removed, similar in shape to a partial-compaction reclaim.

### Publish

`L*(published_state)`: a full re-emission containing no octet of any removed, superseded, or erased unit, verified by octet-level search including layout advances, cached renderings, index entries, and preview payloads (FR-078); no actor-identity-carrying field values (FR-080); every custody and fixity value (signatures, time attestations) carried from the input unchanged. Cost: O(retained content), a whole-artifact re-emission, explicit and infrequent by design, same cost class as full compaction.

### Merge

1. Locate the common ancestor from the two files' own ancestry chains (`parent_state_id` in the commit ring, `causal_predecessor_state_id` in HistorySegment records), no external ancestor store needed.
2. Classify every pair of concurrent operations per DP-015: DISJOINT-COMMUTE, R1 SEQUENCE-ORDER, R2 TOTAL-ORDER-TIEBREAK, or R3 DELETE-DOMINATES.
3. On a genuine conflict (neither side's disposition is decidable by the classification, e.g. FR-116's non-ancestor supersession case), exit non-zero naming both values; never select a side, concatenate, or interleave (HC-026).
4. Refuse a merge across a retention point (CON-024) or between branches declaring different history modes (CON-025), naming the point or both modes.

Cost: O(operation count in both traces since the common ancestor); no whole-document rewrite is required unless the merge's own commit triggers the ordinary Edit write path above.

### Migrate

1. Refusal-first: enumerate every unrepresentable construct and its location before writing any output octet (FR-121). A single unrepresentable construct halts migration entirely; there is no approximation.
2. If clean, emit `L*(migrate(state))` as a FRESH file; every content-unit identifier (`run_id`, `unit_id`) passes through unchanged (FR-120).
3. Pre-migration signatures are retained and reported as covering the pre-migration state (FR-122), never the migrated content.
4. At a major-version boundary only, a custodian may additionally invoke RESCIND-AND-RESIGN (DP-017).

Cost: whole-document re-emission, same class as compact/publish, explicit and infrequent (only at a major-version bump).

---

## 4. Interfaces

### Package boundaries (Go module `Protodoc`, go 1.25)

Dependency direction runs top-to-bottom; a package never imports one below it in this list. Packages marked "core" fall under CP-010's standard-library-only constraint.

| Package | Responsibility | Depends on | Core (stdlib-only)? |
|---|---|---|---|
| `pdlfmt` | PDL-VARINT encode/decode, PDL-TLV field reader/writer, the sorted-vector byte-lexicographic comparator | stdlib only (`encoding/binary`, `bytes`) | yes |
| `container` | Prefix structs (Header, CommitRingRecord, SegmentTableSlot, Frontmatter) as fixed-layout Go structs with named offset constants; ring-winner selection; open/commit primitives | `pdlfmt`, `crypto/sha256` | yes |
| `ledger` | Segment framing (header, frames, frame directory), the `place()` function (no-op, edit, compact, partial-compact), the append-only writer | `container`, `pdlfmt` | yes |
| `content` | TextBlock, Run, Annotation/Range; the run-merge predicate; hand-written NFC quick-check (no `golang.org/x/text`, per CP-010); identity minting via `crypto/rand` | `ledger` | yes |
| `integrity` | T_S and T_C construction and verification; CoverageDescriptor; `structure_digest`; Signature construction and verification (the EdDSA-Protodoc-1 wrapper over `crypto/ed25519`); RedactionCommitment; PresentationArtefact; RescindResignRecord | `container`, `ledger`, `content` | yes |
| `history` | HistorySegment (RLE op batches classified via DP-015), ErasureRecord, retention-point trim | `content`, `integrity` (for the R2 `state_id` comparator) | yes |
| `merge` | The DP-015 4-way classifier, the R1 Fugue-lineage non-interleaving order, ancestor discovery | `history`, `content` | yes |
| `extract` | The extraction view: reading-order text walk over CONTENT segments only, locator emission, streaming and abandonable | `container`, `pdlfmt`, `content` (TextBlock/Run decode only) | yes; explicitly excludes `integrity`, `render`, `merge` (TR-011) |
| `validate` | The full 13-step validation pipeline; the generated ceiling-table constants with a CI equality check against spec text; 5-edge-kind cycle detection | `container`, `ledger`, `content`, `integrity` | yes |
| `render` | PresentationArtefact resolution, the PLP-1 codec (bit-exact decode, discretionary encode), the restricted-PNG lossless codec, the exact-rational rasterizer, Knuth-Plass reflow, fixed pagination | `validate`, `content`, `integrity` | outside CP-010's five core paths; still no third-party dependency without a recorded owner exception |
| `migrate` | Refusal-first two-phase major-version migration; RESCIND-AND-RESIGN | `validate`, `integrity` | yes |
| `registry` | RegistryExcerpt embedding and validation; extension-token space partitioning; EXT_ENVELOPE disposition resolution | `container`, `ledger` | yes |
| `canon` | The canonical depth-first traversal (`L*`) used by compact, publish, migrate, and fresh writes. The only package that produces a full re-serialization | everything above except `render`, `extract`, `merge` | yes |
| `cmd/protodoc` | The CLI binary | all of the above | n/a (composition root) |

### Selected type-and-name-level signatures

`pdlfmt`:
- `EncodeVarint(n uint64) []byte`; `DecodeVarint(r io.Reader) (value uint64, consumed int, err error)`, rejecting non-minimal forms.
- `type FieldReader` with `Next() (tag uint8, value []byte, ok bool, err error)`, enforcing strictly-ascending tags.
- `CompareSortedVectorElement(a, b []byte) int`, byte-lexicographic.

`container`:
- `type Header struct { ... }`; `ReadHeader(prefix []byte) (Header, error)`.
- `type CommitRingRecord struct { ... }`; `SelectWinner(ring [7]CommitRingRecord, fileLength uint64) (winner CommitRingRecord, index int, err error)`, returning a distinct tie error type for PD-RING-001.
- `type SegmentTableSlot struct { ... }`; `ReadSegmentTable(prefix []byte) ([16384]SegmentTableSlot, error)`.

`ledger`:
- `type Segment struct { Header SegmentHeader; Frames []Frame; Directory []FrameDirEntry }`.
- `type Operation interface{}` with concrete cases `NoOp`, `Edit{...}`, `Compact`, `PartialCompact`.
- `Place(prior *File, op Operation) (next *File, writes WriteSet, err error)`.

`integrity`:
- `ComputeTS(slots [16384]SegmentTableSlot) (root [32]byte)`.
- `ComputeTC(units []ContentUnit) (root [32]byte, tree *TCTree)`.
- `StructureDigest(headerPrefix480 []byte, ledgerLength uint64, frontmatterMeta []byte, bitmask uint8, covered []CoveredEntry) [32]byte`.
- `Sign(priv ed25519.PrivateKey, signedObject [32]byte, coverage CoverageDescriptor, presentation PresentationArtefact) (Signature, error)`.
- `Verify(doc *Document, sig Signature) (Verdict, error)`; `type Verdict int` enumerating `Valid`, `AttestedWithDeclaredOmissions`, `Unverified`, `CoveringUnavailableState`.
- `VerifyEdDSAProtodoc1(pub ed25519.PublicKey, sig [64]byte, msg []byte) error`, the 7-step wrapper.

`merge`:
- `Classify(a, b Operation) MergeClass`; `type MergeClass int` enumerating `DisjointCommute`, `R1SequenceOrder`, `R2TotalOrderTiebreak`, `R3DeleteDominates`.
- `Merge(a, b *Document) (*Document, []Conflict, error)`.

`extract`:
- `NewExtractor(doc *Document) *Extractor`.
- `(e *Extractor) Next() (unit Unit, locator Locator, ok bool, err error)`, streaming and abandonable.

`validate`:
- `Validate(raw []byte) (Report, error)`, running the 13-step pipeline; `Report` carries the verdict category plus any `PD-*` rule identifier.

`render`:
- `ResolvePresentation(doc *Document, sig *Signature) (*PresentationArtefact, error)`.
- `RasterizePage(doc *Document, pageUnit ContentUnitID, dpi int) (*Raster, error)`.
- `DecodePLP1(payload []byte) (*Raster, error)`, bit-exact.
- `Reflow(doc *Document, viewportWidthEMU int64) (*Layout, error)`, Knuth-Plass.

`migrate`:
- `Migrate(doc *Document, targetMajor uint16) (*Document, error)`, refusal-first; returns a `MigrationRefusal` error type naming the construct and its location on any unrepresentable construct.
- `RescindAndResign(migrated *Document, prior Signature, newScheme ParamSet, signer crypto.Signer) (RescindResignRecord, error)`.

`registry`:
- `ValidateRegistryExcerpt(doc *Document) error`.
- `ResolveDisposition(env ExtEnvelope, readerGeneration uint16) (active Payload, err error)`.

`canon`:
- `Canonicalize(state *Document, w io.Writer) error`, streaming `C(S)` without materializing it.
- `Compact(doc *Document) (*File, error)`; `Publish(doc *Document, decisions PublishDecisions) (*File, error)`.

### CLI verb surface (`cmd/protodoc`), 11 verbs (TR-012)

| Verb | Purpose |
|---|---|
| `validate <file>` | Structural pipeline only (steps 1-9 of the 13-step order); exits with a `PD-*` code |
| `inspect <file>` | Prints the bounded-prefix inventory (TR-006/007/008) with zero decode |
| `extract <file> [--from-unit=ID] [--limit=N]` | Streaming text and locators |
| `verify <file> [--sig-index=N]` | Full verification; prints the `Verdict` enum |
| `diff <fileA> <fileB>` | Per-content-unit differing-unit enumeration |
| `merge <fileA> <fileB> -o <out>` | DP-015 classifier; nonzero exit and named conflict on a real conflict |
| `project <file> -o <text>` | The deterministic text projection (TR-004/005), non-normative, round-trips, never accepted back as input |
| `redact <file> --unit=<id>... -o <out>` | Subtree removal and salt destruction |
| `publish <file> -o <out>` | `L*(published_state)` |
| `sign <file> --key=<path> --coverage=total\|subset --range=...` | Appends an ATTEST segment |
| `migrate <file> --target-major=N [--rescind-and-resign --new-scheme=...] -o <out>` | Refusal-first two-phase migration; RESCIND-AND-RESIGN is a flag on `migrate`, not a 12th verb, keeping TR-012's count at 11 |

---

## 5. Numeric budget compliance

Every stated budget from `spec.md`, cross-checked against the mechanisms in Sections 1-4. "Achieved" states the actual verdict, including caveats; nothing is rounded up to "satisfied" where the evidence does not support it. Misses are marked **MISS** in bold.

| Requirement | Budget | Achieved | Arithmetic |
|---|---|---|---|
| FR-006/007/008/009/011 | 512 octets yield identity, class, major version, capability generations, durable claim; zero content-derived values | satisfied | Header = magic(8)+major(2)+minor(2)+class(2)+cap_written(2)+cap_required(2)+durable(1)+history_mode(1)+unicode_id(2)+shaping_id(2)+layout_id(8) = 32B core, reserved to 480 (MBZ), header_digest 32B = 512B exact |
| FR-051/053/054 | 262,144 octets: preview bound to a digest, staleness detection, title/pages/dims/language | satisfied for in-window fields; preview raster fidelity to page-1 content is NOT independently verifiable by a bounded-prefix consumer, see Section 6 | Frontmatter ends at exactly 262,144: preview raster <=131,072 + snapshot <=4,096 + metadata <=16,384 + retired-token list <=16,384, well inside the 258,048-octet region |
| TR-006/007/008 | 1,048,576 octets yield type/length/digest per unit, the coverage set, and the no-executable-construct determination, all without decoding | satisfied | SegmentTable (786,432) plus Header (512) plus CommitRing (3,584) plus Frontmatter (258,048) = 1,048,576 exact; inertness = closed 4-value segment-type enum, zero executing kinds |
| NFR-008 | at most 8K+262,144 octets written per K-octet edit, at 1MB/50MB/500MB | satisfied | Worst case: text-block rewrite <=65,536 + 4 T_S nodes (512) + segment-table slot (48) + ring slot (512) + index leaf update (16,384) = 82,992+K <= 262,152+K |
| FR-057 | 262,144-octet combined index+integrity delta per edit | satisfied | 4 T_S nodes (~2,052) + segment-table slot (48) + T_C path (~5,165) + ring slot (512) + index leaf (16,384) + coverage-descriptor delta (worst case ~2,000) = ~26,161, about 10x headroom |
| NFR-009/010 | novel chunks <= max(1,048,576, 8K) at 8KiB average CDC, incl. insert-at-head at 10,000 pages | satisfied | Approximately 19 chunks, ~152,000B at 1/50/500MB; head-insert ~165,000B via the content-identity-keyed page directory, independent of page count |
| NFR-004 | 500 saves produce zero compactions; 100 compaction attempts on signed documents are refused | satisfied for FULL compaction as originally specified; PARTIAL compaction is a new, distinct operation NOT refused merely because a signature is present | Full compaction: `place(BOTTOM,S)`, unconditionally refused whenever any signature is present, exactly as before. Partial compaction: explicit-only, restricted to ordinals outside every present signature's covered range, checked before I/O |
| FR-055 | at most 3 dependent reads, no more than max(1,048,576, 1% of file) beyond the unit | satisfied under the reading that the read-count clause includes the final read of the unit; see Section 9 for the alternate reading | read1 = prefix incl. index-route; read2 = index leaf; read3 = unit at the leaf's recorded offset; 3 reads total |
| FR-056 | 100% of extents with no modified unit byte-identical before/after every operation | satisfied by construction | Sealed segments are never rewritten; the writer's only operations on an existing extent are append elsewhere, a fixed-slot flag flip, or a whole-segment relocation by partial compaction restricted to uncovered ordinals only |
| FR-058 | 1,000 edits incl. renames: ordinal sequence and offsets of unmodified extents unchanged, 0 re-layouts | satisfied | Storage ordinal is a monotonic counter independent of name, digest, or content; partial compaction is excluded from this test's edit corpus since it is a separate, explicit operation |
| NFR-012/013/014/TR-011 | <=15% of file read, <=20s CPU, <=33,554,432B peak, extractor <1,000 source lines, no font/graphics dependency | satisfied | ~72MB / 6.7% at 1GB/10,000 pages; extractor estimated at ~770 lines (header/ring ~80, segment-table walk ~60, PDL-VARINT/TLV decode ~150, text/run/NFC ~200, locator emission ~100, extension pass-through ~80, CLI/IO glue ~100); zero font/shaping/graphics/crypto dependency |
| FR-041/035/042/047/048 | 5,000-doc exact extraction and scalar equality; 10,000-substring locator resolution; streaming/abandonment at N={1,10,100,10000} | satisfied by construction | Pure structural walk over NFC-validated, never-renormalized run text; locator = (block_id, scalar-index) computed at emit time |
| NFR-015/016 | 300ms cold-cache preview to first page image, <=67,108,864B peak | satisfied | 1 sequential 262,144B read plus bounded decode at <=1240x1754; decode is O(pixels) with a small constant, roughly 90ms decode plus 2ms read, well under 300ms |
| NFR-017/018 | <=209,715,200B peak to open and render any page of a 1GB/10,000-page doc; <=8,388,608B read for page N plus its resources | satisfied | Page N cost approximately 93,000B independent of document size (page-directory path 12,288 + text units 65,536 + T_C sibling path ~5,160 + content frames ~10,000), plus referenced resources under MAX_DECODED_UNIT |
| NFR-030/FR-106 | peak resident <=4x input octet length, any input; rejection-path peak <16,777,216B on the bomb corpus | satisfied ONLY under a stated reinterpretation (heap for document data, with a floor exemption below 1,048,577 octets); **NOT satisfied against the literal text**, see Section 9 | No real process validates a 1-octet input within 4 octets of resident memory in any language; validator arena fixed at prefix-size + one bounded window (65,536) + cycle-detection colour array (4,096) is approximately 1.12MB |
| FR-108/CON-007/CON-008 | 10^9 executions: 0 resolved external references, 0 process launches, 0 identifier-equality errors | satisfied by construction | Closed segment-type enum defines zero executing or outward-dereferencing kinds; identifier equality is exact-octet comparison only |
| NFR-019/023/024 | 300dpi tolerance-zero digest equality across 2 implementations; identical rasters with/without font instructions; 100% durable-profile raster equality | satisfied on paper, algorithm fully named; **cross-implementation empirical agreement remains unverified until a second implementation exists** | Rasterizer per DP-010: active-edge signed-area coverage, de Casteljau flattening at 762 base units, depth 16, round-half-to-even, no half-pixel offset, area-weighted exact-rational resampling. No instruction-stream interpreter exists in the format |
| FR-089 | exactly 7 recorded values per font | satisfied | name, version, digest, numeric variation-axis coordinates, required code-point set, required layout-feature set, embedding-permission values |
| CON-012/013 | 1/914,400-inch base unit for every persisted geometric value; 100% identical resolved geometry across 2 implementations | satisfied | No float type anywhere in the schema; proportional split = `floor(total*prefix_i/W) - floor(total*prefix_{i-1}/W)`, exact integer, remainder to the final share |
| CON-014 | exactly 1 colour representation | satisfied | sRGB primaries, D65 white point, exact piecewise transfer function, u16 components, linear-light 32-bit fixed-point compositing, mismatch rejected, never converted |
| FR-098/100 | reflow without 2D scrolling down to 320 reference pixels; identical line-break positions at {320,480,768,1024} across 2 implementations | satisfied, algorithm now named | Knuth-Plass integer-demerits DP over the in-document break/hyphenation table; fixed candidate-enumeration order for ties makes the optimum identical across implementations by construction |
| FR-023/019/020/021/022/024 | 10^8 mints/1,000 replicas 0 collisions; fork-merge and edit trials 0 collisions/losses; delete-create 0 reissue; duplication 0 clones; colliding merges 100% refused | satisfied | 128-bit CSPRNG `run_id`, P(collision) < 2^-60 below 2^34 mints/lineage, approximately 1.5e-23 at the stated scale |
| FR-025/CON-001 | 10,000 randomized non-intersecting edits: identical resolved endpoints in 100%; 0 persisted count-based position constructs | satisfied | `birth_ordinal` assignable only by mint/split; audited under the 7-tag A-FIELD-ROLE vocabulary, which states the discriminating rule (IDENTITY-COMPONENT fields never used for positional arithmetic from a sequence start) |
| FR-026/027 | exactly 4 boundary-behaviour values, declared independently per end, preserved across save/load/merge | satisfied | Closed enum {inside, outside, inside-if-inserted-before, inside-if-inserted-after} |
| FR-028/029/030 | 100 orphaned annotations remain readable, resolve deterministically, retain identities losslessly | satisfied | Orphan carriage captured at orphaning time; also the R3 DELETE-DOMINATES disposition target |
| FR-033 | 500 mixed-direction documents: identical visual order across 2 implementations; 0 in-band directional constructs | satisfied | Direction is a unit-level declared property; PDL-TLV has no field kind capable of expressing an in-band bidi override |
| FR-092..096/116/TR-010/HC-026 | 1,000-trial tests per concurrent-op category: 0 implementation-defined cases across every pair | satisfied via the DP-015 exhaustive classification | Every operation is exhaustively a position-claim, a value-claim, or a delete; the 3-primitive-plus-default partition covers every pair without an N-by-N table |
| FR-075 | omitted redactable content resists exhaustive search over 2^80; 500 published docs, 0 recovered | satisfied | T_C redactable-leaf commitment = `H(0x02 \|\| 32-octet CSPRNG salt \|\| canon(subtree))`; search cost after removal is 2^256 |
| FR-078 | 500-doc redaction corpus, 0 octet-level hits incl. layout advances, caches, index, preview | satisfied | Publish = fresh `L*(published_state)` emission touching only reachable content; orphan `quoted_text` overlapping a redacted subtree is redacted alongside it, see Section 6 for the residual scanning-completeness caveat |
| FR-063/115 | 100 adversarial docs: 0 unqualified valid verdicts; manipulated/partial/retired-parameter docs: 0 signer identities | satisfied via EdDSA-Protodoc-1 | 7-step procedure per DP-007, roughly 60-100 new lines, 100% stdlib, verified against go1.25.1's actual `crypto/ed25519` surface |
| NFR-006 | 1,000 sign-twice trials, identical octets in 100% of cases | satisfied | Ed25519's nonce is derived from `SHA-512(key-prefix, message)`; Go's `Sign` API takes no randomness parameter |
| FR-074/076/077 | 200 signed docs with designated redactable subtrees; identical omission enumerations across 2 implementations; undesignated omissions 100% unverified | satisfied | T_C_root is the signed commitment; recomputing it using retained salted leaves for declared omissions reproduces the exact signed root |
| FR-066/067/068 | signed-then-edited docs: signature/state/presentation survive octet-identically; per-state attestation; zero omission/spurious diffs | satisfied | SIGNATURE segment is ATTEST-typed and untouched by any subsequent append-only edit; partial compaction, restricted to uncovered ordinals, cannot touch it either |
| CON-016 | 200 re-protected docs, all prior verdicts preserved, 0 re-signings | satisfied for hash-family weakening; extended via RESCIND-AND-RESIGN for a scheme break, at a major-version boundary only | SHA3-256 outer digest layer plus a genuine TSA-signed fresh time attestation over the unchanged original signature; RESCIND-AND-RESIGN assembled from existing mechanisms, does not retroactively protect a v1 document never migrated before its scheme breaks |
| FR-119/120/121/122 | canonical octets identical across 2 implementations per version pair; identifiers preserved; unrepresentable construct halts before output; pre-migration signatures report the pre-migration state | satisfied for COMPLETE-HISTORY and RETAINED-FROM-POINT modes; satisfied as a disclosed degraded-verdict case for NO_HISTORY mode | For NO_HISTORY, no prior state, including a just-signed pre-migration one, survives past the next commit by definition of the mode; composing FR-122 with FR-062 gives the correct verdict at the very next commit, not a defect |
| FR-072/073/062 | adversarial-time/compromise/severed-state docs: 100% unverified or unavailable-state, 0 identities | satisfied | Three independent deterministic checks: time-attestation interval, revocation evidence, severed-state enumeration lookup |
| NFR-005 | exactly 4 permitted non-deterministic sites; 20 environments yield exactly 1 distinct digest | satisfied | `canon(S)` is a pure function of an already-fixed state S with no clock, RNG, or filesystem-metadata read in its own definition; the 4 sites (minting, redaction salts, signature values, time attestations) act only at authoring/signing time, never during canonicalisation of a fixed state |
| NFR-032/033 | saved size <=2.0x a no-history equivalent at 250,000 ops/100,000 chars; 10x op-count difference opens within 1.5x time/memory | satisfied for realistic burst-typed traces (approximately 1.1x measured); **MISS for adversarial scattered single-character-edit traces** (no automatic repair path exists, since re-basing runs would detach every anchor built on them, which HC-011 forbids) | Burst-typed ratio ~1.1x; open cost O(content), not O(operation count) |
| CON-022/023/024/025 | exactly 3 history modes, immutable at creation; refusals name mode/retention-point/both-modes | satisfied | One HISTORY-segment-plus-retention_point representation serves all 3 modes; each refusal decidable from two files' headers alone |
| FR-013/107/014 | exactly 3 dispositions; single defined behaviour per disposition; missing disposition 100% rejected naming the token | satisfied | Closed enum {ignore, degrade, refuse}; missing disposition rejected naming the token |
| FR-010/123 | inverted-pair docs 100% rejected naming both values; future-major docs 100% declined naming the version, 0 dispositions applied | satisfied | Both decidable from octets `[0,32)` alone |
| FR-017/018 | future-generation constructs preserve digests before/after save or the save is refused; merges preserve digests or refuse, 0 silent drops | satisfied | Unimplemented constructs are opaque segments the append-only writer never touches |
| FR-090/091 | unknown embedded objects render at declared extent, payloads rewritten octet-identically; opaque payloads expose kind/length/digest with no decode | satisfied | Authored extent is independent of the resource; payload is an untouched opaque segment; kind/length/digest are ordinary SegmentTableSlot fields |
| FR-111..114 | missing/failing/undecodable resource produces a mark at declared extent, non-decorative, with a text alternative; substitution in the render report; pagination marked non-authoritative; no host substitution, no reflow | satisfied | Every resource reference carries its authored rendered extent independent of the resource itself |
| FR-102..105/110/TR-009 | malformed-document rejection at the correct offset; 0 recoveries; 0 precedence selections/renames; forged inventories 100% detected before use | satisfied | Single authoritative inventory (the segment table) means no second structure can disagree with it; a forged inventory is detected because the verifier recomputes `structure_digest` fresh from actual octets, never trusting a stored digest |
| FR-117/TR-010 | writer killed at random points: exactly 1 complete decidable state every time; conditional-replacement writes refuse naming the current holder on mismatch | satisfied | Highest-sequence, self-digest-valid ring slot wins; equal-sequence tie is a distinct rejection; `ledger_length` is part of `structure_digest`'s preimage, so truncation to an earlier boundary invalidates any signature made over a later one via two independent mechanisms |
| CON-003 | 50 non-NFC inputs, 100% rejected, 0 conversions | satisfied | NFC quick-check at validation; the writer refuses rather than renormalizes |
| NFR-001/003 | 100% digest equality across 2 implementations, 2 architectures; 100% octet-identical open-save-compare | satisfied, the specific divergence source (underspecified TLV grammar) is closed | `canon(S)`'s order is fixed to content digest and the R1/R2/R3 classification, never storage ordinal; PDL-TLV's tag width, PDL-VARINT, and the sort comparator are all fixed with zero remaining free parameters |
| NFR-028/CP-011 | 2 implementations, >=200 hostile cases, 0 output/verdict divergences | satisfied for the core paths this repair pass addresses; rendering-role schedule risk tracked separately, not a correctness gap | Every specific divergence source a prior refutation demonstrated (TLV free parameters, tree filler, Ed25519 procedure, merge classification, missing rasterizer/reflow algorithms, AV1 infeasibility) is closed by a named mechanism |
| NFR-026/027 | 5 working days for extracting-and-validating; 30 working days for rendering | extracting/validating-and-verifying satisfied within budget; **rendering role AT RISK**, materially improved but not yet closed | Revised rendering estimate: shaping reproduction (irreducible oracle-matching under CQ-006 option B) approximately 15-25 days; rasterizer approximately 5-10 days; PLP-1 approximately 3-6 days; remaining mechanisms approximately 5-8 days; revised total approximately 28-49 days versus the original 60-200+ day estimate under AV1 |
| NFR-029/031 | 100% statement-to-case coverage as a publish gate; >=80% of accessibility conditions software-decidable | satisfied (process targets) | CI-enforced process property; structural content model is machine-checkable directly from segment/unit structure |
| CON-009/010/011 | exact-decimal ceiling table with a stable identifier; at-limit and over-limit conformance docs with identical verdicts; a local-budget refusal uses a distinct status | satisfied | Ceiling table generated into checked-in Go constants with a CI equality check against the specification text; `PD-BUDGET-*` distinct from every validity verdict |
| CON-017/018/005 | exactly 1 writer conformance class; exactly 3 reader roles; at most 1 normative representation per capability | satisfied | Role boundaries per Section 4's package map; the extension-envelope fallback pair is excluded from the one-representation audit by an explicit, declared selection rule, never left ambiguous |
| TR-004/005 | 100% of the conformance corpus round-trips from the text projection to identical canonical octets; the projection sits outside the conformance surface | satisfied | The text projection is a pure deterministic function of `canon(S)`, never accepted as document input |
| TR-012 | 11 CLI operations at first stable release | satisfied | validate/inspect/extract/verify/diff/merge/project/redact/publish/sign/migrate; RESCIND-AND-RESIGN is a flag on `migrate`, not a 12th verb |
| FR-124 | 1 document per degenerate case, identical verdicts and outputs across 2 implementations | satisfied | An empty-content state is valid under `place(BOTTOM,S)`; `L*` emits a well-defined minimal prefix |
| CP-012 | 90-day disclosure clock; fuzzing oracle triggers above 4x input peak memory | satisfied (process target); fuzzing oracle threshold uses the same reinterpreted 4x bound flagged for NFR-030, not the literal text | Process commitment; see Section 9 for the NFR-030 conflict this inherits |
| CP-010 | extraction implementable in <1,000 lines with no font/graphics dependency; cross-compiles to darwin/linux/windows on amd64/arm64; stdlib-only in the 5 core paths | satisfied | Reference extractor approximately 770 lines; the 5 core paths use only `crypto/sha256`, `crypto/sha3` (present in installed go1.25.1 stdlib), `crypto/ed25519`, `crypto/sha512`, `math/big`, `encoding/binary`, `unicode/utf8`, `crypto/rand`; zero third-party modules, zero cgo |
| CP-002 | per-role normative-statement ceilings computed in CI (no literal number in the constitution; this plan declares the numbers) | satisfied, ceilings newly declared with headroom | Extracting <=150 (current estimate ~65, 57% headroom); validating-and-verifying <=400 cumulative (current estimate ~220, 45% headroom); rendering <=500 cumulative (current estimate ~321, 36% headroom). These are architecture-phase estimates; the CI-enforced literal count requires the tasks/analyze-phase prose to exist |
| Constitution quality bars | >=80% test coverage on business logic; 100% on canonicalisation, digest computation, signature verification, ceiling enforcement, extraction view | satisfied (process target, mechanically reachable) | The 5 named modules (canonicalizer, T_S/T_C digest computation, the EdDSA-Protodoc-1 wrapper, the generated ceiling-table checker, the extractor) are each pure functions over octets, no I/O, no clock |

---

## 6. Threat model

Mandatory per the constitution's quality bars ("threat model is mandatory in plan.md for signature, redaction, publish, and parsing work").

### Assets

- Document content octets, including octets a user has asked to remove (redacted or erased content, whose absence is itself a claim the format must be able to prove).
- Actor-identity-carrying fields (author, device, session), which the format must be able to strip completely at publish (FR-080).
- The signature verdict and any signer identity displayed to a reader; a false positive here (a signer identity shown over unverified or partially-covered content) is the single worst failure mode this format exists to close, per US-003.
- The T_S and T_C digests and the segment-table inventory, which every downstream determination (coverage, type, length, inertness) is derived from.
- Extension payload octets and their disposition metadata.
- The preview raster and Frontmatter metadata, consumed by low-trust, high-volume, automated actors before any full verification occurs.

Signing keys themselves are never a Protodoc asset: the format never stores a private key, only public-key references inside a credential chain.

### Actors

- The document's legitimate author or editor.
- A reviewer or collaborator with edit access to a subset of content.
- A counterparty who receives a signed document and must decide whether to trust it, without necessarily trusting the sender's tooling.
- A compliance officer publishing a redacted copy for external release.
- A malicious document author, who crafts a hostile file to defeat a downstream reader.
- A malicious network or storage intermediary, who tampers with a file already at rest (a forwarded email attachment, a synced file, a shared drive copy).
- A bounded-prefix consumer: a mail gateway, file browser, or thumbnailer, an automated, low-trust actor that reads only the leading 1,048,576 octets and typically runs in a sandbox with no network access (TR-006/007/008's intended audience).
- An independent second implementer (CQ-012), whose interest is divergence from the reference implementation, not malice, but whose findings have the same practical consequence as an attack if the specification under-determines behavior.

### Trust boundaries

1. **File-at-rest versus validated state.** Nothing is trusted until the structural and digest checks in the validation pipeline (Section 3, Verify) pass. A codec is never constructed before its input's digest is checked (CP-006, HC-008).
2. **Bounded-prefix read versus full-file read.** A gateway that reads only 1,048,576 octets can determine type, length, digest, and coverage-hint per TR-006/007/008, but it can never perform a cryptographic verification (verification needs the referenced ATTEST segment and, for TOTAL coverage, potentially the whole covered content set) and it cannot authenticate the preview raster against actual page-1 content, only against a cached digest it cannot independently recompute (see the preview-staleness gap below).
3. **Signed-covered octets versus uncovered octets.** A verifier must never treat content outside a signature's `covered_segment_ranges` as attested, in either TOTAL or SUBSET mode. This boundary is what `structure_digest`'s DP-013 redefinition exists to make mechanically checkable.
4. **Signer versus counterparty.** The classic two-party signature trust boundary: a counterparty's reader must derive its verdict entirely from the file plus the public key, never from the sender's claims about the file.
5. **Reader-generation boundary.** `capability_required` versus `capability_written`: an older reader must never silently apply a disposition it does not understand, and an EXT_ENVELOPE's active payload must be a pure function of `(disposition, reader generation)`, never reader discretion.

### Attack surfaces and specific findings

Prior adversarial review rounds against this exact architecture (recorded in `research.md`) found and, in most cases, closed the following. Each row states the current status honestly; three remain open and are not hidden.

| Threat | Vector | Mitigation | Requirement |
|---|---|---|---|
| Ring-slot rollback / truncation replay | Truncate the file to an earlier, still-digest-valid commit boundary so a reader silently opens a stale, possibly still-signed state | `ledger_length` is bound twice: once as the winning ring record's own field (checked against actual file length at open), once inside `structure_digest`'s preimage (checked at every signature verification). Both must agree with reality | FR-117, TR-010, DP-013 |
| Ring-slot tie at equal sequence | Craft two self-digest-valid ring slots sharing the highest sequence number, each naming a different state, so two conforming readers open different documents from one file | PD-RING-001: an equal-sequence tie at the highest valid sequence is a structural reject naming both slot indices, never resolved by slot position or wall-clock | FR-105, FR-117 |
| T_C leaf/internal domain confusion | Craft a redactable leaf whose salted preimage collides with an internal node's preimage shape, letting one T_C root describe two structurally different trees | Three distinct, non-confusable domain tags: 0x02 (redactable leaf), 0x07 (non-redactable leaf), 0x08 (internal), plus reserved 0x00 for `ABSENT_CHILD_DIGEST`; no real preimage ever starts with 0x00 | HC-014, DP-006 |
| Ed25519 verification divergence | Exploit non-canonical point encodings or small-order keys, on which different Ed25519 libraries are documented to disagree, to get one verdict from one implementation and a different verdict from another, or to achieve a key-substitution / message-binding break | The closed 7-step EdDSA-Protodoc-1 procedure (DP-007) rejects every documented divergence case with plain integer or byte comparisons before any curve arithmetic runs, so two implementations see byte-identical behavior on every input | FR-063, FR-115, NFR-028 |
| Self-coverage circularity | A signature's own octets, or another signature's octets, influence the validation verdict (FR-002's own wording), so they would need to be inside their own covered set, which is circular | ATTEST-typed segments (signatures, time attestations, severance and RESCIND-AND-RESIGN records) are categorically excluded from every signature's own and every co-existing signature's coverable set, by segment type, not by convention | FR-002, HC-015 |
| Forged inventory | Present a self-consistent but false segment-table entry (wrong type, length, or digest) to a bounded-prefix consumer, so it clears a document it never actually inspected | `structure_digest` recomputes every covered ordinal's type, length, and digest FRESH from the file's current octets at verify time; it never trusts a stored slot value. A full verifier's T_S recomputation independently catches inventory tampering outside any signature's coverage | TR-009, FR-104, FR-105 |
| Extension-envelope fallback cycle | Two or more EXT_ENVELOPE frames whose `fallback_ref` fields point to each other, so disposition resolution loops without bound on a structurally valid, non-oversized document | FR-109's cycle detection runs over 5 named edge kinds, including extension-envelope fallback references, before any structural ceiling is evaluated | FR-109, HC-013, HC-025 |
| Frame-directory offset arithmetic | An attacker-controlled `frame_count` and `segment_length` pair causes the frame-directory offset computation to wrap or land outside the segment | `frame_count*48 + 32 <= segment_length - 64` is checked with overflow-safe arithmetic before the directory offset is computed, before any frame is read | HC-008, FR-106 |
| Decode-before-verify on media | Construct a decoder (PLP-1, restricted-PNG) on attacker-controlled octets whose declared pixel dimensions are far larger than the encoded payload implies, before any authentication has occurred | Every declared decoded size is checked against `MAX_DECODED_UNIT = 268,435,456` before a framebuffer is allocated. This specific check (declared width times height times bytes-per-sample against the ceiling, at the codec entry point, before allocation) is a named task-phase obligation; it is not yet exercised by a conformance vector, and is called out here so it is not lost | CP-006, HC-008, NFR-030 |
| Redaction residue via orphan carriage | An annotation's orphan-carriage `quoted_text` field snapshots the plaintext of a deleted anchor range at the moment of orphaning; if that range later falls inside a redacted subtree, the quoted text is a separate copy of the same content that a per-text-block residue scan might not visit | RED-ALIGN requires every redactable subtree boundary to land on a text-block boundary, and an annotation's own comment body is itself a TextBlock subject to the same scan. Whether the orphan-carriage `quoted_text` field specifically is included in FR-078's octet-level residue scan at publish time is not yet demonstrated by a conformance vector. This is flagged as an open item, not claimed closed | FR-078, FR-030, FR-075 |
| Bounded-prefix preview authentication gap | A bounded-prefix-only consumer (mail gateway, thumbnailer) is asked by FR-053 to detect a stale preview by comparing a recorded input digest against the actual render inputs, but the actual render inputs (page-1 content, fonts, resources) live in the ledger, past the 262,144-octet window the check is scoped to | Not fully closed. The digest check is genuinely computable only for the in-window subset of inputs (title, page count, dimensions, language); it cannot authenticate the raster itself against actual content from within the window alone. This is disclosed as spec conflict item 5's sibling gap in Section 9, and any automated, unattended consumer of an untrusted file must run its raster decoders in a memory- and time-bounded sandbox regardless of signature status, as a deployment requirement, not a format guarantee | FR-052, FR-053, TR-006 |
| Severed-state dictionary attack | FR-061's literal text calls for a bare, unsalted digest per unreconstructable published state, which is a brute-force oracle over exactly the low-entropy content (names, dates, amounts) FR-075 spends 256 bits of salt to protect against for redaction | This plan ships the salted-commitment form (`severance_commitment = H(severance-domain \|\| severance_salt \|\| state_digest)`, salt destroyed at trim) as its answer. This is a proposed amendment to FR-061's text, not yet approved; see Section 9 | FR-061, FR-075 |
| Signature-scheme break (not a hash break) | Ed25519/Curve25519 itself, not merely a hash, is broken, and no mechanism connects an old signature to a new scheme | RESCIND-AND-RESIGN (DP-017), available at a major-version migration boundary: a new signature under a new allowlist scheme over the migrated content, with the old signature retained. The mechanism exists but nothing in the format self-triggers it; see Section 7 and Section 8 | CON-016, DP-017 |
| Unknown-construct silent drop | A writer or merge tool discards a construct it does not understand instead of preserving it or refusing | Every non-core construct travels in EXT_ENVELOPE with a mandatory core-expressible fallback for ignore and degrade dispositions; a writer either reproduces unimplemented constructs octet-for-octet on save and merge, or refuses, naming what it could not preserve. An unregistered tag is rejected, never skipped | CP-013, FR-017, FR-018 |
| Merge classifier edge case | An operation kind, or a range list at exactly the `MAX_SEGMENTS` boundary, that does not cleanly resolve to one of DISJOINT-COMMUTE / R1 / R2 / R3 | The classification is claimed exhaustive by construction (every operation is a position-claim, value-claim, or delete), but has not yet been exercised against a property-based or adversarial corpus. `tasks.md` must include a dedicated exhaustive-pair test and a dedicated CoverageDescriptor range-list fuzzer as first-class tasks, not folded into general conformance work | HC-026, FR-092..096, Section 10 |
| Cross-document linkability | Content addressing (CQ-007) makes identical component payloads provably identical to anyone holding two documents, a confidentiality-relevant channel | Explicitly out of v1 scope, consistent with CQ-011's off-critical-path scoping of confidentiality features generally; documented as a residual risk (Section 8), not treated as closed | CQ-007, CQ-011 |

---

## 7. Migration and versioning

**Frozen forever.** Header octets `[0,32)`: `magic`, `format_major`, `format_minor`, `document_class`, `capability_written`, `capability_required`, `durable_claim`, `history_mode`, `unicode_version_id`, `shaping_profile_id`, `prefix_layout_id`. No future major version relocates, resizes, or reinterprets any field inside this window (CP-008, FR-006, FR-010). This is what lets FR-123 determine "can I even read this file's version" before assuming anything else about its layout. `prefix_layout_id` (a `uint64`, reserved, v1 value 1) exists specifically so a future major version can declare a differently-shaped prefix without perturbing this frozen window.

**Changeable within a major version.** `format_minor` and capability generations: new capabilities are added and gated by `capability_written` / `capability_required` arithmetic, resolved through the existing disposition machinery (ignore, degrade, refuse), never a header layout change. PDL-TLV field tags 12-255 within an existing record type's documented reserved tail: safely skippable by a v1 reader via the field's own length prefix. `shaping_profile_id` and `unicode_version_id`: these are registry-issued, per-document values, not frozen-window constants, so a document can pin a newer shaping algorithm or Unicode version without a format version change.

**Changeable only at a major-version boundary.** New `SegmentTableSlot.segment_type` values (the range 5-255 is reserved and a v1 reader MUST reject, not skip, any segment carrying one; segment type governs how T_S and T_C treat a segment's digest and how CoverageDescriptor partitions it, properties CP-004's determinism audit and CON-005's one-representation rule require closed, not writer-extensible within a version). New entries in the CON-015 cryptographic parameter allowlist, paired with RESCIND-AND-RESIGN. A differently-shaped prefix (via a new `prefix_layout_id`).

**Never changeable, full stop.** The 4-site enumerated non-determinism allowlist (CQ-004, a CP-004 constitutional invariant, changeable only by constitution amendment, not by a format version). Segment-type semantics within an already-shipped major version. A permanently-retired extension token (CON-020): never reissued, under any version.

**Migration mechanism.** Refusal-first, two-phase (DP-012). Phase 1 enumerates every construct in the source document with no representation in the target major version, naming the construct and its location; if that enumeration is non-empty, migration halts and writes nothing. Phase 2, only if phase 1 found nothing, emits `L*(migrate(state))` as a fresh file. Every content-unit identifier (`run_id`, `unit_id`) passes through unchanged (FR-120), so comments, suggestions, and cross-references stay attached across the boundary. Pre-migration signatures are retained, unmodified, and are reported as covering the pre-migration state specifically (FR-122); they are never valid over migrated content, and a reader must never suggest otherwise.

**RESCIND-AND-RESIGN**, available only inside phase 2 of a migration, at a major-version boundary: a custodian adds a fresh signature under a new allowlist scheme over the migrated `T_C_root`, while the prior signature and its full metadata remain retained, unmodified, in the pre-migration state that migration already guarantees stays reconstructable. A verifier presented with both reports an explicit two-entry custody chain (old scheme and old time, new scheme and new time), never silently preferring one. This is distinct from CON-016's existing hash-only re-protection layer, which wraps the original signature with a stronger digest and a fresh time attestation but never touches the signing act itself, and so helps only against hash-family weakening, not against the signature scheme itself breaking.

**Disclosed gap, not hidden.** RESCIND-AND-RESIGN supplies the mechanism, not the trigger. Nothing in the format self-initiates a migration-and-rescind when a cryptographic parameter is later deprecated; a v1 document signed only under Ed25519, never migrated before Ed25519 itself breaks, has no protection this mechanism can retroactively supply. Closing this gap is a governance and operational question (a published advisory process, mirroring CP-012's disclosure-clock discipline), outside this plan's authority. See Section 8, risk row on signature-scheme survival.

---

## 8. Risks

| Risk | Severity | Mitigation | Trigger to revisit |
|---|---|---|---|
| TOTAL-mode signed documents get no benefit from partial compaction, since nothing is uncovered to reclaim; `MAX_SEGMENTS = 16,384` under continuous 5-minute autosave still exhausts in roughly 58 days | medium | A distinct refusal status (`PD-CAP-001`) names the ceiling and observed segment count. Tooling guidance treats a TOTAL-mode signature on an actively-edited document as a documented anti-pattern; recommend SUBSET-mode coverage for any document expected to undergo continued heavy editing after signing | First real report of a TOTAL-mode signed document approaching `MAX_SEGMENTS`, or before, at the tasks-phase conformance-vector stage |
| SUBSET-mode signed documents benefit from partial compaction by an amount that depends on the covered-to-uncovered edit ratio, which varies per document and workflow; no single day-count is universal | low | Expose covered-versus-uncovered segment counts as an observable metric; recommend periodic explicit partial compaction for actively-edited SUBSET-signed documents; state the assumed edit ratio in any benchmark that claims a day-count | A conformance benchmark is published implying a universal day-count without stating its assumed ratio |
| NFR-032's 2.0x history-size ratio is missed, not merely degraded, for an adversarially scattered single-character-edit trace, with no repair path (re-basing runs would detach every anchor built on them, forbidden by HC-011) | medium | Expose run-fragmentation as an observable metric; recommend periodic explicit compaction, which does coalesce runs; state the assumed trace shape in any benchmark | A real editing workflow (not just a synthetic fuzzer) is observed producing scattered single-character edits at scale |
| PLP-1 is a from-scratch, previously unbuilt, unbattle-tested codec; small and exactly specified removes the CP-003 infeasibility that killed AV1, but an undiscovered bitstream ambiguity or IDCT shift-schedule off-by-one could still exist | medium | Ship exhaustive decode conformance vectors (every fixed quantization matrix, every Huffman table branch, boundary block sizes, worst-case DPCM chains) as a first-implementation activity, not an afterthought | Before any PLP-1 implementation task begins; conformance vectors must exist first |
| CQ-006 option B's shaping oracle is a named, versioned external implementation; a second, independently-authored implementation must still reproduce its exact output from prose plus conformance vectors alone, the same reproducibility problem CQ-006 option A was rejected to avoid, displaced by one hop | high | A normative per-cluster shaping-trace format is mandatory for the entire durable-profile conformance corpus, not a sample; the shaper's exact source snapshot and a hermetic build recipe are mandated as conformance-suite artefacts, vendored, not merely referenced. This does not close the risk for a lone file divorced from that repository decades out; that residual is real, disclosed, and bounded by CQ-006's already-approved resolution | EX-001's expiry condition (Section 9) is met, or an archival-recovery drill is attempted against a file with no working shaper copy available |
| A bounded-prefix consumer of an unsigned document can check the preview raster's self-consistency but cannot cryptographically authenticate it; a hostile unsigned document can deliver a crafted, memory-unsafe-decoder-targeting raster to such a consumer | medium | Inherent to TR-006/007/008's own requirement of no network access, no decompression, no full-file read for prefix determinations. Any automated, unattended consumer of untrusted files must run its raster decoders in a memory- and time-bounded sandbox regardless of signature status, as a deployment requirement | A production incident involving an automated thumbnailer or gateway |
| Content addressing (CQ-007) makes identical component payloads provably identical to anyone holding two documents, a linkability channel v1 defines no confidentiality model to close | medium | Explicitly out of v1 scope, consistent with CQ-011's off-critical-path scoping of confidentiality features generally | A confidentiality model is proposed for a future version |
| Publish and migration both re-emit the entire document via `L*(state)` as a fresh file, so cost scales with total document size, not with the size of the change | low | Acceptable: both are explicit and infrequent by design, never invoked from an ordinary save. RESCIND-AND-RESIGN inherits this same cost profile since it only occurs at a migration boundary | A workflow emerges that calls publish or migrate at high frequency on large documents |
| RESCIND-AND-RESIGN is a mechanism without a self-triggering condition; the window between a scheme breaking and someone actually performing the migration-plus-rescind is a real exposure | medium | Genuine, disclosed gap. Recommend a governance-level obligation, e.g. a published advisory process that fires a migration recommendation when a v1 cryptographic parameter is deprecated, mirroring CP-012's disclosure-clock discipline. Committing to that process is outside this plan's authority | A v1 cryptographic parameter (Ed25519 or Curve25519 itself) is publicly deprecated |
| Registry Excerpts are mandated only for durable-profile documents; ordinary documents remain fully dependent on the central extension/shaping/Unicode-version registry surviving, with no succession plan named anywhere in this architecture | medium | Accepted, scoped tradeoff: mandating Registry Excerpts universally would burden the overwhelming majority of ordinary documents for a decade-scale guarantee only the durable profile actually claims | Evidence that non-durable documents need decade-scale interpretability in practice |
| The CQ-012/CP-003 two-implementation gate may never be cleared absent a named sponsor, standards body, or bounty; every fix in this pass makes the second implementation's job more tractable but does not manufacture a reason for a second team to exist | high (RESOLVED 2026-09-18) | **Decision recorded 2026-09-18 by Eyvar García: path-forward is a community bounty.** A public bounty will be posted and an implementer recruited from the community to build the second independent implementation per the T-0348 protocol. Fallback if unfunded: no bounty claimant within a reasonable window defers T-0351 indefinitely; v1-stable declaration (T-0358) records this as an open blocking gate rather than proceeding without CP-003 satisfied. See specs/CHANGES.md for the structured entry | Before phase 5 (analyze) closes, per this row's own mitigation |
| CoverageDescriptor's merge-adjacent canonicalisation and the DP-015 R1/R2/R3 merge classification are both new normative surfaces in this pass; neither has been exercised against a real adversarial or property-based test corpus | medium | `tasks.md` must include, as first-class tasks, a dedicated property-based fuzzer for CoverageDescriptor range-list canonicalisation and a dedicated exhaustive-pair test asserting every operation kind resolves to exactly one of {DISJOINT-COMMUTE, R1, R2, R3}, never a fourth case | Before any merge or sign/verify implementation task is marked done |
| Whether an annotation's orphan-carriage `quoted_text` field is included in FR-078's octet-level residue scan at publish time is not yet demonstrated (Section 6) | medium | Add an explicit conformance case: redact a subtree containing an annotation's anchor range, confirm the orphan-carriage `quoted_text` for that annotation is redacted alongside it, not merely the TextBlock content | Before the redaction/publish conformance suite is marked complete |

---

## 9. Conflicts with approved artifacts

Every conflict found between this design and the frozen `spec.md` / `clarify.md` text is recorded here, none hidden. Each needs a ruling from Eyvar, or Eyvar and themis together, before it can be treated as resolved; none is currently approved.

### Conflict 1: FR-061 versus FR-075

**Conflict.** FR-061's literal text (spec.md line 523) requires enumerating "the identifier and digest of every previously published state that it can no longer reconstruct," a bare, unsalted digest. Because the digest is unsalted, it is a brute-force dictionary oracle over exactly the class of content (names, dates, amounts) that FR-075 spends 256 bits of salt to protect against the identical attack in the redaction path. Implementing FR-061 literally reopens, for trimmed history, the exact privacy hole FR-075 exists to close for redaction.

**Proposed resolution.** This architecture ships the salted-commitment form (`severance_commitment = H(severance-domain || severance_salt || state_digest)`, salt destroyed at trim, the same permitted non-deterministic-site category as redaction salts) as its answer. This is a proposed clarifying amendment to FR-061's text, not a silent deviation: it is named here explicitly. **Approval needed**: Eyvar or Eyvar and themis together, before phase 5 (analyze) closes.

### Conflict 2: CON-006 / CP-009 versus CQ-006 option B

**Conflict.** CQ-006 (already approved, not reopened here) resolved to option B, pinning a named, versioned shaping-algorithm implementation as the normative rendering oracle. `clarify.md`'s own text for CQ-006 acknowledges this is "close to the boundary CON-006 draws." CON-006 (spec.md line 1255) requires zero normative statements whose meaning is defined by reference to a named application or application version, gated by an audit (A-VENDOR) that must return zero product-behavior references before release.

**Proposed resolution.** A recorded, one-time exception, **EX-001**, under constitution amendment rule 6:
- **Principle**: CON-006, and its constitution-level mirror CP-009.
- **Reason**: CQ-006 option B is approved, and no normative from-scratch shaping specification exists yet to reference instead.
- **Expiry condition**: superseded when a normative Protodoc shaping specification is adopted (target: a future major version), or when milestone M-SHAPE passes the two-implementation gate.
- **A-VENDOR** is amended to permit exactly this one match at a recorded line, and to fail on any second occurrence.

This exception is **proposed here, not approved**. Per the constitution's amendment process, only Eyvar approves an exception, and an exception with no expiry condition is an amendment rather than an exception; this one carries the expiry condition above so it can be recorded as an exception. Until approved, A-VENDOR blocks every release.

### Conflict 3: NFR-030's literal text

**Conflict.** NFR-030's literal text (spec.md line 1183) reads "SHALL bound a conforming reader's peak resident memory on any input, valid or malformed, at 4 times that input's octet length." This is unsatisfiable by any real process for small inputs: a Go binary's runtime and stack alone occupy megabytes regardless of a tiny input file's size.

**Proposed resolution.** Read the requirement as bounding heap allocated for document data specifically, with an explicit floor exemption for any input shorter than 1,048,577 octets. Recommend the requirement text be corrected to say so explicitly; otherwise CP-012's fuzzing oracle fires on every small, entirely valid input. Every "satisfied" verdict against NFR-030 in Section 5 is conditional on this reading, not on the literal text. **Approval needed**: a text correction to spec.md, or an explicit floor exemption added to it, by Eyvar.

### Conflict 4: NFR-027's rendering-role budget, specifically the shaping share

**Conflict.** The rendering conformance role's 30-working-day external-implementer budget must absorb reproducing a pinned shaper's exact output from prose (Conflict 2, above), a from-scratch exact-rational rasterizer, and a bit-exact custom lossy codec. This pass's repairs (naming the rasterizer algorithm, replacing AV1 with the much smaller PLP-1) cut the credible estimate from 60-200+ days to roughly 28-49 days, but do not close the gap to 30 entirely. The residual overrun is concentrated almost entirely in the shaping sub-component, estimated at 15-25 days on its own, an oracle-matching effort under CQ-006 option B, not a lines-of-code task.

**Proposed resolution.** No further architectural mitigation is available for the shaping share without reopening CQ-006, which this phase has no authority to do. Recommend the day count be revisited specifically for shaping, as an explicit scope decision under CP-002 (the precedent clarify.md's AD-002 already sets: "the correction is a scope decision under CP-002, not a silent re-tiering"), rather than treating the whole rendering role's budget as generally at risk. **Approval needed**: Eyvar, or Eyvar and themis together, to make the scope decision.

### Conflict 5: FR-055's read-count clause versus its octet-ceiling clause

**Conflict.** FR-055's text (spec.md line 481) states one sentence covering both a read-count ceiling ("locate and read... using at most 3 sequential dependent reads") and an octet ceiling ("no more than... beyond the unit itself"). A reading that lets the octet clause's "beyond the unit itself" exclusion also narrow the read-count clause can undercount the hop budget by one: at the stated fanout, reaching the leaf that names a unit and then reading the unit itself could plausibly be read as 4 hops, not 3.

**Proposed resolution.** This architecture reads the two clauses as independently scoped: the read-count budget of 3 includes the final read of the unit itself; the octet-ceiling budget excludes the unit's own bytes from what counts as "beyond" it. Under this reading, the locate path (prefix-including-index-route as read 1, index leaf as read 2, unit as read 3) closes within 3 total reads. **Approval needed**: none strictly, since a reading is available that satisfies the requirement, but recommend the requirement text be restructured into two independently-scoped sentences for future implementers who have not seen this analysis.

---

## 10. Phase 4 inputs

Per AD-002, requirement priority carries no sequencing information (all 197 requirements are `must`). `tasks.md` must therefore produce an explicit dependency ordering on its own; it cannot fall back on priority. This section states that ordering's spine.

### The dependency spine

1. **`pdlfmt` and `container` first, with their own conformance vectors, before anything else is built.** Every other package reads through PDL-VARINT, PDL-TLV, and the fixed-layout prefix structs. A bug here is a canonical-octet divergence that no later layer can mask. This includes the ring-winner-selection tie-break (PD-RING-001) and the bounds-safe `SegmentTableSlot` arithmetic check, both of which need their own at-limit and over-limit conformance files (CON-010) before any content-model work begins.
2. **NFC validation and identity minting (`content` package) next, before any Run, TextBlock, or Annotation task.** Both are foundational to the A-FIELD-ROLE audit (IDENTITY-COMPONENT versus positional arithmetic) that the rest of the content model depends on being correctly enforced.
3. **`extract` can and should be built early, in isolation.** TR-011 requires it to have no font, graphics, or crypto dependency, and its own conformance suite (FR-041/035/042/047/048) does not need `integrity`, `render`, or `merge` to exist first. This is the cheapest early, self-contained conformance win, and validates the prefix and content-model work from step 1-2 end to end before the harder layers begin.
4. **The 7-step EdDSA-Protodoc-1 wrapper is built and vector-tested in complete isolation before any Signature, CoverageDescriptor, sign, verify, redact, publish, or migrate task.** Every one of those depends on it, and its correctness (the canonical and small-order pre-filters specifically) is exactly the kind of narrow, high-consequence unit that should have its own conformance vectors authored before it is wired into anything else.
5. **The generated ceiling table (CON-009/CP-007: checked-in Go constants with a CI equality check against the specification text) must exist before `validate`'s cycle-detection and structural-ceiling tasks**, since those tasks check against it directly.
6. **T_S and T_C construction (including the `ABSENT_CHILD_DIGEST` filler and the three domain tags) come before any Signature task**, since `signed_object` depends on `T_C_root` and `structure_digest`, both of which depend on the trees being correct first.
7. **`history` and `merge` depend on `content` (for run/unit identity) and on `integrity`'s `state_id` comparator (for R2 TOTAL-ORDER-TIEBREAK), so they cannot start before both exist**, even though they do not depend on `render` or `migrate`.
8. **`render` depends on `validate` and `integrity`** (for presentation-artefact pinning) but not on `merge` or `history`; it can proceed in parallel with steps 6-7 once `validate` and `content` are stable.
9. **`migrate` and `registry` depend on `validate` and `integrity`, and come last among the mid-layer packages**, since RESCIND-AND-RESIGN specifically requires a working Signature implementation to retain and a working migration path to attach to.
10. **`canon` (the `L*` traversal used by compact, publish, migrate, and fresh writes) depends on everything except `render`, `extract`, and `merge`, and is therefore one of the last packages to stabilize**, not the first, despite being conceptually central to CQ-003.

### Governance-gated items, and what they do and do not block

- **EX-001 (Conflict 2) gates only the shaping sub-slice of rendering tasks**, specifically anything that reproduces the pinned external shaper's output. It does not gate the rasterizer, PLP-1, reflow, or fixed pagination, all of which are fully self-specified in this plan and can proceed without EX-001's approval.
- **The FR-061 ruling (Conflict 1) gates only `ErasureRecord`'s final wire shape.** Recommend proceeding with the salted-commitment form as a provisional implementation, explicitly flagged and reversible, rather than blocking all of `history` on this ruling.
- **The NFR-030 reading (Conflict 3) gates only the fuzzing harness's peak-memory oracle threshold.** It does not block any other implementation task.
- **The CQ-012 second-implementation funding decision (Section 8's high-severity risk) gates only the "declare v1 stable" milestone.** It does not gate any coding task; coding should proceed in parallel with that governance decision.

### First-class tasks, not folded into general conformance work

Per the open items surfaced in `research.md` Section 8 and repeated in Sections 6 and 8 of this plan, `tasks.md` must include each of the following as its own named task with its own test, not as a sub-bullet of a larger task:

- A property-based fuzzer for `CoverageDescriptor` range-list canonicalisation (merge-adjacent, sorted, exactly-one-valid-encoding), including a range list at exactly the `MAX_SEGMENTS` boundary.
- An exhaustive operation-kind-pair test asserting every operation kind in the actual schema resolves to exactly one of `{DISJOINT-COMMUTE, R1, R2, R3}`, with a fourth-case failure treated as a test failure, not a warning.
- PLP-1 exhaustive decode conformance vectors (every fixed quantization matrix, every canonical-Huffman branch, boundary block sizes, worst-case DC-DPCM chains), authored and merged before PLP-1's decoder implementation begins, not after.
- A conformance case confirming an annotation's orphan-carriage `quoted_text` is included in FR-078's octet-level residue scan when its source range falls inside a redacted subtree.
- A conformance vector exercising the `frame_count*48+32 <= segment_length-64` bound at exactly the ceiling and one octet past it (CON-010's at-limit/over-limit pairing).
- A conformance vector exercising PD-RING-001's equal-sequence tie rejection.
- A conformance vector exercising the extension-envelope fallback-reference cycle case as one of FR-109's 5 named edge kinds.

No task in `tasks.md` may be sequenced by requirement priority. Sequencing comes entirely from the dependency spine above.
