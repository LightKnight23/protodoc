// Bounded-prefix coverage determination (TR-007, contracts/container.abnf
// S4 fm-coverage-summary NORMATIVE comment, S5 slot-flags NORMATIVE
// comment, integrity.abnf S4 coverage-descriptor): a scanner that has read
// only the leading 1,048,576 octets of a document can determine which
// stored units are covered by the document's integrity protection, without
// reading any ATTEST segment body. This package does not define the
// authoritative CoverageDescriptor (data-model.md places it in the
// Integrity layer, built by a later milestone against the SIGNATURE
// segment) — only Frontmatter.fm-coverage-summary, the NON-NORMATIVE,
// digest-bound mirror of it that lives inside the bounded prefix.
package container

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// Closed cd-mode enum (integrity.abnf S4): 0x02-0xFF is a reserved range
// this package rejects, per that section's own "reject" instruction.
type CoverageMode byte

const (
	CoverageModeTotal  CoverageMode = 0
	CoverageModeSubset CoverageMode = 1
)

func (m CoverageMode) valid() bool {
	return m == CoverageModeTotal || m == CoverageModeSubset
}

// Coverage-descriptor prefix-region bitmask bits (integrity.abnf S4
// cd-bitmask NORMATIVE comment): covered_prefix_regions_bitmask.
const (
	CoverageBitHeader       = 1 << 0
	CoverageBitRingWinner   = 1 << 1
	CoverageBitFrontmatter  = 1 << 2
	CoverageBitSegmentTable = 1 << 3
	CoverageBitIntegrity    = 1 << 4
	// coverageBitmaskReservedMask is bits 5-7, MBZ (integrity.abnf S4).
	coverageBitmaskReservedMask = 0xE0
)

// SegmentRange is one half-open segment ordinal range [Start, End)
// (integrity.abnf S4 segment-range: sr-start, sr-end).
type SegmentRange struct {
	Start uint64
	End   uint64
}

var (
	// ErrCoverageInvalidMode is returned when cd-mode is outside the closed {0,1} set.
	ErrCoverageInvalidMode = errors.New("container: coverage-summary-entry mode outside the closed {0,1} set")
	// ErrCoverageBadRange rejects a segment-range with End <= Start (rule PD-COVER-001).
	ErrCoverageBadRange = errors.New("container: coverage segment-range end does not exceed start (PD-COVER-001)")
	// ErrCoverageReservedBitmaskBits rejects a nonzero reserved cd-bitmask bit (5-7).
	ErrCoverageReservedBitmaskBits = errors.New("container: coverage-summary-entry bitmask reserved bits 5-7 are nonzero")
	// ErrCoverageSummaryTruncated is returned when a coverage-summary-entry's declared shape runs past its enclosing field-value.
	ErrCoverageSummaryTruncated = errors.New("container: coverage-summary-entry truncated")
)

// CoverageSummaryEntry is one coverage-summary-entry (contracts/
// container.abnf S4 fm-coverage-summary NORMATIVE comment): a NON-NORMATIVE,
// digest-bound mirror of one currently-present signature's coverage_descriptor
// (integrity.abnf S4), present so TR-007 is answerable from the bounded
// prefix alone without decoding an ATTEST segment that may lie past it.
type CoverageSummaryEntry struct {
	Mode                 CoverageMode
	CoveredRanges        []SegmentRange
	UncoveredRanges      []SegmentRange
	PrefixRegionsBitmask byte
	// DescriptorDigest is SHA-256 over the authoritative coverage_descriptor
	// this entry mirrors (see ComputeCoverageDescriptorDigest), so a full
	// reader that later decodes the real descriptor can detect a stale
	// mirror rather than trusting the cache at face value.
	DescriptorDigest pdlfmt.Digest256
}

