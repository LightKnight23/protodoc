package integrity

import (
	"testing"
	"time"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_070_FullEvidenceChainVerdictStableAcrossTime is T-0175's named
// integration test (FR-070). A signature's complete LTV evidence chain --
// credential chain, revocation evidence, and time attestation with its own
// nested chain and revocation -- verifies OFFLINE, and the verdict is STABLE
// across time: because verification uses only the attested time and the fixed
// local anchors (never the wall clock, never the network), verifying the same
// evidence yields the same verdict no matter when it is run.
func TestFR_070_FullEvidenceChainVerdictStableAcrossTime(t *testing.T) {
	nb := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	na := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	attested := SigningInstant(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC).Unix())

	// Signatory chain.
	sRootDER, sRoot, sRootKey := buildCert(t, "Signer Root", true, nil, nil, nb, na)
	sLeafDER, _, _ := buildCert(t, "Signer Leaf", false, sRoot, sRootKey, nb, na)
	sChain := append(append([]byte(nil), sLeafDER...), sRootDER...)

	// TSA chain.
	tRootDER, tRoot, tRootKey := buildCert(t, "TSA Root", true, nil, nil, nb, na)
	tLeafDER, _, _ := buildCert(t, "TSA Leaf", false, tRoot, tRootKey, nb, na)
	tChain := append(append([]byte(nil), tLeafDER...), tRootDER...)

	credID := aeFixtureID(0x10)
	revID := aeFixtureID(0x20)
	timeID := aeFixtureID(0x30)
	tsaCredID := aeFixtureID(0x40)
	tsaRevID := aeFixtureID(0x50)

	evidence := map[pdlfmt.UnitID]AttestationEvidence{
		credID:    {ID: credID, Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: sChain},
		revID:     {ID: revID, Kind: AeRevocationEvidence, Format: AeFormatOCSP, DEROctets: []byte{0x30, 0x00}},
		tsaCredID: {ID: tsaCredID, Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: tChain},
		tsaRevID:  {ID: tsaRevID, Kind: AeRevocationEvidence, Format: AeFormatCRL, DEROctets: []byte{0x30, 0x00}},
		timeID: {ID: timeID, Kind: AeTimeAttestation, Format: AeFormatTimestamp, DEROctets: []byte{0x30, 0x03, 0x02, 0x01, 0x00},
			NestedCredChain: tsaCredID, NestedRevocation: tsaRevID},
	}
	resolve := func(ref pdlfmt.UnitID) (AttestationEvidence, bool) {
		ae, ok := evidence[ref]
		return ae, ok
	}
	sig := SignatureRecord{CredChainRef: credID, RevocationRef: revID, TimeAttestationRef: timeID}

	anchors, _ := NewTrustAnchors([][]byte{sRootDER, tRootDER})
	opt := OfflineVerifyOptions{Anchors: anchors, CurrentTime: attested}

	// Verify the same evidence many times; the verdict must be identical every
	// time -- it depends only on the attested time and fixed anchors, not on
	// the wall clock. (We call repeatedly at different real moments.)
	first, err := VerifyFullEvidenceChainOffline(sig, opt, resolve)
	if err != nil {
		t.Fatalf("VerifyFullEvidenceChainOffline: %v", err)
	}
	if !first.AllTrusted {
		t.Fatalf("full evidence chain did not verify: %+v", first)
	}
	for i := 0; i < 25; i++ {
		got, err := VerifyFullEvidenceChainOffline(sig, opt, resolve)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if got != first {
			t.Fatalf("iteration %d: verdict %+v differs from first %+v (not stable across time)", i, got, first)
		}
	}

	// Verifying at a DIFFERENT attested time still within cert validity is
	// also all-trusted, and equally stable -- the verdict is a function of the
	// attested time, deterministically.
	optLater := opt
	optLater.CurrentTime = SigningInstant(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
	later, err := VerifyFullEvidenceChainOffline(sig, optLater, resolve)
	if err != nil {
		t.Fatalf("later verify: %v", err)
	}
	if !later.AllTrusted {
		t.Errorf("full evidence chain did not verify at the later attested time")
	}
}
