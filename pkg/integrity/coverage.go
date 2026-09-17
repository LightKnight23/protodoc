// CoverageDescriptor (T-0139, FR-002; integrity.abnf S4): the AUTHORITATIVE
// coverage descriptor carried inside a SIGNATURE record -- distinct from
// container.fm-coverage-summary, which is only the non-normative,
// digest-bound bounded-prefix MIRROR of this. This package is the SINGLE
// canonical CoverageDescriptor wire implementation for the whole codebase:
// M09's FR-063 signature/coverage call sites MUST import and reuse this,
// never re-implement coverage-descriptor encode/decode/validate; a second
// independent implementation of this ABNF production is a defect.
//
// It implements the four PD-COVER structural rules: PD-COVER-001 (no
// zero-length range), PD-COVER-002 (mandatory merge-adjacent canonical
// form), PD-COVER-003 (every ordinal in [0, segmentCount) covered by exactly
// one of the two lists), and PD-COVER-004 (no ATTEST ordinal nameable --
// FR-002's self-coverage circularity closure, via T-0138's predicate).
package integrity

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// CoverageMode and SegmentRange are reused from the container package so
// there is exactly one type for each across the mirror and the authoritative
// descriptor.
type (
	CoverageMode = container.CoverageMode
	SegmentRange = container.SegmentRange
)

const (
	CoverageModeTotal  = container.CoverageModeTotal
	CoverageModeSubset = container.CoverageModeSubset
)

// Prefix-region bitmask bits (integrity.abnf S4 cd-bitmask).
const (
	CoverageBitHeader       = container.CoverageBitHeader
	CoverageBitRingWinner   = container.CoverageBitRingWinner
	CoverageBitFrontmatter  = container.CoverageBitFrontmatter
	CoverageBitSegmentTable = container.CoverageBitSegmentTable
	CoverageBitIntegrity    = container.CoverageBitIntegrity
	coverageBitmaskReserved = 0xE0 // bits 5-7 MBZ
)

// CoverageDescriptor is the authoritative coverage descriptor
// (integrity.abnf S4): cd-mode, cd-covered-ranges, cd-uncovered-ranges,
// cd-bitmask.
type CoverageDescriptor struct {
	Mode      CoverageMode
	Covered   []SegmentRange
	Uncovered []SegmentRange
	Bitmask   byte
}

var (
	// ErrCoverZeroLengthRange is PD-COVER-001.
	ErrCoverZeroLengthRange = errors.New("integrity: coverage range end does not exceed start (PD-COVER-001)")
	// ErrCoverNotCanonical is PD-COVER-002 (unmerged adjacent/overlapping or unsorted ranges).
	ErrCoverNotCanonical = errors.New("integrity: coverage ranges not in merge-adjacent canonical form (PD-COVER-002)")
	// ErrCoverPartition is PD-COVER-003 (an ordinal covered by neither or both lists).
	ErrCoverPartition = errors.New("integrity: an ordinal in [0,segmentCount) is not covered by exactly one list (PD-COVER-003)")
	// ErrCoverAttestNamed is PD-COVER-004 (an ATTEST ordinal named in a list).
	ErrCoverAttestNamed = errors.New("integrity: an ATTEST-typed ordinal is named in a coverage list (PD-COVER-004)")
	// ErrCoverInvalidMode / ErrCoverReservedBits are basic structural rejects.
	ErrCoverInvalidMode   = errors.New("integrity: coverage mode outside the closed {0,1} set")
	ErrCoverReservedBits  = errors.New("integrity: coverage bitmask reserved bits 5-7 are nonzero")
	ErrCoverTruncated     = errors.New("integrity: coverage descriptor encoding truncated")
	ErrCoverTotalNonEmpty = errors.New("integrity: TOTAL-mode descriptor has a non-empty uncovered list")
)

