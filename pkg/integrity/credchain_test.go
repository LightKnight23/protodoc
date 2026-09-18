package integrity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// buildCert creates a certificate signed by parent (or self-signed if parent
// is nil), returning the DER and the parsed cert plus its key.
func buildCert(t *testing.T, cn string, isCA bool, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, notBefore, notAfter time.Time) ([]byte, *x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano() + int64(len(cn))),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	signer := tmpl
	skey := key
	if parent != nil {
		signer = parent
		skey = parentKey
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signer, &key.PublicKey, skey)
	if err != nil {
		t.Fatalf("CreateCertificate(%s): %v", cn, err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("ParseCertificate(%s): %v", cn, err)
	}
	return der, cert, key
}

// TestFR_070_CredentialChainVerifiesOffline is T-0172's named unit test
// (FR-070). A credential-chain ATTESTATION_EVIDENCE (ae-kind=0, X.509 DER,
// concatenated leaf-to-root) verifies OFFLINE against a fixed local
// trust-anchor list at the attested time, and fails against an anchor list
// that does not contain the chain's root -- with no network access.
func TestFR_070_CredentialChainVerifiesOffline(t *testing.T) {
	notBefore := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)
	attested := SigningInstant(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix())

	// Root CA (self-signed) + leaf signed by the root.
	rootDER, rootCert, rootKey := buildCert(t, "Test Root CA", true, nil, nil, notBefore, notAfter)
	leafDER, _, _ := buildCert(t, "Test Leaf", false, rootCert, rootKey, notBefore, notAfter)

	// ae-der-octets = leaf || root (leaf-to-root concatenation).
	chainDER := append(append([]byte(nil), leafDER...), rootDER...)
	ev := AttestationEvidence{
		ID: aeFixtureID(0x01), Kind: AeCredentialChain, Format: AeFormatX509Chain, DEROctets: chainDER,
	}

	// (1) Verifies against the correct local anchor.
	trusted, err := NewTrustAnchors([][]byte{rootDER})
	if err != nil {
		t.Fatalf("NewTrustAnchors: %v", err)
	}
	ok, err := VerifyCredentialChainOffline(ev, OfflineVerifyOptions{Anchors: trusted, CurrentTime: attested})
	if err != nil {
		t.Fatalf("VerifyCredentialChainOffline: %v", err)
	}
	if !ok {
		t.Errorf("credential chain did not verify against its own root anchor")
	}

	// (2) Fails against an anchor list without the chain's root.
	otherRootDER, _, _ := buildCert(t, "Other Root", true, nil, nil, notBefore, notAfter)
	untrusted, _ := NewTrustAnchors([][]byte{otherRootDER})
	ok, err = VerifyCredentialChainOffline(ev, OfflineVerifyOptions{Anchors: untrusted, CurrentTime: attested})
	if err != nil {
		t.Fatalf("VerifyCredentialChainOffline (untrusted): %v", err)
	}
	if ok {
		t.Errorf("credential chain verified against an anchor list without its root")
	}

	// (3) An empty pool trusts nothing.
	if _, err := VerifyCredentialChainOffline(ev, OfflineVerifyOptions{Anchors: TrustAnchors{}, CurrentTime: attested}); err == nil {
		t.Errorf("verification with no anchors should error")
	}

	// (4) Wrong evidence kind/format is rejected.
	badEv := ev
	badEv.Kind = AeRevocationEvidence
	if _, err := VerifyCredentialChainOffline(badEv, OfflineVerifyOptions{Anchors: trusted, CurrentTime: attested}); err == nil {
		t.Errorf("non-credential-chain evidence should be rejected")
	}
}
