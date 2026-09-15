// Package container implements the Protodoc Ledger (PDL) fixed prefix:
// the Header, CommitRingRecord, Frontmatter and SegmentTable structs
// occupying the first 1,048,576 octets of every document
// (contracts/container.abnf S1-S4, data-model.md layer "Container (prefix)").
package container

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// HeaderSize is the fixed size of the Header region, offsets [0, 512).
const HeaderSize = 512

// format-identity occupies offsets [0, 32). This region is FROZEN FOREVER
// starting at format-major=1: no future major version may relocate, resize
// or reinterpret any field inside it (CP-008, FR-006, FR-010).

// headerReservedSize is header-reserved's fixed size, offsets [32, 480).
const headerReservedSize = 448

// Field offsets within format-identity (all within Header's own [0,32)).
const (
	offMagic               = 0   // [0,8)
	offFormatMajor         = 8   // [8,10)
	offFormatMinor         = 10  // [10,12)
	offDocumentClass       = 12  // [12,14)
	offCapabilityWritten   = 14  // [14,16)
	offCapabilityRequired  = 16  // [16,18)
	offDurableClaim        = 18  // [18]
	offHistoryMode         = 19  // [19]
	offUnicodeVersionID    = 20  // [20,22)
	offShapingProfileID    = 22  // [22,24)
	offPrefixLayoutID      = 24  // [24,32)
	offHeaderReservedStart = 32  // [32,480)
	offHeaderDigest        = 480 // [480,512)
)

// Magic is the registered magic constant (FR-125): ASCII "PDL1" followed
// by 4 reserved zero octets. Content-independent, user-supplied-value-free
// (FR-007): no writer may vary it.
var Magic = [8]byte{'P', 'D', 'L', '1', 0, 0, 0, 0}

// IsProtodocMagic reports whether data begins with the registered Protodoc
// magic constant (FR-125). This is the format-identification pattern match
// any consumer (an OS file-type sniffer, a MIME registration probe, a CLI)
// uses to identify a Protodoc file from its leading octets alone, without
// decoding the rest of the header — ahead of the eventual CON-026/CP-014
// registry submission, which is a governance action outside this function's
// scope. It is the sole magic-pattern check in this package (CP-008):
// DecodeHeader calls it too, rather than repeating the comparison.
func IsProtodocMagic(data []byte) bool {
	if len(data) < len(Magic) {
		return false
	}
	return bytes.Equal(data[:len(Magic)], Magic[:])
}

// HistoryMode is the closed 3-value history-retention enum (CON-022).
type HistoryMode uint8

const (
	HistoryComplete          HistoryMode = 0
	HistoryRetainedFromPoint HistoryMode = 1
	HistoryNone              HistoryMode = 2
)

func (m HistoryMode) valid() bool {
	return m == HistoryComplete || m == HistoryRetainedFromPoint || m == HistoryNone
}

// Header is the Protodoc document's fixed 512-octet header (contracts/
// container.abnf S1). It is a raw fixed-layout struct with zero PDL-TLV
// framing: every field has a fixed offset, never a tag.
type Header struct {
	FormatMajor        uint16
	FormatMinor        uint16
	DocumentClass      uint16
	CapabilityWritten  uint16
	CapabilityRequired uint16
	DurableClaim       bool
	HistoryMode        HistoryMode
	UnicodeVersionID   uint16
	ShapingProfileID   uint16
	PrefixLayoutID     uint64
}

var (
	// ErrHeaderTruncated is returned when fewer than HeaderSize octets are available.
	ErrHeaderTruncated = errors.New("container: header truncated, need 512 octets")
	// ErrBadMagic is returned when the leading 8 octets do not match the registered magic.
	ErrBadMagic = errors.New("container: magic constant mismatch")
	// ErrHeaderDigestMismatch is returned when header-digest does not verify (FR-104).
	ErrHeaderDigestMismatch = errors.New("container: header-digest mismatch")
	// ErrHeaderReservedNonzero is rule PD-HDRZERO-001: header-reserved must be all-zero.
	ErrHeaderReservedNonzero = errors.New("container: header-reserved is not all-zero (PD-HDRZERO-001)")
	// ErrInvalidDocumentClass rejects document-class = 0 (reserved/invalid).
	ErrInvalidDocumentClass = errors.New("container: document-class 0 is reserved/invalid")
	// ErrInvalidDurableClaim rejects any durable-claim octet other than 0 or 1 (PD-DURABLE-001).
	ErrInvalidDurableClaim = errors.New("container: durable-claim must be 0 or 1 (PD-DURABLE-001)")
	// ErrInvalidHistoryMode rejects any history-mode octet outside the closed 3-value set.
	ErrInvalidHistoryMode = errors.New("container: history-mode outside the closed {0,1,2} set")
)

// CapabilityMismatchError is FR-010's PD-CAPPAIR-001 rejection: reports
// both the required and written generation numbers.
type CapabilityMismatchError struct {
	Required, Written uint16
}

