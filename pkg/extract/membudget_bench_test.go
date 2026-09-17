package extract_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/extract"
)

// writeCorpusFile writes a generated corpus to a temp file and returns an
// *os.File opened for reading (an io.ReaderAt). Serving the corpus from a
// file keeps its (large) octets OFF the Go heap, so HeapInuse during
// extraction reflects only the extractor's own working set, not the corpus
// buffer.
func writeCorpusFile(b *testing.B, p extract.CorpusProfile) *os.File {
	b.Helper()
	img, err := extract.GenerateCorpus(p)
	if err != nil {
		b.Fatalf("GenerateCorpus: %v", err)
	}
	path := filepath.Join(b.TempDir(), "corpus.pdl")
	if err := os.WriteFile(path, img, 0o644); err != nil {
		b.Fatalf("WriteFile: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	return f
}

// peakExtractionHeap runs a full extraction reading from f and returns the
// peak HeapInuse observed. Because the corpus is file-backed (off-heap),
// this is a clean proxy for extraction's resident working set.
func peakExtractionHeap(b *testing.B, f *os.File) uint64 {
	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		_ = octets[0]
		return "", nil
	}
	runtime.GC()
	ex, err := extract.NewExtractor(f, decode)
	if err != nil {
		b.Fatalf("NewExtractor: %v", err)
	}
	var peak, units uint64
	var ms runtime.MemStats
	for {
		if _, ok := ex.Next(); !ok {
			break
		}
		units++
		if units%256 == 0 {
			runtime.ReadMemStats(&ms)
			if ms.HeapInuse > peak {
				peak = ms.HeapInuse
			}
		}
	}
	runtime.ReadMemStats(&ms)
	if ms.HeapInuse > peak {
		peak = ms.HeapInuse
	}
	if ex.Err() != nil {
		b.Fatalf("extraction error: %v", ex.Err())
	}
	return peak
}

// BenchmarkNFR_014_ExtractionPeakMemoryStaysFlatAndBounded is T-0100's named
// benchmark. It extracts a base corpus and a 4x-scaled corpus (both
// file-backed, so the corpus octets are off-heap) and asserts peak heap-in-
// use stays under 33,554,432 bytes (32 MiB) on both, and that the 4x corpus
// does not materially raise the peak (flatness): the streaming extractor's
// working set is bounded by one segment plus the fixed prefix, not the
// document size (NFR-014).
func BenchmarkNFR_014_ExtractionPeakMemoryStaysFlatAndBounded(b *testing.B) {
	benchconfig.Stamp(b, "NFR-014")

	const budget = uint64(33554432) // 32 MiB

	base := extract.GiBCorpusProfile(1)
	base.Pages = 2500
	scaled := base
	scaled.Pages = base.Pages * 4

	baseF := writeCorpusFile(b, base)
	defer baseF.Close()
	scaledF := writeCorpusFile(b, scaled)
	defer scaledF.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		basePeak := peakExtractionHeap(b, baseF)
		scaledPeak := peakExtractionHeap(b, scaledF)

		if basePeak > budget {
			b.Fatalf("base corpus extraction peak heap %d exceeds NFR-014 budget %d", basePeak, budget)
		}
		if scaledPeak > budget {
			b.Fatalf("4x corpus extraction peak heap %d exceeds NFR-014 budget %d", scaledPeak, budget)
		}
		// Flatness: the 4x corpus must not materially raise the peak.
		if scaledPeak > basePeak*2 && scaledPeak-basePeak > 4*1024*1024 {
			b.Fatalf("peak memory not flat: base %d, 4x %d (grew with document size)", basePeak, scaledPeak)
		}
		b.ReportMetric(float64(scaledPeak)/(1024*1024), "peak_MiB_4x")
	}
}
