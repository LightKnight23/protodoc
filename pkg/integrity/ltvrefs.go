// Signature LTV-reference kind checking (T-0169, FR-070; integrity.abnf S5).
// A SIGNATURE's long-term-validation references each name an
// ATTESTATION_EVIDENCE frame of a SPECIFIC ae-kind: sig-cred-chain-ref must
// resolve to a credential-chain (kind 0), sig-revocation-ref (when non-zero16)
// to revocation-evidence (kind 1), and sig-time-attestation-ref to a
// time-attestation (kind 2). A ref pointing at evidence of the wrong kind, or
// a mandatory ref that does not resolve, is rejected before verification.
package integrity

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// EvidenceResolver returns the ATTESTATION_EVIDENCE addressed by a unit-id, and
// ok=false if no such evidence frame is present. The caller supplies this from
// the current document's ATTEST segments.
type EvidenceResolver func(ref pdlfmt.UnitID) (AttestationEvidence, bool)

var (
	// ErrLtvRefWrongKind is returned when an LTV ref resolves to evidence of
	// the wrong ae-kind.
	ErrLtvRefWrongKind = errors.New("integrity: signature LTV reference resolves to ATTESTATION_EVIDENCE of the wrong ae-kind")
	// ErrLtvRefUnresolved is returned when a mandatory LTV ref does not resolve.
	ErrLtvRefUnresolved = errors.New("integrity: signature LTV reference does not resolve to a present ATTESTATION_EVIDENCE")
)

// CheckSignatureLtvRefs verifies each of a signature's LTV references resolves
// to ATTESTATION_EVIDENCE of the required ae-kind (FR-070):
//
//   - sig-cred-chain-ref (mandatory, non-zero16) -> ae-kind credential-chain
//   - sig-time-attestation-ref (mandatory, non-zero16) -> ae-kind time-attestation
//   - sig-revocation-ref (optional; zero16 when none was current) -> ae-kind
//     revocation-evidence when present
//
// The non-zero16 constraint on cred-chain and time-attestation is enforced by
// the SIGNATURE decoder (ErrSignatureZeroRef); this check adds the kind match.
func CheckSignatureLtvRefs(sig SignatureRecord, resolve EvidenceResolver) error {
	zero := pdlfmt.UnitID{}

	credEv, ok := resolve(sig.CredChainRef)
	if !ok {
		return fmt.Errorf("%w: sig-cred-chain-ref", ErrLtvRefUnresolved)
	}
	if credEv.Kind != AeCredentialChain {
		return fmt.Errorf("%w: sig-cred-chain-ref -> %v, want credential-chain", ErrLtvRefWrongKind, credEv.Kind)
	}

	timeEv, ok := resolve(sig.TimeAttestationRef)
	if !ok {
		return fmt.Errorf("%w: sig-time-attestation-ref", ErrLtvRefUnresolved)
	}
	if timeEv.Kind != AeTimeAttestation {
		return fmt.Errorf("%w: sig-time-attestation-ref -> %v, want time-attestation", ErrLtvRefWrongKind, timeEv.Kind)
	}

	if sig.RevocationRef != zero {
		revEv, ok := resolve(sig.RevocationRef)
		if !ok {
			return fmt.Errorf("%w: sig-revocation-ref", ErrLtvRefUnresolved)
		}
		if revEv.Kind != AeRevocationEvidence {
			return fmt.Errorf("%w: sig-revocation-ref -> %v, want revocation-evidence", ErrLtvRefWrongKind, revEv.Kind)
		}
	}
	return nil
}
