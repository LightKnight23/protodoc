# Protodoc — Clarifications

Status: OPEN — blocks the plan phase | Spec ID: 001-protodoc-format-core | Date: 2026-09-05

No item below may pass into `plan.md` unresolved; each open question is a specification defect until Eyvar records a resolution here.

## Open questions

Ordered by downstream gating weight: the content model first, the programme shape last.

### CQ-001: Persisted position model

- **Question:** How are positions inside text persisted: identity anchors only, integer offsets, or offsets legal only inside an immutable published state?
- **Blocks:** FR-025, FR-026, FR-028, FR-029, FR-042, FR-084, FR-092, FR-108, CON-001. Gates the entire content model in `plan.md` and every entity in `data-model.md`. Every annotation, comment, change record and cross-reference construct is downstream of it.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Identity anchors only: no persisted count-based position anywhere | Anchors survive concurrent and offline edits with no repair function; FR-025 and P-ANCHOR are satisfiable by construction | Every anchor carries an identity token instead of an integer, so at-rest size and index size rise; range endpoints need explicit boundary semantics (FR-026) that offsets get free |
| B. Integer offsets, as in the origin sketch | Smallest at-rest representation; trivially serialised; locators computable without an identity index | A remote insertion before an annotated range silently relocates it, and a file on disk has no transform function available to repair a stale offset; contradicts FR-025 and makes P-ANCHOR unpassable |
| C. Hybrid: counted positions legal only inside an immutable published state, identity anchors mandatory for anything surviving an edit | Compact locators in published and signed artefacts; identity where editing happens | Two position models in one format, so every consumer implements both and CON-005 needs an explicit carve-out; conversion at the publish boundary is a new failure surface |

- **Recommendation:** Option A. No repair function exists at rest, so an offset is a stale pointer the moment anyone else edits; CON-001 then scopes counted positions to non-persisted surfaces only.
- **Resolution:** _pending_

### CQ-002: Smallest unit carrying durable identity

- **Question:** What is the smallest unit carrying durable identity: character, contiguous authored run, or block?
- **Blocks:** FR-019, FR-020, FR-021, FR-022, FR-024, FR-093, NFR-032, NFR-033. Fixes merge granularity, annotation granularity and the at-rest size floor; determines the identity index shape in `data-model.md`.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Character identity, stored per run with run-merging | Phrase-level comments, ranges crossing block boundaries, character-granular merge and non-interleaving (FR-093) | Per-character identity state, so the encoding must carry run-merging from day one or NFR-032's 2.0x history ratio and NFR-033's load ratio are both missed |
| B. Run identity: contiguous authored runs, split on edit | Fewer identities than characters; runs map directly onto formatting ranges | A run splits on every interior edit, minting identities anyway, so the size saving erodes with editing while merge granularity is coarser than the edits users make |
| C. Block identity only: ranges and comments snap to whole blocks | Smallest identity index; simplest merge; smallest at-rest floor | Cannot express a comment on a phrase or a range crossing a block boundary, so US-002 and FR-025 are unsatisfiable and every annotation silently widens to its block |

- **Recommendation:** Option A. It is the only option that expresses a phrase-level comment and a cross-block range, and run-merging holds the demonstrated cost under one extra octet per character.
- **Resolution:** _pending_

### CQ-003: What canonical form is a property of

- **Question:** Do byte-stable canonical form and bounded incremental write apply to the same artefact, or to a canonicalisation function and a file at rest respectively?
- **Blocks:** NFR-001, NFR-002, NFR-003, NFR-004, NFR-008, NFR-009, FR-003, FR-056, FR-057, FR-058, FR-063. Decides what is signed, what is hashed, what is content-addressed and what a no-op save must preserve.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Canonical octets are a function of state, computed without rewriting the file; the file at rest is any member of a defined equivalence class; compaction is explicit and never implicit in a save | NFR-001, NFR-002, NFR-004, NFR-005 and NFR-008 hold simultaneously; one named signable object; bounded per-edit write survives | Two objects to reason about, so the specification must define the equivalence class precisely and every tool needs a canonicalisation pass separate from its writer |
| B. Canonical form is the file; an append region must be compacted before publish and before signing | One artefact, so equality is file equality and content addressing is trivial | Compaction is a whole-file rewrite, so publishing or signing a 500 MB document re-transfers it and NFR-009's novel-chunk bound is breached at exactly the moment signatures are added |
| C. Both are normative file states distinguished by a declared flag, with identical state digests | Writers can pick the cheap state and still publish a canonical one; migration between them is explicit | Two normative file states means two parser paths, two sets of conformance cases and a flag whose mis-setting is a new class of interop divergence, against CON-005 |

