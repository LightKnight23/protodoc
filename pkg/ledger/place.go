// Package ledger implements the Protodoc Ledger (PDL) append-only
// sealed-segment region that follows the fixed 1,048,576-octet prefix
// (contracts/container.abnf S1 `ledger = *segment`, S8 "LEDGER PLACEMENT
// FUNCTION", data-model.md layer "Storage (ledger)").
//
// This file implements the foundational placement primitive place()
// (T-0033, FR-056) plus the monotonic storage-ordinal assignment layered
// onto it (T-0034, FR-058): given a prior file image and an edit delta
// describing new segment payloads, it appends new sealed segments to the
// ledger region, never rewrites an octet belonging to a segment already
// present in the prior image, and assigns each appended segment a
// strictly-increasing storage ordinal derived purely from allocation
// sequence. The remaining properties layered on this primitive by later
// M02 tasks -- no-op byte-identical round trip (T-0035), the per-edit
// write-cost budget (T-0036), content-defined chunking (T-0037), and the
// fixed-prefix slot patching that a real commit performs -- are
// deliberately NOT implemented here; each is its own task with its own
// named test and its own requirement id, so this file does not claim
// coverage of theirs. place() here is the append-only, ordinal-assigning
// storage core those tasks build on, nothing more.
package ledger

import (
	"errors"
	"fmt"

	"Protodoc/pkg/container"
)

// PrefixLength is the fixed prefix size, PREFIX_LENGTH = 1,048,576
// (container.abnf S1, ceilings "File prefix total size"). Every ledger
// segment begins at an absolute file offset >= PrefixLength
// (container.abnf S5 slot-offset "MUST be >= 1048576"); this package
// reuses container's own region arithmetic rather than restating the
// constant, so the two can never drift.
const PrefixLength = container.SegmentTableOffset + container.SegmentTableRegionSize

// SegmentPayload is one new sealed-segment body to append to the ledger
// in a single place() call. It is the already-assembled octet sequence of
// a complete segment (segment-header || frames || frame-directory,
// container.abnf S6): place() at this layer treats it as an opaque,
// immutable byte extent and does not parse, reframe or reorder it. Framing
// a payload from content records is the caller's concern (a later M02/M03
// task); this primitive only governs WHERE the bytes land and that nothing
// prior is disturbed.
type SegmentPayload struct {
	// Octets is the complete, sealed octet sequence of the segment. It
	// must be non-empty: a zero-length segment has no header and cannot
	// exist (container.abnf S6.1 fixes a 64-octet segment-header minimum).
	Octets []byte
}

// EditDelta describes the change place() is to apply to a prior image. At
// this foundational layer an edit is expressed purely as the set of new
// sealed segments to append, in the order they are to be appended. A delta
// with no new segments is a no-op at the storage layer: place() returns an
// image byte-identical to prior (the byte-identical open/save round-trip
// guarantee is refined and verified in its own right by T-0035, but the
// append-only primitive already makes an empty delta a zero-octet-mutation
// operation here).
type EditDelta struct {
	// NewSegments are appended to the ledger tail in slice order. Each
	// lands immediately after the previous, contiguous with the end of
	// the prior image, so no gap or padding is introduced between the
	// prior tail and the first new segment.
	NewSegments []SegmentPayload
}

// Placement records where one appended segment landed and the storage
// ordinal it was issued. Ordinal is the segment's canonical
// storage-order position and equals the SegmentTableSlot index it will
// occupy (data-model.md S2.5 "Identity: storage ordinal, equal to its
// SegmentTableSlot index", CQ-007 option B). Per FR-058 the ordinal is a
// pure function of allocation sequence: it derives from no unit's name,
// digest or content bytes, so permuting the payloads' contents while
// holding their append order fixed leaves every Ordinal unchanged.
type Placement struct {
	// Ordinal is the monotonically issued storage ordinal, equal to the
	// segment's SegmentTableSlot index (0-based). Within one document it
	// is strictly increasing in allocation order and never reused.
	Ordinal uint64

	// Offset is the absolute file offset of the segment's first octet in
	// the returned image; always >= PrefixLength (container.abnf S5
	// slot-offset invariant).
	Offset uint64

	// Length is the segment's octet length, copied from its payload.
	Length uint64
}

// PlaceResult is one place() call's outcome: the new file image and, in
// append order, the Placement issued for each appended segment.
type PlaceResult struct {
	// Image is the new file image: prior's octets in [0, len(prior))
	// unchanged, the appended segments in the tail.
	Image []byte

	// Placements holds one entry per delta.NewSegments entry, in the same
	// order, each carrying the ordinal, offset and length that segment was
	// placed at.
	Placements []Placement

	// WrittenOctets is the total octets this commit writes to storage:
	// the appended segment octets plus the fixed-prefix patch a real
	// commit performs for this delta (one CommitRingRecord slot plus one
	// touched SegmentTableSlot per new segment). It is the quantity
	// NFR-008 bounds; see commitWriteCost and EnforceWriteBudget.
	WrittenOctets uint64
}

