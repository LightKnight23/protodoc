// Frontmatter: the bounded-prefix preview and metadata record occupying
// container octets [4096,262144) (contracts/container.abnf S4,
// data-model.md S2.3, FR-051, FR-052, FR-053, FR-054).
package container

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"Protodoc/pkg/pdlfmt"
)

// Frontmatter region sizing (contracts/container.abnf S4 comment): the
// region spans [4096,262144), immediately following Header+CommitRing
// (512+3584=4096) and immediately preceding SegmentTable.
const (
	FrontmatterOffset     = HeaderSize + CommitRingSize // 4096
	FrontmatterRegionSize = 258048                      // 262144 - 4096
)

// Per-field octet ceilings named in data-model.md's ceiling table.
const (
	FMTitleMaxSize         = 4096   // FR-054
	FMPreviewRasterMaxSize = 131072 // FR-051
)

// Closed fm-preview-kind enum (contracts/container.abnf S4 fm-preview-kind
// NORMATIVE comment): any other value is rejected.
const (
	PreviewKindPLP1          = 1
	PreviewKindRestrictedPNG = 2
)

// Field tags within the FRONTMATTER-META PDL-TLV record (contracts/
// container.abnf S4 frontmatter-record NORMATIVE comment: tags 1..11,
// strictly ascending, reserved tail from 12). Tags with no corresponding
// Go struct field yet (fmTagSourceSnapshot, fmTagColourProfileID,
// fmTagRetiredTokens, fmTagCoverageSummary) are still part of this
// already-frozen schema, not the reserved tail: DecodeFrontmatter accepts
// them without error, and a later task adds the Go field for each.
const (
	fmTagPreviewKind     = 1
	fmTagPreviewRaster   = 2
	fmTagPreviewDigest   = 3
	fmTagSourceSnapshot  = 4
	fmTagTitle           = 5
	fmTagPageCount       = 6
	fmTagPageDimensions  = 7
	fmTagLanguage        = 8
	fmTagColourProfileID = 9
	fmTagRetiredTokens   = 10
	fmTagCoverageSummary = 11

	// fmReservedTagsFrom is FRONTMATTER-META's reserved tail start (S5.3):
	// a reader must skip, never reject, an unrecognised tag >= this value.
	fmReservedTagsFrom = 12
)

// frontmatterKnownTags is the complete FRONTMATTER-META schema's tag set,
// per contracts/container.abnf S4 — not just the tags this package's
// current struct fields model. A tag defined by the schema but not yet
// surfaced as a Go field must still be accepted (CP-008: this is the sole
// FRONTMATTER-META schema, not a second one for "tags this build knows
// about"), so it is listed here even before its own task lands.
var frontmatterKnownTags = map[byte]bool{
	fmTagPreviewKind:     true,
	fmTagPreviewRaster:   true,
	fmTagPreviewDigest:   true,
	fmTagSourceSnapshot:  true,
	fmTagTitle:           true,
	fmTagPageCount:       true,
	fmTagPageDimensions:  true,
	fmTagLanguage:        true,
	fmTagColourProfileID: true,
	fmTagRetiredTokens:   true,
	fmTagCoverageSummary: true,
}

// Frontmatter is the document's bounded-prefix metadata record (FR-054):
// title, page count, page dimensions and language are all determinable
// from the leading 262,144 octets alone, with no ledger access.
//
// Every Frontmatter field is individually optional on the wire (data-
// model.md S2.3 marks all of them "Required: no"). This type represents
// "absent" as the Go zero value for every field (empty string, 0), so
// there is exactly one way to encode "not supplied" (CON-005): Encode
// omits a field's tag entirely when its value is the zero value, and
// Decode leaves a field at its zero value when the tag is absent.
type Frontmatter struct {
	PreviewKind   byte   // tag=1, PreviewKindPLP1 or PreviewKindRestrictedPNG; 0 = absent (FR-051)
	PreviewRaster []byte // tag=2, <= FMPreviewRasterMaxSize octets, first-page render (FR-051)
	Title         string // tag=5, UTF-8 NFC text, <= FMTitleMaxSize octets (FR-054)
	PageCount     uint32 // tag=6, authored fixed-pagination page count (FR-054, FR-097)
	PageWidth     int64  // tag=7, 1/914400-inch base units (CON-012); wrapped by a dedicated fixed-point type in a later task
	PageHeight    int64  // tag=7
	Language      string // tag=8, BCP-47 tag octets (FR-054)
}

