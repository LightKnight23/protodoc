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
// WalkInReadingOrder (FR-036/FR-101, T-0276) is the reading-order counterpart
// that emits CONTENT in authored ROOT_SEQUENCE order rather than storage order.
package extract

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
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

// ContentSegmentsInReadingOrder returns the CONTENT segment placements in the
// document's authored ROOT_SEQUENCE reading order (FR-036, FR-101, T-0276)
// rather than raw storage order, closing the gap where a structural edit that
// inserted a unit mid-document would otherwise emit it at the end (its storage
// ordinal) instead of its authored position.
//
// unitOf maps a CONTENT segment to the unit-id it carries; order is the
// rs-order list of the ROOT_SEQUENCE record (document.abnf S5.1). Segments are
// sorted by their unit-id's position in order; a segment whose unit-id is not
// in order sorts after all sequenced segments, keyed by ascending storage
// ordinal, so the walk stays total and deterministic on an incomplete
// sequence (completeness is enforced separately by semantics PD-A11Y-005).
// It reads only the fixed prefix — no CONTENT payload — to build the
// inventory.
func ContentSegmentsInReadingOrder(r io.ReaderAt, unitOf func(seg ContentSegment) pdlfmt.UnitID, order []pdlfmt.UnitID) ([]ContentSegment, error) {
	segs, err := ContentSegments(r)
	if err != nil {
		return nil, err
	}
	pos := make(map[pdlfmt.UnitID]int, len(order))
	for i, u := range order {
		if _, dup := pos[u]; !dup {
			pos[u] = i
		}
	}
	sort.SliceStable(segs, func(i, j int) bool {
		pi, iok := pos[unitOf(segs[i])]
		pj, jok := pos[unitOf(segs[j])]
		switch {
		case iok && jok:
			return pi < pj
		case iok != jok:
			return iok
		default:
			return segs[i].Ordinal < segs[j].Ordinal
		}
	})
	return segs, nil
}

// WalkInReadingOrder walks the document's CONTENT segments in ROOT_SEQUENCE
// reading order (FR-036, FR-101, T-0276), invoking visit once per CONTENT
// segment with a reader scoped exactly to that segment's octets. It is the
// reading-order counterpart of Walk: the ordering derives from ROOT_SEQUENCE
// (via unitOf + order), not storage order, so extraction text is emitted in
// authored reading order even after a mid-document structural insertion.
// RESOURCE, HISTORY and ATTEST payloads are never read.
func WalkInReadingOrder(r io.ReaderAt, unitOf func(seg ContentSegment) pdlfmt.UnitID, order []pdlfmt.UnitID, visit func(seg ContentSegment, segReader io.Reader) error) error {
	segs, err := ContentSegmentsInReadingOrder(r, unitOf, order)
	if err != nil {
		return err
	}
	for _, seg := range segs {
		sr := io.NewSectionReader(r, int64(seg.Offset), int64(seg.Length))
		if err := visit(seg, sr); err != nil {
			return err
		}
	}
	return nil
}
