// Bounded unit lookup (T-0042, FR-055): locating any single addressable
// unit costs at most 3 sequential dependent reads, and the octet volume
// read beyond those 3 reads is bounded and documented. The read chain is
// exactly the one container.abnf S3's index-route comment specifies:
//
//	read 1: the fixed 1,048,576-octet prefix (already resident on open);
//	        the winning CommitRingRecord's index-route bucket gives the
//	        UnitIndexLeaf segment's ordinal, and that ordinal's
//	        SegmentTableSlot (also in the prefix) gives the leaf's
//	        offset/length -- no extra read, both live in the prefix.
//	read 2: the UnitIndexLeaf segment; binary search gives the unit's
//	        absolute offset and length.
//	read 3: the content unit itself.
//
// Three dependent reads, no more, for a document up to 1 GiB.
package ledger

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// FR055MaxDependentReads is the FR-055 dependent-read ceiling: at most 3
// sequential dependent reads to locate and read any single addressable
// unit.
const FR055MaxDependentReads = 3

// countingReaderAt wraps a byte image and counts each ReadAt call as one
// dependent read, recording the total octets each read returned. It is the
// instrument the FR-055 read-bound test asserts against.
type countingReaderAt struct {
	data       []byte
	reads      int
	octetsRead int
}

func (r *countingReaderAt) readAt(off, length uint64) ([]byte, error) {
	end := off + length
	if end < off || end > uint64(len(r.data)) {
		return nil, fmt.Errorf("ledger: read [%d,%d) out of bounds for %d-octet image", off, end, len(r.data))
	}
	r.reads++
	r.octetsRead += int(length)
	return r.data[off:end], nil
}

// LocatedUnit is the result of a bounded lookup: where the unit lives and
// the cost the lookup incurred, so a caller (and the FR-055 test) can
// verify both clauses of the requirement.
type LocatedUnit struct {
	Offset uint64
	Length uint32
	Kind   uint16
	Digest pdlfmt.Digest256

	// DependentReads is the number of dependent reads the lookup performed
	// (must be <= FR055MaxDependentReads).
	DependentReads int
	// PrefixOctets is the octets read by read 1 (the prefix). Reads beyond
	// the prefix are the leaf segment and the unit; see OctetsBeyondPrefix.
	PrefixOctets int
	// OctetsBeyondPrefix is the octet volume read after read 1 (the leaf
	// segment plus the unit): the FR-055 "octets read beyond" budget figure.
	OctetsBeyondPrefix int
}

var (
	// ErrUnitBucketUnregistered is returned when the index-route bucket for
	// the sought unit is empty (0xFFFF): no unit is registered there.
	ErrUnitBucketUnregistered = errors.New("ledger: index-route bucket holds no unit-index leaf")

	// ErrUnitNotFound is returned when the unit-id is not present in the
	// leaf its bucket routes to.
	ErrUnitNotFound = errors.New("ledger: unit-id not found in its index leaf")
)

// LeafPlacement tells the lookup where a bucket's UnitIndexLeaf segment
// lives (its absolute offset and length in the image), the information a
// real reader gets from the winning CommitRingRecord's index-route bucket
// plus that ordinal's SegmentTableSlot -- both resident in the prefix, so
// obtaining it costs no read beyond read 1.
type LeafPlacement struct {
	Offset uint64
	Length uint64
}

// LocateUnit performs the bounded 3-dependent-read lookup for id over the
// document image. indexRoute is the winning record's 16-bucket route;
// leafPlacements maps a leaf-segment ordinal to its offset/length (derived
// from the prefix's SegmentTable, so free of an extra read); decodeLeaf
// turns a leaf segment's octets into its UnitIndexLeaf. The image's leading
// PrefixLength octets model read 1 (the resident prefix). It returns the
// located unit and the measured read cost, or a structured error.
//
// The bound is enforced defensively: if the chain would exceed
// FR055MaxDependentReads the lookup fails rather than silently doing more
// work, so a regression that adds a read is caught, not absorbed.
func LocateUnit(
	image []byte,
	id pdlfmt.UnitID,
	indexRoute [IndexRouteBuckets]uint16,
	leafPlacements map[uint16]LeafPlacement,
	decodeLeaf func(segment []byte) (UnitIndexLeaf, error),
) (LocatedUnit, error) {
	if len(image) < PrefixLength {
		return LocatedUnit{}, fmt.Errorf("%w: image %d octets", ErrPriorTooShort, len(image))
	}
	r := &countingReaderAt{data: image}

	// Read 1: the resident prefix. A real reader already holds this on
	// open; we model it as one read of the leading PrefixLength octets so
	// the read counter reflects the container.abnf S3 accounting.
	if _, err := r.readAt(0, PrefixLength); err != nil {
		return LocatedUnit{}, err
	}
	prefixOctets := r.octetsRead

	bucket := bucketOf(id)
	leafOrdinal := indexRoute[bucket]
	if leafOrdinal == 0xFFFF {
		return LocatedUnit{}, fmt.Errorf("%w: bucket %d", ErrUnitBucketUnregistered, bucket)
	}
	placement, ok := leafPlacements[leafOrdinal]
	if !ok {
		return LocatedUnit{}, fmt.Errorf("ledger: index-route bucket %d names leaf ordinal %d with no SegmentTable placement", bucket, leafOrdinal)
	}

	// Read 2: the UnitIndexLeaf segment.
	leafOctets, err := r.readAt(placement.Offset, placement.Length)
	if err != nil {
		return LocatedUnit{}, fmt.Errorf("reading unit-index leaf: %w", err)
	}
	leaf, err := decodeLeaf(leafOctets)
	if err != nil {
		return LocatedUnit{}, fmt.Errorf("decoding unit-index leaf: %w", err)
	}
	entry, found := leaf.Lookup(id)
	if !found {
		return LocatedUnit{}, fmt.Errorf("%w: %x", ErrUnitNotFound, id)
	}

	// Read 3: the content unit itself.
	if _, err := r.readAt(entry.Offset, uint64(entry.Length)); err != nil {
		return LocatedUnit{}, fmt.Errorf("reading content unit: %w", err)
	}

	if r.reads > FR055MaxDependentReads {
		return LocatedUnit{}, fmt.Errorf("ledger: lookup used %d dependent reads, exceeds FR-055 ceiling %d", r.reads, FR055MaxDependentReads)
	}

	return LocatedUnit{
		Offset:             entry.Offset,
		Length:             entry.Length,
		Kind:               entry.Kind,
		Digest:             entry.Digest,
		DependentReads:     r.reads,
		PrefixOctets:       prefixOctets,
		OctetsBeyondPrefix: r.octetsRead - prefixOctets,
	}, nil
}
