package container

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"Protodoc/pkg/benchconfig"
)

// coldCachePreviewFixtureSize is the total file length NFR-015/016 name:
// a 10,000-page, 1,073,741,824-octet (1 GiB) synthetic document. This
// milestone (M01) implements only the fixed 1,048,576-octet prefix, not
// ledger segment authoring, so the ledger region past the prefix is a
// genuinely sparse hole rather than fabricated page content: it exercises
// exactly what cold-cache preview production touches (Header, CommitRing,
// Frontmatter — never SegmentTable, never a ledger segment), and the
// fixture's on-disk SIZE still matches the stated 1 GiB document honestly.
const coldCachePreviewFixtureSize = 1073741824

// buildColdCachePreviewFixture writes a real, self-consistent
// Header+CommitRing+Frontmatter+SegmentTable prefix to a fresh temp file
// (ring slot 0 the sole eligible winner, its frontmatter-digest and
// segment-table-digest matching the encoded regions byte-for-byte), then
// sparse-extends the file to coldCachePreviewFixtureSize. It returns the
// file's path.
func buildColdCachePreviewFixture(tb testing.TB) string {
	tb.Helper()

	header := &Header{
		FormatMajor:        1,
		FormatMinor:        0,
		DocumentClass:      1,
		CapabilityWritten:  1,
		CapabilityRequired: 1,
		UnicodeVersionID:   1,
		ShapingProfileID:   1,
		PrefixLayoutID:     1,
		HistoryMode:        HistoryNone,
	}
	headerBytes := header.Encode(nil)

	fm := &Frontmatter{
		PreviewKind:   PreviewKindPLP1,
		PreviewRaster: bytes.Repeat([]byte{0xAB}, FMPreviewRasterMaxSize),
		Title:         "NFR-015/016 cold-cache benchmark synthetic document",
		PageCount:     10000,
		PageWidth:     8 * 914400,
		PageHeight:    11 * 914400,
		Language:      "en-US",
	}
	fm.PreviewDigest = ComputeFrontmatterPreviewDigest(fm)
	fmBytes, err := fm.Encode(nil)
	if err != nil {
		tb.Fatalf("Frontmatter.Encode: %v", err)
	}
	fmDigest := sha256.Sum256(fmBytes)

	// No ledger segments exist yet in this milestone: every slot unused.
	segTableBytes, err := EncodeSegmentTable(nil, nil)
	if err != nil {
		tb.Fatalf("EncodeSegmentTable: %v", err)
	}
	segTableDigest := sha256.Sum256(segTableBytes)

	ring := CommitRingRecord{
		Sequence:           1,
		LedgerLength:       coldCachePreviewFixtureSize,
		SegmentCount:       0,
		FrontmatterDigest:  fmDigest,
		SegmentTableDigest: segTableDigest,
	}
	ringRegion := make([]byte, CommitRingSize) // slots 1-6 stay genuinely all-zero (never-written)
	copy(ringRegion[0:RingSlotSize], ring.Encode(nil))

	path := filepath.Join(tb.TempDir(), "coldcache.pdl")
	f, err := os.Create(path)
	if err != nil {
		tb.Fatalf("os.Create: %v", err)
	}
	defer f.Close()

	for _, region := range [][]byte{headerBytes, ringRegion, fmBytes, segTableBytes} {
		if _, err := f.Write(region); err != nil {
			tb.Fatalf("write prefix region: %v", err)
		}
	}
	const wantPrefixLength = SegmentTableOffset + SegmentTableRegionSize // 1,048,576
	if off, _ := f.Seek(0, io.SeekCurrent); off != wantPrefixLength {
		tb.Fatalf("assembled prefix is %d octets, want %d", off, wantPrefixLength)
	}
	if err := f.Truncate(coldCachePreviewFixtureSize); err != nil {
		tb.Fatalf("Truncate to sparse %d octets: %v", coldCachePreviewFixtureSize, err)
	}
	return path
}

