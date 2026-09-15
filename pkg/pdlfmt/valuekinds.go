package pdlfmt

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf8"
)

// Digest256 is a SHA-256 output (contracts/container.abnf: digest256 = 32OCTET).
type Digest256 [32]byte

// ErrInvalidUTF8 is returned when a decoded nfc-string is not well-formed UTF-8.
var ErrInvalidUTF8 = errors.New("pdlfmt: nfc-string value is not valid UTF-8")

// ErrSortOrder is returned when a sorted-vec's elements are not in strict
// ascending unsigned byte-lexicographic order over their complete encoded
// octets (rule PD-SORT-001): equal-adjacent or descending-adjacent
// elements are both rejected.
var ErrSortOrder = errors.New("pdlfmt: sorted-vec elements not strictly ascending")

// AppendDigest256 appends d's 32 raw octets to dst. digest256 has no length
// prefix of its own: it is always a fixed 32 octets, whether used as a
// whole field-value or nested in a larger structure.
func AppendDigest256(dst []byte, d Digest256) []byte {
	return append(dst, d[:]...)
}

// DecodeDigest256 decodes a fixed 32-octet digest from the start of src,
// returning the value and octets consumed (always 32 on success).
func DecodeDigest256(src []byte) (Digest256, int, error) {
	var d Digest256
	if len(src) < 32 {
		return d, 0, fmt.Errorf("pdlfmt: digest256 needs 32 octets, got %d", len(src))
	}
	copy(d[:], src[:32])
	return d, 32, nil
}

// maxUint48 is the largest value representable in 48 bits (2^48 - 1).
const maxUint48 = 1<<48 - 1

// AppendUint48 appends v's big-endian 48-bit encoding (6 octets) to dst,
// per contracts/container.abnf: u48 = 6OCTET. It errors rather than
// silently truncating when v exceeds the 48-bit range.
func AppendUint48(dst []byte, v uint64) ([]byte, error) {
	if v > maxUint48 {
		return nil, fmt.Errorf("pdlfmt: u48 value %d exceeds 48-bit range", v)
	}
	return append(dst,
		byte(v>>40), byte(v>>32), byte(v>>24),
		byte(v>>16), byte(v>>8), byte(v)), nil
}

// DecodeUint48 decodes a fixed 6-octet big-endian u48 from the start of
// src, returning the value and octets consumed (always 6 on success).
func DecodeUint48(src []byte) (uint64, int, error) {
	if len(src) < 6 {
		return 0, 0, fmt.Errorf("pdlfmt: u48 needs 6 octets, got %d", len(src))
	}
	v := uint64(src[0])<<40 | uint64(src[1])<<32 | uint64(src[2])<<24 |
		uint64(src[3])<<16 | uint64(src[4])<<8 | uint64(src[5])
	return v, 6, nil
}

// AppendNFCString appends s's nfc-string encoding (varint octet length,
// then that many UTF-8 octets) to dst, per contracts/document.abnf:
// nfc-string = varint *OCTET. The caller is responsible for s already
// being NFC-normalized (CON-002/CON-003): a conforming writer rejects
// non-NFC input rather than converting it, so this function does not
// normalize — it only rejects invalid UTF-8.
func AppendNFCString(dst []byte, s string) ([]byte, error) {
	if !utf8.ValidString(s) {
		return nil, ErrInvalidUTF8
	}
	dst = AppendVarint(dst, uint64(len(s)))
	dst = append(dst, s...)
	return dst, nil
}

// DecodeNFCString decodes an nfc-string from the start of src, returning
// the string, octets consumed, and an error if the length is truncated
// or the octets are not valid UTF-8. NFC-normalization itself is a
// separate, higher-layer check (CON-002/CON-003): this function only
// decodes the wire shape and validates UTF-8 well-formedness.
func DecodeNFCString(src []byte) (string, int, error) {
	length, n, err := DecodeVarint(src)
	if err != nil {
		return "", 0, fmt.Errorf("pdlfmt: nfc-string length: %w", err)
	}
	end := n + int(length)
	if length > uint64(len(src)) || end < n || end > len(src) {
		return "", 0, fmt.Errorf("%w: declares %d octets, %d remaining", ErrTruncatedField, length, len(src)-n)
	}
	b := src[n:end]
	if !utf8.Valid(b) {
		return "", 0, ErrInvalidUTF8
	}
	return string(b), end, nil
}

// AppendPlainSeq appends a plain-seq-of-X value: a varint element count
// followed by each element's own encoding concatenated in order (DP-002).
// A plain sequence preserves authored/resolved order with no ordering
// validity rule beyond each element's own well-formedness.
func AppendPlainSeq(dst []byte, elements [][]byte) []byte {
	dst = AppendVarint(dst, uint64(len(elements)))
	for _, e := range elements {
		dst = append(dst, e...)
	}
	return dst
}

// AppendSortedVec appends a sorted-vec-of-X value: as AppendPlainSeq, but
// the caller-supplied elements must already be in strict ascending
// unsigned byte-lexicographic order over their complete encoded octets
// (Go bytes.Compare semantics). AppendSortedVec enforces this rather than
// silently sorting, for the same reason EncodeRecord enforces ascending
// tags: a writer producing out-of-order elements has a bug the sort would
// hide (rule PD-SORT-001).
func AppendSortedVec(dst []byte, elements [][]byte) ([]byte, error) {
	for i := 1; i < len(elements); i++ {
		if bytes.Compare(elements[i-1], elements[i]) >= 0 {
			return nil, fmt.Errorf("%w: element %d does not strictly follow element %d", ErrSortOrder, i, i-1)
		}
	}
	return AppendPlainSeq(dst, elements), nil
}

// DecodeSeqCount decodes a plain-seq-of-X or sorted-vec-of-X's leading
// varint element count. The caller then decodes exactly that many
// elements of its own known shape from the remainder, since element
// shape (X) is defined per use site, not by this generic primitive.
func DecodeSeqCount(src []byte) (count uint64, consumed int, err error) {
	return DecodeVarint(src)
}

// CheckSortedVecOrder verifies elements are in strict ascending unsigned
// byte-lexicographic order over their complete encoded octets, for use
// after a caller has decoded a sorted-vec-of-X's elements with its own
// element-shape decoder (rule PD-SORT-001).
func CheckSortedVecOrder(elements [][]byte) error {
	for i := 1; i < len(elements); i++ {
		if bytes.Compare(elements[i-1], elements[i]) >= 0 {
			return fmt.Errorf("%w: element %d does not strictly follow element %d", ErrSortOrder, i, i-1)
		}
	}
	return nil
}