- **Recommendation:** Option A. It is the only assignment under which byte-stability and bounded write are jointly satisfiable, and it names one signable object without forcing a rewrite.
- **Resolution:** _pending_

### CQ-004: Permitted non-deterministic inputs

- **Question:** What non-deterministic inputs are permitted, given that identifier minting, redaction commitments and signatures all need values the determinism rule bans?
- **Blocks:** NFR-005, NFR-006, NFR-007, FR-023, FR-070, FR-074, CP-004. Without an answer the format cannot mint an identifier, redact safely or sign at all.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Enumerated allowlist of exactly four permitted sites (identifier minting, redaction salts, signature values, timestamp attestations), with deterministic signature schemes mandated | Determinism audit stays mechanical: any ambient value outside the four sites is a defect found by A-AMBIENT rather than argued about | The list is closed, so any future capability needing entropy is a major version change; deterministic signature schemes narrow the CON-015 allowlist and exclude common hardware modules that randomise |
| B. Blanket ban with a general user-requested-content exception | Shortest rule; no enumeration to maintain | The exception is unbounded, so an ambient value can be justified anywhere and the static audit degrades to a judgement call, which is how clock and session values leaked into incumbent content |
| C. Determinism required only of content components, with integrity and signature components excluded wholesale | Simple partition; signatures never fight the audit | Identifier minting sits in content and still needs entropy, so the rule fails on its first real case; a wholesale exclusion also hides ambient values inside integrity components where no one audits them |

- **Recommendation:** Option A. A closed, named list keeps the audit mechanical, and deterministic signing means one state signed twice with one key yields identical octets (NFR-006).
- **Resolution:** _pending_

### CQ-005: Signature survival across publish and redaction

- **Question:** What happens to existing signatures when publish removes content or strips actor identity?
- **Blocks:** FR-063, FR-066, FR-067, FR-074, FR-075, FR-076, FR-077, FR-078, FR-080, FR-081. Determines whether US-004 is achievable at all and how `plan.md`'s threat model is shaped.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Redaction-compatible signature: subtrees designated redactable at signing may be omitted later; verifier reports a distinct positive verdict enumerating declared omissions | A published redacted copy still carries the original attestation; undeclared omission yields unverified, closing the incumbent attack class | The signer must anticipate what may later be redacted, and the construction needs per-subtree hiding commitments plus salts, adding cryptographic surface and a second verdict for every consumer to render honestly |
| B. Publish always invalidates prior signatures; the published copy must be re-signed by the publisher | Simplest verification: coverage is total, one verdict, no commitments | The publisher must hold a signing key and becomes the attesting party, so the original signatory's attestation is destroyed by redaction and US-004 is unmet for anyone who cannot re-sign |
| C. Publish is prohibited on signed documents; redaction must precede signing | No signature ever spans removed content; smallest specification | Real workflows sign first and redact later for disclosure, so the format refuses the archetypal compliance case and users fall back to printing and re-scanning |

- **Recommendation:** Option A, with hiding commitments so omitted content is not recoverable by search over retained digests. It is the only option that lets a compliance officer publish a redacted copy that still carries the original attestation.
- **Resolution:** _pending_

### CQ-006: Rendering conformance strictness

- **Question:** Is rendering conformance byte-identical, or tolerance-bounded, given that no normative text-shaping specification exists anywhere to reference?
- **Blocks:** FR-035, FR-064, FR-065, NFR-019, NFR-020, NFR-021, NFR-022, NFR-023, NFR-024, NFR-027. Prices the largest single deliverable in the programme.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Write a normative shaping specification, algorithm, feature set and application order pinned to exact Unicode and font-technology versions | Full independence: byte-identity holds with no external oracle, and the format owns its own correctness | A multi-year deliverable on the critical path, and one no existing standards body has completed; it also enlarges the rendering role far beyond NFR-027's 30-day implementer budget |
| B. Normatively reference one named, versioned shaping algorithm implementation as the oracle, keeping byte-identity relative to that version | Byte-identical rasterisation is achievable now; conformance vectors pin the behaviour so the reference cannot drift silently | A normative dependency on an external artefact, close to the boundary CON-006 draws against product-defined behaviour, and a version bump upstream is a format version event |
| C. Keep rasterisation tolerance-bounded, stating per-channel and per-sample tolerances, and reserve byte-identity for text-free content | Removes the shaping problem from the critical path entirely; implementers can use existing text stacks | The pinned-presentation signature claim weakens to approximately what the signatory saw, and NFR-019's zero-tolerance raster equality plus the durable-profile claim both have to be restated |

