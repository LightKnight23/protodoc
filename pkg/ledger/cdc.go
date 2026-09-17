// Content-defined chunking (CDC) for edit-delta novel-chunk boundary
// determination (T-0037, NFR-009, plan.md Section 5 "8KiB average CDC").
// The boundary set a document's content produces is a pure function of the
// content octets alone: it depends on no edit offset, no prior edit count,
// and no document-position alignment, so the same final content reached by
// two different edit histories yields an identical boundary set. This is
// what keeps novel-chunk output proportional to the edit rather than to the
// document (NFR-009), and it is a deterministic computation with no ambient
// input (CP-004).
//
// The algorithm is a Gear-hash rolling CDC: a 64-bit rolling hash is
// updated one octet at a time as h = (h << 1) + gearTable[b]; a chunk
// boundary is cut at the first position at or after MinChunkSize where the
// top MaskBits of the hash are zero, or forced at MaxChunkSize. Gear hashing
// is chosen over Rabin fingerprinting for being branch-light and using only
// a fixed 256-entry table, no polynomial-field arithmetic; the table is a
// compile-time constant derived from a fixed seed (below), never an ambient
// or content-derived value.
package ledger

// CDC chunk-size parameters. Average chunk size is 2^MaskBits = 8192 octets
// (plan.md "8KiB average CDC"): a boundary is cut when the top MaskBits of
// the rolling hash are zero, which happens on average once per 2^MaskBits
// octets over uniformly distributed content. MinChunkSize and MaxChunkSize
// bound the realised chunk length so a pathological run of non-matching
// content cannot produce an unbounded chunk and a run of matches cannot
// produce a flood of tiny ones.
const (
	// MaskBits is the number of high hash bits required to be zero at a
	// cut point; 2^MaskBits = 8192 is the target average chunk size.
	MaskBits = 13

	// MinChunkSize is the smallest chunk the cutter will emit before it
	// begins testing the boundary condition (2 KiB): below this the
	// rolling hash has too little content mixed in to place a
	// content-stable boundary.
	MinChunkSize = 2 * 1024

	// MaxChunkSize is the largest chunk the cutter will emit; a boundary
	// is forced here even if the hash condition has not been met (64 KiB),
	// bounding worst-case chunk length.
	MaxChunkSize = 64 * 1024
)

// cutMask has its top MaskBits bits set; a cut point is where
// (hash & cutMask) == 0.
const cutMask = ((uint64(1) << MaskBits) - 1) << (64 - MaskBits)

// gearTable maps each possible octet value to a 64-bit mixing constant. It
// is a fixed, content-independent table produced deterministically at
// package init from a constant seed via a splitmix64 generator, so it is
// identical in every build and every independent implementation that
// follows this same construction (CP-004: a compile-time-fixed constant,
// not an ambient value). It is never regenerated at runtime and never
// depends on any document's content.
var gearTable [256]uint64

func init() {
	// splitmix64 with a fixed seed: a well-specified, portable generator
	// whose output is identical on every platform, so the gear table is a
	// reproducible constant rather than a platform- or run-dependent value.
	const seed = 0x9E3779B97F4A7C15
	x := uint64(seed)
	for i := range gearTable {
		x += 0x9E3779B97F4A7C15
		z := x
		z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
		z = (z ^ (z >> 27)) * 0x94D049BB133111EB
		z = z ^ (z >> 31)
		gearTable[i] = z
	}
}

// Chunk describes one content-defined chunk: its half-open octet range
// [Offset, Offset+Length) within the input the cutter was run over.
type Chunk struct {
	Offset uint64
	Length uint64
}

// ChunkBoundaries runs the Gear-hash CDC over data and returns the chunks
// it cuts, in order, covering data exactly ([0,len(data))) with no gap or
// overlap. The boundary set is a pure function of data's octets: it does
// not depend on where in a larger document data sits, on any prior edit
// count, or on any offset alignment. Two inputs with identical octets
// produce identical boundary sets.
//
// An empty input produces no chunks. A cut is placed at the first position
// at or after MinChunkSize where the rolling hash's top MaskBits are zero,
// and is forced at MaxChunkSize; the final trailing octets always form a
// last chunk even if no cut condition was met.
func ChunkBoundaries(data []byte) []Chunk {
	if len(data) == 0 {
		return nil
	}

	var chunks []Chunk
	var hash uint64
	chunkStart := 0
	chunkLen := 0

	for i := 0; i < len(data); i++ {
		hash = (hash << 1) + gearTable[data[i]]
		chunkLen++

		atMin := chunkLen >= MinChunkSize
		atMax := chunkLen >= MaxChunkSize
		if (atMin && (hash&cutMask) == 0) || atMax {
			chunks = append(chunks, Chunk{Offset: uint64(chunkStart), Length: uint64(chunkLen)})
			chunkStart = i + 1
			chunkLen = 0
			hash = 0
		}
	}

	// Any trailing octets that did not reach a cut form the final chunk.
	if chunkLen > 0 {
		chunks = append(chunks, Chunk{Offset: uint64(chunkStart), Length: uint64(chunkLen)})
	}

	return chunks
}

// BoundaryOffsets returns just the cut offsets (each chunk's start offset)
// of ChunkBoundaries(data), which is the boundary SET the NFR-009 property
// test compares: two documents with identical content must yield identical
// boundary offsets regardless of the edit history that produced them.
func BoundaryOffsets(data []byte) []uint64 {
	chunks := ChunkBoundaries(data)
	offsets := make([]uint64, len(chunks))
	for i, c := range chunks {
		offsets[i] = c.Offset
	}
	return offsets
}

// NovelChunkOctets returns the total octet length of the chunks the CDC
// cuts over after that do not appear (by exact content) among the chunks
// it cuts over before. This is the "novel chunk total" NFR-009 bounds: the
// volume a content-defined sync layer must transfer to move a peer holding
// before to after. Because CDC boundaries are content-defined, a small
// edit perturbs only the chunk(s) overlapping it plus at most the shifted
// boundary chunks adjacent to it, so novel octets stay proportional to the
// edit rather than to the document size.
//
// Chunk identity is by exact content bytes: two chunks are "the same"
// (non-novel) when their octets are identical, matched with multiplicity
// so a chunk value present k times in before covers at most k occurrences
// in after. This models a chunk store keyed by content digest.
func NovelChunkOctets(before, after []byte) uint64 {
	haveCounts := make(map[string]int)
	for _, c := range ChunkBoundaries(before) {
		haveCounts[string(before[c.Offset:c.Offset+c.Length])]++
	}

	var novel uint64
	for _, c := range ChunkBoundaries(after) {
		key := string(after[c.Offset : c.Offset+c.Length])
		if haveCounts[key] > 0 {
			haveCounts[key]--
			continue
		}
		novel += c.Length
	}
	return novel
}
