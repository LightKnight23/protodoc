package extract_test

import (
	"io"
	"sync/atomic"
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/extract"
)

// countingReaderAt wraps a byte image and counts total octets read, for the
// NFR-012 octets-read budget benchmark.
type countingReaderAt struct {
	data  []byte
	octet int64
}

func (c *countingReaderAt) ReadAt(p []byte, off int64) (int, error) {
	n := copy(p, c.data[off:])
	atomic.AddInt64(&c.octet, int64(n))
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// BenchmarkNFR_012_ExtractionReadsWithin15PercentOfFileOctets is T-0098's
// named benchmark. It extracts the full text of the 1 GiB / 10,000-page
// corpus and asserts, as a hard failure, that the total octets read from
// storage is at most 15% of the file's octet length (NFR-012) -- extraction
// reads the fixed prefix plus the CONTENT segments only, skipping the
// resource octets that dominate the file.
func BenchmarkNFR_012_ExtractionReadsWithin15PercentOfFileOctets(b *testing.B) {
	benchconfig.Stamp(b, "NFR-012")

	img, err := extract.GenerateCorpus(extract.GiBCorpusProfile(1))
	if err != nil {
		b.Fatalf("GenerateCorpus: %v", err)
	}
	fileSize := int64(len(img))

	decode := func(seg extract.ContentSegment, octets []byte) (string, error) {
		// Touch the octets so the read is not elided, but produce no
		// allocation-heavy result.
		_ = octets[0]
		return "", nil
	}

	b.ResetTimer()
	var lastRatio float64
	for i := 0; i < b.N; i++ {
		r := &countingReaderAt{data: img}
		ex, err := extract.NewExtractor(r, decode)
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
		lastRatio = float64(r.octet) / float64(fileSize)
		if lastRatio > 0.15 {
			b.Fatalf("extraction read %d of %d octets (%.4f), exceeds NFR-012 15%% budget", r.octet, fileSize, lastRatio)
		}
	}
	b.StopTimer()
	b.ReportMetric(lastRatio*100, "pct_octets_read")
}
