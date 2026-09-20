// Command protodoc-gensample writes protodoc-sample.pdl: the minimal, valid,
// BOTTOM-state Protodoc document used as the PRONOM/IANA registration sample
// (magic "PDL1" + four zeros at offset 0, the byte signature DROID matches).
//
// It is a genuine minimal document: the fixed 1,048,576-octet prefix (header,
// self-digesting commit ring, frontmatter, segment table) with zero segments.
// Unlike the original hand-built sample, it seals the winning commit-ring
// record's ledger_root to T_S over the (empty) full segment table and its
// frontmatter/segment-table/structure digests to their real values, so the
// sample passes `protodoc validate`'s storage_integrity_tree check (T-0396).
// The magic-header bytes at offset 0 are unchanged, so the PRONOM signature
// (50 44 4C 31 00 00 00 00) still matches.
package main

import (
	"crypto/sha256"
	"fmt"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
)

// sha256Of returns the SHA-256 of b as a [32]byte.
func sha256Of(b []byte) [32]byte { return sha256.Sum256(b) }

func main() {
	out := "protodoc-sample.pdl"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	img, err := buildSample()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gensample:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, img, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gensample: write:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d octets)\n", out, len(img))
}

func buildSample() ([]byte, error) {
	prefixLen := container.SegmentTableOffset + container.SegmentTableRegionSize

	h := &container.Header{
		FormatMajor: 1, FormatMinor: 0, DocumentClass: 1,
		CapabilityWritten: 1, CapabilityRequired: 1,
		HistoryMode: container.HistoryNone, UnicodeVersionID: 1,
		ShapingProfileID: 1, PrefixLayoutID: 1,
	}
	fm := &container.Frontmatter{
		Title:     "Protodoc minimal sample document",
		PageCount: 1,
		Language:  "en-US",
	}

	// Zero segments (BOTTOM state). T_S is computed over the FULL decoded
	// segment-table width (all MaxSegments slots absent/zero), matching what
	// `validate`'s storage_integrity_tree check recomputes.
	fullSlots := make([]container.SegmentTableSlot, container.MaxSegments)
	ledgerRoot, err := integrity.TSRoot(fullSlots)
	if err != nil {
		return nil, fmt.Errorf("T_S root: %w", err)
	}

	// Encode header + frontmatter + segment table so their real digests can be
	// sealed into the ring winner.
	headerBytes := h.Encode(nil)
	fmBytes, err := fm.Encode(nil)
	if err != nil {
		return nil, fmt.Errorf("frontmatter encode: %w", err)
	}
	stBytes, err := container.EncodeSegmentTable(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("segment table encode: %w", err)
	}
	fmDigest := container.ComputeFrontmatterPreviewDigest(fm)
	segTableDigest := integrity.Digest(sha256Of(stBytes))

	// Build the winning commit-ring record with the sealed digests. It gets the
	// UNIQUE highest sequence so SelectWinner picks it unambiguously (a ring
	// with several equal-highest sequences is a PD-RING-001 tie).
	const winnerSeq = uint64(container.CommitRingSlots) // 7: unique highest
	winner := container.CommitRingRecord{
		Sequence:     winnerSeq,
		LedgerLength: uint64(prefixLen),
		SegmentCount: 0,
	}
	copy(winner.LedgerRoot[:], ledgerRoot[:])
	copy(winner.FrontmatterDigest[:], fmDigest[:])
	copy(winner.SegmentTableDigest[:], segTableDigest[:])

	// structure_digest over header + ring-winner + ledger-length + (empty)
	// segment-table summaries, matching the writer convention.
	var ringWinnerBytes [container.RingSlotSize]byte
	winner.Encode(ringWinnerBytes[:])
	sd, err := integrity.StructureDigest(integrity.StructureDigestInput{
		Bitmask:         integrity.CoverageBitHeader | integrity.CoverageBitRingWinner | integrity.CoverageBitSegmentTable,
		HeaderBytes:     padTo(headerBytes, 480),
		RingWinnerBytes: ringWinnerBytes[:480],
		LedgerLength:    uint64(prefixLen),
	})
	if err != nil {
		return nil, fmt.Errorf("structure digest: %w", err)
	}
	copy(winner.StructureDigest[:], sd[:])

	// Slot 0 holds the unique-highest winner; slots 1..6 hold prior records
	// with distinct lower sequences (a plausible single-writer commit history).
	// Encode self-seals each slot's record digest.
	var ring [container.CommitRingSlots]container.CommitRingRecord
	ring[0] = winner
	for i := 1; i < container.CommitRingSlots; i++ {
		prior := winner
		prior.Sequence = uint64(i) // 1..6, all below the winner's 7
		ring[i] = prior
	}

	img := make([]byte, 0, prefixLen)
	img = append(img, headerBytes...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	img = append(img, fmBytes...)
	img = append(img, stBytes...)
	if len(img) != prefixLen {
		return nil, fmt.Errorf("assembled prefix %d octets, want %d", len(img), prefixLen)
	}
	return img, nil
}

func padTo(b []byte, n int) []byte {
	if len(b) >= n {
		return b[:n]
	}
	out := make([]byte, n)
	copy(out, b)
	return out
}
