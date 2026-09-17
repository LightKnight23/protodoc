package validate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
)

// writeValidatorInput writes a size-octet input to a temp file (off-heap, so
// the validator's own arena is what the peak-heap measurement reflects). It
// builds a valid PDL image sized up to `size` via the extract corpus
// generator when size >= the prefix, else a short (malformed) input.
func writeValidatorInput(b *testing.B, size int) (*os.File, int64) {
	b.Helper()
	prefixLen := container.SegmentTableOffset + container.SegmentTableRegionSize
	var img []byte
	if size < prefixLen {
		img = make([]byte, size) // short/malformed
	} else {
		// A valid-shaped corpus close to the requested size.
		p := extract.GiBCorpusProfile(1)
		p.Pages = size / (200 * 1024)
		if p.Pages < 1 {
			p.Pages = 1
		}
		g, err := extract.GenerateCorpus(p)
		if err != nil {
			b.Fatalf("GenerateCorpus: %v", err)
		}
		img = g
	}
	path := filepath.Join(b.TempDir(), "vinput.pdl")
	if err := os.WriteFile(path, img, 0o644); err != nil {
		b.Fatalf("WriteFile: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	return f, int64(len(img))
}

// BenchmarkNFR_030_ValidatorPeakMemoryWithinAdoptedBound is T-0121's named
// benchmark. It measures peak HeapInuse of ValidateBytes across 1 KB / 1 MB
// / ~64 MB inputs and a malformed (truncated) variant, asserting each stays
// within the adopted floor+4x NFR-030 bound. The divergence from NFR-030's
// literal zero-floor text is documented in arena.go, citing plan.md Section
// 9 Conflict 3.
func BenchmarkNFR_030_ValidatorPeakMemoryWithinAdoptedBound(b *testing.B) {
	benchconfig.Stamp(b, "NFR-030")

	sizes := []int{1024, 1024 * 1024, 64 * 1024 * 1024}
	for _, sz := range sizes {
		f, actual := writeValidatorInput(b, sz)
		bound := AdoptedMemoryBound(int(actual))

		// Measure the bytes ONE ValidateBytes call allocates, via the
		// TotalAlloc delta across a single call. Unlike live HeapInuse,
		// TotalAlloc is deterministic and not subject to GC-retention timing:
		// it is exactly the working memory the bounded arena touches, which
		// must stay within the adopted floor+4x bound regardless of file
		// size (the file data stays on the ReaderAt, off the validator heap).
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		_ = ValidateBytes(f, actual)
		runtime.ReadMemStats(&after)
		allocated := after.TotalAlloc - before.TotalAlloc

		if allocated > bound {
			b.Fatalf("size %d: ValidateBytes allocated %d octets, exceeds adopted bound %d (floor %d)", actual, allocated, bound, uint64(ArenaFloor))
		}
		f.Close()
		b.ReportMetric(float64(allocated)/(1024*1024), "alloc_MiB")
	}
}
