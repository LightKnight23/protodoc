package container

import (
	"reflect"
	"testing"
)

// TestTR_007_CoverageHintFromBoundedPrefix is T-0022's named test.
// Implements: TR-007.
//
// It builds a fixture with a SUBSET-mode signature (5 populated slots:
// CONTENT, CONTENT, RESOURCE, ATTEST, CONTENT, covering ordinals {0,1,4}
// and leaving ordinal 2 declared uncovered), wires fm-coverage-summary and
// SegmentTableSlot.Flags to agree with it, and confirms the bounded-prefix
// determination matches what decoding "the actual CoverageDescriptor"
// (simulated here, since the Signature/CoverageDescriptor types are a later
// milestone's integrity-layer task, not yet built) would produce.
func TestTR_007_CoverageHintFromBoundedPrefix(t *testing.T) {
	slots := []SegmentTableSlot{
		fixtureSlot(1), // ordinal 0: CONTENT, covered
		fixtureSlot(2), // ordinal 1: CONTENT, covered
		fixtureSlot(3), // ordinal 2: RESOURCE, uncovered
		fixtureSlot(4), // ordinal 3: ATTEST, excluded from coverage entirely
		fixtureSlot(5), // ordinal 4: CONTENT, covered
	}
	slots[2].SegmentType = SegmentTypeResource
	slots[3].SegmentType = SegmentTypeAttest

	// The "actual CoverageDescriptor" a full reader would decode from the
	// covering SIGNATURE segment (integrity.abnf S4): SUBSET mode, covering
	// ordinals 0-1 and 4, leaving ordinal 2 uncovered. Ordinal 3 (ATTEST) is
	// never nameable in either list (PD-COVER-004), so it appears in neither.
	actualCovered := []SegmentRange{{Start: 0, End: 2}, {Start: 4, End: 5}}
	actualUncovered := []SegmentRange{{Start: 2, End: 3}}
	const actualBitmask = CoverageBitHeader | CoverageBitSegmentTable

	// Set each covered slot's coverage-hint bit (bit0) to match the
	// descriptor this fixture mirrors, per contracts/container.abnf S5:
	// the hint is meant to agree with the authoritative coverage, even
	// though it is never trusted on its own.
	for _, i := range []int{0, 1, 4} {
		slots[i].Flags |= SlotFlagCoverageHint
	}
	slots[2].Flags &^= SlotFlagCoverageHint

	descriptorDigest := ComputeCoverageDescriptorDigest(CoverageModeSubset, actualCovered, actualUncovered, actualBitmask)
	summary := []CoverageSummaryEntry{{
		Mode:                 CoverageModeSubset,
		CoveredRanges:        actualCovered,
		UncoveredRanges:      actualUncovered,
		PrefixRegionsBitmask: actualBitmask,
		DescriptorDigest:     descriptorDigest,
	}}

	segTable, err := EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	fm := &Frontmatter{Title: "coverage fixture", CoverageSummary: summary}
	fmEnc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}

	prefix := make([]byte, SegmentTableOffset+SegmentTableRegionSize)
	copy(prefix[FrontmatterOffset:], fmEnc)
	copy(prefix[SegmentTableOffset:], segTable)

	got, err := StoredUnitCoverageFromBoundedPrefix(prefix)
	if err != nil {
		t.Fatalf("StoredUnitCoverageFromBoundedPrefix: %v", err)
	}

	wantCovered := []int{0, 1, 4}
	wantUncovered := []int{2}
	if !reflect.DeepEqual(got.Covered, wantCovered) {
		t.Errorf("Covered = %v, want %v", got.Covered, wantCovered)
	}
	if !reflect.DeepEqual(got.Uncovered, wantUncovered) {
		t.Errorf("Uncovered = %v, want %v", got.Uncovered, wantUncovered)
	}
	if len(got.HintMismatches) != 0 {
		t.Errorf("HintMismatches = %v, want none (fixture's flags agree with the summary)", got.HintMismatches)
	}

	// Cross-check against the "actual CoverageDescriptor" independently:
	// re-deriving covered/uncovered from actualCovered/actualUncovered
	// directly must agree with what the bounded-prefix path produced.
	wantFromDescriptor := map[int]bool{0: true, 1: true, 4: true}
	for _, ord := range got.Covered {
		if !wantFromDescriptor[ord] {
			t.Errorf("ordinal %d reported covered, not covered per the actual descriptor", ord)
		}
	}
	if len(got.Covered) != len(wantFromDescriptor) {
		t.Errorf("got %d covered ordinals, descriptor names %d", len(got.Covered), len(wantFromDescriptor))
	}

	// Ordinal 3 (ATTEST) must appear in neither list.
	for _, ord := range append(append([]int{}, got.Covered...), got.Uncovered...) {
		if ord == 3 {
			t.Errorf("ATTEST ordinal 3 must never appear in either covered or uncovered set (PD-COVER-004)")
		}
	}
}