// ComputeCoverageDescriptorDigest computes the digest a CoverageSummaryEntry
// binds itself to: SHA-256 over the canonical encoding of the coverage
// descriptor's own fields (mode, both range lists, bitmask). A mirror whose
// DescriptorDigest does not match this recomputed over the actual
// coverage_descriptor it claims to mirror is stale (contracts/
// container.abnf S4: "A full reader detecting a mismatch between this
// mirror and the authoritative descriptor treats the mirror as stale").
func ComputeCoverageDescriptorDigest(mode CoverageMode, covered, uncovered []SegmentRange, bitmask byte) pdlfmt.Digest256 {
	var buf []byte
	buf = append(buf, byte(mode))
	buf = appendSegmentRanges(buf, covered)
	buf = appendSegmentRanges(buf, uncovered)
	buf = append(buf, bitmask)
	return pdlfmt.Digest256(sha256.Sum256(buf))
}

func appendSegmentRanges(dst []byte, ranges []SegmentRange) []byte {
	elements := make([][]byte, len(ranges))
	for i, r := range ranges {
		var e []byte
		e = pdlfmt.AppendVarint(e, r.Start)
		e = pdlfmt.AppendVarint(e, r.End)
		elements[i] = e
	}
	return pdlfmt.AppendPlainSeq(dst, elements)
}

// validateSegmentRanges rejects any range with End <= Start (rule
// PD-COVER-001): a zero-length or inverted range is a structural
// rejection, never silently dropped, at encode time as well as decode time.
func validateSegmentRanges(ranges []SegmentRange) error {
	for i, r := range ranges {
		if r.End <= r.Start {
			return fmt.Errorf("%w: range %d is [%d,%d)", ErrCoverageBadRange, i, r.Start, r.End)
		}
	}
	return nil
}

func decodeSegmentRanges(src []byte) ([]SegmentRange, int, error) {
	count, n, err := pdlfmt.DecodeSeqCount(src)
	if err != nil {
		return nil, 0, fmt.Errorf("container: coverage range count: %w", err)
	}
	offset := n
	if count == 0 {
		return nil, offset, nil
	}
	ranges := make([]SegmentRange, 0, count)
	for i := uint64(0); i < count; i++ {
		if offset >= len(src) {
			return nil, 0, ErrCoverageSummaryTruncated
		}
		start, n1, err := pdlfmt.DecodeVarint(src[offset:])
		if err != nil {
			return nil, 0, fmt.Errorf("container: coverage range %d start: %w", i, err)
		}
		offset += n1
		end, n2, err := pdlfmt.DecodeVarint(src[offset:])
		if err != nil {
			return nil, 0, fmt.Errorf("container: coverage range %d end: %w", i, err)
		}
		offset += n2
		if end <= start {
			return nil, 0, fmt.Errorf("%w: range %d is [%d,%d)", ErrCoverageBadRange, i, start, end)
		}
		ranges = append(ranges, SegmentRange{Start: start, End: end})
	}
	return ranges, offset, nil
}

// encode appends e's canonical encoding (mode, covered ranges, uncovered
// ranges, bitmask, digest) to dst. The encoding is self-delimiting: a
// decoder never needs a separate length prefix to know where one entry
// ends and the next (plain-seq-of(coverage-summary-entry)'s next element)
// begins.
func (e CoverageSummaryEntry) encode(dst []byte) ([]byte, error) {
	if !e.Mode.valid() {
		return nil, ErrCoverageInvalidMode
	}
	if e.PrefixRegionsBitmask&coverageBitmaskReservedMask != 0 {
		return nil, ErrCoverageReservedBitmaskBits
	}
	if err := validateSegmentRanges(e.CoveredRanges); err != nil {
		return nil, err
	}
	if err := validateSegmentRanges(e.UncoveredRanges); err != nil {
		return nil, err
	}
	dst = append(dst, byte(e.Mode))
	dst = appendSegmentRanges(dst, e.CoveredRanges)
	dst = appendSegmentRanges(dst, e.UncoveredRanges)
	dst = append(dst, e.PrefixRegionsBitmask)
	dst = pdlfmt.AppendDigest256(dst, e.DescriptorDigest)
	return dst, nil
}