func (e *CapabilityMismatchError) Error() string {
	return fmt.Sprintf("container: capability-required (%d) exceeds capability-written (%d), rule PD-CAPPAIR-001", e.Required, e.Written)
}

// Encode writes h's canonical 512-octet encoding to dst, which must be at
// least HeaderSize octets. It returns dst[:HeaderSize]. Encode does not
// validate h's field values against the closed enums (durable-claim,
// history-mode, capability ordering) — Decode does, on the assumption
// that a caller constructing a Header directly is responsible for supplying
// already-valid values; validation belongs at the trust boundary.
func (h *Header) Encode(dst []byte) []byte {
	if cap(dst) < HeaderSize {
		dst = make([]byte, HeaderSize)
	} else {
		dst = dst[:HeaderSize]
		for i := range dst {
			dst[i] = 0
		}
	}
	copy(dst[offMagic:offMagic+8], Magic[:])
	binary.BigEndian.PutUint16(dst[offFormatMajor:], h.FormatMajor)
	binary.BigEndian.PutUint16(dst[offFormatMinor:], h.FormatMinor)
	binary.BigEndian.PutUint16(dst[offDocumentClass:], h.DocumentClass)
	binary.BigEndian.PutUint16(dst[offCapabilityWritten:], h.CapabilityWritten)
	binary.BigEndian.PutUint16(dst[offCapabilityRequired:], h.CapabilityRequired)
	if h.DurableClaim {
		dst[offDurableClaim] = 1
	}
	dst[offHistoryMode] = byte(h.HistoryMode)
	binary.BigEndian.PutUint16(dst[offUnicodeVersionID:], h.UnicodeVersionID)
	binary.BigEndian.PutUint16(dst[offShapingProfileID:], h.ShapingProfileID)
	binary.BigEndian.PutUint64(dst[offPrefixLayoutID:], h.PrefixLayoutID)
	// dst[32:480] (header-reserved) is already zero from the fill above.
	digest := sha256.Sum256(dst[:offHeaderDigest])
	copy(dst[offHeaderDigest:HeaderSize], digest[:])
	return dst
}

// DecodeHeader decodes and validates a Header from the leading HeaderSize
// octets of src. Per FR-104, header-digest is verified BEFORE any other
// field is trusted, and before any decoder for later prefix regions is
// constructed (CP-006). Validation order after the digest check: magic,
// header-reserved all-zero (PD-HDRZERO-001), document-class != 0,
// durable-claim in {0,1} (PD-DURABLE-001), history-mode in {0,1,2}, then
// capability-required <= capability-written (PD-CAPPAIR-001, reported last
// since it is the one FR-010 explicitly requires naming both values).
func DecodeHeader(src []byte) (*Header, error) {
	if len(src) < HeaderSize {
		return nil, ErrHeaderTruncated
	}
	src = src[:HeaderSize]

	// Verify header-digest before trusting anything else in the header,
	// including the magic bytes: an attacker's whole point of corrupting
	// the header is to make it look plausible.
	wantDigest := src[offHeaderDigest:HeaderSize]
	gotDigest := sha256.Sum256(src[:offHeaderDigest])
	if !bytes.Equal(gotDigest[:], wantDigest) {
		return nil, fmt.Errorf("%w: expected %x, got %x", ErrHeaderDigestMismatch, wantDigest, gotDigest[:])
	}

	if !IsProtodocMagic(src) {
		return nil, fmt.Errorf("%w: got %x", ErrBadMagic, src[offMagic:offMagic+8])
	}

	for _, b := range src[offHeaderReservedStart : offHeaderReservedStart+headerReservedSize] {
		if b != 0 {
			return nil, ErrHeaderReservedNonzero
		}
	}

	h := &Header{
		FormatMajor:        binary.BigEndian.Uint16(src[offFormatMajor:]),
		FormatMinor:        binary.BigEndian.Uint16(src[offFormatMinor:]),
		DocumentClass:      binary.BigEndian.Uint16(src[offDocumentClass:]),
		CapabilityWritten:  binary.BigEndian.Uint16(src[offCapabilityWritten:]),
		CapabilityRequired: binary.BigEndian.Uint16(src[offCapabilityRequired:]),
		UnicodeVersionID:   binary.BigEndian.Uint16(src[offUnicodeVersionID:]),
		ShapingProfileID:   binary.BigEndian.Uint16(src[offShapingProfileID:]),
		PrefixLayoutID:     binary.BigEndian.Uint64(src[offPrefixLayoutID:]),
	}

	if h.DocumentClass == 0 {
		return nil, ErrInvalidDocumentClass
	}

	switch src[offDurableClaim] {
	case 0:
		h.DurableClaim = false
	case 1:
		h.DurableClaim = true
	default:
		return nil, ErrInvalidDurableClaim
	}

	h.HistoryMode = HistoryMode(src[offHistoryMode])
	if !h.HistoryMode.valid() {
		return nil, ErrInvalidHistoryMode
	}

	if h.CapabilityRequired > h.CapabilityWritten {
		return nil, &CapabilityMismatchError{Required: h.CapabilityRequired, Written: h.CapabilityWritten}
	}

	return h, nil
}
