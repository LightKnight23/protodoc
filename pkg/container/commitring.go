// Commit ring: the 7-slot self-digesting ring occupying container octets
// [512, 4096) (contracts/container.abnf S3, data-model.md ceiling table
// "CommitRing slot size"/"CommitRing region size", HC-027).
package container

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// Commit-ring sizing, per contracts/container.abnf S3 and data-model.md's
// ceiling table (HC-027): 7 slots x 512 octets = 3584 octets, immediately
// following Header at offset 512.
const (
	CommitRingSlots = 7
	RingSlotSize    = 512
	CommitRingSize  = CommitRingSlots * RingSlotSize
)

// ring-fields (284 octets) + ring-reserved (196 octets, MBZ) + record-digest
// (32 octets) = 512 (contracts/container.abnf S3 ring-slot comment).
const (
	ringFieldsSize   = 284
	ringReservedSize = 196
)

// Field offsets within a single 512-octet ring slot (contracts/
// container.abnf S3 ring-fields NORMATIVE layout comment).
const (
	roRingMagic            = 0   // [0,4)
	roSequence             = 4   // [4,12)
	roLedgerLength         = 12  // [12,20)
	roLedgerRoot           = 20  // [20,52)
	roSegmentCount         = 52  // [52,54)
	roStateID              = 54  // [54,86)
	roFrontmatterDigest    = 86  // [86,118)
	roSegmentTableDigest   = 118 // [118,150)
	roTCRoot               = 150 // [150,182)
	roParentStateID        = 182 // [182,214)
	roRetentionPoint       = 214 // [214,216)
	roCompactionGeneration = 216 // [216,220)
	roIndexRoute           = 220 // [220,252)
	roStructureDigest      = 252 // [252,284)
	roRingReservedStart    = 284 // [284,480)
	roRecordDigest         = 480 // [480,512)
)

// RingMagic is the ring-slot self-identification constant (contracts/
// container.abnf S3): ASCII "PDR1".
var RingMagic = [4]byte{'P', 'D', 'R', '1'}

// StateID is a document-state identifier (FR-003): the state's T_C root,
// 32 octets. A root state's parent-state-id is the zero value (32 zero
// octets, contracts/container.abnf S3 parent-state-id NORMATIVE comment).
type StateID [32]byte

// IndexRouteBuckets is the fixed bucket count of a ring slot's index-route
// routing table (contracts/container.abnf S3 index-route NORMATIVE comment).
const IndexRouteBuckets = 16

// CommitRingRecord is one commit-ring slot's logical fields (contracts/
// container.abnf S3 ring-fields). Ring-magic and record-digest are wire
// artifacts Encode produces and are not struct fields, matching Header's
// treatment of its own magic and header-digest.
//
// Encode/Decode here perform no self-digest or ring-magic validation:
// per tasks.md T-0007/T-0009, this struct defines the fixed layout only,
// and self-digest validation (detecting a torn or partial slot write) is
// a distinct, later capability.
type CommitRingRecord struct {
	Sequence             uint64
	LedgerLength         uint64
	LedgerRoot           [32]byte
	SegmentCount         uint16
	StateID              StateID
	FrontmatterDigest    [32]byte
	SegmentTableDigest   [32]byte
	TCRoot               [32]byte
	ParentStateID        StateID
	RetentionPoint       uint16
	CompactionGeneration uint32
	IndexRoute           [IndexRouteBuckets]uint16
	StructureDigest      [32]byte
}

var (
	// ErrRingSlotTruncated is returned when fewer than RingSlotSize octets are available.
	ErrRingSlotTruncated = errors.New("container: ring slot truncated, need 512 octets")
	// ErrCommitRingTruncated is returned when fewer than CommitRingSize octets are available.
	ErrCommitRingTruncated = errors.New("container: commit ring truncated, need 3584 octets")
)

