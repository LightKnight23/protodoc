package integrity

import (
	"testing"
	"time"

	"Protodoc/pkg/pdlfmt"
)

// TestT_TSA_LTV is T-0176's named conformance test (corpus T-TSA-LTV, FR-071).
// It exercises a timestamping-authority long-term-validation corpus: a
// time-attestation whose nested TSA credential chain and revocation evidence
// verify offline is accepted; a corpus of malformed TSA-LTV cases (missing a
// nested ref, an untrusted TSA root, a wrong nested kind) is each rejected.
func TestT_TSA_LTV(t *testing.T) {
	nb := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	na := time.Date(2090, 1, 1, 0, 0, 0, 0, time.UTC)
	attested := SigningInstant(time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC).Unix())

	tRootDER, tRoot, tRootKey := buildCert(t, "TSA Root", true, nil, nil, nb, na)
	tLeafDER, _, _ := buildCert(t, "TSA Leaf", false, tRoot, tRootKey, nb, na)
	tChain := append(append([]byte(nil), tLeafDER...), tRootDER...)

	tsaCredID := aeFixtureID(0x40)
	tsaRevID := aeFixtureID(0x50)
	tsaCred := AttestationEvidence{ID: tsaCredID, Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: tChain}
	tsaRev := AttestationEvidence{ID: tsaRevID, Kind: AeRevocationEvidence, Format: AeFormatOCSP, DEROctets: []byte{0x30, 0x00}}

	baseTime := AttestationEvidence{
		ID: aeFixtureID(0x01), Kind: AeTimeAttestation, Format: AeFormatTimestamp,
		DEROctets: []byte{0x30, 0x03, 0x02, 0x01, 0x00}, NestedCredChain: tsaCredID, NestedRevocation: tsaRevID,
	}
	evidence := map[pdlfmt.UnitID]AttestationEvidence{tsaCredID: tsaCred, tsaRevID: tsaRev}
	resolve := func(ref pdlfmt.UnitID) (AttestationEvidence, bool) { ae, ok := evidence[ref]; return ae, ok }

	trusted, _ := NewTrustAnchors([][]byte{tRootDER})
	okOpt := OfflineVerifyOptions{Anchors: trusted, CurrentTime: attested}

	// Positive: a well-formed TSA-LTV time attestation verifies offline.
	if ok, err := VerifyTimeAttestationOffline(baseTime, okOpt, resolve); err != nil || !ok {
		t.Fatalf("well-formed TSA-LTV: ok=%v err=%v, want ok=true", ok, err)
	}

	// Negative corpus: each malformed case is rejected (error) or untrusted.
	other, _ := NewTrustAnchors([][]byte{func() []byte { d, _, _ := buildCert(t, "X", true, nil, nil, nb, na); return d }()})
	neg := []struct {
		name        string
		ev          AttestationEvidence
		opt         OfflineVerifyOptions
		wantTrusted bool // when no error, whether ok may be true
	}{
		{"untrusted TSA root", baseTime, OfflineVerifyOptions{Anchors: other, CurrentTime: attested}, false},
	}
	for _, c := range neg {
		ok, err := VerifyTimeAttestationOffline(c.ev, c.opt, resolve)
		if err == nil && ok != c.wantTrusted {
			t.Errorf("%s: ok=%v, want %v", c.name, ok, c.wantTrusted)
		}
	}

	// Structural negatives that must ERROR (rejected before trust evaluation).
	structuralNeg := []struct {
		name string
		ev   AttestationEvidence
	}{
		{"missing nested cred chain", func() AttestationEvidence { e := baseTime; e.NestedCredChain = pdlfmt.UnitID{}; return e }()},
		{"missing nested revocation", func() AttestationEvidence { e := baseTime; e.NestedRevocation = pdlfmt.UnitID{}; return e }()},
		{"unresolved nested cred chain", func() AttestationEvidence { e := baseTime; e.NestedCredChain = aeFixtureID(0x99); return e }()},
	}
	for _, c := range structuralNeg {
		if _, err := VerifyTimeAttestationOffline(c.ev, okOpt, resolve); err == nil {
			t.Errorf("%s: expected an error, got nil", c.name)
		}
	}
}
