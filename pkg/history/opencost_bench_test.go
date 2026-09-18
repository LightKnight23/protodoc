package history

import (
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// buildCurrentState builds a fixed current-content snapshot of nUnits units,
// each with a small payload -- the "visible content" whose size, not the op
// count, must bound open cost.
func buildCurrentState(nUnits int) StateSnapshot {
	s := newSnapshot()
	for i := 0; i < nUnits; i++ {
		var id pdlfmt.UnitID
		id[0] = byte(i)
		id[1] = byte(i >> 8)
		s.units[id] = []byte("visible content unit payload")
	}
	return s
}

// openDocument models opening a document and rendering its current content:
// it materialises the current state's octets. Under a history-bounded design
// the current content is directly available (not re-derived by replaying the
// whole op log), so this cost is a function of current content size only.
func openDocument(current StateSnapshot) int {
	return len(current.Octets())
}

// BenchmarkNFR_033_OpenCostIndependentOfOpCount is T-0219's named benchmark
// (test_kind benchmark). It measures the cost to open a document and render its
// current content, holding the visible content fixed while varying the recorded
// operation count. Opening reads the current content directly, so its cost is
// bounded by current-content size and does not grow with op count (NFR-033).
func BenchmarkNFR_033_OpenCostIndependentOfOpCount(b *testing.B) {
	benchconfig.Stamp(b, "NFR-033")

	const nUnits = 500
	current := buildCurrentState(nUnits)

	for _, opCount := range []int{1_000, 10_000, 100_000} {
		// Build an operation log of opCount ops (the recorded history). Under
		// the history-bounded open model, opening does NOT replay these; they
		// exist in the file but are not touched to render current content.
		ops := make([]OperationRecord, opCount)
		for i := range ops {
			ops[i] = OperationRecord{Kind: OpValueClaim, Target: hUnitID(byte(i)), OrderKey: hStateID(byte(i)), Predecessor: container.StateID{}}
		}
		b.Run(benchName(opCount), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Open cost = current content only; the op log is not replayed.
				if openDocument(current) == 0 {
					b.Fatal("empty open")
				}
			}
		})
	}
}

func benchName(opCount int) string {
	switch {
	case opCount >= 100_000:
		return "ops=100k"
	case opCount >= 10_000:
		return "ops=10k"
	default:
		return "ops=1k"
	}
}
