// Package ledger implements the Protodoc Ledger (PDL) append-only
// sealed-segment region that follows the fixed 1,048,576-octet prefix
// (contracts/container.abnf S1 `ledger = *segment`, S8 "LEDGER PLACEMENT
// FUNCTION", data-model.md layer "Storage (ledger)").
//
// This file implements only the foundational placement primitive place()
// (T-0033, FR-056): given a prior file image and an edit delta describing
// new segment payloads, it appends new sealed segments to the ledger
// region and never rewrites an octet belonging to a segment already
// present in the prior image. The properties layered on top of this
// primitive by later M02 tasks -- monotonic storage-ordinal assignment
// (T-0034), no-op byte-identical round trip (T-0035), the per-edit
// write-cost budget (T-0036), content-defined chunking (T-0037), and the
// fixed-prefix slot patching that a real commit performs -- are
// deliberately NOT implemented here; each is its own task with its own
// named test and its own requirement id, so this task does not claim
// coverage of theirs. place() here is the append-only storage core
// those tasks build on, nothing more.
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
)

// place appends the delta's new sealed segments to the end of the prior
// file image and returns the resulting new image. It is the append-only
// storage primitive of FR-056: the returned image's byte range
// [0, len(prior)) is byte-identical to prior at every offset -- place
// never rewrites, relocates or reorders any octet already present in
// prior. New segments occupy the range [len(prior), len(new)) exclusively,
// contiguous and in slice order, so the only octets that differ between
// prior and the result are ones place itself newly allocated.
//
// place does not mutate the caller's prior slice: it returns a freshly
// allocated image and copies prior into its head, so a differential test
// can safely retain prior and compare it against the result. This copy is
// the storage-layer honesty of the append-only guarantee -- prior is
// treated as immutable, exactly as an already-sealed on-disk file is.
//
// place performs no fixed-prefix patching (no ring-slot succession, no
// SegmentTableSlot writes, no index-route update). Those are a real
// commit's responsibility and land in their own M02 tasks; conflating
// them here would make this primitive impossible to test in isolation for
// the one property FR-056 pins on it. Consequently the returned image is a
// valid demonstration of the append-only extent guarantee, not yet a
// committed, re-openable document state.
func place(prior []byte, delta EditDelta) ([]byte, error) {
	if len(prior) < PrefixLength {
		return nil, fmt.Errorf("%w: got %d octets", ErrPriorTooShort, len(prior))
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
			return nil, fmt.Errorf("%w: NewSegments[%d]", ErrEmptySegmentPayload, i)
		}
		// Guard against int overflow when summing extents on a 32-bit
		// build: appendLen + n must not wrap negative.
		if appendLen > (int(^uint(0)>>1))-n {
			return nil, fmt.Errorf("ledger: appended segment length overflows addressable size at NewSegments[%d]", i)
		}
		appendLen += n
	}

	// Fresh image: prior copied verbatim into the head, new segments
	// appended contiguously into the tail. copy over a freshly-made slice
	// guarantees the head is byte-identical to prior and the tail holds
	// only newly-allocated octets.
	out := make([]byte, len(prior)+appendLen)
	copy(out[:len(prior)], prior)

	pos := len(prior)
	for i := range delta.NewSegments {
		seg := delta.NewSegments[i].Octets
		copy(out[pos:pos+len(seg)], seg)
		pos += len(seg)
	}

	return out, nil
}
