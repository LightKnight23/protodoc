package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_020_FuzzRunIDPreservationAcrossOperationSequence is T-0085's named
// test: the M04 exit criterion (CP-012). It drives a fuzz target that
// generates random sequences of split / merge / move / reorder / save+load /
// undo operations over a run sequence and asserts that the SET of run_id
// lineages present is preserved through every step -- no operation loses or
// fabricates a run_id (FR-020). save/load is modelled by an
// encode-then-decode round trip of each run's identity; undo restores the
// previous snapshot.
//
// It is a native Go fuzz target so it participates in the CI fuzzing
// harness; its seed corpus plus the generated campaign give the >=10,000
// operation sequences the DoD requires.
func TestFR_020_FuzzRunIDPreservationAcrossOperationSequence(t *testing.T) {
	// A deterministic smoke campaign so the exit criterion is exercised on
	// an ordinary `go test` run too, not only under -fuzz.
	for seed := int64(0); seed < 12000; seed++ {
		runOperationCampaign(t, seed)
	}
}

// FuzzFR_020_RunIDPreservation is the native fuzz entry point wired into the
// CI fuzzing harness (pkg/fuzzmaturity discovers it). Each input seeds one
// generated operation-sequence campaign.
func FuzzFR_020_RunIDPreservation(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(42))
	f.Add(int64(1000))
	f.Fuzz(func(t *testing.T, seed int64) {
		runOperationCampaign(t, seed)
	})
}

// runOperationCampaign builds an initial run sequence and applies a
// pseudo-random sequence of identity-preserving operations, asserting after
// each that the multiset of run_id LINEAGES is exactly the initial set
// (split can add a piece sharing an existing run_id; merge removes the
// duplicate; move/reorder/save/load/undo change nothing). The invariant
// checked is: every run_id present after any step belongs to the initial
// lineage set, and every initial lineage is still reachable.
func runOperationCampaign(t *testing.T, seed int64) {
	t.Helper()
	rng := newTinyRand(uint64(seed) ^ 0x9E3779B97F4A7C15)

	// Initial runs: 1..6 runs, each a distinct minted lineage.
	n := 1 + int(rng.next()%6)
	runs := make([]Run, n)
	lineage := map[pdlfmt.UnitID]struct{}{}
	for i := range runs {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		runs[i] = Run{RunID: id, BaseOrdinal: 0, Text: "abcdef"[:1+int(rng.next()%6)], LangRef: LangRef(1)}
		lineage[id] = struct{}{}
	}

	assertLineage := func(step int, rs []Run) {
		present := map[pdlfmt.UnitID]struct{}{}
		for _, r := range rs {
			if _, ok := lineage[r.RunID]; !ok {
				t.Fatalf("seed %d step %d: run_id %x is not in the initial lineage set (fabricated)", seed, step, r.RunID)
			}
			present[r.RunID] = struct{}{}
		}
		if len(present) != len(lineage) {
			t.Fatalf("seed %d step %d: lineage set shrank from %d to %d (a run_id was lost)", seed, step, len(lineage), len(present))
		}
	}
	assertLineage(0, runs)

	// A single-level undo snapshot.
	var prev []Run

	steps := 4 + int(rng.next()%20)
	for step := 1; step <= steps; step++ {
		snapshot := append([]Run(nil), runs...)
		switch rng.next() % 6 {
		case 0: // split a random run, then immediately merge it back
			i := int(rng.next() % uint64(len(runs)))
			at := int(rng.next() % uint64(runs[i].ScalarLen()+1))
			l, r, ok := SplitRun(runs[i], at)
			if ok {
				// Replace runs[i] with l, r; lineage set unchanged (both
				// share the original run_id).
				merged := make([]Run, 0, len(runs)+1)
				merged = append(merged, runs[:i]...)
				merged = append(merged, l, r)
				merged = append(merged, runs[i+1:]...)
				runs = merged
			}
		case 1: // merge adjacent same-lineage pieces where possible
			for i := 0; i+1 < len(runs); i++ {
				if m, ok := MergeRuns(runs[i], runs[i+1]); ok {
					runs = append(append(append([]Run(nil), runs[:i]...), m), runs[i+2:]...)
					break
				}
			}
		case 2: // move
			if len(runs) >= 2 {
				from := int(rng.next() % uint64(len(runs)))
				to := int(rng.next() % uint64(len(runs)))
				if out, ok := MoveRun(runs, from, to); ok {
					runs = out
				}
			}
		case 3: // reorder (reverse)
			perm := make([]int, len(runs))
			for i := range perm {
				perm[i] = len(runs) - 1 - i
			}
			if out, ok := ReorderRuns(runs, perm); ok {
				runs = out
			}
		case 4: // save + load: identity survives an encode/decode round trip
			runs = saveLoadRuns(runs)
		case 5: // undo: restore the previous snapshot if any
			if prev != nil {
				runs = append([]Run(nil), prev...)
			}
		}
		prev = snapshot
		assertLineage(step, runs)
	}
}

// saveLoadRuns models a save/load cycle at the identity level: each run's
// run_id is carried through an encode/decode round trip unchanged.
func saveLoadRuns(runs []Run) []Run {
	out := make([]Run, len(runs))
	for i, r := range runs {
		var enc []byte
		enc = pdlfmt.AppendUnitID(enc, r.RunID)
		id, _, err := pdlfmt.DecodeUnitID(enc)
		if err != nil {
			panic("saveLoadRuns: round-trip decode failed: " + err.Error())
		}
		out[i] = Run{RunID: id, BaseOrdinal: r.BaseOrdinal, Text: r.Text, LangRef: r.LangRef}
	}
	return out
}

// tinyRand is a small deterministic splitmix64 PRNG (no external dependency,
// no global rand state) used to drive the operation campaign reproducibly.
type tinyRand struct{ state uint64 }

func newTinyRand(seed uint64) *tinyRand { return &tinyRand{state: seed} }

func (r *tinyRand) next() uint64 {
	r.state += 0x9E3779B97F4A7C15
	z := r.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}
