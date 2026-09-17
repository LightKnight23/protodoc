// ATTESTATION_EVIDENCE (T-0167, FR-070/FR-071; integrity.abnf S9): long-term
// validation material carried INSIDE the document -- a credential chain,
// revocation evidence, or a time attestation -- so a signature stays verifiable
// offline decades hence, when none of the cited authorities will answer a live
// query. A PDL-TLV record (discriminant 0x42) legal only in an ATTEST segment.
// The (ae-kind, ae-format) pairing is closed: any other pairing is rejected.
// This file is the single canonical ATTESTATION_EVIDENCE wire implementation.
package integrity

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ATTESTATION_EVIDENCE discriminant + TLV tags (integrity.abnf S9).
const (
	attestationEvidenceDiscriminant = 0x42
	aeTagDiscriminant               = 0
	aeTagID                         = 1 // ae-id, unit-id
	aeTagKind                       = 2 // ae-kind, u8
	aeTagFormat                     = 3 // ae-format, u8
	aeTagDEROctets                  = 4 // ae-der-octets, opaque []byte
	aeTagNestedCredChain            = 5 // ae-nested-cred-chain, unit-id or zero16
	aeTagNestedRevocation           = 6 // ae-nested-revocation, unit-id or zero16
)

// AeKind is ae-kind (integrity.abnf S9): the closed three-value evidence kind.
type AeKind uint8

const (
	AeCredentialChain    AeKind = 0x00
	AeRevocationEvidence AeKind = 0x01
	AeTimeAttestation    AeKind = 0x02
)

func (k AeKind) String() string {
	switch k {
	case AeCredentialChain:
		return "credential-chain"
	case AeRevocationEvidence:
		return "revocation-evidence"
	case AeTimeAttestation:
		return "time-attestation"
	default:
		return fmt.Sprintf("reserved(0x%02x)", uint8(k))
	}
}

// AeFormat is ae-format (integrity.abnf S9): the external standard the DER
// octets conform to, constrained by ae-kind.
type AeFormat uint8

const (
	AeFormatX509Chain AeFormat = 0x00 // X.509 v3 chain (RFC 5280), kind=credential-chain
	AeFormatOCSP      AeFormat = 0x01 // OCSP response (RFC 6960), kind=revocation-evidence
	AeFormatCRL       AeFormat = 0x02 // CRL (RFC 5280), kind=revocation-evidence
	AeFormatTimestamp AeFormat = 0x03 // TimeStampToken (RFC 3161), kind=time-attestation
)

// AttestationEvidence is the decoded ATTESTATION_EVIDENCE record.
type AttestationEvidence struct {
	ID               pdlfmt.UnitID
	Kind             AeKind
	Format           AeFormat
	DEROctets        []byte        // opaque; never parsed as a wire structure by this package
	NestedCredChain  pdlfmt.UnitID // -> ae-kind=0 evidence, or zero16
	NestedRevocation pdlfmt.UnitID // -> ae-kind=1 evidence, or zero16
}

var (
	// ErrAeDiscriminant is returned for a wrong ae-discriminant.
	ErrAeDiscriminant = errors.New("integrity: record discriminant is not ATTESTATION_EVIDENCE (0x42)")
	// ErrAeKindReserved is returned for an ae-kind outside {0,1,2}.
	ErrAeKindReserved = errors.New("integrity: ae-kind outside the closed set {credential-chain, revocation-evidence, time-attestation}")
	// ErrAeKindFormatPairing is returned for an invalid (ae-kind, ae-format) pairing.
	ErrAeKindFormatPairing = errors.New("integrity: invalid (ae-kind, ae-format) pairing")
	// ErrAeNestedRule is returned when a nested reference violates FR-071's
	// mandatory/forbidden rule for the record's kind.
	ErrAeNestedRule = errors.New("integrity: ae nested-reference rule violated for this ae-kind (FR-071)")
	// ErrAeMissingField is returned for a missing required field.
	ErrAeMissingField = errors.New("integrity: ATTESTATION_EVIDENCE missing a required field")
)

// ValidateKindFormatPairing enforces the closed (ae-kind, ae-format) pairing
// (integrity.abnf S9): credential-chain<->X509Chain, revocation-evidence<->
// {OCSP,CRL}, time-attestation<->Timestamp. Any other pairing is rejected.
func ValidateKindFormatPairing(kind AeKind, format AeFormat) error {
	switch kind {
	case AeCredentialChain:
		if format != AeFormatX509Chain {
			return fmt.Errorf("%w: credential-chain requires X.509 chain (0x00), got 0x%02x", ErrAeKindFormatPairing, uint8(format))
		}
	case AeRevocationEvidence:
		if format != AeFormatOCSP && format != AeFormatCRL {
			return fmt.Errorf("%w: revocation-evidence requires OCSP (0x01) or CRL (0x02), got 0x%02x", ErrAeKindFormatPairing, uint8(format))
		}
	case AeTimeAttestation:
		if format != AeFormatTimestamp {
			return fmt.Errorf("%w: time-attestation requires TimeStampToken (0x03), got 0x%02x", ErrAeKindFormatPairing, uint8(format))
		}
	default:
		return fmt.Errorf("%w: 0x%02x", ErrAeKindReserved, uint8(kind))
	}
	return nil
}

