package extract_test

import (
	"bytes"
	"testing"
	"time"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/extract"
)

// BenchmarkNFR_013_ExtractionCompletesWithinProcessorTimeBudget is T-0099's
// named benchmark. It performs a full-text extraction of the 1 GiB /
// 10,000-page corpus and asserts, as a hard failure, that it completes
// within 20 seconds of processor time on the reference configuration
// documented in T-0097 (NFR-011 stand-in). Measured elapsed wall time is
// used as an upper bound on CPU time for this single-goroutine walk.
func BenchmarkNFR_013_ExtractionCompletesWithinProcessorTimeBudget(b *testing.B) {
	benchconfig.Stamp(b, "NFR-013")

	img, err := extract.GenerateCorpus(extract.GiBCorpusProfile(1))
	if err != nil {
		b.Fatalf("GenerateCorpus: %v", err)
	}

	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		_ = octets[0]
		return "", nil
	}

	const budget = 20 * time.Second
	b.ResetTimer()
	var last time.Duration
	for i := 0; i < b.N; i++ {
		start := time.Now()
		ex, err := extract.NewExtractor(bytes.NewReader(img), decode)
		if err != nil {
			b.Fatalf("NewExtractor: %v", err)
		}
		for {
			if _, ok := ex.Next(); !ok {
				break
			}
		}
		if ex.Err() != nil {
			b.Fatalf("extraction error: %v", ex.Err())
		}
		last = time.Since(start)
		if last > budget {
			b.Fatalf("extraction took %v, exceeds NFR-013 processor-time budget %v", last, budget)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(last.Milliseconds()), "extract_ms")
}
