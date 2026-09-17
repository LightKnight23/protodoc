package mint

import (
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

// makeRunSeq builds n runs with distinct minted run_ids and marker text.
func makeRunSeq(t *testing.T, n int) []content.Run {
	t.Helper()
	runs := make([]content.Run, n)
	for i := range runs {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		runs[i] = content.Run{RunID: id, BaseOrdinal: uint32(i * 10), Text: string(rune('A' + i)), LangRef: content.LangRef(1)}
	}
	return runs
}

// TestFR_022_PasteMintsFreshIdentity is T-0072's named test. PasteContent and
// DuplicateRange always produce runs whose run_id is freshly minted and
// distinct from every source run_id, for single-run and multi-run payloads,
// while carrying the source text/base_ordinal through unchanged.
func TestFR_022_PasteMintsFreshIdentity(t *testing.T) {
	cases := []struct {
		name string
		n    int
	}{
		{"single run", 1},
		{"multi run", 8},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			source := makeRunSeq(t, c.n)
			srcIDs := sourceRunIDSet(source)

			for _, op := range []struct {
				label string
				fn    func([]content.Run) ([]content.Run, error)
			}{
				{"DuplicateRange", DuplicateRange},
				{"PasteContent", PasteContent},
			} {
				out, err := op.fn(source)
				if err != nil {
					t.Fatalf("%s: %v", op.label, err)
				}
				if len(out) != c.n {
					t.Fatalf("%s: produced %d runs, want %d", op.label, len(out), c.n)
				}
				produced := make(map[pdlfmt.UnitID]struct{}, c.n)
				for i, r := range out {
					if _, isSource := srcIDs[r.RunID]; isSource {
						t.Fatalf("%s: produced run %d reused a source run_id", op.label, i)
					}
					if _, dup := produced[r.RunID]; dup {
						t.Fatalf("%s: produced run %d duplicates another produced run_id", op.label, i)
					}
					produced[r.RunID] = struct{}{}
					if r.Text != source[i].Text || r.BaseOrdinal != source[i].BaseOrdinal {
						t.Fatalf("%s: produced run %d changed content/base_ordinal", op.label, i)
					}
				}
			}
		})
	}
}
