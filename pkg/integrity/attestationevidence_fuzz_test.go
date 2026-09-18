package integrity

import "testing"

// FuzzAttestationEvidenceDecode is T-0182's native fuzz target (feeding
// CP-012's continuous fuzz). It drives arbitrary bytes through
// DecodeAttestationEvidence, asserting the decoder never panics and that any
// accepted record is a stable canonical form (decode -> encode -> decode
// yields an equal record) that also satisfies its own structural validation.
// Malformed inputs are rejected with an error, never a panic.
func FuzzAttestationEvidenceDecode(f *testing.F) {
	// Seed corpus (>=2 seeds so fuzz maturity treats this as a real target).
	seed := func(ae AttestationEvidence) []byte {
		b, err := ae.Encode()
		if err != nil {
			f.Fatalf("seed encode: %v", err)
		}
		return b
	}
	f.Add(seed(AttestationEvidence{ID: aeFixtureID(1), Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: []byte{0x30, 0x00}}))
	f.Add(seed(AttestationEvidence{ID: aeFixtureID(2), Kind: AeTimeAttestation, Format: AeFormatTimestamp, DEROctets: []byte{0x30, 0x00}, NestedCredChain: aeFixtureID(3), NestedRevocation: aeFixtureID(4)}))
	f.Add([]byte{})
	f.Add([]byte{attestationEvidenceDiscriminant})

	f.Fuzz(func(t *testing.T, data []byte) {
		ae, err := DecodeAttestationEvidence(data) // must never panic
		if err != nil {
			return
		}
		// An accepted record satisfies its own validation.
		if verr := ae.Validate(); verr != nil {
			t.Fatalf("accepted record fails Validate: %v", verr)
		}
		// Stable canonical form.
		enc, err := ae.Encode()
		if err != nil {
			t.Fatalf("accepted record failed to re-encode: %v", err)
		}
		ae2, err := DecodeAttestationEvidence(enc)
		if err != nil {
			t.Fatalf("re-encoded record failed to decode: %v", err)
		}
		if ae2.ID != ae.ID || ae2.Kind != ae.Kind || ae2.Format != ae.Format ||
			ae2.NestedCredChain != ae.NestedCredChain || ae2.NestedRevocation != ae.NestedRevocation ||
			string(ae2.DEROctets) != string(ae.DEROctets) {
			t.Fatalf("decode/encode/decode not stable")
		}
	})
}