// decodeCoverageSummaryEntry decodes one coverage-summary-entry from the
// start of src, returning the entry and the number of octets consumed.
func decodeCoverageSummaryEntry(src []byte) (CoverageSummaryEntry, int, error) {
	var e CoverageSummaryEntry
	if len(src) < 1 {
		return e, 0, ErrCoverageSummaryTruncated
	}
	mode := CoverageMode(src[0])
	if !mode.valid() {
		return e, 0, ErrCoverageInvalidMode
	}
	e.Mode = mode
	offset := 1

	covered, n, err := decodeSegmentRanges(src[offset:])
	if err != nil {
		return CoverageSummaryEntry{}, 0, err
	}
	e.CoveredRanges = covered
	offset += n

	uncovered, n, err := decodeSegmentRanges(src[offset:])
	if err != nil {
		return CoverageSummaryEntry{}, 0, err
	}
	e.UncoveredRanges = uncovered
	offset += n

	if offset >= len(src) {
		return CoverageSummaryEntry{}, 0, ErrCoverageSummaryTruncated
	}
	bitmask := src[offset]
	if bitmask&coverageBitmaskReservedMask != 0 {
		return CoverageSummaryEntry{}, 0, ErrCoverageReservedBitmaskBits
	}
	e.PrefixRegionsBitmask = bitmask
	offset++

	digest, n, err := pdlfmt.DecodeDigest256(src[offset:])
	if err != nil {
		return CoverageSummaryEntry{}, 0, fmt.Errorf("container: coverage-summary-entry digest: %w", err)
	}
	e.DescriptorDigest = digest
	offset += n

	return e, offset, nil
}

// EncodeCoverageSummary encodes entries as fm-coverage-summary's
// plain-seq-of(coverage-summary-entry) field-value.
func EncodeCoverageSummary(entries []CoverageSummaryEntry) ([]byte, error) {
	elements := make([][]byte, len(entries))
	for i, e := range entries {
		enc, err := e.encode(nil)
		if err != nil {
			return nil, fmt.Errorf("container: coverage-summary-entry %d: %w", i, err)
		}
		elements[i] = enc
	}
	return pdlfmt.AppendPlainSeq(nil, elements), nil
}

// DecodeCoverageSummary decodes fm-coverage-summary's field-value
// (plain-seq-of(coverage-summary-entry)) back into its entries. It rejects
// any trailing octets past the last decoded entry (a malformed count or an
// oversized field-value), the same "no ambiguity in the boundary" rule
// DecodeRecord already applies to record framing.
func DecodeCoverageSummary(value []byte) ([]CoverageSummaryEntry, error) {
	count, n, err := pdlfmt.DecodeSeqCount(value)
	if err != nil {
		return nil, fmt.Errorf("container: fm-coverage-summary entry count: %w", err)
	}
	offset := n
	entries := make([]CoverageSummaryEntry, 0, count)
	for i := uint64(0); i < count; i++ {
		e, consumed, err := decodeCoverageSummaryEntry(value[offset:])
		if err != nil {
			return nil, fmt.Errorf("container: fm-coverage-summary entry %d: %w", i, err)
		}
		entries = append(entries, e)
		offset += consumed
	}
	if offset != len(value) {
		return nil, fmt.Errorf("%w: %d trailing octets after %d entries", ErrCoverageSummaryTruncated, len(value)-offset, count)
	}
	return entries, nil
}

// CoverageResult is TR-007's bounded-prefix determination: which stored
// units (SegmentTable ordinals) are covered by the document's integrity
// protection, and which are not, as derivable from Frontmatter.fm-coverage-
// summary alone. ATTEST-typed and unused ordinals are excluded from both
// lists: an ATTEST segment is categorically never nameable in a coverage
// descriptor (integrity.abnf S4 PD-COVER-004), and an unused slot is not a
// stored unit at all (S5.1).
type CoverageResult struct {
	Covered   []int
	Uncovered []int
	// HintMismatches lists ordinals where the SegmentTableSlot.Flags
	// coverage-hint bit (slot-flags bit0) disagrees with the covered/
	// uncovered classification derived from fm-coverage-summary. The hint
	// is never authoritative on its own (contracts/container.abnf S5
	// slot-flags NORMATIVE comment), so a mismatch here is reported, not
	// treated as an error: it is exactly the kind of drift TR-007's
	// "wiring together" of the two signals exists to surface.
	HintMismatches []int
}

