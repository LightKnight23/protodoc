// Offline LTV verification (T-0171, FR-070; integrity.abnf S9's LTV clause).
// Verification of the full evidence chain MUST succeed with ALL network
// interfaces disabled and a FIXED, locally-held trust-anchor list -- none of
// the cited authorities is expected to answer a live query decades hence. This
// file holds the trust-anchor list model and the offline-verification entry
// points; it makes no network call, and the T-0171 audit enforces that the
// LTV verification code imports no networking package.
package integrity

import (
	"crypto/x509"
	"errors"
	"fmt"
)

// TrustAnchors is a fixed, locally-held set of trust anchors (root
// certificates) used for offline evidence verification. It is supplied by the
// caller, never fetched; an empty pool verifies nothing (a closed-world,
// no-network posture).
type TrustAnchors struct {
	Roots *x509.CertPool
}

// NewTrustAnchors builds a TrustAnchors from locally-held DER root
// certificates. It never consults the system root store or the network; only
// the roots passed in are trusted.
func NewTrustAnchors(derRoots [][]byte) (TrustAnchors, error) {
	pool := x509.NewCertPool()
	for i, der := range derRoots {
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return TrustAnchors{}, errors.New("integrity: trust anchor " + itoa(i) + " is not a parseable certificate")
		}
		pool.AddCert(cert)
	}
	return TrustAnchors{Roots: pool}, nil
}

// itoa is a tiny stdlib-free int formatter for the error path above (avoids
// pulling in strconv just for one message).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// OfflineVerifyOptions are the x509 verify options pinned for offline LTV
// verification: the caller's fixed roots, and NO current-time fetch, NO OCSP
// fetch, NO CRL fetch, NO system-root fallback. The verify time is the caller-
// supplied attested time, never the wall clock.
type OfflineVerifyOptions struct {
	Anchors     TrustAnchors
	CurrentTime SigningInstant // the time to verify certificate validity at
}

// ErrNoTrustAnchors is returned when an offline verification is attempted with
// an empty trust-anchor pool (closed-world: nothing is trusted by default).
var ErrNoTrustAnchors = errors.New("integrity: offline verification requires a non-empty locally-held trust-anchor list")

// ErrEmptyCredentialChain is returned when a credential-chain's DER octets
// contain no certificate.
var ErrEmptyCredentialChain = errors.New("integrity: credential-chain DER octets contain no certificate")

// parseConcatenatedDERCerts parses a concatenation of DER certificates
// (leaf-to-root, as ae-der-octets carries for a credential-chain, RFC 5280).
// Go's x509.ParseCertificate rejects trailing data, so we frame each
// certificate by its own ASN.1 SEQUENCE length (tag 0x30 + definite length)
// and parse each framed block individually.
func parseConcatenatedDERCerts(der []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	rest := der
	for len(rest) > 0 {
		n, err := derSequenceLen(rest)
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(rest[:n])
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		rest = rest[n:]
	}
	if len(certs) == 0 {
		return nil, ErrEmptyCredentialChain
	}
	return certs, nil
}

// derSequenceLen returns the total octet length (header + content) of the
// leading DER SEQUENCE in b, supporting the definite-length short and long
// forms. It rejects a non-SEQUENCE tag, an indefinite length, and a length
// that runs past b (a framing error, checked before any use).
func derSequenceLen(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, errors.New("integrity: credential-chain DER truncated at TLV header")
	}
	if b[0] != 0x30 { // universal, constructed, SEQUENCE
		return 0, errors.New("integrity: credential-chain element is not a DER SEQUENCE")
	}
	lenByte := b[1]
	if lenByte < 0x80 { // short form
		total := 2 + int(lenByte)
		if total > len(b) {
			return 0, errors.New("integrity: credential-chain DER length runs past input")
		}
		return total, nil
	}
	if lenByte == 0x80 { // indefinite length not allowed in DER
		return 0, errors.New("integrity: credential-chain DER uses indefinite length (not DER)")
	}
	nBytes := int(lenByte & 0x7F)
	if nBytes > 4 || 2+nBytes > len(b) {
		return 0, errors.New("integrity: credential-chain DER long-form length malformed")
	}
	length := 0
	for i := 0; i < nBytes; i++ {
		length = length<<8 | int(b[2+i])
	}
	total := 2 + nBytes + length
	if total < 0 || total > len(b) {
		return 0, errors.New("integrity: credential-chain DER length runs past input")
	}
	return total, nil
}

// VerifyCredentialChainOffline verifies a credential-chain ATTESTATION_EVIDENCE
// (ae-kind=0, ae-format=X.509 chain) OFFLINE against a fixed local trust-anchor
// list at the attested time. It parses the concatenated leaf-to-root DER
// certificates from ae-der-octets, treats every cert after the leaf as an
// intermediate, and verifies the leaf chains to one of the local anchors --
// with no network, no OCSP/CRL fetch, and no system-root fallback. It returns
// whether the chain is trusted, or an error only for a structurally malformed
// chain or an absent anchor pool.
func VerifyCredentialChainOffline(ev AttestationEvidence, opt OfflineVerifyOptions) (bool, error) {
	if ev.Kind != AeCredentialChain || ev.Format != AeFormatX509Chain {
		return false, fmt.Errorf("%w: not a credential-chain/X.509 evidence", ErrAeKindFormatPairing)
	}
	if opt.Anchors.Roots == nil {
		return false, ErrNoTrustAnchors
	}
	certs, err := parseConcatenatedDERCerts(ev.DEROctets)
	if err != nil {
		return false, err
	}
	leaf := certs[0]
	intermediates := x509.NewCertPool()
	for _, c := range certs[1:] {
		intermediates.AddCert(c)
	}
	vopts := x509.VerifyOptions{
		Roots:         opt.Anchors.Roots,
		Intermediates: intermediates,
		CurrentTime:   opt.CurrentTime.Time(),
	}
	if _, err := leaf.Verify(vopts); err != nil {
		return false, nil
	}
	return true, nil
}

// verifyLeafOffline verifies a leaf certificate against the fixed anchors at
// the given time, with the network disabled: it builds x509.VerifyOptions with
// the caller's Roots (never the system pool) and never sets any callback that
// would fetch. It returns whether the leaf chains to a trusted anchor. This is
// the single offline verification primitive; higher-level evidence checks call
// it and never reach for the network.
func verifyLeafOffline(leaf *x509.Certificate, opt OfflineVerifyOptions) (bool, error) {
	if opt.Anchors.Roots == nil {
		return false, ErrNoTrustAnchors
	}
	vopts := x509.VerifyOptions{
		Roots:       opt.Anchors.Roots,
		CurrentTime: opt.CurrentTime.Time(),
		// Intermediates would be supplied from the credential chain's own DER
		// octets by the caller; no network, no system roots.
	}
	if _, err := leaf.Verify(vopts); err != nil {
		return false, nil // untrusted, but not an operational error
	}
	return true, nil
}