- **Recommendation:** Option B for v1: pin a named versioned shaping algorithm and an exact Unicode version, prohibit execution of any font instruction stream, ship decode-and-shape conformance vectors. Option A stays the long-term goal, off the critical path.
- **Resolution:** _pending_

### CQ-007: Component addressing

- **Question:** Are components addressed by name or by the digest of their own octets?
- **Blocks:** FR-055, FR-056, FR-057, FR-058, FR-104, CON-008, TR-006, TR-009. Determines the container layout and whether deduplication and incremental verification share one mechanism.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Name-addressed with a separate digest field | No cross-document linkability; storage order trivially independent of content; simplest inventory | Deduplication, integrity and incremental verification stay three separate mechanisms, each with its own rules and its own divergence risk between implementations |
| B. Content-addressed with a name-to-digest index and storage order assigned by monotone ordinal | Deduplication, integrity and incremental verification fall out of one mechanism; FR-058's stable ordinals keep re-layout off the save path | Identical payloads are provably identical to anyone holding two documents, which is a linkability channel a later confidentiality model must revisit; the index adds one indirection to every fetch |
| C. Content-addressed with a per-document salt | Removes cross-document linkability while keeping one integrity mechanism | Cross-document deduplication is forfeited entirely, and the salt is a per-document ambient value needing its own site in the CQ-004 allowlist |

- **Recommendation:** Option B, paired with FR-059 so canonical storage order never depends on a digest. Linkability is acceptable while v1 defines no confidentiality model, and revisiting it is a stated dependency of any later one.
- **Resolution:** _pending_

### CQ-008: Default history mode

- **Question:** What is the default history mode for a newly created document, given that the mode is immutable at creation?
- **Blocks:** CON-022, CON-023, CON-024, CON-025, FR-059, FR-060, FR-061, FR-062, NFR-032, NFR-033. The default governs almost every document that will ever exist.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. History retained from a declared point, complete history opt-in | Review, audit and three-way merge over the working life, with a lawful in-place removal path intact | Trimming severs states, so FR-061 enumeration and FR-062's unavailable-state verdict become everyday cases rather than edge cases, and merge across a retention point must be refused (CON-024) |
| B. Complete history by default | Strongest audit and reconstruction story; FR-059 holds for every document; merge always has an ancestor | Every document becomes a standing erasure liability under European data-protection law, and in-place removal is refused permanently (CON-023) on documents created before anyone thought about it |
| C. No history by default | Smallest files; no erasure liability; simplest reader | Forfeits the review, merge and audit differentiator on the default path, so the properties the format exists for are opt-in and mostly unused |

- **Recommendation:** Option A. It preserves review and merge for a document's working life while leaving a lawful in-place removal path; complete history stays available for regulated and archival use.
- **Resolution:** _pending_

### CQ-009: Unicode normalization scope

- **Question:** Which Unicode normalization form is mandated, and is it a property of each text segment or of the concatenated stream?
- **Blocks:** CON-001, CON-002, CON-003, CON-004, FR-092, FR-093. Determines whether concurrent insertion at a combining-mark boundary is a reachable state.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. NFC, normalised per addressable text segment, no renormalisation across boundaries, with a named outcome at a concurrent-insertion boundary | Concurrent insertion at a combining-mark boundary stays serialisable; anchors never relocate; NFC is the conventional form for document text and identifiers | The concatenated stream may not itself be NFC, so extraction consumers comparing extracted text against externally normalised text see mismatches the format has to document |
| B. NFC over the concatenated stream, with writers required to reject | Extracted text is uniformly NFC, matching what search and comparison stacks expect | The forms are not closed under concatenation, so two authors each inserting valid text at adjacent positions produce a state no conforming writer may serialise, and the merge has no legal output |
| C. NFD per segment | Combining sequences are decomposed, making per-character identity and mark-level anchoring uniform | Diverges from the conventional interchange form, inflates scalar counts and at-rest size, and forces every consumer to recompose before comparison |

- **Recommendation:** Option A. Per-segment scoping is what makes concurrent insertion at a combining-mark boundary a reachable, serialisable state.
- **Resolution:** _pending_

