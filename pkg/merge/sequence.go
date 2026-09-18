// Sequence merge (T-0227/T-0228, FR-093). Text inserted by one author at one
// position in one contiguous burst must appear as an UNBROKEN substring in
// every merged result -- Fugue's proven non-interleaving property. A burst is a
// run of characters sharing one run_id with consecutive base_ordinals; a
// concurrent insertion from another author at the same anchor is ordered
// ENTIRELY before or after the burst (by the R2 state-id tiebreak on the
// bursts' authoring states), never split into it. This models the merge order
// over sequence elements; it does not re-implement all of Fugue, but enforces
// the specific non-interleaving guarantee FR-093 requires: same-run_id runs
// stay contiguous.
package merge

import (
	"bytes"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// SeqElement is one inserted sequence element (a character or run-element): its
// run_id (shared across a contiguous burst), its base_ordinal within that run
// (consecutive within a burst), the authoring state-id (the burst's R2 tiebreak
// key), and its scalar value.
type SeqElement struct {
	RunID       pdlfmt.UnitID
	BaseOrdinal uint32
	StateID     [32]byte // authoring state; the R2 tiebreak across bursts
	Value       byte
}

// MergeSequence orders a set of concurrently-inserted sequence elements into
// the single Fugue-consistent merged order. Elements of the same run_id (one
// author's contiguous burst) are kept together in base_ordinal order; distinct
// runs are ordered against each other by their authoring state-id (unsigned
// big-endian, the R2 tiebreak), so a whole burst sorts before or after another
// whole burst -- never interleaved. The result is deterministic.
func MergeSequence(elems []SeqElement) []SeqElement {
	// Group by run_id, preserving each run's internal base_ordinal order.
	runs := map[pdlfmt.UnitID][]SeqElement{}
	order := []pdlfmt.UnitID{}
	for _, e := range elems {
		if _, ok := runs[e.RunID]; !ok {
			order = append(order, e.RunID)
		}
		runs[e.RunID] = append(runs[e.RunID], e)
	}
	for _, r := range order {
		g := runs[r]
		sort.SliceStable(g, func(i, j int) bool {
			if g[i].BaseOrdinal != g[j].BaseOrdinal {
				return g[i].BaseOrdinal < g[j].BaseOrdinal
			}
			// Two elements sharing (run_id, base_ordinal) is a malformed input
			// (a character identity is unique), but the merge must still be
			// TOTAL and deterministic: tiebreak by authoring state-id, then by
			// value, so the order never depends on input order.
			if c := bytes.Compare(g[i].StateID[:], g[j].StateID[:]); c != 0 {
				return c < 0
			}
			return g[i].Value < g[j].Value
		})
		runs[r] = g
	}
	// Order the runs against each other by their authoring state-id (the
	// state-id of the run's first element -- a burst shares one authoring
	// state), unsigned big-endian. This total order places each whole burst
	// as a unit, so no burst is split by another.
	sort.SliceStable(order, func(i, j int) bool {
		a := runs[order[i]][0].StateID
		b := runs[order[j]][0].StateID
		if c := bytes.Compare(a[:], b[:]); c != 0 {
			return c < 0
		}
		// Two distinct runs sharing one authoring state-id tie on R2; break
		// deterministically by run_id (unique per run, unsigned big-endian) so
		// the whole-burst order never depends on input order.
		ri, rj := order[i], order[j]
		return bytes.Compare(ri[:], rj[:]) < 0
	})
	// Concatenate the ordered runs.
	var out []SeqElement
	for _, r := range order {
		out = append(out, runs[r]...)
	}
	return out
}

// BurstSubstring returns the value bytes of the run with the given run_id, in
// base_ordinal order -- the burst's own text, for substring checks.
func BurstSubstring(elems []SeqElement, runID pdlfmt.UnitID) []byte {
	var g []SeqElement
	for _, e := range elems {
		if e.RunID == runID {
			g = append(g, e)
		}
	}
	sort.SliceStable(g, func(i, j int) bool {
		if g[i].BaseOrdinal != g[j].BaseOrdinal {
			return g[i].BaseOrdinal < g[j].BaseOrdinal
		}
		if c := bytes.Compare(g[i].StateID[:], g[j].StateID[:]); c != 0 {
			return c < 0
		}
		return g[i].Value < g[j].Value
	})
	var b []byte
	for _, e := range g {
		b = append(b, e.Value)
	}
	return b
}

// MergedText returns the merged sequence's value bytes in order.
func MergedText(merged []SeqElement) []byte {
	b := make([]byte, len(merged))
	for i, e := range merged {
		b[i] = e.Value
	}
	return b
}
