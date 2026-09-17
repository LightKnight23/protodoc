package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_022_PasteMintsFreshIdentity is T-0072's named test. It confirms
// PasteContent and DuplicateRange always produce runs whose run_id is
// freshly minted and distinct from every source run_id, for single-run and
// multi-run payloads, while carrying the source text/base_ordinal through
// unchanged.
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
				fn    func([]Run) ([]Run, error)
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
					// Fresh id: not equal to ANY source id.
					if _, isSource := srcIDs[r.RunID]; isSource {
						t.Fatalf("%s: produced run %d reused a source run_id", op.label, i)
					}
					// Distinct among produced runs too.
					if _, dup := produced[r.RunID]; dup {
						t.Fatalf("%s: produced run %d duplicates another produced run_id", op.label, i)
					}
					produced[r.RunID] = struct{}{}
					// Content rides through unchanged.
					if r.Text != source[i].Text || r.BaseOrdinal != source[i].BaseOrdinal {
						t.Fatalf("%s: produced run %d changed content/base_ordinal", op.label, i)
					}
				}
			}
		})
	}
}
