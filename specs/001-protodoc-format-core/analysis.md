# Protodoc: Cross-Artifact Analysis

Status: APPROVED — Eyvar, 2026-09-15 | Spec ID: 001-protodoc-format-core | Phase: 5 (analyze) | Date: 2026-09-14

---

## 1. Scope and method

This phase checks consistency across all phase 0-4 artifacts together: the constitution, `spec.md`, `clarify.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`, and `tasks.md`. Per the global workflow, the gate is: no orphan requirement, no orphan task, no constitution violation.

Two gaps `tasks.md` itself disclosed in section 5 were resolved as part of this phase rather than carried forward again: the M03 milestone-dependency omission (GAP A) and stale pre-renumbering task ids inside task description prose (GAP B). Both are closed below.

Three cross-check lenses ran in parallel over the full artifact set (orphan detection, constitution-violation detection, contract/data-model/task consistency), and every blocker- or major-severity finding was then adversarially re-reviewed by a skeptic instructed to refute it — a finding survives only if the skeptic could not explain it away as stale, already-resolved, or hallucinated.

---

## 2. GAP A: milestone spine correction

`tasks.md` section 1's milestone table listed M03 depending on M01, M02 only. Two of M03's tasks (T-0056, T-0064) were patched during phase 4 repair to depend on M07's T-0118 (the reference-graph cycle detector), which the milestone table never reflected. Corrected dependency list for M03: **M01, M02, M07**.

The cross-check contract-consistency lens confirmed M07 is actually capable of supplying what M03 needs (see finding log below if this was flagged as anything other than resolved).

---

## 3. GAP B: stale task-id references resolved

42 pre-renumbering id citations were found in task description prose. 42 were resolved with confirmed or high confidence; 0 remain unresolved and are carried forward below rather than guessed.

### 3.1 Resolved

| In task | Raw old id cited | Resolved to | Confidence | Evidence |
|---|---|---|---|---|
| T-0051 | T-300 | T-0050 | high | T-0050 is titled 'ext-token type: encode/decode + owner-id partition', exact phrase match; M03 block offset raw-250=new confirmed by multiple other pairs. |
| T-0052 | T-304 | T-0054 | high | T-0054 'ext-disposition closed 3-value enum' = the disposition value's decode; M03 offset -250 (304-250=54). |
| T-0052 | T-307 | T-0057 | high | T-0057 'PD-EXT-003: fallback minimum-content check' fits the fallback-check range; offset -250 (307-250=57). |
| T-0052 | T-303 | T-0053 | high | T-0053 'ext-payload-digest verification' is literally 'the digest verification'; offset -250 (303-250=53). |
| T-0055 | T-306 | T-0056 | high | T-0056 'Fallback reachability audit' is exactly the 'reachability' half named; offset -250 (306-250=56). |
| T-0059 | T-306 | T-0056 | high | Same fallback-reachability task ('the audited fallback') as resolved for T-0055; offset -250 consistent. |
| T-0059 | T-304 | T-0054 | high | 'disposition value's own decode' matches T-0054's ext-disposition enum; offset -250 (304-250=54). |
| T-0060 | T-301 | T-0051 | high | T-0051 'Frontmatter.fm-retired-tokens field' is the literal 'fm-retired-tokens' referent; offset -250 (301-250=51). |
| T-0066 | T-312 | T-0062 | high | T-0062 'Conformance suite: one round-trip fixture per disposition' is a conformance-fixture source; offset -250 (312-250=62). |
| T-0066 | T-313 | T-0063 | high | T-0063 'missing-disposition envelope rejected' conformance vector; offset -250 (313-250=63). |
| T-0066 | T-315 | T-0065 | high | T-0065 'owner-id partition boundary conformance vectors'; offset -250 (315-250=65). |
| T-0082 | T-415 | T-0081 | high | T-0081 'Wire exactly-one-language-tag field' is the field PD-LANG-001 validates; M04 offset -334 (415-334=81), same offset confirmed by T-402 pair. |
| T-0083 | T-402 | T-0068 | high | T-0068 'base_ordinal ... positioning arithmetic' is 'the field's own arithmetic'; M04 offset -334 (402-334=68). |
| T-0106 | T-604 | T-0105 | high | T-0105 'Steps 1-2 canonical/small-order check for A' is what R's checks 'mirror'; M06 offset -499 (604-499=105), confirmed by T-607/T-0108 pair. |
| T-0108 | T-604 | T-0105 | high | First of 'Steps 1-5' composed = Steps1-2/A = T-0105; offset -499 (604-499=105). |
| T-0110 | T-607 | T-0108 | high | T-0108 'Verify() 7-step orchestration' is literally '...via T-607's Verify()'; offset -499 (607-499=108). |
| T-0119 | T-314 | T-0064 | high | 'another milestone's inventory' = M03; T-0064 'extension-envelope fallback-reference cycle rejected' matches this task's own subject; M03 offset -250 (314-250=64). |
| T-0129 | T-801 | T-0130 | high | T-0130 is the T_S tree builder fed by this constants task; M08 offset -671 (801-671=130), confirmed by T-0131 pair. |
| T-0129 | T-803 | T-0132 | high | T-0132 'T_C leaf-redactable node encoding' is first of the T_C builder chain fed; offset -671 (803-671=132). |
| T-0129 | T-807 | T-0136 | high | T-0136 'T_C_root end-to-end computation' is last of the fed chain; offset -671 (807-671=136). |
| T-0130 | T-800 | T-0129 | high | T-0129 defines ABSENT_CHILD_DIGEST itself, the constant referenced; offset -671 (800-671=129). |
| T-0131 | T-801 | T-0130 | high | T-0130 is the T_S builder whose output T-0131 wires into TSRoot; offset -671 (801-671=130). |
| T-0133 | T-803 | T-0132 | high | T-0132 'leaf-redactable node encoding' is the sibling task whose tc-canon rule is reused; offset -671 (803-671=132). |
| T-0134 | T-806 | T-0135 | high | T-0135 'T_C internal node encoding builder' is a traverser of this ordinal order; offset -671 (806-671=135). |
| T-0134 | T-807 | T-0136 | high | T-0136 'T_C_root computation' also traverses in this order; offset -671 (807-671=136). |
| T-0135 | T-805 | T-0134 | high | T-0134 'T_C subtree traversal order' is the source of the 'subtree ordinal order' used here; offset -671 (805-671=134). |
| T-0135 | T-803 | T-0132 | high | T-0132 leaf-redactable digest source; offset -671 (803-671=132). |
| T-0135 | T-804 | T-0133 | high | T-0133 leaf-nonredactable digest source, paired with T-803/T-0132; offset -671 (804-671=133). |
| T-0136 | T-806 | T-0135 | high | T-0135 is the internal-node builder wired into TCRoot; offset -671 (806-671=135). |
| T-0136 | T-805 | T-0134 | high | T-0134 traversal order defines subtree ordinal 0 referenced here; offset -671 (805-671=134). |
| T-0136 | T-809 | T-0138 | high | T-0138 'ATTEST-segment exclusion predicate' is what excludes ATTEST-typed segments here; offset -671 (809-671=138). |
| T-0137 | T-807 | T-0136 | high | T-0136 'T_C_root end-to-end computation' is 'the tree computation' being closed the loop from; offset -671 (807-671=136). |
| T-0137 | T-811 | T-0140 | confirmed | Text itself states 'structure_digest (T-811 / T-0140)', an explicit pairing. |
| T-0138 | T-806 | T-0135 | high | T-0135 internal-node builder, paired with T-807/T-0136 as the traversers into T_C's subtree set; offset -671 (806-671=135). |
| T-0138 | T-807 | T-0136 | high | T-0136 T_C_root computation, the other traverser named; offset -671 (807-671=136). |
| T-0138 | T-810 | T-0139 | high | T-0139 'CoverageDescriptor wire encode/decode' is literally 'CoverageDescriptor construction'; offset -671 (810-671=139). |
| T-0139 | T-809 | T-0138 | high | T-0138's ATTEST-exclusion predicate is reused for PD-COVER-004; offset -671 (809-671=138). |
| T-0140 | T-802 | T-0131 | high | T-0131 'T_S_root end-to-end computation' is the 'T_S root' recomputed here; offset -671 (802-671=131). |
| T-0141 | T-811 | T-0140 | high | T-0140 is the structure_digest fresh-recomputation task invoked here; offset -671 (811-671=140), same referent explicitly confirmed elsewhere as T-811/T-0140. |
| T-0141 | T-802 | T-0131 | high | T-0131 T_S_root recomputation, invoked inside T-0140's algorithm; offset -671 (802-671=131). |
| T-0143 | T-813 | T-0142 | high | T-0142 'two-run determinism of T_S_root and T_C_root' is exactly 'T-813's determinism check'; offset -671 (813-671=142). |
| T-0143 | T-811 | T-0140 | confirmed | Text itself states 'structure_digest (T-811 / T-0140)', an explicit pairing. |

