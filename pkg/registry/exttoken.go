// Package registry implements Protodoc's extension-token space and the
// ExtensionEnvelope carrier for non-core constructs. The extension-token space
// (CON-020, container.abnf S5.2) is partitioned into registered, owner-scoped,
// and permanently-retired sets; a retired token is never reissued, and no
// experimental or unregistered prefix exists anywhere in the space.
package registry

import "encoding/binary"

// ExtToken is an 8-octet extension token (container.abnf S5.2): a 4-octet
// big-endian owner-id followed by a 4-octet big-endian owner-local-seq. Two
// tokens are equal iff their 8 octets are equal; the type is a fixed array so
// comparison and map-keying are exact-octet.
type ExtToken [8]byte

// OwnerID returns the token's 4-octet owner-id (the high 4 octets).
func (t ExtToken) OwnerID() uint32 { return binary.BigEndian.Uint32(t[0:4]) }

// OwnerLocalSeq returns the token's 4-octet owner-local-seq (the low 4 octets).
func (t ExtToken) OwnerLocalSeq() uint32 { return binary.BigEndian.Uint32(t[4:8]) }

// NewExtToken builds an ExtToken from an owner-id and owner-local-seq.
func NewExtToken(ownerID, ownerLocalSeq uint32) ExtToken {
	var t ExtToken
	binary.BigEndian.PutUint32(t[0:4], ownerID)
	binary.BigEndian.PutUint32(t[4:8], ownerLocalSeq)
	return t
}

// TokenPartition is the CON-020 partition an ext-token's owner-id falls into.
type TokenPartition int

const (
	// PartitionReservedInvalid is owner-id 0x00000000: never assigned, marks
	// an invalid/absent token.
	PartitionReservedInvalid TokenPartition = iota
	// PartitionRegistered is owner-id 0x00000001..0x7FFFFFFF: issued by the
	// Protodoc token registry (CON-020; the registry-turnaround SLA is a
	// separate governance requirement owned by T-0061).
	PartitionRegistered
	// PartitionOwnerScoped is owner-id 0x80000000..0xFFFFFFFE: self-issued by
	// the owner without a registry round trip.
	PartitionOwnerScoped
	// PartitionRetired is owner-id 0xFFFFFFFF: the permanently-retired marker
	// namespace; only ever a tombstone, never a live extension.
	PartitionRetired
)

func (p TokenPartition) String() string {
	switch p {
	case PartitionReservedInvalid:
		return "reserved-invalid"
	case PartitionRegistered:
		return "registered"
	case PartitionOwnerScoped:
		return "owner-scoped"
	case PartitionRetired:
		return "retired"
	default:
		return "unknown"
	}
}

// Owner-id partition boundaries (container.abnf S5.2), as named constants so
// the boundaries are checkable rather than magic numbers.
const (
	OwnerIDReservedInvalid = 0x00000000
	OwnerIDRegisteredLo    = 0x00000001
	OwnerIDRegisteredHi    = 0x7FFFFFFF
	OwnerIDOwnerScopedLo   = 0x80000000
	OwnerIDOwnerScopedHi   = 0xFFFFFFFE
	OwnerIDRetired         = 0xFFFFFFFF
)

// PartitionOf returns the CON-020 partition of ownerID. The whole uint32
// range is covered by exactly one partition, with no gap and no experimental
// or unregistered prefix (CON-020).
func PartitionOf(ownerID uint32) TokenPartition {
	switch {
	case ownerID == OwnerIDReservedInvalid:
		return PartitionReservedInvalid
	case ownerID >= OwnerIDRegisteredLo && ownerID <= OwnerIDRegisteredHi:
		return PartitionRegistered
	case ownerID >= OwnerIDOwnerScopedLo && ownerID <= OwnerIDOwnerScopedHi:
		return PartitionOwnerScoped
	default: // ownerID == OwnerIDRetired (0xFFFFFFFF)
		return PartitionRetired
	}
}

// Partition returns the token's partition (by its owner-id).
func (t ExtToken) Partition() TokenPartition { return PartitionOf(t.OwnerID()) }

// IsLiveExtension reports whether t may name a live extension: only registered
// and owner-scoped tokens can. A reserved-invalid or retired token never
// names a live extension (CON-020).
func (t ExtToken) IsLiveExtension() bool {
	switch t.Partition() {
	case PartitionRegistered, PartitionOwnerScoped:
		return true
	default:
		return false
	}
}
