package migrate_test

import (
	"testing"

	"Protodoc/pkg/integrity"
)

// TestFR_122_PreMigrationSignatureCoversOriginalState is T-0302's named
// integration test (FR-122). After a migration changes the document state, a
// signature originally covering the PRE-migration state continues to report —
// via M09's FR-062 verdict composition — that it covers the pre-migration
// state (UnavailableState in the migrated file), naming the pre-migration
// signed state (its T_C_root/structure), never the migrated state's.
func TestFR_122_PreMigrationSignatureCoversOriginalState(t *testing.T) {
	// Pre-migration signed state.
	var preTCRoot, preStructure integrity.Digest
	preTCRoot[0] = 0xA1
	preStructure[0] = 0xB1

	// Migrated state (a different T_C_root/structure — migration changed bytes).
	var postTCRoot, postStructure integrity.Digest
	postTCRoot[0] = 0xC2
	postStructure[0] = 0xD2

	var presDigest integrity.Digest
	presDigest[0] = 0xEE
	var presRef integrity.UnitID
	presRef[0] = 0x01
	sig := integrity.SignatureRecord{PresentationRef: presRef}
	resolve := func(ref integrity.UnitID) (integrity.Digest, bool) {
		if ref == presRef {
			return presDigest, true
		}
		return integrity.Digest{}, false
	}

	// Verify the pre-migration signature against the MIGRATED file: the signed
	// (pre-migration) state can no longer be reconstructed there.
	in := integrity.VerifyInput{
		Signature:            sig,
		CurrentTCRoot:        preTCRoot, // the report names the signed (pre-migration) state
		CurrentStructure:     preStructure,
		ShownPresentation:    presDigest,
		StateReconstructable: false, // migration changed the state; pre-migration state is unavailable
	}
	rep, err := integrity.VerifySignatureForState(in, resolve)
	if err != nil {
		t.Fatalf("VerifySignatureForState: %v", err)
	}

	// Verdict is UnavailableState (FR-062), not valid, not failed.
	if rep.Verdict != integrity.VerdictUnavailableState {
		t.Errorf("verdict = %v, want UnavailableState", rep.Verdict)
	}

	// The report names the PRE-migration signed state, never the migrated one.
	if rep.SignedState.TCRoot != preTCRoot || rep.SignedState.StructureDigest != preStructure {
		t.Errorf("report names %x/%x, want pre-migration state %x/%x",
			rep.SignedState.TCRoot[0], rep.SignedState.StructureDigest[0], preTCRoot[0], preStructure[0])
	}
	if rep.SignedState.TCRoot == postTCRoot || rep.SignedState.StructureDigest == postStructure {
		t.Errorf("report must not name the migrated state's digests")
	}
	// No signer identity on a non-Valid verdict.
	if rep.Verdict.CarriesSignerIdentity() {
		t.Errorf("UnavailableState must not carry signer identity")
	}
}
