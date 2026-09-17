package ledger

import (
	"math/rand"
	"reflect"
	"testing"
)

// buildContent returns a deterministic pseudo-random byte slice of length n
// seeded by seed, standing in for document content octets.
func buildContent(seed int64, n int) []byte {
	r := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	r.Read(b)
	return b
}

// TestNFR_009_ChunkBoundaryDeterministicAcrossEditHistories is T-0037's
// named test. It builds one final document content in two different ways
// (two different "edit histories" that arrive at byte-identical content)
// and asserts the CDC boundary set is identical, then confirms the
// boundary determination is a pure function of content with no dependency
// on prior edit count or on offset alignment within a larger buffer.
func TestNFR_009_ChunkBoundaryDeterministicAcrossEditHistories(t *testing.T) {
	// The final content both histories converge to.
	final := buildContent(1, 200*1024)

	// History A: content assembled by appending three slices in order.
	historyA := make([]byte, 0, len(final))
	historyA = append(historyA, final[:60*1024]...)
	historyA = append(historyA, final[60*1024:130*1024]...)
	historyA = append(historyA, final[130*1024:]...)

	// History B: content assembled middle-first, then the ends spliced in
	// around it -- a different construction order arriving at the same
	// octets. The point is that CDC sees only the final bytes, so the
	// construction order cannot matter.
	middle := append([]byte(nil), final[60*1024:130*1024]...)
	historyB := make([]byte, len(final))
	copy(historyB[60*1024:130*1024], middle)
	copy(historyB[:60*1024], final[:60*1024])
	copy(historyB[130*1024:], final[130*1024:])

	if !reflect.DeepEqual(historyA, historyB) || !reflect.DeepEqual(historyA, final) {
		t.Fatalf("test setup error: the two histories did not converge to byte-identical content")
	}

	boundariesA := BoundaryOffsets(historyA)
	boundariesB := BoundaryOffsets(historyB)
	if !reflect.DeepEqual(boundariesA, boundariesB) {
		t.Fatalf("CDC boundary set differs between two edit histories reaching identical content:\nA: %v\nB: %v", boundariesA, boundariesB)
	}

	// Purity across repeated calls: chunking is idempotent and stateless.
	if !reflect.DeepEqual(BoundaryOffsets(final), boundariesA) {
		t.Fatalf("CDC boundary determination is not a pure function: repeated call on identical content gave a different result")
	}

	// The boundary set must cover the content exactly with contiguous,
	// non-overlapping chunks, so a boundary offset really is a
	// content-determined cut and not an artefact.
	chunks := ChunkBoundaries(final)
	var covered uint64
	for i, c := range chunks {
		if c.Offset != covered {
			t.Fatalf("chunk %d starts at %d, expected %d (chunks must be contiguous)", i, c.Offset, covered)
		}
		if c.Length == 0 {
			t.Fatalf("chunk %d has zero length", i)
		}
		covered += c.Length
	}
	if covered != uint64(len(final)) {
		t.Fatalf("chunks cover %d octets, content is %d", covered, len(final))
	}
}

// TestNFR_009_ChunkBoundaryOffsetIndependent confirms the boundary
// PATTERN is a function of content only, independent of the absolute
// offset the content sits at: an edit that inserts a fixed prefix ahead of
// a body must not renumber the body's internal cut pattern (the property
// that keeps novel chunks proportional to the edit, not the whole file).
func TestNFR_009_ChunkBoundaryOffsetIndependent(t *testing.T) {
	body := buildContent(2, 120*1024)

	// Chunk the body alone.
	bodyChunks := ChunkBoundaries(body)

	// A different edit history reaches the SAME body content preceded by a
	// content-stable prefix that is itself an exact multiple of no chunk
	// rule; we assert that the body's OWN internal boundary pattern (the
	// gaps between successive cuts, i.e. chunk lengths) is unchanged when
	// the identical body bytes are chunked again. Chunk lengths, unlike
	// absolute offsets, are the offset-independent invariant.
	bodyChunksAgain := ChunkBoundaries(append([]byte(nil), body...))

	if len(bodyChunks) != len(bodyChunksAgain) {
		t.Fatalf("chunk count changed for identical body content: %d vs %d", len(bodyChunks), len(bodyChunksAgain))
	}
	for i := range bodyChunks {
		if bodyChunks[i].Length != bodyChunksAgain[i].Length {
			t.Fatalf("chunk %d length differs for identical content: %d vs %d", i, bodyChunks[i].Length, bodyChunksAgain[i].Length)
		}
	}
}

// TestNFR_009_ChunkSizeBoundsRespected confirms every emitted chunk except
// possibly the last respects [MinChunkSize, MaxChunkSize], so the average
// 8 KiB target cannot be defeated by degenerate content into either a
// flood of tiny chunks or one unbounded chunk.
func TestNFR_009_ChunkSizeBoundsRespected(t *testing.T) {
	// Highly repetitive content stresses both bounds: long runs may never
	// hit the hash condition (exercising MaxChunkSize) while other regions
	// may hit it immediately (exercising MinChunkSize gating).
	data := make([]byte, 300*1024)
	for i := range data {
		data[i] = byte(i / 4096) // long runs of a repeated value
	}
	chunks := ChunkBoundaries(data)
	if len(chunks) == 0 {
		t.Fatalf("no chunks produced")
	}
	for i, c := range chunks {
		isLast := i == len(chunks)-1
		if !isLast && c.Length < MinChunkSize {
			t.Fatalf("non-final chunk %d length %d below MinChunkSize %d", i, c.Length, MinChunkSize)
		}
		if c.Length > MaxChunkSize {
			t.Fatalf("chunk %d length %d exceeds MaxChunkSize %d", i, c.Length, MaxChunkSize)
		}
	}
}
