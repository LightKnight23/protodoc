package migrate

import "testing"

// Additional unrepresentable construct classes for the golden corpus.
const (
	kindOverCeiling   ConstructKind = 0x1003 // exceeds target major's structural ceiling
	kindDroppedRecord ConstructKind = 0x1004 // a record kind removed in the target major
)

// TestFR_121_ConformanceVectors_UnrepresentableConstructs is T-0297's named
// conformance test (FR-121). A golden corpus of source documents, each with
// exactly one class of unrepresentable construct, each paired with its expected
// RefusalReport (construct kind + location). A fully-representable control
// proceeds.
func TestFR_121_ConformanceVectors_UnrepresentableConstructs(t *testing.T) {
	target := targetV1() // representable set = {TEXT_BLOCK}

	type vector struct {
		name       string
		source     []SourceConstruct
		wantRefuse bool
		wantKind   ConstructKind
		wantSeg    uint16
		wantOffset uint64
	}

	vectors := []vector{
		{
			name:       "control-representable",
			source:     []SourceConstruct{{Kind: kindTextBlock, KindName: "TEXT_BLOCK", Location: Location{SegmentOrdinal: 0, IntraOffset: 0}}},
			wantRefuse: false,
		},
		{
			name: "retired-extension-token",
			source: []SourceConstruct{
				{Kind: kindTextBlock, KindName: "TEXT_BLOCK", Location: Location{SegmentOrdinal: 0, IntraOffset: 0}},
				{Kind: kindRetiredExtTok, KindName: "RETIRED_EXT_TOKEN", Location: Location{SegmentOrdinal: 0, IntraOffset: 128}},
			},
			wantRefuse: true, wantKind: kindRetiredExtTok, wantSeg: 0, wantOffset: 128,
		},
		{
			name: "out-of-allowlist-crypto",
			source: []SourceConstruct{
				{Kind: kindBadCrypto, KindName: "OUT_OF_ALLOWLIST_CRYPTO", Location: Location{SegmentOrdinal: 3, IntraOffset: 0}},
			},
			wantRefuse: true, wantKind: kindBadCrypto, wantSeg: 3, wantOffset: 0,
		},
		{
			name: "exceeds-structural-ceiling",
			source: []SourceConstruct{
				{Kind: kindOverCeiling, KindName: "OVER_STRUCTURAL_CEILING", Location: Location{SegmentOrdinal: 1, IntraOffset: 4096}},
			},
			wantRefuse: true, wantKind: kindOverCeiling, wantSeg: 1, wantOffset: 4096,
		},
		{
			name: "dropped-record-kind",
			source: []SourceConstruct{
				{Kind: kindDroppedRecord, KindName: "DROPPED_RECORD_KIND", Location: Location{SegmentOrdinal: 2, IntraOffset: 64}},
			},
			wantRefuse: true, wantKind: kindDroppedRecord, wantSeg: 2, wantOffset: 64,
		},
	}

	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			r := Scan(v.source, target)
			if r.Refused != v.wantRefuse {
				t.Fatalf("%s: Refused=%v, want %v (report %+v)", v.name, r.Refused, v.wantRefuse, r)
			}
			if !v.wantRefuse {
				return
			}
			if r.Kind != v.wantKind || r.Location.SegmentOrdinal != v.wantSeg || r.Location.IntraOffset != v.wantOffset {
				t.Errorf("%s: report = {kind %d @ seg %d off %d}, want {kind %d @ seg %d off %d}",
					v.name, r.Kind, r.Location.SegmentOrdinal, r.Location.IntraOffset, v.wantKind, v.wantSeg, v.wantOffset)
			}
		})
	}
}
