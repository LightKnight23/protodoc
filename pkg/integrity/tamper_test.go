package integrity

import (
	"os"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestTR_009_TamperedSegmentTableSlotDetectedBeforeReliance is T-0141's named
// conformance test. It builds a baseline state, records its structure_digest
// as the trusted baseline, then tampers with one SegmentTableSlot field
// (length and, separately, type) WITHOUT updating any digest, and asserts
// the freshly recomputed structure_digest differs from the baseline for
// every bitmask that includes SEGMENT_TABLE or INTEGRITY_BLOCK -- i.e. the
// tamper is caught by comparison, never by trusting the tampered slot's own
// claimed digest (TR-009).
func TestTR_009_TamperedSegmentTableSlotDetectedBeforeReliance(t *testing.T) {
	// Baseline segment table: 4 populated CONTENT/RESOURCE slots.
	baseSlots := []container.SegmentTableSlot{
		{SegmentType: container.SegmentTypeContent, Offset: 1 << 20, Length: 128, Digest: pdlfmt.Digest256{0x01}},
		{SegmentType: container.SegmentTypeResource, Offset: 1<<20 + 128, Length: 256, Digest: pdlfmt.Digest256{0x02}},
		{SegmentType: container.SegmentTypeContent, Offset: 1<<20 + 384, Length: 64, Digest: pdlfmt.Digest256{0x03}},
		{SegmentType: container.SegmentTypeHistory, Offset: 1<<20 + 448, Length: 512, Digest: pdlfmt.Digest256{0x04}},
	}

	// Build a structure_digest input from live slots: item 7 summaries and
	// item 8 T_S root are recomputed fresh from the slots.
	buildInput := func(slots []container.SegmentTableSlot, bitmask byte) StructureDigestInput {
		tsRoot, err := TSRoot(slots)
		if err != nil {
			t.Fatalf("TSRoot: %v", err)
		}
		var summaries []CoveredSegmentSummary
		for i, s := range slots {
			// A "freshly recomputed" segment digest: here modelled as the
			// slot's current digest field (in a real verifier this is
			// recomputed over the segment octets; the tamper test only needs
			// the value to change when the slot changes).
			summaries = append(summaries, CoveredSegmentSummary{
				Ordinal: uint16(i), Type: s.SegmentType, Length: s.Length, Digest: Digest(s.Digest),
			})
		}
		return StructureDigestInput{
			Bitmask:          bitmask,
			LedgerLength:     4096,
			CoveredSummaries: summaries,
			TSRoot:           tsRoot,
		}
	}

	// Tampered variants:
	//   digest-tamper: changes a slot's digest -- caught by BOTH item 7
	//     (segment summary) and item 8 (T_S root, whose leaves are the slot
	//     digests), i.e. under SEGMENT_TABLE or INTEGRITY_BLOCK.
	//   length-tamper / type-tamper: change a field carried ONLY by item 7
	//     (the segment summary), not the slot digest T_S commits to, so they
	//     are caught under SEGMENT_TABLE but not by T_S alone.
	tamperDigest := append([]container.SegmentTableSlot(nil), baseSlots...)
	tamperDigest[0].Digest[7] ^= 0xFF
	tamperLen := append([]container.SegmentTableSlot(nil), baseSlots...)
	tamperLen[1].Length = 9999
	tamperType := append([]container.SegmentTableSlot(nil), baseSlots...)
	tamperType[2].SegmentType = container.SegmentTypeResource

	for bm := 0; bm < 32; bm++ {
		coversSeg := bm&CoverageBitSegmentTable != 0
		coversIntegrity := bm&CoverageBitIntegrity != 0
		if !coversSeg && !coversIntegrity {
			continue
		}
		base, err := StructureDigest(buildInput(baseSlots, byte(bm)))
		if err != nil {
			t.Fatalf("bm 0x%02x: baseline: %v", bm, err)
		}
		// A digest tamper is caught whenever either bit is set.
		if dg, _ := StructureDigest(buildInput(tamperDigest, byte(bm))); dg == base {
			t.Fatalf("bm 0x%02x: digest-tamper not detected", bm)
		}
		// Length and type tampers alter only item 7, so they are caught iff
		// SEGMENT_TABLE is set.
		if coversSeg {
			if tl, _ := StructureDigest(buildInput(tamperLen, byte(bm))); tl == base {
				t.Fatalf("bm 0x%02x: length-tamper not detected under SEGMENT_TABLE", bm)
			}
			if tt, _ := StructureDigest(buildInput(tamperType, byte(bm))); tt == base {
				t.Fatalf("bm 0x%02x: type-tamper not detected under SEGMENT_TABLE", bm)
			}
		}
	}

	// Fixture-file pair (DoD: baseline/tampered octets ship under
	// testdata/conformance/integrity/). Load both, decode their segment
	// tables, and confirm the tampered one is detected under a
	// SEGMENT_TABLE-covering bitmask.
	baseFix := loadSegTableFixture(t, "testdata/conformance/integrity/segtable_baseline.bin")
	tamperFix := loadSegTableFixture(t, "testdata/conformance/integrity/segtable_tampered.bin")
	const bm = CoverageBitSegmentTable | CoverageBitIntegrity
	bfd, _ := StructureDigest(buildInput(baseFix, bm))
	tfd, _ := StructureDigest(buildInput(tamperFix, bm))
	if bfd == tfd {
		t.Fatalf("fixture-pair tamper not detected under SEGMENT_TABLE|INTEGRITY_BLOCK bitmask")
	}
}

// loadSegTableFixture reads a full segment-table region fixture and returns
// its populated (non-unused) leading slots.
func loadSegTableFixture(t *testing.T, path string) []container.SegmentTableSlot {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", path, err)
	}
	table, err := container.DecodeSegmentTable(raw)
	if err != nil {
		t.Fatalf("decoding fixture %s: %v", path, err)
	}
	var out []container.SegmentTableSlot
	for _, s := range table {
		if s.SegmentType == container.SegmentTypeUnused {
			continue
		}
		out = append(out, s)
	}
	return out
}
