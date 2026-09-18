package migrate

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// kinds used in the tests.
const (
	kindTextBlock     ConstructKind = 0x01
	kindRetiredExtTok ConstructKind = 0x1001
	kindBadCrypto     ConstructKind = 0x1002
)

func mkUnit(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = seed
	return id
}

// targetV1 represents a target major version representing only TextBlock.
func targetV1() TargetProfile {
	return TargetProfile{Major: 1, Representable: map[ConstructKind]bool{kindTextBlock: true}}
}

// TestFR_121_RefusalHaltsAtFirstUnrepresentableConstruct is T-0296's named unit
// test (FR-121). Scan returns an empty report when all constructs are
// representable; names the single unrepresentable construct with its exact
// location; and with two unrepresentable constructs names only the first in
// traversal order.
func TestFR_121_RefusalHaltsAtFirstUnrepresentableConstruct(t *testing.T) {
	target := targetV1()

	// Zero unrepresentable -> empty report (proceed).
	clean := []SourceConstruct{
		{Kind: kindTextBlock, KindName: "TEXT_BLOCK", Location: Location{SegmentOrdinal: 0, IntraOffset: 0}},
		{Kind: kindTextBlock, KindName: "TEXT_BLOCK", Location: Location{SegmentOrdinal: 0, IntraOffset: 64}},
	}
	if r := Scan(clean, target); r.Refused {
		t.Errorf("clean source should proceed, got refusal %+v", r)
	}

	// Exactly one unrepresentable -> named with location.
	one := []SourceConstruct{
		{Kind: kindTextBlock, KindName: "TEXT_BLOCK", Location: Location{SegmentOrdinal: 0, IntraOffset: 0}},
		{Kind: kindRetiredExtTok, KindName: "RETIRED_EXT_TOKEN", Location: Location{SegmentOrdinal: 1, IntraOffset: 32, UnitID: mkUnit(0x7), HasUnitID: true}},
	}
	r := Scan(one, target)
	if !r.Refused || r.Kind != kindRetiredExtTok || r.Location.SegmentOrdinal != 1 || r.Location.IntraOffset != 32 {
		t.Fatalf("single unrepresentable: report = %+v, want refusal naming the retired token at seg 1 off 32", r)
	}

	// Two distinct unrepresentable in traversal order -> only the FIRST named.
	two := []SourceConstruct{
		{Kind: kindBadCrypto, KindName: "OUT_OF_ALLOWLIST_CRYPTO", Location: Location{SegmentOrdinal: 0, IntraOffset: 16}},
		{Kind: kindRetiredExtTok, KindName: "RETIRED_EXT_TOKEN", Location: Location{SegmentOrdinal: 2, IntraOffset: 8}},
	}
	r2 := Scan(two, target)
	if !r2.Refused || r2.Kind != kindBadCrypto {
		t.Errorf("two unrepresentable: report = %+v, want only the first (crypto) named", r2)
	}
}
