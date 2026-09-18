package render

import (
	"testing"

	"Protodoc/pkg/benchconfig"
)

// BenchmarkNFR_018_RenderPageOctetReadBudget is T-0263's named benchmark
// (NFR-018). It measures the octets read to render a single page and asserts
// they stay within the base per-page read budget (8 MiB) plus the page's own
// declared resource octets -- the read cost is proportional to the page and its
// resources, not the whole document.
func BenchmarkNFR_018_RenderPageOctetReadBudget(b *testing.B) {
	benchconfig.Stamp(b, "NFR-018")

	page := buildRenderPage(1024, 768, 64)

	// The page's declared resource octets (the sum of its object payloads).
	var resourceOctets int
	for _, obj := range page.Objects {
		resourceOctets += len(obj.Payload)
	}
	budget := PageOctetReadBudgetBytes + resourceOctets

	b.ResetTimer()
	var lastRead int
	for i := 0; i < b.N; i++ {
		_, read, err := RenderPage(page)
		if err != nil {
			b.Fatalf("RenderPage: %v", err)
		}
		lastRead = read
	}
	b.StopTimer()

	if lastRead > budget {
		b.Fatalf("page render read %d octets, exceeds NFR-018 budget %d (8 MiB + %d resource octets)",
			lastRead, budget, resourceOctets)
	}
	// The read is exactly the page's resource octets -- nothing document-wide.
	if lastRead != resourceOctets {
		b.Fatalf("page render read %d octets, want exactly the page's resource octets %d", lastRead, resourceOctets)
	}
	b.ReportMetric(float64(lastRead)/(1024*1024), "read_MiB")
}
