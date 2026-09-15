// Package pdlfmt implements PDL-VARINT and PDL-TLV, the two encoding
// primitives every Protodoc fixed-prefix struct and every TLV record
// composes through (contracts/container.abnf S2).
package pdlfmt

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrNonMinimalVarint is returned when a decoded PDL-VARINT uses a wider
// form than the minimal encoding for its value (rule PD-VARINT-001).
var ErrNonMinimalVarint = errors.New("pdlfmt: non-minimal varint encoding")

// ErrTruncatedVarint is returned when the input ends before a varint's
// trailing octets are fully present.
var ErrTruncatedVarint = errors.New("pdlfmt: truncated varint")

// Varint bounds, per container.abnf S2:
//
//	0x00-0xFC        value = octet itself, 0..252
//	0xFD + 2 octets  value = big-endian u16, MUST be > 252
//	0xFE + 4 octets  value = big-endian u32, MUST be > 65535
//	0xFF + 8 octets  value = big-endian u64, MUST be > 4294967295
const (
	varintU16Prefix = 0xFD
	varintU32Prefix = 0xFE
	varintU64Prefix = 0xFF

	varintU8Max  = 0xFC
	varintU16Max = 0xFFFF
	varintU32Max = 0xFFFFFFFF
)

// AppendVarint appends the minimal PDL-VARINT encoding of v to dst and
// returns the extended slice.
func AppendVarint(dst []byte, v uint64) []byte {
	switch {
	case v <= varintU8Max:
		return append(dst, byte(v))
	case v <= varintU16Max:
		var buf [2]byte
		binary.BigEndian.PutUint16(buf[:], uint16(v))
		return append(append(dst, varintU16Prefix), buf[:]...)
	case v <= varintU32Max:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(v))
		return append(append(dst, varintU32Prefix), buf[:]...)
	default:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], v)
		return append(append(dst, varintU64Prefix), buf[:]...)
	}
}

// EncodeVarint returns the minimal PDL-VARINT encoding of v as a new slice.
func EncodeVarint(v uint64) []byte {
	return AppendVarint(nil, v)
}

// DecodeVarint decodes a PDL-VARINT from the start of src. It returns the
// decoded value and the number of octets consumed. It rejects truncated
// input (ErrTruncatedVarint) and any non-minimal encoding (ErrNonMinimalVarint)
// per rule PD-VARINT-001 — a wider form than necessary for the value is a
// structural validity failure, not merely discouraged.
func DecodeVarint(src []byte) (value uint64, consumed int, err error) {
	if len(src) < 1 {
		return 0, 0, ErrTruncatedVarint
	}
	lead := src[0]
	switch {
	case lead <= varintU8Max:
		return uint64(lead), 1, nil
	case lead == varintU16Prefix:
		if len(src) < 3 {
			return 0, 0, ErrTruncatedVarint
		}
		v := binary.BigEndian.Uint16(src[1:3])
		if v <= varintU8Max {
			return 0, 0, fmt.Errorf("%w: u16 form encodes %d, fits in 1 octet", ErrNonMinimalVarint, v)
		}
		return uint64(v), 3, nil
	case lead == varintU32Prefix:
		if len(src) < 5 {
			return 0, 0, ErrTruncatedVarint
		}
		v := binary.BigEndian.Uint32(src[1:5])
		if v <= varintU16Max {
			return 0, 0, fmt.Errorf("%w: u32 form encodes %d, fits in a narrower form", ErrNonMinimalVarint, v)
		}
		return uint64(v), 5, nil
	default: // lead == varintU64Prefix (0xFF)
		if len(src) < 9 {
			return 0, 0, ErrTruncatedVarint
		}
		v := binary.BigEndian.Uint64(src[1:9])
		if v <= varintU32Max {
			return 0, 0, fmt.Errorf("%w: u64 form encodes %d, fits in a narrower form", ErrNonMinimalVarint, v)
		}
		return v, 9, nil
	}
}

// VarintLen returns the number of octets AppendVarint/EncodeVarint would
// produce for v, without allocating.
func VarintLen(v uint64) int {
	switch {
	case v <= varintU8Max:
		return 1
	case v <= varintU16Max:
		return 3
	case v <= varintU32Max:
		return 5
	default:
		return 9
	}
}
