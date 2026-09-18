package integrity

import (
	"testing"
	"time"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_070_TimeAttestationParsesAndVerifiesOffline is T-0174's named unit
// test (FR-070/FR-071). A time-attestation (ae-kind=2, TimeStampToken)
// verifies OFFLINE together with the attesting authority's OWN nested
// credential chain and revocation evidence: both must be present, resolve, and
// verify offline. No network is used; the RFC 3161 token DER stays opaque.
func TestFR_070_TimeAttestationParsesAndVerifiesOffline(t *testing.T) {
	notBefore := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)
	attested := SigningInstant(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix())

	// The TSA's own root CA + leaf (its credential chain).
	rootDER, rootCert, rootKey := buildCert(t, "TSA Root", true, nil, nil, notBefore, notAfter)
	leafDER, _, _ := buildCert(t, "TSA Leaf", false, rootCert, rootKey, notBefore, notAfter)
	tsaChainDER := append(append([]byte(nil), leafDER...), rootDER...)

	credID := aeFixtureID(0x40)
	revID := aeFixtureID(0x50)
	credEv := AttestationEvidence{ID: credID, Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: tsaChainDER}
	revEv := AttestationEvidence{ID: revID, Kind: AeRevocationEvidence, Format: AeFormatCRL, DEROctets: []byte{0x30, 0x00}}

	timeEv := AttestationEvidence{
		ID: aeFixtureID(0x01), Kind: AeTimeAttestation, Format: AeFormatTimestamp,
		DEROctets:       []byte{0x30, 0x03, 0x02, 0x01, 0x00}, // opaque TimeStampToken-ish SEQUENCE
		NestedCredChain: credID, NestedRevocation: revID,
	}

	evidence := map[pdlfmt.UnitID]AttestationEvidence{credID: credEv, revID: revEv}
	resolve := func(ref pdlfmt.UnitID) (AttestationEvidence, bool) {
		ae, ok := evidence[ref]
		return ae, ok
	}

	trusted, _ := NewTrustAnchors([][]byte{rootDER})
	opt := OfflineVerifyOptions{Anchors: trusted, CurrentTime: attested}

	// (1) Full offline verify succeeds with the TSA root trusted.
	ok, err := VerifyTimeAttestationOffline(timeEv, opt, resolve)
	if err != nil {
		t.Fatalf("VerifyTimeAttestationOffline: %v", err)
	}
	if !ok {
		t.Errorf("time attestation with a trusted nested chain did not verify")
	}

	// (2) With the TSA root NOT in the anchors, the nested chain is untrusted.
	otherRootDER, _, _ := buildCert(t, "Other Root", true, nil, nil, notBefore, notAfter)
	untrusted, _ := NewTrustAnchors([][]byte{otherRootDER})
	ok, err = VerifyTimeAttestationOffline(timeEv, OfflineVerifyOptions{Anchors: untrusted, CurrentTime: attested}, resolve)
	if err != nil {
		t.Fatalf("VerifyTimeAttestationOffline (untrusted): %v", err)
	}
	if ok {
		t.Errorf("time attestation verified with an untrusted TSA chain")
	}

	// (3) An unresolved nested credential chain is rejected.
	broken := timeEv
	broken.NestedCredChain = aeFixtureID(0x99) // not in the map
	if _, err := VerifyTimeAttestationOffline(broken, opt, resolve); err == nil {
		t.Errorf("unresolved nested credential chain was accepted")
	}

	// (4) A non-time-attestation is rejected.
	notTime := credEv
	if _, err := VerifyTimeAttestationOffline(notTime, opt, resolve); err == nil {
		t.Errorf("credential-chain accepted as a time attestation")
	}
}
