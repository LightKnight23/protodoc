// SIGNATURE record (T-0153, FR-063/FR-064; integrity.abnf S5): a PDL-TLV
// record (discriminant 0x40) carried only in an ATTEST-typed segment. It binds
// a signature over signed_object (S3.2) to its coverage descriptor and the
// attestation-evidence references that make long-term verification possible.
// The presentation_artefact_digest is never stored inline here: sig-value's
// signed_object is computed with the referenced PRESENTATION_ARTEFACT's own
// slot digest. This file is the single canonical SIGNATURE wire implementation.
package integrity

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// SigIntent is sig-intent (integrity.abnf S10.3): the closed signing-intent
// set. A v1 reader MUST reject a value outside 0x00-0x03.
type SigIntent uint8

const (
	IntentAuthorApproval     SigIntent = 0x00 // author signing their own work
	IntentWitnessAttestation SigIntent = 0x01 // witnessed the document's execution
	IntentNotarization       SigIntent = 0x02 // identity verification, not content approval
	IntentCustodialTransfer  SigIntent = 0x03 // custody continuity (RESCIND_RESIGN's new signature)
)

// ErrSigIntentOutOfRange is returned for a sig-intent outside the closed set
// {0x00..0x03}.
var ErrSigIntentOutOfRange = errors.New("integrity: sig-intent outside the closed set {author-approval, witness-attestation, notarization, custodial-transfer}")

// ValidateSigIntent enforces the closed sig-intent set (S10.3): 0x00-0x03 are
// the four defined intents; 0x04-0xFF are reserved and rejected, never
// defaulted (non-negotiable #8: a closed enum rejects rather than defaults).
func ValidateSigIntent(intent SigIntent) error {
	if intent > IntentCustodialTransfer {
		return fmt.Errorf("%w: got 0x%02x", ErrSigIntentOutOfRange, uint8(intent))
	}
	return nil
}

// SIGNATURE discriminant (integrity.abnf S5) and TLV tags.
const (
	signatureDiscriminant = 0x40
	sigTagDiscriminant    = 0 // sig-discriminant, value = 0x40
	sigTagParamSet        = 1 // sig-param-set, u16
	sigTagSignedObject    = 2 // sig-signed-object, digest256
	sigTagValue           = 3 // sig-value, 64 octets R||S
	sigTagCoverage        = 4 // sig-coverage, coverage-descriptor
	sigTagCredChainRef    = 5 // sig-cred-chain-ref, unit-id (nonzero)
	sigTagRevocationRef   = 6 // sig-revocation-ref, unit-id or zero16
	sigTagTimeAttestRef   = 7 // sig-time-attestation-ref, unit-id (nonzero)
	sigTagIntent          = 8 // sig-intent, u8
	sigTagPresentationRef = 9 // sig-presentation-ref, unit-id -> PRESENTATION_ARTEFACT
)

// SigValueLen is the fixed length of sig-value: R (32) || S (32).
const SigValueLen = 64

// MaxSignatures is MAX_SIGNATURES (integrity.abnf S5): the per-document upper
// bound on SIGNATURE frames across all ATTEST segments.
const MaxSignatures = 64

// zeroUnitID is container.abnf S0's zero16 sentinel.
var zeroUnitID = pdlfmt.UnitID{}

// SignatureRecord is the decoded SIGNATURE record (integrity.abnf S5).
type SignatureRecord struct {
	ParamSet           uint16
	SignedObject       Digest
	Value              [SigValueLen]byte // R||S
	Coverage           CoverageDescriptor
	CredChainRef       pdlfmt.UnitID // MUST NOT be zero16
	RevocationRef      pdlfmt.UnitID // unit-id or zero16
	TimeAttestationRef pdlfmt.UnitID // MUST NOT be zero16
	Intent             uint8
	PresentationRef    pdlfmt.UnitID // -> PRESENTATION_ARTEFACT
}

var (
	// ErrSignatureDiscriminant is returned when a record's discriminant is not 0x40.
	ErrSignatureDiscriminant = errors.New("integrity: record discriminant is not SIGNATURE (0x40)")
	// ErrSignatureMissingField is returned when a required SIGNATURE field is absent.
	ErrSignatureMissingField = errors.New("integrity: SIGNATURE record missing a required field")
	// ErrSignatureZeroRef is returned when a must-not-be-zero16 reference is zero16.
	ErrSignatureZeroRef = errors.New("integrity: SIGNATURE reference that MUST NOT be zero16 is zero16")
)