### CQ-010: Raster encoding admitted in v1

- **Question:** Does v1 admit a lossy raster encoding, or exactly one lossless encoding?
- **Blocks:** FR-090, NFR-015, NFR-016, NFR-018, NFR-019, NFR-024. Determines whether documents containing photographs are competitive against the incumbents.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Admit exactly one lossy codec whose decode is bit-exact by specification, with decode conformance vectors in the corpus | Photographic documents stay near incumbent sizes while raster byte-identity survives; NFR-015's 300 ms and NFR-018's per-page read bound stay reachable | A second codec to specify, test and fuzz, and bit-exact decode pins the codec version permanently, since a decoder change is a rendering change |
| B. Lossless only, with the size penalty recorded as accepted and photographic content removed from the mobile and sync benchmark claims | One codec, uniform byte-identity, smallest conformance surface | A single full-page 300 dpi photograph runs 10 to 25 MB against an 8 MB per-page read budget: a 10x to 100x size penalty versus both incumbents, which prices the format out for the integrator audience |
| C. Lossless only in the durable profile, one bit-exact lossy codec elsewhere | Archival documents stay maximally faithful while everyday documents stay small | Two raster profiles, so a document changes size class when it claims the durable profile, and conversion at that boundary is lossy in one direction and untestable for equality |

- **Recommendation:** Option A. A bit-exact decoder profile preserves the raster byte-identity claim while removing the size penalty; leaving it out prices photographs out of the format.
- **Resolution:** _pending_

### CQ-011: v1 deliverable scope

- **Question:** What ships in v1, given that the drafted deliverable set is a multi-year programme?
- **Blocks:** FR-118, FR-125, TR-012, NFR-026, NFR-027, NFR-028, CON-026, plus the whole `tasks.md` critical path.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. v1 = specification, reference library, validate/inspect/extract/verify/diff tooling, conformance and negative corpora, governance artefacts, and a second implementation of container, validator and canonical serialiser; converters and full-scale rendering oracle deferred to a named post-v1 milestone | Every gate testing a property that cannot be retrofitted stays on the critical path; the two deliverables that are products in their own right move off it | v1 ships with no path in from either incumbent, so early adoption is limited to documents authored natively, and TR-012's full operation list has to be trimmed or restaged |
| B. v1 as drafted, including both converters | Adopters can bring existing corpora on day one; the loss-report story is demonstrable at launch | Reading either incumbent well enough to enumerate everything dropped is a multi-year effort, so the release gates get waived, after which they gate nothing |
| C. v1 = specification and corpora only; all tooling deferred | Fastest to a frozen specification; smallest team | Every asserted property becomes undemonstrable by a user, and CP-003's two-implementation gate has nothing to run against, so the specification freezes untested |

- **Recommendation:** Option A. It keeps every gate that tests a non-retrofittable property and moves the two product-scale deliverables off the critical path.
- **Resolution:** _pending_

### CQ-012: Second implementation before v1

- **Question:** Is a second implementation, in a different language by a different author, committed before v1 is declared stable?
- **Blocks:** NFR-001, NFR-019, NFR-028, CON-019, CP-003, release gate G-INTEROP.
- **Options:**

| Option | What it buys | What it costs |
|---|---|---|
| A. Yes: fund or recruit a second implementation of container, validator and canonical serialiser before v1 | The byte-identity and verdict-equality gates become executable, and divergence is found while files do not yet exist | Real funding or recruitment on the critical path, and a schedule dependency on a party outside the design team, which is the hardest item in the plan to compress |
| B. No: single implementation at v1, second implementation post-v1 | Fastest route to a shipped v1 with no external dependency | The format is that implementation's file format whatever the specification says, and divergence surfaces after real files exist, when it is unrecoverable; CP-003 must be suspended to release |
| C. Partial: second implementation of the validator only | Verdict-equality is testable at modest cost; catches the parser-differential class | Canonical octet equality stays untested, so the property the whole determinism programme rests on ships unverified across implementations |

- **Recommendation:** Option A, scoped to container, validator and canonical serialiser. Those are exactly the layers the byte-identity and verdict-equality gates test, and the layers where divergence is unrecoverable once files exist.
- **Resolution:** _pending_

## Resolved

None yet. Each resolution is appended here with its date, the chosen option, and the requirement IDs amended as a result; the corresponding open question is then struck from the section above.