// Encode writes r's canonical 512-octet slot encoding to dst, which must
// be at least RingSlotSize octets. It returns dst[:RingSlotSize]. Ring-
// reserved [284,480) is always zero-filled; record-digest [480,512) is
// always computed over the preceding 480 octets (self-digesting, per
// contracts/container.abnf S3 record-digest NORMATIVE comment), so every
// slot Encode produces is well-formed even though this task defers
// verifying that digest on decode to T-0009.
func (r *CommitRingRecord) Encode(dst []byte) []byte {
	if cap(dst) < RingSlotSize {
		dst = make([]byte, RingSlotSize)
	} else {
		dst = dst[:RingSlotSize]
		for i := range dst {
			dst[i] = 0
		}
	}
	copy(dst[roRingMagic:roRingMagic+4], RingMagic[:])
	binary.BigEndian.PutUint64(dst[roSequence:], r.Sequence)
	binary.BigEndian.PutUint64(dst[roLedgerLength:], r.LedgerLength)
	copy(dst[roLedgerRoot:roLedgerRoot+32], r.LedgerRoot[:])
	binary.BigEndian.PutUint16(dst[roSegmentCount:], r.SegmentCount)
	copy(dst[roStateID:roStateID+32], r.StateID[:])
	copy(dst[roFrontmatterDigest:roFrontmatterDigest+32], r.FrontmatterDigest[:])
	copy(dst[roSegmentTableDigest:roSegmentTableDigest+32], r.SegmentTableDigest[:])
	copy(dst[roTCRoot:roTCRoot+32], r.TCRoot[:])
	copy(dst[roParentStateID:roParentStateID+32], r.ParentStateID[:])
	binary.BigEndian.PutUint16(dst[roRetentionPoint:], r.RetentionPoint)
	binary.BigEndian.PutUint32(dst[roCompactionGeneration:], r.CompactionGeneration)
	for i, v := range r.IndexRoute {
		binary.BigEndian.PutUint16(dst[roIndexRoute+2*i:], v)
	}
	copy(dst[roStructureDigest:roStructureDigest+32], r.StructureDigest[:])
	// dst[roRingReservedStart:roRecordDigest] (ring-reserved) is already
	// zero from the fill above.
	digest := sha256.Sum256(dst[:roRecordDigest])
	copy(dst[roRecordDigest:RingSlotSize], digest[:])
	return dst
}

// DecodeCommitRingRecord decodes one 512-octet ring slot's fields at their
// fixed offsets. It performs no self-digest or ring-magic validation
// (T-0009's scope): an all-zero (never-written) slot decodes without
// error, same as a written one, since distinguishing the two is exactly
// what self-digest validation is for.
func DecodeCommitRingRecord(src []byte) (*CommitRingRecord, error) {
	if len(src) < RingSlotSize {
		return nil, ErrRingSlotTruncated
	}
	src = src[:RingSlotSize]

	r := &CommitRingRecord{
		Sequence:             binary.BigEndian.Uint64(src[roSequence:]),
		LedgerLength:         binary.BigEndian.Uint64(src[roLedgerLength:]),
		SegmentCount:         binary.BigEndian.Uint16(src[roSegmentCount:]),
		RetentionPoint:       binary.BigEndian.Uint16(src[roRetentionPoint:]),
		CompactionGeneration: binary.BigEndian.Uint32(src[roCompactionGeneration:]),
	}
	copy(r.LedgerRoot[:], src[roLedgerRoot:roLedgerRoot+32])
	copy(r.StateID[:], src[roStateID:roStateID+32])
	copy(r.FrontmatterDigest[:], src[roFrontmatterDigest:roFrontmatterDigest+32])
	copy(r.SegmentTableDigest[:], src[roSegmentTableDigest:roSegmentTableDigest+32])
	copy(r.TCRoot[:], src[roTCRoot:roTCRoot+32])
	copy(r.ParentStateID[:], src[roParentStateID:roParentStateID+32])
	for i := range r.IndexRoute {
		r.IndexRoute[i] = binary.BigEndian.Uint16(src[roIndexRoute+2*i:])
	}
	copy(r.StructureDigest[:], src[roStructureDigest:roStructureDigest+32])

	return r, nil
}

// EncodeCommitRing writes all 7 ring slots' canonical encoding to dst,
// which must be at least CommitRingSize octets, in slot order. It returns
// dst[:CommitRingSize].
func EncodeCommitRing(ring *[CommitRingSlots]CommitRingRecord, dst []byte) []byte {
	if cap(dst) < CommitRingSize {
		dst = make([]byte, CommitRingSize)
	} else {
		dst = dst[:CommitRingSize]
	}
	for i := range ring {
		ring[i].Encode(dst[i*RingSlotSize : (i+1)*RingSlotSize])
	}
	return dst
}

// DecodeCommitRing decodes all 7 ring slots from the leading CommitRingSize
// octets of src, at their fixed offsets.
func DecodeCommitRing(src []byte) (*[CommitRingSlots]CommitRingRecord, error) {
	if len(src) < CommitRingSize {
		return nil, ErrCommitRingTruncated
	}
	var ring [CommitRingSlots]CommitRingRecord
	for i := range ring {
		rec, err := DecodeCommitRingRecord(src[i*RingSlotSize : (i+1)*RingSlotSize])
		if err != nil {
			return nil, fmt.Errorf("container: ring slot %d: %w", i, err)
		}
		ring[i] = *rec
	}
	return &ring, nil
}
