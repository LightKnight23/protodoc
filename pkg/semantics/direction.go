// Base-writing-direction field (T-0268/T-0269; FR-032). A mandatory closed-enum
// base writing direction on every text-container, mirrored from data-model.md
// into document.abnf's text-block / table productions (tb-direction /
// tbl-direction, tag=5) as a PDL-TLV field. Provisional per the T-0267 ruling
// (clarify-002.md, OPEN); reversible until Eyvar rules.
package semantics

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// Direction is the base writing direction, a CLOSED value set {LTR, RTL}
// matching document.abnf's tb-direction-value / tbl-direction (0 = LTR,
// 1 = RTL). Any other octet is rejected.
type Direction uint8

const (
	// DirectionLTR is left-to-right base direction (wire value 0x00).
	DirectionLTR Direction = 0
	// DirectionRTL is right-to-left base direction (wire value 0x01).
	DirectionRTL Direction = 1
)

// DirectionFieldTag is the PDL-TLV field-kind tag assigned to the direction
// field in the text-block and table productions (tag=5).
const DirectionFieldTag = 5

// ErrDirectionOutOfSet is returned when a direction octet is outside the closed
// value set {0, 1}.
var ErrDirectionOutOfSet = errors.New("semantics: direction outside the closed value set {LTR, RTL}")

// Valid reports whether d is one of the two defined directions.
func (d Direction) Valid() bool { return d == DirectionLTR || d == DirectionRTL }

// EncodeDirectionField encodes the direction as a PDL-TLV field (tag=5, a
// single closed-set value octet), appended to buf.
func EncodeDirectionField(buf []byte, d Direction) ([]byte, error) {
	if !d.Valid() {
		return nil, ErrDirectionOutOfSet
	}
	// PDL-TLV field: varint tag, varint length, value octets.
	buf = pdlfmt.AppendVarint(buf, DirectionFieldTag)
	buf = pdlfmt.AppendVarint(buf, 1) // one value octet
	buf = append(buf, byte(d))
	return buf, nil
}

// DecodeDirectionField decodes a direction PDL-TLV field from buf, returning
// the direction, the bytes consumed, and an error for a bad tag/length or an
// out-of-set value.
func DecodeDirectionField(buf []byte) (Direction, int, error) {
	tag, n, err := pdlfmt.DecodeVarint(buf)
	if err != nil {
		return 0, 0, err
	}
	if tag != DirectionFieldTag {
		return 0, 0, errors.New("semantics: not a direction field tag")
	}
	off := n
	length, n2, err := pdlfmt.DecodeVarint(buf[off:])
	if err != nil {
		return 0, 0, err
	}
	off += n2
	if length != 1 || off+1 > len(buf) {
		return 0, 0, errors.New("semantics: malformed direction field")
	}
	d := Direction(buf[off])
	off++
	if !d.Valid() {
		return 0, 0, ErrDirectionOutOfSet
	}
	return d, off, nil
}
