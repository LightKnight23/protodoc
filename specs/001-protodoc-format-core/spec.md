# Protodoc: Document Format and Tooling — Specification

Status: APPROVED — Eyvar, 2026-09-05 | Spec ID: 001-protodoc-format-core | Date: 2026-09-05

---

## 1. Problem statement

The two dominant document formats each fail in a way the other does not fix, and both failures are structural rather than implementational.

Package-of-markup formats rewrite the whole file on save, stamp per-session random identifiers into content, define parts of their normative behaviour by reference to one vendor's product, and permit signatures covering only an enumerated subset of the package: 2023 work found every tested product version vulnerable to all five signature attack classes.

Fixed-layout formats store the layout result rather than the content, so text recovery is heuristic, reading order is guessed from coordinates, and appended octets render while sitting outside the signature: 2019 and 2021 work broke signature validation in 21 of 22 and 16 of 29 tested viewers.

Both leak identity and deleted content through provenance surfaces that cannot be enumerated, so no tool can prove it removed everything; both are specified at a scale at which exactly one complete implementation has ever existed for one of them.

The result is a world where documents cannot be diffed, merged, cheaply synced, reliably indexed, proven redacted, trusted when marked signed, or verified in thirty years. Protodoc exists to make each of those a checkable property of the file rather than a hope about the application that wrote it.

---

## 2. Vision

Protodoc is a document format whose meaning, integrity and cost are decidable from the file itself. Every document declares what a reader must understand to render it, carries everything needed to render and verify it offline, has exactly one canonical octet sequence per logical state computable without rewriting the file, and states its own resource ceilings so a consumer can refuse it before allocating memory.

The specification is partitioned into reader roles small enough that a team outside the design group can implement one from the text alone, and it ships with the reference library, command-line tooling, an executable conformance corpus and a public negative-test corpus in the same release as the prose.

Its differentiators against the incumbents are structural, not featural: octet-stable saves with bounded write cost (which package formats forfeit by repacking), total or explicitly enumerated signature coverage with a pinned signed presentation and per-state attestation across editing (which fixed-layout formats forfeit by letting unsigned appended octets render), authored character sequences and reading order (which fixed-layout formats forfeit by painting glyphs), identity-anchored annotations (which every offset model forfeits on the first concurrent edit), and rendering determinism closed end to end (pinned Unicode version, pinned shaping, in-document line-breaking data, unhinted outlines and one colour representation), so that two independent implementations produce the same pixels rather than merely the same intent.

---

## 3. Target users

- Document authors who need a file whose saved octets change only where they edited, so version control, review and audit work on documents the way they work on code.
- Reviewers and collaborators who annotate, comment and suggest edits and need those anchors to survive other people's concurrent edits to the surrounding text.
- Engineers building document pipelines (indexers, search, data-loss-prevention scanners, mail gateways, backup and sync clients) who need text, metadata and a safety verdict at bounded cost without implementing a renderer.
- Compliance, legal and records staff who must prove a published document contains no removed content and no author identity, and that a signature covers everything a reader can see.
- Independent implementers building a second reader or writer from the specification text alone, at a role scope they can actually finish.
- Assistive-technology users and the accessibility auditors who certify documents for them, who need reading order, semantics and text alternatives as authored data rather than inferred structure.
- Archivists and preservation institutions who must render and verify a document decades after the software that produced it is gone.

---

## 4. User stories

| ID | As a | I want | So that |
|----|------|--------|---------|
| US-001 | document author using version control | a one-word edit to a large document to produce a small, readable change and an otherwise octet-stable file | I can review, diff, merge and sync documents without transferring or re-reading the whole file |
| US-002 | reviewer | my comment to stay attached to the sentence I commented on after other people edit the paragraphs around it, and to remain readable with its quoted text if that sentence is deleted | review history is not silently destroyed by ordinary editing |
| US-003 | recipient of a signed contract | a verifier that tells me exactly which octets and which rendering the signature covers, which state it covers, and what has changed since | a valid-signature indicator means the content I am reading is the content that was signed |
| US-004 | compliance officer publishing a redacted document | an operation that removes content from the output octets, proves no residue remains, and preserves the original signature over what is left with the omissions declared | redaction is verifiable without forcing a choice between a redacted document and a signed one |
| US-005 | search-platform engineer | ordered text, per-unit language tags and always-resolvable locators from a document without implementing fonts, shaping or layout, at a stated cost ceiling on a stated reference machine | documents are indexed correctly and cheaply on every machine in the fleet |
| US-006 | security engineer running a scanning gateway | to conclude from a bounded prefix read that no construct in the document is executed or dereferenced by a reader, and to see every opaque payload's kind, length and digest so I can apply my own policy | documents can be cleared at line rate without the scanner being told something false about payloads it cannot see |
| US-007 | screen-reader user | authored reading order, headings, table header associations and a real text alternative for every non-decorative object | the document reads correctly without a tool inferring structure from visual formatting |
| US-008 | reader on a phone opening a 1 GB document over a mobile link | the first page to appear within a stated time and any page to be reachable without downloading the file | large documents are usable on constrained devices |
| US-009 | independent implementer | a specification I can implement at a bounded role scope from the text alone, plus an executable conformance corpus and a negative corpus with expected verdicts | a second conforming implementation exists and the format is not defined by one program's behaviour |
| US-010 | archivist accessioning a document for thirty years | the document to render identically with no host fonts, no host text stack and no network, and its signature to remain decidable after the issuing authority is gone | the artefact is preservable and its authenticity is still assessable in decades |
| US-011 | team merging two copies of a document that were edited offline | the merge to find the common ancestor from the two files alone, converge identically in any delivery order, keep each author's typed passage unbroken, and stop rather than guess on a real conflict | offline collaboration does not silently lose or interleave work |

---

## 5. Scope

### 5.1 In scope (v1)

- A single document format specification covering container framing and unit addressing, the abstract content model (structure, text, annotations, tables, notes, cross-references), text semantics (language, direction, reading order, accessibility roles), embedded resource carriage, and integrity and signature carriage.
- Definitions that every other requirement quantifies over: document state, canonical form, state identity and ancestry, the acted-upon octet set, reader conformance roles, and the normative table of numeric ceilings.
- Canonical octets computable without rewriting the file, with bounded per-edit write cost, bounded novel-chunk output, and explicit non-implicit compaction.
- Both an authored fixed pagination and a deterministic reflowable presentation from one file, with identical text and reading order, and one presentation pinned into any signature.
- Identity-based anchoring for every annotation, comment, change record and cross-reference, with declared boundary behaviour and a stated orphan resolution.
- Asynchronous merge of copies that diverged offline, with specification-defined determinism, non-interleaving, a named move-cycle resolution and explicit conflict surfacing.
- Whole-document integrity with total or enumerated signature coverage, per-state attestation across editing, in-file long-term validation evidence including the timestamp authority's own material, and a defined re-protection path.
- A redaction-compatible signature construction with hiding commitments, a distinct attested-with-declared-omissions verdict, and a publish operation producing octet-absent removal and identity stripping while retaining custody and fixity.
- An extraction view (text, structural locators, language, metadata) obtainable without fonts, shaping, layout or graphics, plus a bounded-prefix preview payload and a bounded-prefix inspection inventory bound to integrity protection.
- Rendering determinism as a closed system: pinned Unicode version, pinned shaping, in-document line-breaking and hyphenation data, unhinted outline evaluation, one colour representation, and exact 300 dpi raster equality.
- A durable profile: self-contained rendering with no host fonts and no network, and no construct that executes or resolves outward.
- Deterministic, identity-preserving migration between format major versions with per-state signature semantics.
- A reference library and a command-line tool providing validate, inspect, extract, verify, diff, merge, project, redact, publish, sign and migrate, plus a deterministic text projection outside the conformance surface.
- Conformance corpus, negative corpus, per-role conformance suites, continuous fuzzing, and a governance model covering licence, steward, succession, token registry and deprecation window.

### 5.2 Explicitly out of scope (v1)

- **Import and export converters against the incumbent formats, with per-construct loss reports.** Reading either incumbent well enough to enumerate everything dropped is a product in its own right; v1 carries the inferred-value marker the format needs so converters can be built later without a format change, and the converters themselves are a named post-v1 milestone with their own requirement identifiers.
- **Real-time, sub-second multi-author editing with presence, cursors and a live transport.** v1 specifies asynchronous merge of divergent copies and the identity and anchoring primitives a live engine would need; the live engine is deferred.
- **A formula or calculation profile.** v1 tables carry typed cell values only: no formula language, no evaluation, no dependency graph, no recalculation semantics.
- **Interactive forms and a submission model.** No form-field kinds, no value layer, no declarative validation, and no submission action, since every documented submission construct in the incumbents is an activation primitive.
- **Charts as a live construct bound to source data.** A chart in v1 is an embedded object with a rendered representation, a text inventory and a value-equivalent table region.
- **Print exchange.** Device colour, separations and spot colour, output intents, page boxes beyond a single page geometry, overprint, halftones, imposition and finishing are all excluded.
- **Encryption and confidentiality.** v1 defines no protected-content model; documents are integrity-protected and signable but not encrypted, and encryption granularity interacts with deduplication, partial fetch and the inspection inventory and needs its own threat model.
- **A records-management profile.** Retention classes, disposition dates, legal-hold state and an append-only custody chain as first-class content are deferred; v1 carries an enumerable provenance inventory and publish semantics.
- **Archival-institution accepted-format listing as a v1 gate.** Media-type and format-identification registration is a v1 requirement; preservation-list acceptance depends on demonstrated adoption and is a 5-10 year goal.
- **Vector transforms.** Persisted rotation, scaling and nested affine composition are excluded so that integer geometry survives; v1 persists transformed control points and integer translation only.
- **Math notation, ruby and vertical writing modes, and bibliography or citation management as first-class constructs.** Each is acknowledged as real demand and each is a named later profile.
- **Beating either incumbent on breadth.** v1 deliberately ships fewer capabilities than both, in exchange for properties neither has: octet-stable saves, bounded edit cost, total signature coverage, authored text and reading order, identity-based anchoring, deterministic rendering, and a specification two independent teams can implement.

---

## 6. Requirements

Every requirement below is normative and carries priority **must**. EARS pattern is given in parentheses after each identifier.

### 6.1 Functional (FR-*)

**FR-001** *(ubiquitous, must)*

> The Protodoc format SHALL define the term document state as the complete set of values that determine extraction output, rendered output, validation verdict and reported metadata, such that two states differing in any one of those values are distinct states.

- **Why:** Byte-stability, signing, content addressing, history reconstruction and merge all quantify over "state" or "canonical octets"; every contradiction between them traces to the word meaning something different in each place. It cannot be settled once implementations exist.
- **Verify:** Audit A-STATE: every normative statement using the term resolves to this definition; CI fails on any undefined use. Corpus C-STATE pairs differing in exactly one value are classified distinct by two implementations.

**FR-002** *(ubiquitous, must)*

> The Protodoc format SHALL define the acted-upon set of a document as every octet that can influence extraction output, rendered output, validation verdict or reported metadata.

- **Why:** Signature coverage and the unverified verdict both turn on which octets a reader can act on. Without a definition the coverage criterion is not computable and two verifiers reach different verdicts on one file.
- **Verify:** Audit A-ACTED: the acted-upon set is computed by two implementations for every corpus document and compared; zero divergences.

**FR-003** *(ubiquitous, must)*

> The Protodoc format SHALL assign every document state an identifier that differs whenever any value of that state differs.

- **Why:** Verification, merge and history reconstruction must never confuse two states. An identifier that collides across states makes a signature's subject ambiguous.
- **Verify:** Test T-STATEID over the corpus and 10,000 mutated variants: zero identifier collisions between distinct states, identical identifiers for identical states across two implementations.

**FR-004** *(ubiquitous, must)*

> The Protodoc format SHALL record in every state the identifiers of the states it derives from.

- **Why:** Three-way merge, concurrent-write detection and per-state attestation all require ancestry; a merge tool with no specified way to find a common ancestor from two files cannot be written.
- **Verify:** Test T-ANCESTRY: for 1,000 fork-and-merge traces, recorded predecessors match the generating trace for 100% of states.

**FR-005** *(ubiquitous, must)*

> The Protodoc format SHALL permit the nearest common ancestor state of two divergent copies to be determined from those two files alone, without a coordinating service and without network access.

- **Why:** The archetypal case in scope is two copies emailed or synced offline. If the ancestor is only discoverable from a server, offline three-way merge is impossible.
- **Verify:** Test T-COMMON: 1,000 divergent pairs; the computed ancestor equals the generating trace's ancestor for 100%, under a syscall filter recording zero network calls.

**FR-006** *(event-driven, must)*

> WHEN a consumer has read the leading 512 octets of a document, the Protodoc format SHALL permit it to determine the format identity, the document class and the format major version.

- **Why:** Fixed-offset identification is what makes a file recognisable to operating-system type databases, mail gateways, object stores and preservation pipelines. Obtaining it by convention over a generic archive broke silently whenever a writer varied its output.
- **Verify:** Corpus check C-ID under a reference sniffer that faults on any read past octet 512: 100% identification, zero reads beyond the window.

**FR-007** *(ubiquitous, must)*

> The Protodoc format SHALL exclude from the leading 512 octets every document text value, every user-supplied metadata value and every value derived from document content.

- **Why:** An identification window carrying content leaks it to every intermediary that sniffs the file and makes the window vary per document, defeating pattern registration.
- **Verify:** Static audit A-PREAMBLE over the schema plus corpus scan: zero content-derived or user-supplied values in the window for 100% of documents.

**FR-008** *(event-driven, must)*

> WHEN a consumer has read the leading 512 octets of a document, the Protodoc format SHALL permit it to determine the highest capability generation used to write that document.

- **Why:** A writer must be able to state what it used, so a consumer can decide whether older software still reads the file without opening it.
- **Verify:** Corpus check C-VER-W: the declared written generation equals the generating writer's generation for 100% of corpus documents, read within the 512-octet window.

**FR-009** *(event-driven, must)*

> WHEN a consumer has read the leading 512 octets of a document, the Protodoc format SHALL permit it to determine the lowest capability generation a reader must implement to render the document without applying any disposition.

- **Why:** Two integers turn proceed-or-refuse into an arithmetic comparison instead of a judgement call. A single monotonic version leaves readers guessing, which is how a partially rendered contract becomes indistinguishable from a complete one.
- **Verify:** Corpus check C-VER: the reference reader's proceed-or-refuse decision equals the arithmetic comparison for 100% of documents; zero divergences.

**FR-010** *(unwanted-behavior, must)*

> IF a document declares a minimum required capability generation higher than its declared written capability generation, THEN a conforming reader SHALL reject the document naming both declared values.

- **Why:** The two integers are only meaningful as an ordered pair; an inverted pair is a malformed declaration that would otherwise be interpreted differently by each reader.
- **Verify:** Negative corpus N-VER of 20 inverted-pair documents: 100% rejected with both values named, identical verdicts across two implementations.

**FR-011** *(event-driven, must)*

> WHEN a consumer has read the leading 512 octets of a document, the Protodoc format SHALL permit it to determine whether the document claims the durable profile.

- **Why:** The durable profile is referenced normatively by rendering and outward-reference rules, but an archival ingest pipeline cannot decide conformance to it unless the claim is in the file at a fixed, cheap location.
- **Verify:** Corpus C-DURCLAIM: the claim read from the window equals the writer's declared profile for 100% of documents; durable-profile accept and reject cases both present in the negative corpus.

**FR-012** *(ubiquitous, must)*

> The Protodoc format SHALL require every construct not defined by the core specification to be carried in one universal extension envelope declaring a registered identifier, the payload's total octet length, a disposition value, a fallback reference and a digest over the payload.