// validateCanonicalRanges checks PD-COVER-001 (no zero-length) and
// PD-COVER-002 (strictly ascending AND merge-adjacent: each range's Start is
// strictly greater than the previous range's End, so no two ranges are
// adjacent or overlapping -- exactly one valid encoding per set).
func validateCanonicalRanges(ranges []SegmentRange) error {
	for i, r := range ranges {
		if r.End <= r.Start {
			return fmt.Errorf("%w: range %d [%d,%d)", ErrCoverZeroLengthRange, i, r.Start, r.End)
		}
		if i > 0 {
			// Canonical: previous End < this Start (a gap of >=1). Equal
			// (adjacent) would be non-canonical (should have merged);
			// less-than-or-overlap is out of order.
			if ranges[i-1].End >= r.Start {
				return fmt.Errorf("%w: range %d [%d,%d) is adjacent/overlapping/unsorted vs previous end %d", ErrCoverNotCanonical, i, r.Start, r.End, ranges[i-1].End)
			}
		}
	}
	return nil
}

// Validate checks the descriptor's four PD-COVER rules against a document of
// segmentCount ordinals whose ATTEST ordinals are attestOrdinals (from
// T-0138). PD-COVER-001/002 apply to both range lists; PD-COVER-003 requires
// every ordinal in [0, segmentCount) to be in exactly one list (TOTAL mode
// puts all non-ATTEST ordinals in covered and leaves uncovered empty);
// PD-COVER-004 forbids any ATTEST ordinal in either list.
func (d CoverageDescriptor) Validate(segmentCount int, attestOrdinals map[uint16]bool) error {
	if d.Mode != CoverageModeTotal && d.Mode != CoverageModeSubset {
		return ErrCoverInvalidMode
	}
	if d.Bitmask&coverageBitmaskReserved != 0 {
		return ErrCoverReservedBits
	}
	if err := validateCanonicalRanges(d.Covered); err != nil {
		return err
	}
	if err := validateCanonicalRanges(d.Uncovered); err != nil {
		return err
	}
	if d.Mode == CoverageModeTotal && len(d.Uncovered) != 0 {
		return ErrCoverTotalNonEmpty
	}

	// PD-COVER-004: no ATTEST ordinal named in either list.
	for _, r := range append(append([]SegmentRange{}, d.Covered...), d.Uncovered...) {
		for o := r.Start; o < r.End; o++ {
			if attestOrdinals[uint16(o)] {
				return fmt.Errorf("%w: ordinal %d", ErrCoverAttestNamed, o)
			}
		}
	}

	// PD-COVER-003: every ordinal in [0, segmentCount) is covered by exactly
	// one list. ATTEST ordinals are excluded from the partition universe
	// (they are in neither list by PD-COVER-004 and are not "coverable").
	inCovered := rangesToSet(d.Covered)
	inUncovered := rangesToSet(d.Uncovered)
	for o := 0; o < segmentCount; o++ {
		if attestOrdinals[uint16(o)] {
			continue
		}
		c, u := inCovered[uint64(o)], inUncovered[uint64(o)]
		var covered bool
		if d.Mode == CoverageModeTotal {
			covered = true // TOTAL covers every non-ATTEST ordinal
		} else {
			covered = c
		}
		// Exactly one of {covered, uncovered} must hold.
		if d.Mode == CoverageModeSubset {
			if c == u { // both or neither
				return fmt.Errorf("%w: ordinal %d (covered=%v uncovered=%v)", ErrCoverPartition, o, c, u)
			}
		} else {
			if u {
				return fmt.Errorf("%w: ordinal %d marked uncovered in TOTAL mode", ErrCoverPartition, o)
			}
			_ = covered
		}
	}
	return nil
}

func rangesToSet(ranges []SegmentRange) map[uint64]bool {
	s := map[uint64]bool{}
	for _, r := range ranges {
		for o := r.Start; o < r.End; o++ {
			s[o] = true
		}
	}
	return s
}

// Encode appends the descriptor's canonical wire encoding (integrity.abnf
// S4): cd-mode(1) || covered plain-seq || uncovered plain-seq || cd-bitmask(1).
func (d CoverageDescriptor) Encode(dst []byte) ([]byte, error) {
	if d.Mode != CoverageModeTotal && d.Mode != CoverageModeSubset {
		return nil, ErrCoverInvalidMode
	}
	if d.Bitmask&coverageBitmaskReserved != 0 {
		return nil, ErrCoverReservedBits
	}
	dst = append(dst, byte(d.Mode))
	dst = appendRanges(dst, d.Covered)
	dst = appendRanges(dst, d.Uncovered)
	dst = append(dst, d.Bitmask)
	return dst, nil
}

