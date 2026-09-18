// RescindResignRecord wire shape + duplicate-identifier rule (CON-016, FR-110;
// T-0305). This file gives the RescindResignRecord a concrete canonical wire
// encoding (matching integrity.abnf S8's five fields) with an at-limit /
// one-past-limit boundary and applies FR-110's duplicate-identifier rule to the
// record's identity pair (migration_state_id, new_param_set): two records
// resolving to the same identity are rejected naming BOTH locations, never
// renamed or given precedence.
package migrate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// RescindResignWireMax is the maximum canonical wire size of a
// RescindResignRecord: prior-sig-ref(16) + new-param-set(2) +
// new-signed-object(32) + new-signature-value(64) + migration-state-id(32).
const RescindResignWireMax = 16 + 2 + 32 + 64 + 32 // = 146

// EncodeRescindResign appends the record's canonical fixed-width wire form.
func EncodeRescindResign(dst []byte, r RescindResignRecord) []byte {
	dst = append(dst, r.PriorSignatureRef[:]...)
	dst = append(dst, byte(r.NewParamSetID>>8), byte(r.NewParamSetID))
	dst = append(dst, r.NewSignedObject[:]...)
	dst = append(dst, r.NewSignatureValue[:]...)
	dst = append(dst, r.MigrationStateID[:]...)
	return dst
}

// DecodeRescindResign reads a RescindResignRecord from buf. It rejects a buffer
// that is not exactly RescindResignWireMax octets (a one-past-limit buffer is
// rejected), returning the record and octets consumed.
func DecodeRescindResign(buf []byte) (RescindResignRecord, int, error) {
	if len(buf) != RescindResignWireMax {
		return RescindResignRecord{}, 0, fmt.Errorf("migrate: RescindResignRecord wire length %d, want exactly %d", len(buf), RescindResignWireMax)
	}
	var r RescindResignRecord
	p := 0
	copy(r.PriorSignatureRef[:], buf[p:p+16])
	p += 16
	r.NewParamSetID = uint16(buf[p])<<8 | uint16(buf[p+1])
	p += 2
	copy(r.NewSignedObject[:], buf[p:p+32])
	p += 32
	copy(r.NewSignatureValue[:], buf[p:p+64])
	p += 64
	copy(r.MigrationStateID[:], buf[p:p+32])
	p += 32
	return r, p, nil
}

// RRLocation is a physical location of a RescindResignRecord for duplicate
// reporting.
type RRLocation struct {
	SegmentOrdinal uint16
	IntraOffset    uint64
}

func (l RRLocation) String() string {
	return fmt.Sprintf("segment %d offset %d", l.SegmentOrdinal, l.IntraOffset)
}

// RRIdentity is a record's identity pair (migration_state_id, new_param_set).
type RRIdentity struct {
	MigrationStateID pdlfmt.Digest256
	NewParamSetID    uint16
}

// DuplicateRRError names the shared identity and BOTH locations (FR-110).
type DuplicateRRError struct {
	Identity RRIdentity
	First    RRLocation
	Later    RRLocation
}

func (e *DuplicateRRError) Error() string {
	return fmt.Sprintf("migrate: duplicate RescindResignRecord identity at %s and %s", e.First, e.Later)
}

// PlacedRescindResign is a record with its physical location.
type PlacedRescindResign struct {
	Record   RescindResignRecord
	Location RRLocation
}

// CheckDuplicateRescindResign applies FR-110 to a set of placed records: if two
// resolve to the same identity (migration_state_id, new_param_set), it returns
// a *DuplicateRRError naming both locations, never renaming or applying
// precedence. nil means all identities are distinct.
func CheckDuplicateRescindResign(records []PlacedRescindResign) error {
	seen := make(map[RRIdentity]RRLocation, len(records))
	for _, pr := range records {
		id := RRIdentity{MigrationStateID: pr.Record.MigrationStateID, NewParamSetID: pr.Record.NewParamSetID}
		if first, ok := seen[id]; ok {
			return &DuplicateRRError{Identity: id, First: first, Later: pr.Location}
		}
		seen[id] = pr.Location
	}
	return nil
}
