package integrity

import (
	"bytes"
	"reflect"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestATTEST_EVID_ROUNDTRIP is T-0168's named conformance test (corpus
// ATTEST-EVID-ROUNDTRIP, FR-070). A corpus of ATTESTATION_EVIDENCE records --
// one per ae-kind, with the correct format and nested-ref shape -- must
// round-trip byte-exact (encode -> decode -> re-encode identical) with every
// field surviving and the opaque DER octets preserved verbatim.
func TestATTEST_EVID_ROUNDTRIP(t *testing.T) {
	credRef := aeFixtureID(0x10)
	revRef := aeFixtureID(0x20)

	corpus := []struct {
		name string
		ae   AttestationEvidence
	}{
		{
			name: "credential-chain",
			ae: AttestationEvidence{
				ID: aeFixtureID(0x01), Kind: AeCredentialChain, Format: AeFormatX509Chain,
				DEROctets: []byte{0x30, 0x82, 0x01, 0x0A, 0xDE, 0xAD}, // opaque DER-ish bytes
			},
		},
		{
			name: "revocation-evidence (OCSP)",
			ae: AttestationEvidence{
				ID: aeFixtureID(0x02), Kind: AeRevocationEvidence, Format: AeFormatOCSP,
				DEROctets: []byte{0x30, 0x03, 0x0A, 0x01, 0x00},
			},
		},
		{
			name: "revocation-evidence (CRL)",
			ae: AttestationEvidence{
				ID: aeFixtureID(0x03), Kind: AeRevocationEvidence, Format: AeFormatCRL,
				DEROctets: []byte{0x30, 0x00},
			},
		},
		{
			name: "time-attestation (with mandatory nested refs)",
			ae: AttestationEvidence{
				ID: aeFixtureID(0x04), Kind: AeTimeAttestation, Format: AeFormatTimestamp,
				DEROctets: []byte{0x30, 0x82, 0x02, 0x00}, NestedCredChain: credRef, NestedRevocation: revRef,
			},
		},
		{
			name: "empty DER octets",
			ae: AttestationEvidence{
				ID: aeFixtureID(0x05), Kind: AeCredentialChain, Format: AeFormatX509Chain,
			},
		},
	}

	seenKind := map[AeKind]bool{}
	for _, c := range corpus {
		enc, err := c.ae.Encode()
		if err != nil {
			t.Fatalf("%s: Encode: %v", c.name, err)
		}
		dec, err := DecodeAttestationEvidence(enc)
		if err != nil {
			t.Fatalf("%s: Decode: %v", c.name, err)
		}
		if !reflect.DeepEqual(dec, c.ae) {
			t.Errorf("%s: decoded record differs from original\n got %+v\nwant %+v", c.name, dec, c.ae)
		}
		reEnc, err := dec.Encode()
		if err != nil {
			t.Fatalf("%s: re-Encode: %v", c.name, err)
		}
		if !bytes.Equal(enc, reEnc) {
			t.Errorf("%s: round trip not byte-exact", c.name)
		}
		if !bytes.Equal(dec.DEROctets, c.ae.DEROctets) && !(len(dec.DEROctets) == 0 && len(c.ae.DEROctets) == 0) {
			t.Errorf("%s: opaque DER octets not preserved", c.name)
		}
		seenKind[c.ae.Kind] = true
	}

	for _, k := range []AeKind{AeCredentialChain, AeRevocationEvidence, AeTimeAttestation} {
		if !seenKind[k] {
			t.Errorf("corpus does not cover ae-kind %v", k)
		}
	}
}

func aeFixtureID(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}
