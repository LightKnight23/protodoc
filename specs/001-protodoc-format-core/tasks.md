# Protodoc: Task Breakdown

Status: APPROVED — Eyvar, 2026-09-14 | Spec ID: 001-protodoc-format-core | Phase: 4 (tasks) | Date: 2026-09-13

---

## 1. How this file is ordered

Per AD-002 (accepted deviation, `clarify.md`), all 197 requirements carry uniform priority `must`. Priority therefore carries no sequencing information. Every ordering in this file comes from actual technical dependency: what must already be built and tested for the next thing to be buildable, organized into milestones.

| Milestone | Name | Depends on | Exit criteria |
|---|---|---|---|
| M01 | Core Encoding & Fixed Prefix | None | Header/CommitRing/Frontmatter/SegmentTable structs round-trip byte-exact with PD-RING-001 tie-break and bounds-safe arithmetic passing at-limit/over-limit fixtures. |
| M02 | Ledger & Placement | M01 | A K-octet edit commits within the NFR-008 write-cost and FR-057 index-delta budgets, and a no-op open/save round-trips octet-identical per NFR-003. |
| M03 | Extensibility Envelope | M01, M02, M07 | An envelope with each of the 3 dispositions round-trips, a missing-disposition envelope is rejected naming the token, and a fallback-reference cycle vector is rejected. |
| M04 | Identity & Anchor | M01, M02 | Run split/merge/move/reorder preserve run_id under fuzzing, no identifier is ever reissued, and orphan-carriage retains author/quoted-text/neighbours through the append-only ledger. |
| M05 | Extraction | M01, M02, M04 | A 1 GiB/10,000-page document extracts streaming, abandonable, zero-trust-material within the NFR-012/013/014 octets-read/time/memory budgets, under 1000 lines per TR-011. |
| M06 | Signature Primitive (EdDSA-Protodoc-1) | M01 | Signing one state twice with one key yields identical octets, and every allowlisted parameter set passes its vector suite while off-allowlist parameters are rejected. |
| M07 | Structural Validation Core | M01, M02, M04 | Every named ceiling aborts pre-allocation at exactly its boundary, the first reference-graph cycle names every edge before ceiling checks run, and validator peak memory holds under NFR-030's adopted reading. |
| M08 | Integrity Trees (T_S/T_C) | M01, M02, M04, M06 | T_S and T_C (including ABSENT_CHILD_DIGEST and all 3 domain tags) recompute identically across two runs and change whenever any covered value changes. |
| M09 | Signature & Coverage | M06, M08 | A signed state reports Valid only under its pinned presentation, a covered/uncovered split is enumerable via CoverageDescriptor, and verify never shows a signer identity alongside a non-valid verdict. |
| M10 | Attestation Evidence & LTV | M09 | A signing instant outside the attested interval or a revocation at/before signing both present as unverified, and a nested time-attestation credential chain verifies independently. |
| M11 | Redaction | M08, M09 | A verified signature with declared omissions enumerates them, an undesignated omission presents unverified, and publish output (including orphan-carriage quoted_text) contains zero octets of removed content. |
| M12 | History & Erasure | M04, M08 | All 3 history modes reconstruct their guaranteed state ranges, in-place removal is refused on complete-history documents, and the RLE-batched size overhead meets NFR-032 on the 250k-op benchmark. |
| M13 | Concurrent-Edit / Merge | M04, M12 | The exhaustive operation-kind-pair test shows every pair resolves to exactly one of the 4 classes, and a genuine 3-way conflict exits non-zero naming both values. |
| M14 | Rendering & Resource | M07, M08 | PLP-1/restricted-PNG decode and the exact-rational rasterizer produce tolerance-zero output within the NFR-017/018 page-cost budgets, and a missing resource renders its placeholder with pagination marked non-authoritative. |
| M15 | Accessibility & Semantic Content-Model Extensions | M04 | BLOCKED pending an Eyvar/themis spec ruling adding the missing fields (direction, reading-order, role-map, table scope, alt-text, numbering, xref presentation function, inferred-marker) to data-model.md and document.abnf, after which each gets its own conformance vector. |
| M16 | Evolution / Migration & Registry | M07, M08, M09 | Two implementations migrating the same source document produce identical canonical octets, every content-unit identifier survives unchanged, and pre-migration signatures still report covering the pre-migration state. |
| M17 | Canonicalization | M01, M02, M03, M04, M06, M07, M08, M09, M10, M11, M12, M16 | canon.Canonicalize streams C(S) without materializing it, full compaction is refused whenever a signature is present, and the project verb's text projection round-trips exactly while staying outside the conformance surface. |
| M18 | CLI Surface | M01, M02, M04, M05, M07, M08, M09, M11, M13, M14, M16, M17 | All 11 verbs are present with their documented exit-code precedence and stdout JSON payload shape, and validate's machine-readable report carries rule id, severity, unit id, octet offset, and normative-statement id per finding. |
| M19 | Conformance, Fuzzing & Governance Convergence | M01, M02, M03, M04, M05, M06, M07, M08, M09, M10, M11, M12, M13, M14, M16, M17, M18 | A second independent implementation matches canonical octets and verdicts on container/validator/canonical-serialiser per CP-003, every normative statement maps to a passing conformance case, and the licensing/stewardship/deprecation-window gate is on record before v1 is declared stable. |

---

## 2. Cross-cutting concerns

- **Wire-format conformance vectors (CP-011, CON-010)** (applies to: M01, M02, M03, M04, M07, M08, M09, M10, M11, M12, M13, M14, M16, M17): Any milestone introducing or extending a PDL-TLV/document.abnf/integrity.abnf record shape ships at-limit and one-past-limit conformance fixtures alongside the spec text before it exits, not as a follow-up.
- **argus security review** (applies to: M06, M08, M09, M10, M11, M16): Every milestone touching crypto, signing, redaction, or migration-triggered re-signing gets an argus pass before merge per CP rules on auth/PII/payment-grade changes.
- **Two-implementation gate (CP-003 / NFR-028)** (applies to: M01, M07, M17, M19): Only container, validate, and canonical-serialiser are in scope for the second independent implementation; other milestones are exempt from this specific gate but still block M19's stable-release declaration transitively.
- **Continuous fuzzing (CP-012)** (applies to: M01, M02, M03, M07, M09, M10, M14): Any milestone with a decode path over untrusted bytes (container parse, extension envelope, validator, CoverageDescriptor, LTV evidence, PLP-1/PNG codecs) wires into the fuzzing harness as part of its own exit criteria, feeding M19's harness-maturity check.
- **Governance/licensing (CP-014)** (applies to: M03, M19): Registry token-review turnaround (CON-021) attaches to M03's token governance; media-type registration (FR-125, folded into M01) and the licensing/steward/deprecation-window gate (CON-026) attach to M19 as final release gates, not coding blockers for any other milestone.
- **plan.md Section 10 first-class tasks (a)-(g)** (applies to: M09, M13, M14, M11, M01, M03, M07): CoverageDescriptor range-list fuzzer to M09, exhaustive op-kind-pair test to M13, PLP-1 decode vectors authored before its decoder to M14, orphan-carriage redaction residue vector to M11, frame_count boundary vector and PD-RING-001 tie vector both to M01, extension-envelope fallback-cycle vector to M03/M07's cycle-detection work.
- **Disclosed unresolved conflicts gating specific milestone closes** (applies to: M11, M14, M07, M19, M15): FR-061 salted-form ruling gates M11's final wire shape, EX-001 shaping-oracle exception gates only M14's shaping sub-slice, the NFR-030 memory-floor reading gates M07's close, the NFR-027 budget risk gates M14/M19 schedule sign-off, and the accessibility field gaps mean M15 cannot exit without a spec amendment — surface this to Eyvar before phase 5 analyze closes.

---

## 3. Milestones and tasks

### M01: Core Encoding & Fixed Prefix

pdlfmt+container is read by every other package; canonical-octet bugs here are unmaskable downstream, so it must be built and conformance-vectored first.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0001 | PDL-VARINT canonical encode/decode primitive | NFR-001 | None | hephaestus | `TestNFR_001_VarintMinimalEncodingRoundTrip` (unit) |
| T-0002 | PDL-TLV closed frame primitive | NFR-001 | T-0001 | hephaestus | `TestNFR_001_TLVClosedGrammarRoundTrip` (unit) |
| T-0003 | Header struct: fixed 512-octet layout, format identity, class, version | FR-006, FR-007, NFR-020 | T-0001, T-0002 | hephaestus | `TestFR_007_HeaderContentIndependent` (unit) |
| T-0004 | Capability generation fields and required>written rejection | FR-008, FR-009, FR-010 | T-0003 | hephaestus | `TestFR_010_CapabilityRequiredExceedsWritten` (unit) |
| T-0005 | Durable-profile claim field | FR-011 | T-0003 | hephaestus | `TestFR_011_DurableClaimReadFromHeader` (unit) |
| T-0006 | Reserved magic constant and format-identification pattern | FR-125 | T-0003 | hephaestus | `TestFR_125_MagicConstantIdentification` (unit) |
| T-0007 | CommitRingRecord struct: 7-slot fixed layout and parent_state_id | FR-004, FR-117 | T-0001, T-0002 | hephaestus | `TestFR_004_ParentStateIdRoundTrip` (unit) |
| T-0008 | PD-RING-001 winner selection and equal-sequence tie-break | FR-117, CON-010 | T-0007 | argus | `TestFR_117_PDRING001_EqualSequenceTie` (conformance) |
| T-0009 | Self-digesting commit-ring record validation | FR-117 | T-0007 | argus | `TestFR_117_SelfDigestDetectsTornSlot` (unit) |
| T-0010 | Crash-atomicity conformance suite: interrupted commit at each of 7 slots | FR-117, CON-010 | T-0008, T-0009 | argus | `TestFR_117_InterruptedCommitSingleReadableState` (conformance) |
| T-0011 | Frontmatter struct: fixed layout and document_metadata fields | FR-054 | T-0001, T-0002 | mnemosyne | `TestFR_054_MetadataFromBoundedPrefix` (unit) |
| T-0012 | Frontmatter preview payload fields | FR-051 | T-0011 | mnemosyne | `TestFR_051_PreviewPayloadWithinBoundedPrefix` (unit) |
| T-0013 | Preview digest binding over complete render-input set | FR-052 | T-0012 | argus | `TestFR_052_PreviewDigestCoversRenderInputs` (unit) |
| T-0014 | Preview staleness reporting on digest mismatch | FR-053 | T-0013 | hephaestus | `TestFR_053_PreviewDigestMismatchReportsStale` (unit) |
| T-0015 | Cold-cache preview time and memory budget | NFR-015, NFR-016 | T-0012, T-0013 | hephaestus | `BenchmarkNFR_015_ColdCachePreviewLatency` (benchmark) |
| T-0016 | SegmentTableSlot struct: fixed array layout, type/length/digest/flags | FR-091 | T-0001, T-0002 | mnemosyne | `TestFR_091_SegmentTableSlotFieldsExposed` (unit) |
| T-0017 | Generated structural ceiling table with CI spec-text equality check | CON-009 | None | prometheus | `TestCON_009_CeilingTableMatchesSpecText` (unit) |
| T-0018 | Bounds-safe SegmentTableSlot arithmetic check with boundary vectors | CON-010 | T-0016, T-0017 | momus | `TestCON_010_FrameCountBoundaryAndOverByOne` (conformance) |
| T-0019 | Digest mismatch abort before decode | FR-104 | T-0016 | argus | `TestFR_104_DigestMismatchAbortsBeforeDecode` (unit) |
| T-0020 | Single stored-unit inventory design closes dual-inventory class | FR-105 | T-0016 | hephaestus | `TestFR_105_ExactlyOneStoredUnitInventoryType` (unit) |
| T-0021 | TR-006 read API: enumerate every stored unit's type/length/digest from bounded prefix | TR-006 | T-0016 | hephaestus | `TestTR_006_EnumerateStoredUnitsFromBoundedPrefix` (unit) |
| T-0022 | TR-007: coverage-hint exposure from bounded prefix | TR-007 | T-0016, T-0011 | mnemosyne | `TestTR_007_CoverageHintFromBoundedPrefix` (unit) |
| T-0023 | Closed 4-value segment-type enum: zero executing kinds | TR-008 | T-0016 | argus | `TestTR_008_SegmentTypeEnumClosedNoExecutingKind` (unit) |
| T-0024 | CON-007 audit: no field executes or resolves a network/filesystem location | CON-007 | T-0002 | argus | `TestCON_007_NoFieldResolvesLocationOrExecutes` (unit) |
| T-0025 | CON-008: opaque unit-id token type | CON-008 | T-0001 | hephaestus | `TestCON_008_UnitIdOpaqueExactEqualityOnly` (unit) |
| T-0026 | CON-005 audit: at most one normative representation per capability | CON-005 | T-0001, T-0002 | hephaestus | `TestCON_005_ExactlyOneRepresentationPerCapability` (unit) |
| T-0027 | CON-012: fixed-point geometric value type (1/914400 inch units) | CON-012 | T-0011 | hephaestus | `TestCON_012_GeometricValueIsFixedPointInteger` (unit) |
| T-0028 | CON-013: proportional-size rounding rule and accumulation order | CON-013 | T-0027 | hephaestus | `TestCON_013_ProportionalSizeRoundingRemainderToFinalShare` (unit) |
| T-0029 | CON-014: single colour representation definition | CON-014 | T-0011 | hephaestus | `TestCON_014_SingleColourRepresentationDefined` (unit) |
| T-0030 | NFR-001 cross-platform canonical-octet determinism test | NFR-001 | T-0003, T-0007, T-0011, T-0016 | hephaestus | `TestNFR_001_FixedPrefixCanonicalOctetsDeterministic` (unit) |
| T-0031 | NFR-020: Unicode-version binding validation | NFR-020 | T-0003 | hephaestus | `TestNFR_020_UnicodeVersionBoundToFormatMajor` (unit) |
| T-0032 | Fixed-prefix full round-trip conformance suite (milestone exit gate) | NFR-001, CON-010 | T-0004, T-0005, T-0006, T-0008, T-0009, T-0010, T-0013, T-0014, T-0018, T-0019, T-0020, T-0021, T-0022, T-0023, T-0024, T-0025, T-0026, T-0027, T-0028, T-0029, T-0030, T-0031 | momus | `TestM01_FixedPrefixExitConformanceSuite` (conformance) |
| T-0359 | CP-012 continuous fuzz harness for untrusted-byte decode entry points |  | T-0002, T-0003, T-0007, T-0011, T-0016, T-0023 | argus | `FuzzCP_012_DecodeUntrustedBytes` (fuzz) |

**T-0001** PDL-VARINT canonical encode/decode primitive

> Implement the PDL-VARINT unsigned-integer encoding primitive in pkg pdlfmt: minimal-length encoding only, decoder rejects any non-minimal (over-long) encoding. This is the base primitive every fixed-prefix struct and every TLV length/tag field encodes through, so a bug here is a canonical-octet divergence per plan.md Section 10 item 1.

- **Implements:** NFR-001
- **Depends on:** None
- **DoD:** Encode/decode round-trips for 0, 1, max representable value, and mid-range values; decoder returns a named error for any non-minimal encoding; property test over random uint64 values confirms encode(decode(x))==x with zero free parameters.
- **Test:** `TestNFR_001_VarintMinimalEncodingRoundTrip` (unit)
- **Owner:** hephaestus

**T-0002** PDL-TLV closed frame primitive

> Implement the PDL-TLV (tag, length, value) frame primitive with a closed, enumerated field-kind vocabulary and zero free/optional parameters, per DP-002. Every record shape in container.abnf/document.abnf/integrity.abnf composes from this.

- **Implements:** NFR-001
- **Depends on:** T-0001
- **DoD:** TLV frame encodes/decodes byte-exact for every closed field-kind currently enumerated in container.abnf; decoding an unknown tag with no matching field-kind is a structural rejection, not a silent skip.
- **Test:** `TestNFR_001_TLVClosedGrammarRoundTrip` (unit)
- **Owner:** hephaestus

**T-0003** Header struct: fixed 512-octet layout, format identity, class, version

> Define the Header struct occupying container octets [0,512) with fixed-offset fields for format identity, document class, format major version, and unicode-version-id (container.abnf S2). Enforce by construction that no field derives from user/content-supplied values (FR-007) — the struct has no such field and a fuzz-seeded content difference produces byte-identical Header octets when no header field itself changes.

- **Implements:** FR-006, FR-007, NFR-020
- **Depends on:** T-0001, T-0002
- **DoD:** Header round-trips byte-exact for a table of fixture documents; a test asserts two documents differing only in content body produce identical Header octets.
- **Test:** `TestFR_007_HeaderContentIndependent` (unit)
- **Owner:** hephaestus

**T-0004** Capability generation fields and required>written rejection

> Add Header.capability-written and Header.capability-required fields and implement the PD-CAPPAIR-001 rule: reject a document whose required capability exceeds its written capability, naming both values in the error.

- **Implements:** FR-008, FR-009, FR-010
- **Depends on:** T-0003
- **DoD:** Reading the leading 512 octets yields both capability values with no further decode; a fixture with required>written is rejected and the error names both the required and written generation numbers.
- **Test:** `TestFR_010_CapabilityRequiredExceedsWritten` (unit)
- **Owner:** hephaestus

**T-0005** Durable-profile claim field

> Add Header.durable-claim boolean field; wire the invariant that a durable-claim document gates a mandatory RegistryExcerpt presence (DP-017) — this task only implements the header-level flag and its read path, not RegistryExcerpt content validation (owned by a later milestone).

- **Implements:** FR-011
- **Depends on:** T-0003
- **DoD:** Leading 512 octets alone determine durable-claim true/false with zero further decode; test confirms flag round-trips for both values.
- **Test:** `TestFR_011_DurableClaimReadFromHeader` (unit)
- **Owner:** hephaestus

**T-0006** Reserved magic constant and format-identification pattern

> Reserve the fixed magic-constant octets at the start of Header and implement the format-identification pattern match used by any consumer (OS file-type sniffers, MIME registration) ahead of the eventual CON-026/CP-014 registry submission. Registration itself is a governance action outside this task's scope.

- **Implements:** FR-125
- **Depends on:** T-0003
- **DoD:** A byte-pattern matcher correctly identifies a valid Protodoc file from its first bytes and rejects a non-matching file, with the exact reserved octet sequence checked into a shared constant used by both writer and reader.
- **Test:** `TestFR_125_MagicConstantIdentification` (unit)
- **Owner:** hephaestus

**T-0007** CommitRingRecord struct: 7-slot fixed layout and parent_state_id

> Define the CommitRingRecord struct and the 7-slot ring occupying container octets [512,4096), including the parent_state_id field (DP-001) that every state records to name the state(s) it derives from.

- **Implements:** FR-004, FR-117
- **Depends on:** T-0001, T-0002
- **DoD:** All 7 slots encode/decode byte-exact at fixed offsets; parent_state_id round-trips for a root state (empty) and a derived state (non-empty).
- **Test:** `TestFR_004_ParentStateIdRoundTrip` (unit)
- **Owner:** hephaestus

**T-0008** PD-RING-001 winner selection and equal-sequence tie-break

> Implement PD-RING-001: the highest-sequence, self-digest-valid ring slot is the winner; on an equal-sequence tie, apply the deterministic tie-break rule. This is a first-class plan.md Section 10(g)-adjacent task requiring its own equal-sequence-tie conformance vector.

- **Implements:** FR-117, CON-010
- **Depends on:** T-0007
- **DoD:** A conformance vector with two ring slots at equal sequence number resolves deterministically to the same winner across repeated runs and across two independent invocations of the selection function.
- **Test:** `TestFR_117_PDRING001_EqualSequenceTie` (conformance)
- **Owner:** argus

**T-0009** Self-digesting commit-ring record validation

> Implement the self-digest computation and validation for a CommitRingRecord: each slot's digest covers its own slot content so a torn or partial write to that slot is detectable without consulting any other slot.

- **Implements:** FR-117
- **Depends on:** T-0007
- **DoD:** A slot with a corrupted single octet fails self-digest validation; an untouched slot passes; validation requires no read outside the slot's own bytes.
- **Test:** `TestFR_117_SelfDigestDetectsTornSlot` (unit)
- **Owner:** argus

**T-0010** Crash-atomicity conformance suite: interrupted commit at each of 7 slots

> Author and run a conformance suite that simulates an interrupted write terminating mid-write at each of the 7 ring slots in turn, asserting the file remains readable at exactly one complete state and any signature over that state still verifies.

- **Implements:** FR-117, CON-010
- **Depends on:** T-0008, T-0009
- **DoD:** All 7 torn-write fixtures resolve to exactly one complete, self-digest-valid state with no ambiguous or dual-valid outcome.
- **Test:** `TestFR_117_InterruptedCommitSingleReadableState` (conformance)
- **Owner:** argus

**T-0011** Frontmatter struct: fixed layout and document_metadata fields

> Define the Frontmatter struct occupying octets [4096,262144) with fm-title, fm-page-count, fm-page-dimensions, fm-language fields, readable with zero network access.

- **Implements:** FR-054
- **Depends on:** T-0001, T-0002
- **DoD:** All 4 metadata fields decode from the leading 262,144 octets alone, with no I/O beyond the local file and no network call attempted.
- **Test:** `TestFR_054_MetadataFromBoundedPrefix` (unit)
- **Owner:** mnemosyne

**T-0012** Frontmatter preview payload fields

> Add fm-preview-kind and fm-preview-raster fields to Frontmatter, bounded entirely within the leading 262,144-octet region.

- **Implements:** FR-051
- **Depends on:** T-0011
- **DoD:** A preview payload of the maximum allowed size still fits within the 262,144-octet bound; encode/decode round-trips byte-exact.
- **Test:** `TestFR_051_PreviewPayloadWithinBoundedPrefix` (unit)
- **Owner:** mnemosyne

**T-0013** Preview digest binding over complete render-input set

> Compute fm-preview-digest over the complete set of the preview's render inputs and bind it into Frontmatter, per the threat-model's disclosed scope (only in-window inputs are authenticable).

- **Implements:** FR-052
- **Depends on:** T-0012
- **DoD:** Changing any in-window render input changes fm-preview-digest; digest computation is documented as covering only the bounded-prefix inputs per the disclosed threat-model scope.
- **Test:** `TestFR_052_PreviewDigestCoversRenderInputs` (unit)
- **Owner:** argus

**T-0014** Preview staleness reporting on digest mismatch

> A bounded-prefix consumer that recomputes fm-preview-digest and finds a mismatch reports the preview stale rather than displaying it or re-deriving it silently.

- **Implements:** FR-053
- **Depends on:** T-0013
- **DoD:** A fixture with a deliberately mismatched fm-preview-digest yields a distinct 'stale' status, never the raw preview bytes and never a silently re-rendered preview.
- **Test:** `TestFR_053_PreviewDigestMismatchReportsStale` (unit)
- **Owner:** hephaestus

**T-0015** Cold-cache preview time and memory budget

> Benchmark cold-cache preview production (one sequential 262,144-octet read plus bounded decode) against a 10,000-page/1 GiB synthetic document, verifying the 300 ms time bound and 67,108,864-octet peak resident memory bound.

- **Implements:** NFR-015, NFR-016
- **Depends on:** T-0012, T-0013
- **DoD:** Benchmark run on the reference build produces first-page image in <=300ms and peak RSS <=67,108,864 octets across 10 repeated runs.
- **Test:** `BenchmarkNFR_015_ColdCachePreviewLatency` (benchmark)
- **Owner:** hephaestus

**T-0016** SegmentTableSlot struct: fixed array layout, type/length/digest/flags

> Define the SegmentTableSlot struct and its fixed array occupying octets [262144,1048576), with segment_type, length, digest, and a flags field (used later for TR-007's coverage-hint), one slot per stored unit.

- **Implements:** FR-091
- **Depends on:** T-0001, T-0002
- **DoD:** Slot array round-trips byte-exact for 0, 1, and MAX_SEGMENTS slots; every opaque embedded payload's kind, length, and digest is readable from the slot alone.
- **Test:** `TestFR_091_SegmentTableSlotFieldsExposed` (unit)
- **Owner:** mnemosyne

**T-0017** Generated structural ceiling table with CI spec-text equality check

> Generate a checked-in Go constants file for every structural ceiling named in spec.md, with a CI step that diffs the constants against the spec text and fails the build on any drift, per CP-007.

- **Implements:** CON-009
- **Depends on:** None
- **DoD:** The ceiling table lists every spec-named ceiling as an exact decimal integer with its own stable id; the CI job fails when a constant is edited without a matching spec.md change.
- **Test:** `TestCON_009_CeilingTableMatchesSpecText` (unit)
- **Owner:** prometheus

**T-0018** Bounds-safe SegmentTableSlot arithmetic check with boundary vectors

> Implement the bounds-safe check frame_count*48+32<=segment_length-64 using the generated ceiling constants, aborting before any proportional allocation. Author conformance vectors at exactly the boundary and one octet past it, per plan.md Section 10(e) first-class task.

- **Implements:** CON-010
- **Depends on:** T-0016, T-0017
- **DoD:** The at-limit fixture passes the bounds check; the one-octet-over fixture is rejected before any allocation proportional to frame_count occurs, in both cases with zero large allocation performed.
- **Test:** `TestCON_010_FrameCountBoundaryAndOverByOne` (conformance)
- **Owner:** momus

**T-0019** Digest mismatch abort before decode

> When a SegmentTableSlot's stored digest disagrees with the segment's own self-digest, abort before decoding the segment body, reporting expected/actual digest and the segment identifier.

- **Implements:** FR-104
- **Depends on:** T-0016
- **DoD:** A fixture with a corrupted segment body (digest mismatch) aborts before any body-decode call executes, and the error carries both expected and actual digest plus the segment id.
- **Test:** `TestFR_104_DigestMismatchAbortsBeforeDecode` (unit)
- **Owner:** argus

**T-0020** Single stored-unit inventory design closes dual-inventory class

> Confirm and lock in, by construction and by a schema-level test, that SegmentTable is the sole stored-unit inventory (CQ-007) — no second inventory type exists in the container schema, so the 'two disagreeing inventories' failure class cannot arise.

- **Implements:** FR-105
- **Depends on:** T-0016
- **DoD:** A reflective/static test enumerates all container-layer types and asserts exactly one type satisfies the 'stored-unit inventory' role; adding a second such type without updating this test fails CI.
- **Test:** `TestFR_105_ExactlyOneStoredUnitInventoryType` (unit)
- **Owner:** hephaestus

**T-0021** TR-006 read API: enumerate every stored unit's type/length/digest from bounded prefix

> Expose a read-only enumeration API over the leading 1,048,576 octets that yields, for every stored unit, its declared type, octet length, and digest, with no decode of any segment body.

- **Implements:** TR-006
- **Depends on:** T-0016
- **DoD:** Enumeration over a multi-segment fixture yields all N slots' type/length/digest with the process reading only the fixed 1,048,576-octet prefix.
- **Test:** `TestTR_006_EnumerateStoredUnitsFromBoundedPrefix` (unit)
- **Owner:** hephaestus

**T-0022** TR-007: coverage-hint exposure from bounded prefix

> Wire SegmentTableSlot.flags coverage-hint bits together with Frontmatter.fm-coverage-summary so the leading 1,048,576 octets alone yield which stored units are integrity-covered, without reading any ATTEST segment body.

- **Implements:** TR-007
- **Depends on:** T-0016, T-0011
- **DoD:** A fixture with a SUBSET-mode signature yields the correct covered/uncovered slot set from the bounded prefix alone, matching the actual CoverageDescriptor once decoded (cross-checked only in the test, not required at read time).
- **Test:** `TestTR_007_CoverageHintFromBoundedPrefix` (unit)
- **Owner:** mnemosyne

**T-0023** Closed 4-value segment-type enum: zero executing kinds

> Define the closed segment_type enum {CONTENT, RESOURCE, HISTORY, ATTEST} such that the leading 1,048,576 octets prove, by the enum's own closure, that no construct is executed, interpreted, or dereferenced by a conforming reader.

- **Implements:** TR-008
- **Depends on:** T-0016
- **DoD:** Attempting to decode an out-of-enum segment_type value is a structural rejection; a static test enumerates the type and asserts no case triggers any exec/interpret/dereference code path.
- **Test:** `TestTR_008_SegmentTypeEnumClosedNoExecutingKind` (unit)
- **Owner:** argus

**T-0024** CON-007 audit: no field executes or resolves a network/filesystem location

> Audit the closed PDL-TLV field-kind vocabulary (from T-0002) and container's fixed-prefix field set, confirming none is executed, interpreted, or resolved as a network or filesystem location, and no extension point is capable of one.

- **Implements:** CON-007
- **Depends on:** T-0002
- **DoD:** A test enumerates every field-kind in the closed vocabulary and asserts none matches a location/executable-valued category; the audit is re-run automatically whenever a new field-kind is added.
- **Test:** `TestCON_007_NoFieldResolvesLocationOrExecutes` (unit)
- **Owner:** argus

**T-0025** CON-008: opaque unit-id token type

> Define the unit-id type as a 16-octet opaque token supporting only exact-comparison equality, with no hierarchy, case-folding, or traversal semantics.

- **Implements:** CON-008
- **Depends on:** T-0001
- **DoD:** unit-id exposes only an Equal method and byte accessors; no ordering, prefix, or case-insensitive comparison method exists on the type; a static test fails if one is added.
- **Test:** `TestCON_008_UnitIdOpaqueExactEqualityOnly` (unit)
- **Owner:** hephaestus

**T-0026** CON-005 audit: at most one normative representation per capability

> Audit the pdlfmt/container field-kind vocabulary for dual-mechanism duplicates (e.g. no map type alongside a repeated-key-value form, no float type alongside the fixed-point integer form), confirming exactly one representation exists per semantic capability.

- **Implements:** CON-005
- **Depends on:** T-0001, T-0002
- **DoD:** The audit enumerates every field-kind category (numeric, repeated-value, geometric) and asserts exactly one representation form exists per category.
- **Test:** `TestCON_005_ExactlyOneRepresentationPerCapability` (unit)
- **Owner:** hephaestus

**T-0027** CON-012: fixed-point geometric value type (1/914400 inch units)

> Define the geometric-value type used by Frontmatter.fm-page-dimensions (and later by PresentationArtefact) as an integer multiple of 1/914400 inch, with no floating-point representation anywhere in the type.

- **Implements:** CON-012
- **Depends on:** T-0011
- **DoD:** The geometric-value type stores only an int64 count of 1/914400-inch units; encode/decode round-trips exactly for representative page sizes; no float32/float64 field exists anywhere on the type.
- **Test:** `TestCON_012_GeometricValueIsFixedPointInteger` (unit)
- **Owner:** hephaestus

**T-0028** CON-013: proportional-size rounding rule and accumulation order

> Implement the stated rounding rule for proportional/automatic size resolution: share_i = floor(total*prefix_i/W) - floor(total*prefix_{i-1}/W), with any remainder assigned to the final share.

- **Implements:** CON-013
- **Depends on:** T-0027
- **DoD:** For a table of (total, weights) fixtures, computed shares sum exactly to total with the remainder landing on the final share, matching the documented formula bit-for-bit.
- **Test:** `TestCON_013_ProportionalSizeRoundingRemainderToFinalShare` (unit)
- **Owner:** hephaestus

**T-0029** CON-014: single colour representation definition

> Define the one normative colour representation: sRGB primaries, D65 white point, u16 per-channel components, linear-light 32-bit fixed-point compositing space, plus Frontmatter.fm-colour-profile-id identifying it.

- **Implements:** CON-014
- **Depends on:** T-0011
- **DoD:** The colour type exposes exactly the defined component width and compositing space with no alternate colour-space branch reachable in the decoder; fm-colour-profile-id round-trips.
- **Test:** `TestCON_014_SingleColourRepresentationDefined` (unit)
- **Owner:** hephaestus

**T-0030** NFR-001 cross-platform canonical-octet determinism test

> Verify that encoding one fixed input state through Header, CommitRing, Frontmatter, and SegmentTable produces byte-identical output regardless of process map/struct field iteration order, confirming the closed PDL-TLV grammar has zero free parameters end to end across the fixed prefix.

- **Implements:** NFR-001
- **Depends on:** T-0003, T-0007, T-0011, T-0016
- **DoD:** The same logical state, encoded 100 times with randomized internal map/slice iteration order (test harness deliberately randomizes it), produces byte-identical output every time.
- **Test:** `TestNFR_001_FixedPrefixCanonicalOctetsDeterministic` (unit)
- **Owner:** hephaestus

**T-0031** NFR-020: Unicode-version binding validation

> Implement the rule that each format major version binds to exactly one Unicode version: reject a document whose Header.unicode-version-id does not match the single Unicode version pinned for its declared format-major, naming both.

- **Implements:** NFR-020
- **Depends on:** T-0003
- **DoD:** A fixture with a mismatched unicode-version-id for its format-major is rejected, naming the declared and expected Unicode versions; a matching fixture passes.
- **Test:** `TestNFR_020_UnicodeVersionBoundToFormatMajor` (unit)
- **Owner:** hephaestus

**T-0032** Fixed-prefix full round-trip conformance suite (milestone exit gate)

> Assemble the milestone exit-criteria conformance suite: encode a corpus of representative documents through the full 1,048,576-octet fixed prefix (Header+CommitRing+Frontmatter+SegmentTable) and confirm byte-exact round-trip, PD-RING-001 tie-break correctness, and bounds-safe arithmetic passing at-limit/over-limit fixtures, all in one CI-run suite gating M01's close.

- **Implements:** NFR-001, CON-010
- **Depends on:** T-0004, T-0005, T-0006, T-0008, T-0009, T-0010, T-0013, T-0014, T-0018, T-0019, T-0020, T-0021, T-0022, T-0023, T-0024, T-0025, T-0026, T-0027, T-0028, T-0029, T-0030, T-0031
- **DoD:** The full corpus (including the ring-tie and frame-count-boundary fixtures from T-0008/T-0018) round-trips byte-exact and the suite is wired as a required CI check before M01 is marked closed.
- **Test:** `TestM01_FixedPrefixExitConformanceSuite` (conformance)
- **Owner:** momus

**T-0359** CP-012 continuous fuzz harness for untrusted-byte decode entry points

> M01 owns every untrusted-input decode boundary in the fixed prefix (PDL-VARINT, PDL-TLV, Header, CommitRingRecord, Frontmatter, SegmentTable) and is listed in the constitution's continuous-fuzzing cross-cutting concern (CP-012) applies_to_milestones. No task in the original M01 set registered a Fuzz-kind harness for these entry points, so this task is intentionally requirement-less: it implements no FR/NFR/CON/TR, it satisfies CP-012 directly. Seed corpora from the existing conformance fixtures (ring-tie, frame-count boundary, torn-write slots).

- **Implements:** None (see description)
- **Depends on:** T-0002, T-0003, T-0007, T-0011, T-0016, T-0023
- **DoD:** Each of the 6 decode entry points has a registered go test -fuzz target seeded from existing conformance fixtures; CI runs all 6 for a bounded continuous duration on every build per CP-012; zero crashes/panics/unrecovered errors across the seeded corpus and a fixed fuzz-time budget.
- **Test:** `FuzzCP_012_DecodeUntrustedBytes` (fuzz)
- **Owner:** argus

---

### M02: Ledger & Placement

Append-only segment placement (place()) and the fixed per-edit write/index budgets are the storage substrate every content, integrity, and history feature appends into.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0033 | Implement append-only segment placement core (place()) | FR-056 | T-0032 | hephaestus | `TestFR_056_PlaceNeverMutatesExistingOctets` (unit) |
| T-0034 | Assign monotonic storage ordinal to each placed segment | FR-058 | T-0033 | mnemosyne | `TestFR_058_StorageOrdinalMonotonicIndependentOfContent` (unit) |
| T-0035 | Guarantee no-op open/save byte-identical round trip | NFR-003 | T-0033 | hephaestus | `TestNFR_003_NoOpSaveByteIdentical` (integration) |
| T-0036 | Enforce and benchmark per-edit write-cost budget | NFR-008 | T-0033 | hephaestus | `BenchmarkNFR_008_EditWriteCostWithinBudget` (benchmark) |
| T-0037 | Implement content-defined chunking (CDC) boundary algorithm for edit deltas | NFR-009 | T-0033 | hephaestus | `TestNFR_009_ChunkBoundaryDeterministicAcrossEditHistories` (unit) |
| T-0038 | Benchmark novel-chunk-total budget across 1/50/500 MB documents | NFR-009 | T-0037 | hephaestus | `BenchmarkNFR_009_NovelChunkTotalAcrossDocSizes` (benchmark) |
| T-0039 | Preserve unimplemented constructs octet-for-octet on write, or refuse naming them | FR-017 | T-0033 | hephaestus | `TestFR_017_OpaqueConstructPreservedOrRefused` (unit) |
| T-0040 | Verify opaque-segment retention invariant holds across two divergent copies | FR-018 | T-0039 | hephaestus | `TestFR_018_OpaqueSegmentSurvivesTwoDivergentCopies` (integration) |
| T-0041 | Construct UnitIndexLeaf derived index structure | FR-055 | T-0033, T-0034 | mnemosyne | `TestFR_055_UnitIndexLeafConstructionRebuildsFromContent` (unit) |
| T-0042 | Wire CommitRingRecord.index-route to a <=3-dependent-read bounded lookup path | FR-055 | T-0041 | hephaestus | `TestFR_055_BoundedLookupWithinThreeReads` (integration) |
| T-0043 | Enforce combined index+integrity per-edit delta budget (<=262,144 octets) | FR-057 | T-0042, T-0036 | hephaestus | `TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced` (unit) |
| T-0044 | Define ConditionalWriter interface for external shared-storage writes | TR-010 | T-0033 | hephaestus | `TestTR_010_ConditionalWriterInterfaceContract` (unit) |
| T-0045 | Refuse conditional write on token mismatch, naming current holder | TR-010 | T-0044 | hephaestus | `TestTR_010_ConditionalWriteRefusalNamesCurrentHolder` (unit) |
| T-0046 | Enforce sealed-segment write-once immutability | FR-056 | T-0033 | hephaestus | `TestFR_056_SealedSegmentRejectsMutation` (unit) |
| T-0047 | Wire ledger segment/frame decode path into continuous fuzzing harness | FR-056 | T-0033, T-0046 | prometheus | `FuzzFR_056_SegmentFrameDecode` (fuzz) |
| T-0048 | Author conformance vectors at the FR-057 262,144-octet delta boundary | FR-057 | T-0043 | momus | `CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY` (conformance) |
| T-0049 | Document TR-010 vs FR-117 mechanism distinction and interface contract | TR-010 | T-0044, T-0045 | clio | `TestTR_010_DesignNoteContainsRequiredSections` (unit) |

**T-0033** Implement append-only segment placement core (place())

> Implement the `place()` placement function in the `ledger` package: given a prior file image and an edit delta, appends new sealed segments to the ledger region and never rewrites any octet belonging to a segment already present in the prior image. Requires M01's Header/CommitRing/SegmentTable structs (container.abnf) to already round-trip via T-0032's fixed-prefix conformance suite. This is the foundational primitive every other M02 task builds on.

- **Implements:** FR-056
- **Depends on:** T-0032
- **DoD:** place(prior, delta) returns a new image whose byte range [0, len(prior)) is byte-identical to prior at every offset not newly allocated; a differential test writing N edits and diffing intermediate images confirms zero mutation of previously committed octets.
- **Test:** `TestFR_056_PlaceNeverMutatesExistingOctets` (unit)
- **Owner:** hephaestus

**T-0034** Assign monotonic storage ordinal to each placed segment

> Add a strictly-increasing storage-ordinal counter to place()'s allocation path (CQ-007 option B) so every SegmentTableSlot receives an ordinal derived purely from allocation sequence, never from segment name, digest, or content bytes.

- **Implements:** FR-058
- **Depends on:** T-0033
- **DoD:** Two segments with byte-identical content placed at different times receive different, strictly increasing ordinals; a segment's ordinal is provably a pure function of allocation order (property test permutes content while holding order fixed and asserts ordinal sequence is unchanged).
- **Test:** `TestFR_058_StorageOrdinalMonotonicIndependentOfContent` (unit)
- **Owner:** mnemosyne

**T-0035** Guarantee no-op open/save byte-identical round trip

> Verify and, where needed, adjust place() so that place(prior, no-op-edit) == prior exactly (no re-serialization, no re-ordinal-assignment, no incidental digest recomputation touching stored octets).

- **Implements:** NFR-003
- **Depends on:** T-0033
- **DoD:** Opening a corpus document and immediately saving with zero edits produces a file whose SHA-256 matches the input file's SHA-256, across all M01 conformance fixtures.
- **Test:** `TestNFR_003_NoOpSaveByteIdentical` (integration)
- **Owner:** hephaestus

**T-0036** Enforce and benchmark per-edit write-cost budget

> Instrument place() to track total octets written per commit (appended segment + patched fixed-prefix slots) and assert it never exceeds 8K+262,144 for a K-octet edit, at document sizes up to 1 GiB, per plan.md Section 5 row 4 arithmetic.

- **Implements:** NFR-008
- **Depends on:** T-0033
- **DoD:** Benchmark suite committing K-octet edits (K = 1B, 1KB, 1MB) against synthetic 1KB/1MB/1GiB documents reports write-octet totals all <= 8K+262144, with the bound enforced as a hard assertion (commit refuses/panics-in-debug-build on violation) rather than only measured.
- **Test:** `BenchmarkNFR_008_EditWriteCostWithinBudget` (benchmark)
- **Owner:** hephaestus

**T-0037** Implement content-defined chunking (CDC) boundary algorithm for edit deltas

> Implement the rolling-hash content-defined-chunking algorithm used to decide novel-chunk boundaries when an edit is placed, so that chunk boundaries are a deterministic function of content rather than of edit-offset alignment.

- **Implements:** NFR-009
- **Depends on:** T-0033
- **DoD:** Given the same document content reached via two different edit histories, the resulting chunk boundary set is identical (property test); chunk boundary determination is a pure function with no dependency on prior edit count or offset.
- **Test:** `TestNFR_009_ChunkBoundaryDeterministicAcrossEditHistories` (unit)
- **Owner:** hephaestus

**T-0038** Benchmark novel-chunk-total budget across 1/50/500 MB documents

> Run the CDC algorithm from T-0037 against the plan.md Section 5 row 6 benchmark scenario (single-character-style edits against 1MB/50MB/500MB baseline documents) and assert novel-chunk total per edit stays within max(1 MiB, 8K).

- **Implements:** NFR-009
- **Depends on:** T-0037
- **DoD:** Benchmark reports novel bytes written per edit at each of the 3 document sizes, all <= max(1 MiB, 8K); results recorded and compared against the ~152KB/19-chunk figures cited in plan.md Section 5 row 6 as a regression baseline.
- **Test:** `BenchmarkNFR_009_NovelChunkTotalAcrossDocSizes` (benchmark)
- **Owner:** hephaestus

**T-0039** Preserve unimplemented constructs octet-for-octet on write, or refuse naming them

> In place()'s write path, ensure any segment/construct the current writer does not recognize is carried through unmodified as an opaque byte range; if for some reason a construct cannot be carried through unmodified, the writer must refuse the commit and name the specific construct that could not be preserved rather than silently dropping or approximating it.

- **Implements:** FR-017
- **Depends on:** T-0033
- **DoD:** A corpus document containing a synthetic unrecognized segment type round-trips through open+edit+save with that segment's octets byte-identical; a fault-injected unpreservable-construct case produces a refusal error naming the construct's token/id, not a silent write.
- **Test:** `TestFR_017_OpaqueConstructPreservedOrRefused` (unit)
- **Owner:** hephaestus

**T-0040** Verify opaque-segment retention invariant holds across two divergent copies

> Establish (at the ledger primitive level, ahead of M13's merge classifier) the invariant merge will rely on: given two independently-edited copies of a document each containing the same unrecognized construct, both copies retain that construct's octets unchanged after their respective independent edits. This is the storage-layer guarantee FR-018 needs; the merge algorithm itself is out of scope here (M13).

- **Implements:** FR-018
- **Depends on:** T-0039
- **DoD:** Test forks a document with an opaque segment into two copies, applies independent unrelated edits to each via place(), and asserts the opaque segment's octets are byte-identical in both resulting images.
- **Test:** `TestFR_018_OpaqueSegmentSurvivesTwoDivergentCopies` (integration)
- **Owner:** hephaestus

**T-0041** Construct UnitIndexLeaf derived index structure

> Build the UnitIndexLeaf derived (non-normative, rebuilt-on-open) index entity that maps content-unit identity to storage-ordinal/offset, as the lookup structure FR-055's bounded-read path depends on.

- **Implements:** FR-055
- **Depends on:** T-0033, T-0034
- **DoD:** UnitIndexLeaf construction from a SegmentTable produces one entry per addressable unit, rebuildable from authoritative content alone (no persisted-only fields), and round-trips through a rebuild-and-compare test.
- **Test:** `TestFR_055_UnitIndexLeafConstructionRebuildsFromContent` (unit)
- **Owner:** mnemosyne

**T-0042** Wire CommitRingRecord.index-route to a <=3-dependent-read bounded lookup path

> Implement the read path: CommitRingRecord.index-route -> SegmentTableSlot -> UnitIndexLeaf -> content segment, and confirm both FR-055 clauses: the read chain never exceeds 3 dependent reads, and total octets read beyond those 3 reads stay within a stated, documented bound.

- **Implements:** FR-055
- **Depends on:** T-0041
- **DoD:** Locating and reading an arbitrary addressable unit in a synthetic 1GiB/10,000-unit document completes in exactly <=3 sequential dependent reads (instrumented read-counter assertion) and the total octet volume read beyond those 3 reads is measured and documented as the FR-055 budget figure.
- **Test:** `TestFR_055_BoundedLookupWithinThreeReads` (integration)
- **Owner:** hephaestus

**T-0043** Enforce combined index+integrity per-edit delta budget (<=262,144 octets)

> Extend the write-cost accounting from T-0036 to separately track the index-update delta (UnitIndexLeaf/SegmentTable patch size) and reserve headroom for the integrity-tree delta using the worst-case T_S/T_C update-size formula from plan.md Section 5 row 5 (integrity trees themselves land in M08; this task uses the documented worst-case estimate as a conservative stand-in), asserting the combined total never exceeds 262,144 octets per edit. Flag for follow-up: a full end-to-end re-verification against the real integrity-tree implementation should be scheduled once M08 lands.

- **Implements:** FR-057
- **Depends on:** T-0042, T-0036
- **DoD:** Per-edit accounting reports index-delta-octets + estimated-integrity-delta-octets <= 262,144 for edits at 1B/1KB/1MB granularity against 1KB/1MB/1GiB documents, matching or improving on the ~26,161-octet worst-case / ~10x-headroom figure in plan.md Section 5 row 5; a tracking note is left for the M08 follow-up integration test.
- **Test:** `TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced` (unit)
- **Owner:** hephaestus

**T-0044** Define ConditionalWriter interface for external shared-storage writes

> Define a storage-backend abstraction (e.g. ledger.ConditionalWriter) distinct from FR-117's file-internal commit ring: an interface for writing a whole file to a possibly concurrently-written external store (bucket/filesystem) conditionally on an expected prior state (ETag/generation-number-style token), plus a local-filesystem reference adapter. NOTE: plan.md currently describes no such abstraction (TR-010 is repeatedly conflated with FR-117 in the frozen plan); this task implements the minimal missing piece and the gap should be raised for a plan.md amendment before M18's CLI surface is built on top of it.

- **Implements:** TR-010
- **Depends on:** T-0033
- **DoD:** ConditionalWriter interface compiles with a Write(expectedToken, newContent) (actualToken, error) signature; a local-filesystem adapter implementing it passes a test performing a successful conditional write followed by a conflicting concurrent write attempt.
- **Test:** `TestTR_010_ConditionalWriterInterfaceContract` (unit)
- **Owner:** hephaestus

**T-0045** Refuse conditional write on token mismatch, naming current holder

> Implement the refusal path of ConditionalWriter: on an expected-token mismatch, the write is refused and the error names the current holder's token/identity rather than silently overwriting (last-writer-wins).

- **Implements:** TR-010
- **Depends on:** T-0044
- **DoD:** A conditional write with a stale expected-token against the local-filesystem adapter returns a refusal error whose message/fields include the actual current token, and no bytes of the target file are modified.
- **Test:** `TestTR_010_ConditionalWriteRefusalNamesCurrentHolder` (unit)
- **Owner:** hephaestus

**T-0046** Enforce sealed-segment write-once immutability

> Add an explicit guard (distinct from T-0033's happy-path append logic) that rejects any code path attempting to reopen or mutate a segment already marked sealed in the SegmentTable, surfacing a defensive error rather than allowing accidental in-place writes.

- **Implements:** FR-056
- **Depends on:** T-0033
- **DoD:** A fault-injection test that attempts to write through a handle to an already-sealed segment's byte range receives an explicit immutability-violation error, verified via both a direct API-misuse test and a randomized fuzz-style sequence of interleaved seal/write attempts.
- **Test:** `TestFR_056_SealedSegmentRejectsMutation` (unit)
- **Owner:** hephaestus

**T-0047** Wire ledger segment/frame decode path into continuous fuzzing harness

> Per CP-012 (continuous fuzzing applies to M02's ledger decode path over untrusted bytes), add a fuzz target that feeds arbitrary byte sequences into the segment-table/frame decoder and asserts it never panics, never mutates octets outside the decode buffer, and either parses successfully or returns a structured error.

- **Implements:** FR-056
- **Depends on:** T-0033, T-0046
- **DoD:** Fuzz target registered with the project's continuous fuzzing harness (go test -fuzz), seeded with the M01 at-limit/over-limit conformance fixtures, runs crash-free for the CI-mandated fuzz duration, feeding results into M19's harness-maturity check.
- **Test:** `FuzzFR_056_SegmentFrameDecode` (fuzz)
- **Owner:** prometheus

**T-0048** Author conformance vectors at the FR-057 262,144-octet delta boundary

> Per CON-010/CP-011 (at-limit and one-past-limit fixtures ship alongside the mechanism, not as a follow-up), author a conformance corpus case whose edit produces a combined index+integrity delta of exactly 262,144 octets, and a sibling case one octet past it, asserting the first commits and the second is refused.

- **Implements:** FR-057
- **Depends on:** T-0043
- **DoD:** Two corpus fixtures exist (at exactly 262,144 octets and at 262,145 octets combined delta) with recorded expected verdicts (commit / refused-naming-ceiling); both fixtures are wired into the conformance test runner and pass.
- **Test:** `CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY` (conformance)
- **Owner:** momus

**T-0049** Document TR-010 vs FR-117 mechanism distinction and interface contract

> Write a short design note clarifying that FR-117's 7-slot commit ring solves file-internal torn-write detection while TR-010's ConditionalWriter solves external shared-storage last-writer-wins avoidance, documenting the ConditionalWriter contract for CLI implementers (M18) and flagging the plan.md gap (no storage-backend abstraction was described in the frozen plan) for an Eyvar-approved plan.md amendment.

- **Implements:** TR-010
- **Depends on:** T-0044, T-0045
- **DoD:** A recorded note (plan.md addendum pending Eyvar approval, or a linked design doc referenced from plan.md) exists distinguishing FR-117 from TR-010 and describing ConditionalWriter's contract; a doc-lint unit test (ledger/doc_test.go) asserts the note file exists at its documented path and contains both required section headers ('FR-117 vs TR-010 distinction' and 'ConditionalWriter contract'), failing if either section or the file itself is missing. M18's CLI task list can cite this doc rather than re-deriving the distinction.
- **Test:** `TestTR_010_DesignNoteContainsRequiredSections` (unit)
- **Owner:** clio

---

### M03: Extensibility Envelope

Every non-core construct funnels through EXT_ENVELOPE, so forward-compat and unimplemented-construct preservation must exist before any content-model feature that could be non-core.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0050 | ext-token type: encode/decode + owner-id partition | CON-020 | None | hephaestus | `TestCON_020_ExtTokenOwnerIdPartition` (unit) |
| T-0051 | Frontmatter.fm-retired-tokens field (tag=10) | CON-020 | T-0050 | hephaestus | `TestCON_020_FmRetiredTokensSizeCap` (conformance) |
| T-0052 | ExtensionEnvelope record struct and PDL-TLV codec | FR-012 | T-0050 | hephaestus | `TestFR_012_ExtEnvelopeRoundTrip` (unit) |
| T-0053 | ext-payload-digest verification before disposition applies | FR-012, FR-107 | T-0052 | argus | `TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition` (unit) |
| T-0054 | ext-disposition closed 3-value enum with structural-reject on missing/out-of-range | FR-013, FR-014 | T-0052 | hephaestus | `TestFR_013_014_ExtDispositionStructuralReject` (unit) |
| T-0055 | PD-EXT-002: ext-fallback-ref mandatory-when-{ignore,degrade} | FR-015 | T-0054 | hephaestus | `TestFR_015_PD_EXT_002_FallbackMandatoryWhenIgnoreOrDegrade` (unit) |
| T-0056 | Fallback reachability audit (fallback/primary exclusion) | FR-015 | T-0055, T-0118 | hephaestus | `TestFR_015_FallbackReachabilityExclusion` (integration) |
| T-0057 | PD-EXT-003: fallback minimum-content check, and close the contracts/README rule-registry gap | FR-016 | T-0056 | hephaestus | `TestFR_016_PD_EXT_003_FallbackMinimumContent` (unit) |
| T-0058 | ext-position-key canonical ordering among unrecognised constructs | FR-012 | T-0052 | hephaestus | `TestFR_012_ExtPositionKeyCanonicalOrder` (unit) |
| T-0059 | Reader disposition-enforcement policy (ignore/degrade/refuse runtime behaviour) | FR-107 | T-0054, T-0055 | hephaestus | `TestFR_107_DispositionRuntimeEnforcement` (integration) |
| T-0060 | Writer-side retirement enforcement: a retired token is never reissued | CON-020 | T-0050, T-0051 | hephaestus | `TestCON_020_RetiredTokenNeverReissued` (unit) |
| T-0061 | Registry review turnaround SLA published and tracked (CON-021) | CON-021 | T-0050 | clio | `TestCON_021_RegistryTurnaroundSLAPublished` (unit) |
| T-0062 | Conformance suite: one round-trip fixture per disposition value | FR-012, FR-013 | T-0052, T-0054 | momus | `TestFR_012_013_DispositionRoundTripCorpus` (conformance) |
| T-0063 | Conformance vector: missing-disposition envelope rejected naming token | FR-014 | T-0054 | momus | `TestFR_014_MissingDispositionRejectedNamingToken` (conformance) |
| T-0064 | Conformance vector: extension-envelope fallback-reference cycle rejected | FR-015 | T-0056, T-0118 | momus | `TestFR_015_FR_109_ExtFallbackCycleRejected` (conformance) |
| T-0065 | ext-token owner-id partition boundary conformance vectors | CON-020 | T-0050 | momus | `TestCON_020_OwnerIdPartitionBoundaries` (conformance) |
| T-0066 | Continuous fuzzing harness for ExtensionEnvelope decode (CP-012) | FR-107, FR-014 | T-0052, T-0053, T-0054 | prometheus | `FuzzExtEnvelopeDecode` (fuzz) |

**T-0050** ext-token type: encode/decode + owner-id partition

> Implement the 8-octet ext-token primitive (4-octet owner-id + 4-octet owner-local-seq) per container.abnf S5.2: owner-id 0x00000000 reserved/invalid, 0x00000001-0x7FFFFFFF registered (registry-tracked), 0x80000000-0xFFFFFFFE owner-scoped (self-issued, no registry round trip), 0xFFFFFFFF permanently-retired tombstone namespace. No experimental/unregistered prefix exists anywhere in the space. Requires M01's PDL-TLV field/varint primitives to already exist (pdlfmt package). Pure value type, no I/O.

- **Implements:** CON-020
- **Depends on:** None
- **DoD:** ExtToken type classifies any 8-octet value into exactly one of {reserved-invalid, registered, owner-scoped, retired-tombstone}; an owner-id a reader does not recognise as registered still decodes as well-formed owner-scoped per the NORMATIVE note; round-trip encode/decode is lossless for all 4 partition classes.
- **Test:** `TestCON_020_ExtTokenOwnerIdPartition` (unit)
- **Owner:** hephaestus

**T-0051** Frontmatter.fm-retired-tokens field (tag=10)

> Implement Frontmatter field tag=10 per container.abnf S4: plain-seq-of(ext-token), <=16384 octets total, the set of tokens this document has permanently retired. Enforce the size cap at decode time (structural reject, not truncation) and reject any entry whose owner-id partition (T-0050) is not the 0xFFFFFFFF tombstone namespace.

- **Implements:** CON-020
- **Depends on:** T-0050
- **DoD:** A Frontmatter with exactly 16384 octets of retired-token entries decodes; one octet past that cap is rejected naming the field; an entry with a non-tombstone owner-id is rejected.
- **Test:** `TestCON_020_FmRetiredTokensSizeCap` (conformance)
- **Owner:** hephaestus

**T-0052** ExtensionEnvelope record struct and PDL-TLV codec

> Implement the ext-envelope record shape from document.abnf S6: ext-discriminant (0x06), ext-tok, ext-payload-length, ext-payload-ref, ext-payload-digest, ext-disposition, ext-fallback-ref, ext-position-key. Legal in CONTENT and RESOURCE segments. Requires M02's unit-id/segment addressing (place()) to already exist for ext-payload-ref to resolve. This task covers only the 7-field record's own encode/decode round-trip, not the semantic checks on disposition/fallback (T-0054..T-0057) or the digest verification (T-0053).

- **Implements:** FR-012
- **Depends on:** T-0050
- **DoD:** An ext-envelope record with arbitrary valid field values round-trips to identical canonical octets; a record with an unrecognised tag in 8-255 is skipped via field-len per the per-record-type reserved-tail rule, never rejected solely for its presence.
- **Test:** `TestFR_012_ExtEnvelopeRoundTrip` (unit)
- **Owner:** hephaestus

**T-0053** ext-payload-digest verification before disposition applies

> A reader MUST verify ext-payload-digest (SHA-256 over the payload octets) against the referenced RESOURCE segment's own SegmentTableSlot digest, and reject a mismatch, BEFORE applying any disposition -- this is the one check every reader performs even when ext-tok is unrecognised, since the payload octets stay opaque. This is the mechanism that lets an envelope with an unimplemented token still apply its declared disposition (never silent omission) per FR-107.

- **Implements:** FR-012, FR-107
- **Depends on:** T-0052
- **DoD:** A payload whose octets have been altered so ext-payload-digest no longer matches the SegmentTableSlot digest is rejected before disposition dispatch, for both a recognised and an unrecognised ext-tok.
- **Test:** `TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition` (unit)
- **Owner:** argus

**T-0054** ext-disposition closed 3-value enum with structural-reject on missing/out-of-range

> Implement ext-disposition (tag=5) as exactly {0x00 ignore, 0x01 degrade, 0x02 refuse}; 0x03-0xFF reserved. A missing field-value or an out-of-range value is a structural reject naming ext-tok -- never a reader-chosen default (e.g. silently treating a missing disposition as ignore is a defect this task's test must catch).

- **Implements:** FR-013, FR-014
- **Depends on:** T-0052
- **DoD:** Each of the 3 valid disposition values decodes correctly; a missing disposition field and each of the 253 reserved values (spot-checked at 0x03, 0x7F, 0xFF) are rejected, with the rejection error naming the envelope's ext-tok.
- **Test:** `TestFR_013_014_ExtDispositionStructuralReject` (unit)
- **Owner:** hephaestus

**T-0055** PD-EXT-002: ext-fallback-ref mandatory-when-{ignore,degrade}

> Implement validator rule PD-EXT-002: ext-fallback-ref MUST be non-zero16 when ext-disposition is ignore or degrade; a zero16 fallback under either of those two dispositions is rejected. This is the 'declared fallback exists' half of FR-015; reachability (the other half) is T-0056.

- **Implements:** FR-015
- **Depends on:** T-0054
- **DoD:** An envelope with disposition=ignore or disposition=degrade and ext-fallback-ref=zero16 is rejected citing PD-EXT-002; disposition=refuse with ext-fallback-ref=zero16 is accepted (fallback is optional under refuse).
- **Test:** `TestFR_015_PD_EXT_002_FallbackMandatoryWhenIgnoreOrDegrade` (unit)
- **Owner:** hephaestus

**T-0056** Fallback reachability audit (fallback/primary exclusion)

> A declared fallback (ext-fallback-ref) must be a construct expressible entirely in the core feature set AND must have no other live path into the document than as this envelope's fallback -- audited so the fallback/primary exclusion pair is a genuine checked relationship, not a declared-but-unverified role. Requires M07's referential-graph walker (validate package, T-0118) to exist for the 'no other live path' half of the check; this task implements the audit rule itself and its unit-level fixture, coordinating with M07's structural-validation core.

- **Implements:** FR-015
- **Depends on:** T-0055, T-0118
- **DoD:** A fallback unit-id that is also referenced from a live (non-envelope) position elsewhere in the document is rejected as not exclusively reachable via the envelope; a fallback referenced only from the envelope is accepted.
- **Test:** `TestFR_015_FallbackReachabilityExclusion` (integration)
- **Owner:** hephaestus

**T-0057** PD-EXT-003: fallback minimum-content check, and close the contracts/README rule-registry gap

> Spec names validator rule PD-EXT-003 (a declared fallback must yield at least one extractable text unit or non-decorative mark) but the contracts pass never carried this rule id into contracts/README.md's generated rule registry (known gap, confirmed by direct grep of the frozen artifacts). This task both implements the check and files the correction to contracts/README.md's rule registry so CON-009/NFR-029's coverage audit can see PD-EXT-003. Do not silently invent the fix inside code only -- the registry document itself is out of sync with spec.md and that is a defect to close, not defer.

- **Implements:** FR-016
- **Depends on:** T-0056
- **DoD:** A fallback resolving to zero extractable text units and zero non-decorative marks is rejected citing PD-EXT-003; PD-EXT-003 appears in contracts/README.md's rule registry with a mapped conformance case, closing NFR-029's coverage gap for this specific rule id.
- **Test:** `TestFR_016_PD_EXT_003_FallbackMinimumContent` (unit)
- **Owner:** hephaestus

**T-0058** ext-position-key canonical ordering among unrecognised constructs

> ext-position-key (tag=7) reuses anchor-point verbatim. Canonical order among unrecognised constructs sharing one position is (ext-tok, ext-position-key). Implement the comparator and its use in canonical serialisation ordering; this is a pure ordering/comparator concern here, distinct from M13's later R1 SEQUENCE-ORDER merge classification which this ordering is stated to be 'identical to' for insertion semantics -- that equivalence is verified once M13 exists, not blocking this task.

- **Implements:** FR-012
- **Depends on:** T-0052
- **DoD:** Two envelopes sharing one anchor-point serialise in strict (ext-tok, ext-position-key) ascending order, deterministically, across repeated canonicalisation runs.
- **Test:** `TestFR_012_ExtPositionKeyCanonicalOrder` (unit)
- **Owner:** hephaestus

**T-0059** Reader disposition-enforcement policy (ignore/degrade/refuse runtime behaviour)

> An unimplemented construct or capability applies its declared disposition exactly, never silent omission or heuristic approximation. Implement the three runtime behaviours in the reader-facing decode path: ignore = skip the construct's presentation contribution while still emitting it as opaque data for extraction/preservation; degrade = substitute the audited fallback (T-0056) in its place; refuse = decline to render/extract past that point, naming the token, per the disposition's own semantics -- this is a distinct runtime-behaviour concern from the disposition value's own decode (T-0054).

- **Implements:** FR-107
- **Depends on:** T-0054, T-0055
- **DoD:** For each of the 3 dispositions, a reader-side integration test confirms the exact behaviour (opaque-preserve/skip, fallback-substitute, or refuse-naming-token) with no fourth, undocumented outcome ever produced.
- **Test:** `TestFR_107_DispositionRuntimeEnforcement` (integration)
- **Owner:** hephaestus

**T-0060** Writer-side retirement enforcement: a retired token is never reissued

> Beyond storing fm-retired-tokens (T-0051), the writer must refuse to mint a fresh registered- or owner-scoped-tier ext-tok whose owner-id+owner-local-seq pair already appears in the document's (or, for registered tier, the registry's) retired set. This closes the end-to-end 'never reissued' guarantee CP-011 non-negotiable #5 requires, mirroring the discipline already used for run_id/unit_id retirement.

- **Implements:** CON-020
- **Depends on:** T-0050, T-0051
- **DoD:** Attempting to write an ExtensionEnvelope whose ext-tok matches an entry already present in fm-retired-tokens fails writer-side with an error naming the retired token, before any encode occurs.
- **Test:** `TestCON_020_RetiredTokenNeverReissued` (unit)
- **Owner:** hephaestus

**T-0061** Registry review turnaround SLA published and tracked (CON-021)

> CON-021 is a governance/process obligation, not a wire mechanism: publish a maximum registry-review turnaround in business days for registered-tier ext-token issuance, and track observed turnaround against it. Attaches to M03's token-governance scope per plan.md's cross-cutting concerns (governance/licensing CP-014 gate). Deliver a checked-in policy document naming the SLA figure and the tracking mechanism (e.g. an issue-tracker label/metric), not application code.

- **Implements:** CON-021
- **Depends on:** T-0050
- **DoD:** A named maximum turnaround (business days) for registered-tier token review requests is committed to docs/token-registry-governance.md, cross-referenced from contracts/README.md's registry section, with a stated mechanism for recording observed turnaround per request; verified by a documentation-completeness check (required sections/figure present), not a wire-format conformance fixture.
- **Test:** `TestCON_021_RegistryTurnaroundSLAPublished` (unit)
- **Owner:** clio

**T-0062** Conformance suite: one round-trip fixture per disposition value

> Ship golden conformance fixtures (CON-010 style) for an ext-envelope with disposition=ignore, disposition=degrade (with a valid audited fallback), and disposition=refuse, each round-tripping to identical canonical octets. This directly satisfies M03's exit criterion 'an envelope with each of the 3 dispositions round-trips.'

- **Implements:** FR-012, FR-013
- **Depends on:** T-0052, T-0054
- **DoD:** All 3 fixtures pass round-trip (decode-then-reencode-then-byte-compare) in CI, and are registered in the golden conformance corpus index alongside their FR-012/FR-013 rule-id mapping.
- **Test:** `TestFR_012_013_DispositionRoundTripCorpus` (conformance)
- **Owner:** momus

**T-0063** Conformance vector: missing-disposition envelope rejected naming token

> Ship a golden conformance fixture with the ext-disposition field entirely absent from an otherwise well-formed ext-envelope, and confirm rejection names the offending ext-tok. Directly satisfies M03's exit criterion for FR-014.

- **Implements:** FR-014
- **Depends on:** T-0054
- **DoD:** The fixture is rejected in CI with an error message containing the fixture's known ext-tok value; two independent decode calls on the same fixture produce byte-identical error output.
- **Test:** `TestFR_014_MissingDispositionRejectedNamingToken` (conformance)
- **Owner:** momus

**T-0064** Conformance vector: extension-envelope fallback-reference cycle rejected

> Plan.md Section 10 first-class task (g): a fallback-reference cycle (envelope A's fallback resolves to envelope B, B's fallback resolves back to A) must be rejected as one of FR-109's 5 named cycle-detection edge kinds -- 'extension-envelope fallback references (S6)' is explicitly one of the 5. This task builds the envelope-side fixture and asserts rejection naming every edge of the cycle; the generic cycle-detector traversal itself is M07's Structural Validation Core deliverable (T-0118), so this task is dual-owned with M07 per the spine's cross-cutting-concerns note and must not duplicate that traversal, only exercise it against an envelope-shaped cycle. NOTE: the milestone spine currently lists M03 as depending only on M01/M02; this task's dependency on M07's T-0118 is a genuine build-order gap in the spine, flagged for the spine owner, not resolved here beyond the task-level edge.

- **Implements:** FR-015
- **Depends on:** T-0056, T-0118
- **DoD:** A 2-envelope fallback cycle fixture is rejected before any structural-ceiling check runs, naming both edges of the cycle; verified against the same cycle-detector M07 ships (no envelope-local reimplementation).
- **Test:** `TestFR_015_FR_109_ExtFallbackCycleRejected` (conformance)
- **Owner:** momus

**T-0065** ext-token owner-id partition boundary conformance vectors

> CON-010 requires an at-limit and one-past-limit fixture at every named boundary. Ship fixtures at the 4 owner-id partition boundaries: 0x00000000/0x00000001 (reserved-invalid to registered), 0x7FFFFFFF/0x80000000 (registered to owner-scoped), 0xFFFFFFFE/0xFFFFFFFF (owner-scoped to retired-tombstone), and confirm two independent decodes agree on the verdict at each boundary.

- **Implements:** CON-020
- **Depends on:** T-0050
- **DoD:** All 6 boundary values (3 pairs) decode to their documented partition class with two independent decode runs agreeing byte-for-byte on the classification, checked into the golden corpus.
- **Test:** `TestCON_020_OwnerIdPartitionBoundaries` (conformance)
- **Owner:** momus

**T-0066** Continuous fuzzing harness for ExtensionEnvelope decode (CP-012)

> EXT_ENVELOPE decodes untrusted bytes (arbitrary ext-tok, malformed ext-payload-length, disposition out-of-range, garbage fallback-ref) and per CP-012 must be wired into the continuous fuzzing harness as part of its own exit criteria. Seed corpus from T-0062/T-0063/T-0065's conformance fixtures; the fuzzer must never find a panic, unbounded allocation, or a decode that silently accepts an out-of-range disposition.

- **Implements:** FR-107, FR-014
- **Depends on:** T-0052, T-0053, T-0054
- **DoD:** go test -fuzz=FuzzExtEnvelopeDecode runs for the CI-mandated minimum duration with zero crashes/panics and zero allocations exceeding the MAX_DECODED_UNIT ceiling, feeding M19's harness-maturity check.
- **Test:** `FuzzExtEnvelopeDecode` (fuzz)
- **Owner:** prometheus

---

### M04: Identity & Anchor

NFC-scoped text and CSPRNG-minted run/unit identity are the addressing scheme every Run, Annotation, history op, and merge classification depends on.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0067 | Implement CSPRNG identity minting primitive (run_id/unit_id) | FR-019, FR-021, FR-023 | None | hephaestus | `TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded` (unit) |
| T-0068 | Implement base_ordinal Unicode-scalar-value positioning within a run | CON-001 | T-0067 | hephaestus | `TestCON_001_BaseOrdinalIsScalarValueScopedToSegment` (unit) |
| T-0069 | Implement Run split preserving run_id, shifting only base_ordinal | FR-020 | T-0067, T-0068 | hephaestus | `TestFR_020_SplitPreservesRunID` (unit) |
| T-0070 | Implement Run merge as a pure syntactic predicate over matching lineage | FR-020 | T-0067, T-0068 | hephaestus | `TestFR_020_MergeIsPureSyntacticPredicate` (unit) |
| T-0071 | Implement Run move/reorder preserving run_id across storage relocation | FR-020 | T-0067, T-0068 | hephaestus | `TestFR_020_MoveReorderPreservesRunID` (unit) |
| T-0072 | Implement fresh-mint-on-duplicate/paste burst detection | FR-022 | T-0067 | hephaestus | `TestFR_022_PasteMintsFreshIdentity` (unit) |
| T-0073 | Implement per-run/per-block independent NFC scoping | CON-002 | None | hephaestus | `TestCON_002_NFCScopingIsPerRunNotCrossBoundary` (unit) |
| T-0074 | Implement writer-side rejection of non-NFC input | CON-003 | T-0073 | hephaestus | `TestCON_003_WriterRejectsNonNFCRatherThanConverting` (unit) |
| T-0075 | Implement content-identity addressing type (run_id, base_ordinal) | FR-025 | T-0067, T-0068 | hephaestus | `TestFR_025_AddressingIsContentIdentityOnly` (unit) |
| T-0076 | Implement Annotation anchor boundary-behaviour closed enum | FR-026 | T-0075 | hephaestus | `TestFR_026_BoundaryBehaviourIsClosedFourValueEnum` (conformance) |
| T-0077 | Verify boundary-behaviour persistence unchanged across save/load | FR-027 | T-0076 | hephaestus | `TestFR_027_BoundaryBehaviourRoundTripsUnchanged` (integration) |
| T-0078 | Implement orphan-carriage retention when anchored content is fully deleted | FR-028 | T-0076 | hephaestus | `TestFR_028_FullDeletionOrphansRatherThanDrops` (unit) |
| T-0079 | Implement deterministic nearest-surviving-unit resolution for orphans | FR-029 | T-0078 | hephaestus | `TestFR_029_OrphanResolutionIsDeterministic` (unit) |
| T-0080 | Implement orphan record field population (author, quoted text, neighbours) | FR-030 | T-0078, T-0079 | hephaestus | `TestFR_030_OrphanRecordCapturesAllFourFields` (unit) |
| T-0081 | Wire exactly-one-language-tag field onto TextBlock/Run | FR-031 | T-0068 | hephaestus | `TestFR_031_TextSpanRequiresLanguageTagField` (unit) |
| T-0082 | Implement PD-LANG-001 validator rule: reject unresolvable language tag | FR-031 | T-0081 | hephaestus | `TestFR_031_PDLANG001RejectsUnresolvableTag` (conformance) |
| T-0083 | A-FIELD-ROLE audit: assert no persisted counted position exists in content package | CON-001 | T-0068 | momus | `TestCON_001_AFieldRoleAuditHasNoPersistedCountedPosition` (unit) |
| T-0084 | Conformance: run_id uniqueness across document copy, fork, and branch | FR-019 | T-0067 | momus | `TestFR_019_IdentifierUniqueAcrossCopyForkBranch` (conformance) |
| T-0085 | Fuzz: run_id preservation across split/merge/move/reorder/save/load/undo | FR-020 | T-0069, T-0070, T-0071 | prometheus | `TestFR_020_FuzzRunIDPreservationAcrossOperationSequence` (fuzz) |
| T-0086 | Conformance: orphan-carriage survives append-only ledger reload | FR-027, FR-028, FR-029, FR-030 | T-0077, T-0080 | momus | `TestM04_OrphanCarriageSurvivesLedgerReload` (conformance) |

**T-0067** Implement CSPRNG identity minting primitive (run_id/unit_id)

> 128-bit token minted via crypto/rand for every content unit/run. Minting API carries no actor/device/clock/session parameter (DP-003, NFR-005 allowlist exclusion). Document and enforce the collision bound (P<2^-60 below 2^34 mints/lineage, Section 5 row 22). Requires M01's pdlfmt varint/TLV encoding for the 16-octet opaque token wire form to exist first, but no M04-internal predecessor.

- **Implements:** FR-019, FR-021, FR-023
- **Depends on:** None
- **DoD:** MintID() returns a 128-bit crypto/rand-sourced token; its signature accepts no actor/device/clock/session argument; a documented comment and a probability-bound unit test assert collision probability < 2^-60 at 2^34 simulated mints (statistical, not exhaustive); no code path re-mints or reissues a previously returned value.
- **Test:** `TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded` (unit)
- **Owner:** hephaestus

**T-0068** Implement base_ordinal Unicode-scalar-value positioning within a run

> base_ordinal counts Unicode scalar values within one segment only (CON-001); it is a derived offset, never a document-wide persisted counted position. Implements the arithmetic used by split/merge/addressing tasks.

- **Implements:** CON-001
- **Depends on:** T-0067
- **DoD:** Run.BaseOrdinal is computed in Unicode scalar values scoped to its own segment; no field or function in the content package accepts or returns a cross-segment absolute character count.
- **Test:** `TestCON_001_BaseOrdinalIsScalarValueScopedToSegment` (unit)
- **Owner:** hephaestus

**T-0069** Implement Run split preserving run_id, shifting only base_ordinal

> Splitting a run at a scalar-value boundary produces two Run records sharing the original run_id; only base_ordinal differs between the resulting pieces (DP-003/DP-004).

- **Implements:** FR-020
- **Depends on:** T-0067, T-0068
- **DoD:** SplitRun(r, at) returns two Runs whose run_id both equal r.run_id and whose base_ordinal values are r.base_ordinal and r.base_ordinal+at respectively, for every valid split point in a property test over random run lengths.
- **Test:** `TestFR_020_SplitPreservesRunID` (unit)
- **Owner:** hephaestus

**T-0070** Implement Run merge as a pure syntactic predicate over matching lineage

> Two adjacent Run records merge back into one iff they share run_id and are base_ordinal-contiguous; merge is a pure function with no side channel (matches split's inverse).

- **Implements:** FR-020
- **Depends on:** T-0067, T-0068
- **DoD:** CanMergeRuns(a,b) returns true iff a.run_id==b.run_id and b.base_ordinal==a.base_ordinal+len(a); MergeRuns is the exact left-inverse of SplitRun for all generated split points.
- **Test:** `TestFR_020_MergeIsPureSyntacticPredicate` (unit)
- **Owner:** hephaestus

**T-0071** Implement Run move/reorder preserving run_id across storage relocation

> Moving or reordering a run within the document (changing its position in storage/traversal order) never alters run_id; only positional metadata external to the run record changes.

- **Implements:** FR-020
- **Depends on:** T-0067, T-0068
- **DoD:** MoveRun/ReorderRuns operations round-trip run_id unchanged for every generated move/reorder sequence in a table-driven test covering single-move, multi-move, and cyclic-reorder cases.
- **Test:** `TestFR_020_MoveReorderPreservesRunID` (unit)
- **Owner:** hephaestus

**T-0072** Implement fresh-mint-on-duplicate/paste burst detection

> A duplicate or paste operation is treated as a new contiguous typing burst and therefore mints a fresh run_id rather than reusing the source run's identifier (DP-003 burst-boundary rule).

- **Implements:** FR-022
- **Depends on:** T-0067
- **DoD:** PasteContent()/DuplicateRange() always produce Run records with a freshly minted run_id distinct from every source run_id, verified for single-run and multi-run paste payloads.
- **Test:** `TestFR_022_PasteMintsFreshIdentity` (unit)
- **Owner:** hephaestus

**T-0073** Implement per-run/per-block independent NFC scoping

> Every text segment (Run.run-text) and every identifier string is independently NFC-normalized at its own boundary; normalization never spans across a run/block boundary (CQ-009).

- **Implements:** CON-002
- **Depends on:** None
- **DoD:** IsNFCScoped(run) validates a run's text is NFC in isolation; a two-run adversarial pair whose concatenation would normalize differently than each run alone is proven NOT to be renormalized across the boundary.
- **Test:** `TestCON_002_NFCScopingIsPerRunNotCrossBoundary` (unit)
- **Owner:** hephaestus

**T-0074** Implement writer-side rejection of non-NFC input

> A writer presented with non-NFC text input refuses the write and reports the offending unit rather than silently renormalizing it (PD-NFC-001/002).

- **Implements:** CON-003
- **Depends on:** T-0073
- **DoD:** WriteRun() returns a named rejection error (not a normalized/mutated value) for any input string that is not already in NFC form, for a corpus of decomposed/non-NFC Unicode test strings.
- **Test:** `TestCON_003_WriterRejectsNonNFCRatherThanConverting` (unit)
- **Owner:** hephaestus

**T-0075** Implement content-identity addressing type (run_id, base_ordinal)

> A shared (run_id, base_ordinal) address/anchor type used by every consumer that must address a range/point in text by content identity: formatting ranges, comments, links, cross-references, bookmarks, change records. No consumer is permitted to persist a counted absolute text-unit position instead.

- **Implements:** FR-025
- **Depends on:** T-0067, T-0068
- **DoD:** Anchor{RunID, BaseOrdinal} is the sole exported addressing type in the content package; a static grep-based test asserts no other exported struct field in content/, document-model consumers (annotation, xref, table) carries an int-typed absolute position field.
- **Test:** `TestFR_025_AddressingIsContentIdentityOnly` (unit)
- **Owner:** hephaestus

**T-0076** Implement Annotation anchor boundary-behaviour closed enum

> Each annotation endpoint declares exactly one of the 4 closed boundary behaviours {inside, outside, inside-if-before, inside-if-after} per document.abnf S3.1.

- **Implements:** FR-026
- **Depends on:** T-0075
- **DoD:** AnchorBoundary is a closed 4-value enum; decoding an Annotation record with a 5th/undefined boundary value is a structural rejection, exercised by an at-limit (value 3) and one-past-limit (value 4) fixture pair.
- **Test:** `TestFR_026_BoundaryBehaviourIsClosedFourValueEnum` (conformance)
- **Owner:** hephaestus

**T-0077** Verify boundary-behaviour persistence unchanged across save/load

> Because sealed ledger segments are never rewritten in place (M02 append-only guarantee), an Annotation's boundary-behaviour value read back after a save/load cycle is byte-identical to the value written. This task adds the round-trip conformance test confirming that invariant for annotations specifically (the merge-preservation half is verified again at M13 once the merge classifier exists).

- **Implements:** FR-027
- **Depends on:** T-0076
- **DoD:** A save-then-load round-trip test over documents containing all 4 boundary-behaviour values at both endpoints shows zero value drift.
- **Test:** `TestFR_027_BoundaryBehaviourRoundTripsUnchanged` (integration)
- **Owner:** hephaestus

**T-0078** Implement orphan-carriage retention when anchored content is fully deleted

> Deleting every content unit an annotation anchors to does not delete the annotation; it converts to an orphan-carriage record instead (DP-004 orphan carriage).

- **Implements:** FR-028
- **Depends on:** T-0076
- **DoD:** DeleteRange() that removes 100% of an annotation's anchored span leaves the annotation present in the document model, flagged orphaned, never silently dropped.
- **Test:** `TestFR_028_FullDeletionOrphansRatherThanDrops` (unit)
- **Owner:** hephaestus

**T-0079** Implement deterministic nearest-surviving-unit resolution for orphans

> An orphaned annotation resolves to the nearest surviving unit via orphan-prev/orphan-next fields (document.abnf S3.2), deterministically for a given document state.

- **Implements:** FR-029
- **Depends on:** T-0078
- **DoD:** ResolveOrphan(a) returns the same (prev,next) surviving-unit pair on every call for a fixed document state, including the boundary cases of orphaning at document start and document end.
- **Test:** `TestFR_029_OrphanResolutionIsDeterministic` (unit)
- **Owner:** hephaestus

**T-0080** Implement orphan record field population (author, quoted text, neighbours)

> At the moment of orphaning, capture orphan-author, orphan-quoted (the deleted anchored text), orphan-prev and orphan-next into the orphan record so none of the four values is lost.

- **Implements:** FR-030
- **Depends on:** T-0078, T-0079
- **DoD:** Immediately after an orphaning delete, the resulting orphan record's author, quoted_text, orphan-prev and orphan-next fields are all non-empty/populated and match the pre-delete state's author and deleted text exactly.
- **Test:** `TestFR_030_OrphanRecordCapturesAllFourFields` (unit)
- **Owner:** hephaestus

**T-0081** Wire exactly-one-language-tag field onto TextBlock/Run

> Data-model half of FR-031: every text span (TextBlock.tb-language-ref / Run.lang-ref) resolves to exactly one language tag reference; the field is mandatory, not optional.

- **Implements:** FR-031
- **Depends on:** T-0068
- **DoD:** TextBlock and Run structs require a non-nil language-tag reference at construction time; a decode of a record lacking the field is rejected before any resolution logic runs.
- **Test:** `TestFR_031_TextSpanRequiresLanguageTagField` (unit)
- **Owner:** hephaestus

**T-0082** Implement PD-LANG-001 validator rule: reject unresolvable language tag

> Closes the gap the inventory flags as PARTIAL: spec's PD-LANG-001 rule ('reject if unresolvable') exists as a field per T-0081 but was never wired into a validation check. Adds the actual validator rule and its entry in the generated rule registry (contracts/README.md S7) so it participates in the NFR-029 coverage check.

- **Implements:** FR-031
- **Depends on:** T-0081
- **DoD:** A text span whose language-tag reference does not resolve to any RegistryExcerpt/BCP-47 entry in the document is rejected by the validator naming PD-LANG-001 and the offending unit id; PD-LANG-001 appears in contracts/README.md's rule registry with >=1 passing conformance case.
- **Test:** `TestFR_031_PDLANG001RejectsUnresolvableTag` (conformance)
- **Owner:** hephaestus

**T-0083** A-FIELD-ROLE audit: assert no persisted counted position exists in content package

> Per plan.md Section 10 item 2, base_ordinal and run_id must pass the A-FIELD-ROLE audit distinguishing IDENTITY-COMPONENT fields from positional arithmetic. This task builds the checked audit (not just the field's own arithmetic from T-0068) as a standing test/lint rule the rest of the content model is built against.

- **Implements:** CON-001
- **Depends on:** T-0068
- **DoD:** A registry-driven audit enumerates every content-package struct field and classifies it IDENTITY-COMPONENT or POSITIONAL-ARITHMETIC; the audit fails CI if any field is untagged, and passes with zero fields classified as a persisted document-wide counted position.
- **Test:** `TestCON_001_AFieldRoleAuditHasNoPersistedCountedPosition` (unit)
- **Owner:** momus

**T-0084** Conformance: run_id uniqueness across document copy, fork, and branch

> Dedicated conformance corpus proving a content unit's identifier stays unique across the document and every copy/fork/branch derived from it, for the lifetime of the lineage (not just within one file).

- **Implements:** FR-019
- **Depends on:** T-0067
- **DoD:** A 3-file fixture set (original, copy, fork) is decoded together and shown to contain zero colliding run_id/unit_id values across all three, checked into the conformance corpus with a case id.
- **Test:** `TestFR_019_IdentifierUniqueAcrossCopyForkBranch` (conformance)
- **Owner:** momus

**T-0085** Fuzz: run_id preservation across split/merge/move/reorder/save/load/undo

> The milestone's own exit criterion: a continuous fuzzing target that generates random sequences of split/merge/move/reorder/save/load/undo operations and asserts run_id is preserved through every step, per CP-012.

- **Implements:** FR-020
- **Depends on:** T-0069, T-0070, T-0071
- **DoD:** A go-fuzz/native fuzz target (FuzzRunIDPreservation) runs a corpus-seeded campaign of >=10,000 generated operation sequences with zero run_id-loss failures, and is wired into the CI fuzzing harness.
- **Test:** `TestFR_020_FuzzRunIDPreservationAcrossOperationSequence` (fuzz)
- **Owner:** prometheus

**T-0086** Conformance: orphan-carriage survives append-only ledger reload

> End-to-end exit-criteria vector: orphan an annotation, save, reload from the append-only ledger, and confirm author, quoted_text, both neighbour identifiers, and the original boundary-behaviour value are all bit-identical to the pre-save state.

- **Implements:** FR-027, FR-028, FR-029, FR-030
- **Depends on:** T-0077, T-0080
- **DoD:** The round-trip fixture (orphan -> save -> reload) shows zero-diff on all four FR-030 fields plus the FR-026/FR-027 boundary-behaviour value, checked into the conformance corpus.
- **Test:** `TestM04_OrphanCarriageSurvivesLedgerReload` (conformance)
- **Owner:** momus

---

### M05: Extraction

TR-011's font/shaping/layout/crypto-free reference reader is the cheapest early win that end-to-end validates M01-M04 before harder layers begin.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0087 | extract package walker core (CONTENT-only, streaming, dependency-isolated) | FR-041 | None | hephaestus | `TestFR_041_WalksContentSegmentsOnlyNoHeavyDeps` (integration) |
| T-0088 | Structural locator emission (unit identity + live scalar position) | FR-042 | T-0087 | hephaestus | `TestFR_042_EmitsUnitIdentityAndScalarPositionLocator` (unit) |
| T-0089 | Reading-order text fidelity (exact authored scalar sequence) | FR-035 | T-0087, T-0088 | hephaestus | `TestFR_035_ExtractionReproducesAuthoredScalarSequenceExactly` (conformance) |
| T-0090 | Per-unit language tag emission | FR-043 | T-0087 | hephaestus | `TestFR_043_EmitsExactlyOneLanguageTagPerUnit` (conformance) |
| T-0091 | Extraction-view descriptive metadata (font/shaping/layout-free) | FR-044 | T-0087 | hephaestus | `TestFR_044_EmitsDescriptiveMetadataWithoutHeavyDecoders` (unit) |
| T-0092 | Page adjunct emission when pagination artefact is current | FR-045 | T-0088 | hephaestus | `TestFR_045_EmitsPageAdjunctWhenPaginationCurrent` (conformance) |
| T-0093 | Page adjunct omission and staleness reporting | FR-046 | T-0092 | hephaestus | `TestFR_046_OmitsPageAdjunctAndReportsStaleWhenPaginationAbsentOrStale` (conformance) |
| T-0094 | Streaming, early-emission extraction API | FR-047 | T-0087, T-0088, T-0089 | hephaestus | `TestFR_047_EmitsEarlierUnitTextBeforeReadingLaterUnitOctets` (integration) |
| T-0095 | Abandonable extraction with no remainder-read obligation | FR-048 | T-0094 | hephaestus | `TestFR_048_AbandoningExtractionReadsNoFurtherOctets` (integration) |
| T-0096 | Zero-trust-material extraction path | FR-049 | T-0087 | hephaestus | `TestFR_049_ExtractionSucceedsWithZeroTrustMaterial` (integration) |
| T-0097 | Synthetic 1 GiB / 10,000-page benchmark corpus generator |  | None | prometheus | `TestFixtureGen_1GiB10000PageCorpusIsDeterministic` (integration) |
| T-0098 | NFR-012 octets-read budget benchmark | NFR-012 | T-0087, T-0089, T-0097 | prometheus | `BenchmarkNFR_012_ExtractionReadsWithin15PercentOfFileOctets` (benchmark) |
| T-0099 | NFR-013 processor-time budget benchmark | NFR-013 | T-0087, T-0089, T-0097 | prometheus | `BenchmarkNFR_013_ExtractionCompletesWithin20SecondsProcessorTime` (benchmark) |
| T-0100 | NFR-014 peak-memory budget benchmark | NFR-014 | T-0087, T-0089, T-0097 | prometheus | `BenchmarkNFR_014_ExtractionPeakMemoryStaysFlatAndBounded` (benchmark) |
| T-0101 | TR-011 line-budget and dependency-isolation CI gate | TR-011 | T-0087, T-0096 | prometheus | `TestTR_011_ExtractPackageUnder1000LinesNoHeavyDeps` (unit) |

**T-0087** extract package walker core (CONTENT-only, streaming, dependency-isolated)

> Implement extract.Walk(r io.ReaderAt) as the base streaming primitive: walks only CONTENT-typed segments in storage order, reading through the M01 SegmentTableSlot inventory and the M02 ledger/place() segment reader, skipping RESOURCE/HISTORY/ATTEST segments entirely. The extract package must have zero import edge to integrity, render, or merge (module-boundary requirement underlying TR-011 and FR-049). Depends on M01 container structs and M02 ledger read path existing, and on M04's run/unit identity being mintable and readable, though those tasks are owned by earlier milestones.

- **Implements:** FR-041
- **Depends on:** None
- **DoD:** extract package compiles with no import edge to integrity/render/merge, verified by a `go list -deps ./extract/...` CI check; Walk() visits every CONTENT segment in ascending storage ordinal on a mixed-segment-type fixture and never touches a RESOURCE/HISTORY/ATTEST segment's payload bytes.
- **Test:** `TestFR_041_WalksContentSegmentsOnlyNoHeavyDeps` (integration)
- **Owner:** hephaestus

**T-0088** Structural locator emission (unit identity + live scalar position)

> Emit an extract.Locator{UnitID, ScalarOffset} per text unit during the walk. ScalarOffset is computed live by counting Unicode scalar values within the current unit only, per CON-001 — never read from a persisted counted-position field, since none exists.

- **Implements:** FR-042
- **Depends on:** T-0087
- **DoD:** Locator.UnitID matches the unit's run_id/unit_id and Locator.ScalarOffset matches an independently-computed scalar count for 3 multi-run fixtures, including one containing combining-mark clusters.
- **Test:** `TestFR_042_EmitsUnitIdentityAndScalarPositionLocator` (unit)
- **Owner:** hephaestus

**T-0089** Reading-order text fidelity (exact authored scalar sequence)

> Concatenate Run.run-text in reading order per text unit, emitting the authored Unicode scalar sequence exactly as stored. Text is NFC-scoped at write time by M04 and must never be renormalized at extraction time.

- **Implements:** FR-035
- **Depends on:** T-0087, T-0088
- **DoD:** Extracted text is byte-for-byte identical to the golden corpus text for every FR-035 conformance fixture, including one fixture with pre-composed scalars and one with adjacent-but-distinct valid scalar sequences.
- **Test:** `TestFR_035_ExtractionReproducesAuthoredScalarSequenceExactly` (conformance)
- **Owner:** hephaestus

**T-0090** Per-unit language tag emission

> Resolve TextBlock.tb-language-ref / Run.lang-ref to exactly one language tag per emitted text unit via extract.LanguageOf(unit). This closes the previously-uncited FR-043 mechanism gap (plan.md never wires the extraction-view's per-unit-language emission to anything) by giving it a concrete call site and conformance case.

- **Implements:** FR-043
- **Depends on:** T-0087
- **DoD:** Every unit in a 3-language mixed-language fixture reports exactly one resolved tag; a unit whose language reference is unresolvable is reported via a distinct per-unit error value in the result rather than silently omitted.
- **Test:** `TestFR_043_EmitsExactlyOneLanguageTagPerUnit` (conformance)
- **Owner:** hephaestus

**T-0091** Extraction-view descriptive metadata (font/shaping/layout-free)

> Implement extract.Metadata(doc) returning title, page count, language, and other descriptive Frontmatter-sourced metadata, computed without constructing any font/shaping/layout/graphics decoder. Library-level only: wiring this into the actual `extract` CLI verb's stdout payload (currently mapped only to `inspect` per the plan.md review's disclosed gap) belongs to M18's CLI surface, not this milestone — flag that wiring gap forward to the M18 CLI task.

- **Implements:** FR-044
- **Depends on:** T-0087
- **DoD:** Metadata() returns non-empty title/page-count/language fields for a fixture with populated Frontmatter and the documented zero-value for a fixture with an absent field, with zero calls into `render` or `integrity`.
- **Test:** `TestFR_044_EmitsDescriptiveMetadataWithoutHeavyDecoders` (unit)
- **Owner:** hephaestus

**T-0092** Page adjunct emission when pagination artefact is current

> When a PageDirectory artifact's input-digest matches current content (not stale), emit a page-number adjunct alongside each Locator. Library-level only; wiring this into the `extract` CLI verb's stdout payload belongs to M18's CLI surface (same disclosed plan.md gap as FR-044/FR-046) — flag forward to the M18 CLI task rather than wiring it here.

- **Implements:** FR-045
- **Depends on:** T-0088
- **DoD:** A fixture with a fresh PageDirectory yields a non-nil page adjunct on every Locator, matching the golden page map exactly.
- **Test:** `TestFR_045_EmitsPageAdjunctWhenPaginationCurrent` (conformance)
- **Owner:** hephaestus

**T-0093** Page adjunct omission and staleness reporting

> When no PageDirectory exists, or its input-digest mismatches current content, omit page adjuncts entirely and set a PaginationStale flag on the extraction result rather than emitting a guessed page number. Library-level only; CLI-payload wiring for this flag belongs to M18's CLI surface (same disclosed plan.md gap as FR-044/FR-045) — flag forward to the M18 CLI task rather than wiring it here.

- **Implements:** FR-046
- **Depends on:** T-0092
- **DoD:** A fixture with a digest-mismatched PageDirectory and a fixture with none both yield nil page adjuncts on every Locator and PaginationStale=true on the result.
- **Test:** `TestFR_046_OmitsPageAdjunctAndReportsStaleWhenPaginationAbsentOrStale` (conformance)
- **Owner:** hephaestus

**T-0094** Streaming, early-emission extraction API

> Expose extraction as a pull-based iterator/channel so a caller receives the complete text of unit N before any octet of unit N+1's segment is read from storage. Verify with an instrumented ReaderAt recording read offsets.

- **Implements:** FR-047
- **Depends on:** T-0087, T-0088, T-0089
- **DoD:** For a 3-unit fixture, the recorded read-offset trace shows unit 1's full text delivered to the caller strictly before any read into unit 2's or unit 3's segment byte range.
- **Test:** `TestFR_047_EmitsEarlierUnitTextBeforeReadingLaterUnitOctets` (integration)
- **Owner:** hephaestus

**T-0095** Abandonable extraction with no remainder-read obligation

> Support caller-initiated cancellation (context cancellation or early iterator break) after any unit, guaranteeing the walker performs zero further reads once cancelled.

- **Implements:** FR-048
- **Depends on:** T-0094
- **DoD:** Cancelling after unit 2 of a 5-unit fixture results in zero additional Read calls beyond what units 1-2 required, verified by call-count assertion on the instrumented reader.
- **Test:** `TestFR_048_AbandoningExtractionReadsNoFurtherOctets` (integration)
- **Owner:** hephaestus

**T-0096** Zero-trust-material extraction path

> Guarantee extraction succeeds with no signature/trust material loaded and no verification performed. Enforce at compile time that package `extract` has no import edge to `integrity`, and at runtime that extraction of a document with a corrupted or absent ATTEST segment still succeeds.

- **Implements:** FR-049
- **Depends on:** T-0087
- **DoD:** `go list -deps ./extract/...` contains no `integrity` package edge (CI-enforced); extracting a fixture with a zeroed-out ATTEST segment returns full, correct text with no error and no verification call recorded.
- **Test:** `TestFR_049_ExtractionSucceedsWithZeroTrustMaterial` (integration)
- **Owner:** hephaestus

**T-0097** Synthetic 1 GiB / 10,000-page benchmark corpus generator

> Build a deterministic generator producing a valid 1 GiB, 10,000-page PDL document (varied unit sizes, mixed languages, embedded RESOURCE segments) as the shared fixture for the NFR-012/013/014 benchmarks (T-0098..T-0100). NFR-011's reference measurement configuration is not named anywhere in the frozen artifacts (spec/plan/data-model/contracts) despite roughly a dozen NFRs depending on it; this task must pick and document a concrete stand-in machine spec (CPU, cores, RAM, storage class, OS) pending an Eyvar/themis ruling on NFR-011, and flag that dependency in the fixture's own README. Intentionally requirement-less (implements: []): this task is shared test-infrastructure — a fixture generator, not itself a functional or non-functional requirement — under the ISO/IEC/IEEE 29119 test-infrastructure cross-cutting concern, existing solely to make T-0098/T-0099/T-0100 buildable and repeatable.

- **Implements:** None (see description)
- **Depends on:** None
- **DoD:** Generator produces a byte-identical fixture across two runs on the same seed, the fixture validates clean against the M07 structural validator once available, and the fixture plus its documented reference-machine stand-in are checked into the conformance corpus.
- **Test:** `TestFixtureGen_1GiB10000PageCorpusIsDeterministic` (integration)
- **Owner:** prometheus

**T-0098** NFR-012 octets-read budget benchmark

> Benchmark full-text extraction of the T-0097 corpus and assert total octets read from storage is <=15% of the file's total octet length.

- **Implements:** NFR-012
- **Depends on:** T-0087, T-0089, T-0097
- **DoD:** Benchmark run against the corpus reports octets-read/file-size <= 0.15, recorded in CI benchmark output; a regression above 0.15 fails the build.
- **Test:** `BenchmarkNFR_012_ExtractionReadsWithin15PercentOfFileOctets` (benchmark)
- **Owner:** prometheus

**T-0099** NFR-013 processor-time budget benchmark

> Benchmark full-text extraction wall/CPU time of the T-0097 corpus and assert completion within 20 seconds of processor time on the reference configuration documented in T-0097.

- **Implements:** NFR-013
- **Depends on:** T-0087, T-0089, T-0097
- **DoD:** Benchmark reports processor time <=20s on the documented reference machine spec; CI fails on regression above budget.
- **Test:** `BenchmarkNFR_013_ExtractionCompletesWithin20SecondsProcessorTime` (benchmark)
- **Owner:** prometheus

**T-0100** NFR-014 peak-memory budget benchmark

> Benchmark peak resident memory during extraction of the T-0097 corpus and assert it stays <=33,554,432 octets, then confirm the bound stays flat (not proportional) when re-run against a 4x-scaled synthetic corpus.

- **Implements:** NFR-014
- **Depends on:** T-0087, T-0089, T-0097
- **DoD:** Peak RSS measured via runtime/pprof memory profiling stays <=33,554,432 bytes on both the base and 4x-scaled corpus.
- **Test:** `BenchmarkNFR_014_ExtractionPeakMemoryStaysFlatAndBounded` (benchmark)
- **Owner:** prometheus

**T-0101** TR-011 line-budget and dependency-isolation CI gate

> Add a CI check enforcing the `extract` package's reference implementation stays under 1000 non-test source lines (gofmt-normalized) and has zero import edges to font/shaping/layout/graphics or crypto packages, composing with the integrity-isolation check into one gate.

- **Implements:** TR-011
- **Depends on:** T-0087, T-0096
- **DoD:** CI job fails if `extract`'s non-test source line count exceeds 1000 or if `go list -deps` surfaces a forbidden import edge; passes on the current tree.
- **Test:** `TestTR_011_ExtractPackageUnder1000LinesNoHeavyDeps` (unit)
- **Owner:** prometheus

---

### M06: Signature Primitive (EdDSA-Protodoc-1)

The 7-step wrapper is vector-tested in isolation because every Signature, CoverageDescriptor, sign/verify/redact/publish/migrate task depends on it.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0102 | ParamSet allowlist type and v1 registry | CON-015 | None | argus | `TestCON_015_ParamSetAllowlist` (unit) |
| T-0103 | Deterministic Sign() wrapper over stdlib crypto/ed25519 | NFR-006 | None | argus | `TestNFR_006_SignDeterministic` (unit) |
| T-0104 | Small-order point constant table with pinned digest | CON-015 | None | argus | `TestCON_015_SmallOrderTableDigestPinned` (unit) |
| T-0105 | Steps 1-2: canonical-encoding and small-order check for A | CON-015 | T-0104 | argus | `TestCON_015_Step1Step2_PublicKeyChecks` (unit) |
| T-0106 | Steps 3-4: canonical-encoding and small-order check for R | CON-015 | T-0104 | argus | `TestCON_015_Step3Step4_RChecks` (unit) |
| T-0107 | Step 5: scalar S range check against L | CON-015 | None | argus | `TestCON_015_Step5_ScalarRangeCheck` (unit) |
| T-0108 | Verify() 7-step orchestration with short-circuit and stdlib delegation | CON-015 | T-0103, T-0105, T-0106, T-0107 | argus | `TestCON_015_Verify7StepShortCircuit` (unit) |
| T-0109 | EdDSA-Protodoc-1 conformance vector corpus | CON-015 | T-0108 | argus | `EDDSA_P1_CONFORMANCE_CORPUS_V1` (conformance) |
| T-0110 | Sign-twice determinism vector suite | NFR-006 | T-0103, T-0108 | argus | `TestNFR_006_SignTwiceIdenticalOctets` (unit) |
| T-0111 | Off-allowlist sig-param-set rejection gate | CON-015 | T-0102, T-0108 | argus | `TestCON_015_OffAllowlistParamSetRejected` (unit) |
| T-0112 | argus security review of the EdDSA-Protodoc-1 primitive | CON-015, NFR-006 | T-0108, T-0109, T-0110, T-0111 | argus | `TestCON_015_ArgusSecurityReviewChecklist` (integration) |

**T-0102** ParamSet allowlist type and v1 registry

> Define the param_set_id type (u16) per integrity.abnf S10.2: 0x0000 reserved-invalid, 0x0001 = v1 allowlist entry {Ed25519/EdDSA-Protodoc-1, SHA-256}, 0x0002-0xFFFF reserved for future major-version allowlist entries. Provide Validate() that accepts only 0x0001 for the current major version and rejects everything else with a named error (never a warning-level accept).

- **Implements:** CON-015
- **Depends on:** None
- **DoD:** ParamSet type exists with a Validate() error method; 0x0000, 0x0002, and 0xFFFF are rejected, 0x0001 is accepted; constant names cite integrity.abnf S10.2 in a doc comment.
- **Test:** `TestCON_015_ParamSetAllowlist` (unit)
- **Owner:** argus

**T-0103** Deterministic Sign() wrapper over stdlib crypto/ed25519

> Implement Sign(priv ed25519.PrivateKey, msg [32]byte) [64]byte as a thin wrapper over stdlib crypto/ed25519.Sign, taking no additional randomness parameter, per CQ-004 site #3: the only non-deterministic input at signing time is the private key material itself.

- **Implements:** NFR-006
- **Depends on:** None
- **DoD:** Sign() returns the 64-octet R\|\|S RFC 8032 encoding; function signature accepts no seed/nonce/rand.Reader parameter beyond the key.
- **Test:** `TestNFR_006_SignDeterministic` (unit)
- **Owner:** argus

**T-0104** Small-order point constant table with pinned digest

> Populate the fixed 8-entry table of edwards25519 small-order compressed-point encodings verbatim from RFC 8032 section 5.1.3 and Chalkias/Garillot/Nikolaenko 'Taming the many EdDSAs' (2020) Table 1, per integrity.abnf S6's cited requirement. Pin the table's SHA-256 digest as a checked-in constant so an accidental future edit is caught by CI equality check (mirrors the CON-009 ceiling-table pattern).

- **Implements:** CON-015
- **Depends on:** None
- **DoD:** smallOrderPoints [8][32]byte table exists with inline citation comments to both sources; a test computes SHA-256 over the table and asserts equality with the pinned value in testdata/eddsa-protodoc-1/small_order_table.sha256.
- **Test:** `TestCON_015_SmallOrderTableDigestPinned` (unit)
- **Owner:** argus

**T-0105** Steps 1-2: canonical-encoding and small-order check for A

> Implement checkPublicKey(A [32]byte) error: reject if the 255-bit magnitude (top bit of octet 31 masked off) is >= p = 2^255-19 (non-canonical encoding), then reject if A exactly matches any of the 8 small-order table entries.

- **Implements:** CON-015
- **Depends on:** T-0104
- **DoD:** checkPublicKey rejects magnitude==p, magnitude==p+1, and each of the 8 small-order encodings; accepts a valid random-looking canonical key.
- **Test:** `TestCON_015_Step1Step2_PublicKeyChecks` (unit)
- **Owner:** argus

**T-0106** Steps 3-4: canonical-encoding and small-order check for R

> Implement checkR(R [32]byte) error mirroring T-0105's canonical-magnitude and small-order checks, applied to the R half of sig-value.

- **Implements:** CON-015
- **Depends on:** T-0104
- **DoD:** checkR rejects the same non-canonical and small-order encoding classes as checkPublicKey, exercised via its own table-driven test.
- **Test:** `TestCON_015_Step3Step4_RChecks` (unit)
- **Owner:** argus

**T-0107** Step 5: scalar S range check against L

> Implement checkScalarS(S [32]byte) error: decode S as little-endian 256-bit unsigned integer, reject if S >= L = 2^252 + 27742317777372353535851937790883648493 (RFC 8032 exact value).

- **Implements:** CON-015
- **Depends on:** None
- **DoD:** checkScalarS rejects S==L, S==L+1, S==2^256-1, accepts S==L-1 and S==0, using exact big.Int or fixed-width comparison (no float arithmetic).
- **Test:** `TestCON_015_Step5_ScalarRangeCheck` (unit)
- **Owner:** argus

**T-0108** Verify() 7-step orchestration with short-circuit and stdlib delegation

> Compose Steps 1-5 (T-0105/605/606) in order; on the first failure return false immediately without calling stdlib. On all 5 passing, delegate Step 6/7 entirely to unmodified stdlib crypto/ed25519.Verify(A, msg, R\|\|S) and return its boolean result verbatim. Step 6's challenge computation is never independently recomputed outside that call, per integrity.abnf S6's exposition note.

- **Implements:** CON-015
- **Depends on:** T-0103, T-0105, T-0106, T-0107
- **DoD:** Verify(A, msg, sig) returns false for a bit-flipped signature and true for a genuine one; an instrumented stdlib-call counter in the test proves stdlib Verify is never invoked when any of Steps 1-5 already rejected the input.
- **Test:** `TestCON_015_Verify7StepShortCircuit` (unit)
- **Owner:** argus

**T-0109** EdDSA-Protodoc-1 conformance vector corpus

> Author the golden conformance corpus for the full 7-step procedure: a genuinely valid RFC 8032 test vector cross-checked against stdlib; non-canonical A; non-canonical R; each of the 8 small-order encodings applied to A and separately to R (16 cases); S==L and S>L. Each fixture records its expected accept/reject verdict.

- **Implements:** CON-015
- **Depends on:** T-0108
- **DoD:** A corpus-driven test iterates every fixture file under testdata/eddsa-protodoc-1/ and asserts Verify()'s result matches the fixture's recorded expected verdict, with zero mismatches.
- **Test:** `EDDSA_P1_CONFORMANCE_CORPUS_V1` (conformance)
- **Owner:** argus

**T-0110** Sign-twice determinism vector suite

> Sign a fixed set of representative messages (0-octet, 1-octet, 4096-octet) twice each with the same key and assert the raw 64-octet sig-value output is byte-identical both times, then confirm both outputs verify true via T-0108's Verify(). This is the literal NFR-006 acceptance test.

- **Implements:** NFR-006
- **Depends on:** T-0103, T-0108
- **DoD:** All 3 message-size fixtures produce byte-identical sig-value across two independent Sign() calls, and both signatures pass Verify().
- **Test:** `TestNFR_006_SignTwiceIdenticalOctets` (unit)
- **Owner:** argus

**T-0111** Off-allowlist sig-param-set rejection gate

> Implement VerifyWithParamSet(paramSet ParamSet, A, msg, sig) that calls ParamSet.Validate() first and returns rejected (never a positive or warning verdict) for 0x0000 or any 0x0002-0xFFFF value, short-circuiting before Verify() is ever invoked, so an off-allowlist parameter set can never reach the Ed25519 arithmetic.

- **Implements:** CON-015
- **Depends on:** T-0102, T-0108
- **DoD:** A call-counting stub proves Verify() is never invoked when paramSet fails Validate(); table-driven test enumerates 0x0000, 0x0002, and 0xFFFF as rejected and 0x0001 as the only value that proceeds to Verify().
- **Test:** `TestCON_015_OffAllowlistParamSetRejected` (unit)
- **Owner:** argus

**T-0112** argus security review of the EdDSA-Protodoc-1 primitive

> Security review of the complete eddsa package before any downstream Signature/CoverageDescriptor/sign/verify/redact/publish/migrate task is allowed to depend on it: confirm no non-crypto RNG is used anywhere in the signing path, no private key material is logged or returned in error strings, the short-circuit structure introduces no new secret-dependent branch beyond the 5 specified checks, and the small-order table's provenance citations are intact.

- **Implements:** CON-015, NFR-006
- **Depends on:** T-0108, T-0109, T-0110, T-0111
- **DoD:** Review checklist filed with zero open HIGH/CRITICAL findings (any MEDIUM findings tracked with an owner); a static-check test greps the package for math/rand usage and %v-formatted key/error logging and fails if either is found.
- **Test:** `TestCON_015_ArgusSecurityReviewChecklist` (integration)
- **Owner:** argus

---

### M07: Structural Validation Core

The generated ceiling table and the validation pipeline's error-precedence/cycle-detection/referential-integrity checks must exist before integrity, rendering, or migration can be validated against a document.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0113 | Validation pipeline orchestrator with error-precedence ordering | FR-103 | None | hephaestus | `TestFR_103_ErrorPrecedenceReportsEarliestFailureOnly` (unit) |
| T-0114 | Structural-failure diagnostic reporter (offset/unit/rule id) | FR-102 | T-0113 | hephaestus | `TestFR_102_StructuralFailureReportsOffsetUnitRule` (conformance) |
| T-0115 | Generated ceiling table pre-allocation guard | FR-106 | T-0113 | hephaestus | `TestFR_106_CeilingPreAllocationAbortsBeforeAllocation` (conformance) |
| T-0116 | CON-010 at-limit/over-limit fixture authoring for M07-owned ceilings | FR-106 | T-0115 | momus | `TestFR_106_CON010_CeilingFixtureCorpus` (conformance) |
| T-0117 | Referential-integrity check for xref/index/range/identity resolution | FR-108 | T-0113 | hephaestus | `TestFR_108_UnresolvedOrAmbiguousReferenceRejected` (conformance) |
| T-0118 | Reference-graph cycle detection over the 5 named edge kinds | FR-109 | T-0113, T-0115 | hephaestus | `TestFR_109_CycleDetectionNamesAllEdgesBeforeCeilingChecks` (conformance) |
| T-0119 | Extension-envelope fallback-reference cycle vector | FR-109 | T-0118 | momus | `TestFR_109_ExtensionEnvelopeFallbackCycleVector` (conformance) |
| T-0120 | Duplicate-identifier rejection across units and index entries | FR-110 | T-0113 | hephaestus | `TestFR_110_DuplicateIdentifierNamesBothLocations` (conformance) |
| T-0121 | Validator memory-arena design under the adopted NFR-030 bound | NFR-030 | T-0113 | hephaestus | `BenchmarkNFR_030_ValidatorPeakMemoryWithinAdoptedBound` (benchmark) |
| T-0122 | Continuous fuzz target for validator memory-bound and crash safety | NFR-030 | T-0121 | prometheus | `FuzzValidate_NFR030_MemoryBoundAndNoCrash` (fuzz) |
| T-0123 | Verification-before-decode ordering guarantee | NFR-034 | T-0113 | hephaestus | `TestNFR_034_NoDecoderConstructedDuringStructuralValidation` (integration) |
| T-0124 | Admissible-scalar validation (CON-004) | CON-004 | T-0113 | hephaestus | `TestCON_004_ExcludedScalarClassesRejected` (conformance) |
| T-0125 | OVER_BUDGET verdict and exit-code precedence | CON-011 | T-0113 | hephaestus | `TestCON_011_OverBudgetExitCodePrecedence` (integration) |
| T-0126 | Single writer-conformance-class enforcement | CON-017 | T-0113 | clio | `TestCON_017_SingleWriterConformanceClassAsserted` (unit) |
| T-0127 | Reader conformance-role assignment audit (A-ROLES) | CON-018 | T-0113 | clio | `TestCON_018_EveryReaderStatementAssignedToOneOfThreeRoles` (integration) |
| T-0128 | 13-step pipeline ordering golden corpus | FR-102, FR-103, FR-106, FR-108, FR-109, FR-110 | T-0113, T-0114, T-0115, T-0117, T-0118, T-0120 | momus | `TestM07_PipelineStepOrderingGoldenCorpus` (conformance) |
| T-0360 | Document CON-017 enforcement rationale for phase-5 analyze inputs | CON-017 | T-0126 | clio | `TestCON_017_AnalyzeInputNotePresent` (unit) |
| T-0361 | Record NFR-030 memory-floor disclosed-conflict ruling request in clarify.md | NFR-030 | T-0121 | clio | `TestNFR_030_ClarifyRulingRequestRecorded` (unit) |

**T-0113** Validation pipeline orchestrator with error-precedence ordering

> Build the `validate` package's pipeline executor running the 13 ordered steps from data-model.md S7 (depends on M01's Header/CommitRing/SegmentTable structs, M02's place()/segment reads, and M04's run/unit identity being available to reference). Enforce that a structural parse failure yields exactly one verdict — the earliest-ordered failing step — with no partial presentation output and no heuristic reconstruction attempted past that point.

- **Implements:** FR-103
- **Depends on:** None
- **DoD:** Pipeline executes all 13 steps in the documented fixed order; given a fixture violating steps 3 and 8 simultaneously, only step 3's verdict is returned and step 8 is never evaluated; no code path returns a partially-rendered/heuristically-repaired result on structural failure.
- **Test:** `TestFR_103_ErrorPrecedenceReportsEarliestFailureOnly` (unit)
- **Owner:** hephaestus

**T-0114** Structural-failure diagnostic reporter (offset/unit/rule id)

> For truncation, oversized-length-field, and unparseable-TLV failures, populate the pipeline Finding with the octet offset of the first divergence, the enclosing unit id (or an explicit NONE for pre-unit prefix failures), and the violated rule id, in the shape TR-001's report schema expects.

- **Implements:** FR-102
- **Depends on:** T-0113
- **DoD:** Three corpus fixtures (truncated segment, oversized declared length, malformed TLV field) each yield a Finding with the exact expected offset, unit id, and rule id, byte-for-byte matched against golden expectations.
- **Test:** `TestFR_102_StructuralFailureReportsOffsetUnitRule` (conformance)
- **Owner:** hephaestus

**T-0115** Generated ceiling table pre-allocation guard

> Before allocating any buffer sized from a declared count (frame_count, segment_count, run count, etc.), check the declared value against M01's CON-009 generated ceiling constants using overflow-safe arithmetic, aborting and naming the ceiling id plus the observed value before any proportional allocation occurs. The frame_count*48+32<=segment_length-64 boundary vector itself is M01's task per the spine's cross-cutting note; this task covers the remaining ceilings this milestone is responsible for guarding.

- **Implements:** FR-106
- **Depends on:** T-0113
- **DoD:** For every non-M01-owned ceiling in CON-009's table, an at-limit fixture allocates and passes, and a one-past-limit fixture aborts before any large allocation (verified via an allocation-count instrumentation asserting zero oversized allocations on the failing path).
- **Test:** `TestFR_106_CeilingPreAllocationAbortsBeforeAllocation` (conformance)
- **Owner:** hephaestus

**T-0116** CON-010 at-limit/over-limit fixture authoring for M07-owned ceilings

> Author the golden conformance corpus files (per CON-010's 'ship files at each ceiling and one unit beyond it' rule) for every ceiling guarded by T-0115, kept as a separate authoring task from the guard implementation itself so the corpus can be independently reviewed and extended.

- **Implements:** FR-106
- **Depends on:** T-0115
- **DoD:** One at-limit and one over-limit fixture file exists per M07-owned ceiling; both implementations (this repo's validator, run twice) agree on the pass/abort verdict for each fixture.
- **Test:** `TestFR_106_CON010_CeilingFixtureCorpus` (conformance)
- **Owner:** momus

**T-0117** Referential-integrity check for xref/index/range/identity resolution

> Validate that every xref-target, index-entry, annotation range endpoint (run_id + base_ordinal, from M04), and identity reference resolves to exactly one currently-present unit. Unresolvable or multiply-resolving references are rejected outright — never silently dropped, defaulted, or auto-repaired.

- **Implements:** FR-108
- **Depends on:** T-0113
- **DoD:** Fixtures for a dangling xref, a dangling annotation anchor, and a reference resolving to two units are all rejected citing FR-108; a fully-resolved valid document passes this step.
- **Test:** `TestFR_108_UnresolvedOrAmbiguousReferenceRejected` (conformance)
- **Owner:** hephaestus

**T-0118** Reference-graph cycle detection over the 5 named edge kinds

> Build graph construction over data-model.md S7 step 8's 5 named edge kinds and run cycle detection strictly before the ceiling-check step in pipeline order. On the first cycle found, name every edge participating in it. Note the frozen, self-disclosed gap (contracts/README.md / data-model.md S7): a CROSS_REFERENCE edge is NOT currently one of the 5 kinds even though spec.md's FR-109 text names cross-references as one of its 4 abstract graph categories — this task implements the 5-kind design as currently approved and records the gap in a code comment rather than unilaterally expanding the list.

- **Implements:** FR-109
- **Depends on:** T-0113, T-0115
- **DoD:** A single-edge self-cycle and a multi-edge cycle each produce a rejection naming every edge in the cycle; an ordering test asserts the cycle-detection step runs and rejects before any ceiling-check counter is incremented.
- **Test:** `TestFR_109_CycleDetectionNamesAllEdgesBeforeCeilingChecks` (conformance)
- **Owner:** hephaestus

**T-0119** Extension-envelope fallback-reference cycle vector

> Author the conformance vector for an extension-envelope fallback-reference (ext-fallback-ref) forming a cycle — one of FR-109's 5 edge kinds — per plan.md Section 10 first-class task (g), which explicitly requires this NOT be folded into general conformance work. Note: a possibly-overlapping fixture is referenced elsewhere under raw id T-0064 in another milestone's inventory; this task is kept as specified pending phase-5 analyze reconciliation (see milestone notes) rather than unilaterally merged or dropped.

- **Implements:** FR-109
- **Depends on:** T-0118
- **DoD:** The fixture is committed to the golden corpus and fails against T-0118's detector prior to this task's edge-kind wiring, and passes (rejects, naming all edges) after.
- **Test:** `TestFR_109_ExtensionEnvelopeFallbackCycleVector` (conformance)
- **Owner:** momus

**T-0120** Duplicate-identifier rejection across units and index entries

> Detect any two stored units or index entries resolving to the same identifier and reject, naming both physical locations (segment ordinal + intra-segment offset for each); never silently rename either or apply precedence between them.

- **Implements:** FR-110
- **Depends on:** T-0113
- **DoD:** A fixture with two RUN records sharing a run_id and a fixture with two SegmentTableSlot entries sharing a unit id both reject naming both locations; a single-instance document is unaffected.
- **Test:** `TestFR_110_DuplicateIdentifierNamesBothLocations` (conformance)
- **Owner:** hephaestus

**T-0121** Validator memory-arena design under the adopted NFR-030 bound

> Implement the validator's memory arena so peak resident memory for any input (valid or malformed) stays within the adopted floor-exempted, heap-for-document-data reading disclosed in plan.md Section 9 Conflict 3, since NFR-030's literal 'peak memory <= 4x input octet length' with zero floor is unsatisfiable by any real process on small inputs.

- **Implements:** NFR-030
- **Depends on:** T-0113
- **DoD:** Benchmark harness measures peak RSS across input sizes 1 KB / 1 MB / 1 GiB and malformed variants of each; all measurements fall within the adopted floor+4x bound; the divergence from NFR-030's literal text is recorded in a code comment citing plan.md Section 9 Conflict 3.
- **Test:** `BenchmarkNFR_030_ValidatorPeakMemoryWithinAdoptedBound` (benchmark)
- **Owner:** hephaestus

**T-0122** Continuous fuzz target for validator memory-bound and crash safety

> Wire a Go native fuzz target over arbitrary byte inputs into the validator entrypoint, asserting peak memory never exceeds T-0121's bound and the process never panics, feeding CP-012's continuous fuzzing harness (checked at M19).

- **Implements:** NFR-030
- **Depends on:** T-0121
- **DoD:** `FuzzValidate` runs for a minimum 5-minute local CI budget with zero bound violations and zero crashes; the fuzz corpus is committed as a seed corpus for CI reuse.
- **Test:** `FuzzValidate_NFR030_MemoryBoundAndNoCrash` (fuzz)
- **Owner:** prometheus

**T-0123** Verification-before-decode ordering guarantee

> Enforce in the pipeline orchestrator that structural validation and integrity/signature verification steps (T_C/T_S/signature checks per data-model.md S7) complete before any font/image/audio/video decoder is constructed, per CP-006. Add an architectural check that the `validate` pipeline's source order never constructs a codec object ahead of its verification steps.

- **Implements:** NFR-034
- **Depends on:** T-0113
- **DoD:** A fixture with a malformed embedded raster image but otherwise valid structure completes validation without the image decoder ever being invoked (verified via a decode-call counter mock asserting zero invocations); a CI lint rule fails the build if a codec package is imported before the verification step in source order.
- **Test:** `TestNFR_034_NoDecoderConstructedDuringStructuralValidation` (integration)
- **Owner:** hephaestus

**T-0124** Admissible-scalar validation (CON-004)

> Implement the scalar-admissibility check that has zero mechanism anywhere in plan.md/data-model.md/contracts (a confirmed spec-to-implementation gap, not an ambiguity): reject any text scalar unassigned in the bound Unicode version (NFR-020's unicode-version-id) or falling into the excluded classes named by spec.md CON-004 (bidi-control, deprecated, noncharacter, surrogate, private-use, most control characters). Because no downstream artifact records this mechanism, implement directly against spec.md's CON-004 text and the bound UCD, and flag to Eyvar/themis in a code comment plus a phase-5 analysis note that data-model.md and contracts/README.md need amendment to record it (the requirement, spec.md's own PD-SCALAR-001 rule id, and CON-004 itself are otherwise never cited downstream).

- **Implements:** CON-004
- **Depends on:** T-0113
- **DoD:** Conformance fixtures covering one scalar from each excluded class (surrogate, noncharacter, private-use, a deprecated scalar, a disallowed control) are each rejected naming the offending scalar and its position; an all-admissible fixture passes; the gap-and-flag comment is present in the source.
- **Test:** `TestCON_004_ExcludedScalarClassesRejected` (conformance)
- **Owner:** hephaestus

**T-0125** OVER_BUDGET verdict and exit-code precedence

> Introduce a reader resource-budget-refusal status (OVER_BUDGET) distinct from every validity verdict, wired into cli.md's exit-code precedence: ranked above UNAVAILABLE/UNVERIFIED/REFUSED but below INVALID.

- **Implements:** CON-011
- **Depends on:** T-0113
- **DoD:** A fixture that is both structurally INVALID and over-budget reports INVALID; a structurally-valid but over-budget fixture reports OVER_BUDGET; an UNVERIFIED-and-over-budget fixture reports OVER_BUDGET — both precedence directions verified in one table-driven test.
- **Test:** `TestCON_011_OverBudgetExitCodePrecedence` (integration)
- **Owner:** hephaestus

**T-0126** Single writer-conformance-class enforcement

> Confirm and enforce that exactly one writer conformance class exists: any document produced by any conforming writer must satisfy the identical validation pipeline, with no separate lenient/strict writer-class flag or branch anywhere in `validate`. This is satisfied by the absence of a second class, not by new logic — add a CI grep/lint guard against introducing one.

- **Implements:** CON-017
- **Depends on:** T-0113
- **DoD:** CI lint confirms zero conditional branches keyed on any 'writer class' concept anywhere in the `validate` package.
- **Test:** `TestCON_017_SingleWriterConformanceClassAsserted` (unit)
- **Owner:** clio

**T-0127** Reader conformance-role assignment audit (A-ROLES)

> Build the A-ROLES audit tooling that walks every normative reader-binding requirement (any FR-*/NFR-* whose plan_mechanism cites `extract`, `validate`, or `render`) and asserts it is assigned to at least one of exactly 3 defined reader conformance roles (extracting, validating-and-verifying, rendering); fail the audit on zero-assignment or on any role outside this closed set of 3.

- **Implements:** CON-018
- **Depends on:** T-0113
- **DoD:** The audit script runs over the requirement-to-package traceability data (this milestone's inventory plus sibling milestones') and produces a report with zero unassigned-role and zero outside-closed-set findings; wired into CI as part of the M07 exit gate.
- **Test:** `TestCON_018_EveryReaderStatementAssignedToOneOfThreeRoles` (integration)
- **Owner:** clio

**T-0128** 13-step pipeline ordering golden corpus

> Ship a golden corpus exercising the full 13-step pipeline order end-to-end: one fixture failing at each individual step in isolation, plus one 'everything wrong at once' fixture asserting only the earliest-ordered eligible failure is reported (FR-103), satisfying CP-011's ship-corpus-with-spec-text requirement for this milestone's cross-cutting concern.

- **Implements:** FR-102, FR-103, FR-106, FR-108, FR-109, FR-110
- **Depends on:** T-0113, T-0114, T-0115, T-0117, T-0118, T-0120
- **DoD:** 14 corpus fixtures (13 individual + 1 combined) are committed; CI runs all of them against the pipeline and asserts each produces exactly its expected step's verdict and no other.
- **Test:** `TestM07_PipelineStepOrderingGoldenCorpus` (conformance)
- **Owner:** momus

**T-0360** Document CON-017 enforcement rationale for phase-5 analyze inputs

> Split out of T-0126's definition of done: that task's CI lint guard is independently CI-checkable, but the accompanying 'note added to phase-5 analyze inputs' clause is a separate documentation deliverable that cannot be verified by the same test. This task adds the phase-5 analyze input note recording that CON-017 (single writer conformance class) is satisfied by the absence of a second class, evidenced by T-0126's passing lint guard.

- **Implements:** CON-017
- **Depends on:** T-0126
- **DoD:** A note file is committed under the phase-5 analyze inputs path stating CON-017 is satisfied by absence of a second writer-conformance class and citing T-0126's CI lint result; zeus's analyze phase can consume it without further clarification.
- **Test:** `TestCON_017_AnalyzeInputNotePresent` (unit)
- **Owner:** clio

**T-0361** Record NFR-030 memory-floor disclosed-conflict ruling request in clarify.md

> M07's NFR-030 memory-floor reading (the adopted floor-exempted, heap-for-document-data bound from plan.md Section 9 Conflict 3) gates this milestone's close, but unlike the analogous disclosed-conflict governance pattern used elsewhere in this spec (raw ids T-1106/T-1412/T-1501/T-1806 in sibling milestones M11/M14/M15/M19), no task in M07 records an open ruling request for it in clarify.md. Add the matching clarify.md entry so governance tracking of this deviation is consistent across milestones.

- **Implements:** NFR-030
- **Depends on:** T-0121
- **DoD:** clarify.md contains a new open-ruling-request entry for M07's NFR-030 floor-exempted reading, matching the format of the sibling entries, dated and cross-referencing plan.md Section 9 Conflict 3 and T-0121.
- **Test:** `TestNFR_030_ClarifyRulingRequestRecorded` (unit)
- **Owner:** clio

---

### M08: Integrity Trees (T_S/T_C)

structure_digest and state identity both derive from T_C_root, so the digest trees must be correct before any Signature work can be built on top of them.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0129 | ABSENT_CHILD_DIGEST constant and T_S/T_C domain-tag registry | FR-003, TR-009 | None | argus | `TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed` (unit) |
| T-0130 | T_S tree: leaf and internal node encoding, arity-16/depth-4 builder | TR-009 | T-0129 | argus | `TestTR_009_TSTreeFixedShapeAndAbsentChildFill` (unit) |
| T-0131 | T_S_root end-to-end computation with never-trust-stored recomputation rule | TR-009 | T-0130 | argus | `TestTR_009_TSRootAlwaysRecomputedNeverCached` (unit) |
| T-0132 | T_C leaf-redactable node encoding (domain tag 0x02) | FR-003 | T-0129 | argus | `TestFR_003_TCLeafRedactableUsesStoredFrameVerbatim` (unit) |
| T-0133 | T_C leaf-nonredactable node encoding (domain tag 0x07) | FR-003 | T-0129 | argus | `TestFR_003_TCLeafNonredactableDomainSeparatedFromRedactable` (unit) |
| T-0134 | T_C subtree traversal order (interim ascending unit-id byte-lexicographic rule) | FR-003 | None | argus | `TestFR_003_TCTraversalOrderDeterministicUnitIDLexicographic` (unit) |
| T-0135 | T_C internal node encoding, arity-16/depth<=5 builder | FR-003 | T-0129, T-0132, T-0133, T-0134 | argus | `TestFR_003_TCTreeFixedShapeAndAbsentChildFill` (unit) |
| T-0136 | T_C_root end-to-end computation (state identity) | FR-001, FR-003 | T-0135 | argus | `TestFR_001_TCRootIsCompleteStateValueSet` (unit) |
| T-0137 | Wire CommitRingRecord.state-id-field = T_C_root at commit time | FR-001, FR-003 | T-0136, T-0140 | argus | `TestFR_003_CommitRingStateIDMatchesTCRoot` (integration) |
| T-0138 | ATTEST-segment exclusion predicate (shared by T_C, CoverageDescriptor, and no-op-save comparison) | NFR-007, FR-002 | T-0129 | argus | `TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare` (unit) |
| T-0139 | CoverageDescriptor wire encode/decode and PD-COVER-001..004 structural validity (canonical shared implementation) | FR-002 | T-0138 | argus | `TestFR_002_CoverageDescriptorCanonicalImplementationRejectsEachPDCoverViolation` (unit) |
| T-0140 | structure_digest fresh-recomputation algorithm (8-item conditional preimage) | FR-002, TR-009 | T-0131, T-0139 | argus | `TestFR_002_StructureDigestPreimageMatchesBitmaskExactly` (unit) |
| T-0141 | Conformance test: altered stored inventory detected before being relied upon | TR-009 | T-0131, T-0140 | momus | `TestTR_009_TamperedSegmentTableSlotDetectedBeforeReliance` (conformance) |
| T-0142 | Conformance test: two-run determinism of T_S_root and T_C_root | FR-003 | T-0131, T-0136 | momus | `TestFR_003_TSAndTCRootsDeterministicAcrossTwoRuns` (conformance) |
| T-0143 | Conformance test: root changes whenever any covered value changes | FR-003 | T-0131, T-0136, T-0140 | momus | `TestFR_003_RootChangesForEveryCoveredValueMutation` (conformance) |
| T-0144 | At-limit and one-past-limit conformance fixtures for T_S/T_C capacity | FR-003, TR-009 | T-0131, T-0136 | momus | `TestTR_009_TSTreeAtLimitAndOnePastLimit` (conformance) |
| T-0362 | Milestone-exit argus security review of T_S/T_C integrity trees and structure_digest | FR-001, FR-002, FR-003, NFR-007, TR-009 | T-0136, T-0137, T-0138, T-0139, T-0140, T-0141, T-0142, T-0143, T-0144 | argus | `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` (conformance) |

**T-0129** ABSENT_CHILD_DIGEST constant and T_S/T_C domain-tag registry

> Checked-in Go constants for ABSENT_CHILD_DIGEST = SHA-256(0x00) and the domain-tag registry values used by the two trees: 0x01 (T_S internal), 0x02 (T_C leaf-redactable), 0x07 (T_C leaf-nonredactable), 0x08 (T_C internal) per contracts/README.md S3.2's consolidated registry and integrity.abnf S2. A single shared table prevents two implementations from disagreeing on which byte prefixes a real digest can never start with (required so ABSENT_CHILD_DIGEST can never collide with a genuine node/leaf digest, per S2.3's own claim). This is pure constant/utility code with no tree-shape logic yet -- feeds T-0130 and T-0132..T-0136.

- **Implements:** FR-003, TR-009
- **Depends on:** None
- **DoD:** ABSENT_CHILD_DIGEST computed as SHA-256 of the single octet 0x00 and asserted against a hard-coded expected hex value; a table-driven test confirms all 8 registry values (0x00 plus the 7 real preimage-start octets named in the file, 0x01/0x02/0x04/0x07/0x08/0x09/0x0A) are pairwise distinct.
- **Test:** `TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed` (unit)
- **Owner:** argus

**T-0130** T_S tree: leaf and internal node encoding, arity-16/depth-4 builder

> Implement t-s-leaf (= the SegmentTableSlot.slot-digest for the segment at that storage-order position, no domain tag of its own) and t-s-internal (0x01 \|\| 16 child digest256, ALWAYS a full 513-octet preimage regardless of how many children are real) per integrity.abnf S2.1. Build the fixed arity-16, depth-4 tree (16^4 = 65536 leaf capacity) over a live []SegmentTableSlot from M01's container package, filling any tree position with no corresponding slot with ABSENT_CHILD_DIGEST from T-0129, at every level, not only the root.

- **Implements:** TR-009
- **Depends on:** T-0129
- **DoD:** Given N populated SegmentTableSlots (0 <= N <= 16384), the builder produces a tree of exactly depth 4 with every internal-node preimage exactly 513 octets, and every unpopulated child slot at every level equals ABSENT_CHILD_DIGEST.
- **Test:** `TestTR_009_TSTreeFixedShapeAndAbsentChildFill` (unit)
- **Owner:** argus

**T-0131** T_S_root end-to-end computation with never-trust-stored recomputation rule

> Wire T-0130's builder into a single TSRoot(slots []SegmentTableSlot) [32]byte entry point that always recomputes from the SegmentTable's current live octets. Per data-model.md 2.11 note 2 and integrity.abnf S2.1, T_S's root is NEVER read from CommitRingRecord.ledger_root as authoritative -- ledger_root is only compared against a freshly recomputed value, and divergence is a distinct verdict, never silently accepted.

- **Implements:** TR-009
- **Depends on:** T-0130
- **DoD:** TSRoot() takes no cached-root parameter and has no code path that returns a stored value without recomputation; a unit test tampers with one SegmentTableSlot's stored slot-digest in-memory and confirms TSRoot() reflects the tampered value rather than any cached root.
- **Test:** `TestTR_009_TSRootAlwaysRecomputedNeverCached` (unit)
- **Owner:** argus

**T-0132** T_C leaf-redactable node encoding (domain tag 0x02)

> Implement t-c-leaf-redactable = 0x02 \|\| tc-salt \|\| tc-canon per integrity.abnf S2.2. tc-salt is a 32-octet CSPRNG value minted once per designated redactable subtree at signing time (crypto/rand, consistent with M04's run_id minting discipline); tc-canon is the exact stored PDL-TLV frame bytes of the one content-model record this subtree ordinal names (document.abnf S0's `record` grammar: discriminant + fields at strictly ascending tags), never a re-derived or re-normalised copy.

- **Implements:** FR-003
- **Depends on:** T-0129
- **DoD:** Given a stored content-model record's raw frame bytes and a 32-octet salt, the function returns SHA-256(0x02\|\|salt\|\|frame) with frame taken byte-for-byte from storage (no re-encode step); changing one octet of the stored frame changes the leaf digest.
- **Test:** `TestFR_003_TCLeafRedactableUsesStoredFrameVerbatim` (unit)
- **Owner:** argus

**T-0133** T_C leaf-nonredactable node encoding (domain tag 0x07)

> Implement t-c-leaf-nonredactable = 0x07 \|\| tc-canon per integrity.abnf S2.2, for content-model records not designated redactable at signing time. Same tc-canon exact-frame-bytes rule as T-0132, without a salt.

- **Implements:** FR-003
- **Depends on:** T-0129
- **DoD:** Given a stored content-model record's raw frame bytes, the function returns SHA-256(0x07\|\|frame); a table-driven test confirms this never collides with the 0x02-tagged leaf digest of the identical frame bytes plus any possible salt (domain separation holds).
- **Test:** `TestFR_003_TCLeafNonredactableDomainSeparatedFromRedactable` (unit)
- **Owner:** argus

**T-0134** T_C subtree traversal order (interim ascending unit-id byte-lexicographic rule)

> Implement the subtree ordinal assignment that T-0135/T-0136 traverse in: ascending unsigned byte-lexicographic order of each content-model record's own unit-id (document.abnf S0), per integrity.abnf S2.2.1's INTERIM RULE. This is a FLAGGED, self-disclosed gap in the frozen contracts: the interim rule is deterministic and satisfies the negative constraint (never storage ordinal, never a digest) but explicitly does NOT satisfy the aspirational 'reflects the document's logical/reading-order structure' requirement -- a CSPRNG-derived unit-id order carries no relationship to reading order. Do not silently 'fix' this by inventing a reading-order traversal; implement exactly the interim rule as written and carry the same code-comment flag forward, since a future ROOT_SEQUENCE-based redefinition (integrity.abnf S2.2.1's own recommendation) is additive future work, not this task's scope.

- **Implements:** FR-003
- **Depends on:** None
- **DoD:** Given an unordered set of content-model records (each carrying a 16-octet unit-id from M04), the function returns a deterministic ordinal assignment identical to sorting unit-ids as unsigned big-endian byte strings ascending; two independent invocations on a shuffled input slice produce identical ordinal assignments; code comment explicitly cites the FLAGGED interim-rule status.
- **Test:** `TestFR_003_TCTraversalOrderDeterministicUnitIDLexicographic` (unit)
- **Owner:** argus

**T-0135** T_C internal node encoding, arity-16/depth<=5 builder

> Implement t-c-internal = 0x08 \|\| 16 child digest256 (always a full 513-octet preimage) and assemble the complete T_C tree (arity 16, depth <= 5) from T-0134's subtree ordinal order and T-0132/T-0133's leaf digests, filling any position with no corresponding subtree with ABSENT_CHILD_DIGEST at every level per integrity.abnf S2.2/data-model.md 2.10 note 1.

- **Implements:** FR-003
- **Depends on:** T-0129, T-0132, T-0133, T-0134
- **DoD:** Given a mixed set of redactable and nonredactable content-model records, the builder produces a tree whose depth never exceeds 5, every internal-node preimage is exactly 513 octets, and absent children at every level equal ABSENT_CHILD_DIGEST.
- **Test:** `TestFR_003_TCTreeFixedShapeAndAbsentChildFill` (unit)
- **Owner:** argus

**T-0136** T_C_root end-to-end computation (state identity)

> Wire T-0135's builder into a single TCRoot(records []ContentRecord) [32]byte entry point defining state identity per DP-006/FR-003: 'every state gets an identifier differing whenever any value of that state differs.' Per data-model.md 2.10 note 3/5, T_C_root is the node at subtree ordinal 0 in the traversal order (T-0134) and is signed directly inside signed_object, never routed through T_S. This is also the concrete realization of FR-001's abstract 'document state = complete value set' definition: TCRoot's input set IS the value set that determines extraction/render/verdict/metadata (excluding ATTEST-typed segments, enforced by T-0138).

- **Implements:** FR-001, FR-003
- **Depends on:** T-0135
- **DoD:** TCRoot() over a fixed record set is reproducible; changing any one record's stored frame bytes, adding a record, or removing a record all change TCRoot()'s return value; TCRoot() never reads or depends on any SegmentTableSlot or T_S value.
- **Test:** `TestFR_001_TCRootIsCompleteStateValueSet` (unit)
- **Owner:** argus

**T-0137** Wire CommitRingRecord.state-id-field = T_C_root at commit time

> Close the loop from the tree computation (T-0136) to the persisted, addressable state identity: at every commit, M01's CommitRingRecord.state-id-field (container.abnf S3) must be set to the freshly computed T_C_root of the state being committed, never a stale or separately-derived value. This is the concrete mechanism cited by plan.md for both FR-001 and FR-003. Added dependency on structure_digest (T-0140 / T-0140): the commit-time wiring must use the same freshly-recomputed T_C_root and structure_digest values that feed the signed_object preimage, never a value cached from an earlier stage.

- **Implements:** FR-001, FR-003
- **Depends on:** T-0136, T-0140
- **DoD:** An integration test commits two states differing in exactly one content record and asserts the two resulting CommitRingRecord.state-id-field values differ and each equals that state's independently-computed T_C_root.
- **Test:** `TestFR_003_CommitRingStateIDMatchesTCRoot` (integration)
- **Owner:** argus

**T-0138** ATTEST-segment exclusion predicate (shared by T_C, CoverageDescriptor, and no-op-save comparison)

> Implement one shared predicate IsAttestTyped(slot SegmentTableSlot) bool (true iff slot-segment-type = ATTEST/4) and use it to exclude every ATTEST-typed segment (SIGNATURE, RESCIND_RESIGN, ATTESTATION_EVIDENCE frames) from three call sites: (1) T_C's subtree set (T-0135/T-0136 never traverse into ATTEST-segment content), (2) the no-op-save octet-identity comparator (NFR-003/NFR-007: signature/time-attestation/revocation octets never participate in 'did anything change'), and (3) CoverageDescriptor construction (T-0139). A single predicate, not three separately-maintained checks, is required so the three guarantees cannot silently drift apart.

- **Implements:** NFR-007, FR-002
- **Depends on:** T-0129
- **DoD:** A table-driven test over all 4 SegmentTableSlot.slot-segment-type values (CONTENT/RESOURCE/HISTORY/ATTEST) confirms IsAttestTyped is true only for ATTEST, and the three call sites listed above are proven (by grep/call-graph assertion in the test, or by direct invocation) to route through this single function rather than a local re-implementation.
- **Test:** `TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare` (unit)
- **Owner:** argus

**T-0139** CoverageDescriptor wire encode/decode and PD-COVER-001..004 structural validity (canonical shared implementation)

> Implement coverage-descriptor = cd-mode cd-covered-ranges cd-uncovered-ranges cd-bitmask encode/decode (integrity.abnf S4) and its four structural validity rules: PD-COVER-001 (sr-end > sr-start, reject zero-length ranges), PD-COVER-002 (mandatory merge-adjacent canonicalisation -- exactly one valid encoding per coverage set), PD-COVER-003 (every ordinal in [0, segment_count) covered by exactly one of the two lists, no gap, no overlap), PD-COVER-004 (no ATTEST-typed ordinal nameable in either list, using T-0138's predicate -- this is FR-002's self-coverage circularity closure). Scope is structural encode/decode/validate only; signature binding and the range-list property-based fuzzer at the MAX_SEGMENTS boundary are M09's first-class task, not this one. This package is the SINGLE canonical CoverageDescriptor wire implementation for the whole codebase: FR-063 (M09 signature/coverage call sites) names the identical wire mechanism and MUST import and reuse this package rather than re-implementing coverage-descriptor encode/decode/validate; a second independent implementation of this ABNF production is a defect, not an acceptable split.

- **Implements:** FR-002
- **Depends on:** T-0138
- **DoD:** A TOTAL-mode descriptor covering every non-ATTEST ordinal and no others validates; a SUBSET-mode descriptor with an unmerged-but-technically-sorted adjacent pair, a zero-length range, a gap, an overlap, or an ATTEST ordinal named in either list each fail validation with a distinct, named reason; the package is exported from a stable path so M09's FR-063 call sites can import it directly, and a code comment on the package doc marks it as the canonical FR-002/FR-063 shared implementation.
- **Test:** `TestFR_002_CoverageDescriptorCanonicalImplementationRejectsEachPDCoverViolation` (unit)
- **Owner:** argus

**T-0140** structure_digest fresh-recomputation algorithm (8-item conditional preimage)

> Implement the exact fixed-order preimage assembly of integrity.abnf S3.1: (1) domain tag 0x0A always; (2) header_bytes[0,480) iff bit HEADER; (3) winning CommitRingRecord's own octets[0,480) iff bit RING_WINNER; (4) ledger-length always, unconditionally; (5) frontmatter document_metadata field bytes in fixed tag order iff bit FRONTMATTER (fm-preview-raster/fm-preview-digest/fm-source-snapshot NEVER participate regardless of bit); (6) covered_prefix_regions_bitmask itself, always; (7) sorted covered-segment-summary vec (ordinal, type, length, digest -- digest RECOMPUTED FRESH per segment, never read from stored slot-digest or trailing self-digest) iff bit SEGMENT_TABLE; (8) T_S root (T-0131) RECOMPUTED FRESH iff bit INTEGRITY_BLOCK. Hash the assembled buffer once with SHA-256. This is the concrete mechanism FR-002 and TR-009 both cite: the acted-upon set is exactly this preimage's covered items, and every value in it is freshly recomputed, never trusted from storage.

- **Implements:** FR-002, TR-009
- **Depends on:** T-0131, T-0139
- **DoD:** For each of the 32 bitmask combinations of the 5 real bits (HEADER/RING_WINNER/FRONTMATTER/SEGMENT_TABLE/INTEGRITY_BLOCK), the assembled preimage contains exactly the items that bit combination specifies and omits the rest, with items 1/4/6 always present; a reserved bit (5-7) set causes rejection before assembly begins.
- **Test:** `TestFR_002_StructureDigestPreimageMatchesBitmaskExactly` (unit)
- **Owner:** argus

**T-0141** Conformance test: altered stored inventory detected before being relied upon

> Dedicated conformance test for TR-009's literal guarantee: 'bind the prefix inventory to integrity protection so an altered inventory is detected before any listed value is relied upon.' Construct a document state, compute structure_digest and record it as the trusted baseline, then tamper with one stored SegmentTableSlot field (length or type) without updating any digest, and assert that recomputing structure_digest (via T-0140, which recomputes T_S fresh via T-0131 and segment digests fresh) produces a different value than the trusted baseline -- i.e. the tamper is caught by comparison, never by trusting the tampered slot's own claimed digest.

- **Implements:** TR-009
- **Depends on:** T-0131, T-0140
- **DoD:** A conformance fixture pair (baseline octets, tampered octets differing in exactly one SegmentTableSlot field) ships under testdata/conformance/integrity/, and the test asserts structure_digest(baseline) != structure_digest(tampered) for every bitmask that includes SEGMENT_TABLE or INTEGRITY_BLOCK.
- **Test:** `TestTR_009_TamperedSegmentTableSlotDetectedBeforeReliance` (conformance)
- **Owner:** momus

**T-0142** Conformance test: two-run determinism of T_S_root and T_C_root

> Exit-criteria-mandated determinism check: T_S and T_C must recompute identically across two runs. Run TSRoot() and TCRoot() twice on the identical fixed input (same SegmentTableSlots / same content records, freshly constructed in a second process invocation, not just a repeated in-process call) and assert byte-identical results, closing NFR-001's canonical-octet-determinism concern as it applies specifically to the two integrity trees.

- **Implements:** FR-003
- **Depends on:** T-0131, T-0136
- **DoD:** A golden fixture's TSRoot and TCRoot values are checked into testdata/conformance/integrity/determinism/expected.json; the test recomputes both from the fixture's raw inputs via a fresh process invocation (go test -run, separate subprocess or cache-cleared call) and asserts an exact match against the checked-in values.
- **Test:** `TestFR_003_TSAndTCRootsDeterministicAcrossTwoRuns` (conformance)
- **Owner:** momus

**T-0143** Conformance test: root changes whenever any covered value changes

> Exit-criteria-mandated sensitivity check, complementing T-0142's determinism check. For a base fixture, generate one mutated variant per leaf/domain-tag kind covered by the two trees (T_S leaf digest, T_C redactable leaf frame byte, T_C redactable leaf salt, T_C nonredactable leaf frame byte, one internal-node position becoming ABSENT_CHILD_DIGEST) and assert the corresponding root changes for every variant -- a single unmutated control case must NOT change, ruling out a test that trivially always reports 'changed.' Added dependency on structure_digest (T-0140 / T-0140): the sensitivity check verifies the same fresh T_C_root/structure_digest values Verify() recomputes, not a cached copy.

- **Implements:** FR-003
- **Depends on:** T-0131, T-0136, T-0140
- **DoD:** At least 5 mutation cases (one per kind listed above) each produce a root digest differing from the base fixture's root, and a 6th identical-copy control case produces the identical root, all checked into one table-driven test.
- **Test:** `TestFR_003_RootChangesForEveryCoveredValueMutation` (conformance)
- **Owner:** momus

**T-0144** At-limit and one-past-limit conformance fixtures for T_S/T_C capacity

> Per the CON-010 cross-cutting rule (wire-format conformance vectors ship at-limit and one-past-limit alongside the spec text, not as a follow-up) applied to this milestone: ship a fixture with exactly MAX_SEGMENTS = 16384 populated T_S leaves (still within the 16^4 = 65536 depth-4 capacity, exercising the full headroom claim) and confirm TSRoot computes without error at exactly that count; ship a one-past-limit fixture (16385 populated slots) and confirm the builder reports a named capacity error rather than silently truncating or producing a wrong-shape tree. Structural ceiling REJECTION semantics (aborting before allocation) belong to M07's validator; this task only proves the tree builder itself behaves correctly at and just past the boundary.

- **Implements:** FR-003, TR-009
- **Depends on:** T-0131, T-0136
- **DoD:** TSRoot() over the 16384-slot fixture returns a valid depth-4 root with no error; TSRoot() over the 16385-slot fixture returns a named, distinguishable capacity error (not a panic, not silent truncation).
- **Test:** `TestTR_009_TSTreeAtLimitAndOnePastLimit` (conformance)
- **Owner:** momus

**T-0362** Milestone-exit argus security review of T_S/T_C integrity trees and structure_digest

> Consolidated milestone-exit argus review, matching the pattern already used elsewhere (e.g. M06/M10/M11/M16's consolidated review tasks) but currently absent for this integrity-tree milestone. Covers the whole M08 surface as one gate: domain-tag/ABSENT_CHILD_DIGEST collision-freedom (T-0129), never-trust-stored recomputation for both T_S and T_C (T-0131, T-0136), ATTEST-exclusion single-predicate discipline across all three call sites (T-0138), CoverageDescriptor's four PD-COVER structural rules including the FR-002 self-coverage circularity closure (T-0139), and the structure_digest 8-item conditional preimage's exact bitmask semantics (T-0140), against CP-checklist items and the OWASP ASVS/ISO 27034 controls this milestone touches (domain separation, no-trust-stored-digest, least-exposure of redactable salts).

- **Implements:** FR-001, FR-002, FR-003, NFR-007, TR-009
- **Depends on:** T-0136, T-0137, T-0138, T-0139, T-0140, T-0141, T-0142, T-0143, T-0144
- **DoD:** A written review sign-off exists covering every task T-0129..T-0144 by ID, with zero unresolved P0/P1 findings before this milestone's code can be merged; any P0/P1 found is filed as a new task rather than fixed silently inside the review artifact.
- **Test:** `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` (conformance)
- **Owner:** argus

---

### M09: Signature & Coverage

CoverageDescriptor, signed_object, and the presentation-pinning/verdict machinery are the direct consumers of the crypto primitive and the integrity trees.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0145 | CoverageDescriptor wire struct encode/decode | FR-063 | T-0139 | hephaestus | `TestFR_063_CoverageDescriptorRoundTrip` (unit) |
| T-0146 | PD-COVER-001: reject zero-length segment ranges | FR-063 | T-0145 | argus | `TestFR_063_PDCover001RejectsZeroLengthRange` (conformance) |
| T-0147 | PD-COVER-002: mandatory merge-adjacent canonicalisation | FR-063 | T-0145 | argus | `TestFR_063_PDCover002RejectsUnmergedAdjacentRanges` (conformance) |
| T-0148 | PD-COVER-003: no-gap, no-overlap, full-domain well-formedness check | FR-063 | T-0145 | argus | `TestFR_063_PDCover003RejectsGapOrOverlap` (conformance) |
| T-0149 | PD-COVER-004: ATTEST ordinals never nameable in any CoverageDescriptor | FR-063 | T-0145 | argus | `TestFR_063_PDCover004RejectsAttestOrdinal` (conformance) |
| T-0150 | Property-based fuzzer for CoverageDescriptor range-list canonicalisation | FR-063 | T-0145, T-0146, T-0147, T-0148, T-0149 | prometheus | `FuzzFR_063_CoverageDescriptorCanonicalisation` (fuzz) |
| T-0151 | coverage-descriptor-digest computation | FR-063 | T-0145 | hephaestus | `TestFR_063_CoverageDescriptorDigestDeterministic` (unit) |
| T-0152 | signed_object preimage assembly and computation | FR-063 | T-0136, T-0151 | argus | `TestFR_063_SignedObjectPreimageAssembly` (unit) |
| T-0153 | SIGNATURE record wire struct encode/decode | FR-063, FR-064 | T-0151, T-0152 | hephaestus | `TestFR_063_SignatureRecordRoundTrip` (unit) |
| T-0154 | PresentationArtefact minimal record (profile version, page geometry, font set) | FR-064 | None | hephaestus | `TestFR_064_PresentationArtefactRoundTrip` (unit) |
| T-0155 | sig-presentation-ref binding and presentation-artefact-digest sourcing | FR-064 | T-0152, T-0154 | argus | `TestFR_064_SignedObjectSourcesPresentationDigestFromReferencedSlot` (unit) |
| T-0156 | Presentation mismatch detection reports unattested | FR-065 | T-0154, T-0155 | argus | `TestFR_065_MismatchedPresentationReportsUnattested` (integration) |
| T-0157 | Verdict enum type definition | FR-050 | None | hephaestus | `TestFR_050_VerdictIsClosedFourValueEnum` (unit) |
| T-0158 | Verify() core orchestration: recompute inputs, assemble signed_object, delegate to EdDSA-Protodoc-1 | FR-063, FR-067 | T-0136, T-0151, T-0153, T-0155, T-0157 | argus | `TestFR_067_VerifyReportsSignedStateSignerAndPresentation` (integration) |
| T-0159 | StateAvailability interface and CoveringUnavailableState verdict | FR-062 | T-0157, T-0158 | argus | `TestFR_062_UnavailableStateReportsDistinctVerdict` (integration) |
| T-0160 | Signature survives an append-only edit to a signed document | FR-066 | T-0153 | hephaestus | `TestFR_066_SignaturePreservedAcrossSubsequentEdit` (integration) |
| T-0161 | Per-state attestation reporting API with divergence-to-non-attestation | FR-067 | T-0158, T-0159, T-0156 | argus | `TestFR_067_PerStateAttestationReportCoversAllOutcomes` (integration) |
| T-0162 | Content-unit diff enumeration between a signed state and current state | FR-068 | T-0153 | hephaestus | `TestFR_068_EnumeratesUnitsDifferingFromSignedState` (unit) |
| T-0163 | Presentation-artefact staleness exemption for signed states | FR-069 | T-0154, T-0156 | hephaestus | `TestFR_069_SignedPresentationExemptFromStaleness` (unit) |
| T-0164 | Verify() failure classification and signer-identity display invariant | FR-115 | T-0158 | argus | `TestFR_115_UnverifiedNeverCarriesSignerIdentity` (integration) |
| T-0165 | Ambient-value exclusion audit for CoverageDescriptor and SIGNATURE encoding paths | NFR-005 | T-0145, T-0151, T-0152, T-0153 | argus | `TestNFR_005_NoAmbientValuesOutsideAllowlistedSites` (unit) |
| T-0166 | Golden conformance corpus: CoverageDescriptor and SIGNATURE at-limit/over-limit fixtures | FR-063 | T-0145, T-0150, T-0153 | momus | `TestFR_063_ConformanceCorpusAtLimitAndOverLimitFixtures` (conformance) |
| T-0363 | argus milestone-exit security review of M09 (coverage/signature/verify) | FR-063, FR-067, FR-115, NFR-005 | T-0153, T-0158, T-0159, T-0161, T-0164, T-0165, T-0166 | argus | `TestM09_ArgusMilestoneExitReviewRecorded` (conformance) |

**T-0145** CoverageDescriptor wire struct encode/decode

> Implement the coverage-descriptor grammar production (integrity.abnf S4: cd-mode, cd-covered-ranges, cd-uncovered-ranges, cd-bitmask) as a Go struct in the `integrity` package with PDL-TLV encode/decode via `pdlfmt` (M01) primitives. cd-mode is a closed u8 enum {0x00 TOTAL, 0x01 SUBSET}, reject 0x02-0xFF. segment-range is sr-start/sr-end varints, half-open [start,end). cd-bitmask is the covered_prefix_regions_bitmask (bit0 HEADER..bit4 INTEGRITY_BLOCK, bits5-7 reserved MBZ). M08's T-0139 already implements this identical wire shape and its PD-COVER-001..004 structural rules under FR-002 -- this task MUST NOT re-implement the struct or codec independently; it depends on and reuses T-0139's type, adding only the FR-063 signature-binding surface (round-trip codec exposure needed by the SIGNATURE record). Well-formedness rules PD-COVER-001..004 are enforced by T-0139 already; T-0146..T-0149 in this milestone are conformance re-verification of those same rules against the shared type from the signature-binding call sites, not a second implementation.

- **Implements:** FR-063
- **Depends on:** T-0139
- **DoD:** CoverageDescriptor (the type from M08's T-0139, imported/aliased, not re-declared) encodes and decodes byte-identically for both TOTAL and SUBSET mode fixtures when exercised via this package's signature-binding entry points; a reserved cd-mode value (0x02) and a nonzero reserved bitmask bit are rejected at decode time with a named rule id; a test asserts exactly one CoverageDescriptor Go type exists across the M08 and M09 packages (no duplicate declaration).
- **Test:** `TestFR_063_CoverageDescriptorRoundTrip` (unit)
- **Owner:** hephaestus

**T-0146** PD-COVER-001: reject zero-length segment ranges

> Enforce integrity.abnf S4 rule PD-COVER-001: sr-end MUST be strictly greater than sr-start; a zero-length range (sr-start == sr-end) is rejected, never silently dropped from the range list.

- **Implements:** FR-063
- **Depends on:** T-0145
- **DoD:** A CoverageDescriptor containing a range with sr-start==sr-end is rejected with rule_id PD-COVER-001 naming the offending range; a range with sr-end==sr-start+1 (minimal valid) is accepted.
- **Test:** `TestFR_063_PDCover001RejectsZeroLengthRange` (conformance)
- **Owner:** argus

**T-0147** PD-COVER-002: mandatory merge-adjacent canonicalisation

> Enforce integrity.abnf S4 PD-COVER-002: two adjacent entries (a,b) and (b,c) in the same range list (both covered, or both uncovered) MUST be encoded as one entry (a,c). A technically-sorted but unmerged adjacent pair is a decode-time rejection, not an alternate valid encoding -- exactly one valid encoding exists per coverage set.

- **Implements:** FR-063
- **Depends on:** T-0145
- **DoD:** A fixture with two mergeable adjacent ranges in one list is rejected naming PD-COVER-002; the single merged-entry encoding of the identical coverage set is accepted.
- **Test:** `TestFR_063_PDCover002RejectsUnmergedAdjacentRanges` (conformance)
- **Owner:** argus

**T-0148** PD-COVER-003: no-gap, no-overlap, full-domain well-formedness check

> Enforce integrity.abnf S4 PD-COVER-003: no ordinal may appear in both cd-covered-ranges and cd-uncovered-ranges, and every ordinal in [0, segment_count) at the winning CommitRingRecord's segment-count must be covered by exactly one of the two lists -- a gap or overlap is rejected before any signature verification step runs. Also encodes the TOTAL-mode consequence stated in cd-mode's own comment: TOTAL requires cd-covered-ranges to name every currently-populated non-ATTEST ordinal and cd-uncovered-ranges to be empty, derived from this same well-formedness rule rather than coded as a separate special case.

- **Implements:** FR-063
- **Depends on:** T-0145
- **DoD:** A descriptor with an uncovered gap ordinal, one with an ordinal in both lists, and a TOTAL-mode descriptor with a nonempty uncovered list are each rejected naming PD-COVER-003; a well-formed SUBSET split and a well-formed TOTAL descriptor both pass.
- **Test:** `TestFR_063_PDCover003RejectsGapOrOverlap` (conformance)
- **Owner:** argus

**T-0149** PD-COVER-004: ATTEST ordinals never nameable in any CoverageDescriptor

> Enforce integrity.abnf S4 PD-COVER-004: no range in cd-covered-ranges or cd-uncovered-ranges may name an ordinal whose current SegmentTableSlot.slot-segment-type = ATTEST (4), for any signature -- this closes FR-002's self-coverage circularity by construction. Requires the SegmentTable slot-type lookup from M01.

- **Implements:** FR-063
- **Depends on:** T-0145
- **DoD:** A descriptor naming an ATTEST-typed ordinal in either list is rejected naming PD-COVER-004, verified against a fixture with a real ATTEST segment present in the SegmentTable.
- **Test:** `TestFR_063_PDCover004RejectsAttestOrdinal` (conformance)
- **Owner:** argus

**T-0150** Property-based fuzzer for CoverageDescriptor range-list canonicalisation

> plan.md Section 10 first-class task (a), assigned to M09 per the spine's cross-cutting concerns: a property-based fuzzer generating arbitrary covered/uncovered range-list pairs, asserting the canonicalisation invariant (exactly one valid encoding per coverage set) and specifically exercising a range list at exactly the MAX_SEGMENTS (16,384) boundary as its own seed corpus entry, feeding the CP-012 continuous-fuzzing harness (M19).

- **Implements:** FR-063
- **Depends on:** T-0145, T-0146, T-0147, T-0148, T-0149
- **DoD:** Fuzz target registered and runnable via `go test -fuzz`; a corpus seed exists at exactly frame count = MAX_SEGMENTS; 10 minutes of local fuzzing produces zero crashes and zero canonicalisation-invariant violations.
- **Test:** `FuzzFR_063_CoverageDescriptorCanonicalisation` (fuzz)
- **Owner:** prometheus

**T-0151** coverage-descriptor-digest computation

> Implement integrity.abnf S3.2's coverage-descriptor-digest: SHA-256 over the coverage-descriptor structure's own canonical encoded octets exactly as carried inside the SIGNATURE record, so any change to a range entry or the bitmask changes this digest even though the descriptor is also carried verbatim alongside it.

- **Implements:** FR-063
- **Depends on:** T-0145
- **DoD:** Two descriptors differing in exactly one range boundary produce different digests; the identical descriptor encoded twice produces the identical digest (determinism).
- **Test:** `TestFR_063_CoverageDescriptorDigestDeterministic` (unit)
- **Owner:** hephaestus

**T-0152** signed_object preimage assembly and computation

> Implement integrity.abnf S3.2's signed-object-preimage = 0x04 \|\| t-c-root \|\| structure-digest \|\| presentation-artefact-digest \|\| coverage-descriptor-digest (129 octets), hashed once with SHA-256. t-c-root and structure-digest are supplied fresh at call time from M08's T_C/T_S implementation (T-0136, recomputed over the covered subtree set, never read from a stored field); presentation-artefact-digest comes from T-0155. This task owns only the 129-octet concatenation-and-hash step, not the four inputs' own computation.

- **Implements:** FR-063
- **Depends on:** T-0136, T-0151
- **DoD:** Given four fixed 32-octet test inputs, the preimage is exactly 129 octets in the documented field order and signed_object equals the independently-computed SHA-256 of that exact byte sequence.
- **Test:** `TestFR_063_SignedObjectPreimageAssembly` (unit)
- **Owner:** argus

**T-0153** SIGNATURE record wire struct encode/decode

> Implement the `signature` record (integrity.abnf S5): sig-discriminant (0x40), sig-param-set, sig-signed-object, sig-value (R\|\|S, 64 octets), sig-coverage (embeds T-0145's CoverageDescriptor), sig-cred-chain-ref, sig-revocation-ref, sig-time-attestation-ref, sig-intent, sig-presentation-ref (tags 1-9, 10-255 reserved). Reference-field validity (non-zero16 where mandatory) is checked structurally here; the LTV semantics of the referenced ATTESTATION_EVIDENCE frames are M10's job. Signature identity is the pair (t-c-root, structure-digest) at signing time; two SIGNATUREs over the identical covered state are distinguishable only by sig-param-set and sig-value.

- **Implements:** FR-063, FR-064
- **Depends on:** T-0151, T-0152
- **DoD:** A SIGNATURE record round-trips byte-identically through encode/decode with all 9 fields populated; sig-cred-chain-ref and sig-time-attestation-ref being zero16 are each rejected (MUST NOT be zero16 per S5); a MAX_SIGNATURES=64 at-limit and 65-signature over-limit fixture is included as a secondary conformance case per CON-010.
- **Test:** `TestFR_063_SignatureRecordRoundTrip` (unit)
- **Owner:** hephaestus

**T-0154** PresentationArtefact minimal record (profile version, page geometry, font set)

> Implement the PresentationArtefact record (document.abnf S7.4, data-model.md 2.15) to the extent FR-064 requires: a profile-version field, page geometry, and the complete set of fonts used, stored as a distinct segment addressable by unit-id and content-addressed via its own SegmentTableSlot.slot-digest. Full rendering-side consumption of this record (pagination, font subsetting details) is M14's scope; this task only builds the record shape and its digest-addressability needed for signature binding.

- **Implements:** FR-064
- **Depends on:** None
- **DoD:** A PresentationArtefact with profile version, page geometry, and a non-empty font list round-trips through encode/decode, and its SegmentTableSlot.slot-digest changes whenever any of the three fields changes.
- **Test:** `TestFR_064_PresentationArtefactRoundTrip` (unit)
- **Owner:** hephaestus

**T-0155** sig-presentation-ref binding and presentation-artefact-digest sourcing

> Wire SIGNATURE.sig-presentation-ref (a unit-id pointing at a PRESENTATION_ARTEFACT segment) so that signed_object's presentation-artefact-digest input (T-0152) is always read as the referenced segment's own current SegmentTableSlot.slot-digest, never separately stored inside the SIGNATURE record -- per S5's field comment, this closes the gap where data-model.md's Signature entity named the digest as an input to signed_object without ever stating which PresentationArtefact it belongs to.

- **Implements:** FR-064
- **Depends on:** T-0152, T-0154
- **DoD:** Changing the referenced PresentationArtefact's content (and therefore its slot-digest) changes signed_object's presentation-artefact-digest input on the next fresh computation, without the SIGNATURE record's own bytes changing.
- **Test:** `TestFR_064_SignedObjectSourcesPresentationDigestFromReferencedSlot` (unit)
- **Owner:** argus

**T-0156** Presentation mismatch detection reports unattested

> A reader presenting a signed document under any presentation other than its pinned PRESENTATION_ARTEFACT (T-0154/T-0155) MUST report the presentation as unattested rather than silently rendering under the mismatched presentation and claiming attestation.

- **Implements:** FR-065
- **Depends on:** T-0154, T-0155
- **DoD:** Given a signature pinned to presentation A, a verify call supplying presentation B (differing profile version or page geometry) returns an explicit unattested result naming both the pinned and the presented artefact ids.
- **Test:** `TestFR_065_MismatchedPresentationReportsUnattested` (integration)
- **Owner:** argus

**T-0157** Verdict enum type definition

> Define the verification-status type as a value distinct from document content: `Verdict` with exactly the 4 named members {Valid, AttestedWithDeclaredOmissions, Unverified, CoveringUnavailableState}. This task defines the closed enum and its String()/JSON mapping (matching cli.md S5's 3 externally-visible verdict strings plus the unavailable-state case surfaced via exit code 4); it does not implement the branching logic that produces each value -- that is T-0158/T-0159/T-0164, and the AttestedWithDeclaredOmissions branch's actual production is completed in M11 (redaction).

- **Implements:** FR-050
- **Depends on:** None
- **DoD:** Verdict is a closed Go type with exactly 4 defined constants, no fifth value constructible outside the package, and each has a stable JSON/string mapping matching cli.md's verdict vocabulary.
- **Test:** `TestFR_050_VerdictIsClosedFourValueEnum` (unit)
- **Owner:** hephaestus

**T-0158** Verify() core orchestration: recompute inputs, assemble signed_object, delegate to EdDSA-Protodoc-1

> Implement `integrity.Verify(doc, signatureID)`: recompute T_C_root and structure_digest fresh (M08's T-0136), recompute presentation-artefact-digest (T-0155) and coverage-descriptor-digest (T-0151), assemble signed_object (T-0152), and delegate to the EdDSA-Protodoc-1 7-step procedure (M06) with the signer's public key taken from the referenced credential chain. Returns Verdict.Valid on a positive result, naming the signed state, the signer, and the pinned presentation on success (FR-067's per-state attestation report), or Verdict.Unverified on any negative result at this stage (the off-allowlist/failure-specific handling is detailed further in T-0164).

- **Implements:** FR-063, FR-067
- **Depends on:** T-0136, T-0151, T-0153, T-0155, T-0157
- **DoD:** Verify() on a validly-signed fixture returns Valid naming signed state id, signer, and pinned presentation id; Verify() on a fixture with one tampered content octet inside the covered set returns Unverified.
- **Test:** `TestFR_067_VerifyReportsSignedStateSignerAndPresentation` (integration)
- **Owner:** argus

**T-0159** StateAvailability interface and CoveringUnavailableState verdict

> A signature covering a state the current file can no longer reconstruct MUST report a distinct 'covering an unavailable state' verdict, never 'failed' or 'valid' (integrity.abnf composed with FR-122's post-migration NO_HISTORY case). Since M09 depends only on M06/M08 and the actual reconstructability check lives in the History layer (M12, built later in the spine), this task defines a `StateAvailability` interface (`IsReconstructable(stateID) bool`) that Verify() consults before running EdDSA-Protodoc-1, and wires a test double for this milestone's own conformance; M12 supplies the concrete ErasureRecord-backed implementation as a follow-on task and this interface's contract is frozen now so M12 has a fixed target.

- **Implements:** FR-062
- **Depends on:** T-0157, T-0158
- **DoD:** Verify() with a StateAvailability double reporting `false` for the signed state's own t-c-root returns Verdict.CoveringUnavailableState without attempting cryptographic verification; with a double reporting `true` it proceeds to T-0158's normal path.
- **Test:** `TestFR_062_UnavailableStateReportsDistinctVerdict` (integration)
- **Owner:** argus

**T-0160** Signature survives an append-only edit to a signed document

> Committing an edit to a signed document must preserve the signature bound to its covered state and pinned presentation, unmodified. Because SIGNATURE is ATTEST-typed and the ledger (M02) is append-only, this task is a verification/regression test proving the property holds end-to-end rather than new production logic: a sealed SIGNATURE segment's bytes and its own referenced state must be byte-identical, and independently re-verifiable, before and after an unrelated content edit is committed.

- **Implements:** FR-066
- **Depends on:** T-0153
- **DoD:** Signing a document, then committing an unrelated content edit, then re-reading the original SIGNATURE segment's octets and re-running Verify() against the pre-edit state both yield the identical Valid result and identical signature bytes as before the edit.
- **Test:** `TestFR_066_SignaturePreservedAcrossSubsequentEdit` (integration)
- **Owner:** hephaestus

**T-0161** Per-state attestation reporting API with divergence-to-non-attestation

> Expose a reader-facing report combining T-0158's Valid-path fields (signed state, signer, pinned presentation) with T-0156's mismatch handling and T-0159's unavailable-state handling into one per-signature attestation report structure, so a caller gets exactly one of: attested (state+signer+presentation named), unattested (divergence named), or unavailable -- never an ambiguous partial result.

- **Implements:** FR-067
- **Depends on:** T-0158, T-0159, T-0156
- **DoD:** For each of the three outcome classes (attested, presentation-diverged, state-unavailable) a table-driven test asserts the single report shape returned names exactly the fields required for that class and no signer identity is present outside the attested case.
- **Test:** `TestFR_067_PerStateAttestationReportCoversAllOutcomes` (integration)
- **Owner:** argus

**T-0162** Content-unit diff enumeration between a signed state and current state

> Implement the library primitive enumerating every content unit (by unit_id) that differs between a signature's covered state and the document's current state (generalises to any two states of one lineage, per cli.md S6's `diff` verb, which wraps this primitive at the CLI layer in M18 -- CLI wiring itself is out of scope here). Compares T_C leaf sets keyed by unit_id, reporting added/removed/changed sets.

- **Implements:** FR-068
- **Depends on:** T-0153
- **DoD:** Given a signed-state snapshot and a current-state snapshot differing by one changed unit, one added unit, and one removed unit, the primitive returns exactly those three sets with correct unit_ids and reports an empty diff as a normal, non-error result when the two states are identical.
- **Test:** `TestFR_068_EnumeratesUnitsDifferingFromSignedState` (unit)
- **Owner:** hephaestus

**T-0163** Presentation-artefact staleness exemption for signed states

> A PresentationArtefact bound to a signed state (T-0155's sig-presentation-ref) is exempt from the general derived-artefact staleness-refusal rule (the pattern M14 applies to other derived artefacts): its digest-binding to a signature is itself sufficient authority, so it must never be refused as 'stale' merely because its own recomputed input digest check would otherwise apply. Implemented as a guard checked before any staleness refusal path runs against a PresentationArtefact.

- **Implements:** FR-069
- **Depends on:** T-0154, T-0156
- **DoD:** A PresentationArtefact referenced by a valid SIGNATURE is never flagged stale even when its independent input-digest check would (hypothetically) fail; an unsigned PresentationArtefact with the same hypothetical mismatch is still flagged stale, proving the exemption is signature-scoped, not global.
- **Test:** `TestFR_069_SignedPresentationExemptFromStaleness` (unit)
- **Owner:** hephaestus

**T-0164** Verify() failure classification and signer-identity display invariant

> A signature failure (EdDSA-Protodoc-1 rejects), any octet outside the claimed coverage_descriptor being relied upon, or an off-allowlist param_set_id (not in S10.2's registry) MUST each produce Verdict.Unverified with no signer identity and no positive badge attached to the result. This is the non-negotiable cli.md restates: 'verify never displays a signer identity alongside anything other than valid.'

- **Implements:** FR-115
- **Depends on:** T-0158
- **DoD:** Three fixtures (bad signature bytes, an off-allowlist param_set_id, and a tampered out-of-coverage octet relied upon) each produce Unverified with the report's signer-identity field empty/absent; a positive-badge field is never set alongside Unverified in any of the three cases.
- **Test:** `TestFR_115_UnverifiedNeverCarriesSignerIdentity` (integration)
- **Owner:** argus

**T-0165** Ambient-value exclusion audit for CoverageDescriptor and SIGNATURE encoding paths

> Audit every field written by T-0145/T-0151/T-0152/T-0153's encoders against CQ-004's 4-site ambient-value allowlist (identifier minting, redaction salt, signature value, time attestation): confirm no clock, machine, user, process, or session-derived value is ever written into CoverageDescriptor or SIGNATURE octets except sig-value itself (the signature value site, one of the 4 allowed sites) and the referenced time-attestation site (M10's field, referenced but not populated here).

- **Implements:** NFR-005
- **Depends on:** T-0145, T-0151, T-0152, T-0153
- **DoD:** A static field-by-field audit table enumerating every field in coverage-descriptor and signature is checked into the test, each row marked ambient-free or citing its allowlist site by name; a test asserts two independent encodings of the identical logical SIGNATURE (excluding sig-value's own nondeterministic-key-material dependency) produce identical non-sig-value octets.
- **Test:** `TestNFR_005_NoAmbientValuesOutsideAllowlistedSites` (unit)
- **Owner:** argus

**T-0166** Golden conformance corpus: CoverageDescriptor and SIGNATURE at-limit/over-limit fixtures

> Per CP-011/CON-010's cross-cutting requirement (every milestone introducing or extending a record shape ships at-limit and one-past-limit conformance fixtures before it exits), assemble the golden corpus for this milestone's two new record shapes: CoverageDescriptor (TOTAL mode, SUBSET mode, and the MAX_SEGMENTS-boundary range list) and SIGNATURE (at MAX_SIGNATURES=64 and at 65, one past limit). Each fixture carries its own expected verdict for two-implementation agreement (M19/NFR-028 scope).

- **Implements:** FR-063
- **Depends on:** T-0145, T-0150, T-0153
- **DoD:** The corpus directory contains one fixture file per named case (TOTAL, SUBSET, MAX_SEGMENTS-boundary, MAX_SIGNATURES-at-limit, MAX_SIGNATURES-over-limit) each paired with an expected-verdict JSON sidecar, and the test suite consumes every fixture in the directory (no fixture silently unexercised).
- **Test:** `TestFR_063_ConformanceCorpusAtLimitAndOverLimitFixtures` (conformance)
- **Owner:** momus

**T-0363** argus milestone-exit security review of M09 (coverage/signature/verify)

> Per CP-011's cross-cutting requirement, and matching the precedent set by M06's dedicated review task (T-0112, argus review of EdDSA-Protodoc-1), M09 currently has no consolidated milestone-exit security review -- only per-task argus-authored conformance tests. Run one holistic argus review across CoverageDescriptor well-formedness (T-0145..T-0149), signed_object/SIGNATURE construction (T-0151..T-0153), Verify() orchestration and failure classification (T-0158, T-0159, T-0161, T-0164), and the NFR-005 ambient-value audit (T-0165) before this milestone's code is treated as mergeable.

- **Implements:** FR-063, FR-067, FR-115, NFR-005
- **Depends on:** T-0153, T-0158, T-0159, T-0161, T-0164, T-0165, T-0166
- **DoD:** A single argus-authored review record enumerates every file touched by T-0145..T-0166, confirms every T-0165 ambient-audit row was checked and every finding either resolved or explicitly waived by Eyvar, and is committed to the repo before M09 is marked done; any unresolved finding blocks merge.
- **Test:** `TestM09_ArgusMilestoneExitReviewRecorded` (conformance)
- **Owner:** argus

---

### M10: Attestation Evidence & LTV

Long-term-validation material (credential chains, revocation, time attestation) is a distinct testable layer above bare signing that needs its own adversarial corpus.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0167 | ATTESTATION_EVIDENCE ae-kind/ae-format closed-enum pairing validation | FR-070 | None | hephaestus | `TestFR_070_AeKindFormatPairingRejectsInvalid` (unit) |
| T-0168 | ATTESTATION_EVIDENCE opaque DER blob carriage (encode/decode) | FR-070 | T-0167 | hephaestus | `ATTEST-EVID-ROUNDTRIP` (conformance) |
| T-0169 | Wire Signature.sig-cred-chain-ref / sig-revocation-ref / sig-time-attestation-ref | FR-070 | T-0167, T-0168, T-0138, T-0153 | hephaestus | `TestFR_070_SignatureLtvRefsRequireCorrectKind` (unit) |
| T-0170 | Mandatory nested cred-chain/revocation on time-attestation records | FR-071 | T-0167, T-0168 | hephaestus | `TestFR_071_NestedChainMandatoryForTimeAttestation` (unit) |
| T-0171 | Fixed, locally-held trust-anchor list for offline LTV verification | FR-070 | None | argus | `TestFR_070_TrustAnchorsNoNetworkCalls` (unit) |
| T-0172 | X.509 credential-chain offline parse and verify (ae-kind=0x00) | FR-070 | T-0168, T-0171 | argus | `TestFR_070_CredentialChainVerifiesOffline` (unit) |
| T-0173 | Revocation evidence (OCSP/CRL) offline parse and compromise-time extraction | FR-070 | T-0168, T-0171 | argus | `TestFR_070_RevocationEvidenceParsesOffline` (unit) |
| T-0174 | RFC 3161 TimeStampToken offline parse and verify against its own nested chain | FR-070, FR-071 | T-0170, T-0172 | argus | `TestFR_070_TimeAttestationParsesAndVerifiesOffline` (unit) |
| T-0175 | Full LTV evidence-chain offline verification pipeline (T-LTV) | FR-070 | T-0171, T-0172, T-0173, T-0174 | argus | `TestFR_070_FullEvidenceChainVerdictStableAcrossTime` (integration) |
| T-0176 | TSA own-chain independent verification corpus (T-TSA-LTV) | FR-071 | T-0170, T-0174 | argus | `T-TSA-LTV` (conformance) |
| T-0177 | sig-intent closed-enum validation (0x00-0x03) | FR-070 | None | hephaestus | `TestFR_070_SigIntentRejectsOutOfRange` (unit) |
| T-0178 | Signing-instant vs. attested-interval comparator (FR-072) | FR-072 | T-0174, T-0175 | argus | `TestFR_072_SigningInstantOutsideIntervalUnverified` (integration) |
| T-0179 | N-TIME negative conformance corpus (40 documents) | FR-072 | T-0178 | momus | `N-TIME` (conformance) |
| T-0180 | Revocation-compromise-time vs. signing-instant comparator (FR-073) | FR-073 | T-0173, T-0175 | argus | `TestFR_073_RevocationAtOrBeforeSigningUnverified` (integration) |
| T-0181 | N-COMPROMISE negative conformance corpus (30 documents) | FR-073 | T-0180 | momus | `N-COMPROMISE` (conformance) |
| T-0182 | Continuous fuzz target for ATTESTATION_EVIDENCE DER decode paths | FR-070 | T-0168, T-0172, T-0173, T-0174 | prometheus | `FuzzAttestationEvidenceDecode` (fuzz) |
| T-0183 | argus security review of the LTV/attestation-evidence subsystem | FR-070, FR-071, FR-072, FR-073 | T-0171, T-0175, T-0178, T-0180 | argus | `TestArgus_M10_LtvSecurityReviewChecklist` (integration) |
| T-0184 | ATTESTATION_EVIDENCE at-limit / one-past-limit conformance fixtures | FR-070 | T-0168 | momus | `ATTEST-EVID-CEILING-BOUNDARY` (conformance) |

**T-0167** ATTESTATION_EVIDENCE ae-kind/ae-format closed-enum pairing validation

> Implement the ae-kind (0x00 credential-chain, 0x01 revocation-evidence, 0x02 time-attestation; 0x03-0xFF reserved) and ae-format closed-enum validator per integrity.abnf S9/S10.4: reject ae-kind outside 0x00-0x02; reject any (ae-kind, ae-format) pair other than the 3 allowed combinations (0x00->0x00, 0x01->{0x01,0x02}, 0x02->0x03). Depends on M09's Signature/record decode scaffolding existing (not a task in this milestone).

- **Implements:** FR-070
- **Depends on:** None
- **DoD:** All 3 valid (ae-kind, ae-format) pairs decode successfully; every invalid pairing and every ae-kind in 0x03-0xFF is rejected with a distinct error identifying the offending value; no pairing silently defaults.
- **Test:** `TestFR_070_AeKindFormatPairingRejectsInvalid` (unit)
- **Owner:** hephaestus

**T-0168** ATTESTATION_EVIDENCE opaque DER blob carriage (encode/decode)

> Implement ae-discriminant/ae-id/ae-der-octets encode+decode per integrity.abnf S9: the record carries an opaque DER blob whole and unexecuted (CP-005) for all 3 ae-kind values (X.509 chain leaf-to-root RFC 5280, OCSP response RFC 6960 or CRL RFC 5280, TimeStampToken RFC 3161). Author one conformance vector per ae-kind/ae-format combination.

- **Implements:** FR-070
- **Depends on:** T-0167
- **DoD:** Byte-exact round trip (decode then re-encode reproduces identical octets) for one conformance vector per each of the 3 valid ae-kind/ae-format combinations; ae-der-octets is never parsed/interpreted by this encode/decode layer itself.
- **Test:** `ATTEST-EVID-ROUNDTRIP` (conformance)
- **Owner:** hephaestus

**T-0169** Wire Signature.sig-cred-chain-ref / sig-revocation-ref / sig-time-attestation-ref

> Wire the 3 Signature reference fields (integrity.abnf lines ~299-323) to resolved ATTESTATION_EVIDENCE records: sig-cred-chain-ref MUST NOT be zero16 and MUST resolve to ae-kind=0x00; sig-time-attestation-ref MUST NOT be zero16 and MUST resolve to ae-kind=0x02; sig-revocation-ref MUST resolve to ae-kind=0x01 OR be zero16 (permitted only when no revocation evidence was current at signing time). Depends on the Signature entity already built by M09 (T-0138) and the wire struct from T-0153.

- **Implements:** FR-070
- **Depends on:** T-0167, T-0168, T-0138, T-0153
- **DoD:** A Signature with a zero16 cred-chain-ref or time-attestation-ref is rejected naming the field; a Signature whose refs resolve to the wrong ae-kind is rejected naming the mismatch; a zero16 revocation-ref is accepted without further check.
- **Test:** `TestFR_070_SignatureLtvRefsRequireCorrectKind` (unit)
- **Owner:** hephaestus

**T-0170** Mandatory nested cred-chain/revocation on time-attestation records

> Implement the ae-nested-cred-chain/ae-nested-revocation mandatory-when-ae-kind=0x02 rule (integrity.abnf S9): both MUST be non-zero16 and MUST resolve to ae-kind 0x00 and 0x01 respectively when the parent record's ae-kind=0x02; both MUST be zero16 when the parent's ae-kind is 0x00 or 0x01 (no second level of nesting).

- **Implements:** FR-071
- **Depends on:** T-0167, T-0168
- **DoD:** A time-attestation record with either nested field zero16 is rejected naming which field is missing; a credential-chain or revocation-evidence record with a non-zero16 nested field is rejected; correctly-nested records pass.
- **Test:** `TestFR_071_NestedChainMandatoryForTimeAttestation` (unit)
- **Owner:** hephaestus

**T-0171** Fixed, locally-held trust-anchor list for offline LTV verification

> Build the trust-anchor component required by FR-070's zero-network, 30-year LTV guarantee: a fixed, locally-held x509.CertPool loaded from an in-repo/config-supplied anchor set, with no code path that performs a network fetch, live OCSP query, or CRL download during verification.

- **Implements:** FR-070
- **Depends on:** None
- **DoD:** Verification using this component succeeds and fails deterministically using only the local anchor set; a test harness that fails any attempted outbound network call (net.Dial stubbed to error) still passes verification of a valid corpus document.
- **Test:** `TestFR_070_TrustAnchorsNoNetworkCalls` (unit)
- **Owner:** argus

**T-0172** X.509 credential-chain offline parse and verify (ae-kind=0x00)

> Parse the concatenated leaf-to-root DER X.509 chain (RFC 5280) carried in ae-der-octets when ae-format=0x00, and verify it offline against the trust-anchor list from T-0171 using crypto/x509, rejecting malformed DER, wrong leaf-to-root ordering, and chains that do not terminate at a trusted anchor.

- **Implements:** FR-070
- **Depends on:** T-0168, T-0171
- **DoD:** A well-formed valid chain verifies offline with no network access attempted; a malformed, reordered, or untrusted-root chain is rejected with a distinct error, without panicking on truncated DER.
- **Test:** `TestFR_070_CredentialChainVerifiesOffline` (unit)
- **Owner:** argus

**T-0173** Revocation evidence (OCSP/CRL) offline parse and compromise-time extraction

> Parse ae-format=0x01 (OCSP response, RFC 6960) and ae-format=0x02 (CRL, RFC 5280) DER payloads, verify the responder/issuer signature offline against the trust-anchor list, and extract the revoked-credential compromise/revocation time for downstream use by T-0180 (FR-073).

- **Implements:** FR-070
- **Depends on:** T-0168, T-0171
- **DoD:** Both OCSP and CRL forms parse and verify offline; the extracted compromise time matches a hand-computed expected value on 2 fixture vectors (one OCSP, one CRL); no live responder/CRL-distribution-point fetch is attempted.
- **Test:** `TestFR_070_RevocationEvidenceParsesOffline` (unit)
- **Owner:** argus

**T-0174** RFC 3161 TimeStampToken offline parse and verify against its own nested chain

> Parse ae-format=0x03 TimeStampToken (RFC 3161) DER when ae-kind=0x02, extract the attested time interval, and verify the TSA's signature offline using the record's own ae-nested-cred-chain (via T-0172's chain verifier) rather than any external chain.

- **Implements:** FR-070, FR-071
- **Depends on:** T-0170, T-0172
- **DoD:** A valid TimeStampToken with a valid nested chain verifies offline and yields the correct attested interval on 2 fixture vectors; a token whose nested chain fails to verify is rejected as unverified, independent of the primary signer's own chain state.
- **Test:** `TestFR_070_TimeAttestationParsesAndVerifiesOffline` (unit)
- **Owner:** argus

**T-0175** Full LTV evidence-chain offline verification pipeline (T-LTV)

> Compose T-0171/T-0172/T-0173 into the complete FR-070 pipeline for one signature: verify credential chain + revocation evidence + time attestation (including the time attestation's own nested chain/revocation from T-0170/T-0174) fully offline against the fixed trust-anchor list, with an expired signing credential still yielding the same verdict as verification performed at signing time. This is the pipeline T-0178/T-0180 attach their comparators to.

- **Implements:** FR-070
- **Depends on:** T-0171, T-0172, T-0173, T-0174
- **DoD:** Running the T-LTV corpus with all network interfaces disabled and a fixed trust-anchor list yields, for 100% of the corpus, the identical verdict as verification recorded at signing time, including for documents whose signing credential has since expired.
- **Test:** `TestFR_070_FullEvidenceChainVerdictStableAcrossTime` (integration)
- **Owner:** argus

**T-0176** TSA own-chain independent verification corpus (T-TSA-LTV)

> Author and run the T-TSA-LTV conformance corpus: offline verification of every time attestation in the signature corpus succeeds using only its own nested credential chain and revocation evidence, independent of the primary signer's chain, including cases where the TSA's own credential has since expired.

- **Implements:** FR-071
- **Depends on:** T-0170, T-0174
- **DoD:** 100% of the signature corpus's time attestations verify offline via their own nested chain/revocation alone, including at least one fixture with an expired TSA credential that still verifies as current-at-attestation-time.
- **Test:** `T-TSA-LTV` (conformance)
- **Owner:** argus

**T-0177** sig-intent closed-enum validation (0x00-0x03)

> Implement the plan-assigned sig-intent closed set (integrity.abnf S10.3): 0x00 author-approval, 0x01 witness-attestation, 0x02 notarization, 0x03 custodial-transfer (RESCIND_RESIGN only); reject any value in 0x04-0xFF, satisfying FR-070's 'signing-intent value from a closed set.'

- **Implements:** FR-070
- **Depends on:** None
- **DoD:** Each of the 4 assigned sig-intent values decodes and round-trips; every value 0x04-0xFF is rejected naming the offending value.
- **Test:** `TestFR_070_SigIntentRejectsOutOfRange` (unit)
- **Owner:** hephaestus

**T-0178** Signing-instant vs. attested-interval comparator (FR-072)

> Implement the comparator that, given a signature's recorded signing instant and its verified time attestation's attested interval (from an offline-verified chain per T-0174), reports the document as Unverified whenever the signing instant lies outside that interval, with no signer identity shown alongside a non-valid verdict.

- **Implements:** FR-072
- **Depends on:** T-0174, T-0175
- **DoD:** A signing instant inside the attested interval yields the normal verdict chain unaffected; a signing instant before, after, or in a mismatched interval yields Unverified with zero signer identity fields populated in the output.
- **Test:** `TestFR_072_SigningInstantOutsideIntervalUnverified` (integration)
- **Owner:** argus

**T-0179** N-TIME negative conformance corpus (40 documents)

> Author the N-TIME negative corpus of 40 documents covering unsupported, future, and mismatched signing instants relative to their attested interval, and wire it into CI against T-0178's comparator.

- **Implements:** FR-072
- **Depends on:** T-0178
- **DoD:** All 40 N-TIME documents report Unverified with zero signer identity displayed, checked in CI on every change to integrity/ltv_verify.go.
- **Test:** `N-TIME` (conformance)
- **Owner:** momus

**T-0180** Revocation-compromise-time vs. signing-instant comparator (FR-073)

> Implement the comparator that reports the document as Unverified whenever the offline-verified revocation evidence's recorded compromise/revocation time is at or before the signature's recorded signing instant (the backdating case), with no signer identity shown.

- **Implements:** FR-073
- **Depends on:** T-0173, T-0175
- **DoD:** A revocation time strictly after the signing instant leaves the verdict chain unaffected; a revocation time at or before the signing instant yields Unverified with zero signer identity fields populated.
- **Test:** `TestFR_073_RevocationAtOrBeforeSigningUnverified` (integration)
- **Owner:** argus

**T-0181** N-COMPROMISE negative conformance corpus (30 documents)

> Author the N-COMPROMISE negative corpus of 30 documents with revocation compromise times at or before their signature's signing instant, and wire it into CI against T-0180's comparator. Includes the cross-implementation identical-verdict check named by FR-073's Verify clause (this task authors the fixture pair; producing the second implementation's run belongs to M19's CP-003 gate).

- **Implements:** FR-073
- **Depends on:** T-0180
- **DoD:** All 30 N-COMPROMISE documents report Unverified with zero signer identity displayed, checked in CI on every change to integrity/ltv_verify.go.
- **Test:** `N-COMPROMISE` (conformance)
- **Owner:** momus

**T-0182** Continuous fuzz target for ATTESTATION_EVIDENCE DER decode paths

> Wire a Go native fuzz target over the ATTESTATION_EVIDENCE decode path and its 3 embedded DER parsers (X.509 chain, OCSP/CRL, TimeStampToken) into the CP-012 continuous fuzzing harness, per the cross-cutting concern flagging M10 as a decode-path-over-untrusted-bytes milestone.

- **Implements:** FR-070
- **Depends on:** T-0168, T-0172, T-0173, T-0174
- **DoD:** FuzzAttestationEvidenceDecode runs continuously in CI's fuzzing job with zero panics/crashes on the seed corpus plus 1 hour of fuzzing; every malformed/truncated input is rejected with a typed error, never a panic.
- **Test:** `FuzzAttestationEvidenceDecode` (fuzz)
- **Owner:** prometheus

**T-0183** argus security review of the LTV/attestation-evidence subsystem

> Mandatory pre-merge argus security review of the full M10 change set per the crypto-touching-milestone rule: check for DER-parser malleability/ambiguity accepted as valid, trust-anchor-list injection or mutation risk, timing side channels in the interval/compromise-time comparators, and that no verified-vs-unverified branch ever displays a signer identity outside the Valid verdict.

- **Implements:** FR-070, FR-071, FR-072, FR-073
- **Depends on:** T-0171, T-0175, T-0178, T-0180
- **DoD:** Review checklist completed with zero open findings above low severity, or all findings triaged and either fixed or explicitly accepted-with-rationale by Eyvar before the M10 branch merges.
- **Test:** `TestArgus_M10_LtvSecurityReviewChecklist` (integration)
- **Owner:** argus

**T-0184** ATTESTATION_EVIDENCE at-limit / one-past-limit conformance fixtures

> Per CON-010/CP-011's ship-conformance-at-every-ceiling rule, author fixtures for the ATTESTATION_EVIDENCE frame at its applicable structural ceilings (ae-der-octets maximum length from the M01 generated ceiling table, and reserved ae-kind/ae-format boundary values) plus one octet/one value past each limit.

- **Implements:** FR-070
- **Depends on:** T-0168
- **DoD:** An at-limit ae-der-octets fixture decodes successfully; a one-octet-over fixture is rejected naming the exceeded ceiling and the observed value, matching FR-106's abort-before-allocation contract.
- **Test:** `ATTEST-EVID-CEILING-BOUNDARY` (conformance)
- **Owner:** momus

---

### M11: Redaction

Salted-commitment redactable subtrees and the redact/publish residue guarantees are the consumer of the T_C tagging and coverage machinery just built.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0185 | Redactable-subtree designation surface in signing flow | FR-074 | T-0138, T-0153 | argus | `TestFR_074_DesignateRedactableSubtree` (unit) |
| T-0186 | CSPRNG salt minting stored inside the redactable subtree | FR-074 | T-0185 | argus | `TestFR_074_SaltStoredInsideSubtree` (unit) |
| T-0187 | Salted-commitment digest construction H(0x02\|\|salt\|\|canon(subtree)) | FR-074, FR-075 | T-0185, T-0186 | argus | `TestFR_074_SaltedCommitmentDigest` (unit) |
| T-0188 | Conformance vectors for redactable-subtree designation and commitment round-trip | FR-074 | T-0185, T-0186, T-0187 | momus | `redact-designate-conformance-001` (conformance) |
| T-0189 | Hiding-floor analysis: resistance to 2^80 exhaustive search over omitted content | FR-075 | T-0187 | argus | `TestFR_075_HidingBoundAgainstBruteForce` (unit) |
| T-0190 | Track outstanding FR-061/FR-075 salted-commitment ruling before M11 close | FR-075 | T-0189 | clio | `gov-checklist-FR061-FR075-ruling-recorded` (integration) |
| T-0191 | Redact operation: subtree removal replacing plaintext with commitment leaf, T_C_root-preserving | FR-076 | T-0187 | argus | `TestFR_076_RedactPreservesTCRoot` (unit) |
| T-0192 | Verify() reports AttestedWithDeclaredOmissions enumerating every omitted unit | FR-076 | T-0191 | argus | `TestFR_076_VerifyReportsDeclaredOmissions` (integration) |
| T-0193 | Conformance vectors for declared-omission verdict enumeration | FR-076 | T-0191, T-0192 | momus | `redact-declared-omission-conformance-001` (conformance) |
| T-0194 | Undesignated omission causes T_C_root mismatch and Unverified verdict | FR-077 | T-0187, T-0191 | argus | `TestFR_077_UndesignatedOmissionUnverified` (unit) |
| T-0195 | Conformance vectors for undesignated-omission-to-Unverified, including near-miss cases | FR-077 | T-0194 | momus | `redact-undesignated-omission-conformance-001` (conformance) |
| T-0196 | Publish-time zero-residue guarantee for removed/superseded content units | FR-078 | T-0191 | hephaestus | `TestFR_078_PublishOmitsRemovedUnitOctets` (integration) |
| T-0197 | Orphan-carriage quoted_text included in the FR-078 redaction residue scan | FR-078 | T-0196 | momus | `redact-orphan-residue-conformance-001` (conformance) |
| T-0198 | Standalone adversarial residue scanner as a CI safety net | FR-078 | T-0196 | prometheus | `TestFR_078_ResidueScannerFindsNoMatches` (fuzz) |
| T-0199 | Document the FR-079 actor-identity inventory gap and request a spec ruling | FR-079 | None | clio | `TestGOV_FR079_ClarificationRecorded` (integration) |
| T-0200 | Actor-identity field inventory enumeration scoped to currently-existing fields | FR-079 | T-0199 | hephaestus | `TestFR_079_ActorFieldInventoryEnumeratesKnownFields` (unit) |
| T-0201 | Publish-time actor-identity value stripping | FR-080 | T-0200, T-0196 | hephaestus | `TestFR_080_PublishStripsActorIdentityValues` (integration) |
| T-0202 | Conformance vectors for actor-identity stripping on publish | FR-080 | T-0201 | momus | `redact-actor-strip-conformance-001` (conformance) |
| T-0203 | Custody/fixity value preservation check on publish | FR-081 | T-0196 | hephaestus | `TestFR_081_PublishPreservesCustodyFixityValues` (integration) |
| T-0204 | Conformance vectors for custody/fixity preservation across redact+publish | FR-081 | T-0203 | momus | `redact-custody-fixity-conformance-001` (conformance) |
| T-0205 | argus security review of the full redaction path | FR-074, FR-075, FR-076, FR-077, FR-078 | T-0187, T-0191, T-0192, T-0194, T-0196, T-0201, T-0203 | argus | `TestSEC_M11_ArgusReviewPassed` (integration) |
| T-0206 | Wire-format conformance corpus for the RedactionCommitment record shape | FR-074 | T-0187, T-0191 | momus | `redact-commitment-boundary-conformance-001` (conformance) |

**T-0185** Redactable-subtree designation surface in signing flow

> At signing time (integrity package, built on M08's T_C tree and M09's Signature flow), add a designation API letting the signer mark specific T_C subtrees as redactable. A designated subtree's T_C leaf is tagged leaf-redactable (tag 0x02, one of the 3 domain tags established in M08) instead of the plain content tag. Designation must be validated against the actual T_C tree shape (rejecting a designation naming a non-existent or non-leaf-aligned unit).

- **Implements:** FR-074
- **Depends on:** T-0138, T-0153
- **DoD:** Signing a state with a designated subtree list produces a T_C tree where exactly the designated subtrees carry tag 0x02, verified by direct tree inspection; designating a non-existent unit id returns an explicit error.
- **Test:** `TestFR_074_DesignateRedactableSubtree` (unit)
- **Owner:** argus

**T-0186** CSPRNG salt minting stored inside the redactable subtree

> For each redactable subtree, mint a fresh crypto/rand salt (>=128 bits) and store it inside the subtree's own serialized content (not in a side table), so that removing the subtree's plaintext also removes its salt as a side effect, and so an unredacted reader can recompute the same commitment.

- **Implements:** FR-074
- **Depends on:** T-0185
- **DoD:** Two designations of the same content produce different salts; the salt round-trips through subtree serialization and is recoverable only while the subtree is present.
- **Test:** `TestFR_074_SaltStoredInsideSubtree` (unit)
- **Owner:** argus

**T-0187** Salted-commitment digest construction H(0x02\|\|salt\|\|canon(subtree))

> Implement the commitment function used as the T_C leaf value for a redactable subtree: digest = H(0x02 \|\| salt \|\| canon(subtree)), using the domain-tag/canonicalization primitives from M08. This is the load-bearing construction both FR-074 (commitment form) and FR-075 (hiding floor) depend on.

- **Implements:** FR-074, FR-075
- **Depends on:** T-0185, T-0186
- **DoD:** Commitment digest is deterministic for identical (salt, subtree) input, differs for any single-bit change in either input, and matches the value integrity.abnf documents for a leaf-redactable T_C leaf.
- **Test:** `TestFR_074_SaltedCommitmentDigest` (unit)
- **Owner:** argus

**T-0188** Conformance vectors for redactable-subtree designation and commitment round-trip

> Ship a golden corpus exercising: a document with zero redactable subtrees, one, and several, each verified to produce the documented wire shape and T_C leaf tag per contracts/integrity.abnf, per the CP-011/CON-010 cross-cutting requirement that any milestone extending a wire shape ships its own conformance fixtures.

- **Implements:** FR-074
- **Depends on:** T-0185, T-0186, T-0187
- **DoD:** All corpus cases pass in two independent decode passes with identical verdicts; corpus checked into the conformance directory referenced by contracts/README.md.
- **Test:** `redact-designate-conformance-001` (conformance)
- **Owner:** momus

**T-0189** Hiding-floor analysis: resistance to 2^80 exhaustive search over omitted content

> Verify the salted-commitment construction from T-1103 meets FR-075's 2^80 hiding floor: assert salt entropy (>=128 bits, CSPRNG-sourced), assert commitment digests for differing subtrees/salts carry no exploitable structural correlation, and document the post-removal search cost (2^256 per plan.md Section 5 row 28). Explicitly flag: this is the FLAGGED DEVIATION from FR-061's literal bare-unsalted-digest text (plan.md Section 9 Conflict 1, self-disclosed, unresolved) — do not silently resolve the conflict, proceed only under the provisional salted form plan.md already adopted.

- **Implements:** FR-075
- **Depends on:** T-0187
- **DoD:** Test suite asserts salt entropy floor and absence of cross-subtree digest correlation across >=10,000 sampled pairs; the FR-061/FR-075 tension is recorded in the task's test file comment referencing plan.md Section 9 Conflict 1.
- **Test:** `TestFR_075_HidingBoundAgainstBruteForce` (unit)
- **Owner:** argus

**T-0190** Track outstanding FR-061/FR-075 salted-commitment ruling before M11 close

> Governance tracking task, not a code change: plan.md Section 9 Conflict 1 is unresolved — FR-061's literal text demands a bare unsalted severed-state digest, which conflicts with FR-075's 2^80 hiding floor. A salted-commitment form ships provisionally. Per the spine's disclosed-conflicts note, this gates M11's final wire-shape close. Record an explicit request for an Eyvar/themis ruling in clarify.md before phase 5 (analyze) closes; do not let this milestone's wire shape be treated as final until the ruling lands.

- **Implements:** FR-075
- **Depends on:** T-0189
- **DoD:** clarify.md contains a dated entry naming Conflict 1 (FR-061 vs FR-075), stating the provisional salted-commitment resolution in force, and marked open pending an Eyvar/themis ruling; presence verified by a CI doc-lint check that greps clarify.md for the entry, not by a Go test, since this task changes no code.
- **Test:** `gov-checklist-FR061-FR075-ruling-recorded` (integration)
- **Owner:** clio

**T-0191** Redact operation: subtree removal replacing plaintext with commitment leaf, T_C_root-preserving

> Implement the core `redact` package operation: given a signed state and a set of previously-designated redactable unit ids, remove their plaintext octets from the working ledger representation and replace the corresponding T_C leaves with only the salted-commitment digest from T-1103, leaving T_C_root (and therefore signed structure_digest) numerically unchanged from the original signed value.

- **Implements:** FR-076
- **Depends on:** T-0187
- **DoD:** Redacting any subset of designated subtrees yields a T_C_root bit-identical to the pre-redaction signed T_C_root, verified across single- and multi-subtree redaction cases.
- **Test:** `TestFR_076_RedactPreservesTCRoot` (unit)
- **Owner:** argus

**T-0192** Verify() reports AttestedWithDeclaredOmissions enumerating every omitted unit

> Extend the M09 Verify() pipeline so that when it encounters redaction-commitment leaves left by T-1107, it recomputes T_C_root successfully, emits the AttestedWithDeclaredOmissions verdict (not Valid, not Unverified), and enumerates every omitted unit id in the verdict payload.

- **Implements:** FR-076
- **Depends on:** T-0191
- **DoD:** Verifying a redacted-but-signed document returns AttestedWithDeclaredOmissions with an omission list matching exactly the redacted unit ids, for 1-, 2-, and N-subtree redactions.
- **Test:** `TestFR_076_VerifyReportsDeclaredOmissions` (integration)
- **Owner:** argus

**T-0193** Conformance vectors for declared-omission verdict enumeration

> Golden corpus covering zero, one, and multiple redacted subtrees signed and re-verified, asserting the exact AttestedWithDeclaredOmissions enumeration in each case matches the corpus fixture.

- **Implements:** FR-076
- **Depends on:** T-0191, T-0192
- **DoD:** Corpus cases pass under two independent decode/verify passes with identical enumerations.
- **Test:** `redact-declared-omission-conformance-001` (conformance)
- **Owner:** momus

**T-0194** Undesignated omission causes T_C_root mismatch and Unverified verdict

> Ensure the verification pipeline's T_C_root recomputation treats any missing or altered subtree that lacks a valid leaf-redactable (tag 0x02) commitment as a structural mismatch, producing the Unverified verdict rather than AttestedWithDeclaredOmissions or Valid. This distinguishes lawful declared-omission (FR-076, T-1108) from unlawful undesignated removal.

- **Implements:** FR-077
- **Depends on:** T-0187, T-0191
- **DoD:** Removing a subtree's plaintext without going through T-1107's redact operation (no commitment leaf, or an untagged leaf) causes Verify() to return Unverified in every tested case, never AttestedWithDeclaredOmissions.
- **Test:** `TestFR_077_UndesignatedOmissionUnverified` (unit)
- **Owner:** argus

**T-0195** Conformance vectors for undesignated-omission-to-Unverified, including near-miss cases

> Corpus cases: plain deletion with no commitment leaf; a leaf with the correct salt but a missing/corrupted tag byte; a leaf using the wrong domain tag. All must resolve to Unverified.

- **Implements:** FR-077
- **Depends on:** T-0194
- **DoD:** All near-miss corpus cases resolve to Unverified under two independent verify passes.
- **Test:** `redact-undesignated-omission-conformance-001` (conformance)
- **Owner:** momus

**T-0196** Publish-time zero-residue guarantee for removed/superseded content units

> Implement the redact package's own scoped fresh re-emission path (ahead of the full canon.L* traversal landing in M17) that walks only reachable, non-removed content when producing publish output, guaranteeing zero octets of any removed or superseded content unit appear in the emitted file. This is deliberately narrower than the general-purpose canon package and is expected to be superseded/subsumed once M17 lands.

- **Implements:** FR-078
- **Depends on:** T-0191
- **DoD:** For a document with N redacted units, byte-searching the published output for any octet run unique to the pre-redaction plaintext of those units returns zero matches, across single- and multi-unit redaction cases.
- **Test:** `TestFR_078_PublishOmitsRemovedUnitOctets` (integration)
- **Owner:** hephaestus

**T-0197** Orphan-carriage quoted_text included in the FR-078 redaction residue scan

> First-class task named explicitly in plan.md Section 10 item (d) and assigned to this milestone: build the corpus case where a redacted unit had an orphaned annotation (FR-030's orphan-quoted field) carrying its quoted_text, and confirm the publish residue scan from T-1112 detects and excludes that quoted_text octet-for-octet, since it is a secondary carrier of the same content the threat model (plan.md Section 6) flags as not yet demonstrated.

- **Implements:** FR-078
- **Depends on:** T-0196
- **DoD:** The corpus case's published output contains zero octets of the orphan-quoted_text field that duplicated the redacted unit's content; test fails if the scan is scoped only to the primary content unit.
- **Test:** `redact-orphan-residue-conformance-001` (conformance)
- **Owner:** momus

**T-0198** Standalone adversarial residue scanner as a CI safety net

> plan.md Section 6 flags the residue-scan guarantee as not yet exercised against an adversarial or property-based corpus. Build an independent post-hoc scanner (outside the publish code path itself, to avoid sharing a blind spot) that greps published output for any octet substring matching removed content taken from the pre-redaction state, used both as a test oracle here and wired into CP-012 continuous fuzzing.

- **Implements:** FR-078
- **Depends on:** T-0196
- **DoD:** The scanner runs against a corpus of randomly generated redact operations and reports zero residue matches across >=1000 fuzz iterations with no false negatives on seeded-defect mutants.
- **Test:** `TestFR_078_ResidueScannerFindsNoMatches` (fuzz)
- **Owner:** prometheus

**T-0199** Document the FR-079 actor-identity inventory gap and request a spec ruling

> KNOWN GAP, flagged explicitly rather than silently resolved: the only actor/author-carrying field found anywhere in data-model.md/document.abnf is Annotation.orphan.author_ref (populated only once an annotation is orphaned). document.abnf's own comment asserts a live (non-orphaned) annotation's authorship is 'carried by the identity apparatus of its ann-body-block's own runs,' but Run's record shape (document.abnf S2.1) carries no author/actor field, and FR-023/CQ-004 explicitly bar run_id from carrying any actor-derived value -- so that comment's own claim is unsupported by the record it points to. This task does NOT unilaterally add an author field to Run (that would reopen an approved, frozen data-model outside phase-4 authority); it records the inconsistency as a clarification request for Eyvar/themis.

- **Implements:** FR-079
- **Depends on:** None
- **DoD:** clarify.md carries a dated, explicit open item describing the Run-authorship / orphan.author_ref inconsistency and requesting a ruling on where live-annotation authorship is actually carried, referenced by FR-079/080/081.
- **Test:** `TestGOV_FR079_ClarificationRecorded` (integration)
- **Owner:** clio

**T-0200** Actor-identity field inventory enumeration scoped to currently-existing fields

> Implement a single enumerable, machine-readable inventory (e.g. content.ActorFieldInventory()) covering every actor/device-identity-carrying field that actually exists today: Annotation.orphan.author_ref, plus any actor-bearing fields already defined in Signature/ATTESTATION_EVIDENCE from M09/M10 (credential chain, revocation evidence references). Explicitly documented in code comments as incomplete pending T-1115's clarification ruling -- this is a best-effort inventory over a near-empty foundation, not a claim of completeness.

- **Implements:** FR-079
- **Depends on:** T-0199
- **DoD:** ActorFieldInventory() returns every field enumerated above with its record path; a new actor-bearing field added anywhere in the schema and not registered here fails a static-registry-completeness check.
- **Test:** `TestFR_079_ActorFieldInventoryEnumeratesKnownFields` (unit)
- **Owner:** hephaestus

**T-0201** Publish-time actor-identity value stripping

> Using the inventory from T-1116, strip every actor/device-identity-carrying field value from publish output (e.g. zero/omit Annotation.orphan.author_ref values in published documents). Because the inventory is bounded by T-1116's known-incomplete scope, this task's completeness is likewise bounded -- a newly-discovered actor field surfaced by T-1115's ruling becomes a follow-up task, not a silent gap in this one.

- **Implements:** FR-080
- **Depends on:** T-0200, T-0196
- **DoD:** Publishing a document containing every inventoried actor-identity field yields output where none of those field values are present or recoverable, verified by direct byte inspection of the published output.
- **Test:** `TestFR_080_PublishStripsActorIdentityValues` (integration)
- **Owner:** hephaestus

**T-0202** Conformance vectors for actor-identity stripping on publish

> Golden corpus: a document with populated orphan.author_ref and other inventoried actor fields, published, asserting zero octets of the pre-publish actor values remain while the document stays structurally valid.

- **Implements:** FR-080
- **Depends on:** T-0201
- **DoD:** Corpus cases pass under two independent publish/inspect passes with zero actor-value residue and a structurally valid output document.
- **Test:** `redact-actor-strip-conformance-001` (conformance)
- **Owner:** momus

**T-0203** Custody/fixity value preservation check on publish

> Define the concrete set of 'custody and fixity' fields (segment/slot digests, signature and CoverageDescriptor values, T_S/T_C-derived digests not themselves stripped by redaction) that MUST survive publish unchanged, and implement a byte-comparison assertion that every such field's value in published output equals its corresponding pre-publish value.

- **Implements:** FR-081
- **Depends on:** T-0196
- **DoD:** For a redact+publish round trip, every field in the defined custody/fixity set is byte-identical between pre-publish and published states; a diff on any such field fails the check.
- **Test:** `TestFR_081_PublishPreservesCustodyFixityValues` (integration)
- **Owner:** hephaestus

**T-0204** Conformance vectors for custody/fixity preservation across redact+publish

> Golden corpus exercising a signed, then redacted, then published document, asserting every custody/fixity field defined in T-1119 is preserved exactly while actor-identity fields (T-1117) and removed content (T-1112) are absent.

- **Implements:** FR-081
- **Depends on:** T-0203
- **DoD:** Corpus cases pass under two independent decode passes with identical custody/fixity field values.
- **Test:** `redact-custody-fixity-conformance-001` (conformance)
- **Owner:** momus

**T-0205** argus security review of the full redaction path

> Cross-cutting concern (spine: argus security review applies to M11 as a crypto/signing/redaction milestone). Review salt entropy and CSPRNG source, commitment binding/hiding properties, residue-scan completeness (including the orphan-quoted_text and adversarial-scanner tasks), and the undesignated-omission detection path, before this milestone's work merges.

- **Implements:** FR-074, FR-075, FR-076, FR-077, FR-078
- **Depends on:** T-0187, T-0191, T-0192, T-0194, T-0196, T-0201, T-0203
- **DoD:** A recorded argus review sign-off exists covering every listed dependency task, with zero open high/critical findings, before any M11 task is considered merge-ready.
- **Test:** `TestSEC_M11_ArgusReviewPassed` (integration)
- **Owner:** argus

**T-0206** Wire-format conformance corpus for the RedactionCommitment record shape

> Cross-cutting concern (CP-011/CON-010): any milestone introducing or extending a PDL-TLV/integrity.abnf record shape ships at-limit and one-past-limit conformance fixtures before it exits. Ship the RedactionCommitment wire-shape corpus at zero redacted subtrees, one, and a count adjacent to the relevant structural ceiling.

- **Implements:** FR-074
- **Depends on:** T-0187, T-0191
- **DoD:** At-limit and one-past-limit RedactionCommitment fixtures are checked in and pass under two independent decode passes with identical verdicts.
- **Test:** `redact-commitment-boundary-conformance-001` (conformance)
- **Owner:** momus

---

### M12: History & Erasure

HistorySegment reconstruction and retention-point gating need run/unit identity (M04) and the state_id comparator from integrity (M08) before they are meaningful.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0207 | Implement closed HistoryMode enum with immutability enforcement | CON-022 | None | hephaestus | `TestCON_022_HistoryModeClosedEnumAndImmutable` (unit) |
| T-0208 | Implement HistorySegment RLE-batched operation-batch encoder | FR-059, FR-060 | None | hephaestus | `TestHistorySegment_RLEBatchEncoding` (unit) |
| T-0209 | Implement HistorySegment decoder with malformed/at-limit rejection | FR-059, FR-060 | T-0208 | hephaestus | `TestHistorySegment_DecodeConformanceAtAndOverLimit` (conformance) |
| T-0210 | Wire causal_predecessor_state_id for offline lineage walking | FR-005 | T-0208 | hephaestus | `TestFR_005_PredecessorChainWalkable` (unit) |
| T-0211 | Implement offline nearest-common-ancestor determination | FR-005 | T-0210 | hephaestus | `TestFR_005_NearestCommonAncestorOffline` (integration) |
| T-0212 | Implement COMPLETE_HISTORY-mode state reconstruction | FR-059 | T-0207, T-0209 | hephaestus | `TestFR_059_CompleteHistoryReconstructsEveryState` (integration) |
| T-0213 | Wire retention_point field with mode-gated monotonicity | CON-022 | T-0207 | hephaestus | `TestCON_022_RetentionPointModeGated` (unit) |
| T-0214 | Implement RETAINED_FROM_POINT-mode state reconstruction | FR-060 | T-0213, T-0209 | hephaestus | `TestFR_060_RetainedFromPointReconstructsGuaranteedRange` (integration) |
| T-0215 | Enforce NO_HISTORY-mode current-state-only retention | CON-022 | T-0207 | hephaestus | `TestCON_022_NoHistoryModeRetainsOnlyCurrentState` (integration) |
| T-0216 | Refuse in-place content removal on COMPLETE_HISTORY documents | CON-023 | T-0207 | hephaestus | `TestCON_023_RefusesInPlaceRemovalOnCompleteHistory` (unit) |
| T-0217 | Implement ErasureRecord wire shape (provisional salted-commitment form) | FR-061 | T-0209, T-0187 | mnemosyne | `TestFR_061_ErasureRecordSaltedCommitmentForm` (unit) |
| T-0218 | Emit ErasureRecord enumeration on lawful trim | FR-061 | T-0217, T-0213 | mnemosyne | `TestFR_061_ErasureEnumeratedOnRetentionAdvance` (integration) |
| T-0219 | Separate materialized current-state read path from history region | NFR-033 | T-0208 | hephaestus | `TestNFR_033_OpenCostIndependentOfOpCount` (benchmark) |
| T-0220 | Benchmark: 10x op-count difference opens within 1.5x | NFR-033 | T-0219 | prometheus | `TestNFR_033_TenXOpCountOpensWithin1_5x` (benchmark) |
| T-0221 | Tune RLE run-merging to meet the 250k-op size-overhead budget | NFR-032 | T-0208 | hephaestus | `TestNFR_032_CompleteHistoryOverheadBudget250kOps` (benchmark) |
| T-0222 | Lock in the disclosed adversarial-trace overhead MISS as a regression guard | NFR-032 | T-0221 | momus | `TestNFR_032_AdversarialScatteredTraceDocumentedMiss` (benchmark) |
| T-0223 | Ship HistorySegment/ErasureRecord conformance corpus | FR-059, FR-060, FR-061 | T-0209, T-0217 | momus | `TestConformance_HistorySegmentAndErasureCorpus` (conformance) |
| T-0224 | Finalize history package public API and HC-* traceability comments | FR-059, FR-060, CON-022 | T-0212, T-0214, T-0215 | hephaestus | `TestHistoryPackage_PublicAPIStableAndTraceable` (unit) |

**T-0207** Implement closed HistoryMode enum with immutability enforcement

> Implement history.Mode type {CompleteHistory=0, RetainedFromPoint=1, NoHistory=2} read from Header.history-mode (container.abnf S2, field byte-layout frozen by M01; M12 owns the semantic enforcement). Reject any other stored value at parse time, naming the observed value. Reject any commit that would change history-mode after document creation, refusing and naming both the original and attempted value.

- **Implements:** CON-022
- **Depends on:** None
- **DoD:** Unit tests cover all 3 valid enum values, one invalid stored value (rejected, value named), and one post-creation mode-change attempt (rejected, both values named).
- **Test:** `TestCON_022_HistoryModeClosedEnumAndImmutable` (unit)
- **Owner:** hephaestus

**T-0208** Implement HistorySegment RLE-batched operation-batch encoder

> Implement history.EncodeSegment producing the RLE-batched HistorySegment record per document.abnf S9 -- one representation serving all three CON-022 history modes. Apply DP-003 run-merging so contiguous same-burst edits collapse into a single RLE entry rather than one per keystroke. Each batch entry carries its causal_predecessor_state_id.

- **Implements:** FR-059, FR-060
- **Depends on:** None
- **DoD:** Encoding a synthetic 10,000-op trace with 50 contiguous typing bursts produces exactly 50 RLE batch entries, each carrying a populated causal_predecessor_state_id.
- **Test:** `TestHistorySegment_RLEBatchEncoding` (unit)
- **Owner:** hephaestus

**T-0209** Implement HistorySegment decoder with malformed/at-limit rejection

> Implement history.DecodeSegment, rejecting truncated/oversized-length/unparseable HistorySegment octets per the validation pipeline's error-precedence rule, with at-limit and one-past-limit fixtures per CON-010.

- **Implements:** FR-059, FR-060
- **Depends on:** T-0208
- **DoD:** Decoder round-trips T-1201's encoder output byte-exact; the at-limit fixture is accepted and the one-past-limit fixture is rejected with the correct rule id.
- **Test:** `TestHistorySegment_DecodeConformanceAtAndOverLimit` (conformance)
- **Owner:** hephaestus

**T-0210** Wire causal_predecessor_state_id for offline lineage walking

> Populate and expose HistorySegment.causal_predecessor_state_id (op-predecessor, document.abnf S9) per commit, linked to the CommitRingRecord.parent_state_id chain, so every state's ancestor chain is walkable entirely from within the file.

- **Implements:** FR-005
- **Depends on:** T-0208
- **DoD:** For a synthetic 5-commit lineage, walking causal_predecessor_state_id from the tip reaches every ancestor state_id in order using only in-file data.
- **Test:** `TestFR_005_PredecessorChainWalkable` (unit)
- **Owner:** hephaestus

**T-0211** Implement offline nearest-common-ancestor determination

> Implement history.NearestCommonAncestor(fileA, fileB), determining the NCA of two divergent copies from the two files' parent_state_id/causal_predecessor_state_id chains alone, with no network access and no external ancestor store.

- **Implements:** FR-005
- **Depends on:** T-0210
- **DoD:** Given two files forked from a common state then diverged independently by 3 commits each, the function returns the correct common ancestor state_id using only the two files as input.
- **Test:** `TestFR_005_NearestCommonAncestorOffline` (integration)
- **Owner:** hephaestus

**T-0212** Implement COMPLETE_HISTORY-mode state reconstruction

> Implement history.Reconstruct(state_id) for documents declaring history-mode=COMPLETE_HISTORY, replaying every retained HistorySegment batch to rebuild any previously published state.

- **Implements:** FR-059
- **Depends on:** T-0207, T-0209
- **DoD:** For a 20-commit COMPLETE_HISTORY document, every one of the 20 published state_ids reconstructs to content matching the snapshot captured at commit time.
- **Test:** `TestFR_059_CompleteHistoryReconstructsEveryState` (integration)
- **Owner:** hephaestus

**T-0213** Wire retention_point field with mode-gated monotonicity

> Wire CommitRingRecord.retention-point into the history package for RETAINED_FROM_POINT-mode documents. Validate retention_point only advances monotonically and only under RETAINED_FROM_POINT mode.

- **Implements:** CON-022
- **Depends on:** T-0207
- **DoD:** Setting retention_point on a COMPLETE_HISTORY or NO_HISTORY document is rejected naming the declared mode; a backward move of retention_point is rejected.
- **Test:** `TestCON_022_RetentionPointModeGated` (unit)
- **Owner:** hephaestus

**T-0214** Implement RETAINED_FROM_POINT-mode state reconstruction

> Implement reconstruction for RETAINED_FROM_POINT-mode documents: every state at or after retention_point reconstructs successfully; a request for a state before retention_point yields the defined unavailable result rather than a partial or fabricated reconstruction.

- **Implements:** FR-060
- **Depends on:** T-0213, T-0209
- **DoD:** For a document with retention_point set after commit 5 of 10, states 5-10 reconstruct correctly and states 1-4 each return the defined unavailable result.
- **Test:** `TestFR_060_RetainedFromPointReconstructsGuaranteedRange` (integration)
- **Owner:** hephaestus

**T-0215** Enforce NO_HISTORY-mode current-state-only retention

> For history-mode=NO_HISTORY documents, verify no HistorySegment beyond the current state is retained or emitted, and any reconstruction request for a non-current state_id returns the defined unavailable result.

- **Implements:** CON-022
- **Depends on:** T-0207
- **DoD:** A NO_HISTORY document commits 5 edits; only the current state reconstructs, the 4 prior states each return the unavailable result, and no HistorySegment octets for superseded states exist in the file.
- **Test:** `TestCON_022_NoHistoryModeRetainsOnlyCurrentState` (integration)
- **Owner:** hephaestus

**T-0216** Refuse in-place content removal on COMPLETE_HISTORY documents

> Refuse any writer-initiated in-place content removal that would make prior published octets structurally unreconstructable, on a document declaring history-mode=COMPLETE_HISTORY, naming the declared mode in the refusal.

- **Implements:** CON-023
- **Depends on:** T-0207
- **DoD:** An in-place removal attempt against a COMPLETE_HISTORY document is refused with an error naming COMPLETE_HISTORY; the identical operation against RETAINED_FROM_POINT or NO_HISTORY documents proceeds.
- **Test:** `TestCON_023_RefusesInPlaceRemovalOnCompleteHistory` (unit)
- **Owner:** hephaestus

**T-0217** Implement ErasureRecord wire shape (provisional salted-commitment form)

> Implement ErasureRecord (data-model.md 2.17) using the provisional salted-commitment digest form per plan.md Section 9 Conflict 1, NOT FR-061's literal bare unsalted digest. Reuse the salted-commitment digest formula already built by M11 (T-0187) rather than reimplementing it independently. This is a disclosed, unresolved conflict (FR-061's literal text vs FR-075's 2^80 hiding-floor requirement) pending an explicit Eyvar/themis ruling before phase 5 (analyze) closes -- do not silently resolve by picking a side. Implement exactly the salted form plan.md proposes and record the open conflict in the record's doc comment.

- **Implements:** FR-061
- **Depends on:** T-0209, T-0187
- **DoD:** ErasureRecord encodes/decodes the salted-commitment form byte-exact per plan.md's Conflict-1 proposal, reusing M11's (T-0187) digest formula rather than a reimplementation; the doc comment and this task's test both cite the open FR-061/FR-075 conflict so it is not mistaken for a settled reading.
- **Test:** `TestFR_061_ErasureRecordSaltedCommitmentForm` (unit)
- **Owner:** mnemosyne

**T-0218** Emit ErasureRecord enumeration on lawful trim

> On any lawful trim that advances retention_point (RETAINED_FROM_POINT mode) or otherwise makes a previously published state no longer reconstructable, emit one ErasureRecord per such state naming its state identifier and salted-commitment digest.

- **Implements:** FR-061
- **Depends on:** T-0217, T-0213
- **DoD:** Advancing retention_point past 3 previously-published states produces exactly 3 ErasureRecords naming the correct state_id and digest each; no ErasureRecord is produced for states that remain reconstructable.
- **Test:** `TestFR_061_ErasureEnumeratedOnRetentionAdvance` (integration)
- **Owner:** mnemosyne

**T-0219** Separate materialized current-state read path from history region

> Structure storage so the current-content read path never traverses the HistorySegment region (DP-008 rationale): current-state open/access cost is a function of current content size only, never of total retained operation count.

- **Implements:** NFR-033
- **Depends on:** T-0208
- **DoD:** A profiling test shows octets read on open for the current state stay within 5% across two documents with identical current content but a 10x difference in total historical operation count.
- **Test:** `TestNFR_033_OpenCostIndependentOfOpCount` (benchmark)
- **Owner:** hephaestus

**T-0220** Benchmark: 10x op-count difference opens within 1.5x

> CI benchmark that opens two COMPLETE_HISTORY documents with identical current content but a 10x difference in total historical operation count, asserting open wall-clock time and peak memory of the larger document are within 1.5x of the smaller.

- **Implements:** NFR-033
- **Depends on:** T-0219
- **DoD:** Benchmark passes in CI with a recorded ratio <=1.5x for both time and memory; the ratios are logged as build artifacts.
- **Test:** `TestNFR_033_TenXOpCountOpensWithin1_5x` (benchmark)
- **Owner:** prometheus

**T-0221** Tune RLE run-merging to meet the 250k-op size-overhead budget

> Apply DP-003 run-merging plus RLE batching so a non-adversarial 250,000-operation/100,000-character reference edit trace produces a HistorySegment size no more than 2.0x the same final content saved with history-mode=NO_HISTORY.

- **Implements:** NFR-032
- **Depends on:** T-0208
- **DoD:** Running the 250k-op/100k-char reference benchmark trace yields a COMPLETE_HISTORY file <=2.0x the size of the NO_HISTORY file for identical final content.
- **Test:** `TestNFR_032_CompleteHistoryOverheadBudget250kOps` (benchmark)
- **Owner:** hephaestus

**T-0222** Lock in the disclosed adversarial-trace overhead MISS as a regression guard

> Add a permanent regression case for the plan.md Section 5-disclosed MISS: an adversarially scattered single-character-edit trace (edits scattered across many distant runs rather than contiguous bursts) exceeds the 2.0x budget, with no repair path since HC-011 (mapped to CON-005/DP-003's one-representation-per-run discipline) forbids re-basing runs after the fact. This task does not fix the MISS -- none exists per plan.md -- it locks in the documented, accepted limitation so a future change cannot silently regress it further unreviewed.

- **Implements:** NFR-032
- **Depends on:** T-0221
- **DoD:** The adversarial-trace benchmark runs in CI, its measured overhead ratio is recorded and asserted to not regress beyond the currently-measured value, and specs/CHANGES.md carries a one-line entry recording this as an accepted, not fixed, limitation.
- **Test:** `TestNFR_032_AdversarialScatteredTraceDocumentedMiss` (benchmark)
- **Owner:** momus

**T-0223** Ship HistorySegment/ErasureRecord conformance corpus

> Ship the golden conformance corpus for the HistorySegment and ErasureRecord record shapes per CP-011/CON-010: one fixture per record variant (COMPLETE_HISTORY, RETAINED_FROM_POINT, NO_HISTORY HistorySegment; ErasureRecord), each with an at-limit and one-past-limit pair, so a second independent implementation can validate against the same corpus.

- **Implements:** FR-059, FR-060, FR-061
- **Depends on:** T-0209, T-0217
- **DoD:** Corpus files exist for all 3 HistorySegment mode variants and for ErasureRecord, each with an at-limit and over-limit pair, and the reference implementation passes every fixture with the documented verdict.
- **Test:** `TestConformance_HistorySegmentAndErasureCorpus` (conformance)
- **Owner:** momus

**T-0224** Finalize history package public API and HC-* traceability comments

> Finalize the `history` package's exported API (Mode, Reconstruct, NearestCommonAncestor, ErasureRecord accessors) for consumption by `merge` (M13), `validate` (M07), and `canon` (M17). Ensure every doc comment resolves any HC-* legacy identifier it carries to its governing FR-/NFR-/CON- id, per the analysis.md traceability requirement that HC-* never be treated as an independent requirement axis.

- **Implements:** FR-059, FR-060, CON-022
- **Depends on:** T-0212, T-0214, T-0215
- **DoD:** Package compiles with a stable exported API consumed by a stub caller in merge/validate scaffolding; a repo grep confirms zero bare HC-* references without an adjacent FR-/NFR-/CON- id in this package.
- **Test:** `TestHistoryPackage_PublicAPIStableAndTraceable` (unit)
- **Owner:** hephaestus

---

### M13: Concurrent-Edit / Merge

The exhaustive DISJOINT-COMMUTE/R1/R2/R3 classification needs content identity and the history layer's causal-predecessor/state_id data to exist first.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0225 | Operation-kind taxonomy and classifier skeleton | FR-092 | None | hephaestus | `TestFR_092_ClassifyDispatchHasNoWildcardCase` (unit) |
| T-0226 | Implement DISJOINT-COMMUTE classification rule | FR-092 | T-0225 | hephaestus | `TestFR_092_DisjointCommuteOrderIndependent` (unit) |
| T-0227 | Implement R1-SEQUENCE-ORDER rule via Fugue | FR-092, FR-093 | T-0225 | hephaestus | `TestFR_093_ContiguousBurstSurvivesAsSubstring` (unit) |
| T-0228 | R1 non-interleaving property fuzzer | FR-093 | T-0227 | momus | `FuzzFR_093_BurstNonInterleaving` (fuzz) |
| T-0229 | Implement generic R2-TOTAL-ORDER-TIEBREAK rule | FR-092 | T-0225 | hephaestus | `TestFR_092_R2TotalOrderTiebreakDeterministic` (unit) |
| T-0230 | Apply R2 tiebreak to concurrent move-cycles | FR-094 | T-0229 | hephaestus | `TestFR_094_MoveCycleRetainsSmallerStateId` (unit) |
| T-0231 | Implement R3-DELETE-DOMINATES classification rule | FR-092 | T-0225 | hephaestus | `TestFR_092_DeleteDominatesConcurrentOps` (unit) |
| T-0232 | Exhaustive operation-kind-pair conformance test | FR-092 | T-0226, T-0227, T-0229, T-0231 | momus | `TestFR_092_ExhaustiveOperationKindPairClassification` (conformance) |
| T-0233 | Merge-time duplicate-identifier (CSPRNG collision) refusal | FR-024 | T-0225 | hephaestus | `TestFR_024_RefuseMergeOnDuplicateIdentifier` (unit) |
| T-0234 | Buffer or refuse a change with a missing causal predecessor | FR-095 | None | hephaestus | `TestFR_095_MissingCausalPredecessorBufferedOrRefused` (unit) |
| T-0235 | Refuse replay of erased content | FR-096 | None | hephaestus | `TestFR_096_RefuseReplayOfErasedUnit` (integration) |
| T-0236 | Refuse merge across a retention-point boundary | CON-024 | None | hephaestus | `TestCON_024_RefuseMergeAcrossRetentionPoint` (integration) |
| T-0237 | Refuse merge between mismatched history modes | CON-025 | None | hephaestus | `TestCON_025_RefuseMergeOnHistoryModeMismatch` (unit) |
| T-0238 | Record superseded state and resolution on non-ancestor commit | FR-116 | None | hephaestus | `TestFR_116_RecordsSupersededStateAndResolution` (unit) |
| T-0239 | Construct-level diff engine | TR-002 | None | hephaestus | `TestTR_002_DiffReportsConstructsNotStorageUnits` (unit) |
| T-0240 | Three-way merge conflict reporting with no auto-resolution | TR-003 | T-0229, T-0230, T-0231, T-0232 | hephaestus | `TestTR_003_ConflictNamesBothValuesNoAutoResolve` (unit) |
| T-0241 | End-to-end merge orchestrator integration test | TR-002, TR-003 | T-0233, T-0234, T-0235, T-0236, T-0237, T-0238, T-0239, T-0240 | hephaestus | `TestM13_MergeOrchestratorEndToEnd` (integration) |
| T-0364 | At-limit/over-limit wire-shape conformance vectors for merge guards | CON-024, CON-025 | T-0236, T-0237 | momus | `TestCON_024_CON_025_GuardsAtDeclaredWireLimit` (conformance) |

**T-0225** Operation-kind taxonomy and classifier skeleton

> Define the closed OperationKind enum (insert, delete, move, format-range-op, annotation-anchor-op, etc., matching data-model.md's concurrent-edit-relevant operation set) and the Classify(opA, opB) dispatch function signature in the merge package. The dispatch table must have no default/wildcard branch — every declared pair is resolved explicitly to one of {DISJOINT-COMMUTE, R1-SEQUENCE-ORDER, R2-TOTAL-ORDER-TIEBREAK, R3-DELETE-DOMINATES} by later tasks wiring into this skeleton.

- **Implements:** FR-092
- **Depends on:** None
- **DoD:** OperationKind enum and Classify() signature compile; the dispatch switch has zero default/wildcard case; package builds with `go build ./internal/merge/...`.
- **Test:** `TestFR_092_ClassifyDispatchHasNoWildcardCase` (unit)
- **Owner:** hephaestus

**T-0226** Implement DISJOINT-COMMUTE classification rule

> Implement the DISJOINT-COMMUTE branch: two concurrent operations whose target content-unit id sets are disjoint commute — applying them in either order yields the identical resulting T_C_root.

- **Implements:** FR-092
- **Depends on:** T-0225
- **DoD:** For a fixture pair of operations on two distinct unit ids, applying opA-then-opB and opB-then-opA produce byte-identical T_C_root values.
- **Test:** `TestFR_092_DisjointCommuteOrderIndependent` (unit)
- **Owner:** hephaestus

**T-0227** Implement R1-SEQUENCE-ORDER rule via Fugue

> Implement R1 using the Fugue algorithm's proven non-interleaving insert-ordering rule for concurrent inserts anchored at the same position, so that a contiguous authored burst by one actor is never split by an interleaved concurrent burst from another actor.

- **Implements:** FR-092, FR-093
- **Depends on:** T-0225
- **DoD:** Two concurrent single-actor bursts inserted at the same anchor merge deterministically per Fugue ordering; each burst appears as one unbroken substring in the merged run sequence.
- **Test:** `TestFR_093_ContiguousBurstSurvivesAsSubstring` (unit)
- **Owner:** hephaestus

**T-0228** R1 non-interleaving property fuzzer

> Property-based fuzz target generating randomized sets of concurrent authored bursts sharing anchors and merge-application orderings, asserting every burst remains an unbroken substring of the merged text under every generated ordering.

- **Implements:** FR-093
- **Depends on:** T-0227
- **DoD:** Fuzz corpus reaches >=10,000 randomized burst/anchor/ordering configurations with zero interleaving violations found; target registered in the CI fuzzing harness (CP-012 lane).
- **Test:** `FuzzFR_093_BurstNonInterleaving` (fuzz)
- **Owner:** momus

**T-0229** Implement generic R2-TOTAL-ORDER-TIEBREAK rule

> Implement the generic R2 branch: for a concurrent operation pair whose outcome depends on application order (e.g. two concurrent format-range writes on the same run), the operation whose committing state's state_id is lexicographically smaller wins, deterministically and independent of local processing order.

- **Implements:** FR-092
- **Depends on:** T-0225
- **DoD:** For a fixture pair with reversed state_id ordering fed to the classifier in both possible local-processing orders, the same operation wins both times.
- **Test:** `TestFR_092_R2TotalOrderTiebreakDeterministic` (unit)
- **Owner:** hephaestus

**T-0230** Apply R2 tiebreak to concurrent move-cycles

> Apply the same R2 state_id comparator from T-0229, uniformly, to a concurrent move-cycle (unit A moved into B's prior location and vice versa in the same window): retain the move committed under the lexicographically smaller state_id, and record the discarded move as a conflict entry rather than silently dropping it.

- **Implements:** FR-094
- **Depends on:** T-0229
- **DoD:** A synthesized 2-cycle and a synthesized 3-cycle move conflict both retain exactly the smaller-state_id move and each produces exactly one conflict record naming the discarded move(s).
- **Test:** `TestFR_094_MoveCycleRetainsSmallerStateId` (unit)
- **Owner:** hephaestus

**T-0231** Implement R3-DELETE-DOMINATES classification rule

> Implement R3: a concurrent delete of a unit dominates any other concurrent operation targeting that same unit (format-range write, move, annotation-anchor adjust). The delete's outcome always wins; the other operation's effect is discarded or redirected into orphan-carriage per DP-004 where the other operation was an annotation anchor.

- **Implements:** FR-092
- **Depends on:** T-0225
- **DoD:** For every other operation kind X in the taxonomy, the concurrent pair (delete, X) resolves to the delete's outcome with X's effect either discarded or orphan-carried, never partially applied.
- **Test:** `TestFR_092_DeleteDominatesConcurrentOps` (unit)
- **Owner:** hephaestus

**T-0232** Exhaustive operation-kind-pair conformance test

> plan.md Section 10 mandatory first-class task (b): a table-driven test enumerating the full N x N cross product of every operation kind in the T-0225 taxonomy against itself and every other kind, asserting Classify() resolves each pair to exactly one of {DISJOINT-COMMUTE, R1, R2, R3}. Any pair reaching an unclassified/4th case must fail the test loudly, not skip or TODO.

- **Implements:** FR-092
- **Depends on:** T-0226, T-0227, T-0229, T-0231
- **DoD:** Test iterates all NxN pairs from the taxonomy, asserts non-empty classification for each, zero pairs unclassified; test is registered as a required CI check (not marked skip/pending).
- **Test:** `TestFR_092_ExhaustiveOperationKindPairClassification` (conformance)
- **Owner:** momus

**T-0233** Merge-time duplicate-identifier (CSPRNG collision) refusal

> Before classification runs, scan both merge inputs for two distinct content units sharing one identifier — a CSPRNG collision, not an ordinary concurrent edit per FR-024's own distinction. Refuse the merge naming both unit locations; do not proceed to classification.

- **Implements:** FR-024
- **Depends on:** T-0225
- **DoD:** A fixture with two unrelated units forced to share one id is refused with an error naming both unit locations, and no classification call occurs for that pair; an unrelated collision-free fixture proceeds past this check.
- **Test:** `TestFR_024_RefuseMergeOnDuplicateIdentifier` (unit)
- **Owner:** hephaestus

**T-0234** Buffer or refuse a change with a missing causal predecessor

> When applying an incoming change (operation or HistorySegment batch) whose causal_predecessor_state_id / op-predecessor is not present locally, either buffer the change pending its predecessor's arrival or refuse it outright — pick and document one policy — naming the missing predecessor state_id in the buffered-pending list or the refusal error either way.

- **Implements:** FR-095
- **Depends on:** None
- **DoD:** An operation citing an unknown predecessor id is neither silently applied nor silently dropped: it appears in a pending-buffer list or a refusal error, in both cases naming the missing predecessor id explicitly.
- **Test:** `TestFR_095_MissingCausalPredecessorBufferedOrRefused` (unit)
- **Owner:** hephaestus

**T-0235** Refuse replay of erased content

> Reject any transmitted change whose application would re-derive content already covered by an ErasureRecord (data-model.md 2.17 invariant #2), naming the erased unit's identifier in the refusal.

- **Implements:** FR-096
- **Depends on:** None
- **DoD:** Replaying the original insert operation for a unit that has a corresponding ErasureRecord is refused, naming that unit's id; replaying an unrelated, non-erased operation against the same document succeeds.
- **Test:** `TestFR_096_RefuseReplayOfErasedUnit` (integration)
- **Owner:** hephaestus

**T-0236** Refuse merge across a retention-point boundary

> Refuse a merge when either input carries changes whose causal predecessor predates the other input's (possibly trimmed) retention_point, since the content needed to reconstruct their common ancestor may no longer be reconstructable.

- **Implements:** CON-024
- **Depends on:** None
- **DoD:** A fixture pairing a full-history branch with a trimmed branch whose retention_point falls after the full branch's divergence point is refused; a fixture where retention_point precedes divergence passes this check.
- **Test:** `TestCON_024_RefuseMergeAcrossRetentionPoint` (integration)
- **Owner:** hephaestus

**T-0237** Refuse merge between mismatched history modes

> Refuse a merge between two documents/branches whose Header.history-mode values differ, naming both declared modes in the refusal.

- **Implements:** CON-025
- **Depends on:** None
- **DoD:** Merging a COMPLETE_HISTORY-mode document with a RETAINED_FROM_POINT-mode document is refused with an error naming both modes; merging two same-mode documents passes this check.
- **Test:** `TestCON_025_RefuseMergeOnHistoryModeMismatch` (unit)
- **Owner:** hephaestus

**T-0238** Record superseded state and resolution on non-ancestor commit

> When a newly committed state's parent_state_id does not name the current head (a concurrent-commit case — the new state supersedes a state that is not its ancestor), record the superseded state's identifier plus a resolution value drawn from a closed enum (e.g. superseded-by-merge, rebased) in the commit metadata, rather than silently overwriting it.

- **Implements:** FR-116
- **Depends on:** None
- **DoD:** Committing over a non-ancestor head state produces a commit record carrying both the superseded state_id and a closed-enum resolution value; the superseded state remains reconstructable per its declared history mode.
- **Test:** `TestFR_116_RecordsSupersededStateAndResolution` (unit)
- **Owner:** hephaestus

**T-0239** Construct-level diff engine

> Implement diff(stateA, stateB) walking both states' content-unit sets and emitting differences expressed as document constructs (a changed Run, a moved unit, an added/removed Annotation, etc.), never as raw storage-segment or byte-offset deltas. Any construct both inputs agree on (identical content and identity) is omitted from the report.

- **Implements:** TR-002
- **Depends on:** None
- **DoD:** Diffing two states differing only in one annotation's boundary-behaviour field emits exactly one Annotation-level diff entry and zero entries for every unchanged Run or other construct.
- **Test:** `TestTR_002_DiffReportsConstructsNotStorageUnits` (unit)
- **Owner:** hephaestus

**T-0240** Three-way merge conflict reporting with no auto-resolution

> When classification yields a genuine conflict (an R2 case the comparator marks conflicting, or another rule's explicit conflict disposition per DP-015), the merge entry point returns a structured MergeConflict result — mapped by the future CLI layer (M18) to a non-zero exit — naming both competing values verbatim. Merge MUST NOT select one value, concatenate them, or interleave them silently.

- **Implements:** TR-003
- **Depends on:** T-0229, T-0230, T-0231, T-0232
- **DoD:** A fixture with two concurrent conflicting format-range edits on the same run produces a MergeConflict result whose payload contains both original values verbatim and no synthesized third value; no code path returns an auto-merged value for this fixture.
- **Test:** `TestTR_003_ConflictNamesBothValuesNoAutoResolve` (unit)
- **Owner:** hephaestus

**T-0241** End-to-end merge orchestrator integration test

> Wire the classification rules (T-0226, T-0227, T-0229, T-0231), refusal guards (T-0233..T-0237), supersession recording (T-0238), diff engine (T-0239), and conflict reporting (T-0240) into one merge.Merge(a, b, ancestor) entry point, and validate the milestone's exit criteria end-to-end on a realistic multi-construct document with a mix of disjoint, sequenced, tiebroken, delete-dominated, and genuinely conflicting concurrent edits in one scenario.

- **Implements:** TR-002, TR-003
- **Depends on:** T-0233, T-0234, T-0235, T-0236, T-0237, T-0238, T-0239, T-0240
- **DoD:** A single integration fixture exercising all four classification outcomes plus one genuine conflict in one merge call produces the expected per-construct result for the non-conflicting parts and a MergeConflict naming both values for the conflicting part, with no case falling through unclassified.
- **Test:** `TestM13_MergeOrchestratorEndToEnd` (integration)
- **Owner:** hephaestus

**T-0364** At-limit/over-limit wire-shape conformance vectors for merge guards

> Wire-format boundary conformance gap identified for this milestone: T-0236 and T-0237 each have fixtures for clearly-valid and clearly-invalid inputs but none pinned exactly at the documented boundary. Add at-limit fixtures (retention_point exactly at the divergence point; the maximal single history-mode value set consistently across both branches) that must pass, and one-unit-over-limit fixtures (retention_point one change past divergence; a single differing history-mode field) that must be refused with the same named-boundary error shape T-0236/T-0237 already produce. Note: the cross-cutting wire-format conformance concern referenced in review (tracked elsewhere as CON-010/CP-011) lists this merge milestone among its applicable milestones, but CON-010 is not one of this milestone's assigned requirement ids — this task closes the gap using the guard requirements this milestone actually owns (CON-024, CON-025) rather than claiming CON-010 outright; ownership of CON-010's milestone assignment should be confirmed against the concern registry rather than resolved here.

- **Implements:** CON-024, CON-025
- **Depends on:** T-0236, T-0237
- **DoD:** At-limit fixtures for both guards pass; over-limit-by-one fixtures for both guards are refused with the same named-boundary error shape as T-0236/T-0237; test registered as a required CI check.
- **Test:** `TestCON_024_CON_025_GuardsAtDeclaredWireLimit` (conformance)
- **Owner:** momus

---

### M14: Rendering & Resource

Font-less deterministic rendering needs validate (M07) for structural safety and integrity (M08) for presentation pinning, but not merge or history, so it proceeds in parallel with M12-M13.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0242 | PageDirectory entity keyed by content identity, not absolute page ordinal | NFR-010 | None | hephaestus | `TestNFR_010_PageDirectoryKeyedByContentIdentity` (unit) |
| T-0243 | Bind PageDirectory to a digest over its complete input set | FR-086 | T-0242 | hephaestus | `TestFR_086_PageDirectoryDigestBinding` (unit) |
| T-0244 | Mark PageDirectory non-normative and enforce staleness refusal/re-derivation | FR-087, FR-088 | T-0243 | hephaestus | `TestFR_088_StalePageDirectoryRefusedAndRederived` (integration) |
| T-0245 | FontRecord entity with exactly 7 declared fields | FR-089 | None | mnemosyne | `font_record_seven_fields` (conformance) |
| T-0246 | Font subset embedding + RegistryExcerpt wiring for durable-profile offline identity | NFR-024 | T-0245 | hephaestus | `TestNFR_024_DurableProfileRendersIdenticallyOffline` (integration) |
| T-0247 | Express authored fixed pagination directly from the file, no conversion step | FR-097 | T-0242 | hephaestus | `TestFR_097_FixedPaginationNoConversionStep` (integration) |
| T-0248 | Exact-rational active-edge rasterizer core | NFR-019 | None | hephaestus | `TestNFR_019_RasterizerExactRationalDeterminism` (unit) |
| T-0249 | Rasterizer excludes grid-fitting and font-instruction execution | NFR-023 | T-0248 | hephaestus | `TestNFR_023_NoGridFittingNoInstructionExecution` (unit) |
| T-0250 | Knuth-Plass reflow engine over in-document break/hyphenation table | FR-100, NFR-022 | None | hephaestus | `TestFR_100_ReflowDeterministicLineBreaks` (unit) |
| T-0251 | Reflow viewport support down to 320 reference pixels with no undeclared 2D scroll | FR-098 | T-0250 | hephaestus | `TestFR_098_ReflowNoScrollBelow320px` (integration) |
| T-0252 | Deterministic glyph shaping via pinned, versioned external oracle | NFR-021 | T-0245, T-0253 | hephaestus | `TestNFR_021_ShapingDeterministicGivenPinnedOracle` (unit) |
| T-0253 | Record the CON-006/CQ-006 shaping-oracle exception (EX-001) | CON-006 | None | clio | `TestCON_006_ShapingExceptionRecorded` (conformance) |
| T-0254 | Author PLP-1 exhaustive decode conformance vectors before the decoder exists | NFR-017 | None | momus | `plp1_decode_conformance_corpus_frozen` (conformance) |
| T-0255 | PLP-1 lossy codec decoder | NFR-017 | T-0254 | hephaestus | `TestPLP1_DecoderPassesConformanceCorpus` (conformance) |
| T-0256 | Restricted-PNG lossless profile decoder | NFR-017 | None | hephaestus | `TestRestrictedPNG_DecodeConformance` (conformance) |
| T-0257 | Materialised rendered representation for unimplementable embedded-object kinds | FR-090 | T-0255, T-0256 | hephaestus | `TestFR_090_UnimplementableObjectUsesMaterialisedRepresentation` (integration) |
| T-0258 | Placeholder mark for missing/failed embedded resource at recorded extent | FR-111 | T-0257 | hephaestus | `TestFR_111_MissingResourceRendersPlaceholder` (integration) |
| T-0259 | Machine-readable render report records resource substitution | FR-112 | T-0258 | hephaestus | `TestFR_112_RenderReportRecordsSubstitution` (integration) |
| T-0260 | Resource substitution marks resulting pagination non-authoritative | FR-113 | T-0258, T-0247 | hephaestus | `TestFR_113_SubstitutionMarksPaginationNonAuthoritative` (integration) |
| T-0261 | Resource substitution never uses a host resource and never triggers reflow | FR-114 | T-0258 | hephaestus | `TestFR_114_SubstitutionNoHostResourceNoReflow` (unit) |
| T-0262 | Open-and-render-any-page peak memory budget (<=200 MiB) | NFR-017 | T-0242, T-0247, T-0250 | hephaestus | `BenchmarkNFR_017_OpenAndRenderPagePeakMemory` (benchmark) |
| T-0263 | Per-page render octet-read budget (<=8 MiB + resources) | NFR-018 | T-0262 | hephaestus | `BenchmarkNFR_018_RenderPageOctetReadBudget` (benchmark) |
| T-0264 | At-limit/over-limit conformance fixtures for render-layer record shapes | FR-089, NFR-010 | T-0242, T-0245 | momus | `render_layer_at_limit_conformance_corpus` (conformance) |
| T-0265 | Continuous fuzzing harness for PLP-1 and restricted-PNG decode paths | NFR-017 | T-0255, T-0256 | prometheus | `FuzzPLP1Decode` (fuzz) |
| T-0266 | Isolate the shaping-oracle sub-slice so it cannot block the core render path (EX-001 blast-radius) | CON-006 | T-0248, T-0250, T-0255, T-0256, T-0247, T-0252 | hephaestus | `TestEX001_ShapingIsolationDoesNotBlockCoreRenderPath` (integration) |

**T-0242** PageDirectory entity keyed by content identity, not absolute page ordinal

> Implement the PageDirectory derived entity (data-model.md 2.22) as a rank-augmented structure over content-unit identity. A single-page content change must touch only that page's entry, never renumber the rest. Functionally depends on M07's validated content model and M04's run/unit identity, but no intra-milestone predecessor.

- **Implements:** NFR-010
- **Depends on:** None
- **DoD:** PageDirectory entries are keyed by content-unit identity; a table-driven test with fixed insert/remove-at-position fixtures (start, middle, end, single-entry corpus) shows only the affected entry's rank changes, all others retain their prior keys.
- **Test:** `TestNFR_010_PageDirectoryKeyedByContentIdentity` (unit)
- **Owner:** hephaestus

**T-0243** Bind PageDirectory to a digest over its complete input set

> Compute a digest over the complete declared input set (page geometry, referenced content-unit identities, font digests) that PageDirectory was derived from, following the same input-digest-binding pattern already used for PresentationArtefact (M09) and Frontmatter preview (M01).

- **Implements:** FR-086
- **Depends on:** T-0242
- **DoD:** Recomputing the digest over an unchanged input set yields an identical value; changing any one input (geometry, a referenced unit, a font digest) changes the digest.
- **Test:** `TestFR_086_PageDirectoryDigestBinding` (unit)
- **Owner:** hephaestus

**T-0244** Mark PageDirectory non-normative and enforce staleness refusal/re-derivation

> Tag PageDirectory as non-normative relative to source content (data-model.md Layer map convention) and implement the refusal path: on open, if the stored digest (T-0243) does not match a fresh recomputation, the stored PageDirectory is discarded and rebuilt from authoritative CONTENT segments before any page renders.

- **Implements:** FR-087, FR-088
- **Depends on:** T-0243
- **DoD:** A document with a hand-corrupted PageDirectory digest triggers full re-derivation before the first page render, and the render output is identical to a document that never had a stale directory.
- **Test:** `TestFR_088_StalePageDirectoryRefusedAndRederived` (integration)
- **Owner:** hephaestus

**T-0245** FontRecord entity with exactly 7 declared fields

> Implement FontRecord (data-model.md 2.21, document.abnf S7.4.1) carrying exactly: name, version, digest, variation coordinates, code-point set, layout-feature set, embedding permissions. No additional or missing field. Functionally depends on M01's container primitives.

- **Implements:** FR-089
- **Depends on:** None
- **DoD:** FontRecord round-trips through its wire form with exactly the 7 named fields present; a conformance fixture with a missing or extra field is rejected.
- **Test:** `font_record_seven_fields` (conformance)
- **Owner:** mnemosyne

**T-0246** Font subset embedding + RegistryExcerpt wiring for durable-profile offline identity

> Wire FontRecord-based subset embedding together with DP-017's RegistryExcerpt so a durable-profile document renders identically whether host fonts/network are available or not.

- **Implements:** NFR-024
- **Depends on:** T-0245
- **DoD:** The same durable-profile document rendered once with host fonts/network available and once with both unavailable produces byte-identical raster output for every page.
- **Test:** `TestNFR_024_DurableProfileRendersIdenticallyOffline` (integration)
- **Owner:** hephaestus

**T-0247** Express authored fixed pagination directly from the file, no conversion step

> Read Frontmatter.fm-page-count/fm-page-dimensions plus PageDirectory (T-0242) to produce the document's authored fixed pagination with zero intermediate transcoding or conversion pass.

- **Implements:** FR-097
- **Depends on:** T-0242
- **DoD:** Fixed pagination output is produced by a single read-and-project path with no intermediate document representation; a code-path trace shows zero conversion steps between file read and pagination output.
- **Test:** `TestFR_097_FixedPaginationNoConversionStep` (integration)
- **Owner:** hephaestus

**T-0248** Exact-rational active-edge rasterizer core

> Implement the reference 300 dpi rasterizer using only integer/rational arithmetic (de Casteljau curve flattening at 762 base units, round-half-to-even), per DP-010. Zero floating-point anywhere in the arithmetic path. Functionally depends on M08 for presentation-artefact pinning inputs.

- **Implements:** NFR-019
- **Depends on:** None
- **DoD:** Rasterizing the same input twice, and on two separate build/OS targets, produces bit-identical raster output; a static grep/lint check confirms no float32/float64 arithmetic in the rasterizer package.
- **Test:** `TestNFR_019_RasterizerExactRationalDeterminism` (unit)
- **Owner:** hephaestus

**T-0249** Rasterizer excludes grid-fitting and font-instruction execution

> Confirm and enforce that outline evaluation never applies grid-fitting and that no font-instruction-stream interpreter exists in the binary; fs-font-octets is read as opaque outline data only, per CQ-006/CP-005.

- **Implements:** NFR-023
- **Depends on:** T-0248
- **DoD:** A crafted font subset carrying an instruction stream renders identically to the same subset with the instruction stream stripped; static analysis confirms zero instruction-interpreter code paths.
- **Test:** `TestNFR_023_NoGridFittingNoInstructionExecution` (unit)
- **Owner:** hephaestus

**T-0250** Knuth-Plass reflow engine over in-document break/hyphenation table

> Implement the integer-demerits Knuth-Plass DP (DP-010) that reads line-break and hyphenation candidates exclusively from the in-document, digest-identified break/hyphenation table, never a host dictionary or locale service. Functionally depends on M04 content identity.

- **Implements:** FR-100, NFR-022
- **Depends on:** None
- **DoD:** Reflowing identical content at a fixed width twice (including across two separate processes) produces byte-identical line-break positions; a test with the host locale/dictionary swapped shows no change in output.
- **Test:** `TestFR_100_ReflowDeterministicLineBreaks` (unit)
- **Owner:** hephaestus

**T-0251** Reflow viewport support down to 320 reference pixels with no undeclared 2D scroll

> Extend the reflow engine to succeed at every viewport width >=320 reference pixels and confirm the produced layout needs no horizontal/vertical scrolling outside any declared 2D region. Note: FR-099's 2D-region declaration construct itself does not yet exist in the data model (M15 gap); this task treats 'no declared regions present' as the default case and does not block on that gap.

- **Implements:** FR-098
- **Depends on:** T-0250
- **DoD:** Reflow succeeds and requires zero out-of-region scrolling for every tested width in [320, 1920] reference pixels on a representative document corpus.
- **Test:** `TestFR_098_ReflowNoScrollBelow320px` (integration)
- **Owner:** hephaestus

**T-0252** Deterministic glyph shaping via pinned, versioned external oracle

> Implement glyph selection/positioning as a pure function of (font digest, variation axes, scalars, layout features, language) against the CQ-006 option-B pinned shaping oracle identified by Header.shaping-profile-id. BLOCKED: cannot close until T-0253's EX-001 exception is on record (plan.md Section 9 Conflict 2, CON-006 vs CQ-006).

- **Implements:** NFR-021
- **Depends on:** T-0245, T-0253
- **DoD:** Given a fixed input tuple and pinned oracle version, shaping output is byte-identical across repeated invocations and across two build targets; task cannot be marked done until EX-001 is approved.
- **Test:** `TestNFR_021_ShapingDeterministicGivenPinnedOracle` (unit)
- **Owner:** hephaestus

**T-0253** Record the CON-006/CQ-006 shaping-oracle exception (EX-001)

> Surface plan.md's self-disclosed Conflict 2 (CON-006's zero-named-reference rule vs. CQ-006 option B's pinned external shaping oracle) to Eyvar for an explicit, recorded, expiring exception (EX-001) scoped only to the shaping algorithm. Do not resolve the conflict unilaterally — flag and wait for ruling per this phase's constraints.

- **Implements:** CON-006
- **Depends on:** None
- **DoD:** clarify.md contains a committed EX-001 record with non-empty scope, expiry, approver, and date fields, verified by an automated fixture check against the committed file; absent that record, the check fails and NFR-021/T-0252 stay blocked.
- **Test:** `TestCON_006_ShapingExceptionRecorded` (conformance)
- **Owner:** clio

**T-0254** Author PLP-1 exhaustive decode conformance vectors before the decoder exists

> Plan.md Section 10 first-class task (c): author the complete PLP-1 decode conformance corpus (every syntactic decode case, including malformed/truncated inputs) before writing any PLP-1 decoder code. Pure spec/test-authoring task, feeds T-0255's NFR-017 decoder; it has no FR/NFR of its own beyond being the upstream artifact for that requirement.

- **Implements:** NFR-017
- **Depends on:** None
- **DoD:** A frozen, reviewed PLP-1 decode conformance corpus is committed to the repo with zero decoder implementation present yet.
- **Test:** `plp1_decode_conformance_corpus_frozen` (conformance)
- **Owner:** momus

**T-0255** PLP-1 lossy codec decoder

> Implement the PLP-1 decoder against the frozen corpus from T-0254, including the MAX_DECODED_UNIT pre-allocation ceiling check before any decode buffer is allocated, contributing to the NFR-017 page-render memory budget.

- **Implements:** NFR-017
- **Depends on:** T-0254
- **DoD:** Decoder passes 100% of T-0254's conformance corpus, and an over-ceiling input is rejected before any proportional memory allocation occurs.
- **Test:** `TestPLP1_DecoderPassesConformanceCorpus` (conformance)
- **Owner:** hephaestus

**T-0256** Restricted-PNG lossless profile decoder

> Implement the restricted-PNG lossless resource profile: accept only the allowed chunk/color-type subset, reject everything else, decode to an exact pixel-identical raster.

- **Implements:** NFR-017
- **Depends on:** None
- **DoD:** Decoder round-trips a golden restricted-PNG corpus bit-exact and rejects every disallowed chunk/color-type as a conformance fixture.
- **Test:** `TestRestrictedPNG_DecodeConformance` (conformance)
- **Owner:** hephaestus

**T-0257** Materialised rendered representation for unimplementable embedded-object kinds

> For an embedded object of a kind the renderer cannot natively decode, render from its stored materialised representation (PLP-1 or restricted-PNG) placed at exactly the object's declared authored extent, independent of the native resource format's own support.

- **Implements:** FR-090
- **Depends on:** T-0255, T-0256
- **DoD:** An object of a deliberately unimplementable kind renders using only its materialised representation, positioned at its declared extent, with no dependency on decoding the native format.
- **Test:** `TestFR_090_UnimplementableObjectUsesMaterialisedRepresentation` (integration)
- **Owner:** hephaestus

**T-0258** Placeholder mark for missing/failed embedded resource at recorded extent

> When a resource fails to load or decode, render a non-decorative placeholder mark at exactly the resource's recorded extent. The text-alternative content itself is a stub pending M15's FR-040 alt-text field, which does not yet exist in the data model — flagged, does not block this task.

- **Implements:** FR-111
- **Depends on:** T-0257
- **DoD:** A deliberately corrupted/missing resource reference renders a placeholder occupying exactly the recorded extent and is marked non-decorative in the render output.
- **Test:** `TestFR_111_MissingResourceRendersPlaceholder` (integration)
- **Owner:** hephaestus

**T-0259** Machine-readable render report records resource substitution

> Every resource-substitution event during a render run is appended to the render report in the JSON schema referenced by cli.md's render/verify reporting, naming the resource id, failure reason, and substitution kind.

- **Implements:** FR-112
- **Depends on:** T-0258
- **DoD:** A render with N induced resource failures produces a report listing exactly N substitution entries with resource id, reason, and kind populated; a clean render produces zero entries.
- **Test:** `TestFR_112_RenderReportRecordsSubstitution` (integration)
- **Owner:** hephaestus

**T-0260** Resource substitution marks resulting pagination non-authoritative

> Any render that performed at least one resource substitution (T-0258) must set the pagination-authoritative flag to false in the render output/report; a render with zero substitutions leaves fixed pagination (T-0247) authoritative.

- **Implements:** FR-113
- **Depends on:** T-0258, T-0247
- **DoD:** Pagination-authoritative flag is false whenever the render report (T-0259) contains >=1 substitution entry, and true otherwise, across a corpus of both cases.
- **Test:** `TestFR_113_SubstitutionMarksPaginationNonAuthoritative` (integration)
- **Owner:** hephaestus

**T-0261** Resource substitution never uses a host resource and never triggers reflow

> Prove the substitution path (T-0258) never performs a host font/resource lookup (filesystem, network, font directory) and never invokes the reflow engine (T-0250), preserving font-less/network-less rendering discipline (CP-005/CP-010).

- **Implements:** FR-114
- **Depends on:** T-0258
- **DoD:** A sandboxed test harness that fails on any host filesystem/network/font-directory call, or any reflow invocation, passes cleanly while exercising the substitution path.
- **Test:** `TestFR_114_SubstitutionNoHostResourceNoReflow` (unit)
- **Owner:** hephaestus

**T-0262** Open-and-render-any-page peak memory budget (<=200 MiB)

> Benchmark opening a 1 GiB/10,000-page document and rendering an arbitrary page against the NFR-017 peak resident memory ceiling of 209,715,200 octets, on the NFR-011 reference configuration once defined (M19).

- **Implements:** NFR-017
- **Depends on:** T-0242, T-0247, T-0250
- **DoD:** Benchmark harness reports peak RSS <=209,715,200 octets across a representative sample of pages (first, middle, last) of a synthetic 1 GiB/10,000-page corpus.
- **Test:** `BenchmarkNFR_017_OpenAndRenderPagePeakMemory` (benchmark)
- **Owner:** hephaestus

**T-0263** Per-page render octet-read budget (<=8 MiB + resources)

> Instrument I/O to confirm rendering page N of a 1 GiB document reads at most 8,388,608 octets from the ledger plus exactly that page's own referenced resource octets, independent of document size.

- **Implements:** NFR-018
- **Depends on:** T-0262
- **DoD:** Instrumented octet-read count for rendering any single page across a 1 MB, 100 MB, and 1 GiB document stays <=8,388,608 plus that page's resource octets in every case.
- **Test:** `BenchmarkNFR_018_RenderPageOctetReadBudget` (benchmark)
- **Owner:** hephaestus

**T-0264** At-limit/over-limit conformance fixtures for render-layer record shapes

> Cross-cutting CP-011/CON-010 obligation: ship at-limit and one-octet-past-limit conformance fixtures for every render-layer record shape introduced this milestone (PageDirectory entry [NFR-010], FontRecord [FR-089]), feeding the shared golden corpus before milestone exit.

- **Implements:** FR-089, NFR-010
- **Depends on:** T-0242, T-0245
- **DoD:** PageDirectory-entry and FontRecord fixtures exist at exactly their declared ceiling and one octet past it, both agreeing on verdict across two independent decode calls.
- **Test:** `render_layer_at_limit_conformance_corpus` (conformance)
- **Owner:** momus

**T-0265** Continuous fuzzing harness for PLP-1 and restricted-PNG decode paths

> CP-012 obligation: register the PLP-1 (T-0255) and restricted-PNG (T-0256) decoders, both of which implement NFR-017, as continuous fuzz targets since both decode untrusted external bytes, feeding M19's harness-maturity check.

- **Implements:** NFR-017
- **Depends on:** T-0255, T-0256
- **DoD:** Both fuzz targets run in CI's continuous fuzzing job with zero panics/OOM/unbounded-allocation crashes over a 24-hour soak on the current corpus.
- **Test:** `FuzzPLP1Decode` (fuzz)
- **Owner:** prometheus

**T-0266** Isolate the shaping-oracle sub-slice so it cannot block the core render path (EX-001 blast-radius)

> Per plan.md Section 9's disclosed blast-radius note: EX-001 (T-0253, CON-006 exception) gates only the shaping sub-slice, not the rasterizer, PLP-1/PNG decode, reflow, or fixed pagination. Structure the render pipeline so glyph shaping (T-0252) sits behind an interface the rest of the pipeline does not depend on, so those other pieces ship and test green even while T-0252/T-0253 remain blocked on Eyvar's EX-001 ruling.

- **Implements:** CON-006
- **Depends on:** T-0248, T-0250, T-0255, T-0256, T-0247, T-0252
- **DoD:** Building and running the full test suite for rasterizer, PLP-1/PNG decode, reflow, and fixed pagination succeeds with the shaping-oracle interface stubbed/unimplemented.
- **Test:** `TestEX001_ShapingIsolationDoesNotBlockCoreRenderPath` (integration)
- **Owner:** hephaestus

---

### M15: Accessibility & Semantic Content-Model Extensions

16 of 18 spec-named validator rules (direction, reading order, role mapping, table scope, alt-text, numbering, xref presentation, inferred markers) have zero data-model representation, so these fields must be added to the content model before validate/render/extract can claim full coverage of them.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0267 | Obtain Eyvar/themis design ruling on the 8 missing accessibility/semantic mechanism fields |  | None | clio | `ConformanceCase_M15_DesignRulingRecorded` (conformance) |
| T-0268 | Add `direction` field to data-model.md text-container entities | FR-032 | T-0267 | mnemosyne | `TestFR_032_DirectionFieldInDataModel` (unit) |
| T-0269 | Add `direction` field to document.abnf grammar and wire encoding | FR-032 | T-0268 | hephaestus | `CONFORMANCE-BIDI-000-direction-roundtrip` (conformance) |
| T-0270 | Implement PD-BIDI-001 mandatory-presence check for `direction` | FR-032 | T-0269 | hephaestus | `CONFORMANCE-BIDI-001-missing-direction-rejected` (conformance) |
| T-0271 | Extend PD-BIDI-001 to bar in-band directional control scalars | FR-033 | T-0270 | hephaestus | `CONFORMANCE-BIDI-002-inband-control-rejected` (conformance) |
| T-0272 | Define computed-inline isolation wrapper construct | FR-034 | T-0271 | mnemosyne | `TestFR_034_ComputedInlineIsolation` (conformance) |
| T-0273 | Define ROOT_SEQUENCE authored reading-order record | FR-036 | T-0267 | mnemosyne | `TestFR_036_RootSequenceSchema` (unit) |
| T-0274 | Add ROOT_SEQUENCE to document.abnf plus completeness validator | FR-036 | T-0273 | hephaestus | `CONFORMANCE-A11Y-005-rootsequence-completeness` (conformance) |
| T-0275 | Retrofit T_C subtree traversal order to use ROOT_SEQUENCE | FR-036 | T-0274, T-0134, T-0135, T-0149, T-0150 | argus | `TestFR_036_TCTraversalUsesRootSequence` (integration) |
| T-0276 | Retrofit `extract` package walk order to use ROOT_SEQUENCE | FR-036, FR-101 | T-0274, T-0087 | hephaestus | `TestFR_036_ExtractWalksRootSequenceOrder` (integration) |
| T-0277 | Add reading-order-or-decoration field to every renderable mark | FR-037 | T-0274 | hephaestus | `CONFORMANCE-A11Y-001-mark-placement` (conformance) |
| T-0278 | Conformance test: fixed and reflowed presentations agree on text and reading order | FR-101 | T-0275, T-0276, T-0277 | momus | `CONFORMANCE-A11Y-006-fixed-reflow-parity` (conformance) |
| T-0279 | Author closed accessibility-role mapping table | FR-038 | T-0267 | clio | `TestFR_038_RoleMapCompleteness` (unit) |
| T-0280 | Implement PD-A11Y-002 role-resolution validator | FR-038 | T-0279 | hephaestus | `CONFORMANCE-A11Y-002-role-resolution` (conformance) |
| T-0281 | Add header-cell `scope` field to TABLE record | FR-039 | T-0267 | mnemosyne | `CONFORMANCE-TBL-000-scope-roundtrip` (conformance) |
| T-0282 | Implement PD-TBL-001a header-scope validator | FR-039 | T-0281 | hephaestus | `CONFORMANCE-TBL-001-header-scope` (conformance) |
| T-0283 | Implement PD-TBL-001b cell-tiling invariant validator | FR-082 | T-0281 | hephaestus | `CONFORMANCE-TBL-002-cell-tiling` (conformance) |
| T-0284 | Add mandatory alt-text field to non-decorative object records | FR-040 | T-0267 | mnemosyne | `CONFORMANCE-A11Y-000-alttext-roundtrip` (conformance) |
| T-0285 | Implement PD-A11Y-003 alt-text validator | FR-040 | T-0284 | hephaestus | `CONFORMANCE-A11Y-003-alttext-quality` (conformance) |
| T-0286 | Define ordered-sequence numbering construct | FR-083 | T-0272, T-0276 | mnemosyne | `TestFR_083_NumberingLabelDeterministic` (unit) |
| T-0287 | Implement PD-A11Y-004 no-persisted-literal-label validator | FR-083 | T-0286 | hephaestus | `CONFORMANCE-A11Y-004-numbering-literal-rejected` (conformance) |
| T-0288 | Add `presentation-function` field to CROSS_REFERENCE | FR-084 | T-0272 | mnemosyne | `CONFORMANCE-XREF-000-presentation-function-roundtrip` (conformance) |
| T-0289 | Add staleness digest-binding field to CROSS_REFERENCE | FR-085 | T-0288 | hephaestus | `CONFORMANCE-XREF-001-staleness-without-layout` (conformance) |
| T-0290 | Define 2D-presentation-region record | FR-099 | T-0274, T-0284 | mnemosyne | `CONFORMANCE-2D-000-region-roundtrip` (conformance) |
| T-0291 | Implement PD-2D-001 validator | FR-099 | T-0290 | hephaestus | `CONFORMANCE-2D-001-region-fields` (conformance) |
| T-0292 | Add inferred-value marker + basis field to mandated-but-inferable fields | FR-118 | T-0267 | mnemosyne | `TestFR_118_InferredMarkerEnumeration` (unit) |
| T-0293 | Implement PD-INFER-001 validator | FR-118 | T-0292 | hephaestus | `CONFORMANCE-INFER-001-marker-consistency` (conformance) |
| T-0294 | Register all M15-introduced PD-rule ids in contracts/README.md's CI registry | NFR-031 | T-0270, T-0271, T-0272, T-0275, T-0277, T-0280, T-0282, T-0283, T-0285, T-0287, T-0289, T-0291, T-0293 | prometheus | `TestNFR_031_RuleRegistryComplete` (conformance) |
| T-0295 | Conformance suite: >=80% of accessibility failure conditions are software-decidable | NFR-031 | T-0294 | momus | `TestNFR_031_EightyPercentDecidable` (conformance) |

**T-0267** Obtain Eyvar/themis design ruling on the 8 missing accessibility/semantic mechanism fields

> spec.md's FR-032/033/034/036/037/038/039/040/082/083/084/085/099/118/NFR-031 are frozen and correct as WHAT/WHY — the gap is that plan.md/data-model.md/document.abnf (the HOW) never added a mechanism for 16 of 18 spec-named validator rules. Per SDD, this is a plan/data-model amendment, not a requirement reopening: draft a short design-ruling memo enumerating each missing construct (direction field, ROOT_SEQUENCE reading-order record, computed-inline isolation wrapper, accessibility role-mapping table, table header scope + tiling invariant, alt-text field, ordered-sequence numbering construct, xref presentation-function + staleness field, 2D-region record, inferred-value marker) with the proposed field shapes, and get Eyvar's explicit sign-off before any of T-0268..T-0293 touch data-model.md/document.abnf. This also surfaces to Eyvar that M15 depends on M04 alone per the spine, but T-0275/T-0276 below in practice retrofit M08 (T_C traversal) and M05 (extract walk order) — a spine dependency-edge gap worth flagging alongside the ruling request. NOTE: this task is intentionally requirement-less in `implements` — it is a cross-cutting SDD phase-gate governance step (CP-011/AD-002 compliance: no plan/data-model amendment proceeds without recorded sign-off), not an implementation of any single FR/NFR; the 16 requirements it gates are named above and in the tasks it unblocks.

- **Implements:** None (see description)
- **Depends on:** None
- **DoD:** A dated addendum is appended to clarify.md (or a new clarify-002.md) listing all 8 construct groups with Eyvar's written approval recorded per construct; zero remain unresolved.
- **Test:** `ConformanceCase_M15_DesignRulingRecorded` (conformance)
- **Owner:** clio

**T-0268** Add `direction` field to data-model.md text-container entities

> Add a mandatory base-writing-direction field (closed enum, e.g. LTR/RTL) to every text-container entity in data-model.md (TextBlock, Table, top-level document unit) per the T-0267 ruling.

- **Implements:** FR-032
- **Depends on:** T-0267
- **DoD:** Every text-container entity's field table in data-model.md lists a `direction` field with a closed value set and 'mandatory' constraint.
- **Test:** `TestFR_032_DirectionFieldInDataModel` (unit)
- **Owner:** mnemosyne

**T-0269** Add `direction` field to document.abnf grammar and wire encoding

> Mirror T-0268's data-model.md field into document.abnf's TextBlock/Table record productions as a PDL-TLV field, assign it a field-kind tag, and update pdlfmt's generated struct.

- **Implements:** FR-032
- **Depends on:** T-0268
- **DoD:** document.abnf's TextBlock and Table productions include the direction field with an ABNF rule for its closed value set; a round-trip encode/decode of a fixture preserves the value byte-exact.
- **Test:** `CONFORMANCE-BIDI-000-direction-roundtrip` (conformance)
- **Owner:** hephaestus

**T-0270** Implement PD-BIDI-001 mandatory-presence check for `direction`

> Add validator rule PD-BIDI-001 rejecting any text-container record whose direction field is absent, naming the offending unit id and octet offset per FR-102.

- **Implements:** FR-032
- **Depends on:** T-0269
- **DoD:** A fixture with direction omitted is rejected with rule id PD-BIDI-001 and the correct unit id/offset; a fixture with it present passes.
- **Test:** `CONFORMANCE-BIDI-001-missing-direction-rejected` (conformance)
- **Owner:** hephaestus

**T-0271** Extend PD-BIDI-001 to bar in-band directional control scalars

> Add the structural-only-direction half of PD-BIDI-001: reject any run-text scalar sequence containing a bidi control character (LRE/RLE/PDF/LRO/RLO/LRI/RLI/FSI/PDI etc.), coordinating with M07's CON-004/PD-SCALAR-001 excluded-scalar set so the two rules do not double-report the same octet.

- **Implements:** FR-033
- **Depends on:** T-0270
- **DoD:** A run containing an LRO/RLO control scalar is rejected citing PD-BIDI-001 exactly once, not duplicated by PD-SCALAR-001.
- **Test:** `CONFORMANCE-BIDI-002-inband-control-rejected` (conformance)
- **Owner:** hephaestus

**T-0272** Define computed-inline isolation wrapper construct

> Define a generic 'computed-inline' field/wrapper (isolation marker) applicable to any inline object whose text value is computed rather than authored (e.g. a rendered cross-reference or numbering label), guaranteeing its directionality cannot leak into surrounding paragraph reordering. This is the shared primitive that T-0286 (numbering) and T-0288 (xref) will embed.

- **Implements:** FR-034
- **Depends on:** T-0271
- **DoD:** document.abnf defines a COMPUTED_INLINE wrapper field-kind with an isolation flag that a bidi-reordering test fixture proves does not affect neighbouring run ordering.
- **Test:** `TestFR_034_ComputedInlineIsolation` (conformance)
- **Owner:** mnemosyne

**T-0273** Define ROOT_SEQUENCE authored reading-order record

> Add a top-level 'document body sequence' record to data-model.md establishing the authored logical top-to-bottom reading order over content units, resolving the self-disclosed absence integrity.abnf S2.2.1 flags for T_C's own traversal and the plan.md Section 3 Extraction storage-order/reading-order inconsistency.

- **Implements:** FR-036
- **Depends on:** T-0267
- **DoD:** data-model.md defines ROOT_SEQUENCE with an ordered list of unit references, one entry per content unit, and states it is the authoritative reading order independent of storage/append order.
- **Test:** `TestFR_036_RootSequenceSchema` (unit)
- **Owner:** mnemosyne

**T-0274** Add ROOT_SEQUENCE to document.abnf plus completeness validator

> Add the ROOT_SEQUENCE record production to document.abnf and a validator rule requiring every content unit appear in it exactly once (no duplicates, no omissions).

- **Implements:** FR-036
- **Depends on:** T-0273
- **DoD:** A fixture omitting one unit from ROOT_SEQUENCE, and one listing a unit twice, are both rejected naming the unit id; a complete fixture passes.
- **Test:** `CONFORMANCE-A11Y-005-rootsequence-completeness` (conformance)
- **Owner:** hephaestus

**T-0275** Retrofit T_C subtree traversal order to use ROOT_SEQUENCE

> Replace the interim CSPRNG-ordered T_C subtree traversal (integrity.abnf S2.2.1's self-flagged placeholder, built in M08) with a traversal keyed on ROOT_SEQUENCE order. This is a retrofit of already-built M08 integrity-tree code — added explicit deps on the M08 identity/builder/traversal/root tasks (T-0134, T-0135, T-0149, T-0150) since M15's spine lists only M04, which does not carry the T_C code this task modifies. Those ids come from cross-milestone findings and are not present in the M04-only dependency summary handed to this pass; confirm the exact M08 task ids at the analyze/zeus stage and add M08 as an explicit spine dependency for M15.

- **Implements:** FR-036
- **Depends on:** T-0274, T-0134, T-0135, T-0149, T-0150
- **DoD:** T_C_root recomputation over a fixture with reordered ROOT_SEQUENCE (but unchanged content) changes only if ROOT_SEQUENCE itself changed; traversal order is documented as ROOT_SEQUENCE order, not CSPRNG order.
- **Test:** `TestFR_036_TCTraversalUsesRootSequence` (integration)
- **Owner:** argus

**T-0276** Retrofit `extract` package walk order to use ROOT_SEQUENCE

> Change the `extract` package's CONTENT-segment walk (built in M05) from raw storage order to ROOT_SEQUENCE order, so extraction text is emitted in reading order even after a structural edit inserted a unit mid-document (closing the plan.md Section 3 gap this milestone's inventory identified). Added explicit dep on T-0087 (M05 extract walker) since M15's spine lists only M04, which does not carry the extract-walk code this task modifies; T-0087 is not present in the M04-only dependency summary handed to this pass — confirm the exact M05 task id at the analyze/zeus stage and add M05 as an explicit spine dependency for M15.

- **Implements:** FR-036, FR-101
- **Depends on:** T-0274, T-0087
- **DoD:** Extracting a fixture where a unit was inserted logically mid-document (and therefore appended at the end of storage order) yields text in ROOT_SEQUENCE order, not storage order.
- **Test:** `TestFR_036_ExtractWalksRootSequenceOrder` (integration)
- **Owner:** hephaestus

**T-0277** Add reading-order-or-decoration field to every renderable mark

> Add a field to every renderable-mark record requiring it to carry either a ROOT_SEQUENCE position reference or an explicit 'decoration' flag, plus validator rule PD-A11Y-001 rejecting a mark with neither.

- **Implements:** FR-037
- **Depends on:** T-0274
- **DoD:** A mark fixture with neither a reading-order ref nor a decoration flag is rejected as PD-A11Y-001; fixtures with either pass.
- **Test:** `CONFORMANCE-A11Y-001-mark-placement` (conformance)
- **Owner:** hephaestus

**T-0278** Conformance test: fixed and reflowed presentations agree on text and reading order

> Author a conformance suite that renders one state under both fixed pagination and Knuth-Plass reflow and asserts the extracted text and ROOT_SEQUENCE-derived reading order are identical between the two presentations.

- **Implements:** FR-101
- **Depends on:** T-0275, T-0276, T-0277
- **DoD:** The suite passes on at least 3 fixtures spanning single-page, multi-page, and mid-document-inserted-unit cases.
- **Test:** `CONFORMANCE-A11Y-006-fixed-reflow-parity` (conformance)
- **Owner:** momus

**T-0279** Author closed accessibility-role mapping table

> Publish a closed accessibility-role enum and a table mapping every defined construct kind in document.abnf onto exactly one role, in data-model.md.

- **Implements:** FR-038
- **Depends on:** T-0267
- **DoD:** Every construct kind named in document.abnf appears in the mapping table exactly once, with no construct mapped to zero or >1 roles.
- **Test:** `TestFR_038_RoleMapCompleteness` (unit)
- **Owner:** clio

**T-0280** Implement PD-A11Y-002 role-resolution validator

> Add validator rule PD-A11Y-002 asserting every construct instance in a document resolves to exactly one role from T-0279's table; an unmapped or newly-added construct kind fails closed.

- **Implements:** FR-038
- **Depends on:** T-0279
- **DoD:** A fixture containing every construct kind passes with one role each; a synthetic unmapped construct kind is rejected as PD-A11Y-002.
- **Test:** `CONFORMANCE-A11Y-002-role-resolution` (conformance)
- **Owner:** hephaestus

**T-0281** Add header-cell `scope` field to TABLE record

> Add a scope field (declaring which data cells a header cell heads: row/column/row-group/column-group) to TABLE's header-cell shape in data-model.md and document.abnf S4.

- **Implements:** FR-039
- **Depends on:** T-0267
- **DoD:** TABLE's header-cell production carries a mandatory `scope` field with a closed enum; a round-trip fixture preserves it byte-exact.
- **Test:** `CONFORMANCE-TBL-000-scope-roundtrip` (conformance)
- **Owner:** mnemosyne

**T-0282** Implement PD-TBL-001a header-scope validator

> Add validator rule PD-TBL-001a requiring every header cell declare a scope and that the scope resolve to at least one real data cell in the table's grid.

- **Implements:** FR-039
- **Depends on:** T-0281
- **DoD:** A header cell with missing scope, and one whose scope resolves to zero cells, are both rejected as PD-TBL-001a.
- **Test:** `CONFORMANCE-TBL-001-header-scope` (conformance)
- **Owner:** hephaestus

**T-0283** Implement PD-TBL-001b cell-tiling invariant validator

> Add validator rule PD-TBL-001b asserting a table's cells tile its declared grid exactly once: no two cells overlap, no grid position is uncovered.

- **Implements:** FR-082
- **Depends on:** T-0281
- **DoD:** A fixture with an overlapping cell pair and one with an uncovered grid position are both rejected as PD-TBL-001b; a correctly-tiled fixture passes.
- **Test:** `CONFORMANCE-TBL-002-cell-tiling` (conformance)
- **Owner:** hephaestus

**T-0284** Add mandatory alt-text field to non-decorative object records

> Add a non-empty text-alternative field to RASTER_IMAGE, FONT_SUBSET, and every other non-decorative embedded-object record in data-model.md/document.abnf S7.

- **Implements:** FR-040
- **Depends on:** T-0267
- **DoD:** Every non-decorative object record's field table includes a mandatory `alt-text` field; a round-trip fixture preserves a non-ASCII alt-text value byte-exact.
- **Test:** `CONFORMANCE-A11Y-000-alttext-roundtrip` (conformance)
- **Owner:** mnemosyne

**T-0285** Implement PD-A11Y-003 alt-text validator

> Add validator rule PD-A11Y-003 rejecting a non-decorative object whose alt-text is absent, empty, or merely echoes its filename/digest/dimensions.

- **Implements:** FR-040
- **Depends on:** T-0284
- **DoD:** Fixtures with missing, empty, and filename-echoing alt-text are all rejected as PD-A11Y-003; a genuine description passes.
- **Test:** `CONFORMANCE-A11Y-003-alttext-quality` (conformance)
- **Owner:** hephaestus

**T-0286** Define ordered-sequence numbering construct

> Define a SEQUENCE_DEFINITION record plus a per-item order-value field, such that a rendered numbering label is a pure deterministic function of (order-value, referenced sequence definition, ROOT_SEQUENCE position) with no persisted literal label ever authoritative. Rendered labels are carried as COMPUTED_INLINE (T-0272) values.

- **Implements:** FR-083
- **Depends on:** T-0272, T-0276
- **DoD:** SEQUENCE_DEFINITION and per-item order-value fields are defined; a fixture computes an identical numbering label across two independent runs of the label-derivation function.
- **Test:** `TestFR_083_NumberingLabelDeterministic` (unit)
- **Owner:** mnemosyne

**T-0287** Implement PD-A11Y-004 no-persisted-literal-label validator

> Add validator rule PD-A11Y-004 rejecting any ordered-sequence item that carries a persisted literal numbering label not derivable from its SEQUENCE_DEFINITION and order-value.

- **Implements:** FR-083
- **Depends on:** T-0286
- **DoD:** A fixture with a stray literal label field on a numbered item is rejected as PD-A11Y-004.
- **Test:** `CONFORMANCE-A11Y-004-numbering-literal-rejected` (conformance)
- **Owner:** hephaestus

**T-0288** Add `presentation-function` field to CROSS_REFERENCE

> Add a presentation-function field to CROSS_REFERENCE (document.abnf S5) naming how the reference renders (page-number/citation/literal-frozen-text/etc.), carried as a COMPUTED_INLINE value per T-0272, replacing any reliance on literal frozen text.

- **Implements:** FR-084
- **Depends on:** T-0272
- **DoD:** CROSS_REFERENCE's production includes a mandatory presentation-function field with a closed enum; a round-trip fixture preserves it.
- **Test:** `CONFORMANCE-XREF-000-presentation-function-roundtrip` (conformance)
- **Owner:** mnemosyne

**T-0289** Add staleness digest-binding field to CROSS_REFERENCE

> Add a digest-binding field to CROSS_REFERENCE, bound to its target unit's current state the way PresentationArtefact/Frontmatter-preview/PageDirectory/UnitIndex already bind to their inputs, plus a validator that reports staleness from the digest mismatch alone, without computing layout.

- **Implements:** FR-085
- **Depends on:** T-0288
- **DoD:** A fixture where the xref target changed after the xref was authored is reported stale purely from digest comparison, with no layout pass invoked.
- **Test:** `CONFORMANCE-XREF-001-staleness-without-layout` (conformance)
- **Owner:** hephaestus

**T-0290** Define 2D-presentation-region record

> Define a 2D-presentation-region record carrying a mandatory linearised-reading-order reference (into ROOT_SEQUENCE) and a mandatory text-alternative field (reusing T-0284's alt-text field shape), for content that cannot be expressed as pure linear flow.

- **Implements:** FR-099
- **Depends on:** T-0274, T-0284
- **DoD:** The 2D-region record's field table includes both mandatory fields; a round-trip fixture preserves both.
- **Test:** `CONFORMANCE-2D-000-region-roundtrip` (conformance)
- **Owner:** mnemosyne

**T-0291** Implement PD-2D-001 validator

> Add validator rule PD-2D-001 rejecting a 2D-presentation-region missing its linearised-reading-order reference, missing its text-alternative, or whose reading-order reference does not resolve into ROOT_SEQUENCE.

- **Implements:** FR-099
- **Depends on:** T-0290
- **DoD:** Fixtures missing each of the two mandatory fields, and one with a dangling reading-order reference, are all rejected as PD-2D-001.
- **Test:** `CONFORMANCE-2D-001-region-fields` (conformance)
- **Owner:** hephaestus

**T-0292** Add inferred-value marker + basis field to mandated-but-inferable fields

> Enumerate every mandated field in data-model.md capable of being system-inferred rather than authored (e.g. an inferred alt-text, an inferred direction) and add an inferred-marker flag plus a basis-of-inference reference to each.

- **Implements:** FR-118
- **Depends on:** T-0267
- **DoD:** Every field identified in the enumeration carries both the inferred-marker flag and a basis-of-inference field in its data-model.md entry.
- **Test:** `TestFR_118_InferredMarkerEnumeration` (unit)
- **Owner:** mnemosyne

**T-0293** Implement PD-INFER-001 validator

> Add validator rule PD-INFER-001 requiring the inferred-marker+basis pair be present whenever a value in one of T-0292's fields was not explicitly authored, and absent when it was.

- **Implements:** FR-118
- **Depends on:** T-0292
- **DoD:** A fixture with a system-supplied value lacking the marker is rejected as PD-INFER-001; an authored value carrying a spurious inferred-marker is also rejected.
- **Test:** `CONFORMANCE-INFER-001-marker-consistency` (conformance)
- **Owner:** hephaestus

**T-0294** Register all M15-introduced PD-rule ids in contracts/README.md's CI registry

> Add PD-BIDI-001, PD-A11Y-001..004, PD-TBL-001a/b, PD-2D-001, and PD-INFER-001 to contracts/README.md's generated rule-id-to-conformance-case registry with the CI equality check against spec.md text per CP-011/NFR-029, closing the pattern where 16 of 18 spec-named rule ids were previously never carried downstream.

- **Implements:** NFR-031
- **Depends on:** T-0270, T-0271, T-0272, T-0275, T-0277, T-0280, T-0282, T-0283, T-0285, T-0287, T-0289, T-0291, T-0293
- **DoD:** CI's registry-equality check passes with all 9 new rule ids present, each mapped to at least one conformance case id from T-0268..T-0293.
- **Test:** `TestNFR_031_RuleRegistryComplete` (conformance)
- **Owner:** prometheus

**T-0295** Conformance suite: >=80% of accessibility failure conditions are software-decidable

> Enumerate every published accessibility structural+presence failure condition (spec.md's accessibility-relevant FR list) and, for each, determine whether a PD-rule from T-0268..T-0293 decides it purely structurally without human judgment; assert the decidable fraction is >=80% per NFR-031.

- **Implements:** NFR-031
- **Depends on:** T-0294
- **DoD:** A decidability-classification rubric (documented in contracts/README.md) classifies each enumerated accessibility failure condition as 'structurally decidable' only when a specific, named PD-rule id fully determines pass/fail from document structure alone with no human judgment required; applying that rubric across every enumerated condition yields a computed decidable-fraction of at least 0.80, and the computation is wired into CI so a future regression below 0.80 fails the build.
- **Test:** `TestNFR_031_EightyPercentDecidable` (conformance)
- **Owner:** momus

---

### M16: Evolution / Migration & Registry

Refusal-first major-version migration and RESCIND-AND-RESIGN need a working validator and a working signature implementation to retain across the migration boundary.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0296 | Implement Phase 1 refusal-first construct-enumeration scan | FR-121 | None | hephaestus | `TestFR_121_RefusalHaltsAtFirstUnrepresentableConstruct` (unit) |
| T-0297 | Author unrepresentable-construct conformance vectors for Phase 1 refusal | FR-121 | T-0296 | momus | `TestFR_121_ConformanceVectors_UnrepresentableConstructs` (conformance) |
| T-0298 | Implement Phase 2 deterministic canonical-octet migration transform | FR-119 | T-0296 | hephaestus | `TestFR_119_TransformDeterministicAcrossInvocations` (unit) |
| T-0299 | Publish external-trial corpus for cross-implementation migration determinism | FR-119 | T-0298 | momus | `TestFR_119_ExternalTrialCorpusPublished` (external-trial) |
| T-0300 | Preserve content-unit identifiers unchanged through migration | FR-120 | T-0298 | hephaestus | `TestFR_120_IdentifiersSurviveMigration` (fuzz) |
| T-0301 | Enforce unsupported-major-version decline gate | FR-123 | None | hephaestus | `TestFR_123_UnsupportedMajorVersionDeclinedFirst` (conformance) |
| T-0302 | Compose post-migration signature verdict with pre-migration coverage | FR-122 | T-0298, T-0300 | argus | `TestFR_122_PreMigrationSignatureCoversOriginalState` (integration) |
| T-0303 | Implement hash-family re-protection layer (SHA3-256 outer digest) | CON-016 | None | argus | `TestCON_016_ReProtectionPreservesOriginalVerdict` (integration) |
| T-0304 | Implement RESCIND-AND-RESIGN scheme-level break flow | CON-016 | T-0302, T-0303 | argus | `TestCON_016_RescindAndResignAtMajorVersionBoundary` (integration) |
| T-0305 | Ship wire-format conformance vectors for RescindResignRecord | CON-016 | T-0304 | momus | `TestCON_016_RescindResignRecordConformanceVectors` (conformance) |
| T-0306 | argus security review of migration and re-protection flows | CON-016 | T-0298, T-0302, T-0303, T-0304 | argus | `TestCON_016_ArgusReviewNoP0P1Findings` (integration) |

**T-0296** Implement Phase 1 refusal-first construct-enumeration scan

> Build migrate.Scan(source, targetMajor) that walks the source document in a defined, deterministic traversal order (reusing validate's error-precedence ordering pattern from FR-103) and halts at the FIRST construct not representable in targetMajor's construct set, returning a RefusalReport naming the construct kind and its exact location (segment id + octet offset + unit id where applicable). No mutation is performed and no approximation/partial migration is attempted. Conceptually depends on M07 (validate's structural walk) and M08 (T_C-based location naming) being complete, though no intra-milestone task dependency exists.

- **Implements:** FR-121
- **Depends on:** None
- **DoD:** Scan() against a fixture with zero unrepresentable constructs returns an empty RefusalReport (proceed signal); a fixture with exactly one unrepresentable construct returns a RefusalReport naming that construct's exact kind and location; a fixture with two distinct unrepresentable constructs in traversal order returns a report naming only the first one encountered, never both.
- **Test:** `TestFR_121_RefusalHaltsAtFirstUnrepresentableConstruct` (unit)
- **Owner:** hephaestus

**T-0297** Author unrepresentable-construct conformance vectors for Phase 1 refusal

> Ship a golden corpus of source documents, each containing exactly one class of unrepresentable construct (a retired extension token, an out-of-allowlist crypto parameter, a construct exceeding the target major version's structural ceiling, etc.), each paired with its expected RefusalReport (construct kind + exact octet offset + unit id). Satisfies CON-010's at-limit/over-limit fixture discipline for the migrate package specifically.

- **Implements:** FR-121
- **Depends on:** T-0296
- **DoD:** >=5 distinct unrepresentable-construct fixture classes ship with expected RefusalReport outputs; CI runs Scan() against each and asserts byte-exact match on reported location and construct kind.
- **Test:** `TestFR_121_ConformanceVectors_UnrepresentableConstructs` (conformance)
- **Owner:** momus

**T-0298** Implement Phase 2 deterministic canonical-octet migration transform

> Build migrate.Transform(source, targetMajor) that, given a Phase-1-clean source (RefusalReport empty), produces the migrated document's canonical octet sequence deterministically: no map iteration order, no wall-clock, no goroutine/process-order dependency. Document the traversal-ordering contract migrate relies on ahead of canon's own M17 stabilization so a later canon change cannot silently alter migrate's output.

- **Implements:** FR-119
- **Depends on:** T-0296
- **DoD:** Transform() invoked twice in-process on the same input, and once each in two separate process invocations, produces byte-identical canonical octets; the ordering contract is documented in a code comment referencing the canon dependency.
- **Test:** `TestFR_119_TransformDeterministicAcrossInvocations` (unit)
- **Owner:** hephaestus

**T-0299** Publish external-trial corpus for cross-implementation migration determinism

> Publish a golden corpus of (source document, target major version) pairs with their expected canonical migrated octets, for consumption by a second, independently-authored implementation per CP-003/NFR-028's scope. This is the migrate-specific fixture set feeding M19's broader two-implementation gate evaluation; it does not itself run that gate.

- **Implements:** FR-119
- **Depends on:** T-0298
- **DoD:** Corpus ships with >=3 source/target-version pairs and their expected canonical octets, documented as reusable by an independent implementation, and referenced from contracts/README.md's conformance index.
- **Test:** `TestFR_119_ExternalTrialCorpusPublished` (external-trial)
- **Owner:** momus

**T-0300** Preserve content-unit identifiers unchanged through migration

> In Transform(), thread every run_id/unit_id from source to migrated output unchanged: no reissuing, no truncation or reencoding of the 128-bit CSPRNG token. Add a property-based fuzz harness that generates randomized documents with N distinct identifiers and asserts the migrated output's identifier set is exactly equal in value and count to the source's.

- **Implements:** FR-120
- **Depends on:** T-0298
- **DoD:** Fuzz harness runs >=10,000 iterations across randomized documents with zero identifier-mismatch findings; a targeted unit test confirms one specific known run_id survives migration byte-for-byte.
- **Test:** `TestFR_120_IdentifiersSurviveMigration` (fuzz)
- **Owner:** hephaestus

**T-0301** Enforce unsupported-major-version decline gate

> At every entry point that reads a document's Header (the migrate CLI path, and any other reader reusing this gate), check format-major against the implementation's supported set before any other processing. On an unsupported value, decline immediately, naming the version, with zero downstream dispositions applied (no partial header parse side effects, no extension-envelope processing, no validation pipeline steps executed). This is validation pipeline step 1 per data-model.md S7, scoped to M16 because FR-123 is not in M07's requirement list.

- **Implements:** FR-123
- **Depends on:** None
- **DoD:** A fixture with format-major = current+1 is declined before any other pipeline step executes (verified via call-count instrumentation showing zero downstream invocations), and the decline message names the unsupported version exactly.
- **Test:** `TestFR_123_UnsupportedMajorVersionDeclinedFirst` (conformance)
- **Owner:** hephaestus

**T-0302** Compose post-migration signature verdict with pre-migration coverage

> After a successful migration, a signature originally covering the pre-migration state must continue to report (via M09's existing FR-062 CoveringUnavailableState/verdict composition) that it covers the pre-migration state_id specifically, never the migrated state's state_id. Implement the migrate-side half only: Transform() must treat ATTEST-typed segments as opaque and untouched, so no code change is required in M09's verifier for the composed verdict to hold.

- **Implements:** FR-122
- **Depends on:** T-0298, T-0300
- **DoD:** A signed document migrated to the next major version reports its original signature as covering its original pre-migration state_id via the existing verify path, exercised end-to-end, with zero modifications made to M09's verifier code.
- **Test:** `TestFR_122_PreMigrationSignatureCoversOriginalState` (integration)
- **Owner:** argus

**T-0303** Implement hash-family re-protection layer (SHA3-256 outer digest)

> Build the registry package's re-protection mechanism: wrap an existing signed_object with a fresh outer digest computed under a stronger/alternate hash family (SHA3-256) plus a fresh time attestation, without re-signing — the original Ed25519 signature and its verdict remain intact and independently re-checkable. This addresses the hash-family-upgrade half of CON-016, distinct from a scheme-level break.

- **Implements:** CON-016
- **Depends on:** None
- **DoD:** A document re-protected under SHA3-256 still verifies its original signature to the same verdict as before re-protection; the re-protection-aware verifier additionally checks the new SHA3-256 outer digest; applying re-protection twice does not corrupt or invalidate the original signature.
- **Test:** `TestCON_016_ReProtectionPreservesOriginalVerdict` (integration)
- **Owner:** argus

**T-0304** Implement RESCIND-AND-RESIGN scheme-level break flow

> Implement DP-017's RESCIND-AND-RESIGN mechanism as a migrate CLI flag (per cli.md/TR-012's design decision, not a 12th verb), invocable only at a major-version boundary: it rescinds a signature whose cryptographic scheme is considered broken and attaches a fresh signature under the current allowlisted parameter set, recording the rescission (old signature id, reason, timestamp) in a RescindResignRecord so the rescinded signature's historical verdict remains inspectable, never erased. This is the only invocation path (an explicit, operator-supplied flag), addressing the plan.md-disclosed absence of any self-triggering condition by design rather than by policy.

- **Implements:** CON-016
- **Depends on:** T-0302, T-0303
- **DoD:** A document with a signature under a simulated broken-scheme flag, migrated with --rescind-and-resign, produces a document whose new signature verifies Valid under the current parameter set; its RescindResignRecord names the rescinded signature's original id and reason; the rescinded signature's own historical verdict remains queryable, not deleted; invoking the flag outside a major-version boundary is refused.
- **Test:** `TestCON_016_RescindAndResignAtMajorVersionBoundary` (integration)
- **Owner:** argus

**T-0305** Ship wire-format conformance vectors for RescindResignRecord

> Per the wire-format conformance vectors cross-cutting concern applying to M16, ship at-limit and one-past-limit fixtures for RescindResignRecord's wire shape (integrity.abnf S8), including FR-110's duplicate-identifier rule applied concretely to RescindResignRecord identity: two records resolving to the same identifier must be rejected naming both locations, never renamed or given precedence.

- **Implements:** CON-016
- **Depends on:** T-0304
- **DoD:** Conformance corpus includes a valid RescindResignRecord fixture, a duplicate-identifier pair fixture (rejected per FR-110 naming both locations), and a boundary-length fixture one octet past the record's maximum size (rejected); all three assert the documented verdict in CI.
- **Test:** `TestCON_016_RescindResignRecordConformanceVectors` (conformance)
- **Owner:** momus

**T-0306** argus security review of migration and re-protection flows

> Per the cross-cutting rule that every milestone touching crypto, signing, redaction, or migration-triggered re-signing gets an argus pass before merge, run a dedicated security review of the migrate and registry packages: confirm Transform() never mutates ATTEST-segment octets outside the documented re-protection/rescind flows, confirm RESCIND-AND-RESIGN cannot be invoked outside a major-version boundary, and confirm no migration path can silently downgrade or reattribute a signature's declared coverage (cross-checking T-1607's guarantee).

- **Implements:** CON-016
- **Depends on:** T-0298, T-0302, T-0303, T-0304
- **DoD:** argus review completed against migrate/ and registry/ with zero unresolved P0/P1 findings; the findings log is attached to the PR before merge.
- **Test:** `TestCON_016_ArgusReviewNoP0P1Findings` (integration)
- **Owner:** argus

---

### M17: Canonicalization

The L* traversal used by compact/publish/migrate/fresh-writes touches nearly every other package's output, so it stabilizes last among the mid-layer packages despite being conceptually central.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0307 | Canonical depth-first traversal core (L* order) | NFR-002, FR-124 | None | hephaestus | `TestNFR_002_CanonicalTraversalOrderMatchesTC` (unit) |
| T-0308 | Streaming Canonicalize(state, io.Writer) | NFR-002 | T-0307 | hephaestus | `TestNFR_002_CanonicalizeStreamsWithoutMaterializing` (benchmark) |
| T-0309 | Conformance corpus: canonicalization is storage-order independent | NFR-002 | T-0308 | momus | `CONF-CANON-001-storage-order-independence` (conformance) |
| T-0310 | Full compaction (canon.Compact) with unconditional signature-present refusal | NFR-004 | T-0308 | hephaestus | `TestNFR_004_FullCompactionRefusedWhenSignaturePresent` (unit) |
| T-0311 | Conformance corpus: full-compaction refusal exit-criteria (100/500) | NFR-004 | T-0310 | momus | `CONF-COMPACT-001-full-compaction-refusal-corpus` (conformance) |
| T-0312 | Partial compaction (canon.PartialCompact, DP-016) | NFR-004 | T-0307 | hephaestus | `TestNFR_004_PartialCompactionLeavesCoveredSegmentsByteIdentical` (unit) |
| T-0313 | Conformance vector: TOTAL-mode signature yields zero-reclaim partial compaction | NFR-004 | T-0312 | momus | `CONF-COMPACT-002-total-mode-zero-reclaim` (conformance) |
| T-0314 | Explicit-only invocation guard for Compact/PartialCompact | NFR-004 | T-0310, T-0312 | hephaestus | `TestNFR_004_CompactionExplicitOnlyNotInvokedBySave` (integration) |
| T-0315 | Minimal empty/BOTTOM-state definition and L*(BOTTOM,S) emission | FR-124 | T-0307, T-0308 | hephaestus | `TestFR_124_EmptyStateEmitsWellDefinedMinimalPrefix` (unit) |
| T-0316 | Degenerate/empty-content conformance corpus against validate | FR-124 | T-0315 | momus | `CONF-DEGENERATE-001-empty-content-case-corpus` (conformance) |
| T-0317 | Publish normative outcome table for the 5 empty-state reader operations | FR-124 | T-0315 | clio | `TestFR_124_EmptyStateOutcomeTableSchemaComplete` (unit) |
| T-0318 | Verify Publish/Migrate L* call sites delegate to streaming Canonicalize | NFR-002 | T-0308, T-0196, T-0298 | hephaestus | `TestNFR_002_PublishAndMigrateUseStreamingCanonicalize` (integration) |
| T-0319 | Deterministic text projection renderer (--format=text) | TR-004 | T-0307 | hephaestus | `TestTR_004_TextProjectionIsDeterministic` (unit) |
| T-0320 | Deterministic markup projection renderer (--format=html) | TR-004 | T-0307 | hephaestus | `TestTR_004_HtmlProjectionIsDeterministic` (unit) |
| T-0321 | Round-trip recovery utility: projection back to identical canonical octets | TR-004 | T-0319, T-0320 | hephaestus | `TestTR_004_ProjectionRoundTripsToIdenticalCanonicalOctets` (unit) |
| T-0322 | Conformance corpus: 100% round-trip over the full corpus | TR-004 | T-0321 | momus | `CONF-PROJECT-001-full-corpus-round-trip` (conformance) |
| T-0323 | Enforce projection non-reingestability (TR-005) | TR-005 | T-0319, T-0320 | hephaestus | `TestTR_005_ProjectionRejectedAsDocumentInput` (integration) |
| T-0324 | CI coverage gate: 100% test coverage on the canonicalizer core | NFR-002 | T-0307, T-0308 | prometheus | `TestNFR_002_CanonicalizerHas100PercentCoverage` (unit) |

**T-0307** Canonical depth-first traversal core (L* order)

> Implement the canon package's core depth-first traversal that walks a Document state's content subtrees in the same canonical order established by T_C (integrity.abnf S2.2.1's interim ascending-unit-id rule), yielding the ordered sequence of PDL-TLV records constituting C(S). The BOTTOM/empty state must yield the well-defined empty traversal (zero content-subtree records, fixed-prefix skeleton only). This is the shared walk consumed by Canonicalize, Compact, PartialCompact, Publish, Migrate, and the text/html projectors. Requires M08's finalized T_C traversal-order convention as an upstream input (not a task dependency here since M08 predates M17 in the spine).

- **Implements:** NFR-002, FR-124
- **Depends on:** None
- **DoD:** Traversal order is a pure function of state and is byte-identical to T_C's own per-subtree canonical order on every fixture in the shared integrity corpus; traversal over the BOTTOM state yields zero content records; two calls on the same state produce identical ordered record sequences.
- **Test:** `TestNFR_002_CanonicalTraversalOrderMatchesTC` (unit)
- **Owner:** hephaestus

**T-0308** Streaming Canonicalize(state, io.Writer)

> Implement `Canonicalize(state *Document, w io.Writer) error` that streams C(S) directly to w using T-1701's traversal, holding at most one content-subtree's serialized bytes in memory at a time and never materializing the complete canonical sequence as an in-memory buffer or temp file.

- **Implements:** NFR-002
- **Depends on:** T-0307
- **DoD:** Peak heap allocation during Canonicalize() of a 1 GiB document fixture is bounded independent of total document size (scales with the largest single subtree, not total octets, measured via pprof); output written to w is byte-identical to a reference in-memory-buffered serialization across the golden fixture set.
- **Test:** `TestNFR_002_CanonicalizeStreamsWithoutMaterializing` (benchmark)
- **Owner:** hephaestus

**T-0309** Conformance corpus: canonicalization is storage-order independent

> Author a conformance corpus proving Canonicalize(S) is a pure function of logical state S: the same logical state canonicalized from two distinct physical storage-order layouts (e.g. pre- and post- a partial-compaction reshuffle) must yield byte-identical C(S).

- **Implements:** NFR-002
- **Depends on:** T-0308
- **DoD:** Corpus of at least 10 fixture states, each canonicalized from at least 2 distinct physical layouts, all pairs byte-identical.
- **Test:** `CONF-CANON-001-storage-order-independence` (conformance)
- **Owner:** momus

**T-0310** Full compaction (canon.Compact) with unconditional signature-present refusal

> Implement `Compact(doc *Document) (*File, error)` performing full compaction as place(BOTTOM, S) = L*(S) via Canonicalize. Before any I/O, check for any present Signature record and return a named refusal error if one exists; never partially write output on refusal.

- **Implements:** NFR-004
- **Depends on:** T-0308
- **DoD:** Compact() on any document carrying one or more Signature records returns a refusal error with zero bytes written; Compact() on a signature-free document succeeds and its output is byte-identical to a direct Canonicalize() call on the same state.
- **Test:** `TestNFR_004_FullCompactionRefusedWhenSignaturePresent` (unit)
- **Owner:** hephaestus

**T-0311** Conformance corpus: full-compaction refusal exit-criteria (100/500)

> Author the plan.md Section 5 NFR-004 exit-criteria corpus: 100 Compact() attempts against signed documents (must be 100% refused) and 500 ordinary save operations (NoOp/Edit place() calls) that must never internally invoke Compact.

- **Implements:** NFR-004
- **Depends on:** T-0310
- **DoD:** 100/100 signed-document compaction attempts refused; 500/500 ordinary saves recorded as making zero calls into Compact or PartialCompact.
- **Test:** `CONF-COMPACT-001-full-compaction-refusal-corpus` (conformance)
- **Owner:** momus

**T-0312** Partial compaction (canon.PartialCompact, DP-016)

> Implement partial compaction as a distinct, explicit operation permitted while a signature is present, restricted to segments/ordinals named in NO present signature's covered_segment_ranges. It may relocate, coalesce, and reclaim only uncovered slots; every covered segment's ordinal, offset, length, and digest must remain byte-identical.

- **Implements:** NFR-004
- **Depends on:** T-0307
- **DoD:** PartialCompact() on a SUBSET-signed fixture reclaims only uncovered ordinals; a field-by-field diff of every covered segment's slot before/after is empty; PartialCompact() succeeds (is never refused) on a document carrying a present signature, distinguishing it from Compact's unconditional refusal.
- **Test:** `TestNFR_004_PartialCompactionLeavesCoveredSegmentsByteIdentical` (unit)
- **Owner:** hephaestus

**T-0313** Conformance vector: TOTAL-mode signature yields zero-reclaim partial compaction

> Author the fixture proving a TOTAL-mode-signed document's PartialCompact() call is a well-defined no-op (zero uncovered ordinals to reclaim), matching plan.md Section 8's disclosed residual-risk note, and distinguishing this no-op from an error condition.

- **Implements:** NFR-004
- **Depends on:** T-0312
- **DoD:** PartialCompact() on a TOTAL-mode-signed fixture returns success with zero ordinals reclaimed, never an error, and this is asserted as a positive expectation rather than a skipped case.
- **Test:** `CONF-COMPACT-002-total-mode-zero-reclaim` (conformance)
- **Owner:** momus

**T-0314** Explicit-only invocation guard for Compact/PartialCompact

> Enforce and test that neither Compact nor PartialCompact is reachable from the ledger `place()` NoOp/Edit code paths used by an ordinary save; both must be callable only via an explicit top-level API/CLI entry point, matching NFR-004's 'compaction is explicit-only, never part of a save' clause.

- **Implements:** NFR-004
- **Depends on:** T-0310, T-0312
- **DoD:** A fuzzed sequence of 1,000 ordinary edits (NoOp/Edit place() calls) invokes neither Compact nor PartialCompact, verified via call-graph or instrumentation check, not just an absence of observed behavior.
- **Test:** `TestNFR_004_CompactionExplicitOnlyNotInvokedBySave` (integration)
- **Owner:** hephaestus

**T-0315** Minimal empty/BOTTOM-state definition and L*(BOTTOM,S) emission

> Define the canonical empty-content Document state (BOTTOM) and implement its L* emission as a valid, minimal, well-defined file (fixed prefix plus zero content/resource/history segments). Used both for fresh-document creation and as the base case of every place() invocation (place(BOTTOM, no-content) = L*(BOTTOM)).

- **Implements:** FR-124
- **Depends on:** T-0307, T-0308
- **DoD:** L*(BOTTOM,S) for the zero-content state produces a fixed, reproducible minimal file across 2 independent invocations; the emitted file validates successfully under the M07 validation pipeline with zero findings.
- **Test:** `TestFR_124_EmptyStateEmitsWellDefinedMinimalPrefix` (unit)
- **Owner:** hephaestus

**T-0316** Degenerate/empty-content conformance corpus against validate

> Author one fixture per degenerate/empty-content case reachable through canon+validate (empty document, zero-run empty text unit, empty table with zero rows/cols, zero-page pagination declaration) and assert validate produces exactly one defined verdict per fixture, identical across independent runs.

- **Implements:** FR-124
- **Depends on:** T-0315
- **DoD:** At least 6 degenerate fixtures, each yielding exactly one non-ambiguous validate verdict with no crash or undefined case; verdicts match a golden reference across 2 independent same-platform (Go 1.25.1, darwin/arm64) build-and-run passes, since the project's stated toolchain is a single target and no second OS/arch is provisioned.
- **Test:** `CONF-DEGENERATE-001-empty-content-case-corpus` (conformance)
- **Owner:** momus

**T-0317** Publish normative outcome table for the 5 empty-state reader operations

> FR-124 requires exactly one defined outcome for extraction, preview, reflow, rasterisation, and validation of every degenerate/empty-content case. M17 does not depend on M05 (extract) or M14 (render) per the plan.md dependency spine, so this task's scope is limited to publishing the normative reference table (per reader role, the exact required output/verdict against the T-1709 empty-state fixture) into the contracts set; executing extract's and render's own conformance vectors against this fixture and table is M05's and M14's obligation when those packages are built, using this table as their acceptance contract.

- **Implements:** FR-124
- **Depends on:** T-0315
- **DoD:** The outcome table is checked in with one row per reader role (extraction, preview, reflow, rasterisation, validation) naming the exact expected output or verdict for the T-1709 fixture; CI fails the table's own schema check if any of the 5 roles is missing an entry.
- **Test:** `TestFR_124_EmptyStateOutcomeTableSchemaComplete` (unit)
- **Owner:** clio

**T-0318** Verify Publish/Migrate L* call sites delegate to streaming Canonicalize

> Verify that canon.Publish's L*(published_state) emission (M11) and the migration engine's L*(migrate(state)) fresh-file emission (M16) both invoke T-1702's streaming Canonicalize rather than any bespoke re-serialization path, so NFR-002's 'computable without rewriting the file at rest' guarantee holds uniformly at every L* call site, not only canon's own direct API.

- **Implements:** NFR-002
- **Depends on:** T-0308, T-0196, T-0298
- **DoD:** Static/integration check confirms Publish and Migrate delegate their full re-emission to canon.Canonicalize; no duplicate full-state serializer implementation exists elsewhere in the codebase.
- **Test:** `TestNFR_002_PublishAndMigrateUseStreamingCanonicalize` (integration)
- **Owner:** hephaestus

**T-0319** Deterministic text projection renderer (--format=text)

> Implement the canon-to-text projection function mapping C(S) to a deterministic, human-readable plain-text representation, driven by T-1701's traversal, backing the `project` CLI verb's `--format=text` mode.

- **Implements:** TR-004
- **Depends on:** T-0307
- **DoD:** Projecting the same state twice yields byte-identical text output; projecting two distinct states yields distinct text output.
- **Test:** `TestTR_004_TextProjectionIsDeterministic` (unit)
- **Owner:** hephaestus

**T-0320** Deterministic markup projection renderer (--format=html)

> Implement the canon-to-html projection function mapping C(S) to a deterministic markup representation, sharing T-1701's traversal with the text renderer, backing the `project` CLI verb's `--format=html` mode.

- **Implements:** TR-004
- **Depends on:** T-0307
- **DoD:** Projecting the same state twice yields byte-identical HTML output; output is well-formed HTML verified by a parse-and-reserialize idempotency check.
- **Test:** `TestTR_004_HtmlProjectionIsDeterministic` (unit)
- **Owner:** hephaestus

**T-0321** Round-trip recovery utility: projection back to identical canonical octets

> Implement a test-only recovery function that reconstructs canonical octets from a text or html projection, used solely to prove round-trip fidelity per TR-004; this recovery path must never be wired into any conforming reader's document-input path (that exclusion is T-1717's job).

- **Implements:** TR-004
- **Depends on:** T-0319, T-0320
- **DoD:** recovery(project(C(S))) equals C(S) exactly for every fixture in the shared conformance corpus.
- **Test:** `TestTR_004_ProjectionRoundTripsToIdenticalCanonicalOctets` (unit)
- **Owner:** hephaestus

**T-0322** Conformance corpus: 100% round-trip over the full corpus

> Run T-1715's round-trip check across the entire accumulated conformance corpus (every record-type and ceiling fixture produced by M01 through M16's own conformance suites) and track the pass rate in CI.

- **Implements:** TR-004
- **Depends on:** T-0321
- **DoD:** 100% of the accumulated conformance corpus round-trips from text/html projection to identical canonical octets, tracked as a CI gate.
- **Test:** `CONF-PROJECT-001-full-corpus-round-trip` (conformance)
- **Owner:** momus

**T-0323** Enforce projection non-reingestability (TR-005)

> Ensure the text/html projection carries no valid Protodoc magic or header and that validate/extract/verify reject it as a structural failure when presented as document input, keeping the projection outside the conformance surface; the `project` CLI verb's stdout payload always sets `reingestable: false`.

- **Implements:** TR-005
- **Depends on:** T-0319, T-0320
- **DoD:** Feeding a text or html projection file to validate, extract, or verify yields a structural-rejection verdict, never a partial parse or silent acceptance; every `project` verb invocation's payload sets `reingestable:false`.
- **Test:** `TestTR_005_ProjectionRejectedAsDocumentInput` (integration)
- **Owner:** hephaestus

**T-0324** CI coverage gate: 100% test coverage on the canonicalizer core

> Wire the canon package's canonicalizer core (T-1701 traversal, T-1702 Canonicalize) into the project's coverage-tracking CI per the constitution's 100%-coverage-on-canonicalisation quality bar (plan.md Section 5 row 512).

- **Implements:** NFR-002
- **Depends on:** T-0307, T-0308
- **DoD:** CI fails the build if internal/canon/traversal.go or internal/canon/canonicalize.go drops below 100% line coverage.
- **Test:** `TestNFR_002_CanonicalizerHas100PercentCoverage` (unit)
- **Owner:** prometheus

---

### M18: CLI Surface

cmd/protodoc wraps every package above it into the 11 TR-012 verbs and is the first point where an external conditional-replacement write (TR-010) has to be resolved into an actual interface.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0325 | CLI command dispatch & global flags framework | TR-012 | None | hephaestus | `TestTR_012_DispatchRegistersAllElevenVerbs` (unit) |
| T-0326 | Exit-code precedence engine | TR-012 | T-0325 | hephaestus | `TestTR_012_ExitCodePrecedenceTable` (unit) |
| T-0327 | validate verb command | TR-012 | T-0325, T-0326 | hephaestus | `TestTR_012_ValidateVerbInvocationShape` (integration) |
| T-0328 | Validate machine-readable finding schema (5-field report) | TR-001 | T-0327 | hephaestus | `TestTR_001_FindingCarriesAllFiveFields` (conformance) |
| T-0329 | inspect verb command | TR-012 | T-0325, T-0326 | hephaestus | `TestTR_012_InspectVerbBoundedRead` (integration) |
| T-0330 | extract verb command | TR-012 | T-0325, T-0326, T-0091, T-0092, T-0093 | hephaestus | `TestTR_012_ExtractVerbStreaming` (integration) |
| T-0331 | verify verb command | TR-012 | T-0325, T-0326 | argus | `TestTR_012_VerifyVerbVerdictMapping` (integration) |
| T-0332 | diff verb command | TR-012, TR-002 | T-0325, T-0326 | hephaestus | `TestTR_002_DiffVerbConstructLevel` (integration) |
| T-0333 | merge verb command | TR-012, TR-003 | T-0325, T-0326 | hephaestus | `TestTR_003_MergeVerbConflictExit` (integration) |
| T-0334 | project verb command | TR-012, TR-004, TR-005 | T-0325, T-0326 | hephaestus | `TestTR_004_ProjectRoundTripsExactly` (integration) |
| T-0335 | redact verb command | TR-012 | T-0325, T-0326 | argus | `TestTR_012_RedactVerbOmissionEnumeration` (integration) |
| T-0336 | publish verb command | TR-012 | T-0325, T-0326 | hephaestus | `TestTR_012_PublishVerbZeroResidue` (integration) |
| T-0337 | sign verb command | TR-012 | T-0325, T-0326 | argus | `TestTR_012_SignVerbDeterministicOutput` (integration) |
| T-0338 | migrate verb command (with RESCIND-AND-RESIGN flag) | TR-012 | T-0325, T-0326 | hephaestus | `TestTR_012_MigrateVerbRefusalFirst` (integration) |
| T-0339 | CLI requirement-to-verb traceability table | TR-012 | T-0327, T-0329, T-0330, T-0331, T-0332, T-0333, T-0334, T-0335, T-0336, T-0337, T-0338 | clio | `TestTR_012_TraceabilityTableMatchesDispatch` (unit) |
| T-0340 | End-to-end CLI conformance suite | TR-012 | T-0326, T-0327, T-0328, T-0329, T-0330, T-0331, T-0332, T-0333, T-0334, T-0335, T-0336, T-0337, T-0338 | momus | `TestTR_012_CLIConformanceSuite` (conformance) |
| T-0341 | Conditional-replacement write flag for mutating verbs | TR-010 | T-0325, T-0044 | hephaestus | `TestTR_010_ConditionalWriteRefusesOnMismatch` (integration) |

**T-0325** CLI command dispatch & global flags framework

> Build cmd/protodoc's root entrypoint using only Go stdlib (flag/os.Args, per CP-010): a verb subcommand routing table registering all 11 TR-012 verbs (validate, inspect, extract, verify, diff, merge, project, redact, publish, sign, migrate), shared global flags (--json, --help, --version), and a single stdout JSON envelope writer used by every verb per cli.md's invocation-shape section. This is the foundation every other verb task wires into; it contains no verb business logic itself.

- **Implements:** TR-012
- **Depends on:** None
- **DoD:** `protodoc <verb> --help` prints per-verb usage for all 11 registered verb names with none missing; invoking an unregistered verb name exits with the USAGE code without importing or calling any downstream package.
- **Test:** `TestTR_012_DispatchRegistersAllElevenVerbs` (unit)
- **Owner:** hephaestus

**T-0326** Exit-code precedence engine

> Implement the shared exit-code type encoding cli.md's 8 documented exit codes and their precedence order (per CON-011: a resource-budget refusal ranks above UNAVAILABLE/UNVERIFIED/REFUSED but below INVALID), plus a resolver function that, given a set of simultaneously-true conditions from any verb, returns exactly one code deterministically.

- **Implements:** TR-012
- **Depends on:** T-0325
- **DoD:** A table-driven test enumerating every pairwise combination of the 8 documented conditions resolves to the single precedence winner matching cli.md's exit-code table exactly; zero ambiguous or undocumented pairs remain.
- **Test:** `TestTR_012_ExitCodePrecedenceTable` (unit)
- **Owner:** hephaestus

**T-0327** validate verb command

> Wire `protodoc validate` to the M07 validate package, emitting cli.md S2's `checks`/`findings` stdout JSON shape and mapping validator verdicts (pass/fail/budget-refused) through the T-1802 exit-code engine.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Running validate against M07's at-limit and over-limit ceiling conformance fixtures produces JSON matching cli.md's findings schema field-for-field, exiting OK for the at-limit fixture and INVALID for the over-limit one.
- **Test:** `TestTR_012_ValidateVerbInvocationShape` (integration)
- **Owner:** hephaestus

**T-0328** Validate machine-readable finding schema (5-field report)

> Extend the validate finding struct/serializer so every emitted finding always carries all 5 mandated fields: rule id, severity, offending unit id, octet offset, and the normative-statement id (FR-/NFR-/CON-/TR- id) the rule enforces, sourced from a checked-in rule-to-requirement map rather than hardcoded per call site.

- **Implements:** TR-001
- **Depends on:** T-0327
- **DoD:** A golden-file test with one fixture per validator rule category confirms all 5 fields are non-empty and the normative-statement id matches the checked-in map; a finding missing any of the 5 fields fails the build.
- **Test:** `TestTR_001_FindingCarriesAllFiveFields` (conformance)
- **Owner:** hephaestus

**T-0329** inspect verb command

> Wire `protodoc inspect` to read only the fixed 1,048,576-octet prefix (M01 container) and emit the documented segment-inventory/coverage-summary JSON payload without decoding any segment body content.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** inspect on a valid fixture lists every SegmentTableSlot's type/length/digest and coverage-hint field in the documented JSON shape, with an I/O trace confirming reads never extend past octet 1,048,576.
- **Test:** `TestTR_012_InspectVerbBoundedRead` (integration)
- **Owner:** hephaestus

**T-0330** extract verb command

> Wire `protodoc extract` to the M05 extract package's streaming reader, exposing cli.md's documented flags (page range, language filter, locator emission) and JSON/plain-text output modes.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326, T-0091, T-0092, T-0093
- **DoD:** extract on M05's 1 GiB/10,000-page benchmark fixture streams the first unit's text before 15% of the file has been read, exits OK, and imports no non-extract package for its core logic.
- **Test:** `TestTR_012_ExtractVerbStreaming` (integration)
- **Owner:** hephaestus

**T-0331** verify verb command

> Wire `protodoc verify` to M09's Verify()/Verdict machinery and M10's LTV checks, emitting cli.md S5's `signatures` payload (signed state, signer, pinned presentation) and enforcing FR-115's rule that a signer identity is never shown alongside a non-Valid verdict.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Fixtures covering all 4 Verdict values (Valid, AttestedWithDeclaredOmissions, Unverified, CoveringUnavailableState) each produce the documented payload shape and correct exit code; no non-Valid fixture's stdout contains a signer-identity field.
- **Test:** `TestTR_012_VerifyVerbVerdictMapping` (integration)
- **Owner:** argus

**T-0332** diff verb command

> Wire `protodoc diff` to M13's construct-level diff engine, emitting cli.md S6's per-construct change list and reporting no construct both inputs agree on, per TR-002.

- **Implements:** TR-012, TR-002
- **Depends on:** T-0325, T-0326
- **DoD:** Diffing two fixtures differing in exactly one Annotation and one moved Run reports exactly those 2 changed constructs and zero unrelated ones.
- **Test:** `TestTR_002_DiffVerbConstructLevel` (integration)
- **Owner:** hephaestus

**T-0333** merge verb command

> Wire `protodoc merge` to M13's DP-015 merge classifier, surfacing TR-003's non-zero-exit-naming-both-values behavior on a genuine conflict, plus the FR-024/CON-024/CON-025 named refusal cases at the CLI boundary.

- **Implements:** TR-012, TR-003
- **Depends on:** T-0325, T-0326
- **DoD:** A genuine R2/R3 conflict fixture exits CONFLICT naming both source values in stdout JSON; a retention-point-crossing (CON-024) fixture and a history-mode-mismatch (CON-025) fixture each exit REFUSED naming the specific violated condition.
- **Test:** `TestTR_003_MergeVerbConflictExit` (integration)
- **Owner:** hephaestus

**T-0334** project verb command

> Wire `protodoc project` to M17's canon package, emitting TR-004's deterministic human-readable text projection and enforcing TR-005's 'reingestable: false' guarantee so no other verb accepts projection output as document input.

- **Implements:** TR-012, TR-004, TR-005
- **Depends on:** T-0325, T-0326
- **DoD:** `protodoc project` output, fed through the test's own reverse-mapping (used only for verification), round-trips to the identical canonical octet sequence for 3 fixtures of increasing size; every other verb rejects the projection file as malformed rather than as a valid document.
- **Test:** `TestTR_004_ProjectRoundTripsExactly` (integration)
- **Owner:** hephaestus

**T-0335** redact verb command

> Wire `protodoc redact` to M11's redaction machinery, exposing subtree-selection flags and producing output whose subsequent verify reports AttestedWithDeclaredOmissions enumerating exactly the redacted subtree(s).

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Redacting a designated subtree on a signed fixture produces output where verify (T-1807) reports AttestedWithDeclaredOmissions naming exactly that subtree, and a raw byte-grep of the output file finds none of the redacted plaintext.
- **Test:** `TestTR_012_RedactVerbOmissionEnumeration` (integration)
- **Owner:** argus

**T-0336** publish verb command

> Wire `protodoc publish` to M11/M17's canon.Publish path, enforcing FR-078/FR-080/FR-081's zero-residue and custody-preservation guarantees at the CLI boundary.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Publishing a fixture with prior redactions and edit history produces output containing zero octets of removed content (including orphan-carriage quoted_text per FR-078) while every custody/fixity field's value matches the input unchanged.
- **Test:** `TestTR_012_PublishVerbZeroResidue` (integration)
- **Owner:** hephaestus

**T-0337** sign verb command

> Wire `protodoc sign` to M06/M09's EdDSA-Protodoc-1 signing path and M10's ATTESTATION_EVIDENCE attachment, exposing TOTAL/SUBSET coverage-mode flags and credential-chain/revocation/time-attestation input flags.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Signing the same fixture and key twice produces byte-identical signature octets (NFR-006), and a SUBSET-mode sign run produces a CoverageDescriptor whose covered/uncovered split matches the flags passed.
- **Test:** `TestTR_012_SignVerbDeterministicOutput` (integration)
- **Owner:** argus

**T-0338** migrate verb command (with RESCIND-AND-RESIGN flag)

> Wire `protodoc migrate` to M16's refusal-first two-phase migration, including the RESCIND-AND-RESIGN longevity mechanism as a `--rescind-and-resign` flag rather than a separate 12th verb, matching cli.md's design decision.

- **Implements:** TR-012
- **Depends on:** T-0325, T-0326
- **DoD:** Migrating a fixture with one unrepresentable construct halts in phase 1, naming the construct and location, with zero output file written; migrating a clean fixture with --rescind-and-resign produces output whose pre-migration signature still reports covering the pre-migration state (FR-122).
- **Test:** `TestTR_012_MigrateVerbRefusalFirst` (integration)
- **Owner:** hephaestus

**T-0339** CLI requirement-to-verb traceability table

> Maintain cli.md's requirement-to-verb traceability table mapping each FR/NFR/CON/TR id surfaced through the CLI to its owning verb(s), plus a CI check that every table row names a verb present in the T-1801 dispatch registry and every registered verb appears in the table.

- **Implements:** TR-012
- **Depends on:** T-0327, T-0329, T-0330, T-0331, T-0332, T-0333, T-0334, T-0335, T-0336, T-0337, T-0338
- **DoD:** CI check fails if any traceability-table row names a verb absent from dispatch, or any of the 11 registered verbs is absent from the table; check currently passes.
- **Test:** `TestTR_012_TraceabilityTableMatchesDispatch` (unit)
- **Owner:** clio

**T-0340** End-to-end CLI conformance suite

> Build the golden conformance suite invoking all 11 verbs as subprocesses against the shared fixture corpus, asserting invocation shape, stdout JSON schema, and exit-code precedence for every documented case in cli.md; feeds M19's harness-maturity check.

- **Implements:** TR-012
- **Depends on:** T-0326, T-0327, T-0328, T-0329, T-0330, T-0331, T-0332, T-0333, T-0334, T-0335, T-0336, T-0337, T-0338
- **DoD:** The suite exercises every verb against every fixture-reachable documented exit code and all cases pass; suite runs in CI on every PR touching cmd/protodoc.
- **Test:** `TestTR_012_CLIConformanceSuite` (conformance)
- **Owner:** momus

**T-0341** Conditional-replacement write flag for mutating verbs

> Expose a `--if-match=<expected-state-id>` write-conditional flag on every mutating verb (sign, redact, publish, migrate, merge), backed by a storage-backend interface (local-file plus at least one ETag/generation-number-capable backend) that refuses the write and names the current holder on mismatch. TR-010 is formally tracked as an M02 storage-package requirement; this task is the CLI-facing interface the M18 rationale calls out as where it must actually be resolved into a usable surface.

- **Implements:** TR-010
- **Depends on:** T-0325, T-0044
- **DoD:** A concurrent-write simulation against a mock ETag-capable backend causes the second writer to refuse, naming the first writer's resulting state id, for every mutating verb that accepts --if-match.
- **Test:** `TestTR_010_ConditionalWriteRefusesOnMismatch` (integration)
- **Owner:** hephaestus

---

### M19: Conformance, Fuzzing & Governance Convergence

CP-003 (two independent implementations), CP-011 (corpus ships with prose), CP-012 (continuous fuzzing), and CP-014 (governance/licensing) are release gates that can only be evaluated once every package above exists to be fuzzed, ported, and audited against.

| Task | Title | Implements | Depends on | Owner | Test |
|---|---|---|---|---|---|
| T-0342 | Publish NFR-011 reference measurement configuration | NFR-011 | None | clio | `TestNFR_011_ReferenceConfigPublished` (integration) |
| T-0343 | Retrofit existing NFR benchmark suites to cite the NFR-011 reference config | NFR-011 | T-0342 | prometheus | `TestNFR_011_BenchmarksCiteReferenceConfig` (benchmark) |
| T-0344 | Implement per-reader-role normative-statement ceiling CI check | NFR-025 | None | prometheus | `TestNFR_025_RoleStatementCeilingsEnforced` (unit) |
| T-0345 | Conduct external-implementer trial: extracting-and-validating reader, 5-working-day budget | NFR-026 | None | momus | `TestNFR_026_ExternalReaderTrial_5DayBudget` (external-trial) |
| T-0346 | Conduct external-implementer trial: rendering reader, 30-working-day budget | NFR-027 | None | momus | `TestNFR_027_ExternalRenderingTrial_30DayBudget` (external-trial) |
| T-0347 | Record Eyvar's CP-002 scope ruling on the NFR-027 schedule-risk conflict | NFR-027 | T-0346, T-0071, T-0091 | clio | `TestNFR_027_ScheduleRiskRulingRecorded` (integration) |
| T-0348 | Define CP-003/NFR-028 second-implementation scope and acceptance protocol | NFR-028 | None | clio | `TestNFR_028_SecondImplScopeDefined` (integration) |
| T-0349 | Record CQ-012 second-implementation funding/sponsor decision | NFR-028, CON-019 | T-0348 | clio | `TestNFR_028_CQ012FundingDecisionRecorded` (integration) |
| T-0350 | Build cross-implementation differential conformance harness | NFR-028 | T-0348 | prometheus | `TestNFR_028_DifferentialHarnessDiffsCanonicalOctets` (conformance) |
| T-0351 | Commission and execute the second independent implementation conformance run | NFR-028 | T-0349, T-0350 | momus | `TestNFR_028_SecondImplementationMatchesCanonicalOctetsAndVerdicts` (external-trial) |
| T-0352 | Build normative-statement-to-conformance-case traceability coverage checker | NFR-029 | None | prometheus | `TestNFR_029_TraceabilityCheckerParsesAllRuleIDs` (unit) |
| T-0353 | Run traceability coverage audit and file the PD-rule registry gap report | NFR-029 | T-0352 | momus | `TestNFR_029_AllNormativeStatementsHaveConformanceCase` (conformance) |
| T-0354 | Implement the feature register (A-FEATURE audit) gated on two-implementation pass | CON-019 | T-0351 | hephaestus | `TestCON_019_FeatureNormativeStatusGatedOnTwoImplPass` (integration) |
| T-0355 | Draft the irrevocable royalty-free licence with steward, succession, and deprecation-window policy | CON-026 | None | clio | `TestCON_026_LicenceDraftPublished` (integration) |
| T-0356 | Record Eyvar's governance approval of the CON-026 licensing gate | CON-026 | T-0355 | clio | `TestCON_026_GovernanceGateApprovedAndOnRecord` (integration) |
| T-0357 | Aggregate continuous-fuzzing harness maturity report across all fuzz targets (CP-012) | NFR-029 | None | prometheus | `TestNFR_029_FuzzHarnessMaturityMeetsCP012Bar` (fuzz) |
| T-0358 | Capstone: record the v1-stable declaration go/no-go decision | NFR-025, NFR-026, NFR-027, NFR-028, NFR-029, CON-019, CON-026 | T-0344, T-0345, T-0347, T-0351, T-0353, T-0354, T-0356, T-0357, T-0044, T-0365 | clio | `TestM19_V1StableDeclarationGateRecorded` (integration) |
| T-0365 | Record Eyvar's ruling on an NFR-026 external-trial budget miss | NFR-026 | T-0345 | clio | `TestNFR_026_TrialMissRulingRecorded` (integration) |

**T-0342** Publish NFR-011 reference measurement configuration

> NFR-011 requires a named reference measurement configuration (processor, core count, memory, storage class, OS) against which every time/memory bound in the spec (NFR-012..019 etc.) is judged. No such artifact exists anywhere in plan.md/data-model.md/contracts (confirmed gap). Author a non-normative operational doc pinning the exact benchmark host spec used for this repo's own CI (matches the stated toolchain: Go 1.25.1, darwin/arm64, plus concrete CPU model, core count, RAM, storage class, kernel version) and cite it from Section 5's existing raw-arithmetic estimates. This does not reopen or alter any FR/NFR text — it fills an operational parameter the frozen spec already presumes exists.

- **Implements:** NFR-011
- **Depends on:** None
- **DoD:** docs/nfr-011-reference-config.md exists naming processor, core count, RAM, storage class, OS/kernel, Go version; every existing NFR-012..NFR-019 benchmark job config references its identifier; Eyvar has approved the doc since it becomes the baseline every other NFR 'satisfied' verdict depends on.
- **Test:** `TestNFR_011_ReferenceConfigPublished` (integration)
- **Owner:** clio

**T-0343** Retrofit existing NFR benchmark suites to cite the NFR-011 reference config

> The extraction (NFR-012/013/014, M05), preview (NFR-015/016, M01), page-render (NFR-017/018, M14) and rasterizer (NFR-019, M14) benchmark jobs were built before a reference configuration existed and record only raw numbers with no config citation. Update each benchmark's output/report to stamp the T-0342 config id alongside its measured value so every published 'satisfied' verdict is traceable to the configuration it was measured under.

- **Implements:** NFR-011
- **Depends on:** T-0342
- **DoD:** Every NFR-012..019 benchmark report line includes the reference-config id from T-0342; CI fails if a benchmark result is emitted without one.
- **Test:** `TestNFR_011_BenchmarksCiteReferenceConfig` (benchmark)
- **Owner:** prometheus

**T-0344** Implement per-reader-role normative-statement ceiling CI check

> NFR-025 requires each of the 3 reader conformance roles (extracting/validating-and-verifying/rendering, CON-018) to have a stated maximum normative-statement count, tracked in CI. data-model.md's ceiling table already states the numbers (Extraction<=150, Validating-and-verifying<=400, Rendering<=500, CP-002/CON-009); build the CI tool that counts normative statements assigned to each role by walking spec.md's role-binding annotations and fails the build if any role exceeds its ceiling.

- **Implements:** NFR-025
- **Depends on:** None
- **DoD:** CI job counts per-role normative-statement totals against the checked-in ceiling table and fails on any overage; a synthetic fixture that pushes one role one statement over its ceiling is verified to fail the build.
- **Test:** `TestNFR_025_RoleStatementCeilingsEnforced` (unit)
- **Owner:** prometheus

**T-0345** Conduct external-implementer trial: extracting-and-validating reader, 5-working-day budget

> NFR-026 requires that an outside implementer, working from the spec text alone (no access to this reference implementation's source), can build a passing extracting-and-validating reader within 5 working days. Recruit an implementer with no prior exposure to this codebase, hand them only the frozen spec/contracts, time-box the exercise, and run their resulting reader against the extract+validate conformance corpus.

- **Implements:** NFR-026
- **Depends on:** None
- **DoD:** A dated trial report records elapsed working days and the pass/fail result of the implementer's reader against the full extract+validate conformance corpus. If the trial passes within 5 working days, this task is done. If it misses the 5-day budget or fails the corpus, T-NEW-1 (Eyvar's ruling on the NFR-026 miss) must be filed and remain open before this task is considered closed — mirroring how T-0347 handles the analogous NFR-027 miss; a miss with no ruling task filed does not satisfy this DoD.
- **Test:** `TestNFR_026_ExternalReaderTrial_5DayBudget` (external-trial)
- **Owner:** momus

**T-0346** Conduct external-implementer trial: rendering reader, 30-working-day budget

> NFR-027 requires an outside implementer to build a passing rendering reader within 30 working days from the spec text alone. plan.md Section 9 Conflict 4 already discloses this is AT RISK (revised estimate 28-49 days, driven mostly by CQ-006 shaping-oracle matching). Run the same isolated-implementer trial as T-0345 but scoped to the full rendering role (PLP-1, restricted-PNG, rasterizer, Knuth-Plass reflow, fixed pagination, shaping-oracle matching) and record actual elapsed days.

- **Implements:** NFR-027
- **Depends on:** None
- **DoD:** A dated trial report records elapsed working days and the pass/fail result of the implementer's rendering reader against the rendering conformance corpus, explicitly comparing actual elapsed days to the 30-day budget and the plan.md 28-49 day risk estimate.
- **Test:** `TestNFR_027_ExternalRenderingTrial_30DayBudget` (external-trial)
- **Owner:** momus

**T-0347** Record Eyvar's CP-002 scope ruling on the NFR-027 schedule-risk conflict

> plan.md Section 9 Conflict 4 flags NFR-027's 30-day budget as at-risk and states it 'needs an explicit scope ruling under CP-002.' T-0346's trial result either confirms or refutes the risk with real data. Present that result to Eyvar and get an explicit ruling: accept the revised 28-49 day estimate as the new documented bound, descope part of the rendering role, or accept the miss as a known limitation. This does not itself change spec.md (frozen) — it records a governance decision plan.md can carry forward.

- **Implements:** NFR-027
- **Depends on:** T-0346, T-0071, T-0091
- **DoD:** Eyvar's ruling on Conflict 4 is recorded as a structured, dated entry (fields: date, decision-maker, decision, rationale) in plan.md Section 9 and mirrored in specs/CHANGES.md, closing the previously 'unresolved, needs ruling' status; TestNFR_027_ScheduleRiskRulingRecorded parses both files and asserts all four fields are present and decision-maker == 'Eyvar García'.
- **Test:** `TestNFR_027_ScheduleRiskRulingRecorded` (integration)
- **Owner:** clio

**T-0348** Define CP-003/NFR-028 second-implementation scope and acceptance protocol

> NFR-028 and the CP-003 gate require a second, independently-authored implementation to match canonical octets and verdicts before v1 is declared stable, but only for container, validate, and canonical-serialiser (per the two-implementation-gate cross-cutting note — rendering/merge/history/etc. are explicitly exempt from this specific gate). Write the acceptance protocol: what 'independent' means (no shared source access), which corpus is authoritative, what 'match' means byte-for-byte, and how a disagreement is triaged (spec bug vs. implementation bug).

- **Implements:** NFR-028
- **Depends on:** None
- **DoD:** A written, Eyvar-reviewed protocol document exists naming the exact scope (container/validate/canon packages only), the independence criteria, the shared conformance corpus version, and the byte-exact/verdict-exact match definition.
- **Test:** `TestNFR_028_SecondImplScopeDefined` (integration)
- **Owner:** clio

**T-0349** Record CQ-012 second-implementation funding/sponsor decision

> plan.md Section 8's risk table discloses, as high severity, that the CQ-012/CP-003 two-independent-implementation gate has no funding, sponsor, or recruitment strategy on record, and explicitly asks Eyvar to decide before phase 5 (analyze) closes. This decision blocks recruiting the second implementation (T-0351) and blocks CON-019's feature register from ever having pass data to gate on. Bring T-0348's scoped protocol to Eyvar and get a funding/sponsor decision (self-funded contractor, community bounty, partner org, or explicit non-funding with a fallback plan) on record.

- **Implements:** NFR-028, CON-019
- **Depends on:** T-0348
- **DoD:** Eyvar's funding/sponsor decision for the second implementation is recorded as a structured, dated entry (fields: date, decision-maker, path-forward, fallback-if-unfunded) in plan.md Section 8's risk row and specs/CHANGES.md; TestNFR_028_CQ012FundingDecisionRecorded parses both files and asserts all four fields are present, updating the row from 'no funding on record' to the resolved state.
- **Test:** `TestNFR_028_CQ012FundingDecisionRecorded` (integration)
- **Owner:** clio

**T-0350** Build cross-implementation differential conformance harness

> Build the tooling that runs two implementations of container/validate/canon against the shared conformance corpus (including the CON-010 at-limit/over-limit fixtures and the plan.md Section 10 first-class task fixtures: PD-RING-001 tie vector, frame_count boundary vector) and diffs (a) canonical octet output byte-for-byte and (b) reported verdicts, producing a single pass/fail per corpus case with a readable diff on mismatch.

- **Implements:** NFR-028
- **Depends on:** T-0348
- **DoD:** Harness runs both implementations' binaries against every corpus case, reports byte-exact octet diffs and verdict mismatches, and exits non-zero on any disagreement; a deliberately-mismatched fixture pair is verified to be caught.
- **Test:** `TestNFR_028_DifferentialHarnessDiffsCanonicalOctets` (conformance)
- **Owner:** prometheus

**T-0351** Commission and execute the second independent implementation conformance run

> With funding/scope settled (T-0349) and the harness built (T-0350), commission the second independent implementation of container/validate/canon per the T-0348 protocol, then run it through the T-0350 harness against the full corpus. This is the concrete satisfaction of CP-003/NFR-028's release gate.

- **Implements:** NFR-028
- **Depends on:** T-0349, T-0350
- **DoD:** The second implementation passes the differential harness on every corpus case (byte-exact canonical octets, matching verdicts) for container/validate/canon, or every remaining mismatch is triaged and either fixed or recorded as an accepted spec ambiguity for Eyvar/themis ruling.
- **Test:** `TestNFR_028_SecondImplementationMatchesCanonicalOctetsAndVerdicts` (external-trial)
- **Owner:** momus

**T-0352** Build normative-statement-to-conformance-case traceability coverage checker

> NFR-029 requires every normative statement to carry a stable id mapped to at least one conformance case, CI-blocking on any gap. Build a tool that parses every FR-/NFR-/CON-/TR- id and every PD-*-NNN validator-rule id out of spec.md and contracts/*.abnf, cross-references them against the checked-in conformance corpus's own id annotations, and reports every id with zero matching cases.

- **Implements:** NFR-029
- **Depends on:** None
- **DoD:** Tool runs over the frozen spec.md/contracts and the current conformance corpus, correctly identifies every id with at least one case and every id with none; a fixture with one deliberately-orphaned rule id is verified to be flagged.
- **Test:** `TestNFR_029_TraceabilityCheckerParsesAllRuleIDs` (unit)
- **Owner:** prometheus

**T-0353** Run traceability coverage audit and file the PD-rule registry gap report

> Run T-0352's checker against the full 197-requirement set and the current corpus. The inventory already identifies that only 2 of 18 spec-named PD-* validator rule ids (PD-EXT-002, PD-MODE-001) reappear anywhere downstream, and CON-002's own PD-NORM-001 was silently renamed to PD-NFC-001/002 (violating CP-011 non-negotiable #5: rule identifiers never renumbered or reused). Do not silently fix these in the frozen contracts artifacts; file them as an explicit gap report for Eyvar/themis to rule on before release, per CI blocking every 'satisfied' claim that rests on an unmapped id.

- **Implements:** NFR-029
- **Depends on:** T-0352
- **DoD:** A gap report enumerates every normative-statement id with zero conformance cases (including the 16 missing PD-rule ids and the PD-NORM-001/PD-NFC-001 rename) and is handed to Eyvar/themis for a ruling; CI is configured to block a stable-release declaration while the report is non-empty.
- **Test:** `TestNFR_029_AllNormativeStatementsHaveConformanceCase` (conformance)
- **Owner:** momus

**T-0354** Implement the feature register (A-FEATURE audit) gated on two-implementation pass

> CON-019 requires no feature be marked normative until two source-independent implementations pass every conformance case for it. Build a feature register (one entry per named feature/mechanism in data-model.md) that starts every feature as 'provisional' and flips it to 'normative' only when T-0350's differential harness records a passing run from a second implementation covering that feature's conformance cases.

- **Implements:** CON-019
- **Depends on:** T-0351
- **DoD:** Register lists every named feature with a provisional/normative status; status flips to normative only on recorded two-implementation pass evidence; a feature with only single-implementation coverage is verified to remain provisional.
- **Test:** `TestCON_019_FeatureNormativeStatusGatedOnTwoImplPass` (integration)
- **Owner:** hephaestus

**T-0355** Draft the irrevocable royalty-free licence with steward, succession, and deprecation-window policy

> CON-026 requires the specification be published under an irrevocable royalty-free licence with a named steward, a succession plan, and a deprecation window, on record before v1. This is a governance/legal artifact, not a technical mechanism (already noted as outside plan.md's architectural scope). Draft the licence text, name the initial steward, define the succession process if the steward becomes unable to act, and state the minimum deprecation-window period for any future breaking change.

- **Implements:** CON-026
- **Depends on:** None
- **DoD:** LICENSE and GOVERNANCE.md exist in the repo root; GOVERNANCE.md contains four required headed sections (Licence Grant, Named Steward, Succession Process, Deprecation Window Policy), each non-empty, drafted for Eyvar's review; TestCON_026_LicenceDraftPublished parses GOVERNANCE.md and asserts all four sections are present.
- **Test:** `TestCON_026_LicenceDraftPublished` (integration)
- **Owner:** clio

**T-0356** Record Eyvar's governance approval of the CON-026 licensing gate

> The draft licence/governance artifact from T-0355 needs Eyvar's explicit approval before it satisfies CON-026's 'on record before v1' bar. Present the draft, capture the approval (or requested changes, iterate), and mark the gate closed.

- **Implements:** CON-026
- **Depends on:** T-0355
- **DoD:** specs/CHANGES.md records a structured, dated entry (fields: date, decision-maker, decision=approved/changes-requested, referenced-files=[LICENSE, GOVERNANCE.md]); TestCON_026_GovernanceGateApprovedAndOnRecord parses the entry and asserts decision-maker == 'Eyvar García' and decision == 'approved' before CON-026's governance gate is marked satisfied and referenced from the v1-stable decision (T-0358).
- **Test:** `TestCON_026_GovernanceGateApprovedAndOnRecord` (integration)
- **Owner:** clio

**T-0357** Aggregate continuous-fuzzing harness maturity report across all fuzz targets (CP-012)

> The cross-cutting CP-012 continuous-fuzzing concern spans M01/M02/M03/M07/M09/M10/M14's decode paths over untrusted bytes (container parse, extension envelope, validator, CoverageDescriptor range-list canonicalisation, LTV evidence, PLP-1/PNG codecs) and explicitly feeds M19's harness-maturity check before v1 is declared stable. Aggregate crash-free run duration, corpus size, and coverage percentage from every fuzz target across those milestones into one maturity report and check it against a documented minimum bar (run duration, zero unresolved crashes).

- **Implements:** NFR-029
- **Depends on:** None
- **DoD:** Report aggregates all fuzz targets' crash-free duration and coverage into one document; CI fails the report generation if any target has an unresolved crash or falls below the documented minimum crash-free duration.
- **Test:** `TestNFR_029_FuzzHarnessMaturityMeetsCP012Bar` (fuzz)
- **Owner:** prometheus

**T-0358** Capstone: record the v1-stable declaration go/no-go decision

> M19's exit criteria require all of: the CP-003 second-implementation match, full normative-statement-to-conformance-case coverage, and the licensing/stewardship/deprecation-window gate on record, before v1 is declared stable. Aggregate the results of every M19 gate task (role ceilings, both external trials plus any trial-miss rulings, the two-implementation conformance run, the traceability coverage audit, the feature register status, the licensing approval, and the fuzz-maturity report) into one decision document and get Eyvar's explicit go/no-go on declaring v1 stable, per CQ-012's own note that this gate 'blocks declaring v1 stable but not any individual coding task.'

- **Implements:** NFR-025, NFR-026, NFR-027, NFR-028, NFR-029, CON-019, CON-026
- **Depends on:** T-0344, T-0345, T-0347, T-0351, T-0353, T-0354, T-0356, T-0357, T-0044, T-0365
- **DoD:** A single dated decision document lists the pass/fail state of every M19 gate (NFR-025..029, CON-019, CON-026) and carries Eyvar's explicit go/no-go signature for declaring v1 stable; a no-go records the specific blocking gate(s) and the follow-up task(s) that will close them.
- **Test:** `TestM19_V1StableDeclarationGateRecorded` (integration)
- **Owner:** clio

**T-0365** Record Eyvar's ruling on an NFR-026 external-trial budget miss

> Mirrors T-0347's handling of the NFR-027 schedule risk. If T-0345's extracting-and-validating reader trial misses the 5-working-day budget or fails the extract+validate conformance corpus, that outcome alone does not close NFR-026's gate. Present the T-0345 result to Eyvar and obtain an explicit ruling: accept a revised day-budget as the new documented bound, descope part of the extracting-and-validating role, or accept the miss as a known limitation. Does not alter frozen spec.md text; only records a governance decision plan.md/CHANGES.md carries forward. Not applicable if T-0345 passes within budget.

- **Implements:** NFR-026
- **Depends on:** T-0345
- **DoD:** If T-0345 misses budget or fails corpus, Eyvar's ruling is recorded as a structured, dated entry (fields: date, decision-maker, decision, rationale) in plan.md Section 9 and specs/CHANGES.md; TestNFR_026_TrialMissRulingRecorded parses both files and asserts all four fields are present. If T-0345 passes within budget, this task is marked not-applicable with that reason logged in specs/CHANGES.md.
- **Test:** `TestNFR_026_TrialMissRulingRecorded` (integration)
- **Owner:** clio

---

## 4. Traceability summary

| Requirement class | Count | Tasks covering | Orphans |
|---|---|---|---|
| FR | 125 | 125 | 0 |
| NFR | 34 | 34 | 0 |
| CON | 26 | 26 | 0 |
| TR | 12 | 12 | 0 |

---

## 5. Residual gaps

- Three tasks carry empty implements arrays and are intentionally requirement-less, each justified in its own description: T-0359 (CP-012 continuous fuzz harness for untrusted-byte decode entry points), T-0097 (synthetic 1 GiB / 10,000-page benchmark corpus generator), T-0267 (governance task to obtain Eyvar/themis's design ruling on 8 missing accessibility/semantic mechanism fields, gating M15).
- Coverage concentration: FR-070 has 12 implementing tasks, FR-063 has 12, TR-012 has 15, FR-003 has 12 — verify these aren't double-billing the same work across split tasks vs. genuinely incremental build-up.
- Argus/security-review tasks (T-0112, T-0183, T-0205, T-0306, T-0362, T-0363) sit as terminal nodes after all substantive tasks with no modeled downstream remediation task if review fails findings — rework path is implicit, not graphed.
- CP-014 governance: FR-125's actual media-type/format-identification registry submission act (distinct from the in-repo magic-constant mechanism T-0006) has no task performing or tracking that external registration before v1-stable declaration.
- Section 6 below records two findings the M18/M19 patch passes judged mis-attributed to the wrong task during renumbering (T-0347's dependency-graph finding referencing extraction-metadata concerns unrelated to its actual subject; T-0356's stale id-collision note against a pre-renumbering raw id with no live counterpart) — flagged there rather than silently applied, and worth a human sanity check before this file is treated as final.
- M03 -> M07 is a real dependency at the task level (T-0056, T-0064 depend on M07's T-0118) that the milestone spine table in section 1 does not yet reflect (it lists M03 depending only on M01, M02) — the spine should be corrected to match before phase 5 (analyze) runs.
- Known limitation: the id-renumbering pass that gave every task its final `T-0NNN` id remapped only structured fields (`id`, `depends_on_tasks`). It did not rewrite free-text mentions of sibling tasks inside `description` bodies, so a handful of tasks still say "see T-1103" or similar using a pre-renumbering raw id instead of the task's current id. Section 6 documents 5 confirmed old-to-new pairs found this way (e.g. raw T-1103 is now T-0187); at least 50 more such mentions remain unverified and were not blindly remapped, because a wrong guess would silently point a reader at an unrelated task, which is worse than a reference that is visibly broken. Any description containing a bare `T-` id outside the `T-0NNN` format should be treated as referring to a sibling task in the same or an adjacent milestone by subject matter, not by that literal id, until a human or a dedicated pass resolves it.

---

## 6. Repair log

- Renumbered 365 tasks to a globally unique id space, closing 3 collision ranges (T-401-415, T-801-815, T-1801-1817) found by the traceability and dependency-graph audits.
- Remapped every depends_on_tasks reference into the new id space (0 could not be resolved and are listed as residual).
- Patched 19 milestones with a real finding: 16 missing dependency edges added, 15 DoD/test-kind issues rewritten, 10 orphan tasks given a real implements target or an explicit cross-cutting justification, duplicate coverage reviewed and consolidated where it was not genuine split work.
- 0 milestones had zero findings and were left untouched.
- dupLocal findings are false positives, not fixed: sibling_ids_raw (T-107/T-108/T-109/T-110, T-101/T-102/T-131/T-133, T-103/T-132, T-108/T-110/T-119/T-133) are the same tasks' OLD pre-renumbering ids (T-107=T-0007, T-101=T-0001, etc.), not distinct duplicate tasks. Verified by description: T-0007 (struct def), T-0008 (winner selection), T-0009 (self-digest validation), T-0010 (crash-atomicity suite) are genuinely split, non-overlapping work under FR-117 — same pattern holds for the NFR-001/NFR-020/CON-010 groups. Left all as-is per 'genuinely split work, leave both'. Task count unchanged from these findings.
- missingEdgesLocal, qualityLocal, orphanLocal were empty for this milestone — no changes made.
- Addressed the actionable note: added T-NEW-1, a Fuzz-kind conformance task covering CP-012 continuous fuzzing for M01's untrusted-byte decode boundaries (varint, TLV, Header, CommitRingRecord, Frontmatter, SegmentTable). It is intentionally requirement-less (implements: []) since CP-012 is a constitution principle, not a spec.md FR/NFR/CON/TR id; this is stated in its description per the orphan-handling instruction.
- Milestone task count: 32 original (unchanged) + 1 new (T-NEW-1) = 33 total.
- missingEdgesLocal fix: T-0033 now depends on T-0032 (M01 fixed-prefix full round-trip conformance suite), matching the resolved id from dependency_milestone_tasks_summary — T-0033's own text already required M01 Header/CommitRing/SegmentTable to round-trip; T-0032 is the exit-gate task that proves that.
- qualityLocal fix: T-0049 relabeled test_kind integration -> unit and its DoD rewritten to require an actual executable doc-lint check (ledger/doc_test.go asserting the design-note file exists and contains the two required section headers) instead of only 'a document exists'. No enum value in this schema means 'documentation/governance'; unit is the least-wrong fit for an automated presence/content check and keeps the task honestly testable rather than vacuously 'integration'.
- orphanLocal: none found for this milestone, no action taken.
- dupLocal: reviewed all 6 flagged requirement groups (FR-055, FR-056, FR-057, NFR-009, TR-010). In every case the sibling tasks are genuinely split work, not duplicates: FR-055 splits index-structure construction (T-0041) from wiring the bounded read path (T-0042); FR-056 splits core append-only placement (T-0033), the sealed-segment mutation guard (T-0046), and the fuzz harness (T-0047); FR-057 splits budget enforcement (T-0043) from boundary conformance vectors (T-0048); NFR-009 splits the CDC algorithm (T-0037) from its benchmark (T-0038); TR-010 splits the interface definition (T-0044), the refusal-path implementation (T-0045), and the FR-117-vs-TR-010 design note (T-0049). All kept, count unchanged (17 tasks), no merge performed.
- No new tasks added — findings.notes was empty and no finding in this milestone required a new task (e.g. no milestone-exit argus review was flagged here; that would need to come from an explicit finding).
- missingEdgesLocal fixed: T-0056 and T-0064 now depend_on_tasks include T-0118 (M07 referential-graph walker), reflecting that both the reachability audit and the fallback-cycle conformance vector need M07's cycle-detector/graph-walker to exist, not just the local M03 predecessor task.
- Cross-milestone spine gap: the milestone spine lists M03 as depending only on M01/M02, but T-0056/T-0064 now carry a real dependency on M07 (T-0118). That spine-level edge (M03 -> M07) is out of scope for this task-list patch — I can only fix it at the task level within M03. Flagging for whoever owns the milestone spine/dependency graph (zeus) to add M03->M07 to the spine so scheduling tools see the true build order.
- qualityLocal fixed: T-0061 (registry-review SLA) relabeled from test_kind 'conformance' to 'unit' since it verifies a checked-in documentation file's content (a doc-completeness assertion), not a wire-format/golden fixture. definition_of_done reworded to say so explicitly.
- orphanLocal: none reported for this milestone, no action taken.
- dupLocal reviewed for all listed groups (FR-012, FR-013, FR-014, FR-015, FR-107, CON-020, and the FR-109/FR-015 fallback-cycle pair): in every case the sibling tasks cover genuinely distinct mechanisms under one requirement (e.g. FR-012 splits into envelope codec T-0052, digest verification T-0053, position-key ordering T-0058, and conformance corpus T-0062; CON-020 splits into token type T-0050, frontmatter field T-0051, writer enforcement T-0060, and boundary vectors T-0065). None are literal duplicate work — kept all, no merge, task count unchanged at 17.
- No new tasks added (T-NEW-*): none of the findings required a brand-new task; the fixes were dependency edges, a test_kind/DoD relabel, and a no-op confirmation on dup groups.
- dupLocal reviewed (18 entries, all sharing sibling_ids_raw pairs/groups): every case is genuinely split work sharing one requirement, not true duplication. FR-019: T-0067 (mint primitive, unit) vs T-0084 (cross-file uniqueness conformance corpus) — different test kind/scope. FR-020: T-0069/T-0070/T-0071 (split/merge/move implementations) vs T-0085 (cross-operation fuzz exit-criterion) — different code paths and test_kind (unit vs fuzz). FR-027..FR-030: T-0077/T-0078/T-0079/T-0080 (per-requirement unit/integration implementation) vs T-0086 (four-requirement end-to-end reload conformance capstone, explicitly cross-referenced in T-0077's own description) — implementation vs capstone conformance, not redundant. FR-031: T-0081 (data-model field) vs T-0082 (validator rule) — model vs behavior. CON-001: T-0068 (arithmetic implementation) vs T-0083 (standing CI audit/lint) — implementation vs governance check. No merges performed; task count unchanged at 20 (T-0067..T-0086).
- Task-id-collision note (raw T-401..T-415 reused across M04/M05) is not actionable within this milestone's list: it describes a pre-renumbering id scheme visible only in sibling_ids_raw annotations. The milestone's actual task ids (T-0067..T-0086) are already globally unique, so the collision is resolved by the renumbering pass itself and requires no further change here. Flagging for whichever pass owns final id allocation to confirm M05's T-401..T-415 equivalents were likewise renumbered out of collision.
- missingEdgesLocal, qualityLocal, orphanLocal were empty for this milestone — no dependency, DoD/test, or orphan-requirement fixes required.
- No milestone-exit review task added: M04's scope (content identity, run split/merge/move, NFC scoping, annotation anchoring/orphaning, language-tag binding) touches no auth/payment/PII surface per constitution, so no argus gate is warranted. The existing momus conformance tasks (T-0083, T-0084, T-0086) and prometheus fuzz task (T-0085) already serve as this milestone's verify/exit gate.
- No new tasks added (T-NEW-1/2 not needed) — no genuine gap required one.
- orphanLocal fixed: T-0097 kept implements:[] but description now states explicitly why it is intentionally requirement-less (shared test-infrastructure/fixture generator under the ISO/IEC/IEEE 29119 test-infrastructure cross-cutting concern, feeding T-0098/T-0099/T-0100) — not converted to a fake FR/NFR claim.
- notes item on FR-044/045/046 CLI-wiring gap: T-0091 already flagged the M18 forward-wiring gap; T-0092 (FR-045) and T-0093 (FR-046) descriptions were missing the same flag-forward language and now carry it, so all three uncited-mechanism-gap tasks point M18's extract-verb owner at the same disclosed gap. No task or edge added outside this milestone since M18 is out of scope for this call.
- missingEdgesLocal and qualityLocal were empty for this milestone — no changes made there.
- dupLocal was empty — no duplicate tasks found in this milestone, no merge performed.
- Cosmetic-only: corrected stray legacy T-4xx references inside T-0097/T-0098/T-0099/T-0100 descriptions (leftover from pre-renumbering draft) to the actual T-0097/T-0098/T-0099/T-0100 ids so cross-references inside this milestone are internally consistent.
- Task count unchanged at 15; no new tasks added (findings did not genuinely require new governance/review tasks beyond description-level fixes).
- dupLocal resolved as false-positive, no change: every flagged sibling group is genuinely split work, not duplication. NFR-006 group (T-0103 Sign() impl, T-0110 sign-twice determinism test, T-0112 security review) covers implementation/test/audit of one requirement, not repeated work. CON-015 group (T-0102 type, T-0104 table, T-0105/T-0106/T-0107 steps 1-2/3-4/5, T-0108 orchestration, T-0109 conformance corpus, T-0111 allowlist gate, T-0112 review) is the natural decomposition of the 7-step verification procedure into independently testable units required by CON-015's own step structure in integrity.abnf S6 - collapsing any of these would hide a distinct DoD/test behind another task's. Left all 11 tasks unchanged; task count unchanged (11).
- missingEdgesLocal, qualityLocal, orphanLocal were empty for this milestone - no action taken in those categories.
- Not fixed (out of scope per findings list, flagging only): T-0106, T-0108, and T-0110 descriptions still reference stale pre-renumbering task IDs (T-604, T-604/605/606, T-607) instead of the current T-0105/T-0106/T-0107/T-0108 ids. These are dangling references left over from an id remap and should be swept in a later mechanical pass across all milestones, not patched task-by-task here since it wasn't a diagnosed finding for this call.
- No new governance/review task added - T-0112 already serves as the milestone-exit argus security review gate for both NFR-006 and CON-015, and it correctly depends on all 4 of its prerequisite tasks (T-0108, T-0109, T-0110, T-0111) so nothing downstream can depend on the eddsa package before that review lands.
- qualityLocal fix: T-0126's DoD was split. T-0126 now carries only the CI-checkable lint clause; the unverifiable 'note added to phase-5 analyze inputs' clause moved to new task T-NEW-1 (clio), which depends on T-0126.
- notes fix (actionable): added T-NEW-2 (clio) to record the NFR-030 memory-floor disclosed-conflict ruling request in clarify.md, mirroring the pattern used by sibling milestones (raw ids T-1106/T-1412/T-1501/T-1806). Depends on T-0121 since that task establishes the adopted bound being flagged.
- dupLocal — FR-102 (T-0114 vs T-0128), FR-103 (T-0113 vs T-0128), FR-106 (T-0115/T-0116 vs T-0128), FR-108 (T-0117 vs T-0128), FR-109 (T-0118/T-0119 vs T-0128), FR-110 (T-0120 vs T-0128), NFR-030 (T-0121 vs T-0122): all kept as-is. Each pair/group is genuinely split work — mechanism implementation vs fixture authoring vs end-to-end golden corpus vs benchmark-vs-fuzz test kind — not true duplicates. No tasks dropped for these.
- dupLocal — the two 'FR-109/FR-015 extension-envelope fallback-reference cycle vector' entries referencing raw id T-314 in another milestone (M03) are the same finding reported twice. This is a cross-milestone overlap I cannot resolve from within this milestone alone: the dependency-milestone summary provided to me does not include T-314's description, so I cannot confirm true duplication versus split scope (e.g. M03 authoring a general extension-envelope fallback-reference fixture vs M07's T-0119 authoring specifically the cycle-detection conformance vector required by plan.md Section 10 task (g), which explicitly forbids folding this into general work). I left T-0119 in place and added a note in its description flagging the possible overlap with raw id T-314 for phase-5 analyze (zeus) to reconcile with visibility into both milestones' full task text. This is not actionable by me alone since T-314/M03 tasks are outside my scope.
- missingEdgesLocal and orphanLocal were empty — no action needed.
- Task count: 16 original + 2 new (T-NEW-1, T-NEW-2) = 18 tasks returned for this milestone.
- missingEdgesLocal fixed: T-0137 and T-0143 both gained a dependency on structure_digest. The finding's resolved id 'T-0151' does not exist anywhere in this milestone or the dependency-milestone summary; the actual structure_digest task (raw T-811) is already present IN this milestone as T-0140, so both edges now point to T-0140 instead of the nonexistent id.
- qualityLocal fixed: T-0139's DoD, description and test_name rewritten to declare it the single canonical CoverageDescriptor wire implementation that M09's FR-063 call sites must import and reuse, rather than merging tasks across milestones (out of scope for this call -- M09 is handled by a parallel call and should add a dependency on this task's output on its side).
- dupLocal: no drops made. The FR-001/FR-002/FR-003/TR-009 clusters list many of THIS milestone's own tasks as 'siblings' of themselves because those requirements are broad and span the whole T_S/T_C/structure_digest pipeline (constants -> leaf encoding -> traversal order -> internal builder -> root -> commit wiring -> conformance suites); each listed task does genuinely distinct work, so all were left in place. The one cross-milestone true duplicate (T-0139 vs M09's coverage-descriptor tasks) was addressed via the qualityLocal fix above, not by deleting a task in this milestone.
- orphanLocal: empty, no action needed.
- Added T-NEW-1: consolidated milestone-exit argus security review task, matching the M06/M10/M11/M16 pattern the findings flagged as missing for M08.
- Not actionable here: the T-801..T-815 raw-id collision between old M08 and old M09 numbering is a source-data artifact of the pre-renumbering task lists; this milestone's tasks already carry unique T-0129..T-0144 ids, so no further action was taken on that note within this call's scope.
- missingEdgesLocal fixed: T-0152 now depends_on_tasks=[T-0136,T-0151] (was [T-0151]); T-0158 now depends_on_tasks=[T-0136,T-0151,T-0153,T-0155,T-0157] (was [T-0153,T-0155,T-0157]). Both consume T_C_root/structure_digest fresh from M08's T-0136, cited only implicitly in prose before.
- qualityLocal: the finding's task_new_id (T-0154, PresentationArtefact) did not match its own issue text (CoverageDescriptor/PD-COVER-001..004, i.e. T-0145's group) -- applied the fix to T-0145 instead, since that is the task whose content the issue actually describes. T-0145 now depends_on_tasks=[T-0139] (M08's CoverageDescriptor task) and its description/DoD state explicitly that it reuses T-0139's type rather than re-declaring CoverageDescriptor a second time under FR-063. T-0146..T-0149 already depend on T-0145 so they inherit this without further edits. This resolves the cross-milestone FR-002/FR-063 duplicate flagged in dupLocal group D without touching M08's task list (out of scope for this call).
- dupLocal: FR-063 group (T-0145..T-0153,T-0158,T-0166), FR-064 group (T-0153..T-0155), and FR-067 group (T-0158,T-0161) are genuinely split work (descriptor vs digest vs preimage vs record vs verify vs corpus) -- left as-is, no tasks dropped. The FR-002/FR-063 CoverageDescriptor group is addressed via the T-0145 dependency/description fix above rather than dropping a task, since M08's T-0139 is out of this milestone's authority to remove or merge.
- orphanLocal: none reported, no action.
- notes-actionable: added T-NEW-1, a milestone-exit argus review task for M09, mirroring M06's T-0112 pattern, per the reported gap. Depends on the milestone's security-relevant tasks (T-0153,T-0158,T-0159,T-0161,T-0164,T-0165,T-0166) and cites FR-063/FR-067/FR-115/NFR-005, all honestly within this milestone's covered requirement set.
- Task count: 22 original + 1 new (T-NEW-1) = 23 returned; no tasks were dropped since the only true duplicate (CoverageDescriptor vs M08's T-0139) was resolved by an explicit dependency/reuse edge rather than deletion, per the fix option offered in qualityLocal.
- missingEdgesLocal fixed: T-0169 depends_on_tasks now includes T-0138 (M09 Signature entity) and T-0153 (SIGNATURE record wire struct, this milestone's own dependency-milestone task) in addition to its existing T-0167/T-0168 deps.
- qualityLocal and orphanLocal were empty for this milestone - no changes made.
- dupLocal analyzed: sibling_ids_raw across all 22 dupLocal entries are exactly T-1001..T-1018 (18 distinct ids), matching 1:1 the count and FR-070/071/072/073 scope of this milestone's 18 tasks (T-0167..T-0184). This is not genuine duplicate work - it is the same milestone's task set referenced under a prior/old id scheme (T-10xx) before renumbering to T-016x. No sibling task exists outside this list to merge or drop, so task count stays 18 (no merge performed). Internal cross-references inside descriptions/DoD that still pointed at the old T-10xx numbering (e.g. T-1005, T-1006, T-1012, T-1014) were corrected to the corresponding T-016x ids while making the T-0169 edit pass, to prevent this same false-positive from recurring downstream.
- No new governance/review task added: T-0183 (argus security review) and T-0179/T-0181 (negative conformance corpora) already provide milestone-exit verification coverage for FR-070..073; nothing in findings called for an additional task.
- Fixed missingEdgesLocal: T-0185 depends_on_tasks now includes T-0138, T-0153 (was empty) so the designation API's dependency on M08's T_C tree primitives and M09's Signature/SIGNATURE-record flow is now explicit.
- Fixed qualityLocal on T-0190: replaced the fake Go-unit-test framing with a governance-checklist framing. test_name changed from 'TestGOV_FR075_SaltedCommitmentRulingRecorded' to 'gov-checklist-FR061-FR075-ruling-recorded'; definition_of_done rewritten to describe a CI doc-lint grep of clarify.md rather than an implied code test, since this task changes no code. test_kind left as 'integration' (closest fit in the fixed enum) but DoD now matches that reality.
- orphanLocal: empty, no action needed.
- dupLocal: reviewed every FR-074..081 sibling group. All are genuinely split work, not literal duplicates -- e.g. under FR-074, T-0185/T-0186/T-0187 build distinct pipeline stages (designate -> salt -> commitment digest), T-0188 is a general round-trip conformance corpus while T-0206 is specifically the at-limit/one-past-limit boundary corpus for the same wire shape, and T-0205 is the cross-cutting security review referencing all of them. Kept all 22 tasks unchanged in count; no merge performed.
- Note on cross-milestone risk (H(0x02\|\|salt\|\|canon(subtree)) also referenced by M12's T-1210/ErasureRecord with no shared dependency): not actionable from this milestone alone since M12 is owned by a separate parallel call; flagging here only so the M12 pass (or a later cross-milestone analyze step) can add a shared dependency or extract the formula into a common helper referenced by both milestones instead of re-specifying it.
- No new tasks added -- none of the findings required a new task; T-0205 already serves as the milestone-exit argus review.
- missingEdgesLocal fixed: T-0217 depends_on_tasks now [T-0209, T-0187] -- reuses M11's salted-commitment digest formula (T-0187, raw T-1103) instead of reimplementing it; description and DoD updated to say so explicitly.
- qualityLocal: empty, no changes.
- orphanLocal: empty, no changes.
- dupLocal reviewed for all 7 requirement groups (FR-005, FR-059, FR-060, FR-061, NFR-032, NFR-033, CON-022): in every case the task_new_id entries do genuinely-split, non-overlapping work under the same umbrella requirement (e.g. FR-005: T-0210 wires the predecessor chain, T-0211 computes NCA from it -- different deliverables, sequential dependency, not duplicates; NFR-032: T-0221 tunes to hit the budget, T-0222 locks in the disclosed miss as a regression guard -- different deliverables). No task count reduced; count stays at 18.
- No new tasks added -- your_findings.notes was empty and no finding here independently justified a new governance/review task within this milestone's scope.
- dupLocal (FR-092: T-0225/T-0226/T-0227/T-0229/T-0231/T-0232 vs raw T-1301/1302/1303/1305/1307/1308): not a true duplicate. Old-numbering ids map 1:1 onto these six new-numbering tasks (taxonomy skeleton, disjoint-commute, R1, R2, R3, exhaustive test) — six distinct sub-implementations of one requirement, which is expected for a classifier requirement with multiple rules. No merge performed.
- dupLocal (FR-093: T-0227 vs T-0228, raw T-1303/1304): genuinely split — T-0227 implements the R1 rule itself, T-0228 is the separate property-fuzz target over that rule. Left both.
- dupLocal (TR-002: T-0239 vs T-0241, raw T-1315/1317/1808(M18)): genuinely split — T-0239 is the diff engine unit, T-0241 is the end-to-end orchestrator integration test that exercises it; T-1808(M18) is a cross-milestone CLI exit-code task, out of scope here. Left both.
- dupLocal (TR-003: T-0240 vs T-0241, raw T-1316/1317/1809(M18)): same reasoning as TR-002 — implementation task vs integration test, genuinely split. Left both.
- Fixed stale cross-reference: T-0230's description cited 'T-1305' (old numbering) for the R2 comparator it reuses; corrected to T-0229. T-0232's description cited 'T-1301 taxonomy'; corrected to T-0225. T-0241's description cited old-numbering task lists (T-1302/1303/1305/1307, T-1309..1313, T-1314, T-1315, T-1316); corrected to the actual new task ids it depends on.
- notes item on CON-010/CP-011 wire-format conformance gap: added T-NEW-1 covering at-limit/over-limit boundary fixtures for the two guard requirements this milestone actually owns (CON-024, CON-025). Flagging rather than silently claiming CON-010: that requirement id is not in this milestone's assigned set, so its milestone-ownership question should be confirmed against the concern registry, not resolved unilaterally here.
- No missingEdgesLocal, qualityLocal, or orphanLocal findings were reported for this milestone — none required action.
- Task count: 17 -> 18 (one addition, T-NEW-1; no merges performed).
- qualityLocal fix T-0242: DoD reworded from 'fuzz-style' randomized test to a fixed table-driven test (insert/remove at start/middle/end/single-entry) so DoD matches test_kind=unit; no fuzz infra required for this task.
- qualityLocal fix T-0253: test_kind changed integration->conformance and DoD reworded into an automated, checkable assertion (clarify.md carries a non-empty EX-001 record: scope/expiry/approver/date) instead of an unrunnable 'Eyvar's sign-off' condition. The blocking semantics (NFR-021/T-0252 stay blocked absent the record) are preserved.
- orphanLocal fix T-0254: implements set to [NFR-017] — this is the conformance-vector precursor artifact that T-0255's PLP-1 decoder (NFR-017) is built and graded against; it has no FR/NFR of its own but is not requirement-less, it is upstream of NFR-017.
- orphanLocal fix T-0264: implements set to [FR-089, NFR-010] — the at-limit/over-limit fixtures directly target the two render-layer record shapes this milestone owns (FontRecord/FR-089, PageDirectory entry/NFR-010); this is the CP-011/CON-010 cross-cutting obligation applied to those two shapes specifically, not a free-floating task.
- orphanLocal fix T-0265: implements set to [NFR-017] — continuous fuzzing is the CP-012 cross-cutting obligation applied to the two untrusted-byte decoders (PLP-1, restricted-PNG) that themselves implement NFR-017; tied to that requirement rather than left orphaned.
- orphanLocal fix T-0266: implements set to [CON-006] — this task exists solely to manage the blast radius of the CON-006/CQ-006 shaping-oracle exception (EX-001); it is the isolation mechanism that lets the rest of the milestone ship independent of that exception's resolution, so CON-006 is the correct (and only honest) anchor.
- dupLocal NFR-017 (T-0255 PLP-1 decoder, T-0256 restricted-PNG decoder, T-0262 open+render memory-budget benchmark): kept all three, not merged. They are genuinely split work — two distinct untrusted-byte decoders plus a separate end-to-end memory benchmark that exercises both plus the rest of the render path — not duplicate coverage of the same slice of NFR-017.
- missingEdgesLocal: none reported for this milestone, no dependency edges added.
- No new tasks added: none of the findings required a new governance/review task beyond edits to existing tasks.
- Applied missingEdgesLocal: T-0275 now depends on T-0274 plus M08 tasks T-0134/T-0135/T-0149/T-0150 (union of the two candidate sets from findings, since both refer to the same M08 T_C-traversal retrofit); T-0276 now depends on T-0274 plus M05 task T-0087.
- Caveat on the above: T-0134/T-0135/T-0149/T-0150/T-0087 do not appear in the dependency_milestone_tasks_summary handed to this pass (it only lists M04 tasks T-0067..T-0086) — cannot independently verify these ids from what I was given. Recommend the analyze/zeus pass confirm exact M08 and M05 task ids and add M08 and M05 as explicit spine dependencies for M15 (self-disclosed by T-0275/T-0276's own task text).
- Fixed qualityLocal on T-0295: DoD now requires a named decidability-classification rubric (a PD-rule id must fully decide pass/fail with no human judgment) before the >=0.80 fraction is computed, making the check mechanically verifiable.
- Fixed orphanLocal on T-0267: left implements empty by design and added an explicit sentence to its description stating it is a cross-cutting SDD phase-gate governance task (CP-011/AD-002 compliance) gating the 16 FR/NFR ids named in its own text, not an implementation of any single requirement.
- Reviewed all dupLocal entries: every sibling group is genuinely split work (separate schema-definition / grammar-encoding / validator-rule / retrofit tasks per requirement), not a literal duplicate. Left both/all siblings as-is; no merge, no task-count change (still 29 tasks, 0 added, 0 removed).
- Renumbered cross-references inside descriptions that pointed to the raw T-15xx/T-13xx numbering (e.g. 'T-1501 ruling', 'T-1506', 'T-1508', 'T-1509/T-1510', 'T-1518', 'T-1526') to this milestone's actual T-02xx ids so the task list is internally consistent.
- Not actionable here: the note about M15's overall spine gap (M08/M05 missing as milestone-level dependencies) is a cross-milestone graph edit outside a single milestone's task-list patch — flagged above for the analyze/zeus integration pass rather than fixed unilaterally.
- missingEdgesLocal, qualityLocal, orphanLocal: all empty for this milestone, no changes made.
- dupLocal FR-119 (T-0298 vs T-0299): not a true duplicate. T-0298 implements the deterministic Transform() itself (unit-tested in-process); T-0299 publishes the external-trial conformance corpus consumed by a second implementation per CP-003/NFR-028. Distinct deliverables, distinct test_kind (unit vs external-trial). Kept both, no merge.
- dupLocal FR-121 (T-0296 vs T-0297): not a true duplicate. T-0296 implements Scan()/RefusalReport; T-0297 authors the golden conformance corpus (>=5 fixture classes) that exercises it in CI, per CON-010 fixture discipline. Distinct deliverables, distinct test_kind (unit vs conformance). Kept both, no merge.
- dupLocal CON-016 (T-0303, T-0304, T-0305, T-0306): not duplicates. Four non-overlapping deliverables under one requirement: T-0303 hash-family re-protection (no re-sign), T-0304 the operator-invoked RESCIND-AND-RESIGN scheme-break flow, T-0305 wire-format conformance vectors for the resulting RescindResignRecord, T-0306 the mandatory pre-merge argus security review of migrate/+registry/. Kept all four, no merge.
- No new tasks added: T-0306 already functions as the milestone-exit argus review gate for migrate/ and registry/ (depends on T-0298, T-0302, T-0303, T-0304), so a separate governance/review task would be redundant.
- missingEdgesLocal fix: T-0318 now depends on T-0196 (publish-time delegation target) and T-0298 (Phase 2 migration transform, the concrete Migrate call site) in addition to its existing T-0308 dependency. The finding's resolved list only named T-0196; T-0298 was added because the task description explicitly targets 'the migration engine's L*(migrate(state)) fresh-file emission (M16)', and T-0298 is that engine's core transform - without it migrate.go does not exist to verify against.
- qualityLocal fix: T-0316 DoD reworded to drop the '2 OS/arch build combinations' requirement (no second target exists per the stated single-toolchain constraint: Go 1.25.1, darwin/arm64) in favor of 2 independent same-platform runs.
- qualityLocal fix: T-0317 test_kind changed from 'conformance' to 'unit' and test_name changed from 'CONF-DEGENERATE-002-...' to 'TestFR_124_EmptyStateOutcomeTableSchemaComplete', since this task publishes a reference table and its own CI schema check, not a fixture-driven conformance corpus.
- dupLocal: reviewed all 19 flagged pairs. The sibling_ids_raw (T-1701, T-1702, T-1709...T-1810 etc.) are not present anywhere in the actual milestone task lists (ours or the dependency summary) - they are old pre-renumbering IDs that this milestone's own task descriptions still reference as self-citations (e.g. T-0308's description says 'using T-1701's traversal', where T-1701 is T-0307 itself under its old number). No genuine duplicate tasks exist among T-0307..T-0324; every flagged group is legitimate split work (traversal core vs streaming impl vs conformance corpus vs call-site verification, etc.). No merges made; task count unchanged at 18.
- notes item (CON-010/CP-011 wire-format ceiling vectors for M17): not actionable within this milestone's mandate. CON-010 is not among this milestone's covered requirement ids (FR-124, NFR-002, NFR-004, TR-004, TR-005), and the established pattern in the dependency milestone list is that each package owning a newly-introduced structural ceiling authors its own CON-010 at-limit/over-limit fixtures (e.g. T-0018, T-0032 for M01's frame/header ceilings, T-0116 for M07-owned ceilings). M17 (canon) does not introduce new ceiling-bounded wire structures of its own - it only traverses/serializes structures whose ceilings are already owned and fixture-tested elsewhere. Recommend this be raised against whichever milestone actually owns any canon-adjacent ceiling value, not added here; no new task created for it, and the T-NEW budget was left unused since no genuine gap in this milestone's own scope was found.
- missingEdgesLocal fixed: T-0330 (extract verb) now depends_on T-0091 (FR-044 extraction-view metadata), T-0092 (FR-045 page-adjunct emission), T-0093 (FR-046 page-adjunct staleness) instead of the given resolved list ['T-0071','T-0091'], which was wrong -- T-0071 is Run split/merge/move (FR-020), unrelated to extract's metadata/page-adjunct surface. T-0091/92/93 are the actual M05 tasks the CLI wiring needs already built.
- missingEdgesLocal fixed: T-0341 (--if-match flag) now depends_on T-0044 (M02 ConditionalWriter interface), the interface this CLI flag is the usage surface for.
- qualityLocal NOT applied for T-0328, T-0330, T-0332, T-0338, T-0339: the described defects (NFR-026 5-day external-reader-trial governance, plan.md ruling record, funding/sponsor decision record, licence-document drafting task, human sign-off record, 'same as T-1106') do not correspond to anything in these tasks' actual descriptions/DoDs in this milestone (T-0328 is the 5-field finding schema, T-0330 is the extract verb, T-0332 is diff, T-0338 is migrate, T-0339 is the traceability table -- none mention NFR-026, external trials, licences, or sign-off). This looks like stale/misattributed analysis from a different task set carried over during renumbering. Rather than inject fabricated governance content onto unrelated tasks, DoD/test_name/test_kind were left unchanged and this mismatch is flagged as a genuine cross-artifact conflict per instructions -- whoever owns the real T-1106-pattern tasks (if they exist) should get this finding instead.
- dupLocal TR-002/TR-003/TR-004/TR-005/TR-010: kept both sides in every case -- these are genuinely split work, this milestone's task being the CLI-boundary wiring (diff, merge, project, --if-match) and the cited siblings being the core engine/interface tasks in M13/M17/M02 that the CLI task explicitly depends on. No merge.
- dupLocal TR-012 (15 rows spanning T-0325..T-0340 against the same M18 sibling set): no merge -- these are 15 genuinely distinct tasks (dispatch skeleton, exit-code engine, 11 separate verbs, traceability table, conformance suite), not duplicates of each other. The underlying issue is a raw source-id collision (T-1801..T-1817 reused wholesale between an M18 draft and an M19 draft before renumbering); this milestone already resolves it by giving every CLI task a unique id (T-0325-T-0341). Any remaining collision on the M19 side (governance/conformance tasks that reused the same raw numbers) is out of scope for this milestone and should be handled by the M19 pass.
- orphanLocal: empty, nothing to fix.
- No new tasks added -- none of the surviving findings required a genuinely new governance/review task within this milestone's TR-001/TR-012 scope.
- missingEdgesLocal applied: T-0347.depends_on_tasks now includes T-0071, T-0091 (mechanical per finding). Flagging for zeus review: the finding's stated rationale ('extract verb never picked up FR-044/045/046 metadata/page-adjunct emission') describes extraction-metadata concerns that do not match T-0347's actual subject (recording Eyvar's NFR-027 schedule-risk ruling). This looks like a mis-attributed finding, possibly intended for T-0330 (extract verb command) in the dependency milestone. Edge added as instructed but should be sanity-checked before being treated as load-bearing.
- missingEdgesLocal applied: T-0358.depends_on_tasks now includes T-0044 (ConditionalWriter interface, cited as the CLI --if-match usage surface).
- qualityLocal: rewrote T-0347, T-0349, T-0355, T-0356 definitions_of_done to specify structured, machine-checkable fields (date/decision-maker/decision/etc.) so their integration-kind tests can actually assert repo state instead of only trusting a narrative human sign-off.
- qualityLocal: T-0345's DoD no longer treats a budget/corpus miss as a self-sufficient acceptable outcome — it now requires the new T-NEW-1 governance-ruling task (mirroring T-0347) to be filed and open before T-0345 can close on a miss.
- qualityLocal: T-0356's flagged id collision with 'CLI's T-1815' is against a stale pre-renumbering raw id, not present anywhere in the current final id space (this milestone's T-0342..T-0358 vs. the dependency milestone's T-0001..T-0341 both being unique) — no live collision exists, so no renumbering was performed.
- dupLocal: reviewed every listed sibling group (NFR-011, NFR-025, NFR-026, NFR-027, NFR-028, NFR-029, CON-019, CON-026). All are genuinely split pipeline work — distinct verbs across the group (define protocol / fund+scope / build harness / commission+run / record ruling / aggregate capstone) rather than restatements of the same task — so no tasks were merged or dropped; task count only grew by the one new governance task below.
- Added T-NEW-1 (owner clio, NFR-026) as the missing budget-miss governance-ruling task analogous to T-0347/NFR-027, and wired it into T-0345's DoD and T-0358's dependency list so the capstone go/no-go decision accounts for it.
- orphanLocal: empty in findings — no action needed.
</content>