// ComputeStoredUnitCoverage computes CoverageResult from a decoded
// SegmentTable and the Frontmatter's decoded fm-coverage-summary entries.
// A stored, non-ATTEST unit is covered if at least one entry covers it
// (TOTAL mode covers every currently-populated non-ATTEST ordinal; SUBSET
// mode covers exactly its declared covered_segment_ranges); it is
// uncovered otherwise. This is the mirror's determination (contracts/
// container.abnf S4: "MAY report the mirrored coverage as a determination
// but MUST NOT report it as a cryptographic verification") — it performs
// no signature verification and reads no ATTEST segment body.
func ComputeStoredUnitCoverage(table [MaxSegments]SegmentTableSlot, summary []CoverageSummaryEntry) CoverageResult {
	covered := make([]bool, MaxSegments)
	considered := make([]bool, MaxSegments)
	for i, slot := range table {
		considered[i] = slot.SegmentType != SegmentTypeUnused && slot.SegmentType != SegmentTypeAttest
	}

	for _, entry := range summary {
		if entry.Mode == CoverageModeTotal {
			for i := range covered {
				if considered[i] {
					covered[i] = true
				}
			}
			continue
		}
		for _, r := range entry.CoveredRanges {
			markRange(covered, considered, r, true)
		}
	}

	var result CoverageResult
	for i := 0; i < MaxSegments; i++ {
		if !considered[i] {
			continue
		}
		if covered[i] {
			result.Covered = append(result.Covered, i)
		} else {
			result.Uncovered = append(result.Uncovered, i)
		}
		hinted := table[i].Flags&SlotFlagCoverageHint != 0
		if hinted != covered[i] {
			result.HintMismatches = append(result.HintMismatches, i)
		}
	}
	return result
}

// markRange sets covered[i] = value for every considered ordinal in the
// half-open range [r.Start, r.End), clamped to the table's bounds so a
// malformed range naming an ordinal >= MaxSegments cannot index out of
// range.
func markRange(covered, considered []bool, r SegmentRange, value bool) {
	end := r.End
	if end > uint64(len(covered)) {
		end = uint64(len(covered))
	}
	for i := r.Start; i < end; i++ {
		if considered[i] {
			covered[i] = value
		}
	}
}

// StoredUnitCoverageFromBoundedPrefix decodes the SegmentTable and
// Frontmatter regions out of prefix and computes CoverageResult (TR-007).
// prefix must hold at least the leading SegmentTableOffset+
// SegmentTableRegionSize (1,048,576) octets of the document; this function
// reads no octet at or past that bound and decodes no ATTEST segment body.
func StoredUnitCoverageFromBoundedPrefix(prefix []byte) (CoverageResult, error) {
	if len(prefix) < SegmentTableOffset+SegmentTableRegionSize {
		return CoverageResult{}, ErrBoundedPrefixTruncated
	}
	table, err := DecodeSegmentTable(prefix[SegmentTableOffset : SegmentTableOffset+SegmentTableRegionSize])
	if err != nil {
		return CoverageResult{}, fmt.Errorf("container: computing stored-unit coverage: %w", err)
	}
	fm, err := DecodeFrontmatter(prefix[FrontmatterOffset:FrontmatterRegionSize+FrontmatterOffset])
	if err != nil {
		return CoverageResult{}, fmt.Errorf("container: computing stored-unit coverage: %w", err)
	}
	return ComputeStoredUnitCoverage(table, fm.CoverageSummary), nil
}
