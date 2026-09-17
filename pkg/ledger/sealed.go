// Sealed-segment write-once immutability guard (T-0046, FR-056). A segment
// is immutable once sealed (container.abnf S6 "a segment is immutable once
// sealed (never rewritten in place)"): the only lawful writes are appends
// elsewhere, a fixed-prefix slot patch, or a whole-segment relocation by
// partial compaction restricted to uncovered ordinals. This guard is the
// explicit, defensive rejection -- distinct from place()'s happy-path
// append logic -- of any code path that attempts to reopen or write through
// a handle into an already-sealed segment's byte range, so an accidental
// in-place write surfaces a structured error instead of silently corrupting
// a sealed extent.
package ledger

import (
	"errors"
	"fmt"
	"sort"
)

// sealedExtent is one sealed segment's immutable byte range, keyed by its
// storage ordinal.
type sealedExtent struct {
	ordinal uint64
	offset  uint64
	length  uint64
}

// SealedSegmentGuard tracks the byte ranges of sealed segments and rejects
// any write that would touch one. It is append-friendly: Seal records a new
// sealed range, and WriteRange permits any write that lies entirely outside
// every sealed range (an append past the current tail, or a fixed-prefix
// patch) while rejecting any write that overlaps a sealed range.
type SealedSegmentGuard struct {
	// extents is kept sorted by offset so WriteRange can binary-search the
	// candidate overlapping extent rather than scan all of them.
	extents []sealedExtent
}

// ErrSealedSegmentImmutable is the defensive immutability-violation error a
// write into an already-sealed segment's range produces.
var ErrSealedSegmentImmutable = errors.New("ledger: write into an already-sealed segment is refused (FR-056 write-once immutability)")

// ErrSegmentAlreadySealed is returned when Seal is called twice for the
// same ordinal, or for an ordinal whose range overlaps an already-sealed
// one: a sealed segment is never re-sealed or overlaid.
var ErrSegmentAlreadySealed = errors.New("ledger: segment ordinal already sealed")

// SealImmutabilityError carries the ordinal and range a rejected write
// collided with, so a caller sees exactly which sealed segment it tried to
// mutate.
type SealImmutabilityError struct {
	Ordinal     uint64
	SealedStart uint64
	SealedEnd   uint64
	WriteStart  uint64
	WriteEnd    uint64
}

func (e *SealImmutabilityError) Error() string {
	return fmt.Sprintf("ledger: write [%d,%d) overlaps sealed segment ordinal %d at [%d,%d): refused (FR-056 write-once immutability)",
		e.WriteStart, e.WriteEnd, e.Ordinal, e.SealedStart, e.SealedEnd)
}

func (e *SealImmutabilityError) Is(target error) bool {
	return target == ErrSealedSegmentImmutable
}

// Seal records the byte range [offset, offset+length) as belonging to a
// newly sealed segment with the given ordinal. It rejects a zero-length
// range, an ordinal already sealed, and a range overlapping any existing
// sealed extent (a new segment's octets are freshly allocated, so they can
// never overlap a prior one).
func (g *SealedSegmentGuard) Seal(ordinal, offset, length uint64) error {
	if length == 0 {
		return fmt.Errorf("ledger: cannot seal a zero-length segment (ordinal %d)", ordinal)
	}
	end := offset + length
	if end < offset {
		return fmt.Errorf("ledger: sealed range for ordinal %d overflows", ordinal)
	}
	for _, e := range g.extents {
		if e.ordinal == ordinal {
			return fmt.Errorf("%w: ordinal %d", ErrSegmentAlreadySealed, ordinal)
		}
		if rangesOverlap(offset, end, e.offset, e.offset+e.length) {
			return fmt.Errorf("%w: ordinal %d range [%d,%d) overlaps sealed ordinal %d [%d,%d)",
				ErrSegmentAlreadySealed, ordinal, offset, end, e.ordinal, e.offset, e.offset+e.length)
		}
	}
	g.extents = append(g.extents, sealedExtent{ordinal: ordinal, offset: offset, length: length})
	sort.Slice(g.extents, func(i, j int) bool { return g.extents[i].offset < g.extents[j].offset })
	return nil
}

// WriteRange reports whether a write to [offset, offset+length) is
// permitted: it returns a *SealImmutabilityError (matching
// ErrSealedSegmentImmutable) if the write overlaps any sealed segment's
// range, and nil otherwise. A zero-length write touches nothing and is
// always permitted. This is the guard every in-place-write path must
// consult before writing.
func (g *SealedSegmentGuard) WriteRange(offset, length uint64) error {
	if length == 0 {
		return nil
	}
	end := offset + length
	if end < offset {
		return fmt.Errorf("ledger: write range [%d,+%d) overflows", offset, length)
	}
	// Binary search for the first sealed extent whose end is beyond offset;
	// only that extent and its immediate neighbours can overlap.
	i := sort.Search(len(g.extents), func(i int) bool {
		return g.extents[i].offset+g.extents[i].length > offset
	})
	if i < len(g.extents) {
		e := g.extents[i]
		if rangesOverlap(offset, end, e.offset, e.offset+e.length) {
			return &SealImmutabilityError{
				Ordinal:     e.ordinal,
				SealedStart: e.offset,
				SealedEnd:   e.offset + e.length,
				WriteStart:  offset,
				WriteEnd:    end,
			}
		}
	}
	return nil
}

// SealedCount returns how many segments the guard currently protects.
func (g *SealedSegmentGuard) SealedCount() int { return len(g.extents) }

// rangesOverlap reports whether half-open [aStart,aEnd) and [bStart,bEnd)
// intersect.
func rangesOverlap(aStart, aEnd, bStart, bEnd uint64) bool {
	return aStart < bEnd && bStart < aEnd
}
