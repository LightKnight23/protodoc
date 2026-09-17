package integrity

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_070_SignatureLtvRefsRequireCorrectKind is T-0169's named unit test
// (FR-070). Each of a signature's LTV references must resolve to
// ATTESTATION_EVIDENCE of the required ae-kind: cred-chain->credential-chain,
// time-attestation->time-attestation, revocation->revocation-evidence (when
// present). A ref to the wrong kind, or a mandatory ref that does not resolve,
// is rejected.
func TestFR_070_SignatureLtvRefsRequireCorrectKind(t *testing.T) {
	credID := aeFixtureID(0x10)
	revID := aeFixtureID(0x20)
	timeID := aeFixtureID(0x30)
	credNested := aeFixtureID(0x40)
	revNested := aeFixtureID(0x50)

	evidence := map[pdlfmt.UnitID]AttestationEvidence{
		credID:     {ID: credID, Kind: AeCredentialChain, Format: AeFormatX509Chain},
		revID:      {ID: revID, Kind: AeRevocationEvidence, Format: AeFormatOCSP},
		timeID:     {ID: timeID, Kind: AeTimeAttestation, Format: AeFormatTimestamp, NestedCredChain: credNested, NestedRevocation: revNested},
		credNested: {ID: credNested, Kind: AeCredentialChain, Format: AeFormatX509Chain},
		revNested:  {ID: revNested, Kind: AeRevocationEvidence, Format: AeFormatCRL},
	}
	resolve := func(ref pdlfmt.UnitID) (AttestationEvidence, bool) {
		ae, ok := evidence[ref]
		return ae, ok
	}

	good := SignatureRecord{
		CredChainRef: credID, RevocationRef: revID, TimeAttestationRef: timeID,
	}
	if err := CheckSignatureLtvRefs(good, resolve); err != nil {
		t.Fatalf("correct LTV refs rejected: %v", err)
	}

	// zero16 revocation ref (none current at signing) is allowed and skipped.
	noRev := good
	noRev.RevocationRef = pdlfmt.UnitID{}
	if err := CheckSignatureLtvRefs(noRev, resolve); err != nil {
		t.Errorf("zero16 revocation ref rejected: %v", err)
	}

	// cred-chain ref pointing at revocation-evidence -> wrong kind.
	badCred := good
	badCred.CredChainRef = revID
	if err := CheckSignatureLtvRefs(badCred, resolve); !errors.Is(err, ErrLtvRefWrongKind) {
		t.Errorf("cred-chain ref to revocation evidence: err = %v, want ErrLtvRefWrongKind", err)
	}

	// time-attestation ref pointing at credential-chain -> wrong kind.
	badTime := good
	badTime.TimeAttestationRef = credID
	if err := CheckSignatureLtvRefs(badTime, resolve); !errors.Is(err, ErrLtvRefWrongKind) {
		t.Errorf("time ref to credential chain: err = %v, want ErrLtvRefWrongKind", err)
	}

	// revocation ref pointing at time-attestation -> wrong kind.
	badRev := good
	badRev.RevocationRef = timeID
	if err := CheckSignatureLtvRefs(badRev, resolve); !errors.Is(err, ErrLtvRefWrongKind) {
		t.Errorf("revocation ref to time attestation: err = %v, want ErrLtvRefWrongKind", err)
	}

	// A mandatory ref that does not resolve is rejected.
	badResolve := good
	badResolve.CredChainRef = aeFixtureID(0x99) // not in the map
	if err := CheckSignatureLtvRefs(badResolve, resolve); !errors.Is(err, ErrLtvRefUnresolved) {
		t.Errorf("unresolved cred-chain ref: err = %v, want ErrLtvRefUnresolved", err)
	}
}
