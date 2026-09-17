// PresentationArtefact (T-0154, FR-064; document.abnf S7.4) and its embedded
// FontRecord (S7.4.1, FR-089): a RESOURCE-typed record (discriminant 0x0C)
// content-addressed like every other resource. A Signature carries only the
// presentation_artefact_digest (the referenced segment's own slot digest) as
// an input to signed_object; the artefact record itself is never inlined in
// the SIGNATURE. This file is the single canonical wire implementation of both
// records.
//
// pa-page-geometry uses fixed 8-octet big-endian two's-complement i64 (not
// varint) because a coordinate MAY be negative (an object placed partially
// off-page); order is authored page/axis order, never re-sorted.
package integrity

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// PRESENTATION_ARTEFACT discriminant + TLV tags (document.abnf S7.4).
const (
	presentationArtefactDiscriminant = 0x0C
	paTagDiscriminant                = 0 // pa-discriminant, value = 0x0C
	paTagProfileVersion              = 1 // pa-profile-version, u16
	paTagPageGeometry                = 2 // pa-page-geometry, plain-seq-of(i64)
	paTagFontIdentity                = 3 // pa-font-identity, plain-seq-of(font-record)
)

// FontRecord is the embedded font-identity element (document.abnf S7.4.1,
// FR-089): EXACTLY seven fields, no more and no fewer.
type FontRecord struct {
	Name       string   // fr-name, nfc-string
	Version    string   // fr-version, nfc-string
	Digest     Digest   // fr-digest, digest256 = referencing FONT_SUBSET slot digest
	Axes       []int64  // fr-axes: count then that many i64 axis coordinates
	Codepoints []uint32 // fr-codepoints: count then that many varint scalar values
	Features   []uint16 // fr-features: count then that many u16 registry feature ids
	EmbedPerm  uint8    // fr-embed-perm, u8 embedding-permission bitset
}

// PresentationArtefact is the decoded PRESENTATION_ARTEFACT record.
type PresentationArtefact struct {
	ProfileVersion uint16
	PageGeometry   []int64 // authored order, may be negative
	FontIdentity   []FontRecord
}

var (
	// ErrPresentationDiscriminant is returned for a wrong pa-discriminant.
	ErrPresentationDiscriminant = errors.New("integrity: record discriminant is not PRESENTATION_ARTEFACT (0x0C)")
	// ErrPresentationMissingField is returned for a missing required field.
	ErrPresentationMissingField = errors.New("integrity: PRESENTATION_ARTEFACT missing a required field")
	// ErrCodepointRange is returned for a codepoint outside 0..0x10FFFF.
	ErrCodepointRange = errors.New("integrity: font-record codepoint outside 0..0x10FFFF")
)

func appendI64(dst []byte, v int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	return append(dst, b[:]...)
}

func decodeI64(src []byte) (int64, int, error) {
	if len(src) < 8 {
		return 0, 0, fmt.Errorf("integrity: i64 truncated (%d octets)", len(src))
	}
	return int64(binary.BigEndian.Uint64(src[:8])), 8, nil
}

// encodeFontRecord encodes one font-record element (an element of the
// pa-font-identity plain sequence, not itself a discriminated record).
func encodeFontRecord(fr FontRecord) ([]byte, error) {
	var buf []byte
	var err error
	if buf, err = pdlfmt.AppendNFCString(buf, fr.Name); err != nil {
		return nil, fmt.Errorf("integrity: fr-name: %w", err)
	}
	if buf, err = pdlfmt.AppendNFCString(buf, fr.Version); err != nil {
		return nil, fmt.Errorf("integrity: fr-version: %w", err)
	}
	buf = pdlfmt.AppendDigest256(buf, pdlfmt.Digest256(fr.Digest))
	// fr-axes: varint count then that many i64.
	buf = pdlfmt.AppendVarint(buf, uint64(len(fr.Axes)))
	for _, a := range fr.Axes {
		buf = appendI64(buf, a)
	}
	// fr-codepoints: varint count then that many varint scalars.
	buf = pdlfmt.AppendVarint(buf, uint64(len(fr.Codepoints)))
	for _, c := range fr.Codepoints {
		if c > 0x10FFFF {
			return nil, fmt.Errorf("%w: 0x%X", ErrCodepointRange, c)
		}
		buf = pdlfmt.AppendVarint(buf, uint64(c))
	}
	// fr-features: varint count then that many u16.
	buf = pdlfmt.AppendVarint(buf, uint64(len(fr.Features)))
	for _, ft := range fr.Features {
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], ft)
		buf = append(buf, b[:]...)
	}
	buf = append(buf, fr.EmbedPerm)
	return buf, nil
}