// Encode encodes the SIGNATURE record to its byte-exact PDL-TLV wire form,
// the ten fields in strict ascending tag order (0..9). It enforces the
// non-zero16 rule for sig-cred-chain-ref and sig-time-attestation-ref, since a
// writer producing a zero16 there has a bug.
func (s SignatureRecord) Encode() ([]byte, error) {
	if s.CredChainRef == zeroUnitID {
		return nil, fmt.Errorf("%w: sig-cred-chain-ref", ErrSignatureZeroRef)
	}
	if s.TimeAttestationRef == zeroUnitID {
		return nil, fmt.Errorf("%w: sig-time-attestation-ref", ErrSignatureZeroRef)
	}
	cov, err := s.Coverage.Encode(nil)
	if err != nil {
		return nil, fmt.Errorf("integrity: SIGNATURE coverage: %w", err)
	}
	var ps [2]byte
	binary.BigEndian.PutUint16(ps[:], s.ParamSet)

	fields := []pdlfmt.Field{
		{Tag: sigTagDiscriminant, Value: []byte{signatureDiscriminant}},
		{Tag: sigTagParamSet, Value: ps[:]},
		{Tag: sigTagSignedObject, Value: pdlfmt.AppendDigest256(nil, pdlfmt.Digest256(s.SignedObject))},
		{Tag: sigTagValue, Value: append([]byte(nil), s.Value[:]...)},
		{Tag: sigTagCoverage, Value: cov},
		{Tag: sigTagCredChainRef, Value: pdlfmt.AppendUnitID(nil, s.CredChainRef)},
		{Tag: sigTagRevocationRef, Value: pdlfmt.AppendUnitID(nil, s.RevocationRef)},
		{Tag: sigTagTimeAttestRef, Value: pdlfmt.AppendUnitID(nil, s.TimeAttestationRef)},
		{Tag: sigTagIntent, Value: []byte{s.Intent}},
		{Tag: sigTagPresentationRef, Value: pdlfmt.AppendUnitID(nil, s.PresentationRef)},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeSignatureRecord decodes a SIGNATURE record, the byte-exact inverse of
// Encode. It rejects a wrong discriminant, a missing required field, a wrong
// field width, a zero16 in a must-not-be-zero16 reference, and a malformed
// coverage descriptor (structurally; PD-COVER-001..004 are checked separately
// against a segment count via Coverage.Validate).
func DecodeSignatureRecord(src []byte) (SignatureRecord, error) {
	var s SignatureRecord
	known := map[byte]bool{
		sigTagDiscriminant: true, sigTagParamSet: true, sigTagSignedObject: true,
		sigTagValue: true, sigTagCoverage: true, sigTagCredChainRef: true,
		sigTagRevocationRef: true, sigTagTimeAttestRef: true, sigTagIntent: true,
		sigTagPresentationRef: true,
	}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return s, err
	}
	seen := map[byte]bool{}
	for _, f := range fields {
		seen[f.Tag] = true
		switch f.Tag {
		case sigTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != signatureDiscriminant {
				return s, ErrSignatureDiscriminant
			}
		case sigTagParamSet:
			if len(f.Value) != 2 {
				return s, fmt.Errorf("integrity: sig-param-set is %d octets, want 2", len(f.Value))
			}
			s.ParamSet = binary.BigEndian.Uint16(f.Value)
		case sigTagSignedObject:
			d, _, err := pdlfmt.DecodeDigest256(f.Value)
			if err != nil || len(f.Value) != 32 {
				return s, fmt.Errorf("integrity: sig-signed-object digest256: %v", err)
			}
			s.SignedObject = Digest(d)
		case sigTagValue:
			if len(f.Value) != SigValueLen {
				return s, fmt.Errorf("integrity: sig-value is %d octets, want %d", len(f.Value), SigValueLen)
			}
			copy(s.Value[:], f.Value)
		case sigTagCoverage:
			cd, n, err := DecodeCoverageDescriptor(f.Value)
			if err != nil || n != len(f.Value) {
				return s, fmt.Errorf("integrity: sig-coverage: %v (consumed %d of %d)", err, n, len(f.Value))
			}
			s.Coverage = cd
		case sigTagCredChainRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return s, fmt.Errorf("integrity: sig-cred-chain-ref unit-id: %v", err)
			}
			s.CredChainRef = id
		case sigTagRevocationRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return s, fmt.Errorf("integrity: sig-revocation-ref unit-id: %v", err)
			}
			s.RevocationRef = id
		case sigTagTimeAttestRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return s, fmt.Errorf("integrity: sig-time-attestation-ref unit-id: %v", err)
			}
			s.TimeAttestationRef = id
		case sigTagIntent:
			if len(f.Value) != 1 {
				return s, fmt.Errorf("integrity: sig-intent is %d octets, want 1", len(f.Value))
			}
			s.Intent = f.Value[0]
		case sigTagPresentationRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return s, fmt.Errorf("integrity: sig-presentation-ref unit-id: %v", err)
			}
			s.PresentationRef = id
		}
	}
	for tag := byte(sigTagDiscriminant); tag <= sigTagPresentationRef; tag++ {
		if !seen[tag] {
			return s, fmt.Errorf("%w: tag %d", ErrSignatureMissingField, tag)
		}
	}
	// Non-zero16 references (integrity.abnf S5): cred-chain and time-attestation.
	if s.CredChainRef == zeroUnitID {
		return s, fmt.Errorf("%w: sig-cred-chain-ref", ErrSignatureZeroRef)
	}
	if s.TimeAttestationRef == zeroUnitID {
		return s, fmt.Errorf("%w: sig-time-attestation-ref", ErrSignatureZeroRef)
	}
	return s, nil
}
