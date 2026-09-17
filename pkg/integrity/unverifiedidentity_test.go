package integrity

import (
	"crypto/ed25519"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_115_UnverifiedNeverCarriesSignerIdentity is T-0164's named
// integration test (FR-115). Each FR-115 trigger -- signature verification
// fails, a signature names a cryptographic parameter outside the version
// allowlist, or an acted-upon octet lies outside every covered range -- must
// present the document as unverified and MUST NOT display a signer identity, a
// partial-validity indicator, or any positive verification badge.
func TestFR_115_UnverifiedNeverCarriesSignerIdentity(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	var signer SignerIdentity
	copy(signer[:], pub)

	presRef := sigFixtureUnitID(0x30)
	boundPresentation := Digest{0x77}
	resolve := func(ref pdlfmt.UnitID) (Digest, bool) {
		if ref == presRef {
			return boundPresentation, true
		}
		return Digest{}, false
	}
	tcRoot := Digest{0x0A}
	structDigest := Digest{0x0B}
	cov := CoverageDescriptor{Mode: CoverageModeSubset, Covered: []SegmentRange{{Start: 0, End: 4}}, Uncovered: []SegmentRange{{Start: 4, End: 8}}}

	newSig := func() SignatureRecord {
		return SignatureRecord{
			ParamSet:           0x0001,
			Coverage:           cov,
			CredChainRef:       sigFixtureUnitID(0x10),
			TimeAttestationRef: sigFixtureUnitID(0x20),
			PresentationRef:    presRef,
		}
	}

	// Trigger 1: signature verification fails (tampered sig value).
	sig1 := newSig()
	so1, _ := SignedObjectForSignature(sig1, tcRoot, structDigest, resolve)
	sig1.Value = eddsa.Sign(priv, [32]byte(so1))
	sig1.Value[0] ^= 0xFF // corrupt the signature
	rep1, err := VerifySignatureForState(VerifyInput{
		Signature: sig1, SignerKey: signer, CurrentTCRoot: tcRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPresentation, StateReconstructable: true,
	}, resolve)
	if err != nil {
		t.Fatalf("trigger1: %v", err)
	}
	if rep1.Verdict == VerdictValid || rep1.Signer != nil {
		t.Errorf("trigger1 (bad signature): verdict=%v signer=%v; must be unverified with no signer", rep1.Verdict, rep1.Signer)
	}

	// Trigger 2: param set outside the version allowlist.
	sig2 := newSig()
	sig2.ParamSet = 0xFFFF // not on the allowlist
	so2, _ := SignedObjectForSignature(sig2, tcRoot, structDigest, resolve)
	sig2.Value = eddsa.Sign(priv, [32]byte(so2)) // a valid Ed25519 sig, but param set is disallowed
	rep2, err := VerifySignatureForState(VerifyInput{
		Signature: sig2, SignerKey: signer, CurrentTCRoot: tcRoot, CurrentStructure: structDigest,
		ShownPresentation: boundPresentation, StateReconstructable: true,
	}, resolve)
	if err != nil {
		t.Fatalf("trigger2: %v", err)
	}
	if rep2.Verdict == VerdictValid || rep2.Signer != nil {
		t.Errorf("trigger2 (disallowed param set): verdict=%v signer=%v; must be unverified with no signer", rep2.Verdict, rep2.Signer)
	}

	// Trigger 3: an acted-upon octet lies outside every covered range.
	actedUpon := []uint64{0, 1, 5} // 5 is in the uncovered range [4,8), outside coverage
	if ActedUponWithinCoverage(cov, actedUpon) {
		t.Errorf("acted-upon set with an out-of-coverage octet reported as within coverage")
	}
	// And an entirely-covered acted-upon set is within coverage.
	if !ActedUponWithinCoverage(cov, []uint64{0, 1, 2, 3}) {
		t.Errorf("fully-covered acted-upon set reported as outside coverage")
	}

	// The Unverified verdict itself never carries a signer identity.
	if VerdictUnverified.CarriesSignerIdentity() {
		t.Errorf("Unverified must never be an identity-carrying verdict (FR-115)")
	}
}