// ValidateNestedRefs enforces FR-071's nested-reference rule: a
// time-attestation MUST carry both a nested credential chain and a nested
// revocation reference (non-zero16); a credential-chain or revocation-evidence
// record MUST carry neither (both zero16 -- no second level of nesting).
func (a AttestationEvidence) ValidateNestedRefs() error {
	zero := pdlfmt.UnitID{}
	hasCred := a.NestedCredChain != zero
	hasRev := a.NestedRevocation != zero
	switch a.Kind {
	case AeTimeAttestation:
		if !hasCred || !hasRev {
			return fmt.Errorf("%w: time-attestation requires both nested credential-chain and revocation refs", ErrAeNestedRule)
		}
	case AeCredentialChain, AeRevocationEvidence:
		if hasCred || hasRev {
			return fmt.Errorf("%w: ae-kind %v must carry no nested refs", ErrAeNestedRule, a.Kind)
		}
	default:
		return fmt.Errorf("%w: 0x%02x", ErrAeKindReserved, uint8(a.Kind))
	}
	return nil
}

// Validate runs the full structural validation: kind in range, kind/format
// pairing, and the nested-reference rule.
func (a AttestationEvidence) Validate() error {
	if err := ValidateKindFormatPairing(a.Kind, a.Format); err != nil {
		return err
	}
	return a.ValidateNestedRefs()
}

// Encode encodes the record to its byte-exact PDL-TLV wire form. It validates
// the kind/format pairing and nested rule (a writer producing an invalid one
// has a bug).
func (a AttestationEvidence) Encode() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	fields := []pdlfmt.Field{
		{Tag: aeTagDiscriminant, Value: []byte{attestationEvidenceDiscriminant}},
		{Tag: aeTagID, Value: pdlfmt.AppendUnitID(nil, a.ID)},
		{Tag: aeTagKind, Value: []byte{byte(a.Kind)}},
		{Tag: aeTagFormat, Value: []byte{byte(a.Format)}},
		{Tag: aeTagDEROctets, Value: append([]byte(nil), a.DEROctets...)},
		{Tag: aeTagNestedCredChain, Value: pdlfmt.AppendUnitID(nil, a.NestedCredChain)},
		{Tag: aeTagNestedRevocation, Value: pdlfmt.AppendUnitID(nil, a.NestedRevocation)},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeAttestationEvidence decodes an ATTESTATION_EVIDENCE record, the
// byte-exact inverse of Encode. It rejects a wrong discriminant, a missing
// required field, a wrong field width, a reserved ae-kind, an invalid
// kind/format pairing, and a nested-reference rule violation. The DER octets
// are kept opaque (never parsed as a wire structure here).
func DecodeAttestationEvidence(src []byte) (AttestationEvidence, error) {
	var a AttestationEvidence
	known := map[byte]bool{
		aeTagDiscriminant: true, aeTagID: true, aeTagKind: true, aeTagFormat: true,
		aeTagDEROctets: true, aeTagNestedCredChain: true, aeTagNestedRevocation: true,
	}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return a, err
	}
	seen := map[byte]bool{}
	for _, f := range fields {
		seen[f.Tag] = true
		switch f.Tag {
		case aeTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != attestationEvidenceDiscriminant {
				return a, ErrAeDiscriminant
			}
		case aeTagID:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return a, fmt.Errorf("integrity: ae-id unit-id: %v", err)
			}
			a.ID = id
		case aeTagKind:
			if len(f.Value) != 1 {
				return a, fmt.Errorf("integrity: ae-kind is %d octets, want 1", len(f.Value))
			}
			a.Kind = AeKind(f.Value[0])
		case aeTagFormat:
			if len(f.Value) != 1 {
				return a, fmt.Errorf("integrity: ae-format is %d octets, want 1", len(f.Value))
			}
			a.Format = AeFormat(f.Value[0])
		case aeTagDEROctets:
			a.DEROctets = append([]byte(nil), f.Value...)
		case aeTagNestedCredChain:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return a, fmt.Errorf("integrity: ae-nested-cred-chain unit-id: %v", err)
			}
			a.NestedCredChain = id
		case aeTagNestedRevocation:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return a, fmt.Errorf("integrity: ae-nested-revocation unit-id: %v", err)
			}
			a.NestedRevocation = id
		}
	}
	for tag := byte(aeTagDiscriminant); tag <= aeTagNestedRevocation; tag++ {
		if !seen[tag] {
			return a, fmt.Errorf("%w: tag %d", ErrAeMissingField, tag)
		}
	}
	if err := a.Validate(); err != nil {
		return a, err
	}
	return a, nil
}
