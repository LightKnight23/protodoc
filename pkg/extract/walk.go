// Package extract is Protodoc's zero-trust-material extraction view: it
// walks a document's CONTENT and emits ordered text, structural locators and
// per-unit language, reading only what it needs and never constructing a
// font, shaping, layout, graphics, crypto or merge facility (TR-011,
// FR-041). It has, by construction, NO import edge to the integrity,
// render or merge packages -- a module-boundary property a CI check enforces
// (TestFR_041_ExtractHasNoHeavyImportEdges) -- so the extractor stays the
// small, cheap, dependency-isolated reader the format promises. (The
// module-boundary property is also a requirement in its own right, covered
// by its own later extraction task.)
//
// This file implements the base streaming primitive Walk (T-0087, FR-041):
// it walks only CONTENT-typed segments in storage (ascending storage
// ordinal) order via the M01 SegmentTable inventory, skipping RESOURCE,
// HISTORY and ATTEST segments entirely -- their payload bytes are never read.
package extract

import (
	"errors"
	"fmt"
	"io"

	"Protodoc/pkg/container"
)

// PrefixLength is the fixed prefix size (container.abnf S1), the leading
// region Walk reads to obtain the SegmentTable inventory.
const PrefixLength = container.SegmentTableOffset + container.SegmentTableRegionSize

// ContentSegment identifies one CONTENT-typed segment visited by Walk: its
// storage ordinal and the octet range it occupies in the document.
type ContentSegment struct {
	Ordinal uint64
	Offset  uint64
	Length  uint64
}

// ErrPrefixTooShort is returned when the source does not contain a full
// fixed prefix.
var ErrPrefixTooShort = errors.New("extract: source shorter than the 1048576-octet fixed prefix")

// Walk reads the document's SegmentTable from r and invokes visit once per
// CONTENT-typed segment, in ascending storage ordinal order. RESOURCE,
// HISTORY and ATTEST segments are skipped entirely: their payload bytes are
// never read. visit receives the CONTENT segment's placement and a reader
// positioned to read exactly that segment's octets; if visit returns an
// error, the walk stops and returns it (supporting abandonment). Walk itself
// reads only the fixed prefix plus, lazily via the supplied segment reader,
// the CONTENT segments visit chooses to read.
func Walk(r io.ReaderAt, visit func(seg ContentSegment, segReader io.Reader) error) error {
	prefix := make([]byte, PrefixLength)
	n, err := r.ReadAt(prefix, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("extract: reading fixed prefix: %w", err)
	}
	if n < PrefixLength {
		return ErrPrefixTooShort
	}

	// The SegmentTable occupies [SegmentTableOffset, PrefixLength).
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return fmt.Errorf("extract: decoding segment table: %w", err)
	}

	for ordinal := 0; ordinal < len(table); ordinal++ {
		slot := table[ordinal]
		if slot.SegmentType == container.SegmentTypeUnused {
			continue
		}
		if slot.SegmentType != container.SegmentTypeContent {
			continue // RESOURCE / HISTORY / ATTEST: never read the payload
		}
		seg := ContentSegment{Ordinal: uint64(ordinal), Offset: slot.Offset, Length: slot.Length}
		// A section reader scoped exactly to this CONTENT segment's octets,
		// so visit cannot read outside the segment it was handed.
		sr := io.NewSectionReader(r, int64(slot.Offset), int64(slot.Length))
		if err := visit(seg, sr); err != nil {
			return err
		}
	}
	return nil
}

// ContentSegments is a convenience over Walk returning the placements of
// every CONTENT segment in storage order without reading any payload (it
// passes a visit that ignores the reader). Useful for callers that only need
// the inventory.
func ContentSegments(r io.ReaderAt) ([]ContentSegment, error) {
	var out []ContentSegment
	err := Walk(r, func(seg ContentSegment, _ io.Reader) error {
		out = append(out, seg)
		return nil
	})
	return out, err
}