### 3.2 Unresolved (carried forward, not guessed)

None.

---

## 4. Traceability matrix

Full requirement -> task -> test mapping. `files` is empty for every row because phase 6 (implement) has not started; this column is populated as code lands, per requirement, so this same matrix stays the living traceability record through implementation.

| Requirement | Kind | Tasks | Tests | Files |
|---|---|---|---|---|
| FR-001 | FR | T-0136, T-0137, T-0362 | `TestFR_001_TCRootIsCompleteStateValueSet`, `TestFR_003_CommitRingStateIDMatchesTCRoot`, `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` | (pending phase 6) |
| FR-002 | FR | T-0138, T-0139, T-0140, T-0362 | `TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare`, `TestFR_002_CoverageDescriptorCanonicalImplementationRejectsEachPDCoverViolation`, `TestFR_002_StructureDigestPreimageMatchesBitmaskExactly`, `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` | (pending phase 6) |
| FR-003 | FR | T-0129, T-0132, T-0133, T-0134, T-0135, T-0136, T-0137, T-0142, T-0143, T-0144, T-0362 | `TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed`, `TestFR_003_TCLeafRedactableUsesStoredFrameVerbatim`, `TestFR_003_TCLeafNonredactableDomainSeparatedFromRedactable`, `TestFR_003_TCTraversalOrderDeterministicUnitIDLexicographic`, `TestFR_003_TCTreeFixedShapeAndAbsentChildFill`, `TestFR_001_TCRootIsCompleteStateValueSet`, `TestFR_003_CommitRingStateIDMatchesTCRoot`, `TestFR_003_TSAndTCRootsDeterministicAcrossTwoRuns`, `TestFR_003_RootChangesForEveryCoveredValueMutation`, `TestTR_009_TSTreeAtLimitAndOnePastLimit`, `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` | (pending phase 6) |
| FR-004 | FR | T-0007 | `TestFR_004_ParentStateIdRoundTrip` | (pending phase 6) |
| FR-005 | FR | T-0210, T-0211 | `TestFR_005_PredecessorChainWalkable`, `TestFR_005_NearestCommonAncestorOffline` | (pending phase 6) |
| FR-006 | FR | T-0003 | `TestFR_007_HeaderContentIndependent` | (pending phase 6) |
| FR-007 | FR | T-0003 | `TestFR_007_HeaderContentIndependent` | (pending phase 6) |
| FR-008 | FR | T-0004 | `TestFR_010_CapabilityRequiredExceedsWritten` | (pending phase 6) |
| FR-009 | FR | T-0004 | `TestFR_010_CapabilityRequiredExceedsWritten` | (pending phase 6) |
| FR-010 | FR | T-0004 | `TestFR_010_CapabilityRequiredExceedsWritten` | (pending phase 6) |
| FR-011 | FR | T-0005 | `TestFR_011_DurableClaimReadFromHeader` | (pending phase 6) |
| FR-012 | FR | T-0052, T-0053, T-0058, T-0062 | `TestFR_012_ExtEnvelopeRoundTrip`, `TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition`, `TestFR_012_ExtPositionKeyCanonicalOrder`, `TestFR_012_013_DispositionRoundTripCorpus` | (pending phase 6) |
| FR-013 | FR | T-0054, T-0062 | `TestFR_013_014_ExtDispositionStructuralReject`, `TestFR_012_013_DispositionRoundTripCorpus` | (pending phase 6) |
| FR-014 | FR | T-0054, T-0063, T-0066 | `TestFR_013_014_ExtDispositionStructuralReject`, `TestFR_014_MissingDispositionRejectedNamingToken`, `FuzzExtEnvelopeDecode` | (pending phase 6) |
| FR-015 | FR | T-0055, T-0056, T-0064 | `TestFR_015_PD_EXT_002_FallbackMandatoryWhenIgnoreOrDegrade`, `TestFR_015_FallbackReachabilityExclusion`, `TestFR_015_FR_109_ExtFallbackCycleRejected` | (pending phase 6) |
| FR-016 | FR | T-0057 | `TestFR_016_PD_EXT_003_FallbackMinimumContent` | (pending phase 6) |
| FR-017 | FR | T-0039 | `TestFR_017_OpaqueConstructPreservedOrRefused` | (pending phase 6) |
| FR-018 | FR | T-0040 | `TestFR_018_OpaqueSegmentSurvivesTwoDivergentCopies` | (pending phase 6) |
| FR-019 | FR | T-0067, T-0084 | `TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded`, `TestFR_019_IdentifierUniqueAcrossCopyForkBranch` | (pending phase 6) |
| FR-020 | FR | T-0069, T-0070, T-0071, T-0085 | `TestFR_020_SplitPreservesRunID`, `TestFR_020_MergeIsPureSyntacticPredicate`, `TestFR_020_MoveReorderPreservesRunID`, `TestFR_020_FuzzRunIDPreservationAcrossOperationSequence` | (pending phase 6) |
| FR-021 | FR | T-0067 | `TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded` | (pending phase 6) |
| FR-022 | FR | T-0072 | `TestFR_022_PasteMintsFreshIdentity` | (pending phase 6) |
| FR-023 | FR | T-0067 | `TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded` | (pending phase 6) |
| FR-024 | FR | T-0233 | `TestFR_024_RefuseMergeOnDuplicateIdentifier` | (pending phase 6) |
| FR-025 | FR | T-0075 | `TestFR_025_AddressingIsContentIdentityOnly` | (pending phase 6) |
| FR-026 | FR | T-0076 | `TestFR_026_BoundaryBehaviourIsClosedFourValueEnum` | (pending phase 6) |
| FR-027 | FR | T-0077, T-0086 | `TestFR_027_BoundaryBehaviourRoundTripsUnchanged`, `TestM04_OrphanCarriageSurvivesLedgerReload` | (pending phase 6) |
| FR-028 | FR | T-0078, T-0086 | `TestFR_028_FullDeletionOrphansRatherThanDrops`, `TestM04_OrphanCarriageSurvivesLedgerReload` | (pending phase 6) |
| FR-029 | FR | T-0079, T-0086 | `TestFR_029_OrphanResolutionIsDeterministic`, `TestM04_OrphanCarriageSurvivesLedgerReload` | (pending phase 6) |
| FR-030 | FR | T-0080, T-0086 | `TestFR_030_OrphanRecordCapturesAllFourFields`, `TestM04_OrphanCarriageSurvivesLedgerReload` | (pending phase 6) |
| FR-031 | FR | T-0081, T-0082 | `TestFR_031_TextSpanRequiresLanguageTagField`, `TestFR_031_PDLANG001RejectsUnresolvableTag` | (pending phase 6) |
| FR-032 | FR | T-0268, T-0269, T-0270 | `TestFR_032_DirectionFieldInDataModel`, `CONFORMANCE-BIDI-000-direction-roundtrip`, `CONFORMANCE-BIDI-001-missing-direction-rejected` | (pending phase 6) |
| FR-033 | FR | T-0271 | `CONFORMANCE-BIDI-002-inband-control-rejected` | (pending phase 6) |
| FR-034 | FR | T-0272 | `TestFR_034_ComputedInlineIsolation` | (pending phase 6) |
| FR-035 | FR | T-0089 | `TestFR_035_ExtractionReproducesAuthoredScalarSequenceExactly` | (pending phase 6) |
| FR-036 | FR | T-0273, T-0274, T-0275, T-0276 | `TestFR_036_RootSequenceSchema`, `CONFORMANCE-A11Y-005-rootsequence-completeness`, `TestFR_036_TCTraversalUsesRootSequence`, `TestFR_036_ExtractWalksRootSequenceOrder` | (pending phase 6) |
| FR-037 | FR | T-0277 | `CONFORMANCE-A11Y-001-mark-placement` | (pending phase 6) |
| FR-038 | FR | T-0279, T-0280 | `TestFR_038_RoleMapCompleteness`, `CONFORMANCE-A11Y-002-role-resolution` | (pending phase 6) |
| FR-039 | FR | T-0281, T-0282 | `CONFORMANCE-TBL-000-scope-roundtrip`, `CONFORMANCE-TBL-001-header-scope` | (pending phase 6) |
| FR-040 | FR | T-0284, T-0285 | `CONFORMANCE-A11Y-000-alttext-roundtrip`, `CONFORMANCE-A11Y-003-alttext-quality` | (pending phase 6) |
| FR-041 | FR | T-0087 | `TestFR_041_WalksContentSegmentsOnlyNoHeavyDeps` | (pending phase 6) |
| FR-042 | FR | T-0088 | `TestFR_042_EmitsUnitIdentityAndScalarPositionLocator` | (pending phase 6) |
| FR-043 | FR | T-0090 | `TestFR_043_EmitsExactlyOneLanguageTagPerUnit` | (pending phase 6) |
| FR-044 | FR | T-0091 | `TestFR_044_EmitsDescriptiveMetadataWithoutHeavyDecoders` | (pending phase 6) |
| FR-045 | FR | T-0092 | `TestFR_045_EmitsPageAdjunctWhenPaginationCurrent` | (pending phase 6) |
| FR-046 | FR | T-0093 | `TestFR_046_OmitsPageAdjunctAndReportsStaleWhenPaginationAbsentOrStale` | (pending phase 6) |
| FR-047 | FR | T-0094 | `TestFR_047_EmitsEarlierUnitTextBeforeReadingLaterUnitOctets` | (pending phase 6) |
| FR-048 | FR | T-0095 | `TestFR_048_AbandoningExtractionReadsNoFurtherOctets` | (pending phase 6) |
| FR-049 | FR | T-0096 | `TestFR_049_ExtractionSucceedsWithZeroTrustMaterial` | (pending phase 6) |
| FR-050 | FR | T-0157 | `TestFR_050_VerdictIsClosedFourValueEnum` | (pending phase 6) |
| FR-051 | FR | T-0012 | `TestFR_051_PreviewPayloadWithinBoundedPrefix` | (pending phase 6) |
| FR-052 | FR | T-0013 | `TestFR_052_PreviewDigestCoversRenderInputs` | (pending phase 6) |
| FR-053 | FR | T-0014 | `TestFR_053_PreviewDigestMismatchReportsStale` | (pending phase 6) |
| FR-054 | FR | T-0011 | `TestFR_054_MetadataFromBoundedPrefix` | (pending phase 6) |
| FR-055 | FR | T-0041, T-0042 | `TestFR_055_UnitIndexLeafConstructionRebuildsFromContent`, `TestFR_055_BoundedLookupWithinThreeReads` | (pending phase 6) |
| FR-056 | FR | T-0033, T-0046, T-0047 | `TestFR_056_PlaceNeverMutatesExistingOctets`, `TestFR_056_SealedSegmentRejectsMutation`, `FuzzFR_056_SegmentFrameDecode` | (pending phase 6) |
| FR-057 | FR | T-0043, T-0048 | `TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced`, `CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY` | (pending phase 6) |
| FR-058 | FR | T-0034 | `TestFR_058_StorageOrdinalMonotonicIndependentOfContent` | (pending phase 6) |
| FR-059 | FR | T-0208, T-0209, T-0212, T-0223, T-0224 | `TestHistorySegment_RLEBatchEncoding`, `TestHistorySegment_DecodeConformanceAtAndOverLimit`, `TestFR_059_CompleteHistoryReconstructsEveryState`, `TestConformance_HistorySegmentAndErasureCorpus`, `TestHistoryPackage_PublicAPIStableAndTraceable` | (pending phase 6) |
| FR-060 | FR | T-0208, T-0209, T-0214, T-0223, T-0224 | `TestHistorySegment_RLEBatchEncoding`, `TestHistorySegment_DecodeConformanceAtAndOverLimit`, `TestFR_060_RetainedFromPointReconstructsGuaranteedRange`, `TestConformance_HistorySegmentAndErasureCorpus`, `TestHistoryPackage_PublicAPIStableAndTraceable` | (pending phase 6) |
| FR-061 | FR | T-0217, T-0218, T-0223 | `TestFR_061_ErasureRecordSaltedCommitmentForm`, `TestFR_061_ErasureEnumeratedOnRetentionAdvance`, `TestConformance_HistorySegmentAndErasureCorpus` | (pending phase 6) |
| FR-062 | FR | T-0159 | `TestFR_062_UnavailableStateReportsDistinctVerdict` | (pending phase 6) |
| FR-063 | FR | T-0145, T-0146, T-0147, T-0148, T-0149, T-0150, T-0151, T-0152, T-0153, T-0158, T-0166, T-0363 | `TestFR_063_CoverageDescriptorRoundTrip`, `TestFR_063_PDCover001RejectsZeroLengthRange`, `TestFR_063_PDCover002RejectsUnmergedAdjacentRanges`, `TestFR_063_PDCover003RejectsGapOrOverlap`, `TestFR_063_PDCover004RejectsAttestOrdinal`, `FuzzFR_063_CoverageDescriptorCanonicalisation`, `TestFR_063_CoverageDescriptorDigestDeterministic`, `TestFR_063_SignedObjectPreimageAssembly`, `TestFR_063_SignatureRecordRoundTrip`, `TestFR_067_VerifyReportsSignedStateSignerAndPresentation`, `TestFR_063_ConformanceCorpusAtLimitAndOverLimitFixtures`, `TestM09_ArgusMilestoneExitReviewRecorded` | (pending phase 6) |
| FR-064 | FR | T-0153, T-0154, T-0155 | `TestFR_063_SignatureRecordRoundTrip`, `TestFR_064_PresentationArtefactRoundTrip`, `TestFR_064_SignedObjectSourcesPresentationDigestFromReferencedSlot` | (pending phase 6) |
| FR-065 | FR | T-0156 | `TestFR_065_MismatchedPresentationReportsUnattested` | (pending phase 6) |
| FR-066 | FR | T-0160 | `TestFR_066_SignaturePreservedAcrossSubsequentEdit` | (pending phase 6) |
| FR-067 | FR | T-0158, T-0161, T-0363 | `TestFR_067_VerifyReportsSignedStateSignerAndPresentation`, `TestFR_067_PerStateAttestationReportCoversAllOutcomes`, `TestM09_ArgusMilestoneExitReviewRecorded` | (pending phase 6) |
| FR-068 | FR | T-0162 | `TestFR_068_EnumeratesUnitsDifferingFromSignedState` | (pending phase 6) |
| FR-069 | FR | T-0163 | `TestFR_069_SignedPresentationExemptFromStaleness` | (pending phase 6) |
| FR-070 | FR | T-0167, T-0168, T-0169, T-0171, T-0172, T-0173, T-0174, T-0175, T-0177, T-0182, T-0183, T-0184 | `TestFR_070_AeKindFormatPairingRejectsInvalid`, `ATTEST-EVID-ROUNDTRIP`, `TestFR_070_SignatureLtvRefsRequireCorrectKind`, `TestFR_070_TrustAnchorsNoNetworkCalls`, `TestFR_070_CredentialChainVerifiesOffline`, `TestFR_070_RevocationEvidenceParsesOffline`, `TestFR_070_TimeAttestationParsesAndVerifiesOffline`, `TestFR_070_FullEvidenceChainVerdictStableAcrossTime`, `TestFR_070_SigIntentRejectsOutOfRange`, `FuzzAttestationEvidenceDecode`, `TestArgus_M10_LtvSecurityReviewChecklist`, `ATTEST-EVID-CEILING-BOUNDARY` | (pending phase 6) |
| FR-071 | FR | T-0170, T-0174, T-0176, T-0183 | `TestFR_071_NestedChainMandatoryForTimeAttestation`, `TestFR_070_TimeAttestationParsesAndVerifiesOffline`, `T-TSA-LTV`, `TestArgus_M10_LtvSecurityReviewChecklist` | (pending phase 6) |
| FR-072 | FR | T-0178, T-0179, T-0183 | `TestFR_072_SigningInstantOutsideIntervalUnverified`, `N-TIME`, `TestArgus_M10_LtvSecurityReviewChecklist` | (pending phase 6) |
| FR-073 | FR | T-0180, T-0181, T-0183 | `TestFR_073_RevocationAtOrBeforeSigningUnverified`, `N-COMPROMISE`, `TestArgus_M10_LtvSecurityReviewChecklist` | (pending phase 6) |
| FR-074 | FR | T-0185, T-0186, T-0187, T-0188, T-0205, T-0206 | `TestFR_074_DesignateRedactableSubtree`, `TestFR_074_SaltStoredInsideSubtree`, `TestFR_074_SaltedCommitmentDigest`, `redact-designate-conformance-001`, `TestSEC_M11_ArgusReviewPassed`, `redact-commitment-boundary-conformance-001` | (pending phase 6) |
| FR-075 | FR | T-0187, T-0189, T-0190, T-0205 | `TestFR_074_SaltedCommitmentDigest`, `TestFR_075_HidingBoundAgainstBruteForce`, `gov-checklist-FR061-FR075-ruling-recorded`, `TestSEC_M11_ArgusReviewPassed` | (pending phase 6) |
| FR-076 | FR | T-0191, T-0192, T-0193, T-0205 | `TestFR_076_RedactPreservesTCRoot`, `TestFR_076_VerifyReportsDeclaredOmissions`, `redact-declared-omission-conformance-001`, `TestSEC_M11_ArgusReviewPassed` | (pending phase 6) |
| FR-077 | FR | T-0194, T-0195, T-0205 | `TestFR_077_UndesignatedOmissionUnverified`, `redact-undesignated-omission-conformance-001`, `TestSEC_M11_ArgusReviewPassed` | (pending phase 6) |
| FR-078 | FR | T-0196, T-0197, T-0198, T-0205 | `TestFR_078_PublishOmitsRemovedUnitOctets`, `redact-orphan-residue-conformance-001`, `TestFR_078_ResidueScannerFindsNoMatches`, `TestSEC_M11_ArgusReviewPassed` | (pending phase 6) |
| FR-079 | FR | T-0199, T-0200 | `TestGOV_FR079_ClarificationRecorded`, `TestFR_079_ActorFieldInventoryEnumeratesKnownFields` | (pending phase 6) |
| FR-080 | FR | T-0201, T-0202 | `TestFR_080_PublishStripsActorIdentityValues`, `redact-actor-strip-conformance-001` | (pending phase 6) |
| FR-081 | FR | T-0203, T-0204 | `TestFR_081_PublishPreservesCustodyFixityValues`, `redact-custody-fixity-conformance-001` | (pending phase 6) |
| FR-082 | FR | T-0283 | `CONFORMANCE-TBL-002-cell-tiling` | (pending phase 6) |
| FR-083 | FR | T-0286, T-0287 | `TestFR_083_NumberingLabelDeterministic`, `CONFORMANCE-A11Y-004-numbering-literal-rejected` | (pending phase 6) |
| FR-084 | FR | T-0288 | `CONFORMANCE-XREF-000-presentation-function-roundtrip` | (pending phase 6) |
| FR-085 | FR | T-0289 | `CONFORMANCE-XREF-001-staleness-without-layout` | (pending phase 6) |
| FR-086 | FR | T-0243 | `TestFR_086_PageDirectoryDigestBinding` | (pending phase 6) |
| FR-087 | FR | T-0244 | `TestFR_088_StalePageDirectoryRefusedAndRederived` | (pending phase 6) |
| FR-088 | FR | T-0244 | `TestFR_088_StalePageDirectoryRefusedAndRederived` | (pending phase 6) |
| FR-089 | FR | T-0245, T-0264 | `font_record_seven_fields`, `render_layer_at_limit_conformance_corpus` | (pending phase 6) |
| FR-090 | FR | T-0257 | `TestFR_090_UnimplementableObjectUsesMaterialisedRepresentation` | (pending phase 6) |
| FR-091 | FR | T-0016 | `TestFR_091_SegmentTableSlotFieldsExposed` | (pending phase 6) |
| FR-092 | FR | T-0225, T-0226, T-0227, T-0229, T-0231, T-0232 | `TestFR_092_ClassifyDispatchHasNoWildcardCase`, `TestFR_092_DisjointCommuteOrderIndependent`, `TestFR_093_ContiguousBurstSurvivesAsSubstring`, `TestFR_092_R2TotalOrderTiebreakDeterministic`, `TestFR_092_DeleteDominatesConcurrentOps`, `TestFR_092_ExhaustiveOperationKindPairClassification` | (pending phase 6) |
| FR-093 | FR | T-0227, T-0228 | `TestFR_093_ContiguousBurstSurvivesAsSubstring`, `FuzzFR_093_BurstNonInterleaving` | (pending phase 6) |
| FR-094 | FR | T-0230 | `TestFR_094_MoveCycleRetainsSmallerStateId` | (pending phase 6) |
| FR-095 | FR | T-0234 | `TestFR_095_MissingCausalPredecessorBufferedOrRefused` | (pending phase 6) |
| FR-096 | FR | T-0235 | `TestFR_096_RefuseReplayOfErasedUnit` | (pending phase 6) |
| FR-097 | FR | T-0247 | `TestFR_097_FixedPaginationNoConversionStep` | (pending phase 6) |
| FR-098 | FR | T-0251 | `TestFR_098_ReflowNoScrollBelow320px` | (pending phase 6) |
| FR-099 | FR | T-0290, T-0291 | `CONFORMANCE-2D-000-region-roundtrip`, `CONFORMANCE-2D-001-region-fields` | (pending phase 6) |
| FR-100 | FR | T-0250 | `TestFR_100_ReflowDeterministicLineBreaks` | (pending phase 6) |
| FR-101 | FR | T-0276, T-0278 | `TestFR_036_ExtractWalksRootSequenceOrder`, `CONFORMANCE-A11Y-006-fixed-reflow-parity` | (pending phase 6) |
| FR-102 | FR | T-0114, T-0128 | `TestFR_102_StructuralFailureReportsOffsetUnitRule`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-103 | FR | T-0113, T-0128 | `TestFR_103_ErrorPrecedenceReportsEarliestFailureOnly`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-104 | FR | T-0019 | `TestFR_104_DigestMismatchAbortsBeforeDecode` | (pending phase 6) |
| FR-105 | FR | T-0020 | `TestFR_105_ExactlyOneStoredUnitInventoryType` | (pending phase 6) |
| FR-106 | FR | T-0115, T-0116, T-0128 | `TestFR_106_CeilingPreAllocationAbortsBeforeAllocation`, `TestFR_106_CON010_CeilingFixtureCorpus`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-107 | FR | T-0053, T-0059, T-0066 | `TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition`, `TestFR_107_DispositionRuntimeEnforcement`, `FuzzExtEnvelopeDecode` | (pending phase 6) |
| FR-108 | FR | T-0117, T-0128 | `TestFR_108_UnresolvedOrAmbiguousReferenceRejected`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-109 | FR | T-0118, T-0119, T-0128 | `TestFR_109_CycleDetectionNamesAllEdgesBeforeCeilingChecks`, `TestFR_109_ExtensionEnvelopeFallbackCycleVector`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-110 | FR | T-0120, T-0128 | `TestFR_110_DuplicateIdentifierNamesBothLocations`, `TestM07_PipelineStepOrderingGoldenCorpus` | (pending phase 6) |
| FR-111 | FR | T-0258 | `TestFR_111_MissingResourceRendersPlaceholder` | (pending phase 6) |
| FR-112 | FR | T-0259 | `TestFR_112_RenderReportRecordsSubstitution` | (pending phase 6) |
| FR-113 | FR | T-0260 | `TestFR_113_SubstitutionMarksPaginationNonAuthoritative` | (pending phase 6) |
| FR-114 | FR | T-0261 | `TestFR_114_SubstitutionNoHostResourceNoReflow` | (pending phase 6) |
| FR-115 | FR | T-0164, T-0363 | `TestFR_115_UnverifiedNeverCarriesSignerIdentity`, `TestM09_ArgusMilestoneExitReviewRecorded` | (pending phase 6) |
| FR-116 | FR | T-0238 | `TestFR_116_RecordsSupersededStateAndResolution` | (pending phase 6) |
| FR-117 | FR | T-0007, T-0008, T-0009, T-0010 | `TestFR_004_ParentStateIdRoundTrip`, `TestFR_117_PDRING001_EqualSequenceTie`, `TestFR_117_SelfDigestDetectsTornSlot`, `TestFR_117_InterruptedCommitSingleReadableState` | (pending phase 6) |
| FR-118 | FR | T-0292, T-0293 | `TestFR_118_InferredMarkerEnumeration`, `CONFORMANCE-INFER-001-marker-consistency` | (pending phase 6) |
| FR-119 | FR | T-0298, T-0299 | `TestFR_119_TransformDeterministicAcrossInvocations`, `TestFR_119_ExternalTrialCorpusPublished` | (pending phase 6) |
| FR-120 | FR | T-0300 | `TestFR_120_IdentifiersSurviveMigration` | (pending phase 6) |
| FR-121 | FR | T-0296, T-0297 | `TestFR_121_RefusalHaltsAtFirstUnrepresentableConstruct`, `TestFR_121_ConformanceVectors_UnrepresentableConstructs` | (pending phase 6) |
| FR-122 | FR | T-0302 | `TestFR_122_PreMigrationSignatureCoversOriginalState` | (pending phase 6) |
| FR-123 | FR | T-0301 | `TestFR_123_UnsupportedMajorVersionDeclinedFirst` | (pending phase 6) |
| FR-124 | FR | T-0307, T-0315, T-0316, T-0317 | `TestNFR_002_CanonicalTraversalOrderMatchesTC`, `TestFR_124_EmptyStateEmitsWellDefinedMinimalPrefix`, `CONF-DEGENERATE-001-empty-content-case-corpus`, `TestFR_124_EmptyStateOutcomeTableSchemaComplete` | (pending phase 6) |
| FR-125 | FR | T-0006 | `TestFR_125_MagicConstantIdentification` | (pending phase 6) |
| NFR-001 | NFR | T-0001, T-0002, T-0030, T-0032 | `TestNFR_001_VarintMinimalEncodingRoundTrip`, `TestNFR_001_TLVClosedGrammarRoundTrip`, `TestNFR_001_FixedPrefixCanonicalOctetsDeterministic`, `TestM01_FixedPrefixExitConformanceSuite` | (pending phase 6) |
| NFR-002 | NFR | T-0307, T-0308, T-0309, T-0318, T-0324 | `TestNFR_002_CanonicalTraversalOrderMatchesTC`, `TestNFR_002_CanonicalizeStreamsWithoutMaterializing`, `CONF-CANON-001-storage-order-independence`, `TestNFR_002_PublishAndMigrateUseStreamingCanonicalize`, `TestNFR_002_CanonicalizerHas100PercentCoverage` | (pending phase 6) |
| NFR-003 | NFR | T-0035 | `TestNFR_003_NoOpSaveByteIdentical` | (pending phase 6) |
| NFR-004 | NFR | T-0310, T-0311, T-0312, T-0313, T-0314 | `TestNFR_004_FullCompactionRefusedWhenSignaturePresent`, `CONF-COMPACT-001-full-compaction-refusal-corpus`, `TestNFR_004_PartialCompactionLeavesCoveredSegmentsByteIdentical`, `CONF-COMPACT-002-total-mode-zero-reclaim`, `TestNFR_004_CompactionExplicitOnlyNotInvokedBySave` | (pending phase 6) |
| NFR-005 | NFR | T-0165, T-0363 | `TestNFR_005_NoAmbientValuesOutsideAllowlistedSites`, `TestM09_ArgusMilestoneExitReviewRecorded` | (pending phase 6) |
| NFR-006 | NFR | T-0103, T-0110, T-0112 | `TestNFR_006_SignDeterministic`, `TestNFR_006_SignTwiceIdenticalOctets`, `TestCON_015_ArgusSecurityReviewChecklist` | (pending phase 6) |
| NFR-007 | NFR | T-0138, T-0362 | `TestNFR_007_AttestSegmentsExcludedFromStateDigestAndNoOpCompare`, `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` | (pending phase 6) |
| NFR-008 | NFR | T-0036 | `BenchmarkNFR_008_EditWriteCostWithinBudget` | (pending phase 6) |
| NFR-009 | NFR | T-0037, T-0038 | `TestNFR_009_ChunkBoundaryDeterministicAcrossEditHistories`, `BenchmarkNFR_009_NovelChunkTotalAcrossDocSizes` | (pending phase 6) |
| NFR-010 | NFR | T-0242, T-0264 | `TestNFR_010_PageDirectoryKeyedByContentIdentity`, `render_layer_at_limit_conformance_corpus` | (pending phase 6) |
| NFR-011 | NFR | T-0342, T-0343 | `TestNFR_011_ReferenceConfigPublished`, `TestNFR_011_BenchmarksCiteReferenceConfig` | (pending phase 6) |
| NFR-012 | NFR | T-0098 | `BenchmarkNFR_012_ExtractionReadsWithin15PercentOfFileOctets` | (pending phase 6) |
| NFR-013 | NFR | T-0099 | `BenchmarkNFR_013_ExtractionCompletesWithin20SecondsProcessorTime` | (pending phase 6) |
| NFR-014 | NFR | T-0100 | `BenchmarkNFR_014_ExtractionPeakMemoryStaysFlatAndBounded` | (pending phase 6) |
| NFR-015 | NFR | T-0015 | `BenchmarkNFR_015_ColdCachePreviewLatency` | (pending phase 6) |
| NFR-016 | NFR | T-0015 | `BenchmarkNFR_015_ColdCachePreviewLatency` | (pending phase 6) |
| NFR-017 | NFR | T-0254, T-0255, T-0256, T-0262, T-0265 | `plp1_decode_conformance_corpus_frozen`, `TestPLP1_DecoderPassesConformanceCorpus`, `TestRestrictedPNG_DecodeConformance`, `BenchmarkNFR_017_OpenAndRenderPagePeakMemory`, `FuzzPLP1Decode` | (pending phase 6) |
| NFR-018 | NFR | T-0263 | `BenchmarkNFR_018_RenderPageOctetReadBudget` | (pending phase 6) |
| NFR-019 | NFR | T-0248 | `TestNFR_019_RasterizerExactRationalDeterminism` | (pending phase 6) |
| NFR-020 | NFR | T-0003, T-0031 | `TestFR_007_HeaderContentIndependent`, `TestNFR_020_UnicodeVersionBoundToFormatMajor` | (pending phase 6) |
| NFR-021 | NFR | T-0252 | `TestNFR_021_ShapingDeterministicGivenPinnedOracle` | (pending phase 6) |
| NFR-022 | NFR | T-0250 | `TestFR_100_ReflowDeterministicLineBreaks` | (pending phase 6) |
| NFR-023 | NFR | T-0249 | `TestNFR_023_NoGridFittingNoInstructionExecution` | (pending phase 6) |
| NFR-024 | NFR | T-0246 | `TestNFR_024_DurableProfileRendersIdenticallyOffline` | (pending phase 6) |
| NFR-025 | NFR | T-0344, T-0358 | `TestNFR_025_RoleStatementCeilingsEnforced`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| NFR-026 | NFR | T-0345, T-0358, T-0365 | `TestNFR_026_ExternalReaderTrial_5DayBudget`, `TestM19_V1StableDeclarationGateRecorded`, `TestNFR_026_TrialMissRulingRecorded` | (pending phase 6) |
| NFR-027 | NFR | T-0346, T-0347, T-0358 | `TestNFR_027_ExternalRenderingTrial_30DayBudget`, `TestNFR_027_ScheduleRiskRulingRecorded`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| NFR-028 | NFR | T-0348, T-0349, T-0350, T-0351, T-0358 | `TestNFR_028_SecondImplScopeDefined`, `TestNFR_028_CQ012FundingDecisionRecorded`, `TestNFR_028_DifferentialHarnessDiffsCanonicalOctets`, `TestNFR_028_SecondImplementationMatchesCanonicalOctetsAndVerdicts`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| NFR-029 | NFR | T-0352, T-0353, T-0357, T-0358 | `TestNFR_029_TraceabilityCheckerParsesAllRuleIDs`, `TestNFR_029_AllNormativeStatementsHaveConformanceCase`, `TestNFR_029_FuzzHarnessMaturityMeetsCP012Bar`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| NFR-030 | NFR | T-0121, T-0122, T-0361 | `BenchmarkNFR_030_ValidatorPeakMemoryWithinAdoptedBound`, `FuzzValidate_NFR030_MemoryBoundAndNoCrash`, `TestNFR_030_ClarifyRulingRequestRecorded` | (pending phase 6) |
| NFR-031 | NFR | T-0294, T-0295 | `TestNFR_031_RuleRegistryComplete`, `TestNFR_031_EightyPercentDecidable` | (pending phase 6) |
| NFR-032 | NFR | T-0221, T-0222 | `TestNFR_032_CompleteHistoryOverheadBudget250kOps`, `TestNFR_032_AdversarialScatteredTraceDocumentedMiss` | (pending phase 6) |
| NFR-033 | NFR | T-0219, T-0220 | `TestNFR_033_OpenCostIndependentOfOpCount`, `TestNFR_033_TenXOpCountOpensWithin1_5x` | (pending phase 6) |
| NFR-034 | NFR | T-0123 | `TestNFR_034_NoDecoderConstructedDuringStructuralValidation` | (pending phase 6) |
| CON-001 | CON | T-0068, T-0083 | `TestCON_001_BaseOrdinalIsScalarValueScopedToSegment`, `TestCON_001_AFieldRoleAuditHasNoPersistedCountedPosition` | (pending phase 6) |
| CON-002 | CON | T-0073 | `TestCON_002_NFCScopingIsPerRunNotCrossBoundary` | (pending phase 6) |
| CON-003 | CON | T-0074 | `TestCON_003_WriterRejectsNonNFCRatherThanConverting` | (pending phase 6) |
| CON-004 | CON | T-0124 | `TestCON_004_ExcludedScalarClassesRejected` | (pending phase 6) |
| CON-005 | CON | T-0026 | `TestCON_005_ExactlyOneRepresentationPerCapability` | (pending phase 6) |
| CON-006 | CON | T-0253, T-0266 | `TestCON_006_ShapingExceptionRecorded`, `TestEX001_ShapingIsolationDoesNotBlockCoreRenderPath` | (pending phase 6) |
| CON-007 | CON | T-0024 | `TestCON_007_NoFieldResolvesLocationOrExecutes` | (pending phase 6) |
| CON-008 | CON | T-0025 | `TestCON_008_UnitIdOpaqueExactEqualityOnly` | (pending phase 6) |
| CON-009 | CON | T-0017 | `TestCON_009_CeilingTableMatchesSpecText` | (pending phase 6) |
| CON-010 | CON | T-0008, T-0010, T-0018, T-0032 | `TestFR_117_PDRING001_EqualSequenceTie`, `TestFR_117_InterruptedCommitSingleReadableState`, `TestCON_010_FrameCountBoundaryAndOverByOne`, `TestM01_FixedPrefixExitConformanceSuite` | (pending phase 6) |
| CON-011 | CON | T-0125 | `TestCON_011_OverBudgetExitCodePrecedence` | (pending phase 6) |
| CON-012 | CON | T-0027 | `TestCON_012_GeometricValueIsFixedPointInteger` | (pending phase 6) |
| CON-013 | CON | T-0028 | `TestCON_013_ProportionalSizeRoundingRemainderToFinalShare` | (pending phase 6) |
| CON-014 | CON | T-0029 | `TestCON_014_SingleColourRepresentationDefined` | (pending phase 6) |
| CON-015 | CON | T-0102, T-0104, T-0105, T-0106, T-0107, T-0108, T-0109, T-0111, T-0112 | `TestCON_015_ParamSetAllowlist`, `TestCON_015_SmallOrderTableDigestPinned`, `TestCON_015_Step1Step2_PublicKeyChecks`, `TestCON_015_Step3Step4_RChecks`, `TestCON_015_Step5_ScalarRangeCheck`, `TestCON_015_Verify7StepShortCircuit`, `EDDSA_P1_CONFORMANCE_CORPUS_V1`, `TestCON_015_OffAllowlistParamSetRejected`, `TestCON_015_ArgusSecurityReviewChecklist` | (pending phase 6) |
| CON-016 | CON | T-0303, T-0304, T-0305, T-0306 | `TestCON_016_ReProtectionPreservesOriginalVerdict`, `TestCON_016_RescindAndResignAtMajorVersionBoundary`, `TestCON_016_RescindResignRecordConformanceVectors`, `TestCON_016_ArgusReviewNoP0P1Findings` | (pending phase 6) |
| CON-017 | CON | T-0126, T-0360 | `TestCON_017_SingleWriterConformanceClassAsserted`, `TestCON_017_AnalyzeInputNotePresent` | (pending phase 6) |
| CON-018 | CON | T-0127 | `TestCON_018_EveryReaderStatementAssignedToOneOfThreeRoles` | (pending phase 6) |
| CON-019 | CON | T-0349, T-0354, T-0358 | `TestNFR_028_CQ012FundingDecisionRecorded`, `TestCON_019_FeatureNormativeStatusGatedOnTwoImplPass`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| CON-020 | CON | T-0050, T-0051, T-0060, T-0065 | `TestCON_020_ExtTokenOwnerIdPartition`, `TestCON_020_FmRetiredTokensSizeCap`, `TestCON_020_RetiredTokenNeverReissued`, `TestCON_020_OwnerIdPartitionBoundaries` | (pending phase 6) |
| CON-021 | CON | T-0061 | `TestCON_021_RegistryTurnaroundSLAPublished` | (pending phase 6) |
| CON-022 | CON | T-0207, T-0213, T-0215, T-0224 | `TestCON_022_HistoryModeClosedEnumAndImmutable`, `TestCON_022_RetentionPointModeGated`, `TestCON_022_NoHistoryModeRetainsOnlyCurrentState`, `TestHistoryPackage_PublicAPIStableAndTraceable` | (pending phase 6) |
| CON-023 | CON | T-0216 | `TestCON_023_RefusesInPlaceRemovalOnCompleteHistory` | (pending phase 6) |
| CON-024 | CON | T-0236, T-0364 | `TestCON_024_RefuseMergeAcrossRetentionPoint`, `TestCON_024_CON_025_GuardsAtDeclaredWireLimit` | (pending phase 6) |
| CON-025 | CON | T-0237, T-0364 | `TestCON_025_RefuseMergeOnHistoryModeMismatch`, `TestCON_024_CON_025_GuardsAtDeclaredWireLimit` | (pending phase 6) |
| CON-026 | CON | T-0355, T-0356, T-0358 | `TestCON_026_LicenceDraftPublished`, `TestCON_026_GovernanceGateApprovedAndOnRecord`, `TestM19_V1StableDeclarationGateRecorded` | (pending phase 6) |
| TR-001 | TR | T-0328 | `TestTR_001_FindingCarriesAllFiveFields` | (pending phase 6) |
| TR-002 | TR | T-0239, T-0241, T-0332 | `TestTR_002_DiffReportsConstructsNotStorageUnits`, `TestM13_MergeOrchestratorEndToEnd`, `TestTR_002_DiffVerbConstructLevel` | (pending phase 6) |
| TR-003 | TR | T-0240, T-0241, T-0333 | `TestTR_003_ConflictNamesBothValuesNoAutoResolve`, `TestM13_MergeOrchestratorEndToEnd`, `TestTR_003_MergeVerbConflictExit` | (pending phase 6) |
| TR-004 | TR | T-0319, T-0320, T-0321, T-0322, T-0334 | `TestTR_004_TextProjectionIsDeterministic`, `TestTR_004_HtmlProjectionIsDeterministic`, `TestTR_004_ProjectionRoundTripsToIdenticalCanonicalOctets`, `CONF-PROJECT-001-full-corpus-round-trip`, `TestTR_004_ProjectRoundTripsExactly` | (pending phase 6) |
| TR-005 | TR | T-0323, T-0334 | `TestTR_005_ProjectionRejectedAsDocumentInput`, `TestTR_004_ProjectRoundTripsExactly` | (pending phase 6) |
| TR-006 | TR | T-0021 | `TestTR_006_EnumerateStoredUnitsFromBoundedPrefix` | (pending phase 6) |
| TR-007 | TR | T-0022 | `TestTR_007_CoverageHintFromBoundedPrefix` | (pending phase 6) |
| TR-008 | TR | T-0023 | `TestTR_008_SegmentTypeEnumClosedNoExecutingKind` | (pending phase 6) |
| TR-009 | TR | T-0129, T-0130, T-0131, T-0140, T-0141, T-0144, T-0362 | `TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed`, `TestTR_009_TSTreeFixedShapeAndAbsentChildFill`, `TestTR_009_TSRootAlwaysRecomputedNeverCached`, `TestFR_002_StructureDigestPreimageMatchesBitmaskExactly`, `TestTR_009_TamperedSegmentTableSlotDetectedBeforeReliance`, `TestTR_009_TSTreeAtLimitAndOnePastLimit`, `TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff` | (pending phase 6) |
| TR-010 | TR | T-0044, T-0045, T-0049, T-0341 | `TestTR_010_ConditionalWriterInterfaceContract`, `TestTR_010_ConditionalWriteRefusalNamesCurrentHolder`, `TestTR_010_DesignNoteContainsRequiredSections`, `TestTR_010_ConditionalWriteRefusesOnMismatch` | (pending phase 6) |
| TR-011 | TR | T-0101 | `TestTR_011_ExtractPackageUnder1000LinesNoHeavyDeps` | (pending phase 6) |
| TR-012 | TR | T-0325, T-0326, T-0327, T-0329, T-0330, T-0331, T-0332, T-0333, T-0334, T-0335, T-0336, T-0337, T-0338, T-0339, T-0340 | `TestTR_012_DispatchRegistersAllElevenVerbs`, `TestTR_012_ExitCodePrecedenceTable`, `TestTR_012_ValidateVerbInvocationShape`, `TestTR_012_InspectVerbBoundedRead`, `TestTR_012_ExtractVerbStreaming`, `TestTR_012_VerifyVerbVerdictMapping`, `TestTR_002_DiffVerbConstructLevel`, `TestTR_003_MergeVerbConflictExit`, `TestTR_004_ProjectRoundTripsExactly`, `TestTR_012_RedactVerbOmissionEnumeration`, `TestTR_012_PublishVerbZeroResidue`, `TestTR_012_SignVerbDeterministicOutput`, `TestTR_012_MigrateVerbRefusalFirst`, `TestTR_012_TraceabilityTableMatchesDispatch`, `TestTR_012_CLIConformanceSuite` | (pending phase 6) |

