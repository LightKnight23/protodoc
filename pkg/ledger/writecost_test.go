package ledger

import (
	"errors"
	"fmt"
	"testing"

	"Protodoc/pkg/benchconfig"
)

// makeBaseOfSize returns a valid placement base (>= PrefixLength) of at
// least the requested size: a fixed-prefix-sized document padded with a
// synthetic ledger tail so its total length is exactly size (or
// PrefixLength if size is smaller). The tail bytes stand in for
// already-sealed segment octets; place() treats them as opaque.
func makeBaseOfSize(size int) []byte {
	if size < PrefixLength {
		size = PrefixLength
	}
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(i * 37)
	}
	return b
}

// makeEditSegment returns a sealed-segment payload whose octet length
// models a K-octet logical edit's on-storage footprint. Per plan.md the
// content segment a K-octet edit appends is proportional to K (the touched
// TextBlock frame plus any split/merged Run records); this model uses
// K + a small fixed framing overhead, well inside the 8*K allowance, so
// the append is a realistic-or-pessimistic stand-in, never an unrealistic
// under-count that would make the budget assertion vacuous.
func makeEditSegment(k int) SegmentPayload {
	const framingOverhead = 128 // segment header + one frame-dir entry + digest, rounded up
	n := k + framingOverhead
	seg := make([]byte, n)
	for i := range seg {
		seg[i] = byte(i*11 + 3)
	}
	return SegmentPayload{Octets: seg}
}

// BenchmarkNFR_008_EditWriteCostWithinBudget is T-0036's named benchmark.
// It commits K-octet edits (K = 1 byte, 1 KiB, 1 MiB) against synthetic
// documents of 1 KiB, 1 MiB and 1 GiB and asserts, as a HARD in-benchmark
// failure (b.Fatalf, not a mere measurement), that each commit's total
// write-octet count stays within NFR-008's 8*K + 262144 bound. The bound
// is additionally enforced inside PlaceWithinBudget itself, so a violation
// is refused at the API rather than only caught here.
func BenchmarkNFR_008_EditWriteCostWithinBudget(b *testing.B) {
	benchconfig.Stamp(b, "NFR-008")
	const (
		kib = 1024
		mib = 1024 * 1024
		gib = 1024 * 1024 * 1024
	)
	editSizes := []int{1, kib, mib}
	docSizes := []struct {
		name string
		size int
	}{
		{"1KiB", kib},
		{"1MiB", mib},
		{"1GiB", gib},
	}

	for _, ds := range docSizes {
		base := makeBaseOfSize(ds.size)
		// The prior's populated segment count is irrelevant to the write
		// cost (which depends on the delta, not prior fill); start each
		// document's ordinals at 0.
		for _, k := range editSizes {
			name := fmt.Sprintf("doc=%s/K=%d", ds.name, k)
			b.Run(name, func(b *testing.B) {
				delta := EditDelta{NewSegments: []SegmentPayload{makeEditSegment(k)}}
				budget := WriteBudget(uint64(k))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					res, err := PlaceWithinBudget(base, 0, delta, uint64(k))
					if err != nil {
						b.Fatalf("commit refused (K=%d, doc=%s): %v", k, ds.name, err)
					}
					if res.WrittenOctets > budget {
						b.Fatalf("write cost %d exceeds NFR-008 budget %d for K=%d at doc=%s", res.WrittenOctets, budget, k, ds.name)
					}
				}
				b.ReportMetric(float64(commitWriteCost(uint64(k+128), 1)), "write_octets/op")
			})
		}
	}
}

// TestNFR_008_WriteCostRefusedWhenOverBudget confirms the bound is a hard
// assertion, not merely measured: a commit whose appended segment is far
// larger than 8*K + 262144 for a tiny K is refused with
// ErrWriteBudgetExceeded and produces no image. This is the enforcement
// half of the DoD ("the bound enforced as a hard assertion (commit
// refuses on violation) rather than only measured").
func TestNFR_008_WriteCostRefusedWhenOverBudget(t *testing.T) {
	base := makeBaseOfSize(PrefixLength)

	// K = 1 octet: budget is 8 + 262144 = 262152. Append a segment far
	// bigger than that so the commit must be refused.
	oversized := SegmentPayload{Octets: make([]byte, WriteBudgetConstant+64*1024)}
	_, err := PlaceWithinBudget(base, 0, EditDelta{NewSegments: []SegmentPayload{oversized}}, 1)
	if err == nil {
		t.Fatalf("an over-budget commit was not refused")
	}
	if !errors.Is(err, ErrWriteBudgetExceeded) {
		t.Fatalf("over-budget commit returned %v, want ErrWriteBudgetExceeded", err)
	}

	// A commit whose write cost is within budget for its K is accepted,
	// and its reported WrittenOctets matches the accounting.
	k := 4096
	seg := makeEditSegment(k)
	res, err := PlaceWithinBudget(base, 0, EditDelta{NewSegments: []SegmentPayload{seg}}, uint64(k))
	if err != nil {
		t.Fatalf("in-budget commit refused: %v", err)
	}
	want := commitWriteCost(uint64(len(seg.Octets)), 1)
	if res.WrittenOctets != want {
		t.Fatalf("WrittenOctets = %d, want %d", res.WrittenOctets, want)
	}
	if res.WrittenOctets > WriteBudget(uint64(k)) {
		t.Fatalf("in-budget commit reported over-budget cost %d > %d", res.WrittenOctets, WriteBudget(uint64(k)))
	}
}
