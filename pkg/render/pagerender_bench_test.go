package render

import (
	"runtime"
	"testing"

	"Protodoc/pkg/benchconfig"
)

// buildRenderPage builds a page with `n` embedded objects, each a small
// materialised PLP-1 raster placed on the page.
func buildRenderPage(w, h, n int) PageToRender {
	page := PageToRender{Width: w, Height: h}
	payload := buildValidVector("gradient_8x8", 8, 8, 3)
	for i := 0; i < n; i++ {
		page.Objects = append(page.Objects, EmbeddedObject{
			NativeKind:   uint32(i),
			DeclExtent:   Extent{X: (i * 8) % w, Y: 0, Width: 8, Height: 8},
			Materialised: MaterialisedPLP1,
			Payload:      payload,
		})
	}
	return page
}

// BenchmarkNFR_017_OpenAndRenderPagePeakMemory is T-0262's named benchmark
// (NFR-017). It measures the working memory to open and render a single page
// and asserts it stays within the 200 MiB page-render budget, independent of
// how large the surrounding document would be.
func BenchmarkNFR_017_OpenAndRenderPagePeakMemory(b *testing.B) {
	benchconfig.Stamp(b, "NFR-017")

	// A large-ish page with many objects -- the working set is one page raster
	// plus one transient object raster at a time, not the whole document.
	page := buildRenderPage(1024, 768, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := RenderPage(page); err != nil {
			b.Fatalf("RenderPage: %v", err)
		}
	}
	b.StopTimer()

	// Measure the bytes one page render allocates via the TotalAlloc delta.
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, _, err := RenderPage(page); err != nil {
		b.Fatalf("RenderPage: %v", err)
	}
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc

	if allocated > PageRenderBudgetBytes {
		b.Fatalf("page render allocated %d octets, exceeds NFR-017 budget %d (200 MiB)", allocated, PageRenderBudgetBytes)
	}
	b.ReportMetric(float64(allocated)/(1024*1024), "peak_MiB")
}