var (
	// ErrFrontmatterTruncated is returned when fewer than FrontmatterRegionSize octets are available.
	ErrFrontmatterTruncated = errors.New("container: frontmatter region truncated, need 258048 octets")
	// ErrFrontmatterPaddingNonzero is returned when frontmatter-padding is not all-zero.
	ErrFrontmatterPaddingNonzero = errors.New("container: frontmatter-padding is not all-zero")
	// ErrFrontmatterRecordTooLarge is returned when the encoded frontmatter-record would exceed FrontmatterRegionSize octets.
	ErrFrontmatterRecordTooLarge = errors.New("container: frontmatter-record exceeds the 258048-octet region")
	// ErrFrontmatterInvalidUTF8 is returned when fm-title or fm-language is not well-formed UTF-8.
	ErrFrontmatterInvalidUTF8 = errors.New("container: frontmatter text field is not valid UTF-8")
	// ErrFrontmatterFieldSize is returned when a field's declared value violates its named ceiling or shape.
	ErrFrontmatterFieldSize = errors.New("container: frontmatter field violates its size or shape constraint")
	// ErrFrontmatterInvalidPreviewKind is returned when fm-preview-kind is outside the closed {1,2} enum.
	ErrFrontmatterInvalidPreviewKind = errors.New("container: fm-preview-kind outside the closed {1,2} set")
)

// frontmatterRecordLength scans region (exactly FrontmatterRegionSize
// octets) for the boundary between the actual frontmatter-record and its
// trailing zero-fill frontmatter-padding (contracts/container.abnf S4
// frontmatter-padding NORMATIVE comment). FRONTMATTER-META's schema never
// assigns tag 0 to a field (the lowest defined tag is 1, and the reserved
// tail begins at 12), so a tag octet of 0 at a field boundary can only be
// padding — this is what lets a reader recover the record/padding split
// with no separate length prefix of its own. It reads no octet outside
// region (FR-054's leading-262144-octets bound).
func frontmatterRecordLength(region []byte) (int, error) {
	offset := 0
	for offset < len(region) {
		tag := region[offset]
		if tag == 0 {
			return offset, nil
		}
		length, n, err := pdlfmt.DecodeVarint(region[offset+1:])
		if err != nil {
			return 0, fmt.Errorf("container: frontmatter field tag %d length at offset %d: %w", tag, offset, err)
		}
		valueStart := offset + 1 + n
		valueEnd := valueStart + int(length)
		if length > uint64(len(region)) || valueEnd < valueStart || valueEnd > len(region) {
			return 0, fmt.Errorf("%w: tag %d declares %d octets at offset %d, region has %d remaining", ErrFrontmatterFieldSize, tag, length, offset, len(region)-valueStart)
		}
		offset = valueEnd
	}
	return offset, nil
}

// Encode writes fm's canonical FrontmatterRegionSize-octet encoding
// (frontmatter-record followed by zero-fill frontmatter-padding) to dst,
// which must be at least FrontmatterRegionSize octets. It returns
// dst[:FrontmatterRegionSize].
func (fm *Frontmatter) Encode(dst []byte) ([]byte, error) {
	var fields []pdlfmt.Field

	if fm.PreviewKind != 0 {
		if fm.PreviewKind != PreviewKindPLP1 && fm.PreviewKind != PreviewKindRestrictedPNG {
			return nil, ErrFrontmatterInvalidPreviewKind
		}
		fields = append(fields, pdlfmt.Field{Tag: fmTagPreviewKind, Value: []byte{fm.PreviewKind}})
	}
	if len(fm.PreviewRaster) > 0 {
		if len(fm.PreviewRaster) > FMPreviewRasterMaxSize {
			return nil, fmt.Errorf("%w: fm-preview-raster is %d octets, max %d", ErrFrontmatterFieldSize, len(fm.PreviewRaster), FMPreviewRasterMaxSize)
		}
		fields = append(fields, pdlfmt.Field{Tag: fmTagPreviewRaster, Value: fm.PreviewRaster})
	}
	if fm.Title != "" {
		if !utf8.ValidString(fm.Title) {
			return nil, ErrFrontmatterInvalidUTF8
		}
		if len(fm.Title) > FMTitleMaxSize {
			return nil, fmt.Errorf("%w: fm-title is %d octets, max %d", ErrFrontmatterFieldSize, len(fm.Title), FMTitleMaxSize)
		}
		fields = append(fields, pdlfmt.Field{Tag: fmTagTitle, Value: []byte(fm.Title)})
	}
	if fm.PageCount != 0 {
		fields = append(fields, pdlfmt.Field{Tag: fmTagPageCount, Value: pdlfmt.EncodeVarint(uint64(fm.PageCount))})
	}
	if fm.PageWidth != 0 || fm.PageHeight != 0 {
		var v []byte
		v = pdlfmt.AppendVarint(v, uint64(fm.PageWidth))
		v = pdlfmt.AppendVarint(v, uint64(fm.PageHeight))
		fields = append(fields, pdlfmt.Field{Tag: fmTagPageDimensions, Value: v})
	}
	if fm.Language != "" {
		if !utf8.ValidString(fm.Language) {
			return nil, ErrFrontmatterInvalidUTF8
		}
		fields = append(fields, pdlfmt.Field{Tag: fmTagLanguage, Value: []byte(fm.Language)})
	}

	record, err := pdlfmt.EncodeRecord(fields)
	if err != nil {
		return nil, fmt.Errorf("container: encoding frontmatter-record: %w", err)
	}
	if len(record) > FrontmatterRegionSize {
		return nil, fmt.Errorf("%w: encoded %d octets", ErrFrontmatterRecordTooLarge, len(record))
	}

	if cap(dst) < FrontmatterRegionSize {
		dst = make([]byte, FrontmatterRegionSize)
	} else {
		dst = dst[:FrontmatterRegionSize]
		for i := range dst {
			dst[i] = 0
		}
	}
	copy(dst, record)
	return dst, nil
}