- **Why:** Every extensibility obligation presupposes that a reader can find a future construct's extent, disposition and fallback without understanding it. With no mandated envelope, a future construct is unparseable and the extensibility model collapses to old readers refusing new files.
- **Verify:** Validator rule PD-EXT-001 with corpus C-FUTURE: a previous-generation reader determines extent, disposition and fallback for 100% of future constructs; any non-enveloped extension construct is rejected.

**FR-013** *(ubiquitous, must)*

> The Protodoc format SHALL require every extension envelope to declare a disposition selected by the writer from exactly three values: ignore, degrade to the accompanying fallback, or refuse to render.

- **Why:** Writer-authored per-construct handling policy is the one idea worth keeping from the incumbent compatibility layer. A blanket reader-side skip-unknown rule produced decades of mutually incompatible dialects; the writer knows when ignoring changes meaning, the reader never does.
- **Verify:** Corpus N-EXT: one document per disposition value per extension point; observable reader output equals the single defined output for that disposition in 100% of cases.

**FR-014** *(unwanted-behavior, must)*

> IF an extension envelope carries no disposition value, THEN a conforming reader SHALL reject the document naming the envelope's registered identifier.

- **Why:** No reader-selected default is defined for an unrecognised construct, so an envelope without a disposition is undefined behaviour at exactly the version boundary the format claims to solve, and two implementations would choose differently.
- **Verify:** Negative corpus N-EXT-NODISP of 25 documents: 100% rejected with the identifier named, identical verdicts across two implementations.

**FR-015** *(optional-feature, must)*

> WHERE an extension envelope declares the ignore or degrade disposition, the Protodoc format SHALL require an accompanying fallback expressible entirely in the core feature set.

- **Why:** The incumbent markup format has this facility and leaves it optional, so producers ship an empty fallback or none; an older reader then drops the content silently and the user saves the loss back.
- **Verify:** Validator rule PD-EXT-002 with corpus N-EXT-EMPTY: every ignore- or degrade-disposition envelope lacking a fallback is rejected; zero false accepts.

**FR-016** *(optional-feature, must)*

> WHERE an extension envelope declares a fallback, the Protodoc format SHALL require that fallback to yield at least one extractable text unit or at least one mark classified as non-decorative.

- **Why:** A non-emptiness rule alone is satisfied by a single space or a zero-area object, so the drafted requirement did not deliver graceful degradation. A minimum-content criterion is what a validator can actually decide.
- **Verify:** Validator rule PD-EXT-003 with corpus N-EXT-THIN: whitespace-only and zero-extent fallbacks are rejected; zero false accepts.

**FR-017** *(complex, must)*

> WHEN a conforming writer saves a document, IF that document contains constructs the writer does not implement, THEN the writer SHALL reproduce those constructs octet-for-octet in the output or refuse the save naming each construct it could not preserve.

- **Why:** Nothing in the incumbent obliges an editor that ignored an extension region to keep it, so third-party data is destroyed by open-and-save with no diagnostic. A widely used encoding silently discarded unknown fields for about two years, corrupting read-modify-write in every mixed-version fleet.
- **Verify:** Round-trip corpus R-FUT: 50 future-generation documents opened and saved by a previous-generation writer; per-construct digests equal before and after, or the save was refused, in 100% of cases.

**FR-018** *(complex, must)*

> WHEN a conforming implementation merges two documents, IF a merge input contains a construct the implementation does not implement, THEN it SHALL carry that construct and its identity through the merge octet-for-octet or refuse the merge naming each construct it could not carry.

- **Why:** Merge is neither a save nor a render, so an older implementation could drop a newer peer's constructs during merge and then satisfy the save rule afterwards, because the constructs are already gone.
- **Verify:** Test T-MERGE-FUT: 50 merges with future-generation constructs in one input; digests preserved or the merge refused, in 100% of cases; zero silent drops.

**FR-019** *(ubiquitous, must)*

> The Protodoc format SHALL assign every independently addressable content unit an identifier that is unique across the document and every copy, fork and branch derived from it, for the lifetime of that lineage.

- **Why:** Document-scoped uniqueness is not enough: two forks minting identifiers independently can produce a legitimate merge in which two units share an identifier, after which duplicate rejection makes merge impossible for exactly the documents that need it.
- **Verify:** Property test P-ID: 1,000 fork-and-merge trials across independently minting replicas plus 10,000 randomised structural edits; zero identifier collisions.

**FR-020** *(ubiquitous, must)*

> The Protodoc format SHALL preserve a content unit's identifier across splitting, merging, moving, reordering, save, load and undo of surrounding content.

- **Why:** Stable identity is the precondition for comments, suggestions, cross-references, permissions and incremental sync; identity that dies on an ordinary edit anchors nothing.
- **Verify:** Property test P-ID-SURVIVE under 10,000 randomised structural edits: no identifier resolves to different content than at the start; zero losses.

**FR-021** *(ubiquitous, must)*

> The Protodoc format SHALL prohibit the reuse of any identifier that has been assigned to a content unit within a lineage, whether or not that unit still exists.

- **Why:** Reuse silently reattaches every annotation, cross-reference and permission entry pointing at the retired unit, converting a detectable corruption into an invisible one.
- **Verify:** Property test P-ID-REUSE over 10,000 delete-and-create cycles: zero reissued identifiers across two implementations.

**FR-022** *(event-driven, must)*

> WHEN a content unit is produced by duplicating or pasting existing content, the Protodoc format SHALL require a freshly minted identifier for the produced unit.

- **Why:** Copy-paste paths that clone identifiers make every anchor resolve to two places at once, which users see as comments appearing on the wrong copy. It is the commonest way identity models break in practice.
- **Verify:** Test T-DUP: 100 duplications of a 500-unit subtree; zero cloned identifiers, every annotation resolves to exactly one unit.

**FR-023** *(ubiquitous, must)*

> The Protodoc format SHALL define identifier minting as a construction that carries no value derived from actor identity, device identity, wall-clock time or editing session, with a stated collision probability bound across uncoordinated concurrent authors.

- **Why:** Uniqueness normally comes from entropy or from an actor identifier plus a counter; the second would be removed by the identity-stripping publish operation, detaching every anchor. Per-session identifiers stamped into content are the specific incumbent behaviour this format rejects.
- **Verify:** Static audit A-MINT over the minting definition plus test T-MINT: 10^8 identifiers minted across 1,000 simulated uncoordinated replicas produce zero collisions and no recoverable actor, device, clock or session value.

**FR-024** *(unwanted-behavior, must)*

> IF a merge input pair contains two distinct content units carrying the same identifier, THEN a conforming implementation SHALL refuse the merge naming both units and their source documents.

- **Why:** A cross-lineage collision cannot be repaired by renaming without detaching every anchor pointing at the renamed unit, and cannot be resolved by precedence without silently discarding one author's content.
- **Verify:** Negative corpus N-MERGEID of 50 colliding pairs: 100% refused with both units named, zero merged outputs, identical verdicts across two implementations.

**FR-025** *(ubiquitous, must)*

> The Protodoc format SHALL address every formatting range, comment, link, cross-reference, bookmark and change record by reference to content identity, and SHALL define no persisted construct that locates a position by a count of text units from the start of a text sequence.

- **Why:** This rejects the origin sketch's offset-span model. With text "Hello world" and a bold range over the first five characters, a remote insertion of four characters at position 0 leaves the range bolding the wrong text, and a file on disk has no transform function available to repair it.
- **Verify:** Property test P-ANCHOR: over 10,000 randomised edits not intersecting an annotated range, the resolved first and last characters of that range are identical before and after in 100% of trials; static audit confirms zero persisted count-based position constructs.

**FR-026** *(ubiquitous, must)*

> The Protodoc format SHALL require every annotated range to declare, for each of its two ends independently, one boundary behaviour from exactly four values: content inserted at that boundary falls inside the range, falls outside the range, falls inside only when inserted before existing content at that position, or falls inside only when inserted after it.

- **Why:** Formatting kinds genuinely differ: typing at the end of a bold run should continue the bold, typing at the end of a hyperlink should not. Leaving it to the renderer makes two conforming editors extend a hyperlink differently on the same keystroke.
- **Verify:** Conformance suite S-BOUND: for each of the four values, one character inserted at each boundary, locally and as a merged remote edit, yields the declared inclusion result in 100% of cases.

**FR-027** *(ubiquitous, must)*

> The Protodoc format SHALL preserve each declared boundary behaviour unchanged across save, load and merge.

- **Why:** A behaviour that survives editing but not serialisation produces a document that formats differently after a round trip through a conforming tool.
- **Verify:** Round-trip test R-BOUND over the annotation corpus: declared boundary values are identical before and after save, load and merge for 100% of ranges.

**FR-028** *(ubiquitous, must)*

> The Protodoc format SHALL retain every annotation whose entire anchored content has been deleted.

- **Why:** Orphaned comments are the visible bug in every collaborative editor: the anchored text is rewritten and the reviewer's comment vanishes. Whether retention happens must not be an application choice, or review history is destroyed by ordinary editing in half the tools.
- **Verify:** Test T-ORPHAN-KEEP: delete the entire anchored range of each of 100 annotations; all 100 remain readable in the saved file across two implementations.

**FR-029** *(complex, must)*

> WHEN all content an annotation refers to has been deleted, IF a surviving content unit follows the deleted range in document order, THEN the Protodoc format SHALL resolve the annotation to the start of the nearest such unit, and otherwise to the end of the nearest surviving preceding unit.

- **Why:** Leaving the resolution undefined means each client picks differently. Silent re-attachment by similarity heuristics is worse than a stated rule, because a reader then believes a reviewer approved text the reviewer never saw.
- **Verify:** Test T-ORPHAN-POS: 100 orphaned annotations resolve to the specified position deterministically and identically across two implementations.

**FR-030** *(optional-feature, must)*

> WHERE an annotation has been orphaned, the Protodoc format SHALL preserve its author, its recorded quoted text and the identifiers of its nearest surviving preceding and following units.

- **Why:** An orphan with no quoted text and no neighbours is unreviewable; preserving them is what lets a human decide whether the comment still applies.
- **Verify:** Test T-ORPHAN-DATA: for 100 orphaned annotations, author, quoted text and both neighbour identifiers are present and correct in the saved file; zero losses.

**FR-031** *(ubiquitous, must)*

> The Protodoc format SHALL resolve exactly one language tag for every span of document text, and SHALL treat a document containing a span with no resolvable language tag as invalid.

- **Why:** Language is a rendering input. Under unified ideographic encoding one code point selects visibly different glyphs for four language communities; casing, hyphenation, line breaking, quotation marks and digit shaping are all locale-selected. An unknown-language path that still renders makes two machines produce different glyphs from identical octets.
- **Verify:** Validator rule PD-LANG-001 with corpus C-I18N: every text span resolves to exactly one tag; documents with an unresolvable span are rejected, zero false accepts.

**FR-032** *(ubiquitous, must)*

> The Protodoc format SHALL require every text container to declare its base writing direction explicitly.

- **Why:** Inferring base direction from the first strong character makes an empty or digit-initial paragraph render differently in two conforming readers, and changes as the user types.
- **Verify:** Validator rule PD-BIDI-001: every text container in the corpus declares a base direction; documents lacking one are rejected, zero false accepts.

**FR-033** *(ubiquitous, must)*

> The Protodoc format SHALL express directional scope as a structural relationship between content units rather than as characters within a text sequence.

- **Why:** The Unicode bidirectional algorithm deprecated its embedding controls in favour of isolates because embeddings leak ordering into neighbouring text. In-band controls also mean an identity-based edit can split a control pair, and unbalanced overrides are the mechanism of a published source-hiding attack class.
- **Verify:** Static audit A-BIDI plus corpus C-BIDI: zero in-band directional control constructs defined; rendered visual order identical across two implementations for 500 mixed-direction documents.

**FR-034** *(ubiquitous, must)*

> The Protodoc format SHALL isolate every inline object whose value is computed, such that the directionality of that value cannot alter the visual order of surrounding text.

- **Why:** A page number or citation whose value changes must not reorder the punctuation around it; without isolation, re-materialising a reference changes the appearance of text nobody edited.
- **Verify:** Test T-ISOLATE: for every computed inline object in corpus C-BIDI, replacing its value with one of different directionality leaves surrounding visual order identical; zero divergences.

**FR-035** *(ubiquitous, must)*

> The Protodoc format SHALL require that extracting a document's text reproduces the authored Unicode scalar sequence, including word boundaries, in reading order.

- **Why:** The fixed-layout incumbent paints glyph codes at coordinates and makes the reverse mapping optional, so extraction is permanently heuristic, which is the reason copy-paste yields mangled ligatures and invented spaces. Search, indexing, redaction verification and assistive technology all fail silently when the character layer is derived rather than authored.
- **Verify:** Round-trip corpus R-TEXT of 5,000 documents including ligatures, contextual forms and reordered Indic clusters: extracted scalars equal writer input scalar-for-scalar with no normalising step; CI fails on any mismatch.

**FR-036** *(ubiquitous, must)*

> The Protodoc format SHALL require logical reading order to be an authored property of every document.

- **Why:** Reading order inferred from coordinates is wrong on multi-column pages, sidebars and tables, and neither the author nor the consumer can tell when it is wrong.
- **Verify:** Validator rule PD-A11Y-001: every document declares a total reading order over its content units; documents lacking one are rejected, zero false accepts.

**FR-037** *(ubiquitous, must)*

> The Protodoc format SHALL require every rendered mark to be either placed in the document's reading order or explicitly declared decoration excluded from it.

- **Why:** The content-versus-decoration distinction was retrofitted onto the fixed-layout incumbent and unmarked content remains among its commonest accessibility failures. Making the classification exhaustive converts an editorial judgement into a machine-decidable property.
- **Verify:** Validator rule PD-A11Y-002: documents containing an unclassified mark are rejected, zero false accepts, identical verdicts across two implementations.

**FR-038** *(ubiquitous, must)*

> The Protodoc format SHALL map every defined structural element kind, annotation kind, tabular construct and embedded object kind onto exactly one role in a published closed set of accessibility roles and relationships.

- **Why:** The markup incumbent infers headings from a style name, so direct-formatted large bold text is a heading to the eye and nothing to assistive technology; the fixed-layout incumbent's optional parallel structure tree drifts out of agreement with the visible content with nothing validating the binding.
- **Verify:** Traceability audit A-ROLE: the mapping table covers every defined construct with exactly one role; CI fails on any unmapped or doubly mapped construct.

**FR-039** *(ubiquitous, must)*

> The Protodoc format SHALL require every table header cell to declare the scope of the cells it heads.

- **Why:** Header association inferred from position fails on stub columns, spanned headers and nested tables, and is the difference between a navigable table and an unreadable grid of numbers.
- **Verify:** Validator rule PD-A11Y-003 with corpus C-TBL-HDR: header cells lacking a declared scope are rejected; two implementations resolve identical header associations for every data cell.

**FR-040** *(ubiquitous, must)*

> The Protodoc format SHALL require every non-decorative non-text object to carry a text alternative that is non-empty and distinct from the object's identifier, its stored component name, its source file name and its kind name.

- **Why:** An unconstrained alternative field is silenced with a filename or an empty string, which is how real documents pass accessibility checkers today, and it would inflate the machine-decidable conformance share with the commonest failure it claims to catch.
- **Verify:** Validator rule PD-A11Y-004 with negative corpus N-ALT of empty, whitespace-only, filename-echo and kind-name alternatives: 100% rejected, zero false accepts.

**FR-041** *(ubiquitous, must)*

> The Protodoc format SHALL admit an extraction view yielding a document's text in reading order without resolving fonts, performing shaping, computing layout or evaluating the graphics model.

