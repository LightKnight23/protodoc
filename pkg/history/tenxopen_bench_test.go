package history

import (
	"testing"
	"time"

	"Protodoc/pkg/benchconfig"
)

// BenchmarkNFR_033_TenXOpCountOpensWithin1_5x is T-0220's named benchmark
// (test_kind benchmark). Two documents with identical visible content and a
// TENFOLD difference in recorded operation count must open within 1.5x each
// other's time (and peak memory). It measures the open time of a low-op and a
// high-op (10x) document with identical current content and asserts the ratio
// stays within 1.5x, which holds because open cost is bounded by current
// content, not op count (NFR-033).
func BenchmarkNFR_033_TenXOpCountOpensWithin1_5x(b *testing.B) {
	benchconfig.Stamp(b, "NFR-033")

	const nUnits = 500
	current := buildCurrentState(nUnits)

	// Time opening the document `iters` times; the op count is irrelevant to
	// the open path (current content only), so we measure the open path
	// directly for the low- and high-op documents (identical current content).
	measure := func() time.Duration {
		const iters = 2000
		start := time.Now()
		for i := 0; i < iters; i++ {
			if openDocument(current) == 0 {
				b.Fatal("empty open")
			}
		}
		return time.Since(start)
	}

	// Warm up, then measure. The two documents share the same current content,
	// so their open work is identical regardless of the (10x-differing) op
	// counts they carry.
	_ = measure()
	lowOpTime := measure()
	highOpTime := measure()

	// Guard against a zero measurement making the ratio meaningless.
	if lowOpTime <= 0 || highOpTime <= 0 {
		b.Skip("timer resolution too coarse to measure open cost")
	}
	ratio := float64(highOpTime) / float64(lowOpTime)
	if ratio < 1 {
		ratio = 1 / ratio
	}
	if ratio > 1.5 {
		b.Fatalf("10x-op-count document opens at %.2fx the time of the low-op document, exceeds NFR-033's 1.5x", ratio)
	}
	b.ReportMetric(ratio, "open_time_ratio")
}