func appendRanges(dst []byte, ranges []SegmentRange) []byte {
	elems := make([][]byte, len(ranges))
	for i, r := range ranges {
		var e []byte
		e = pdlfmt.AppendVarint(e, r.Start)
		e = pdlfmt.AppendVarint(e, r.End)
		elems[i] = e
	}
	return pdlfmt.AppendPlainSeq(dst, elems)
}

// Digest returns coverage-descriptor-digest (integrity.abnf S3.2): SHA-256
// over the descriptor's own canonical encoded octets exactly as carried inside
// the SIGNATURE record. This is the value that participates in the
// signed_object preimage, so any change to a range entry or the bitmask
// changes the digest even though the descriptor is also carried verbatim
// alongside it. It is deterministic: equal descriptors always yield the same
// digest, because Encode is the single canonical encoding.
func (d CoverageDescriptor) Digest() (Digest, error) {
	enc, err := d.Encode(nil)
	if err != nil {
		return Digest{}, err
	}
	var out Digest
	sum := sha256.Sum256(enc)
	copy(out[:], sum[:])
	return out, nil
}

// DecodeCoverageDescriptor decodes a descriptor from the leading octets of
// src, returning it and the octets consumed. It rejects an invalid mode and
// reserved bitmask bits structurally; PD-COVER-001..004 are checked by
// Validate (some callers validate against a specific segmentCount).
func DecodeCoverageDescriptor(src []byte) (CoverageDescriptor, int, error) {
	var d CoverageDescriptor
	if len(src) < 1 {
		return d, 0, ErrCoverTruncated
	}
	mode := CoverageMode(src[0])
	if mode != CoverageModeTotal && mode != CoverageModeSubset {
		return d, 0, ErrCoverInvalidMode
	}
	d.Mode = mode
	pos := 1

	covered, n, err := decodeRanges(src[pos:])
	if err != nil {
		return CoverageDescriptor{}, 0, err
	}
	d.Covered = covered
	pos += n

	uncovered, n, err := decodeRanges(src[pos:])
	if err != nil {
		return CoverageDescriptor{}, 0, err
	}
	d.Uncovered = uncovered
	pos += n

	if pos >= len(src) {
		return CoverageDescriptor{}, 0, ErrCoverTruncated
	}
	if src[pos]&coverageBitmaskReserved != 0 {
		return CoverageDescriptor{}, 0, ErrCoverReservedBits
	}
	d.Bitmask = src[pos]
	pos++
	return d, pos, nil
}

func decodeRanges(src []byte) ([]SegmentRange, int, error) {
	count, n, err := pdlfmt.DecodeSeqCount(src)
	if err != nil {
		return nil, 0, fmt.Errorf("integrity: coverage range count: %w", err)
	}
	pos := n
	if count == 0 {
		return nil, pos, nil
	}
	// FR-106 / CP-012: a declared element count is checked against the octets
	// actually remaining BEFORE any allocation of that size. Each range is at
	// least two octets (two minimal 1-octet varints), so a count exceeding
	// len(remaining)/2 cannot possibly be satisfied and is rejected rather
	// than used to size an allocation from untrusted input.
	remaining := len(src) - pos
	if count > uint64(remaining/2) {
		return nil, 0, fmt.Errorf("%w: coverage range count %d exceeds the %d octets remaining", ErrCoverTruncated, count, remaining)
	}
	ranges := make([]SegmentRange, 0, count)
	for i := uint64(0); i < count; i++ {
		if pos >= len(src) {
			return nil, 0, ErrCoverTruncated
		}
		start, n1, err := pdlfmt.DecodeVarint(src[pos:])
		if err != nil {
			return nil, 0, fmt.Errorf("integrity: coverage range %d start: %w", i, err)
		}
		pos += n1
		end, n2, err := pdlfmt.DecodeVarint(src[pos:])
		if err != nil {
			return nil, 0, fmt.Errorf("integrity: coverage range %d end: %w", i, err)
		}
		pos += n2
		ranges = append(ranges, SegmentRange{Start: start, End: end})
	}
	return ranges, pos, nil
}