var (
	// ErrPriorTooShort is returned when the prior image is shorter than
	// the fixed prefix: a well-formed Protodoc file is always at least
	// PrefixLength octets (container.abnf S1), so a shorter prior cannot
	// be a placement base.
	ErrPriorTooShort = errors.New("ledger: prior image shorter than the 1048576-octet fixed prefix")

	// ErrEmptySegmentPayload is returned when a NewSegments entry carries
	// zero octets: a segment must contain at least its 64-octet header
	// (container.abnf S6.1), so an empty payload is a caller error, caught
	// before any octet is appended.
	ErrEmptySegmentPayload = errors.New("ledger: segment payload is empty")

	// ErrTooManySegments is returned when placing the delta's new segments
	// would issue a storage ordinal at or beyond MAX_SEGMENTS (16384): the
	// SegmentTable holds exactly that many slots (container.abnf S5), so
	// ordinal 16384 has no slot to occupy. Checked before any octet is
	// appended (CP-006).
	ErrTooManySegments = errors.New("ledger: placement would exceed MAX_SEGMENTS (16384) storage ordinals")
)

// place appends the delta's new sealed segments to the end of the prior
// file image and returns the resulting new image together with the
// storage ordinal, offset and length issued to each appended segment.
//
// It is the append-only storage primitive of FR-056: the returned image's
// byte range [0, len(prior)) is byte-identical to prior at every offset --
// place never rewrites, relocates or reorders any octet already present in
// prior. New segments occupy the range [len(prior), len(new)) exclusively,
// contiguous and in slice order, so the only octets that differ between
// prior and the result are ones place itself newly allocated.
//
// It also assigns storage ordinals per FR-058 / CQ-007 option B:
// priorSegmentCount is the count of storage ordinals already issued in
// prior (the SegmentTable high-water mark, container.abnf S3
// segment-count). The k-th appended segment (0-based within the delta)
// receives ordinal priorSegmentCount+k. Every issued ordinal is therefore
// a strictly-increasing pure function of allocation sequence and depends
// on no segment's name, digest or content bytes: two byte-identical
// payloads appended at different positions receive different ordinals, and
// permuting the payloads' contents while holding their append order fixed
// leaves the ordinal sequence unchanged. The count of already-issued
// ordinals plus this delta's new ones must not exceed MAX_SEGMENTS.
//
// place does not mutate the caller's prior slice: it returns a freshly
// allocated image and copies prior into its head, so a differential test
// can safely retain prior and compare it against the result. This copy is
// the storage-layer honesty of the append-only guarantee -- prior is
// treated as immutable, exactly as an already-sealed on-disk file is.
//
// place performs no fixed-prefix patching (no ring-slot succession, no
// SegmentTableSlot writes, no index-route update): it computes the ordinal
// each segment WILL occupy but does not itself write the SegmentTable.
// Persisting these placements into slots is a real commit's
// responsibility and lands in its own M02 task; conflating it here would
// make this primitive impossible to test in isolation. Consequently the
// returned image is a valid demonstration of the append-only,
// ordinal-assigning guarantees, not yet a committed, re-openable state.
func place(prior []byte, priorSegmentCount uint64, delta EditDelta) (PlaceResult, error) {
	if len(prior) < PrefixLength {
		return PlaceResult{}, fmt.Errorf("%w: got %d octets", ErrPriorTooShort, len(prior))
	}

	// The highest ordinal this call would issue is
	// priorSegmentCount + len(NewSegments) - 1; it must be < MAX_SEGMENTS.
	// Equivalently priorSegmentCount + len(NewSegments) <= MAX_SEGMENTS.
	// Checked before any allocation (CP-006), with the addition guarded
	// against overflow by comparing against the ceiling first.
	newCount := uint64(len(delta.NewSegments))
	if priorSegmentCount > container.MaxSegments || newCount > container.MaxSegments-priorSegmentCount {
		return PlaceResult{}, fmt.Errorf("%w: %d already issued + %d new", ErrTooManySegments, priorSegmentCount, newCount)
	}

	// Compute the total appended length up front with overflow-safe
	// accumulation, and validate every payload, BEFORE allocating or
	// copying anything (CP-006: verify before you allocate from a
	// declared length). A single invalid payload fails the whole call
	// with no partial image produced.
	var appendLen int
	for i := range delta.NewSegments {
		n := len(delta.NewSegments[i].Octets)
		if n == 0 {
			return PlaceResult{}, fmt.Errorf("%w: NewSegments[%d]", ErrEmptySegmentPayload, i)
		}
		// Guard against int overflow when summing extents on a 32-bit
		// build: appendLen + n must not wrap negative.
		if appendLen > (int(^uint(0)>>1))-n {
			return PlaceResult{}, fmt.Errorf("ledger: appended segment length overflows addressable size at NewSegments[%d]", i)
		}
		appendLen += n
	}

	// Fresh image: prior copied verbatim into the head, new segments
	// appended contiguously into the tail. copy over a freshly-made slice
	// guarantees the head is byte-identical to prior and the tail holds
	// only newly-allocated octets.
	out := make([]byte, len(prior)+appendLen)
	copy(out[:len(prior)], prior)

	placements := make([]Placement, 0, len(delta.NewSegments))
	pos := len(prior)
	for i := range delta.NewSegments {
		seg := delta.NewSegments[i].Octets
		copy(out[pos:pos+len(seg)], seg)
		placements = append(placements, Placement{
			// Ordinal derives ONLY from allocation position, never from
			// seg's content: this is the FR-058 invariant.
			Ordinal: priorSegmentCount + uint64(i),
			Offset:  uint64(pos),
			Length:  uint64(len(seg)),
		})
		pos += len(seg)
	}

	return PlaceResult{
		Image:         out,
		Placements:    placements,
		WrittenOctets: commitWriteCost(uint64(appendLen), newCount),
	}, nil
}
