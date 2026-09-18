// FontRecord (T-0245; FR-089). For every font used, the format records EXACTLY
// seven values: the font's name, its version, a digest of its content, the
// numeric variation-axis coordinates used, the set of code points required,
// the set of layout features required, and the embedding-permission values
// observed when the font octets were included. Not six, not eight: the count
// is fixed so two implementations agree on what a font record contains and a
// consumer can reproduce the exact subset a document depends on.
package render

import (
	"errors"
	"reflect"

	"Protodoc/pkg/pdlfmt"
)

// FontRecord records the exactly-seven declared values for a used font
// (FR-089). Adding or removing a field is a conformance break; the count is
// pinned by FontRecordFieldCount and the T-0245 conformance test.
type FontRecord struct {
	// 1. Name: the font's name.
	Name string
	// 2. Version: the font's version.
	Version string
	// 3. ContentDigest: a digest of the font's content.
	ContentDigest pdlfmt.Digest256
	// 4. VariationAxes: the numeric variation-axis coordinates used.
	VariationAxes []int32
	// 5. CodePoints: the set of code points required.
	CodePoints []rune
	// 6. LayoutFeatures: the set of layout features required.
	LayoutFeatures []string
	// 7. EmbeddingPermissions: the embedding-permission values observed when
	//    the font octets were included.
	EmbeddingPermissions []uint16
}

// FontRecordFieldCount is the FR-089 pinned count: exactly seven declared
// values per font.
const FontRecordFieldCount = 7

// fontRecordFieldCount returns the actual number of struct fields on
// FontRecord, used by the conformance test to hold the count at exactly seven.
func fontRecordFieldCount() int {
	return reflect.TypeOf(FontRecord{}).NumField()
}

// ErrFontRecordFieldCount is returned when a decoded font record does not carry
// exactly FontRecordFieldCount fields (missing or extra) -- a conformance
// break (FR-089).
var ErrFontRecordFieldCount = errors.New("render: FONT_RECORD must carry exactly seven fields")

// Encode serializes the font record as exactly seven length-prefixed fields in
// declared order. Each field is a varint length followed by its octets; the
// leading field count is itself written so a decoder can reject a record with
// the wrong number of fields.
func (r FontRecord) Encode() []byte {
	fields := r.encodeFields()
	var buf []byte
	buf = pdlfmt.AppendVarint(buf, uint64(len(fields)))
	for _, f := range fields {
		buf = pdlfmt.AppendVarint(buf, uint64(len(f)))
		buf = append(buf, f...)
	}
	return buf
}

// encodeFields returns the seven field payloads in declared order.
func (r FontRecord) encodeFields() [][]byte {
	var axes []byte
	for _, a := range r.VariationAxes {
		axes = pdlfmt.AppendVarint(axes, uint64(uint32(a)))
	}
	var cps []byte
	for _, c := range r.CodePoints {
		cps = pdlfmt.AppendVarint(cps, uint64(uint32(c)))
	}
	var feats []byte
	for _, f := range r.LayoutFeatures {
		feats = pdlfmt.AppendVarint(feats, uint64(len(f)))
		feats = append(feats, f...)
	}
	var perms []byte
	for _, p := range r.EmbeddingPermissions {
		perms = pdlfmt.AppendVarint(perms, uint64(p))
	}
	return [][]byte{
		[]byte(r.Name),
		[]byte(r.Version),
		r.ContentDigest[:],
		axes,
		cps,
		feats,
		perms,
	}
}

// DecodeFontRecordFieldCount decodes the leading field count and the field
// payloads from a font-record wire buffer, enforcing that EXACTLY
// FontRecordFieldCount fields are present (FR-089). It returns the field
// payloads or ErrFontRecordFieldCount for a missing/extra field. It does not
// re-interpret each field's inner value (the conformance obligation here is the
// field count and round-trip presence).
func DecodeFontRecordFieldCount(buf []byte) ([][]byte, error) {
	count, n, err := pdlfmt.DecodeVarint(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[n:]
	if count != FontRecordFieldCount {
		return nil, ErrFontRecordFieldCount
	}
	fields := make([][]byte, 0, count)
	for i := uint64(0); i < count; i++ {
		l, n, err := pdlfmt.DecodeVarint(buf)
		if err != nil {
			return nil, err
		}
		buf = buf[n:]
		if uint64(len(buf)) < l {
			return nil, errors.New("render: FONT_RECORD field truncated")
		}
		fields = append(fields, buf[:l])
		buf = buf[l:]
	}
	if len(buf) != 0 {
		return nil, errors.New("render: FONT_RECORD trailing octets after seven fields")
	}
	return fields, nil
}