- **Why:** Enterprise consumption is mediated by extraction libraries and operating-system indexing interfaces, none of which will implement a renderer. A format offering only parse-me-fully forces every integrator onto the most expensive contract, which is why one vendor's extraction component was routinely disabled and its documents went unindexed.
- **Verify:** Reference extractor built against this view only, with no font or graphics dependency, reproduces writer input text exactly for 100% of the 5,000-document corpus.

**FR-042** *(ubiquitous, must)*

> The Protodoc format SHALL require the extraction view to emit, for every text unit it emits, a structural locator composed of the unit's identity and a scalar position within that unit.

- **Why:** A locator that depends on pagination is unavailable exactly when the cached layout is stale or absent, which is most of the time for an indexer; a structural locator is always resolvable and always current.
- **Verify:** Test T-LOCATOR-S: for 10,000 randomly chosen substrings across the corpus, the emitted structural locator resolves to a character range equal to the source range; CI fails on any mismatch.

**FR-043** *(ubiquitous, must)*

> The Protodoc format SHALL require the extraction view to emit exactly one language tag for every text unit it emits.

- **Why:** A widely deployed extraction interface emits a locale per text chunk rather than per document, confirming multilingual documents are the normal case; per-document language makes indexing and language-aware search wrong on the common case.
- **Verify:** Test T-EXT-LANG: emitted per-unit tags equal writer input tags for 100% of the corpus; zero unresolved units.

**FR-044** *(ubiquitous, must)*

> The Protodoc format SHALL require the extraction view to emit the document's descriptive metadata without resolving fonts, performing shaping, computing layout or evaluating the graphics model.

- **Why:** The dominant search stack separates give-me-text, give-me-metadata and give-me-pixels into three contracts with three cost profiles; metadata that costs a full parse is metadata nobody reads.
- **Verify:** Test T-EXT-META: emitted metadata equals writer input for 100% of the corpus under a harness that faults on font, layout and graphics entry points.

**FR-045** *(optional-feature, must)*

> WHERE a current pagination artefact is present, the Protodoc format SHALL permit the extraction view to emit a page adjunct on each locator identifying the page on which the located characters are rendered.

- **Why:** Search without hit highlighting is search users do not trust, and every enterprise search product must map an index position to a visual position. Making the page component conditional is what keeps extraction free of layout while still serving highlighting when the artefact is current.
- **Verify:** Test T-LOCATOR-P over documents with a current pagination artefact: resolving each page adjunct through the reference renderer yields the extracted characters on the named page for 100% of cases.

**FR-046** *(unwanted-behavior, must)*

> IF no current pagination artefact is present, THEN the extraction view SHALL emit structural locators with the page adjunct absent and report the document's pagination as stale.

- **Why:** Without this the three requirements collide: page-ordered output demands pagination, extraction may not compute it, and a stale stored artefact must be refused, leaving no defined behaviour at all.
- **Verify:** Test T-LOCATOR-STALE: with the pagination artefact absent or digest-mismatched, extraction succeeds, page adjuncts are absent and pagination is reported stale for 100% of cases.

**FR-047** *(event-driven, must)*

> WHEN a consumer extracts text from a document, the Protodoc format SHALL permit the complete text and locators of every content unit preceding a given point in reading order to be emitted before any octet belonging to a later unit is read.

- **Why:** Mail previewers, loss-prevention gateways and cloud extractors work against byte-range or streaming inputs. A format requiring a trailing index or a whole-file pass before the first character forces a full download in every one of those pipelines.
- **Verify:** Test T-STREAM drives extraction through a reader that faults on any read beyond the highest offset needed for the first N units; must succeed for N in {1, 10, 100, 10000} on the synthetic corpus.

**FR-048** *(event-driven, must)*

> WHEN a consumer abandons extraction after any emitted unit, the Protodoc format SHALL impose no obligation to read the remainder of the file.

- **Why:** Scanners routinely stop early on a match or a budget; a format that requires reading to the end to finish cleanly makes early abandonment indistinguishable from failure.
- **Verify:** Test T-ABANDON: extraction abandoned after unit N under a reader that faults on further reads; succeeds and reports a complete prefix for N in {1, 10, 100}.

**FR-049** *(state-driven, must)*

> WHILE a consumer possesses no signature trust material and performs no verification, a conforming reader SHALL produce complete extraction output.

- **Why:** No indexer, mail previewer or backup scanner holds signer trust anchors, and none will fail a document because a chain cannot be built offline. If verification gates extraction, documents become unindexable in the environments that most need search.
- **Verify:** Test T-NOTRUST: extraction in a process with an empty trust store and no network succeeds and reproduces writer input text for 100% of the corpus.

**FR-050** *(ubiquitous, must)*

> The Protodoc format SHALL require verification status to be reported as a value distinct from extracted content, taking one of the defined verdicts.

- **Why:** Keeping the status separate lets a downstream system require verification when it can afford it, and prevents a tampered document from extracting as though nothing were wrong.
- **Verify:** Test T-STATUS: a tampered file extracts and reports verification-failed, an untrusted file reports unverified, never ok; zero misreports across two implementations.

**FR-051** *(ubiquitous, must)*

> The Protodoc format SHALL require every document to carry, within its leading 262144 octets, a preview payload rendering the first page at no more than the stated maximum pixel dimensions and no more than the stated maximum octet length.

- **Why:** A representative preview cannot be improvised from a bounded prefix when a first page's fonts or artwork alone exceed it. Requiring the writer to emit a bounded payload makes the property a validity rule rather than a reader-side gamble that yields a different picture in every tool.
- **Verify:** Validator rule PD-PREV-001 plus test T-PREFIX under a reader that faults past octet 262144 and on any socket call: payload present and octet-identical across two implementations for 100% of corpus documents including the 10,000-page synthetic.

**FR-052** *(ubiquitous, must)*

> The Protodoc format SHALL bind the preview payload to a digest over the complete set of inputs it was rendered from.

- **Why:** A stored cover image is a second representation of the content; cached previews have repeatedly displayed content that was subsequently redacted or replaced. A digest makes staleness detectable rather than a privacy incident.
- **Verify:** Test T-PREV-STALE: mutating any recorded input class classifies the payload stale in 100% of cases and current in zero.

**FR-053** *(unwanted-behavior, must)*

> IF a consumer restricted to the leading 262144 octets finds the preview payload's input digest mismatched, THEN it SHALL report the preview as stale rather than displaying it or deriving a replacement.

- **Why:** The general rule to refuse a stale artefact and re-derive from authoritative content is unreachable inside a bounded prefix, leaving a prefix consumer with no conforming behaviour at all. A third defined outcome closes it.
- **Verify:** Test T-PREV-VERDICT: 100 documents with mutated inputs and unrefreshed payloads; the prefix consumer reports preview-stale in 100% of cases, renders nothing, reads no further.

**FR-054** *(event-driven, must)*

> WHEN a consumer has read the leading 262144 octets of a document, the Protodoc format SHALL permit it to determine the document's title, page count, page dimensions and language without network access.

- **Why:** Mail gateways, webmail preview panes, object stores and preview hosts issue one bounded range request in a network-denied sandbox with a wall-clock kill; a killed generator yields a generic icon permanently and users conclude the format is broken.
- **Verify:** Test T-PREFIX-META: the four reported values equal writer input for 100% of corpus documents, under a reader that faults past octet 262144 and on any socket call.

**FR-055** *(ubiquitous, must)*

> The Protodoc format SHALL permit a reader to locate and read any single addressable unit using at most 3 sequential dependent reads and no more than the greater of 1048576 octets or 1% of the file size beyond the unit itself, for documents up to 1073741824 octets.

- **Why:** A columnar analytics format meets this with a fixed-size trailer carrying an explicit length; the dominant package container does not, because a variable-length trailing comment forces a backwards scan and different parsers select different candidate records. Bounded round trips are what make a large document openable over a range-request transport.
- **Verify:** Benchmark B-RANDOM against a range-serving harness counting dependent reads and octets: for units on pages 1, 5000 and 10000 of the synthetic corpus, both ceilings hold; CI gate on regression.

**FR-056** *(event-driven, must)*

> WHEN a consumer replaces one addressable unit of a document, the Protodoc format SHALL leave the octets of every stored extent containing no modified unit unchanged.

- **Why:** Per-file transfer is what makes cloud sync expensive; per-extent transfer is what makes it cheap and lets a server answer what changed without a parse.
- **Verify:** Byte-diff test T-COMP over the edit corpus: untouched extents are octet-identical before and after for 100% of operations.

**FR-057** *(event-driven, must)*

> WHEN a consumer replaces one addressable unit of a document, the Protodoc format SHALL bound the combined change to index and integrity data at 262144 octets.

- **Why:** Bounding index churn explicitly is what prevents a single authoritative index from silently reintroducing whole-file rewrite, the measured difference between re-uploading 44 KiB and 63 MB on a 500 MB document.
- **Verify:** Byte-diff test T-INDEX over the edit corpus: index and integrity delta under 262144 octets for every operation; distribution reported per edit class.

**FR-058** *(ubiquitous, must)*

> The Protodoc format SHALL assign every stored extent a position in canonical storage order from a monotonically issued ordinal, and SHALL derive that order from no unit's name, digest or content.

- **Why:** If canonical order follows names or digests, changing one unit moves it in sort order and shifts every subsequent extent, violating the unchanged-extent rule and putting every save into the whole-file-rewrite regime that makes a format un-syncable.
- **Verify:** Test T-ORDER: for 1,000 edits including renames and content changes, the ordinal sequence of unmodified extents is unchanged and their offsets are unchanged; zero re-layouts.

**FR-059** *(optional-feature, must)*

> WHERE a document declares the complete history mode, the Protodoc format SHALL permit every state that document has previously been published in to be reconstructed octet-for-octet from the current file alone.

- **Why:** Unconditional reconstruction contradicts both the no-history mode and in-place removal, so a document could satisfy at most two of the three. Conditioning it on the declared mode is what makes the guarantee true where it is claimed.
- **Verify:** Test T-HIST-FULL: for each complete-history corpus document with N published states, reconstruct all N and compare digests against recorded state identifiers; 100% match.

**FR-060** *(optional-feature, must)*

> WHERE a document declares the history-retained-from-a-declared-point mode, the Protodoc format SHALL permit every state published at or after that point to be reconstructed octet-for-octet from the current file alone.

- **Why:** The default working mode must still support review, audit and three-way merge over the document's working life without carrying a permanent erasure liability.
- **Verify:** Test T-HIST-PART: states at or after the declared point reconstruct with matching digests for 100% of documents; states before it are not reconstructed.

**FR-061** *(ubiquitous, must)*

> The Protodoc format SHALL require a document to enumerate the identifier and a salted-commitment digest of every previously published state that it can no longer reconstruct.