// decodeFontRecord decodes one font-record from the start of src, returning it
// and the octets consumed.
func decodeFontRecord(src []byte) (FontRecord, int, error) {
	var fr FontRecord
	pos := 0
	name, n, err := pdlfmt.DecodeNFCString(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-name: %w", err)
	}
	fr.Name = name
	pos += n
	ver, n, err := pdlfmt.DecodeNFCString(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-version: %w", err)
	}
	fr.Version = ver
	pos += n
	d, n, err := pdlfmt.DecodeDigest256(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-digest: %w", err)
	}
	fr.Digest = Digest(d)
	pos += n

	// fr-axes.
	axisCount, n, err := pdlfmt.DecodeVarint(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-axes count: %w", err)
	}
	pos += n
	if axisCount > uint64((len(src)-pos)/8) {
		return fr, 0, fmt.Errorf("integrity: fr-axes count %d exceeds remaining octets", axisCount)
	}
	for i := uint64(0); i < axisCount; i++ {
		v, n, err := decodeI64(src[pos:])
		if err != nil {
			return fr, 0, err
		}
		fr.Axes = append(fr.Axes, v)
		pos += n
	}

	// fr-codepoints.
	cpCount, n, err := pdlfmt.DecodeVarint(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-codepoints count: %w", err)
	}
	pos += n
	if cpCount > uint64(len(src)-pos) {
		return fr, 0, fmt.Errorf("integrity: fr-codepoints count %d exceeds remaining octets", cpCount)
	}
	for i := uint64(0); i < cpCount; i++ {
		v, n, err := pdlfmt.DecodeVarint(src[pos:])
		if err != nil {
			return fr, 0, fmt.Errorf("integrity: fr-codepoint %d: %w", i, err)
		}
		if v > 0x10FFFF {
			return fr, 0, fmt.Errorf("%w: 0x%X", ErrCodepointRange, v)
		}
		fr.Codepoints = append(fr.Codepoints, uint32(v))
		pos += n
	}

	// fr-features.
	ftCount, n, err := pdlfmt.DecodeVarint(src[pos:])
	if err != nil {
		return fr, 0, fmt.Errorf("integrity: fr-features count: %w", err)
	}
	pos += n
	if ftCount > uint64((len(src)-pos)/2) {
		return fr, 0, fmt.Errorf("integrity: fr-features count %d exceeds remaining octets", ftCount)
	}
	for i := uint64(0); i < ftCount; i++ {
		if pos+2 > len(src) {
			return fr, 0, fmt.Errorf("integrity: fr-feature %d truncated", i)
		}
		fr.Features = append(fr.Features, binary.BigEndian.Uint16(src[pos:pos+2]))
		pos += 2
	}

	if pos >= len(src) {
		return fr, 0, fmt.Errorf("integrity: fr-embed-perm truncated")
	}
	fr.EmbedPerm = src[pos]
	pos++
	return fr, pos, nil
}

// Encode encodes the PresentationArtefact to its byte-exact PDL-TLV wire form.
func (p PresentationArtefact) Encode() ([]byte, error) {
	var ver [2]byte
	binary.BigEndian.PutUint16(ver[:], p.ProfileVersion)

	// pa-page-geometry: plain-seq-of(i64), authored order.
	var geom []byte
	geom = pdlfmt.AppendVarint(geom, uint64(len(p.PageGeometry)))
	for _, g := range p.PageGeometry {
		geom = appendI64(geom, g)
	}

	// pa-font-identity: plain-seq-of(font-record).
	elems := make([][]byte, len(p.FontIdentity))
	for i, fr := range p.FontIdentity {
		b, err := encodeFontRecord(fr)
		if err != nil {
			return nil, err
		}
		elems[i] = b
	}
	fontSeq := pdlfmt.AppendPlainSeq(nil, elems)

	fields := []pdlfmt.Field{
		{Tag: paTagDiscriminant, Value: []byte{presentationArtefactDiscriminant}},
		{Tag: paTagProfileVersion, Value: ver[:]},
		{Tag: paTagPageGeometry, Value: geom},
		{Tag: paTagFontIdentity, Value: fontSeq},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodePresentationArtefact decodes a PresentationArtefact, the byte-exact
// inverse of Encode.
func DecodePresentationArtefact(src []byte) (PresentationArtefact, error) {
	var p PresentationArtefact
	known := map[byte]bool{paTagDiscriminant: true, paTagProfileVersion: true, paTagPageGeometry: true, paTagFontIdentity: true}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return p, err
	}
	seen := map[byte]bool{}
	for _, f := range fields {
		seen[f.Tag] = true
		switch f.Tag {
		case paTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != presentationArtefactDiscriminant {
				return p, ErrPresentationDiscriminant
			}
		case paTagProfileVersion:
			if len(f.Value) != 2 {
				return p, fmt.Errorf("integrity: pa-profile-version is %d octets, want 2", len(f.Value))
			}
			p.ProfileVersion = binary.BigEndian.Uint16(f.Value)
		case paTagPageGeometry:
			count, n, err := pdlfmt.DecodeVarint(f.Value)
			if err != nil {
				return p, fmt.Errorf("integrity: pa-page-geometry count: %w", err)
			}
			pos := n
			if count > uint64((len(f.Value)-pos)/8) {
				return p, fmt.Errorf("integrity: pa-page-geometry count %d exceeds remaining octets", count)
			}
			for i := uint64(0); i < count; i++ {
				v, n, err := decodeI64(f.Value[pos:])
				if err != nil {
					return p, err
				}
				p.PageGeometry = append(p.PageGeometry, v)
				pos += n
			}
			if pos != len(f.Value) {
				return p, fmt.Errorf("integrity: pa-page-geometry has %d trailing octets", len(f.Value)-pos)
			}
		case paTagFontIdentity:
			count, n, err := pdlfmt.DecodeSeqCount(f.Value)
			if err != nil {
				return p, fmt.Errorf("integrity: pa-font-identity count: %w", err)
			}
			pos := n
			for i := uint64(0); i < count; i++ {
				fr, consumed, err := decodeFontRecord(f.Value[pos:])
				if err != nil {
					return p, err
				}
				p.FontIdentity = append(p.FontIdentity, fr)
				pos += consumed
			}
			if pos != len(f.Value) {
				return p, fmt.Errorf("integrity: pa-font-identity has %d trailing octets", len(f.Value)-pos)
			}
		}
	}
	for tag := byte(paTagDiscriminant); tag <= paTagFontIdentity; tag++ {
		if !seen[tag] {
			return p, fmt.Errorf("%w: tag %d", ErrPresentationMissingField, tag)
		}
	}
	return p, nil
}
