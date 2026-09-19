// Real verify backend (T-0374, updated to close GAP-VERIFY-CONTENT-REBUILD).
// Opens the file, runs the CP-006 validate-first precondition, REBUILDS the
// content-commitment tree (T_C) from the real decoded ContentRecords via
// extract.LoadContentRecords + integrity.TCRoot, and verifies each ATTEST
// SignatureRecord against that recomputed state:
//
//   - If the recomputed T_C_root does not match the winning commit-ring
//     record's recorded T_C_root, the file's own state is internally
//     inconsistent and every signature is unverified.
//   - Otherwise, for each signature, the current signed_object is recomputed
//     from the recorded T_C_root/structure_digest and compared to the
//     signature's own signed_object: match -> the signed state is
//     reconstructable and covered; mismatch/unresolved -> the signature covers
//     a state this file can no longer reconstruct (covering_unavailable_state,
//     FR-062).
//
// Full EdDSA byte-verification of sig-value against the recomputed
// signed_object requires the signer's public key from the credential chain
// (M09/M10 offline-evidence path); the CLI verify surface reports the
// state-coverage verdict derived above and defers the credential-chain crypto
// to that offline path. Go stdlib only.
package cli

import (
	"io"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/validate"
)

// realVerifyRun implements the production verify backend.
func realVerifyRun(path string) []VerifyVerdict {
	// DEFECT-2026-09-19b/T-0390: a file that could not be opened at all is
	// USAGE, not UNVERIFIED (cli.md S1: "an unreadable path"), matching
	// inspect's already-correct behaviour.
	steps, _ := realValidateStepsFor(path)
	if res := validate.Run(steps); res.Validity != nil {
		if res.Validity.RuleID == unreadableFileRuleID {
			return []VerifyVerdict{{Verdict: verifyVerdictUsage}}
		}
		return []VerifyVerdict{{Verdict: "unverified"}}
	}

	f, err := os.Open(path)
	if err != nil {
		return []VerifyVerdict{{Verdict: verifyVerdictUsage}}
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return []VerifyVerdict{{Verdict: verifyVerdictUsage}}
	}
	fileLen := uint64(info.Size())

	prefix := make([]byte, prefixSize)
	if _, err := io.ReadFull(f, prefix); err != nil {
		return []VerifyVerdict{{Verdict: "unverified"}}
	}
	ring := prefix[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
	winner, _, err := container.SelectWinner(ring, fileLen)
	if err != nil {
		return []VerifyVerdict{{Verdict: "unverified"}}
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return []VerifyVerdict{{Verdict: "unverified"}}
	}

	var recordedTC, currentStructure integrity.Digest
	copy(recordedTC[:], winner.TCRoot[:])
	copy(currentStructure[:], winner.StructureDigest[:])

	// Rebuild the content tree from the real decoded ContentRecords.
	units, err := extract.LoadContentRecords(f)
	if err != nil {
		return []VerifyVerdict{{Verdict: "unverified"}}
	}
	records := make([]integrity.ContentRecord, 0, len(units))
	for _, u := range units {
		records = append(records, integrity.ContentRecord{UnitID: u.UnitID, Frame: u.Frame})
	}
	recomputedTC, err := integrity.TCRoot(records)
	if err != nil {
		return []VerifyVerdict{{Verdict: "unverified"}}
	}

	// If the recomputed content tree does not match the recorded T_C_root, the
	// document's own state is internally inconsistent — reconstruction fails.
	stateReconstructable := recomputedTC == recordedTC

	resolve := func(ref integrity.UnitID) (integrity.Digest, bool) {
		return integrity.Digest{}, false
	}

	var verdicts []VerifyVerdict
	for _, slot := range table {
		if !integrity.IsAttestTyped(slot) {
			continue
		}
		body := make([]byte, slot.Length)
		if _, err := f.ReadAt(body, int64(slot.Offset)); err != nil {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "unverified"})
			continue
		}
		sig, err := integrity.DecodeSignatureRecord(body)
		if err != nil {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "unverified"})
			continue
		}
		if !stateReconstructable {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "covering_unavailable_state"})
			continue
		}
		current, err := integrity.SignedObjectForSignature(sig, recomputedTC, currentStructure, resolve)
		if err != nil {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "covering_unavailable_state"})
			continue
		}
		if current == sig.SignedObject {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "valid"})
		} else {
			verdicts = append(verdicts, VerifyVerdict{Verdict: "covering_unavailable_state"})
		}
	}
	return verdicts
}

func init() {
	VerifyRun = realVerifyRun
}