// BenchmarkNFR_015_ColdCachePreviewLatency is T-0015's named benchmark.
// Implements: NFR-015, NFR-016.
//
// Verification method per spec.md's Verify clauses for both requirements:
// the OBSERVED MAXIMUM over repeated runs, never a percentile. Run with a
// fixed iteration count rather than adaptive calibration, e.g.:
//
//	go test -run='^$' -bench=BenchmarkNFR_015_ColdCachePreviewLatency -benchtime=10x ./pkg/container/
//
// Each iteration reopens the fixture file and performs exactly the cold-
// cache preview production path: one sequential 262,144-octet read
// (Header+CommitRing+Frontmatter, never SegmentTable, never a ledger
// segment) plus bounded decode, asserting NFR-015's 300ms wall-clock
// ceiling per iteration. True OS page-cache eviction between iterations
// is not forced (no portable stdlib primitive does this across
// darwin/linux/windows — CP-010 forbids reaching for a non-stdlib one for
// it); what this benchmark verifies unconditionally is that the
// implemented decode path itself, on the reference build, stays inside
// both ceilings when driven from a freshly reopened file descriptor each
// time, which is the part actually implemented in this milestone.
//
// NFR-016's peak-resident-memory ceiling is checked via getrusage(2)
// ru_maxrss (peakRSSBytes, rusage_unix.go) on darwin/linux, comparing the
// process-lifetime high-water mark before and after the loop: since
// ru_maxrss never decreases, any growth across the loop is attributable
// to work the loop did. On a platform with no stdlib getrusage access
// (rusage_other.go), this half of the check is skipped and reported as
// unsupported rather than fabricated.
func BenchmarkNFR_015_ColdCachePreviewLatency(b *testing.B) {
	const (
		timeBoundNFR015 = 300 * time.Millisecond
		memBoundNFR016  = 67108864 // octets (NFR-016)
	)

	benchconfig.Stamp(b, "NFR-015, NFR-016")

	path := buildColdCachePreviewFixture(b)
	baselineRSS, rssSupported := peakRSSBytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		f, err := os.Open(path)
		if err != nil {
			b.Fatalf("open: %v", err)
		}
		prefix := make([]byte, SegmentTableOffset) // 262144: Header+CommitRing+Frontmatter only
		if _, err := io.ReadFull(f, prefix); err != nil {
			f.Close()
			b.Fatalf("read leading %d octets: %v", SegmentTableOffset, err)
		}
		fi, err := f.Stat()
		if err != nil {
			f.Close()
			b.Fatalf("stat: %v", err)
		}
		if err := f.Close(); err != nil {
			b.Fatalf("close: %v", err)
		}

		if _, err := DecodeHeader(prefix[:HeaderSize]); err != nil {
			b.Fatalf("DecodeHeader: %v", err)
		}
		winner, _, err := SelectWinner(prefix[HeaderSize:HeaderSize+CommitRingSize], uint64(fi.Size()))
		if err != nil {
			b.Fatalf("SelectWinner: %v", err)
		}
		fmRegion := prefix[FrontmatterOffset : FrontmatterOffset+FrontmatterRegionSize]
		if gotDigest := sha256.Sum256(fmRegion); gotDigest != winner.FrontmatterDigest {
			b.Fatalf("frontmatter-digest mismatch: got %x, want %x", gotDigest, winner.FrontmatterDigest)
		}
		fm, err := DecodeFrontmatter(fmRegion)
		if err != nil {
			b.Fatalf("DecodeFrontmatter: %v", err)
		}
		raster, status := FrontmatterPreview(fm)
		if status != PreviewOK || len(raster) == 0 {
			b.Fatalf("FrontmatterPreview: status=%v raster_len=%d, want PreviewOK with a non-empty raster", status, len(raster))
		}

		if elapsed := time.Since(start); elapsed > timeBoundNFR015 {
			b.Fatalf("NFR-015: iteration %d took %s, want <= %s", i, elapsed, timeBoundNFR015)
		}
	}
	b.StopTimer()

	if !rssSupported {
		b.Logf("NFR-016: peak resident memory measurement unsupported on %s/%s; not checked here", runtime.GOOS, runtime.GOARCH)
		return
	}
	finalRSS, _ := peakRSSBytes()
	var grew uint64
	if finalRSS > baselineRSS {
		grew = finalRSS - baselineRSS
	}
	b.Logf("NFR-016: peak RSS grew %d octets across %d iteration(s) (bound %d)", grew, b.N, uint64(memBoundNFR016))
	if grew > memBoundNFR016 {
		b.Fatalf("NFR-016: peak resident memory grew by %d octets, want <= %d", grew, uint64(memBoundNFR016))
	}
}