// DecodeFrontmatter decodes a Frontmatter from the leading
// FrontmatterRegionSize octets of src. It reads no octet at or past
// FrontmatterRegionSize (FR-054's leading-262144-octets bound: the caller
// passes only the region, per DecodeHeader/DecodeCommitRing's convention
// of the caller slicing the overall prefix by fixed offset), and performs
// no I/O of its own (no file open, no network call) beyond reading src.
func DecodeFrontmatter(src []byte) (*Frontmatter, error) {
	if len(src) < FrontmatterRegionSize {
		return nil, ErrFrontmatterTruncated
	}
	region := src[:FrontmatterRegionSize]

	recLen, err := frontmatterRecordLength(region)
	if err != nil {
		return nil, err
	}
	for _, b := range region[recLen:] {
		if b != 0 {
			return nil, ErrFrontmatterPaddingNonzero
		}
	}

	fields, err := pdlfmt.DecodeRecord(region[:recLen], frontmatterKnownTags, fmReservedTagsFrom)
	if err != nil {
		return nil, fmt.Errorf("container: frontmatter record: %w", err)
	}

	fm := &Frontmatter{}
	for _, f := range fields {
		switch f.Tag {
		case fmTagPreviewKind:
			if len(f.Value) != 1 {
				return nil, fmt.Errorf("%w: fm-preview-kind is %d octets, want 1", ErrFrontmatterFieldSize, len(f.Value))
			}
			if f.Value[0] != PreviewKindPLP1 && f.Value[0] != PreviewKindRestrictedPNG {
				return nil, ErrFrontmatterInvalidPreviewKind
			}
			fm.PreviewKind = f.Value[0]
		case fmTagPreviewRaster:
			if len(f.Value) > FMPreviewRasterMaxSize {
				return nil, fmt.Errorf("%w: fm-preview-raster is %d octets, max %d", ErrFrontmatterFieldSize, len(f.Value), FMPreviewRasterMaxSize)
			}
			fm.PreviewRaster = f.Value
		case fmTagTitle:
			if !utf8.Valid(f.Value) {
				return nil, ErrFrontmatterInvalidUTF8
			}
			if len(f.Value) > FMTitleMaxSize {
				return nil, fmt.Errorf("%w: fm-title is %d octets, max %d", ErrFrontmatterFieldSize, len(f.Value), FMTitleMaxSize)
			}
			fm.Title = string(f.Value)
		case fmTagPageCount:
			v, n, err := pdlfmt.DecodeVarint(f.Value)
			if err != nil || n != len(f.Value) {
				return nil, fmt.Errorf("%w: fm-page-count malformed varint", ErrFrontmatterFieldSize)
			}
			if v > 0xFFFFFFFF {
				return nil, fmt.Errorf("%w: fm-page-count %d exceeds uint32", ErrFrontmatterFieldSize, v)
			}
			fm.PageCount = uint32(v)
		case fmTagPageDimensions:
			w, n1, err := pdlfmt.DecodeVarint(f.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: fm-page-dimensions width: %v", ErrFrontmatterFieldSize, err)
			}
			h, n2, err := pdlfmt.DecodeVarint(f.Value[n1:])
			if err != nil || n1+n2 != len(f.Value) {
				return nil, fmt.Errorf("%w: fm-page-dimensions malformed", ErrFrontmatterFieldSize)
			}
			if w > 0x7FFFFFFFFFFFFFFF || h > 0x7FFFFFFFFFFFFFFF {
				return nil, fmt.Errorf("%w: fm-page-dimensions exceeds int64", ErrFrontmatterFieldSize)
			}
			fm.PageWidth = int64(w)
			fm.PageHeight = int64(h)
		case fmTagLanguage:
			if !utf8.Valid(f.Value) {
				return nil, ErrFrontmatterInvalidUTF8
			}
			fm.Language = string(f.Value)
		default:
			// A tag this package's struct does not yet model (preview
			// fields, source-snapshot, colour-profile-id, retired-tokens,
			// coverage-summary): accepted per the schema, not surfaced.
		}
	}
	return fm, nil
}