- **Why:** A severed state must be distinguishable from a state that never existed, or a verifier presented with a signature over an earlier state cannot say whether the file is the wrong one or the history was lawfully trimmed. Amended 2026-09-15 (Eyvar's ruling, recorded in clarify.md): the original text required a bare digest, which conflicted with FR-075's 2^80 hiding-floor requirement — an unsalted digest is a brute-force oracle over the small candidate space of real-world severed-state identifiers, exactly the exposure FR-075 exists to close. The salted-commitment form (already plan.md's proposed resolution, already what tasks.md builds) is now the frozen text; no implementation changes as a result of this amendment, since the code was already built against this form.
- **Verify:** Test T-SEVER: after publish and after trimming, every severed state is enumerated with its identifier and salted-commitment digest; two implementations produce identical enumerations and identical digests for identical (salt, state) inputs.

**FR-062** *(unwanted-behavior, must)*

> IF a signature covers a state the current file cannot reconstruct, THEN a conforming reader SHALL report that signature as covering an unavailable state rather than as failed or as valid.

- **Why:** Reporting a lawful trim as a verification failure trains users to ignore failures; reporting it as valid attests to content nobody can inspect. A third verdict is the only honest answer.
- **Verify:** Negative corpus N-SEVER of 50 signed-then-trimmed documents: 100% report the unavailable-state verdict, zero failed verdicts, zero positive indicators, identical across two implementations.

**FR-063** *(ubiquitous, must)*

> The Protodoc format SHALL require every signature to cover either the document's complete acted-upon set, or an explicitly enumerated covered subset together with the complete enumeration of the uncovered remainder.

- **Why:** Signer-chosen coverage lists are the shared root cause of both incumbent signature failures: 2023 work found seven attacks in five classes against signed markup documents with every tested version vulnerable, and the appended-update model spoofed 21 of 22 desktop viewers in 2019 and broke 16 of 29 in 2020.
- **Verify:** Negative corpus N-SIG of at least 100 adversarial documents covering each named attack class: a conforming verifier reports total or enumerated-partial coverage for 100%, with zero unqualified valid verdicts.

**FR-064** *(event-driven, must)*

> WHEN a document is signed, the Protodoc format SHALL bind the signature to exactly one presentation artefact identifying the presentation profile version, the page geometry and the identity of every font used.

- **Why:** A signature cannot attest to a rendering recomputed per viewport. The exploitable gap in the fixed-layout incumbent is content the viewer shows that the signature does not cover, which 2020 shadow-attack work turned into working forgeries against more than half the viewers tested.
- **Verify:** Test T-PIN: every signed corpus document carries exactly one bound presentation artefact with all three values; validator rejects any signature lacking one.

**FR-065** *(unwanted-behavior, must)*

> IF a reader presents a signed document in any presentation other than the one bound by its signature, THEN it SHALL report that presentation as not attested by that signature.

- **Why:** Without this the reflowable view inherits the signer's name, and the question of what the signatory saw has no answer.
- **Verify:** Test T-WYSIWYS: verifying a signed document at three viewport widths reports the reflowed view unattested in all three and the pinned presentation attested; identical across two implementations.

**FR-066** *(event-driven, must)*

> WHEN a writer commits an edit to a document carrying a signature, it SHALL preserve that signature bound to the state identifier it covers together with that state's pinned presentation artefact.

- **Why:** Editing after signing is the normal case for an append-shaped format, and discarding the signature destroys the record that the earlier state was attested at all.
- **Verify:** Test T-SIGN-EDIT: 200 signed documents edited once each; the signature, its covered state identifier and its pinned presentation survive octet-identically in 100% of cases.

**FR-067** *(ubiquitous, must)*

> The Protodoc format SHALL require a reader to report attestation per state, naming the signed state, its signer and its pinned presentation, and reporting the current state as not attested when it differs from a signed state.

- **Why:** Applied naively, the first keystroke after signing turns a correctly signed contract into an unsigned one with no way to say what was signed. Per-state reporting keeps the earlier attestation legible without ever showing a signer's name over edited content.
- **Verify:** Test T-ATTEST: for 200 signed-then-edited documents, two implementations produce identical per-state attestation reports and zero document-level positive indicators.

**FR-068** *(ubiquitous, must)*

> The Protodoc format SHALL permit a reader to enumerate every content unit that differs between a signed state and the current state.

- **Why:** Telling a recipient that the current state is unattested is useless without telling them what changed; this is the difference between a warning and an answer.
- **Verify:** Test T-SIGDIFF: for 200 signed-then-edited documents, the enumerated differing units equal the units the generating trace modified; zero omissions, zero spurious entries.

**FR-069** *(optional-feature, must)*

> WHERE a presentation artefact is bound to a signed state, the Protodoc format SHALL exempt it from staleness refusal relative to the current state.

- **Why:** Its inputs are the signed state, not the current one, so the general stale-artefact refusal would destroy the attested rendering the moment anyone edits the document.
- **Verify:** Test T-PIN-STALE: after editing a signed document, the pinned artefact remains servable and is reported current relative to its signed state in 100% of cases.

**FR-070** *(ubiquitous, must)*

> The Protodoc format SHALL carry inside the document, for every signature, the signatory's credential chain, the revocation evidence current at the signing instant, a time attestation from a timestamping authority and a signing-intent value from a closed set.

- **Why:** Long-term validity is a data-locality problem: the revocation responder and issuing authority will not answer in thirty years, which is why long-term signature profiles escalate to embedded validation material and chained archive timestamps.
- **Verify:** Test T-LTV: verification of the signature corpus with all network interfaces disabled, a fixed trust-anchor list and an expired signing credential yields the same verdict as verification at signing time for 100% of documents.

**FR-071** *(ubiquitous, must)*

> The Protodoc format SHALL carry inside the document, for every time attestation, the attesting authority's own credential chain and revocation evidence current at the attested time.

- **Why:** A timestamp token whose own chain cannot be checked offline is unverifiable under exactly the thirty-year zero-network rule imposed on the signer, so the long-term claim fails at the second hop.
- **Verify:** Test T-TSA-LTV: offline verification of every time attestation succeeds with an expired authority credential for 100% of the signature corpus.

**FR-072** *(unwanted-behavior, must)*

> IF a signature's recorded signing instant lies outside the interval attested by a time attestation whose own chain and revocation evidence verify offline, THEN a conforming reader SHALL present the document as unverified.

- **Why:** A recorded instant that no attestation supports is an assertion by the signer about themselves, and trusting it returns a positive verdict on a backdated or future-dated signature.
- **Verify:** Negative corpus N-TIME of 40 documents with unsupported, future and mismatched instants: 100% unverified, zero signer identities displayed.

**FR-073** *(unwanted-behavior, must)*

> IF revocation evidence records a credential compromise time at or before a signature's recorded signing instant, THEN a conforming reader SHALL present the document as unverified.

- **Why:** This is the backdating case: revocation after signing is survivable, revocation for a compromise that predates the signature is not, and a verifier that ignores the compromise time validates the attacker's document.
- **Verify:** Negative corpus N-COMPROMISE of 30 documents: 100% unverified with no signer identity, identical verdicts across two implementations.

**FR-074** *(event-driven, must)*

> WHEN a document is signed, the Protodoc format SHALL permit the signatory to designate content subtrees as redactable, committing each with a hiding and binding commitment whose salt is stored inside the designated subtree.

- **Why:** Without this, publishing a redacted copy of a signed document is impossible: the compliance officer must choose between a provably redacted document and a signed one, and both incumbents' worst attack class is exactly a signature that survives content removal.
- **Verify:** Test T-REDSIG: for 200 signed documents with designated subtrees, omitting any designated subtree leaves the signature verifiable; omitting an undesignated one does not.

**FR-075** *(ubiquitous, must)*

> The Protodoc format SHALL require that the content of an omitted redactable subtree is not recoverable from the retained integrity values by exhaustive search over a candidate set of 2^80 values.

- **Why:** With unsalted digests the retained hashes are a brute-force oracle for exactly the names, dates and amounts redaction was meant to remove; the commitment must hide, not merely bind.
- **Verify:** Test T-REDHIDE: a search harness over dictionaries of names, dates and amounts recovers zero omitted values from 500 published documents; static audit confirms per-subtree salting.

**FR-076** *(optional-feature, must)*

> WHERE a verified signature has designated omissions, a conforming reader SHALL report the verdict attested-with-declared-omissions and enumerate every omission.

- **Why:** A positive verdict that hides what is missing is the incumbent failure in a new costume; enumerating omissions is what lets a recipient decide whether the redactions matter.
- **Verify:** Test T-REDVERDICT: 200 published signed documents report the distinct verdict with a complete omission enumeration; two implementations produce identical enumerations.

**FR-077** *(unwanted-behavior, must)*

> IF a signed document omits content that was not designated redactable at signing time, THEN a conforming reader SHALL present the document as unverified.

- **Why:** Otherwise the omission mechanism becomes a general content-removal channel that preserves the signer's name, which is precisely the attack the coverage rule exists to close.
- **Verify:** Negative corpus N-REDBAD of 50 documents with undesignated omissions: 100% unverified with no signer identity.

**FR-078** *(ubiquitous, must)*

> The Protodoc format SHALL define a publish operation whose output contains no octet sequence belonging to any removed or superseded content unit.

- **Why:** Looks-removed and is-removed are different states and no incumbent tooling distinguishes them: a 2019 federal filing was redacted with drawn rectangles over intact text, and 2022 research recovered redacted names from residual glyph advance widths after the characters were deleted.
- **Verify:** Test T-REDACT over a 500-document redaction corpus: octet-level search of every output for every removed string returns zero hits, including layout advances, cached renderings, index entries and preview payloads.

**FR-079** *(ubiquitous, must)*

> The Protodoc format SHALL expose a single enumerable inventory of every field carrying actor identity, actor device identity or per-actor edit attribution.

- **Why:** In the incumbents the leaking surface is scattered across content attributes, property parts and container metadata with no enumeration, so no tool can prove it removed everything, which is the mechanism by which a 2003 government dossier's authors were reconstructed from its revision log.
- **Verify:** Audit A-IDENT: every schema field carrying such a value appears in the inventory; CI fails on any unlisted field.

**FR-080** *(event-driven, must)*

> WHEN the publish operation runs, the Protodoc format SHALL require the output to contain no value from the identity inventory.

- **Why:** An inventory nothing acts on is documentation; the publish obligation is what turns it into a provable property.
- **Verify:** Test T-SCRUB: after publish, an automated scan finds no inventoried identity value and the document renders identically at 300 dpi; zero residues over the corpus.

**FR-081** *(event-driven, must)*

> WHEN the publish operation runs, the Protodoc format SHALL require every custody and fixity value in the input to be present unchanged in the output.

- **Why:** Preservation reference models make provenance and fixity mandatory package content, so a blanket strip produces an object no trustworthy repository can ingest and no litigant can authenticate.
- **Verify:** Test T-CUSTODY: custody and fixity values are octet-identical between input and output for 100% of the redaction corpus.

**FR-082** *(ubiquitous, must)*

> The Protodoc format SHALL require the cells of every table to tile that table's grid exactly once, with no overlapping and no uncovered position.

- **Why:** The markup incumbent represents vertical merges as a per-cell restart flag with no integrity check, and the web table model tolerates overlapping or missing cells as recoverable errors, so two conforming consumers compute different grids from identical input.
- **Verify:** Validator rule PD-TBL-001 with negative corpus N-TBL: every non-tiling grid is rejected naming the first conflicting coordinate; zero false accepts, zero repair attempts.

**FR-083** *(ubiquitous, must)*

> The Protodoc format SHALL derive every ordered-sequence label from a specified deterministic function of document order and one directly referenced sequence definition, with no intermediate indirection and no persisted literal treated as authoritative.

- **Why:** The markup incumbent routes numbering through a three-level indirection with per-instance overrides, the leading cause of corrupt-list repair prompts and of numbering loss on merge. Numbering is legally load-bearing in contracts and standards documents.
- **Verify:** Conformance suite S-NUM including interrupted, resumed and merged sequences: two implementations compute identical visible labels for every item; zero divergences.

**FR-084** *(ubiquitous, must)*

> The Protodoc format SHALL store every cross-reference as the identity of its target together with a named presentation function, and never as literal frozen text.

- **Why:** The paged-media specification family expresses a reference as a function of its target; the incumbents freeze it and leave it stale. Because the markup incumbent has no first-class place for reference source data, citation managers smuggle it into field codes that other editors drop.
- **Verify:** Test T-XREF: two implementations agree on every resolved reference value over the reference corpus; static audit confirms no persisted authoritative literal.

**FR-085** *(ubiquitous, must)*

> The Protodoc format SHALL permit the staleness state of every cross-reference to be determined without computing layout.

- **Why:** A reference whose staleness can only be learned by re-running layout is a reference every cheap consumer will report wrongly, which is how silently wrong page numbers survive in shipped documents.
- **Verify:** Test T-XREF-STALE: after edits changing target pagination, every affected reference reports stale without any layout entry point being reached; zero false current reports.

**FR-086** *(ubiquitous, must)*

> The Protodoc format SHALL bind every derived artefact to a digest over the complete set of inputs it was computed from.

- **Why:** The markup incumbent writes a marker at every soft page break but records no engine version, font digests, geometry or validity flag, so consumers cannot tell whether it is current and trust it wrongly. Layout results without recorded dependencies are worse than none.
- **Verify:** Test T-STALE: mutating each recorded input class in turn classifies the artefact stale in 100% of cases and current in zero, across two implementations.

**FR-087** *(ubiquitous, must)*

> The Protodoc format SHALL mark every derived artefact as non-normative relative to the content it was derived from.

- **Why:** Two representations of one capability make document equality, deduplication and signature comparison undefined unless exactly one is authoritative.
- **Verify:** Audit A-DERIVED: every derived artefact kind is declared non-normative in the schema; CI fails on any artefact competing with authoritative content.

**FR-088** *(unwanted-behavior, must)*

> IF a derived artefact's recorded input digest does not match its current inputs, THEN a conforming reader SHALL refuse to serve that artefact and derive the value from the authoritative content instead, even when deriving is more expensive.

- **Why:** Serving a stale artefact is both a correctness failure and a disclosure channel: cached thumbnails have repeatedly displayed content that was subsequently redacted.
- **Verify:** Test T-CACHE: mutate content without updating each derived artefact class; every conforming consumer refuses and re-derives in 100% of cases; no redaction-corpus output yields a cached image differing from a fresh render.

**FR-089** *(ubiquitous, must)*

> The Protodoc format SHALL record for every font used exactly seven values: the font's name, its version, a digest of its content, the numeric variation-axis coordinates used, the set of code points required, the set of layout features required, and the embedding-permission values observed when the font octets were included.

- **Why:** Substituting a font changes advance widths, line breaks, pagination and every page-valued reference, so which font exactly is a correctness input. Variation named instances are unsafe to record by name because a foundry can remap them between releases, silently reflowing existing documents.
- **Verify:** Validator rule PD-FONT-001 plus test T-FONTID: every font reference carries all seven values; substituting a font whose digest differs is detected in 100% of cases.

**FR-090** *(ubiquitous, must)*

> The Protodoc format SHALL require every embedded object whose kind a reader may not implement to carry a materialised rendered representation whose extent equals the object's declared extent.

- **Why:** A legacy embedding mechanism's cached presentation is why decades-old documents still show their embedded charts, while a rich-media construct bound to a proprietary runtime lost its content outright when that runtime was withdrawn.
- **Verify:** Corpus C-EMBED with 50 unknown object kinds: a reader implementing none renders all 50 at the declared extent and rewrites their payloads octet-identically; zero failures.

**FR-091** *(ubiquitous, must)*

> The Protodoc format SHALL expose, for every opaque embedded payload, its declared kind, its octet length and a digest of its octets.

- **Why:** A clean verdict about the format's own constructs says nothing about an opaque payload's octets; exposing kind, length and digest is what lets a scanner apply its own policy or recurse instead of being told something false.
- **Verify:** Test T-OPAQUE: the inventory lists every opaque payload with all three values for 100% of corpus documents; negative case containing an executable payload is listed, never reported as containing no payloads.

**FR-092** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly one outcome for every pair of concurrent operation kinds, such that two conforming implementations given the same set of changes in any delivery order produce identical document state and identical annotation resolutions, with zero cases left implementation-defined.

- **Why:** A format whose merge outcome varies by implementation is a suggestion, not a format.
- **Verify:** Test T-MERGE: canonical state digest equality across at least 1,000 random delivery permutations of a 5-author, 20,000-operation trace for two independent implementations.

**FR-093** *(ubiquitous, must)*

> The Protodoc format SHALL require text inserted by one author at one position in one contiguous burst to appear as an unbroken substring in every merged result.

- **Why:** Convergence to identical garbage is still convergence: published work demonstrated two well-known sequence algorithms interleaving concurrently inserted passages character-by-character in real implementations, and even better families interleave on the right-to-left insertion case.
- **Verify:** Test T-NOWEAVE: two authors at one caret, both typing directions, 1,000 trials; 100% unbroken substrings across two implementations.

**FR-094** *(unwanted-behavior, must)*

> IF concurrent move operations would place a content unit within its own subtree, THEN the Protodoc format SHALL retain the move whose originating state identifier is lexicographically smaller, record the other as a conflict against the moved unit's identity, and leave that unit's subtree at its pre-move parent.

- **Why:** Concurrent tree move is the case that quietly assumes a solved research problem: two authors moving A under B and B under A produce a cycle, and every published resolution yields visibly different state, so the specification must name one.
- **Verify:** Test T-MOVECYCLE: two-node and n-node cycles under 1,000 delivery permutations; identical resulting trees and identical recorded conflicts across two implementations.

**FR-095** *(unwanted-behavior, must)*

> IF a peer transmits a change whose causal predecessor a reader does not hold, THEN the reader SHALL buffer or refuse the change without applying it and report the missing predecessor's identifier.

- **Why:** Applying changes out of causal order makes final state a function of network delivery order, converging to different documents on different machines with no error raised.
- **Verify:** Test T-CAUSAL: 1,000 randomised out-of-order deliveries; zero out-of-order applications, missing predecessor named in 100% of buffered cases.

**FR-096** *(unwanted-behavior, must)*

> IF applying a transmitted change would re-derive content that a recorded removal operation has erased, THEN a conforming reader SHALL refuse the change, report the erased unit's identifier, and persist none of the transmitted content.

- **Why:** Content the erasing party removed locally is still held by peers, so without a refusal rule the first synchronisation silently restores it and the removal was cosmetic.
- **Verify:** Test T-ERASE: 1,000 replays of erased units; zero persistences of erased content, erased identifier named in 100% of refusals.

**FR-097** *(ubiquitous, must)*

> The Protodoc format SHALL express a document's authored fixed pagination from the single document file with no conversion step.

- **Why:** A pinned signed presentation, page-valued citations and archival rendering all require the authored pagination to be first-class rather than recomputed per viewer.
- **Verify:** Test T-PAGED: every corpus document renders its authored pagination from the file alone; page geometry and page count match writer input for 100%.

**FR-098** *(ubiquitous, must)*

> The Protodoc format SHALL express a re-laid-out presentation of the same document at any viewport width down to 320 reference pixels without two-dimensional scrolling, except within regions declared to require two-dimensional presentation.

- **Why:** A format whose only normative presentation is a fixed page geometry cannot satisfy the reflow criterion that European procurement rules make a purchasing condition; the incumbent's answer is an unspecified viewer-side mode, so conformance depends on which reader the auditor opened. The exception exists because a twelve-column table cannot be linearised without loss and the source criterion carries the same exception.
- **Verify:** Test T-REFLOW at widths {320, 480, 768, 1024}: no horizontal scrolling for 100% of the corpus apart from declared two-dimensional regions, which the validator enumerates.

**FR-099** *(optional-feature, must)*

> WHERE a region is declared to require two-dimensional presentation, the Protodoc format SHALL require that region to carry a linearised reading order and a text alternative.

- **Why:** Without them the exception becomes a hole through which any content escapes both reflow and assistive technology; with them, extraction and screen-reader output are unaffected by the exception.
- **Verify:** Validator rule PD-2D-001: every declared two-dimensional region carries both; documents lacking either are rejected, zero false accepts.

**FR-100** *(ubiquitous, must)*

> The Protodoc format SHALL require the reflowed presentation at a given viewport width to be a deterministic function of document content, such that two conforming implementations produce identical line breaks and identical content-to-screen division.

- **Why:** Requiring only identical text and reading order would let two readers reflow the same document differently, reproducing the exact defect the fixed-layout incumbent is criticised for: conformance that depends on the auditor's reader rather than on the file.
- **Verify:** Test T-REFLOW-DET at widths {320, 480, 768, 1024}: line-break positions and division points identical across two implementations for 100% of the corpus.

**FR-101** *(ubiquitous, must)*

> The Protodoc format SHALL require the fixed and reflowed presentations of one document to yield identical extracted text and identical reading order.

- **Why:** If the two views can disagree, a signature over one attests nothing about what a reader of the other saw, and accessibility conformance measured on one says nothing about the other.
- **Verify:** Test T-VIEWEQ: extracted scalar sequences and reading orders compared between presentations for 100% of the corpus; zero divergences.

**FR-102** *(unwanted-behavior, must)*

> IF a document is truncated, a declared length exceeds the octets available, or any structural element is unparseable, THEN a conforming reader SHALL report the octet offset, the unit identifier and the stable rule identifier of the first failure.

- **Why:** A rejection without an offset and a rule identifier is a generic unreadable-content dialog, which is why writing a generator for the markup incumbent is largely trial and error.
- **Verify:** Negative corpus N-TRUNC of at least 100 truncated and malformed documents: 100% rejected at the correct offset with the correct rule identifier, identical verdicts across two implementations.

**FR-103** *(unwanted-behavior, must)*

> IF a document fails structural parsing at any offset, THEN a conforming reader SHALL NOT present content derived from or following that offset as complete and SHALL NOT attempt heuristic reconstruction.

- **Why:** Every real viewer of the fixed-layout incumbent implements an undocumented index-rebuilding recovery path that differs per viewer and is a direct signature-bypass vector; one vendor's repair path re-blessed a mangled package under the original signer's name.
- **Verify:** Negative corpus N-RECOVER: zero recoveries and zero partial presentations over 100 malformed documents, identical behaviour across two implementations.

**FR-104** *(unwanted-behavior, must)*

> IF a stored unit's digest does not match its octets, or two independent inventories within the file disagree, THEN a conforming reader SHALL abort before decoding that unit and report the offending identifier with expected and actual values.

- **Why:** Duplicated inventories are a parser-differential primitive: the 2013 mobile-platform signature bypass worked because the verifier read one inventory while the installer read the other, affecting every release from 2009 onward.
- **Verify:** Negative corpus N-INTEG of 100 mismatched-digest and disagreeing-inventory documents: 100% rejected before decode with the identifier named.

**FR-105** *(unwanted-behavior, must)*

> IF two inventories within one document disagree, THEN a conforming reader SHALL NOT select one over the other and SHALL NOT present any positive verification indicator.

- **Why:** Choosing a side by any precedence rule preserves the split that makes one component see different content from another; only rejection closes it.
- **Verify:** Negative corpus N-PRECEDENCE of 50 documents: zero precedence selections, zero positive indicators, identical verdicts across two implementations.

**FR-106** *(unwanted-behavior, must)*

> IF a declared decoded size, aggregate decoded size, expansion ratio, nesting depth, element count, reference count or dereferenced-octet total exceeds its stated ceiling, THEN a conforming reader SHALL abort naming the exceeded ceiling and the observed value, before allocating memory proportional to any declared value.

- **Why:** A 2019 non-recursive construction reaches roughly 281 TB from a 10 MB input in one decompression pass by overlapping entry headers, so per-entry checks are insufficient; downstream library heuristics both fail on crafted inputs and reject legitimate files.
- **Verify:** Negative corpus N-BOMB including overlapping-stream, deep-nesting and forged-length cases: 100% rejected with peak resident memory under 16777216 octets under a memory-capped harness.

**FR-107** *(unwanted-behavior, must)*

> IF a conforming reader encounters a construct it does not implement or a declared required capability generation higher than it implements, THEN it SHALL apply that construct's declared disposition and SHALL NOT silently omit, approximate or render around it.

- **Why:** Silent omission across a version boundary is indistinguishable to the user from content that was never there, and a partially rendered contract looks authoritative while missing clauses.
- **Verify:** Corpus C-FUTURE of 150 documents, 50 per disposition, opened by a previous-generation reader: observable behaviour equals the single defined behaviour in 100% of cases.

**FR-108** *(unwanted-behavior, must)*

> IF any reference, index entry, range endpoint or identity does not resolve to exactly one unit present in the document, THEN a conforming reader SHALL reject the document at validation time naming the offending value and SHALL NOT clamp, snap or otherwise repair it.

- **Why:** An out-of-range reference is the binary equivalent of a dangling pointer, and a clamped endpoint silently reattaches formatting or a comment to the wrong characters, producing a document that is valid, renderable and wrong.
- **Verify:** Fuzz property F-REF: no generated input produces a resolved reference outside the document or a repaired endpoint, over 10^9 executions; negative corpus verdicts identical across two implementations.

**FR-109** *(unwanted-behavior, must)*

> IF the document's reference graph over cross-references, embedded object containment, annotation anchoring and derived-artefact input dependencies contains a cycle, THEN a conforming reader SHALL reject the document naming every edge of the first cycle found, before evaluating any structural ceiling.

- **Why:** Every edge of such a cycle resolves to exactly one present unit, so resolution checks accept it while it generates unbounded work or unbounded expansion, the class the ceilings exist to close, arriving through a door they do not cover.
- **Verify:** Negative corpus N-CYCLE with self-reference, mutual reference and long-cycle cases: 100% rejected with all cycle edges named, identical verdicts across two implementations.

**FR-110** *(unwanted-behavior, must)*

> IF two stored units, inventory entries or content units in one document resolve to the same identifier under the format's equality rule, THEN a conforming reader SHALL reject the document reporting both locations and SHALL NOT resolve the conflict by precedence or by renaming either unit.

- **Why:** Duplicate-identifier ambiguity is a signature-bypass primitive when one component reads the first entry and another the last; renaming detaches every annotation, cross-reference and permission entry pointing at the renamed unit.
- **Verify:** Negative corpus N-DUP of 100 documents: 100% rejected with both locations reported, zero precedence selections, zero renames.

**FR-111** *(unwanted-behavior, must)*

> IF a font, image or other embedded resource required for rendering is absent, fails its digest check or fails to decode, THEN a conforming reader SHALL render in its place a mark occupying exactly the resource's recorded extent, producing at least one non-background sample within that extent in the reference raster, classified as non-decorative and carrying a text alternative naming the missing resource.

- **Why:** Silent substitution with different metrics reflows the document, changing line breaks, page breaks and every page-valued citation. Preserving the recorded extent keeps pagination and any pinned presentation stable while making the loss visible rather than invisible.
- **Verify:** Test T-MISSRES under a font-less network-less harness with each resource class removed in turn: placeholder present at the recorded extent with a non-background sample and a text alternative for 100% of documents.

**FR-112** *(unwanted-behavior, must)*

> IF a required embedded resource is absent or fails verification, THEN a conforming reader SHALL record the substitution in a machine-readable render report.

- **Why:** A visible placeholder tells a human; a report tells a pipeline. Without it, an automated publishing chain cannot detect that it shipped a degraded rendering.
- **Verify:** Test T-RENDERREPORT: every substitution appears in the report with the resource identifier and failure cause; zero unreported substitutions.

**FR-113** *(unwanted-behavior, must)*

> IF a required embedded resource is absent or fails verification, THEN a conforming reader SHALL mark the resulting pagination as non-authoritative.

- **Why:** A rendering with a substituted resource may still paginate identically, but a consumer must be able to tell that its page-valued references were resolved against a degraded rendering.
- **Verify:** Test T-PAGEFLAG: pagination is flagged non-authoritative in 100% of degraded renders and in zero clean renders.

**FR-114** *(unwanted-behavior, must)*

> IF a required embedded resource is absent or fails verification, THEN a conforming reader SHALL NOT substitute a host-system resource and SHALL NOT reflow the content.

- **Why:** Unembedded standard fonts turning a paginated contract into a different document on a different machine is the exact mechanism this rule closes.
- **Verify:** Test T-NOSUB: pagination is octet-identical to the reference for 100% of documents under the font-less harness; zero host resources loaded, verified by syscall filter.

**FR-115** *(unwanted-behavior, must)*

> IF signature verification fails, or any octet of the acted-upon set lies outside every covered range, or a signature names a cryptographic parameter outside the version allowlist, THEN a conforming reader SHALL present the document as unverified and SHALL NOT display a signer identity, a partial-validity indicator or any positive verification badge.

- **Why:** A positive indicator over partial coverage is worse than none: 2023 work found products showing a valid-signature interface for manipulated documents, and on one platform showing signature protection while performing no validation at all.
- **Verify:** Negative corpus N-SIGUI of 100 manipulated, partially covered and retired-parameter documents: zero signer identities and zero positive indicators across two implementations.

**FR-116** *(unwanted-behavior, must)*

> IF a writer produces a state that supersedes a state which is not among its recorded ancestors, THEN the Protodoc format SHALL require that state to record the superseded state's identifier together with a resolution value.

- **Why:** Both incumbents delegate concurrency to the application, so last-writer-wins at the filesystem or sync layer erases one author's work with no record, and destroys the common ancestor a three-way merge needs at the moment nobody is watching. Recording it makes silent overwrite detectable from the file alone.
- **Verify:** Test T-CONC: 1,000 concurrent-commit traces; zero states in which a superseded non-ancestor is unrecorded, verified by two implementations.

**FR-117** *(unwanted-behavior, must)*

> IF a save or commit is interrupted by process termination, power loss or storage failure, THEN the resulting file SHALL be readable at exactly one complete state, either the pre-commit or the post-commit state, with every signature on that state still verifying.

- **Why:** A package format's normal save path is a full repack, so a crash mid-save yields a truncated archive and the previous version is gone. Detectability from the file alone matters because auxiliary lock or journal files do not survive being emailed, synced or copied.
- **Verify:** Test T-CRASH kills the writer at 100 random points during commit for each of 50 documents: every resulting file opens at exactly one complete state with signatures verifying; zero blended or ambiguous outcomes.

**FR-118** *(ubiquitous, must)*

> The Protodoc format SHALL carry, for every value it mandates that was supplied by inference rather than authored, a marker identifying that value as inferred together with the basis of inference.

- **Why:** Every import from an incumbent must supply a language for every span, a reading order, an alternative for every object, a tiling table grid and durable identity, none of which real sources reliably provide. Without a marker a converter either refuses every real document or silently invents the values, which defeats the rationale of every requirement mandating them.
- **Verify:** Validator rule PD-INFER-001: inferred mandatory values are reportable at a distinct severity and enumerable by a reader; corpus C-INFER round-trips markers with zero losses.

**FR-119** *(event-driven, must)*

> WHEN a document is migrated from one format major version to the next, the Protodoc format SHALL require the migration to be a deterministic function of the source document, such that two implementations produce identical canonical octets.

- **Why:** Migration is the only lifecycle stage with no requirement today, yet a major version increment is assumed to come with a mechanical lossless converter and retired identifier ranges only a migration step can retire.
- **Verify:** Test T-MIGRATE: one corpus per version pair migrated by two implementations; canonical octets identical for 100% of documents.

**FR-120** *(event-driven, must)*

> WHEN a document is migrated between format major versions, the Protodoc format SHALL require every content unit identifier to be preserved.

- **Why:** Migration that remints identities detaches every comment, suggestion and cross-reference at the version boundary, the one moment when a whole corpus moves at once.
- **Verify:** Test T-MIGRATE-ID: identifiers compared before and after migration for 100% of corpus documents; zero changes, zero collisions.

**FR-121** *(unwanted-behavior, must)*

> IF a construct in a document being migrated has no representation in the target major version, THEN the migration SHALL refuse naming that construct and its location rather than approximating it.

- **Why:** Approximation at a version boundary is data loss applied to an entire corpus at once, discovered later in documents already signed or filed.
- **Verify:** Test T-MIGRATE-LOSS: every unrepresentable construct produces a refusal naming construct and location; zero silent approximations over the migration corpus.

**FR-122** *(event-driven, must)*

> WHEN a migrated document is verified, a conforming reader SHALL report every signature made before migration as covering the pre-migration state rather than the migrated state.

- **Why:** Migration necessarily changes canonical octets, so without this rule the first version bump either destroys every prior signature or, worse, appears to carry it forward onto content it never covered, and the thirty-year verification claim fails exactly when it matters.
- **Verify:** Test T-MIGRATE-SIG: 200 signed documents migrated; verdicts name the pre-migration state for 100%, with zero positive indicators over migrated content.

**FR-123** *(unwanted-behavior, must)*

> IF a conforming reader encounters a declared format major version it does not implement, THEN it SHALL decline to render the document naming the required major version, and SHALL NOT apply any construct disposition.

- **Why:** A major version boundary is not a capability generation: the envelope and dispositions themselves may differ, so the per-construct machinery cannot be assumed to apply.
- **Verify:** Negative corpus N-MAJOR of 20 future-major-version documents: 100% declined with the version named, zero disposition applications, identical verdicts across two implementations.

**FR-124** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly one outcome for extraction, preview, reflow, rasterisation and validation of a document containing zero content units, zero pages, a text container with no text, or a table with zero rows or columns.

- **Why:** Structural ceilings cover the upper bound of every count and never the lower bound of zero, and the empty case is where implementations diverge silently: an empty document must still report a page count, produce a raster, and receive one agreed validity verdict.
- **Verify:** Corpus C-EMPTY with one document per degenerate case: two implementations produce identical verdicts and identical outputs for 100%.

**FR-125** *(ubiquitous, must)*

> The Protodoc format SHALL have a registered media type and a registered format-identification pattern published before the first stable release.

- **Why:** Bounded-prefix identification depends on a registered pattern, and an unregistered magic number strands every operating-system type database, mail gateway and preservation pipeline the format targets.
- **Verify:** Release gate G-REGISTER: registration records for media type and format identification are attached to the v1.0 release; the pipeline refuses to publish without them.

### 6.2 Non-functional (NFR-*)

**NFR-001** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly one canonical octet sequence for any given document state, such that two implementations sharing no source code produce identical canonical octets for 100% of the conformance corpus across operating systems, architectures, locales and process orderings.

- **Why:** The binary encoding proposed in the origin sketch documents that its serialisation is not canonical and advises against hashing serialised octets; the self-describing alternative has two mutually incompatible canonical profiles. Signing, content addressing, deduplication and octet-identical round-trip all collapse without this, and it cannot be retrofitted after implementations exist.
- **Verify:** Corpus check C-CANON: at least 200 state-and-canonical-octets pairs plus the full corpus, encoded by two independent implementations on two architectures; digest equality required for 100%.

**NFR-002** *(ubiquitous, must)*

> The Protodoc format SHALL permit the canonical octet sequence of a document's current state to be computed without rewriting the file at rest.

- **Why:** This is what makes byte-stability and bounded incremental write compatible: the canonical form is the object that is hashed, signed, addressed and compared, while the file at rest is any member of a defined equivalence class over it.
- **Verify:** Test T-CANON-COMPUTE: canonical octets computed for every corpus document under a harness that faults on any write; 100% success, results equal to those from a freshly written file.

**NFR-003** *(event-driven, must)*

> WHEN a conforming writer opens a document and saves it with zero user edits, it SHALL emit a file octet-identical to the input for 100% of the conformance corpus.

- **Why:** The markup incumbent stamps randomised per-session identifiers on every save and repacks the container, so a one-word change produces a whole-file difference and third-party tools exist solely to strip those identifiers. This single property is what makes version control, construct-level difference, incremental sync, signature survival and audit trails possible.
- **Verify:** Test T-NOOP over the full corpus in CI: open-save-compare yields octet equality for 100%; any inequality fails the build.

**NFR-004** *(event-driven, must)*

> WHEN compaction of a document's stored representation runs, it SHALL be invoked explicitly and SHALL NOT be performed as part of a save, and it SHALL be refused while any signature is present.

- **Why:** Compaction is a whole-file rewrite: implicit compaction would put every save in the regime that makes a format un-syncable, and compacting a signed document rewrites the octets a signature was computed over.
- **Verify:** Test T-COMPACT: 500 saves produce zero compactions; 100 compaction attempts on signed documents are refused; explicit compaction preserves canonical octets exactly.

**NFR-005** *(ubiquitous, must)*

> The Protodoc format SHALL exclude from the octet stream every value derived from wall-clock time, machine or user identity, filesystem metadata, process randomness, uninitialised padding or editing-session history, except at exactly four named sites: identifier minting, redaction commitment salts, signature values and time attestations.

- **Why:** An entire ecosystem of post-processing tools exists purely to normalise archive timestamps, entry ordering, ownership bits and compressor differences, which proves determinism must be a normative property rather than a writer discipline. The four exceptions are enumerated because identifier minting, redaction and signing cannot function without them, and an unenumerated exception is how ambient values leak back into content.
- **Verify:** Static audit A-AMBIENT over every defined field plus test T-ENV: serialising identical content under 20 varied environments yields one distinct digest, and zero ambient-derived values appear outside the four named sites.

**NFR-006** *(ubiquitous, must)*

> The Protodoc format SHALL require signature values to be produced by a deterministic signature scheme, such that signing one document state twice with one key yields identical octets.

- **Why:** Without determinism, re-signing changes the file for no logical reason, breaking the no-op-save property and making content addressing of signed documents unstable.
- **Verify:** Test T-SIGDET: 1,000 sign-twice trials across the signature corpus; identical octets in 100% of cases.

**NFR-007** *(ubiquitous, must)*

> The Protodoc format SHALL exclude signature, time attestation and revocation-evidence components from the comparison used to test no-op-save octet identity and from the canonical state digest.

- **Why:** A signature carries a time attestation and a credential chain by construction, so without an explicit exclusion every signed document is non-conforming to the determinism rule and the static audit flags its own signature fields.
- **Verify:** Audit A-AMBIENT-SCOPE: the excluded component set is enumerated in the specification; T-NOOP and C-CANON apply it and still achieve 100% equality on the signed corpus.

**NFR-008** *(event-driven, must)*

> WHEN a conforming writer commits an edit that changes K octets of logical content, it SHALL write at most 8K plus 262144 octets to storage without rewriting stored extents containing no modified unit, for documents up to 1073741824 octets.

- **Why:** Rewrite-on-save is the defining cost failure of package formats: a one-word change to a 200 MB document rewrites 200 MB and re-transfers it to sync, backup and scanning layers. The bound is stated as a cost rather than a mechanism because the mechanism is deliberately open at this phase.
- **Verify:** Benchmark B-WRITE at the write-syscall level for the edit corpus at 1 MB, 50 MB and 500 MB; CI gate at the stated bound with the distribution reported, not the mean.

**NFR-009** *(event-driven, must)*

> WHEN a document is modified by an edit changing K octets of user content, the resulting file SHALL differ from its predecessor such that content-defined chunking produces novel chunks totalling no more than the greater of 1048576 octets and 8K, for every edit in the defined edit corpus at document sizes of 1 MB, 50 MB and 500 MB.

- **Why:** This is the adoption gate for every sync and backup vendor and it is measurable: on a 32 MiB file a 200-octet insert produced 1 novel chunk, the same insert plus a rewritten trailing index produced 407, and re-encoding everything downstream produced 1,680, which is roughly 44 KiB, 63 MB and 250 MB re-uploaded per save when scaled to 500 MB.
- **Verify:** CI job B-CHUNK chunks each edit-corpus document before and after at an 8 KiB average chunk size, failing if novel-chunk octets exceed the bound at any of the three sizes; distribution reported per edit class.

**NFR-010** *(ubiquitous, must)*

> The Protodoc format SHALL express pagination artefact entries relative to content identity rather than as absolute page ordinals stored per unit, such that repaginating after an edit rewrites artefact entries only for pages containing changed content.

- **Why:** Inserting one line on page 1 of a 10,000-page document changes the page assignment of every later page; an ordinal-keyed artefact is then rewritten end to end, putting the commonest edit class into the un-syncable regime the novel-chunk bound exists to prevent.
- **Verify:** Benchmark B-CHUNK-HEAD adds an insert-at-document-head case at 10,000 pages: novel-chunk octets remain within the NFR-009 bound; CI gate on regression.

**NFR-011** *(ubiquitous, must)*

> The Protodoc specification SHALL define a reference measurement configuration, naming the processor model, core count, memory, storage class and operating system, against which every stated time and memory bound is measured.

- **Why:** A processor described only by clock class is not a reproducible denominator, so a stated seconds budget asserted against it is untestable and every performance gate becomes a matter of whose machine ran it.
- **Verify:** Audit A-REFPLAT: every time or memory bound cites this configuration by identifier; CI benchmarks execute on it and fail if the configuration identifier does not match.

**NFR-012** *(event-driven, must)*

> WHEN a consumer extracts the full text of a 1073741824-octet, 10,000-page document, the reference library SHALL read no more than 15% of the file's octets.

- **Why:** Indexers run extraction on every file on every machine concurrently under watchdogs; the dominant extraction library added process forking and heap caps precisely because parsers had no declared cost bound.
- **Verify:** Benchmark B-EXTRACT-IO on the frozen synthetic corpus asserting octets read against file size on the reference configuration; CI regression gate.

**NFR-013** *(event-driven, must)*

> WHEN a consumer extracts the full text of a 1073741824-octet, 10,000-page document, the reference library SHALL complete within 20 seconds of processor time on the reference configuration.

- **Why:** An undeclared cost is a cost the ecosystem refuses to pay: the dominant search stack truncates extracted text defensively rather than trust an unbounded parser.
- **Verify:** Benchmark B-EXTRACT-CPU asserting processor seconds on the reference configuration; CI regression gate.

**NFR-014** *(event-driven, must)*

> WHEN a consumer extracts the full text of a document of any size, the reference library SHALL hold no more than 33554432 octets of peak resident memory.

- **Why:** Memory that grows with document size is the property that gets extraction processes killed in indexing fleets, and flatness is what makes the bound trustworthy rather than corpus-specific.
- **Verify:** Benchmark B-EXTRACT-MEM with a flatness check that peak memory at 10,000 pages is within 10% of peak memory at 100 pages; CI gate at the stated ceiling.

**NFR-015** *(event-driven, must)*

> WHEN a preview of a 10,000-page, 1073741824-octet document is requested from a cold cache, a conforming reader SHALL produce the first page image within 300 milliseconds of wall-clock time on the reference configuration reading from local storage, for every document in the corpus.

- **Why:** Preview services run in sandboxed short-lived processes with a wall-clock kill, and a killed generator yields a generic icon permanently. A percentile bound is the wrong oracle here because the tail is exactly the case that produces the permanent icon.
- **Verify:** Benchmark B-PREVIEW with cold page cache asserting the observed maximum, not a percentile, over the synthetic corpus; CI regression gate at 300 ms.

**NFR-016** *(event-driven, must)*

> WHEN a preview is requested from a cold cache, a conforming reader SHALL hold no more than 67108864 octets of peak resident memory.

- **Why:** Preview hosts cap memory as well as time, and exceeding the cap produces the same permanent generic icon as exceeding the clock.
- **Verify:** Benchmark B-PREVIEW-MEM asserting observed maximum peak resident memory on the reference configuration; CI gate at the stated ceiling.

**NFR-017** *(event-driven, must)*

> WHEN a conforming reader opens a 1073741824-octet, 10,000-page document and renders any single page, peak resident memory SHALL NOT exceed 209715200 octets.

- **Why:** Mobile readers are terminated by the operating system when they exceed a per-application budget and users blame the format rather than the device; an explicit envelope makes works-on-a-phone a gate rather than an aspiration.
- **Verify:** Benchmark B-MOBILE-MEM on a memory-capped harness rendering pages 1, 5000 and 10000; CI fails on out-of-memory or on memory growth correlated with page number.

**NFR-018** *(event-driven, must)*

> WHEN a conforming reader renders page N of a 1073741824-octet document, it SHALL read no more than 8388608 octets plus the size of the resources page N itself references.

- **Why:** Per-page cost independent of document size is the property the fixed-layout incumbent achieves only through an optional front-loaded layout that its own edit model destroys.
- **Verify:** Benchmark B-MOBILE-IO counting octets read for pages 1, 5000 and 10000 against the stated bound; CI regression gate.

**NFR-019** *(ubiquitous, must)*

> The Protodoc format SHALL determine every component value of the reference raster at 300 dots per inch by integer or exact-rational arithmetic, with the sample-grid origin, curve-flattening tolerance and subdivision order, pixel coverage computation, rounding mode, image resampling filter and compositing arithmetic stated normatively, such that two independent conforming implementations produce identical output for every document in the conformance corpus.

- **Why:** Both incumbents specify no anti-aliasing rule, no flattening tolerance and no sample position, which is why four widely used renderers disagree on the edge pixels of one path, a divergence licensed by the specifications rather than caused by bugs. The precedent for fixing it is video coding, where a tolerance-bounded transform produced decoder drift and was replaced by an exactly specified integer transform.
- **Verify:** Corpus check C-RASTER: 300 dpi rasterisation of every corpus document by two independent implementations; digest equality required for 100%, tolerance zero.

**NFR-020** *(ubiquitous, must)*

> The Protodoc format SHALL bind each format version to exactly one Unicode version, and SHALL require every document to record the Unicode version it was written against.

- **Why:** Bidirectional ordering, line breaking, normalisation and segmentation are all versioned data that has changed materially in recent releases, so two implementations built against different versions diverge on line breaks before a single pixel is computed, taking pagination and every page-valued reference with them.
- **Verify:** Audit A-UNICODE plus corpus C-UNIVER: every document records the version; two implementations built against the bound version produce identical line breaks and identical rasters for the multilingual corpus.

**NFR-021** *(ubiquitous, must)*

> The Protodoc format SHALL require glyph selection and positioning to be a deterministic function of the font digest, the variation-axis coordinates, the scalar sequence, the requested layout feature set and the language tag, with the shaping algorithm and feature application order fixed normatively per format version.

- **Why:** Shaping dominates glyph pixels and is defined today only by one implementation whose output changes between versions, so byte-identical rasterisation across independent implementations is otherwise unachievable for any text containing ligatures, contextual forms or reordered clusters.
- **Verify:** Corpus C-SHAPE of multilingual text with ligatures, contextual forms and Indic reordering: two implementations produce identical glyph identifiers and identical positions for 100% of runs.

**NFR-022** *(ubiquitous, must)*

> The Protodoc format SHALL require line-breaking and hyphenation decisions to be computed from data identified by digest within the document rather than from data supplied by the host.

- **Why:** Hyphenation and line-breaking data supplied by the host makes pagination a property of the machine, which changes what every page-valued citation refers to and defeats both the durable profile and any pinned signed presentation.
- **Verify:** Test T-LINEBREAK: rendering the corpus on two hosts with different installed text stacks produces identical line breaks and identical rasters for 100% of documents.

**NFR-023** *(ubiquitous, must)*

> The Protodoc format SHALL require reference rasterisation to evaluate glyph outlines without grid fitting and without executing any instruction stream contained in an embedded font.

- **Why:** Hinting bytecode is an interpreter embedded in the font, so executing it both contradicts the data-not-a-program rule and makes coverage per glyph per size implementation-dependent; a font-rasteriser overflow exploited in the wild and chained to a kernel bug is the same surface.
- **Verify:** Corpus C-HINT pairing identical fonts with and without instruction streams: 300 dpi rasters are identical in 100% of cases; harness records zero interpreter entries.

**NFR-024** *(optional-feature, must)*

> WHERE a document claims the durable profile, it SHALL render with identical page geometry and identical glyph shapes on a host with no fonts installed and no network access, measured by requiring 300 dpi rasterisation of every page to be identical to rasterisation performed with host fonts and network available.

- **Why:** The fixed-layout incumbent let standard fonts go unembedded, so files repaginate depending on what the host has installed, which is the reason its archival profile mandates embedding every font including subsets.
- **Verify:** Test T-DURABLE renders the durable-profile corpus twice, once under a font-less network-less harness, comparing rasters; 100% equality required.

**NFR-025** *(ubiquitous, must)*

> The Protodoc specification SHALL state, for each defined reader role, a maximum count of normative statements binding that role, and SHALL track those counts in continuous integration.

- **Why:** The markup incumbent's first edition ran to roughly 6,000 pages and exactly one complete implementation has ever existed. Specification size is a hard engineering budget determining how many correct implementations the world will ever have, and page count is not countable deterministically while a statement count is.
- **Verify:** Audit A-SIZE computes per-role normative statement counts on every commit; the release pipeline refuses to publish above any role's budget.

**NFR-026** *(ubiquitous, must)*

> The Protodoc specification SHALL enable an implementer outside the design team to produce an extracting-and-validating reader passing 100% of that role's conformance suite from the normative text alone within 5 working days and with zero clarification requests answered outside the text.

- **Why:** The five-day gate is only meaningful against a bounded role; asserting it over a renderer that must also implement shaping, layout, rasterisation and long-term signature validation guarantees the gate is waived and therefore gates nothing.
- **Verify:** Recorded external-implementer trial for the extracting-and-validating role with elapsed days and question count attached to the release; both gate the v1 freeze.

**NFR-027** *(ubiquitous, must)*

> The Protodoc specification SHALL enable an implementer outside the design team to produce a rendering reader passing 100% of that role's conformance suite from the normative text alone within 30 working days and with zero clarification requests answered outside the text.

- **Why:** Rendering carries shaping, layout and exact rasterisation, so it needs its own honest budget; one number covering all roles measures nothing.
- **Verify:** Recorded external-implementer trial for the rendering role with elapsed days and question count attached to the release; both gate the v1 freeze.

**NFR-028** *(ubiquitous, must)*

> The Protodoc format SHALL NOT be declared stable until two implementations written independently in different languages by different authors produce identical canonical octets on the conformance corpus and identical accept-or-reject verdicts on a negative corpus of at least 200 hostile cases.

- **Why:** A format with one implementation is that implementation's file format regardless of what the specification says, and matching verdicts matter as much as matching output: the 2013 signature bypass was two parsers in one system disagreeing about what a valid file was.
- **Verify:** Release gate G-INTEROP: cross-implementation comparison report attached to the v1.0 release with zero output divergences and zero verdict divergences.

**NFR-029** *(ubiquitous, must)*

> The Protodoc specification SHALL carry a stable identifier on every normative statement mapped to at least one executable case in a conformance corpus released in the same artefact as the prose, and a release SHALL be blocked while any normative statement has zero mapped cases.

- **Why:** A plain-text document specification embeds 655 executable examples in the specification file so prose and tests cannot drift, and it exists only because a decade without one fragmented that syntax into permanently incompatible dialects.
- **Verify:** CI gate G-COVER computes statement-to-case coverage on every commit; the release pipeline refuses to publish below 100%.

**NFR-030** *(ubiquitous, must)*

> The Protodoc format SHALL bound a conforming reader's peak resident memory on any input, valid or malformed, at 4 times that input's octet length.

- **Why:** In a memory-safe implementation, fuzzers find resource exhaustion and non-termination long before memory-safety bugs, and those only count as failures if the bound is a normative property of the format rather than a library default.
- **Verify:** CI gate G-FUZZ: continuous coverage-guided fuzzing seeded from the conformance corpus with crash, timeout and memory-ratio oracles; the release pipeline refuses to publish with any open reproducer.

**NFR-031** *(ubiquitous, must)*

> The Protodoc specification SHALL publish a versioned enumeration of accessibility failure conditions partitioned into structural, presence and semantic-adequacy classes, and SHALL require at least 80% of the structural and presence conditions taken together to be decidable by software without human judgement.

- **Why:** The de facto conformance model for the fixed-layout incumbent decomposes into 136 failure conditions of which only 87 are software-decidable. Authoring structure rather than inferring it removes inference failures but cannot make semantic adequacy decidable, so the target must be stated over a published denominator the project controls or it measures nothing.
- **Verify:** Audit A-A11Y: the published enumeration labels every condition with its class; CI fails if the machine-decidable share of the structural and presence classes drops below 80% or if class sizes drift without a version increment.

**NFR-032** *(optional-feature, must)*

> WHERE a document retains complete edit history, its saved size SHALL be at most 2.0 times the size of the same visible content saved with no history, measured on a text-dominated document of at least 100,000 characters produced by at least 250,000 recorded operations.

- **Why:** Demonstrated achievable rather than aspirational: on a standard editing trace of roughly 260,000 operations, plain text is 107,121 octets and one mature third-party library retaining complete history reaches 129,062 octets, about 1.21 times, while the same logical model with a naive encoding could exceed 1,300 octets per character.
- **Verify:** Benchmark B-HIST on the frozen editing trace reported as a ratio in CI, failing the build above 2.0.

**NFR-033** *(ubiquitous, must)*

> The Protodoc format SHALL bound the time and peak memory to open a document and render its current content by the size of that current content, such that two documents with identical visible content and a tenfold difference in recorded operation count open within 1.5 times each other's time and 1.5 times each other's peak resident memory.

- **Why:** At-rest size and load cost are separate failures: one mature library stored history compactly but expanded it on load, so pasting a novel cost roughly 700 MB and one user reported a document unloaded after 17 hours, while its successor loaded that document in 9 seconds. The reading path for current content must be separable from the reading path for history.
- **Verify:** Benchmark B-LOAD on paired documents at 1x and 10x operation counts; CI gate at the 1.5 ratio for both time and peak resident memory.

**NFR-034** *(ubiquitous, must)*

> The Protodoc format SHALL permit a conforming reader to complete structural validation and full integrity and signature verification without constructing any font, image, audio or video decoder, with embedded media integrity-checked as opaque octet sequences.

- **Why:** The decoder is where remote code execution actually occurs: an image-codec heap overflow in Huffman table construction was a zero-click exploited vulnerability that forced emergency patches across five major products in one week. Verification must precede decoder construction, not merely invocation, because parser state is established at construction.
- **Verify:** Test T-NODECODE runs validation and verification over the corpus under a harness that faults on any decoder entry point; zero decoder constructions permitted, and the red-team media corpus reaches zero code paths before verification.

### 6.3 Constraints (CON-*)

**CON-001** *(ubiquitous, must)*

> The Protodoc format SHALL count text positions in reported locators, diagnostics and non-persisted interfaces in Unicode scalar values within one addressable text segment, and SHALL define no persisted construct that counts text positions in any unit.

- **Why:** Scalar values are the only version-stable, encoding-independent counting unit: sixteen-bit code units admit positions inside a surrogate pair, grapheme clusters are versioned segmentation rules that have changed materially in three recent releases, and octets leak the transport encoding. Scoping the rule to non-persisted surfaces is what keeps it compatible with identity anchoring.
- **Verify:** Corpus C-UNIT with astral-plane characters, combining sequences and multi-codepoint emoji: two implementations report identical locators and identical resolved character ranges for 100% of cases; static audit confirms zero persisted counted positions.

**CON-002** *(ubiquitous, must)*

> The Protodoc format SHALL require every addressable text segment and every identifier string to be in Unicode Normalization Form C independently, with no normalisation applied across segment boundaries.

- **Why:** Normalisation changes scalar counts, so an implicit renormalisation on save relocates anchors with no exception thrown. Per-segment scoping is required because the forms are not closed under concatenation: two authors inserting individually normalised text at adjacent positions would otherwise produce a state no conforming writer may serialise.
- **Verify:** Validator rule PD-NFC-002 (renamed from an earlier draft's PD-NORM-001 at phase 5 analyze; contracts/document.abnf and data-model.md already used PD-NFC-002, per Eyvar's ruling that the contracts/data-model naming is canonical) with negative corpus N-NORM including combining marks and Hangul jamo at concurrent-insertion boundaries: every non-normalised segment is rejected at the first offending position before any digest is computed.

**CON-003** *(ubiquitous, must)*

> The Protodoc format SHALL require a conforming writer to reject text that is not already in the mandated normalization form rather than converting it.

- **Why:** Silent conversion relocates every anchor, and it lets a document be altered after signing and still verify because verifier and signer normalised different inputs.
- **Verify:** Negative corpus N-CONVERT of 50 non-normalised inputs: 100% rejected, zero conversions performed, identical verdicts across two implementations.

**CON-004** *(ubiquitous, must)*

> The Protodoc format SHALL admit in document text and identifier strings only Unicode scalar values assigned in the bound Unicode version, excluding bidirectional and deprecated formatting characters, noncharacters, surrogate scalars, private-use scalars, and control characters outside an enumerated whitelist.

- **Why:** Without a closed permitted set, in-band directional controls reappear as ordinary text and the source-hiding attack class the structural-direction rule exists to close is unenforced; unassigned scalars make language resolution and raster determinism undefined; and invisible or confusable scalars let content evade the octet-level redaction scan while looking identical.
- **Verify:** Validator rule PD-SCALAR-001 with a negative corpus alongside N-NORM: rejection occurs at the first offending scalar before any digest is computed; zero false accepts across two implementations.

**CON-005** *(ubiquitous, must)*

> The Protodoc format SHALL define at most one normative representation for any single semantic or presentational capability, and SHALL treat a document expressing one result through two different mechanisms as invalid.

- **Why:** The markup incumbent normatively contains two graphics vocabularies for one capability, so every reader must implement both permanently, and it permits direct formatting, character styles, paragraph styles and theme values to produce identical bold text, which leaves document equality, deduplication, difference and signature comparison undefined without a canonicalisation nobody agrees on.
- **Verify:** Audit A-DUP maps every core construct to exactly one capability and every capability to exactly one construct; validator rule PD-DUP-001 rejects dual-mechanism documents; CI fails on any duplicate.

**CON-006** *(ubiquitous, must)*

> The Protodoc specification SHALL contain zero normative statements whose meaning is defined by reference to the behaviour of a named application or application version, and SHALL define exactly one observable result for every construct and every combination of constructs.

- **Why:** The markup incumbent carries switches whose normative content is behave the way this closed product did in a given year. No independent implementer can be correct against those, and they are permanent because every future reader inherits every historical layout engine.
- **Verify:** Text search audit A-VENDOR over the normative specification returns zero product-behaviour references, and the conformance suite has exactly one expected output per case; both gate every release.

**CON-007** *(ubiquitous, must)*

> The Protodoc format SHALL contain no field whose value is executed, interpreted, or resolved as a network or filesystem location by any conforming reader, and SHALL define no extension point capable of later carrying one.

- **Why:** A document is data. The 2022 remote-template exploit needed no macros: a document-declared reference fetched attacker content which invoked a diagnostic handler for code execution, exploited in the wild within days. An extension point, once shipped, cannot be removed, which is why the archival profile of the fixed-layout incumbent had to prohibit the whole category while every general reader still ships the parser for it.
- **Verify:** Fuzz property F-INERT: the reference reader over an adversarial corpus under a syscall filter produces zero process launches, zero outbound connections and zero reads outside the document, over 10^9 executions; static audit confirms no location-valued or executable-valued field is definable.

**CON-008** *(ubiquitous, must)*

> The Protodoc format SHALL define every stored unit identifier as an opaque token over a fixed alphabet with a single normalized form whose equality is decided by exact comparison, with no hierarchical, path, case-folding, escaping or parent-traversal semantics.

- **Why:** The package incumbent compares part names case-insensitively while the underlying container compares entry names case-sensitively, so one package can hold two entries that are one ambiguous part. Because the names look like filesystem paths, implementations extract them to disk, which is the archive traversal class that recurred across dozens of libraries in five language ecosystems.
- **Verify:** Fuzz property F-IDENT: no generated pair of distinct identifiers is treated as equal and no identifier designates any location outside the document, over 10^9 executions plus a hostile-identifier corpus every implementation must pass.

**CON-009** *(ubiquitous, must)*

> The Protodoc specification SHALL state as exact decimal integers the maximum nesting depth, identifier length, numeric range and precision, element count, stored unit count, declared and decoded unit sizes, decompression expansion ratio, reference count, total dereferenced octets and rejection-path peak memory, in a normative table carrying its own stable identifier.

- **Why:** When limits are library defaults rather than format rules, the same file is valid in one runtime and invalid in another; the visible symptom is a file that opens in one product and not another with no way to say which side is wrong, and the invisible one is that expansion-bomb thresholds become per-implementation guesses.
- **Verify:** Audit A-LIMIT confirms every listed ceiling carries an exact integer with no unresolved marker; CI fails on any statement deferring to an unstated ceiling.

**CON-010** *(ubiquitous, must)*

> The Protodoc format SHALL ship conformance files at each stated ceiling and one unit beyond it, and SHALL require two conforming implementations to produce identical accept-or-reject verdicts for every such file.

- **Why:** A ceiling with no at-limit and over-limit case is a number nobody has tested, and boundary behaviour is exactly where two parsers disagree about what a valid file is.
- **Verify:** Corpus C-LIMIT with one at-limit and one over-limit document per ceiling: identical verdicts for 100%, with rejection occurring before allocation.

**CON-011** *(ubiquitous, must)*

> The Protodoc format SHALL permit a conforming reader to refuse a valid document exceeding the reader's own resource budget only by reporting a status that names the local budget and the observed value and is distinct from every validity verdict.

- **Why:** A document valid at the format ceilings can still exceed a specific device's memory budget, leaving the reader with no conforming behaviour: refusing would apply a stricter validity limit and accepting would be termination by the operating system. A distinct status closes it without letting local policy pollute verdict-equality testing.
- **Verify:** Test T-BUDGET: 50 valid documents exceeding a constrained harness's budget produce the distinct status in 100% of cases and are accepted by an unconstrained reader; verdict-equality testing treats the status as non-divergent.

**CON-012** *(ubiquitous, must)*

> The Protodoc format SHALL express every persisted geometric value as an integer multiple of one base unit of 1/914400 inch, which divides the inch, the point, the millimetre and the 1/96-inch reference pixel exactly.

- **Why:** Integer geometry is what makes one mature typesetting system bit-reproducible and why two major browser engines compute layout in fixed-point units chosen to avoid floating-point imprecision. Naming the value rather than only its property is what makes the constraint implementable from the text alone.
- **Verify:** Static audit A-GEOM over the schema: zero binary floating-point values in persisted geometry, every geometric field an integer multiple of the base unit; CI fails on any exception.

**CON-013** *(ubiquitous, must)*

> The Protodoc format SHALL state the rounding rule and the accumulation order used when proportional and automatic sizes are resolved.

- **Why:** A fixed-point unit still cannot represent an equal three-way split, so without a specified rule two implementations disagree by one unit per column and the error accumulates across a wide table.
- **Verify:** Corpus C-GEOM of proportional and automatic layouts including three-way and seven-way splits: two implementations produce identical resolved integer geometry for 100% of cases.

**CON-014** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly one colour representation, stating its primaries, white point, transfer function, the numeric precision of stored colour components, the space in which compositing arithmetic is performed, and the treatment of a raster resource whose embedded profile differs from it.

- **Why:** Exactly specified compositing arithmetic over undefined colour values still produces different pixels, and blending in a device space versus a linear-light space diverges visibly on any transparency. A rendering is also not reproducible in thirty years if the colour it names depends on the host's profile handling, and a second colour representation cannot be added within a major version.
- **Verify:** Corpus C-COLOR including transparency, gradients and profile-mismatched rasters: two implementations produce identical 300 dpi rasters for 100% of cases.

**CON-015** *(ubiquitous, must)*

> The Protodoc format SHALL fix a closed allowlist of cryptographic parameters per format version, and SHALL require a verifier to report a document naming any parameter outside that allowlist as invalid rather than valid-with-warning.

- **Why:** Letting the document choose its verification algorithm produced a decade of vulnerabilities in web token formats (a null-algorithm class and a key-confusion class) and required a dedicated best-practice document to correct.
- **Verify:** Negative corpus N-ALG of retired and out-of-allowlist parameters: 100% reported invalid with no signer identity, identical verdicts across two implementations.

**CON-016** *(ubiquitous, must)*

> The Protodoc format SHALL define how a document is re-protected under a stronger parameter set without re-signing its content, such that every prior signature verdict is preserved.

- **Why:** Without re-protection every document written today becomes unverifiable when its algorithm is retired, exactly as happened when one hash function was broken by a practical collision and the migration remains incomplete years later.
- **Verify:** Test T-REPROTECT: 200 signed documents re-protected under a stronger parameter set; all prior verdicts preserved, zero re-signings required, identical results across two implementations.

**CON-017** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly one conformance class for writers.

- **Why:** The markup incumbent shipped a second migration-only class partly defined by one vendor's behaviour; that class has been the dominant writer's default output since 2007, so nearly every file in existence is in the class the standard wanted retired. A class labelled legacy or transitional becomes the de facto format.
- **Verify:** Audit A-CLASS: the specification defines exactly one writer conformance class; CI fails on any second class, and the release gate checks it.

**CON-018** *(ubiquitous, must)*

> The Protodoc format SHALL define exactly three reader conformance roles (extracting, validating-and-verifying, and rendering), and SHALL assign every statement binding a conforming reader to at least one of them.

- **Why:** Thirty-odd statements bind a conforming reader while only writers had a defined class, so an implementer could not tell what they were required to build, and the size budgets, cost bounds and two-implementation gates were stated against an undefined scope. Roles are also what deliver the promise that an indexer or scanner need not implement a renderer.
- **Verify:** Audit A-ROLES: every reader-binding statement carries at least one role assignment and every role has its own conformance suite; CI fails on any unassigned statement.

**CON-019** *(ubiquitous, must)*

> The Protodoc format SHALL NOT mark any feature normative until two implementations sharing no source code pass every conformance case for that feature.

- **Why:** A feature specified but implemented once is a specification of that implementation, and the interoperability convention exists precisely because prose alone does not converge.
- **Verify:** Audit A-FEATURE: the feature register records two independent passing implementations before any feature is marked normative; the release pipeline refuses to publish otherwise.

**CON-020** *(ubiquitous, must)*

> The Protodoc format SHALL partition its extension token space into registered, owner-scoped and permanently retired sets, with a retired token never reissued and no experimental or unregistered prefix defined.

- **Why:** Reusing a retired field number in one major encoding produces silent data corruption rather than an error, and the internet standards community deprecated the experimental-prefix convention outright because unstandardised names leak permanently into the standards space and cannot be reclaimed.
- **Verify:** Audit A-REG on every schema change: rejects any schema reusing a retired token or assigning into a set it does not own; CI gate on every commit.

**CON-021** *(ubiquitous, must)*

> The Protodoc governance model SHALL publish a maximum extension registry review turnaround in business days, and SHALL track the observed turnaround against it.

- **Why:** Registries break when they lag the deployed base: one container standard had to be formally amended because it reserved an identifier that shipped implementations were already using.
- **Verify:** Governance metric G-REG: published turnaround and observed distribution reported per release; the release pipeline records both.

**CON-022** *(ubiquitous, must)*

> The Protodoc format SHALL require every document to declare, at creation and immutably, one history mode from the closed set of complete history, history retained from a declared point, and no history.

- **Why:** Retaining complete edit history, reconstructing every prior published state, and erasing specified content in place are pairwise satisfiable and not jointly satisfiable, and no cryptographic construction escapes it. A declared immutable mode lets an organisation answer the question before data is entered rather than after.
- **Verify:** Validator rule PD-MODE-001: every document declares exactly one mode; documents attempting to change a declared mode are rejected, zero false accepts.

**CON-023** *(unwanted-behavior, must)*

> IF an in-place content removal is attempted on a document declaring complete history, THEN a conforming implementation SHALL refuse the removal naming the declared history mode.

- **Why:** Complete history and in-place erasure are the two goals that cannot coexist; refusing at the point of attempt is what makes the earlier declaration meaningful rather than advisory.
- **Verify:** Test T-MODE-REMOVE: 100 attempted in-place removals on complete-history documents are refused naming the mode; zero removals performed.

**CON-024** *(unwanted-behavior, must)*

> IF a merge input holds changes predating a trimmed document's retention point, THEN a conforming implementation SHALL refuse the merge naming the retention point.

- **Why:** Merging a pre-trim branch silently restores content the trimming party removed, making the trim cosmetic and the retention declaration unenforceable.
- **Verify:** Test T-MODE-MERGE: 100 attempted merges of pre-trim branches produce a named refusal with zero merged outputs.

**CON-025** *(unwanted-behavior, must)*

> IF two documents or branches declaring different history modes are merged, THEN a conforming implementation SHALL refuse the merge naming both declared modes.

- **Why:** The mode is immutable at creation, so a complete-history branch and a no-history branch cannot both be honoured and any silent choice violates one document's declaration.
- **Verify:** Test T-MODE-MIX: 50 cross-mode merge attempts refused naming both modes; zero partial merge outputs, identical behaviour across two implementations.

**CON-026** *(ubiquitous, must)*

> The Protodoc specification SHALL be published under an irrevocable royalty-free licence covering all necessary patent claims, with a named steward, a documented succession arrangement and a deprecation window stated in years, before the first stable release.

- **Why:** An independent implementer needs the patent grant before writing code, and an archival institution needs the steward and succession on record before accessioning; neither can be retrofitted once third parties have implemented against the format.
- **Verify:** Release gate G-GOVERN: licence text, named steward, succession arrangement and deprecation window attached to the v1.0 release; the pipeline refuses to publish without all four.

### 6.4 Tooling and reference implementation (TR-*)

These requirements bind the reference library and command-line tool rather than the format itself. They are in scope for v1 because every format property above is demonstrated to a user through them.

**TR-001** *(event-driven, must)*

> WHEN a document is validated, the Protodoc command-line tool SHALL emit a machine-readable report in which every rule that fired carries a stable rule identifier, a severity from a fixed set, the offending unit identifier, an octet offset and the identifier of the normative statement the rule enforces.

- **Why:** A single missing declaration in the markup incumbent produces a generic dialog with no part name, no rule and no offset, which is why writing a generator for it is largely trial and error. Stable identifiers matter because acceptance pipelines and archival ingest contracts are written against them.
- **Verify:** Audit A-RULE: every defined validity rule has a distinct stable identifier mapped to a normative statement; CI fails on any renumbering, reuse, unmapped rule, or rejection emitted without an identifier and offset.

**TR-002** *(ubiquitous, must)*

> The Protodoc command-line tool SHALL express differences and three-way merges in terms of document constructs rather than storage units, reporting no construct on which the two inputs agree.

- **Why:** Byte-level diffing of a package format reports a one-word edit as a total rewrite, which is why documents in version control are unusable today; version-control systems already supply the hook points for an external difference and merge driver.
- **Verify:** Test T-DIFF over the edit corpus: reported construct count equals the number actually changed for 100% of edits.

**TR-003** *(unwanted-behavior, must)*

> IF a three-way merge encounters a conflict, THEN the Protodoc command-line tool SHALL exit non-zero naming both values and SHALL NOT select a side, concatenate or interleave content.

- **Why:** A merge tool that guesses on a contract clause is worse than one that stops, because the guess is invisible in the output.
- **Verify:** Test T-CONFLICT: every seeded conflicting pair yields a non-zero exit with both values named and no merged output for that construct; zero silent resolutions.

**TR-004** *(ubiquitous, must)*

> The Protodoc command-line tool SHALL generate a human-readable text projection of a document as a deterministic function of that document's canonical octets, from which parsing reproduces those canonical octets exactly for 100% of the conformance corpus.

- **Why:** One widely used binary encoding's text form deliberately injects randomised whitespace so callers cannot depend on byte equality, which makes it useless for review and version control; determinism plus a guaranteed round trip is what makes the projection usable.
- **Verify:** Round-trip test T-PROJ: project and re-parse every corpus document; canonical octets equal the original for 100%, with an empty exception list at v1.

**TR-005** *(ubiquitous, must)*

> The Protodoc command-line tool SHALL treat the text projection as outside the format's conformance surface, and a conforming reader SHALL NOT accept the projection as document input.

- **Why:** A normatively specified, losslessly round-tripping text form is a complete second representation of every capability, which the one-representation rule forbids and which becomes a fork the moment teams start hand-authoring it and demanding it be authoritative.
- **Verify:** Audit A-PROJ: the projection is excluded from the conformance surface and from A-DUP's capability mapping; negative test confirms readers reject projection input for 100% of cases.

**TR-006** *(event-driven, must)*

> WHEN a scanner has read the leading 1048576 octets of a document, the Protodoc format SHALL permit it to determine every stored unit's declared type, octet length and digest.

- **Why:** Inspection pipelines operate at line rate and cannot afford a parse; an inventory available only after decoding is an inventory nobody reads.
- **Verify:** Reference scanner limited to a 1 MiB prefix enumerates all three values for 100% of corpus documents without decompressing or decoding any content.

**TR-007** *(event-driven, must)*

> WHEN a scanner has read the leading 1048576 octets of a document, the Protodoc format SHALL permit it to determine which stored units are covered by the document's integrity protection.

- **Why:** A gateway needs to know which parts of a file are protected before deciding what its verdict is worth; unprotected regions are where an attacker works.
- **Verify:** Test T-SCAN-COVER: covered and uncovered unit sets computed from the prefix match the full-file computation for 100% of corpus documents.

**TR-008** *(event-driven, must)*

> WHEN a scanner has read the leading 1048576 octets of a document, the Protodoc format SHALL permit it to determine that no construct in the document is executed, interpreted or dereferenced by a conforming reader.

- **Why:** The whole history of document-borne code trained inspection pipelines to assume documents carry code; a machine-checkable absence of active Protodoc constructs, cheap enough to assert on every file, is the strongest adoption argument the format has with security teams, but the claim must be scoped to the format's own constructs, because an opaque payload's octets are not covered by it.
- **Verify:** Reference scanner over a red-team corpus of crafted inventories, oversized declarations, expansion bombs and forged digests, plus a benign-looking opaque payload containing an executable: zero verdicts claiming the file contains no payloads, zero false clean verdicts.

**TR-009** *(ubiquitous, must)*

> The Protodoc format SHALL bind the prefix inventory to the document's integrity protection, such that an altered inventory is detected before any listed value is relied upon.

- **Why:** An inventory a scanner trusts is an inventory an attacker will forge; unbound, it is a way to make a gateway clear a file on the strength of the attacker's own description of it.
- **Verify:** Negative corpus N-INVENTORY of 50 forged inventories: 100% detected before any listed value is used, identical verdicts across two implementations.

**TR-010** *(event-driven, must)*

> WHEN writing to storage that supports conditional replacement, the Protodoc command-line tool SHALL perform the write conditionally on the expected prior state and SHALL refuse naming the current holder when the condition fails.

- **Why:** A file format has no shared storage, so a conditional-write obligation cannot be a format requirement; but the tool that writes to shared storage can and must avoid last-writer-wins, which is how one author's work disappears with no record.
- **Verify:** Test T-CONC-CAS: 1,000 concurrent-commit trials against conditional storage; zero unconditional overwrites, current holder named in 100% of refusals.

**TR-011** *(ubiquitous, must)*

> The Protodoc reference extraction implementation SHALL require no font, shaping, layout or graphics dependency and SHALL be implementable in under 1000 source lines.

- **Why:** The extraction contract is only credible if a second party can implement it cheaply; a reference implementation that pulls in a renderer proves nothing about the cost the format imposes on integrators.
- **Verify:** Audit A-EXTRACT-SIZE counts source lines and dependencies in CI; the reference extractor reproduces writer input text exactly for 100% of the 5,000-document corpus.

**TR-012** *(ubiquitous, must)*

> The Protodoc command-line tool SHALL provide validate, inspect, extract, verify, diff, merge, project, redact, publish, sign and migrate operations at the first stable release.

- **Why:** These are the operations every adopted requirement is verified through; shipping the format without them leaves every property asserted and none demonstrable by a user.
- **Verify:** Release gate G-CLI: each named operation exercised by at least one conformance case; the pipeline refuses to publish with any operation missing or failing.

---

## 7. Open questions

None. All twelve clarifications were resolved on 2026-09-05; see `clarify.md` for the decision record, the rejected options and two accepted process deviations (AD-001, AD-002). The table below is retained for traceability.

These are unresolved and recorded in `clarify.md`. Under SDD, any unresolved item here blocks phase 3 (Plan). They are not answered in this document.

| ID | Question | Blocks |
|----|----------|--------|
| CL-001 | How are positions inside text persisted: identity anchors only, integer offsets, or offsets legal only inside an immutable published state? | The whole content model. FR-025 (identity anchoring) and CON-001 (scalar-value counting unit) sit on opposite sides of it; every annotation, comment, change record and cross-reference requirement depends on which is persisted. |
| CL-002 | What is the smallest unit carrying durable identity: character, contiguous authored run, or block? | Merge granularity, annotation granularity and the at-rest size floor. Block identity cannot express a comment on a phrase or a range crossing a block boundary; character identity costs size unless run-merging is designed in from the start. |
| CL-003 | Is rendering conformance byte-identical, or tolerance-bounded, given that no normative text-shaping specification exists anywhere to reference? | NFR-009, NFR-010, NFR-013 and FR-035 all assume byte-identical rasterisation across independent implementations. Byte-identity means writing a normative shaping specification and pricing it as a major deliverable; tolerance weakens the durable and signature-attestation claims. |
| CL-004 | Do byte-stable canonical form and bounded incremental write apply to the same artefact, or to a canonicalisation function and a file at rest respectively? | What is signed, hashed and content-addressed, and what a no-op save must preserve. Byte-identity per logical state and bounded per-edit write cost cannot both hold of the file at rest, because compaction is a whole-file rewrite. |
| CL-005 | What non-deterministic inputs are permitted, given that identifier minting, redaction commitments and signatures all need values the determinism rule bans? | NFR-005 bans wall-clock, identity and randomness from the octet stream; FR-023 needs uncoordinated unique identifiers, FR-074 needs hiding-commitment salts, FR-070 needs timestamp and signature values. Without an explicit allowlist the format cannot mint an identifier, redact safely or sign. |
| CL-006 | What happens to existing signatures when publish removes content or strips actor identity? | Whether a document can be both provably redacted and signed. Either publish invalidates prior signatures, or the format admits a signature verifying over content removed after signing, which is the incumbent attack class the format exists to close. |
| CL-007 | Which Unicode normalization form is mandated, and is it a property of each text segment or of the concatenated stream? | CON-002 and every anchoring requirement. Normalization forms are not closed under concatenation, so two authors inserting valid text at adjacent positions can produce a stream no conforming writer may serialise, with conversion banned because it would relocate anchors. |
| CL-008 | What is the default history mode for a newly created document, given that the mode is immutable at creation? | CON-022, FR-059 and FR-060, and almost every real document. Complete history forecloses in-place erasure and makes every document a standing erasure liability under European data-protection law; no history weakens the review and audit differentiator. |
| CL-009 | Are components addressed by name or by the digest of their own octets? | FR-058 storage ordering, deduplication, integrity and incremental verification. Content addressing makes identical payloads provably identical to anyone holding two documents, and can reintroduce whole-file re-layout if storage order derives from digests. |
| CL-010 | Does v1 admit a lossy raster encoding, or exactly one lossless encoding? | The raster byte-identity claim against the size and cost budgets. One lossless encoding puts a single full-page 300 dpi photograph at 10-25 MB, against an 8 MB per-page read budget and a 300 ms cold first page. |
| CL-011 | What ships in v1, given that the drafted deliverable set is a multi-year programme? | Every release gate. Converters against both incumbents, a rendering oracle at full corpus scale, twelve tool operations and a second independent implementation cannot all sit on the critical path of one release. |
| CL-012 | Is a second implementation, in a different language by a different author, committed before v1 is declared stable? | NFR-028, CON-019 and every verdict-equality gate. A format with a single implementation is that implementation's file format whatever the specification says. |

---

## 8. Traceability

Populated in `analysis.md` during phase 5 (Analyze). No requirement may reach phase 6 (Implement) without a row here naming its task, its test and the file that satisfies it.

| Requirement | Task | Test | File |
|-------------|------|------|------|
| FR-001 .. FR-125 | (phase 4) | (phase 4) | (phase 6) |
| NFR-001 .. NFR-034 | (phase 4) | (phase 4) | (phase 6) |
| CON-001 .. CON-026 | (phase 4) | (phase 4) | (phase 6) |
| TR-001 .. TR-012 | (phase 4) | (phase 4) | (phase 6) |

---

## Editorial notes

Changes made while rendering the approved requirement set into this document. Every change is either removal of implementation technology from a WHAT/WHY artefact, or a presentational normalisation. No requirement identifier, priority, EARS pattern or normative meaning was altered.

1. **Section 5.1, reference library entry.** The approved scope text read "Reference library in Go and a command-line tool". The implementation language is a mechanism decision and does not belong in a spec-phase artefact; it is already settled in the constitution (CP-010, reference implementation in Go 1.25, standard library only in core paths) and belongs in `plan.md`. Rendered here as "A reference library and a command-line tool" with no loss of scope.
2. **TR-010 rationale.** The approved rationale named a concrete storage primitive ("a compare-and-swap obligation"). Replaced with "a conditional-write obligation", matching the statement's own technology-neutral wording ("storage that supports conditional replacement"). The statement itself was already neutral and is unchanged.
3. **Section 6.4 added.** The requested structure named only 6.1 (FR), 6.2 (NFR) and 6.3 (CON), but the approved requirement set also contains twelve TR-* tooling requirements. Dropping them would lose approved content, so they are rendered verbatim in a fourth subsection with a note that they bind the reference tooling rather than the format.
4. **Em dashes removed throughout.** Statements and rationales in the approved set used em dashes as clause separators. These are rendered here as colons, commas, parentheses or restructured clauses per house style. No statement's meaning changed; in particular FR-013's three disposition values, CON-018's three reader roles and NFR-005's four permitted non-deterministic sites are all preserved exactly.
5. **`source_domains` omitted.** Present in the approved requirement set, deliberately not rendered, per the authoring instruction.
6. **Terms left as-is because they are not implementation technology.** Unicode versions and normalization forms, the 1/914400 inch base unit (CON-012), 300 dpi reference rasterisation, content-defined chunking as a measurement oracle (NFR-009), and third-party prior-art measurements cited as evidence in rationales (NFR-032, NFR-033) are all normative properties or evidence, not encoding or container choices. The origin sketch's candidate mechanisms (a specific binary encoding, a zip-of-parts container, an offset-span text model) appear nowhere as adopted choices; they survive only as rejected prior art inside FR-025 and NFR-001 rationales.

