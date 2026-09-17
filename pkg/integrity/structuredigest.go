// structure_digest (T-0140, FR-002/TR-009; integrity.abnf S3.1): a verifier
// computes this FRESH from the current file's octets, restricted to the
// covered_prefix_regions_bitmask a specific signature claims, never reading a
// stored copy as authoritative. The preimage is assembled in a fixed 8-item
// order; each item's presence is conditional on a bitmask bit except items
// 1, 4 and 6 which are always present; the assembled buffer is hashed once
// with SHA-256.
//
// This file assembles the preimage from freshly-recomputed materials the
// caller supplies (segment summary digests and the T_S root MUST be
// recomputed fresh by the caller, never read from stored slot-digests).
package integrity

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// coveredSegmentSummary is item 7's per-covered-ordinal entry (integrity.abnf
// S3.1): ordinal, type, length, and a FRESHLY recomputed digest over the
// segment's current octets (never the stored slot-digest).
type CoveredSegmentSummary struct {
	Ordinal uint16
	Type    byte
	Length  uint64 // u48
	Digest  Digest // recomputed fresh by the caller
}

// StructureDigestInput carries the freshly-recomputed materials the
// structure_digest preimage draws on. Conditional items are included per the
// Bitmask; the caller is responsible for having recomputed HeaderBytes,
// RingWinnerBytes, FrontmatterMetadataBytes, CoveredSummaries and TSRoot from
// current octets.
type StructureDigestInput struct {
	Bitmask                  byte
	HeaderBytes              []byte                  // header[0,480), item 2 (bit HEADER)
	RingWinnerBytes          []byte                  // winning ring record[0,480), item 3 (bit RING_WINNER)
	LedgerLength             uint64                  // item 4, always
	FrontmatterMetadataBytes []byte                  // item 5 (bit FRONTMATTER)
	CoveredSummaries         []CoveredSegmentSummary // item 7 (bit SEGMENT_TABLE), caller-sorted by ordinal
	TSRoot                   Digest                  // item 8 (bit INTEGRITY_BLOCK), recomputed fresh
}

// ErrStructureReservedBits is returned when a reserved bitmask bit (5-7) is
// set: assembly is rejected before it begins.
var ErrStructureReservedBits = errors.New("integrity: structure_digest bitmask reserved bits 5-7 are set")

// AssembleStructureDigestPreimage builds the fixed-order 8-item preimage
// buffer per integrity.abnf S3.1. Items 1 (domain tag 0x0A), 4
// (ledger-length) and 6 (the bitmask octet) are always present; items 2, 3,
// 5, 7, 8 are included iff their bit is set. It rejects a reserved bit
// (5-7) before assembling.
func AssembleStructureDigestPreimage(in StructureDigestInput) ([]byte, error) {
	if in.Bitmask&coverageBitmaskReserved != 0 {
		return nil, fmt.Errorf("%w: 0x%02x", ErrStructureReservedBits, in.Bitmask)
	}
	var buf []byte
	// Item 1: domain tag 0x0A, always.
	buf = append(buf, DomainStructureDigest)
	// Item 2: header[0,480) iff HEADER.
	if in.Bitmask&CoverageBitHeader != 0 {
		buf = append(buf, in.HeaderBytes...)
	}
	// Item 3: winning ring record[0,480) iff RING_WINNER.
	if in.Bitmask&CoverageBitRingWinner != 0 {
		buf = append(buf, in.RingWinnerBytes...)
	}
	// Item 4: ledger-length (8 octets big-endian), always.
	var ll [8]byte
	binary.BigEndian.PutUint64(ll[:], in.LedgerLength)
	buf = append(buf, ll[:]...)
	// Item 5: frontmatter metadata bytes iff FRONTMATTER.
	if in.Bitmask&CoverageBitFrontmatter != 0 {
		buf = append(buf, in.FrontmatterMetadataBytes...)
	}
	// Item 6: the bitmask octet itself, always.
	buf = append(buf, in.Bitmask)
	// Item 7: sorted covered-segment-summary vec iff SEGMENT_TABLE.
	if in.Bitmask&CoverageBitSegmentTable != 0 {
		for _, s := range in.CoveredSummaries {
			var e [2 + 1 + 6 + 32]byte
			binary.BigEndian.PutUint16(e[0:2], s.Ordinal)
			e[2] = s.Type
			// length as u48 big-endian (6 octets)
			putU48(e[3:9], s.Length)
			copy(e[9:41], s.Digest[:])
			buf = append(buf, e[:]...)
		}
	}
	// Item 8: T_S root (recomputed fresh) iff INTEGRITY_BLOCK.
	if in.Bitmask&CoverageBitIntegrity != 0 {
		buf = append(buf, in.TSRoot[:]...)
	}
	return buf, nil
}

// putU48 writes v as 6 big-endian octets into dst[:6].
func putU48(dst []byte, v uint64) {
	dst[0] = byte(v >> 40)
	dst[1] = byte(v >> 32)
	dst[2] = byte(v >> 24)
	dst[3] = byte(v >> 16)
	dst[4] = byte(v >> 8)
	dst[5] = byte(v)
}

// StructureDigest assembles the preimage and hashes it once with SHA-256
// (integrity.abnf S3.1). It errors if a reserved bitmask bit is set.
func StructureDigest(in StructureDigestInput) (Digest, error) {
	preimage, err := AssembleStructureDigestPreimage(in)
	if err != nil {
		return Digest{}, err
	}
	var d Digest
	sum := sha256.Sum256(preimage)
	copy(d[:], sum[:])
	return d, nil
}
