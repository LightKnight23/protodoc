package pdlfmt

import (
	"errors"
	"fmt"
)

// ErrTagNotAscending is returned when a record's field tags are not in
// strictly ascending order (rule PD-TLV-001).
var ErrTagNotAscending = errors.New("pdlfmt: field tags not strictly ascending")

// ErrUnknownTag is returned when a decoded tag has no matching field-kind
// in the caller's schema and falls outside any documented reserved tail.
// Per container.abnf S2, an unrecognised tag is a structural rejection,
// never a silent skip, unless it lies in a record type's documented
// reserved tail (S5.3).
var ErrUnknownTag = errors.New("pdlfmt: unknown field tag")

// ErrTruncatedField is returned when a field's declared length exceeds
// the octets actually remaining in the input.
var ErrTruncatedField = errors.New("pdlfmt: truncated field value")

// MaxFieldsPerRecord is the permanent per-record-type field ceiling
// (CON-009): a record type has at most 256 fields, ever.
const MaxFieldsPerRecord = 256

// Field is one PDL-TLV (tag, length, value) frame. Length is never stored
// explicitly — it is len(Value) — since field-len is a wire-only artifact
// of encoding, not a fact about the decoded value.
type Field struct {
	Tag   byte
	Value []byte
}

// AppendField appends f's TLV encoding (tag, minimal-varint length, value)
// to dst and returns the extended slice.
func AppendField(dst []byte, f Field) []byte {
	dst = append(dst, f.Tag)
	dst = AppendVarint(dst, uint64(len(f.Value)))
	dst = append(dst, f.Value...)
	return dst
}

// EncodeRecord encodes fields as a PDL-TLV record: the concatenation of
// each field's frame, with no in-band terminator (a record is bounded by
// its enclosing frame's declared length, never a terminator octet).
//
// fields must already be in strictly ascending tag order — EncodeRecord
// enforces this rather than silently sorting, because a writer that does
// not already produce ascending tags has a bug the sort would hide.
func EncodeRecord(fields []Field) ([]byte, error) {
	if len(fields) > MaxFieldsPerRecord {
		return nil, fmt.Errorf("pdlfmt: %d fields exceeds the %d-field-per-record-type ceiling (CON-009)", len(fields), MaxFieldsPerRecord)
	}
	var out []byte
	var prevTag int = -1
	for _, f := range fields {
		if int(f.Tag) <= prevTag {
			return nil, fmt.Errorf("%w: tag %d does not follow tag %d", ErrTagNotAscending, f.Tag, prevTag)
		}
		prevTag = int(f.Tag)
		out = AppendField(out, f)
	}
	return out, nil
}

// DecodeRecord decodes every field in src. known lists the tags the
// caller's schema defines for this record type; a decoded tag not in
// known is accepted only when reservedFrom >= 0 and the tag is >=
// reservedFrom (the record type's documented reserved tail, S5.3), in
// which case the field is still returned to the caller (so e.g. a
// forward-compatible writer can preserve it) but is not validated beyond
// well-formedness. Any other unknown tag is ErrUnknownTag: a structural
// rejection, never a silent skip.
//
// DecodeRecord itself enforces: strict ascending tag order (PD-TLV-001,
// a repeat or descending tag is rejected naming both tags), the
// CON-009 256-field ceiling, and that no field's declared length exceeds
// the octets actually remaining (ErrTruncatedField, checked before any
// allocation of that size per FR-106).
func DecodeRecord(src []byte, known map[byte]bool, reservedFrom int) ([]Field, error) {
	var fields []Field
	var prevTag int = -1
	offset := 0
	for offset < len(src) {
		if len(fields) >= MaxFieldsPerRecord {
			return nil, fmt.Errorf("pdlfmt: record exceeds the %d-field-per-record-type ceiling (CON-009)", MaxFieldsPerRecord)
		}
		tag := src[offset]
		if int(tag) <= prevTag {
			return nil, fmt.Errorf("%w: tag %d does not follow tag %d at offset %d", ErrTagNotAscending, tag, prevTag, offset)
		}
		length, n, err := DecodeVarint(src[offset+1:])
		if err != nil {
			return nil, fmt.Errorf("pdlfmt: decoding length for tag %d at offset %d: %w", tag, offset, err)
		}
		valueStart := offset + 1 + n
		valueEnd := valueStart + int(length)
		if length > uint64(len(src)) || valueEnd < valueStart || valueEnd > len(src) {
			// Declared length checked against remaining input BEFORE any
			// allocation of that size, per FR-106.
			return nil, fmt.Errorf("%w: tag %d declares %d octets, offset %d, input has %d remaining", ErrTruncatedField, tag, length, offset, len(src)-valueStart)
		}
		if !known[tag] {
			if reservedFrom < 0 || int(tag) < reservedFrom {
				return nil, fmt.Errorf("%w: tag %d at offset %d", ErrUnknownTag, tag, offset)
			}
		}
		value := make([]byte, length)
		copy(value, src[valueStart:valueEnd])
		fields = append(fields, Field{Tag: tag, Value: value})
		prevTag = int(tag)
		offset = valueEnd
	}
	return fields, nil
}
