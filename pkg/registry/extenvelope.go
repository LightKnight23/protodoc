package registry

import (
	"errors"
	"fmt"

	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

// ExtensionEnvelope (document.abnf S6, ext-envelope; data-model.md 2.20) is the
// single universal carrier for any construct not defined by the core
// specification, so a previous-generation reader can always determine a future
// construct's extent, disposition and fallback without understanding its
// payload (FR-012..018). It is a PDL-TLV record with discriminant 0x06 and the
// fixed field set below; the opaque payload octets live in a separate
// RESOURCE-typed segment addressed by PayloadRef, never inline, so an
// arbitrarily large unknown payload never forces this record past its own
// ceilings (CQ-007).

// ExtEnvelope discriminant (document.abnf S6) and TLV tags.
const (
	extEnvelopeDiscriminant = 0x06
	extTagDiscriminant      = 0 // ext-discriminant, value = 0x06
	extTagTok               = 1 // ext-tok, value = ext-token (8 octets)
	extTagPayloadLength     = 2 // ext-payload-length, value = varint
	extTagPayloadRef        = 3 // ext-payload-ref, value = unit-id
	extTagPayloadDigest     = 4 // ext-payload-digest, value = digest256
	extTagDisposition       = 5 // ext-disposition, value = u8
	extTagFallbackRef       = 6 // ext-fallback-ref, value = unit-id or zero16
	extTagPositionKey       = 7 // ext-position-key, value = anchor-point (22 octets)
)

// ExtDisposition is ext-disposition (document.abnf S6; the disposition enum
// and its structural-reject rule are exercised by T-0054): the
// closed 3-value set telling a reader what to do with a construct it does not
// implement. Exactly these three values are legal; a missing or out-of-range
// disposition is a structural reject, never a reader-chosen default.
type ExtDisposition uint8

const (
	// DispositionIgnore (0x00): render as if the construct were absent.
	DispositionIgnore ExtDisposition = 0x00
	// DispositionDegrade (0x01): degrade to the ext-fallback-ref construct.
	DispositionDegrade ExtDisposition = 0x01
	// DispositionRefuse (0x02): refuse to render the document.
	DispositionRefuse ExtDisposition = 0x02
)

func (d ExtDisposition) String() string {
	switch d {
	case DispositionIgnore:
		return "ignore"
	case DispositionDegrade:
		return "degrade"
	case DispositionRefuse:
		return "refuse"
	default:
		return fmt.Sprintf("invalid(0x%02x)", uint8(d))
	}
}

// IsValid reports whether d is one of the exactly-three legal dispositions.
func (d ExtDisposition) IsValid() bool {
	return d == DispositionIgnore || d == DispositionDegrade || d == DispositionRefuse
}

// Zero16 is container.abnf S0's zero16: 16 zero octets, the "no unit"
// sentinel for a unit-id-typed field (here, an absent ext-fallback-ref).
var Zero16 = pdlfmt.UnitID{}

// ExtEnvelope is the decoded ExtensionEnvelope.
type ExtEnvelope struct {
	Tok           ExtToken
	PayloadLength uint64
	PayloadRef    pdlfmt.UnitID
	PayloadDigest pdlfmt.Digest256
	Disposition   ExtDisposition
	FallbackRef   pdlfmt.UnitID // Zero16 when absent
	PositionKey   content.AnchorPoint
}

// ErrExtEnvelopeDiscriminant is returned when a decoded record's discriminant
// is not 0x06.
var ErrExtEnvelopeDiscriminant = errors.New("registry: record discriminant is not EXT_ENVELOPE (0x06)")

// ErrExtEnvelopeMissingField is returned when a required ext-envelope field is
// absent.
var ErrExtEnvelopeMissingField = errors.New("registry: ext-envelope missing a required field")

// Encode encodes the envelope to its byte-exact PDL-TLV wire form: the eight
// fields in strict ascending tag order (0..7). ext-disposition validity is not
// checked here (a caller building an envelope supplies a valid value);
// Decode validates disposition at the trust boundary.
func (e ExtEnvelope) Encode() ([]byte, error) {
	var payloadLen []byte
	payloadLen = pdlfmt.AppendVarint(payloadLen, e.PayloadLength)

	posKey := content.EncodeAnchorPoint(nil, e.PositionKey)

	fields := []pdlfmt.Field{
		{Tag: extTagDiscriminant, Value: []byte{extEnvelopeDiscriminant}},
		{Tag: extTagTok, Value: e.Tok[:]},
		{Tag: extTagPayloadLength, Value: payloadLen},
		{Tag: extTagPayloadRef, Value: pdlfmt.AppendUnitID(nil, e.PayloadRef)},
		{Tag: extTagPayloadDigest, Value: pdlfmt.AppendDigest256(nil, e.PayloadDigest)},
		{Tag: extTagDisposition, Value: []byte{byte(e.Disposition)}},
		{Tag: extTagFallbackRef, Value: pdlfmt.AppendUnitID(nil, e.FallbackRef)},
		{Tag: extTagPositionKey, Value: posKey},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeExtEnvelope decodes an ExtensionEnvelope, the byte-exact inverse of
// Encode. It rejects a wrong discriminant, a missing required field, and a
// malformed field value; it does NOT itself reject an out-of-range
// disposition or a fallback-rule violation (those are separate structural
// checks, T-0054/T-0055) so the decoder can surface the decoded value for the
// checks to name. It DOES reject any field whose octet width is wrong.
func DecodeExtEnvelope(src []byte) (ExtEnvelope, error) {
	var e ExtEnvelope
	known := map[byte]bool{
		extTagDiscriminant: true, extTagTok: true, extTagPayloadLength: true,
		extTagPayloadRef: true, extTagPayloadDigest: true, extTagDisposition: true,
		extTagFallbackRef: true, extTagPositionKey: true,
	}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return e, err
	}
	seen := map[byte]bool{}
	for _, f := range fields {
		seen[f.Tag] = true
		switch f.Tag {
		case extTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != extEnvelopeDiscriminant {
				return e, ErrExtEnvelopeDiscriminant
			}
		case extTagTok:
			if len(f.Value) != ExtTokenOctets {
				return e, fmt.Errorf("registry: ext-tok is %d octets, want %d", len(f.Value), ExtTokenOctets)
			}
			copy(e.Tok[:], f.Value)
		case extTagPayloadLength:
			v, n, err := pdlfmt.DecodeVarint(f.Value)
			if err != nil || n != len(f.Value) {
				return e, fmt.Errorf("registry: ext-payload-length varint: %v (consumed %d of %d)", err, n, len(f.Value))
			}
			e.PayloadLength = v
		case extTagPayloadRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return e, fmt.Errorf("registry: ext-payload-ref unit-id: %v", err)
			}
			e.PayloadRef = id
		case extTagPayloadDigest:
			d, _, err := pdlfmt.DecodeDigest256(f.Value)
			if err != nil || len(f.Value) != 32 {
				return e, fmt.Errorf("registry: ext-payload-digest digest256: %v", err)
			}
			e.PayloadDigest = d
		case extTagDisposition:
			if len(f.Value) != 1 {
				return e, fmt.Errorf("registry: ext-disposition is %d octets, want 1", len(f.Value))
			}
			e.Disposition = ExtDisposition(f.Value[0])
		case extTagFallbackRef:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return e, fmt.Errorf("registry: ext-fallback-ref unit-id: %v", err)
			}
			e.FallbackRef = id
		case extTagPositionKey:
			ap, n, err := content.DecodeAnchorPoint(f.Value)
			if err != nil || n != len(f.Value) {
				return e, fmt.Errorf("registry: ext-position-key anchor-point: %v (consumed %d of %d)", err, n, len(f.Value))
			}
			e.PositionKey = ap
		}
	}
	// Every field 0..7 is required.
	for tag := byte(extTagDiscriminant); tag <= extTagPositionKey; tag++ {
		if !seen[tag] {
			return e, fmt.Errorf("%w: tag %d", ErrExtEnvelopeMissingField, tag)
		}
	}
	return e, nil
}
