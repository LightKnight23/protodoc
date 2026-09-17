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
