// Bounded-prefix stored-unit enumeration (TR-006, contracts/container.abnf
// S5): a scanner that has read only the leading 1,048,576 octets of a
// document can determine every stored unit's declared type, octet length
// and digest, with no segment body decoded.
package container

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ErrBoundedPrefixTruncated is returned when fewer octets than the fixed
// 1,048,576-octet prefix are supplied to a bounded-prefix reader.
var ErrBoundedPrefixTruncated = fmt.Errorf("container: bounded prefix truncated, need %d octets", SegmentTableOffset+SegmentTableRegionSize)

// StoredUnit is one stored unit's declared identity as read from its
// SegmentTableSlot alone (FR-091, TR-006): type, octet length and digest,
// with no segment body decode involved. Ordinal is the unit's index in the
// segment table (0..MaxSegments-1).
type StoredUnit struct {
	Ordinal     int
	SegmentType byte
	Length      uint64
	Digest      pdlfmt.Digest256
}

// EnumerateStoredUnits returns one StoredUnit per populated slot in table
// (SegmentType != SegmentTypeUnused), in ascending ordinal order. An unused
// slot is the normative all-zero marker (contracts/container.abnf S5.1),
// not a stored unit, so it is omitted rather than returned with a zero
// type. This function performs no I/O and reads no field beyond table
// itself, which the caller has already decoded from the bounded prefix.
func EnumerateStoredUnits(table [MaxSegments]SegmentTableSlot) []StoredUnit {
	var units []StoredUnit
	for i, slot := range table {
		if slot.SegmentType == SegmentTypeUnused {
			continue
		}
		units = append(units, StoredUnit{
			Ordinal:     i,
			SegmentType: slot.SegmentType,
			Length:      slot.Length,
			Digest:      slot.Digest,
		})
	}
	return units
}

// EnumerateStoredUnitsFromBoundedPrefix decodes the SegmentTable region out
// of prefix and enumerates every stored unit's type, length and digest
// (TR-006). prefix must hold at least the leading
// SegmentTableOffset+SegmentTableRegionSize (1,048,576) octets of the
// document; this function reads no octet at or past that bound and
// constructs no segment/font/image/audio/video decoder (CP-006) — it only
// walks the fixed-offset SegmentTableSlot array already defined by
// DecodeSegmentTable.
func EnumerateStoredUnitsFromBoundedPrefix(prefix []byte) ([]StoredUnit, error) {
	if len(prefix) < SegmentTableOffset+SegmentTableRegionSize {
		return nil, ErrBoundedPrefixTruncated
	}
	table, err := DecodeSegmentTable(prefix[SegmentTableOffset : SegmentTableOffset+SegmentTableRegionSize])
	if err != nil {
		return nil, fmt.Errorf("container: enumerating stored units: %w", err)
	}
	return EnumerateStoredUnits(table), nil
}