---

## 5. Orphan check

Requirements with zero covering tasks: **0**.

Gate passes: every one of the 197 requirements has at least one covering task.

---

## 6. Cross-check findings

16 raw findings across 3 lenses. 12 blocker/major findings were adversarially re-reviewed: 11 survived, 1 were refuted as stale-reference artifacts, already-resolved by GAP A/B, or hallucinated ids. Refuted findings are not listed below; they are the difference between the raw and surviving counts and exist only in this run's transcript.

| Severity | Lens | Category | Description | Artifacts | Resolution |
|---|---|---|---|---|---|
| blocker | constitution-violations | constitution-violation | CP-009 (no normative statement defined by another product's behaviour) has no exception clause in its own text or in the amendment process (only principles that explicitly permit one, like CP-004, can take a one-time exception per Amendment rule 6). T-0252 implements glyph shaping as a pure function that must reproduce a 'pinned, versioned external oracle's' exact output, and T-0266 institutionalizes the pipeline around it. plan.md itself names this 'Conflict 2: CON-006/CP-009 versus CQ-006 option B' and only proposes an 'EX-001 exception' -- a mechanism CP-009 does not offer. As written, implementing T-0252 ships a normative construct defined by reference to a named external implementation's behaviour, which is exactly what CP-009's Consequence text says blocks release with zero exceptions. | tasks.md | Do not resolve via a 'one-time exception' record; this needs Eyvar's explicit constitutional amendment (minor version bump relaxing CP-009) before T-0252 can be built, or CQ-006 must be reopened to select a from-scratch normative shaping spec instead. |
| blocker | constitution-violations | constitution-violation | CP-014 requires media-type and format-identification registration filed before the first stable release, as a v1 deliverable with its own gate. T-0006 only builds the in-repo magic-constant mechanism for FR-125; tasks.md's own residual-gaps section (line 4251) confirms no task performs or tracks the actual external registry submission. T-0358's capstone go/no-go gate list (NFR-025/026/027/028/029, CON-019, CON-026) does not include this registration at all, so v1-stable could be declared with FR-125's external registration never filed. | tasks.md | Add a real M19 task (owner clio) to file and record the IANA media-type / format-identification registration, and add it to T-0358's dependency list so the v1-stable capstone gate cannot pass without it. |
| blocker | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | orphan-wire-structure | document.abnf S4/S5 define TABLE (tbl-rows/tbl-columns/tbl-cells/cell-entry), NOTE, and CROSS_REFERENCE (xref-target/xref-kind/xref-anchor) as full record types, and data-model.md has no entity entry for any of the three (they are 'PLAN-LEVEL DESIGN... flagged' additions made only inside the ABNF file). No task in tasks.md builds any of these three base records. The only tasks that touch them (T-0281/T-0282/T-0283 add a `scope` field and tiling validator onto 'TABLE record'; T-0288/T-0289 add fields onto 'CROSS_REFERENCE') assume a base record that no task ever constructs, and they all sit in M15 gated behind T-0267 (a governance sign-off task, not a builder task). | contracts/document.abnf, data-model.md, tasks.md | Add data-model.md entities for Table/Note/CrossReference and add explicit M04-or-earlier tasks that build tbl-discriminant/tbl-rows/tbl-columns/tbl-cells, note-anchor/note-body-block/note-placement, and xref-discriminant/xref-target/xref-kind before any M15 field-addition task can have a real dependency target. |
| blocker | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | orphan-requirement | REGISTRY_EXCERPT (document.abnf S7.3; data-model.md §2.19) is mandatory whenever durable_claim=1 (FR-011) and its completeness is validation step 12, which cli.md's `validate` stdout schema names explicitly as the required check key `registry_excerpt_completeness`. No task builds the RegistryExcerpt record (re-kind/re-value/re-spec-pointer/re-snapshot-digest) or implements the step-12 completeness check. T-0005 explicitly states 'not RegistryExcerpt content validation (owned by a later milestone)' but no milestone or task anywhere owns it; T-0246 only wires RegistryExcerpt into font rendering for NFR-024, never the structural check. | contracts/document.abnf, contracts/cli.md, data-model.md, tasks.md | Add a task (likely M07, alongside T-0117/T-0127) that builds REGISTRY_EXCERPT's wire struct and implements validation step 12 / cli.md's `registry_excerpt_completeness` check, with its own conformance corpus. |
| major | constitution-violations | constitution-violation | CP-004 permits exactly four non-deterministic sites: identifier minting, redaction salts, signature values, and timestamp attestations (TSA-issued, per M10). T-0304's RESCIND-AND-RESIGN description mandates recording '(old signature id, reason, timestamp)' in the persisted RescindResignRecord -- a raw wall-clock value with no TSA attestation behind it, which is not one of the four allowed sites and is exactly the kind of ambient value NFR-005 bars from the octet stream. | tasks.md | Drop the raw timestamp field from RescindResignRecord, or replace it with a proper TSA time attestation reusing M10's ATTESTATION_EVIDENCE mechanism (one of the four allowed sites) rather than an unattested wall-clock value. |
| major | constitution-violations | constitution-violation | CP-001 requires code to follow, never precede, a spec edit ('Change flows downhill... a patch that changes observable behaviour without a preceding spec edit is reverted'). T-0187/T-0217/T-0218 implement FR-061's ErasureRecord using a 'provisional salted-commitment form' that contradicts FR-061's own frozen literal text (bare unsalted digest), justified only by plan.md's self-disclosed, still-open Conflict 1. No task gates M11/M12 completion on the ruling actually landing before this diverging wire format is built and shipped in the conformance corpus. | tasks.md | Sequence T-0190's clarify.md ruling-request ahead of T-0187/T-0217 build work (or gate their conformance-corpus freeze on it), and amend FR-061's text once the salted form is approved rather than shipping the diverging behaviour under the frozen requirement id. |
| major | constitution-violations | constitution-violation | CP-011 non-negotiable #5: rule identifiers are never renumbered or reused. spec.md line 1232 names the validator rule 'PD-NORM-001', but the frozen contracts/document.abnf and data-model.md instead use 'PD-NFC-001'/'PD-NFC-002' for the same rule -- a rename that has already happened across approved artifacts, not a hypothetical task risk. T-0353 self-discloses this exact violation but only files a gap report; the renamed ids remain live and are what T-0073/T-0074/T-0082 build against. | spec.md, contracts/document.abnf, data-model.md, tasks.md | Get Eyvar/themis to rule on which id is canonical (PD-NORM-001 or PD-NFC-001/002) and correct every other frozen artifact to match before any conformance corpus cites either id, rather than leaving two live spellings of one rule. |
| major | constitution-violations | constitution-violation | CP-012 requires 'a named triage owner and a published maximum of 90 days from report to fix or advisory... before enrolment in any public fuzzing service.' Every fuzz-related task (T-0359, T-0047, T-0066, T-0085, T-0122, T-0150, T-0182, T-0198, T-0228, T-0265, T-0300, T-0357) builds fuzzing infrastructure and aggregates crash-free/coverage metrics, but none establishes the named owner or the 90-day disclosure-clock policy itself -- a governance precondition CP-012 states must exist before public enrolment, not just a harness-maturity number. | tasks.md | Add an M19 governance task (owner clio, mirroring T-0355/T-0356's pattern) naming the fuzzing triage owner and publishing the 90-day disclosure SLA, gating any public fuzzing-service enrolment and feeding T-0358's capstone decision. |
| major | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | orphan-wire-structure | data-model.md §2.22 PageDirectory (backing FR-045, FR-046, FR-097, NFR-010's 'pagination artefact') has no assigned record discriminant or field shape anywhere in container.abnf or document.abnf's registries -- unlike UnitIndexLeaf, which got discriminant 0x0A specifically so it could be persisted and located. Tasks T-0242/T-0243/T-0244/T-0247 (M14) build 'PageDirectory entity' and bind it to a digest and staleness rule as if it were an ordinary persisted RESOURCE record, but no such record type exists to build against. | contracts/document.abnf, contracts/container.abnf, data-model.md, tasks.md | Assign PageDirectory a discriminant in document.abnf's RESOURCE registry (e.g. next after 0x0C) with an explicit field list, or state in data-model.md that it is purely in-memory/never persisted (in which case FR-045/046's 'present'/'stale' language needs rewording) before M14 tasks proceed. |
| major | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | data-model-contract-mismatch | data-model.md §2.2 CommitRingRecord types `segment_count` and `retention_point` as `uint32`, but container.abnf S3 explicitly types both `segment-count` and `retention-point` as `u16` (2 octets each), and container.abnf's own arithmetic proof for ring-fields (14 fields summing to exactly 284 octets, hence the 512-octet slot with 196-octet reserved) only balances at 2 octets each -- at uint32 the total would be 288, not 284, breaking the slot layout. Tasks T-0007 (ring struct), T-0042 (index-route wiring) and the byte-exact round-trip tests (T-0008, T-0009, T-0030, T-0032) all depend on the correct width. | contracts/container.abnf, data-model.md | Correct data-model.md's CommitRingRecord table: segment_count and retention_point are uint16, not uint32; also add the missing index_route field (16 x u16) that container.abnf mandates and T-0042 already implements but data-model.md's table omits entirely. |
| major | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | milestone-dependency-gap | cli.md's `validate` stdout schema requires the check key `storage_integrity_tree` (T_S recomputed and compared against CommitRingRecord.ledger_root -- data-model.md §7 validation step 6). T_S is built entirely in M08 (T-0130/T-0131), but M07 (Structural Validation Core, whose T-0113 is the validate pipeline orchestrator) depends only on M01, M02, M04 -- not M08 -- and no M07 task depends on any M08 task. T-0144's own description states 'structural ceiling rejection semantics... belong to M07's validator,' confirming M07 is expected to consume T_S output it has no dependency path to. | contracts/cli.md, data-model.md, tasks.md | Add M08 (or at least T-0131) to M07's milestone-table dependency list, and add a task wiring T_S recomputation into the validate pipeline's reported `storage_integrity_tree` check, the same way GAP A wired M07's cycle detector into M03. |
| minor | constitution-violations | constitution-violation | CP-002 requires each reader role to carry its own independent-implementer trial. CON-018 defines three distinct reader roles (extracting / validating-and-verifying / rendering), and T-0344's own ceiling table gives each a separate normative-statement budget (150/400/500). But only two trial tasks exist: T-0345 tests a merged 'extracting-and-validating' reader and T-0346 tests 'rendering' -- no task independently evidences a standalone trial for the validating-and-verifying role as CON-018 names it, leaving one of the three roles' 'own... trial' obligation unconfirmed. | tasks.md, spec.md | Confirm with Eyvar/themis whether NFR-026's combined 'extracting-and-validating' trial is intended to jointly satisfy both CON-018 roles' trial obligation, and record that reading explicitly in clarify.md/plan.md rather than leaving it implicit. |
| minor | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | data-model-arithmetic-error | data-model.md §2.4 SegmentTableSlot 'Limits' states the region occupies '[262144,1048960)', but 262144 + 786432 (the entity's own stated total size) = 1048576, not 1048960. This also contradicts container.abnf's own reconciled prefix total (512+3584+258048+786432 = 1048576 exactly, stated explicitly in container.abnf's CONTRACTS-PHASE RESOLUTION note). | data-model.md, contracts/container.abnf | Fix the typo: SegmentTable occupies [262144,1048576). |
| minor | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | data-model-completeness | data-model.md §5 states 'This is the complete v1 ceiling table' (per CP-007), but omits NFR-032's 2.0x history-overhead ceiling and NFR-033's 10x-op-count/1.5x-open-time ceiling even though both are implemented as named benchmarks in tasks.md (T-0221/T-0222 for NFR-032, T-0219/T-0220 for NFR-033). | data-model.md, tasks.md | Add rows for NFR-032 (2.0x) and NFR-033 (1.5x at 10x op-count) to data-model.md §5's ceiling table. |
| minor | contract/data-model/task consistency (container.abnf, document.abnf, integrity.abnf, data-model.md vs tasks.md) + GAP A satisfiability verification | gap-a-verification | CONFIRMED, not a defect: M07's T-0118 ('Reference-graph cycle detection over the 5 named edge kinds', implements FR-109, depends on T-0113/T-0115) is exactly the generic cycle detector M03's T-0056/T-0064 need, and extension-envelope fallback references are explicitly one of the 5 named edge kinds per document.abnf S5 / integrity.abnf S8. M07's own dependency set (M01, M02, M04) has no path back into M03, so GAP A's fix (M03 depends on M01, M02, M07) introduces no cycle and is mechanically satisfiable as stated. | tasks.md | No action needed; GAP A's milestone-table correction is sound. |

---

## 7. Constitution compliance

Checked against all 14 principles via the constitution-violations lens. Violations found, if any, are listed in section 6 above tagged with the violated principle id in their category field. No violation found means the current task set, as written, does not describe anything that would break a principle if built as specified — this does not certify the eventual CODE will comply; that is phase 7 (verify)'s job.


---

## 8. Gate verdict

**DOES NOT PASS (as originally run).** 0 orphan requirement(s) and 4 surviving blocker finding(s) had to be closed before phase 6 could begin. See sections 5 and 6 for the full original finding set. Section 9 below records what was fixed immediately afterward and what still blocks the gate.

## 9. Post-analysis fixes applied

Applied directly to `data-model.md` and `tasks.md` after this analysis ran, without waiting for a second full analyze pass, since each fix below is either purely mechanical or was independently re-verified (field-width arithmetic, dependency-cycle check, duplicate-id check) before being applied.

**Blockers closed (3 of 4):**
- CP-014 registration: added T-0371 (files the IANA media-type/format-identification registration), gated into T-0358's capstone decision.
- Orphaned wire structures TABLE/NOTE/CROSS_REFERENCE: added `data-model.md` entities 2.24-2.26 and building tasks T-0367/T-0368/T-0369.
- Orphaned REGISTRY_EXCERPT: added T-0370 (builds the wire struct, implements validation step 12 / cli.md's `registry_excerpt_completeness` check).

**Blocker still open (1 of 4):**
- CP-009 vs the pinned shaping-oracle construct (T-0252): no exception mechanism exists in CP-009's text or the amendment process. Requires a constitutional amendment or reopening CQ-006. Not something this pass can resolve unilaterally.

**Major findings closed (5 of 7):**
- CommitRingRecord field widths: `data-model.md` corrected `segment_count`/`retention_point` from `uint32` to `uint16`, and added the missing `index_route` field, matching container.abnf's ring-field arithmetic (verified: 4+8+8+32+2+32+32+32+32+32+2+4+32+32 = 284 exactly). Also clarified that `integrity_block_digest` already was container.abnf's `t-c-root` under a different name (a naming ambiguity, not a missing field) — this was caught and corrected before being shipped as a duplicate field.
- M07-to-M08 dependency gap: M07's milestone-table dependency list now includes M08; T-0366 wires T-0131's T_S recomputation into the validate pipeline's `storage_integrity_tree` check. Verified no dependency cycle results (M08 does not depend on M07).
- CP-004 RescindResignRecord: T-0304's description no longer claims a free-text "reason" or raw wall-clock "timestamp" field; data-model.md 2.18's actual entity never had either, so the fix was correcting the task's prose to match the already-approved wire shape rather than adding new fields.
- CP-012 fuzzing governance: added T-0372 (names the triage owner, publishes the 90-day disclosure SLA), gated into T-0358.
- CP-001 sequencing risk: T-0217 now depends on T-0190, so the provisional salted-commitment wire format's conformance freeze cannot complete ahead of clarify.md recording the FR-061/FR-075 ruling request.

**Major findings still open (2 of 7):**
- `PD-NORM-001` vs `PD-NFC-001`/`PD-NFC-002`: the same validator rule has two different ids across `spec.md` vs `contracts/document.abnf`/`data-model.md`. Needs a ruling on which is canonical before any conformance corpus cites either.
- PageDirectory's missing wire discriminant (vs. sibling UnitIndexLeaf's `0x0A`): investigation during the fix pass found this is plausibly a legitimate asymmetry — PageDirectory is genuinely lazy-rebuilt and bounded by `MAX_PAGES`, unlike UnitIndex's role in NFR-012's extraction-budget — rather than an oversight. The existing M14 tasks (T-0242-244, T-0247) already correctly implement it as a derived, non-normative, digest-bound-for-staleness structure matching data-model.md 2.22's own existing (and correct) classification. Not mechanically fixed because forcing a discriminant assignment could itself be wrong; needs Eyvar's ruling on whether the asymmetry is intentional.

**Minor findings: 1 of 2 closed, plus 1 confirmation needing no action:**
- SegmentTableSlot arithmetic typo (`1048960` should be `1048576`): fixed.
- NFR-032/NFR-033 ceiling-table rows: added.
- CON-018 trial-coverage question (whether the combined NFR-026 extracting-and-validating trial satisfies both of CON-018's named roles' "own trial" obligation): still open, needs Eyvar/themis confirmation, recorded in `tasks.md` section 5.
- The gap-a-verification finding required no action (it confirmed GAP A's fix was sound).

**Remaining before this gate fully passes (as originally listed):** CP-009 (blocker), the `PD-NORM-001`/`PD-NFC-001` naming conflict, FR-061's own text-vs-provisional-form approval, PageDirectory's discriminant question, and CON-018's trial-coverage confirmation — five items, all requiring Eyvar's explicit ruling rather than further mechanical work.

## 10. Rulings applied (2026-09-15)

Eyvar ruled on all five remaining items 2026-09-15, adopting the recommended option in every case. Full record in `clarify.md` CQ-013..CQ-017.

- **CQ-013 (CP-009 blocker):** amended CP-009 (`.specify/memory/constitution.md` v0.2.0) with a narrow standing exception for a pinned, versioned external artefact cited strictly as a byte-exact determinism oracle. T-0252/T-0253/T-0266 updated; no longer blocked. CON-006 (spec.md's requirement-level mirror) read consistently with the same exception.
- **CQ-014 (rule id conflict):** `PD-NFC-001`/`PD-NFC-002` ruled canonical. `spec.md`'s CON-002 verify clause corrected from `PD-NORM-001` to `PD-NFC-002`.
- **CQ-015 (FR-061 vs FR-075):** FR-061's frozen text amended to require the salted-commitment digest form. No implementation changes result, since `tasks.md` was already built against this form. T-0187, T-0190, T-0217, T-0218 updated to drop "provisional"/"pending ruling" framing.
- **CQ-016 (PageDirectory discriminant):** confirmed intentional. `data-model.md` 2.22 updated with the ruling; no discriminant assigned.
- **CQ-017 (CON-018 trial coverage):** confirmed the combined NFR-026 extracting-and-validating trial (T-0345) satisfies CON-018's obligation for both named roles. No new trial task added.

**Updated gate verdict: PASSES.** 0 orphan requirements, 0 surviving blocker findings (all 4 closed: 3 mechanically, 1 via CQ-013), 0 unresolved major findings requiring further judgment (all 7 closed: 5 mechanically, 2 via CQ-014/CQ-015), and both minor judgment items resolved (CQ-016, CQ-017). Phase 6 (implement) is unblocked.
