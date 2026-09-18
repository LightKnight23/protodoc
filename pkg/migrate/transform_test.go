package migrate

import (
	"bytes"
	"testing"
)

// targetV2Full represents a target that can represent all the kinds the test
// uses, so Transform proceeds.
func targetV2Full() TargetProfile {
	return TargetProfile{Major: 2, Representable: map[ConstructKind]bool{
		kindTextBlock: true, ConstructKind(0x02): true, ConstructKind(0x0E): true,
	}}
}

// TestFR_119_TransformDeterministicAcrossInvocations is T-0298's named unit
// test (FR-119). Transform of a Phase-1-clean source produces byte-identical
// canonical output across invocations, independent of the input slice order.
func TestFR_119_TransformDeterministicAcrossInvocations(t *testing.T) {
	target := targetV2Full()
	src := []SourceConstruct{
		{Kind: ConstructKind(0x0E), Location: Location{SegmentOrdinal: 0, IntraOffset: 200, UnitID: mkUnit(0x3), HasUnitID: true}},
		{Kind: kindTextBlock, Location: Location{SegmentOrdinal: 0, IntraOffset: 0, UnitID: mkUnit(0x1), HasUnitID: true}},
		{Kind: ConstructKind(0x02), Location: Location{SegmentOrdinal: 0, IntraOffset: 100, UnitID: mkUnit(0x2), HasUnitID: true}},
	}

	out1, _, err := Transform(src, target)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	out2, _, err := Transform(src, target)
	if err != nil {
		t.Fatalf("Transform (2nd): %v", err)
	}
	if !bytes.Equal(out1, out2) {
		t.Errorf("Transform not deterministic across invocations")
	}

	// Same constructs in a DIFFERENT input slice order -> identical output
	// (the traversal ordering, not the input order, fixes the bytes).
	shuffled := []SourceConstruct{src[1], src[2], src[0]}
	out3, _, err := Transform(shuffled, target)
	if err != nil {
		t.Fatalf("Transform (shuffled): %v", err)
	}
	if !bytes.Equal(out1, out3) {
		t.Errorf("Transform output depends on input slice order; must depend only on traversal order")
	}

	// A source Scan would refuse cannot be transformed.
	dirty := []SourceConstruct{{Kind: kindRetiredExtTok, Location: Location{SegmentOrdinal: 0, IntraOffset: 0}}}
	if _, _, err := Transform(dirty, targetV1()); err != ErrSourceNotClean {
		t.Errorf("Transform of an unclean source: err = %v, want ErrSourceNotClean", err)
	}
}