// TestTR_007_CoverageHintMismatchReported confirms a slot whose flags
// coverage-hint bit disagrees with fm-coverage-summary's declared coverage
// is reported in HintMismatches, per slot-flags never being authoritative
// on its own.
func TestTR_007_CoverageHintMismatchReported(t *testing.T) {
	slots := []SegmentTableSlot{fixtureSlot(1)}
	slots[0].Flags &^= SlotFlagCoverageHint // hint says "not covered"

	covered := []SegmentRange{{Start: 0, End: 1}} // summary says "covered"
	summary := []CoverageSummaryEntry{{
		Mode:                 CoverageModeSubset,
		CoveredRanges:        covered,
		UncoveredRanges:      nil,
		PrefixRegionsBitmask: 0,
		DescriptorDigest:     ComputeCoverageDescriptorDigest(CoverageModeSubset, covered, nil, 0),
	}}

	var table [MaxSegments]SegmentTableSlot
	table[0] = slots[0]

	result := ComputeStoredUnitCoverage(table, summary)
	if !reflect.DeepEqual(result.Covered, []int{0}) {
		t.Fatalf("Covered = %v, want [0]", result.Covered)
	}
	if !reflect.DeepEqual(result.HintMismatches, []int{0}) {
		t.Fatalf("HintMismatches = %v, want [0]", result.HintMismatches)
	}
}

// TestTR_007_CoverageModeTotalCoversAllNonAttest confirms TOTAL mode
// covers every currently-populated non-ATTEST ordinal (integrity.abnf S4
// cd-mode NORMATIVE comment) without naming ranges explicitly.
func TestTR_007_CoverageModeTotalCoversAllNonAttest(t *testing.T) {
	var table [MaxSegments]SegmentTableSlot
	table[0] = fixtureSlot(1)
	table[1] = fixtureSlot(2)
	table[1].SegmentType = SegmentTypeAttest
	table[2] = fixtureSlot(3)

	summary := []CoverageSummaryEntry{{Mode: CoverageModeTotal}}
	result := ComputeStoredUnitCoverage(table, summary)

	if !reflect.DeepEqual(result.Covered, []int{0, 2}) {
		t.Fatalf("Covered = %v, want [0 2]", result.Covered)
	}
	if len(result.Uncovered) != 0 {
		t.Fatalf("Uncovered = %v, want none", result.Uncovered)
	}
}

// TestCoverageSummaryRoundTrip confirms CoverageSummaryEntry's encoding is
// self-delimiting and round-trips byte-exact through Frontmatter.
func TestCoverageSummaryRoundTrip(t *testing.T) {
	entries := []CoverageSummaryEntry{
		{
			Mode:                 CoverageModeSubset,
			CoveredRanges:        []SegmentRange{{Start: 0, End: 3}, {Start: 10, End: 12}},
			UncoveredRanges:      []SegmentRange{{Start: 3, End: 10}},
			PrefixRegionsBitmask: CoverageBitHeader | CoverageBitFrontmatter,
		},
		{
			Mode: CoverageModeTotal,
		},
	}
	entries[0].DescriptorDigest = ComputeCoverageDescriptorDigest(entries[0].Mode, entries[0].CoveredRanges, entries[0].UncoveredRanges, entries[0].PrefixRegionsBitmask)
	entries[1].DescriptorDigest = ComputeCoverageDescriptorDigest(entries[1].Mode, nil, nil, 0)

	fm := &Frontmatter{CoverageSummary: entries}
	enc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := DecodeFrontmatter(enc)
	if err != nil {
		t.Fatalf("DecodeFrontmatter: %v", err)
	}
	if !reflect.DeepEqual(got.CoverageSummary, entries) {
		t.Fatalf("CoverageSummary round-trip mismatch:\n got=%+v\nwant=%+v", got.CoverageSummary, entries)
	}
}

func TestCoverageSummaryRejectsBadRange(t *testing.T) {
	entries := []CoverageSummaryEntry{{
		Mode:          CoverageModeSubset,
		CoveredRanges: []SegmentRange{{Start: 5, End: 5}}, // zero-length: rejected, not silently dropped
	}}
	if _, err := EncodeCoverageSummary(entries); err == nil {
		t.Fatal("expected error encoding a zero-length segment-range")
	}
}

func TestCoverageSummaryRejectsInvalidMode(t *testing.T) {
	entries := []CoverageSummaryEntry{{Mode: CoverageMode(2)}}
	if _, err := EncodeCoverageSummary(entries); err == nil {
		t.Fatal("expected error encoding an out-of-enum coverage mode")
	}
}

func TestCoverageSummaryRejectsReservedBitmaskBits(t *testing.T) {
	entries := []CoverageSummaryEntry{{Mode: CoverageModeTotal, PrefixRegionsBitmask: 0x20}}
	if _, err := EncodeCoverageSummary(entries); err == nil {
		t.Fatal("expected error encoding a nonzero reserved bitmask bit")
	}
}
