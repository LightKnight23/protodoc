package migrate

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// FuzzFR_120_IdentifiersSurviveMigration is T-0300's identifier-survival fuzz
// harness (FR-120). It generates a document with N distinct identifiers and
// asserts every run_id/unit_id appears in the migrated output UNCHANGED — no
// reissuing, truncation, or re-encoding of the 128-bit token. (The named test
// is a Fuzz func per the fuzz test_kind.)
func FuzzFR_120_IdentifiersSurviveMigration(f *testing.F) {
	f.Add([]byte{0x01, 0x02, 0x03})
	f.Add([]byte{0xFF, 0x00, 0x7A, 0x10, 0x22})
	f.Add([]byte{})

	target := TargetProfile{Major: 2, Representable: map[ConstructKind]bool{0x01: true, 0x02: true, 0x0E: true}}

	f.Fuzz(func(t *testing.T, seeds []byte) {
		if len(seeds) == 0 {
			return
		}
		// Build N distinct identifiers from the fuzz seeds; skip if a
		// collision would make two ids equal (we need distinctness).
		seen := map[pdlfmt.UnitID]bool{}
		var src []SourceConstruct
		for i, s := range seeds {
			var id pdlfmt.UnitID
			id[0] = s
			id[1] = byte(i)
			id[15] = s ^ 0x5a
			if seen[id] {
				continue
			}
			seen[id] = true
			src = append(src, SourceConstruct{
				Kind:     0x01,
				Location: Location{SegmentOrdinal: 0, IntraOffset: uint64(i * 32), UnitID: id, HasUnitID: true},
			})
		}
		if len(src) == 0 {
			return
		}

		_, migrated, err := Transform(src, target)
		if err != nil {
			t.Fatalf("Transform: %v", err)
		}

		// Every source id must appear, unchanged and exactly once, in output.
		outIDs := map[pdlfmt.UnitID]int{}
		for _, m := range migrated {
			outIDs[m.UnitID]++
		}
		for _, c := range src {
			id := c.Location.UnitID
			if outIDs[id] != 1 {
				t.Errorf("identifier %x survived %d times, want exactly 1 (no reissue/truncation)", id[:4], outIDs[id])
			}
		}
		if len(migrated) != len(src) {
			t.Errorf("migrated %d constructs, want %d (no drop/duplication)", len(migrated), len(src))
		}
	})
}
