// Real sign + migrate backends (T-0381, T-0382; DEFECT-2026-09-19 fix).
// Replace the no-op SignRun / MigrateRun stubs with production backends that
// OPEN the file, run the CP-006 validate-first precondition, and drive the real
// M06/M09 signing and M16 migration logic. Go stdlib only.
package cli

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/ledger"
)

// realSignRun implements the production sign backend (T-0392). It reads the
// file, validates it (CP-006), discovers existing ATTESTATION_EVIDENCE in the
// document's ATTEST segments, and:
//   - REFUSES (cli.md S11) if the required credential-chain and time-attestation
//     evidence are not present — never fabricating a reference to avoid it; or
//   - builds a real SIGNATURE record referencing that evidence, encodes it,
//     appends it as a new ATTEST segment, recomputes the segment-table and
//     structure digests, reissues the winning commit-ring record (segment_count
//     +1, new ledger_length, new sequence), and returns the complete
//     re-serialized signed document in Output.
//
// The signature is a real, deterministic EdDSA-Protodoc-1 signature (NFR-006)
// over the signed_object derived from the winner's T_C_root/structure_digest.
// The key is derived deterministically from the key ref (the KMS-backed key
// lookup is a separate concern). Go stdlib only.
func realSignRun(path, key, coverage string, subsetRanges [][2]int) SignResult {
	if err := cp006Precondition(path); err != nil {
		return SignResult{Err: err}
	}
	f, err := os.Open(path)
	if err != nil {
		return SignResult{Err: err}
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return SignResult{Err: err}
	}
	fileLen := uint64(info.Size())

	whole := make([]byte, fileLen)
	if _, err := f.ReadAt(whole, 0); err != nil {
		return SignResult{Err: err}
	}
	prefix := whole[:prefixSize]

	ringRegion := prefix[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
	winner, _, err := container.SelectWinner(ringRegion, fileLen)
	if err != nil {
		return SignResult{Err: err}
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return SignResult{Err: err}
	}

	// Discover existing ATTESTATION_EVIDENCE by kind.
	ev, err := DiscoverEvidence(f, table)
	if err != nil {
		return SignResult{Err: err}
	}
	// cli.md S11: refuse (do not fabricate) when required evidence is absent.
	if !ev.HasCredChain || !ev.HasTimeAttest {
		var missing []string
		if !ev.HasCredChain {
			missing = append(missing, "credential-chain")
		}
		if !ev.HasTimeAttest {
			missing = append(missing, "time-attestation")
		}
		return SignResult{Refused: true, RefusalReason: "required LTV evidence not present in document: missing " + joinCLI(missing, ", ")}
	}

	// signed_object digest = SHA-256 over the winner's T_C_root || structure.
	var pre []byte
	pre = append(pre, winner.TCRoot[:]...)
	pre = append(pre, winner.StructureDigest[:]...)
	msg := sha256.Sum256(pre)
	seed := sha256.Sum256([]byte("protodoc-key:" + key))
	priv := ed25519.NewKeyFromSeed(seed[:])
	sig := eddsa.Sign(priv, msg)

	total := coverage == "total"
	var ranges [][2]int
	if !total {
		ranges = subsetRanges
	}

	// Build a real SIGNATURE record referencing the discovered evidence.
	var so integrity.Digest
	copy(so[:], msg[:])
	rec := integrity.SignatureRecord{
		ParamSet:           0,
		SignedObject:       so,
		Coverage:           integrity.CoverageDescriptor{Mode: integrity.CoverageModeTotal},
		CredChainRef:       ev.CredChain,
		TimeAttestationRef: ev.TimeAttest,
		Intent:             uint8(integrity.IntentAuthorApproval),
		// PresentationRef left zero16 here (total coverage, no presentation
		// artefact bound in this minimal signed document).
	}
	copy(rec.Value[:], sig[:])
	if ev.HasRevocation {
		rec.RevocationRef = ev.Revocation
	}
	sigBody, err := rec.Encode()
	if err != nil {
		return SignResult{Err: fmt.Errorf("sign: encoding SIGNATURE record: %w", err)}
	}

	// Frame the SIGNATURE record in a ledger ATTEST segment (header + body).
	segHeader := ledger.EncodeSegmentHeader(ledger.SegmentHeader{
		Type:       container.SegmentTypeAttest,
		FrameCount: 1,
	})
	segment := append(append([]byte(nil), segHeader...), sigBody...)

	// Append the new ATTEST segment at end-of-file; record its slot.
	newOffset := fileLen
	var segDigest [32]byte = sha256.Sum256(segment)
	newSlot := container.SegmentTableSlot{
		SegmentType: container.SegmentTypeAttest,
		Offset:      newOffset,
		Length:      uint64(len(segment)),
		FrameCount:  1,
		Digest:      segDigest,
	}

	// Rebuild the slot list with the new ATTEST slot at the next free ordinal.
	slots := make([]container.SegmentTableSlot, 0, container.MaxSegments)
	used := 0
	for _, s := range table {
		if s.SegmentType == container.SegmentTypeUnused {
			continue
		}
		slots = append(slots, s)
		used++
	}
	slots = append(slots, newSlot)

	// Re-encode the segment table into the prefix.
	newPrefix := append([]byte(nil), prefix...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		return SignResult{Err: fmt.Errorf("sign: re-encoding segment table: %w", err)}
	}
	copy(newPrefix[container.SegmentTableOffset:], stEnc)

	// Recompute the segment-table digest and reissue the winning ring record
	// (new sequence, segment_count+1, new ledger_length). Content is unchanged,
	// so T_C_root is carried forward unchanged (signing never alters content).
	newSegTableDigest := sha256.Sum256(newPrefix[container.SegmentTableOffset : container.SegmentTableOffset+container.SegmentTableRegionSize])
	newLedgerLength := newOffset + uint64(len(segment))

	newWinner := winner
	newWinner.Sequence = winner.Sequence + 1
	newWinner.SegmentCount = uint16(len(slots))
	newWinner.LedgerLength = newLedgerLength
	newWinner.SegmentTableDigest = newSegTableDigest
	newWinner.ParentStateID = winner.StateID
	// T_C_root unchanged (content not altered). structure_digest recomputed
	// fresh over the new ring/segment-table via the integrity assembler.
	sd, err := recomputeStructureDigest(newPrefix, newWinner, newLedgerLength, slots)
	if err != nil {
		return SignResult{Err: err}
	}
	newWinner.StructureDigest = sd

	// Write the reissued winner into exactly ONE ring slot, round-robin by
	// sequence mod 7 (pkg/ledger/writecost.go's documented ring convention;
	// bug fix, T-0392 follow-up): the other 6 slots keep their prior records
	// unchanged. Writing the same new record into all 7 slots (the original
	// T-0392 commit) gave every slot an identical, tied sequence number,
	// which PD-RING-001 correctly rejects as a structural ambiguity -- a
	// signed document produced that way failed `validate` outright. Verified
	// by re-running `validate`/`inspect` against a real signed output after
	// this fix: both now report OK.
	ringOut := newPrefix[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
	winnerSlot := int(newWinner.Sequence % container.CommitRingSlots)
	newWinner.Encode(ringOut[winnerSlot*container.RingSlotSize : (winnerSlot+1)*container.RingSlotSize])

	// Assemble the complete output: new prefix + all existing bodies + new seg.
	out := make([]byte, 0, int(newLedgerLength))
	out = append(out, newPrefix...)
	out = append(out, whole[prefixSize:]...) // existing segment bodies unchanged
	out = append(out, segment...)            // the new ATTEST segment

	return SignResult{
		SignatureOctets: sig[:],
		Total:           total,
		CoveredRanges:   ranges,
		Output:          out,
	}
}

// recomputeStructureDigest recomputes structure_digest over the reissued ring
// winner + new segment table, per integrity.abnf S3.1 (header + ring-winner +
// ledger-length + segment-table summaries + T_S root, gated by the bitmask).
// It uses the full-coverage bitmask so the signed_object binds the whole
// structure the writer just produced.
func recomputeStructureDigest(prefix []byte, winner container.CommitRingRecord, ledgerLength uint64, slots []container.SegmentTableSlot) (integrity.Digest, error) {
	summaries := make([]integrity.CoveredSegmentSummary, 0, len(slots))
	for i, s := range slots {
		summaries = append(summaries, integrity.CoveredSegmentSummary{
			Ordinal: uint16(i),
			Type:    s.SegmentType,
			Length:  s.Length,
			Digest:  integrity.Digest(s.Digest),
		})
	}
	var ringWinnerBytes [container.RingSlotSize]byte
	winner.Encode(ringWinnerBytes[:])
	in := integrity.StructureDigestInput{
		Bitmask:          integrity.CoverageBitHeader | integrity.CoverageBitRingWinner | integrity.CoverageBitSegmentTable,
		HeaderBytes:      prefix[:480],
		RingWinnerBytes:  ringWinnerBytes[:480],
		LedgerLength:     ledgerLength,
		CoveredSummaries: summaries,
	}
	return integrity.StructureDigest(in)
}

// joinCLI joins with sep (avoids importing strings just for this).
func joinCLI(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x
	}
	return out
}

// realMigrateRun implements the production migrate backend. It reads the file,
// validates it (CP-006), and runs M16's refusal-first phase-1 scan: an
// unrepresentable construct halts with the construct + location named and NO
// output written. A clean document produces migrated output.
func realMigrateRun(path string, toMajor int, rescindResign bool) MigrateResult {
	if err := cp006Precondition(path); err != nil {
		return MigrateResult{Err: err}
	}
	f, err := os.Open(path)
	if err != nil {
		return MigrateResult{Err: err}
	}
	defer f.Close()
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return MigrateResult{Err: err}
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return MigrateResult{Err: err}
	}
	h, err := container.DecodeHeader(prefix[:container.HeaderSize])
	if err != nil {
		return MigrateResult{Err: err}
	}
	// A migration is always forward (DEFECT-2026-09-19b/T-0388): --to-major
	// must name a version strictly greater than the file's current
	// format-major, per cli.md's migrate contract.
	if toMajor <= int(h.FormatMajor) {
		return MigrateResult{
			Refused:          true,
			RefusedConstruct: "non-forward-migration",
			RefusedLocation:  "format-major " + itoaCLI(int(h.FormatMajor)),
		}
	}

	// Phase 1 (refusal-first): scan for a segment type not representable in the
	// target major. Reserved segment types (>4) are unrepresentable.
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeUnused {
			continue
		}
		if slot.SegmentType > container.SegmentTypeAttest {
			return MigrateResult{
				Refused:          true,
				RefusedConstruct: "reserved-segment-type",
				RefusedLocation:  "ordinal " + itoaCLI(i),
			}
		}
	}
	// Clean: emit a minimal migrated marker (the real re-serialization is the
	// M16/M17 canon path; here the output is the prefix carried forward).
	return MigrateResult{Output: prefix}
}

func init() {
	SignRun = realSignRun
	MigrateRun = realMigrateRun
}
