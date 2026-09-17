// Package integrity implements Protodoc's two Merkle integrity trees -- T_S
// (storage integrity, arity-16/depth-4 over the SegmentTable) and T_C
// (content commitment / redaction, arity-16/depth<=5 over content-model
// records) -- their roots, the ATTEST-exclusion predicate, the
// CoverageDescriptor wire form, and the structure_digest algorithm
// (integrity.abnf S2-S4). Every value is recomputed fresh from the current
// file's octets; a stored digest is never trusted (data-model.md S2.11).
//
// This file defines the shared domain-tag registry and ABSENT_CHILD_DIGEST
// (T-0129, FR-003/TR-009). A single shared registry (integrity.abnf S2's
// domain-tag table) guarantees no two hash-preimage classes anywhere in the
// format can collide even if their remaining octets coincide, and that
// ABSENT_CHILD_DIGEST (whose preimage is the single reserved octet 0x00) can
// never equal a genuine node/leaf/top-level digest, since every real
// preimage starts with a different, nonzero domain tag.
package integrity

import "crypto/sha256"

// Domain tags: the first octet of every hash preimage defined in
// integrity.abnf S2-S3. This is the single shared, non-overlapping registry.
const (
	// DomainAbsentChild (0x00) is the reserved filler tag; its preimage is
	// the single octet 0x00 (integrity.abnf S2.3). No real digest starts
	// with it.
	DomainAbsentChild byte = 0x00
	// DomainTSInternal (0x01): T_S internal node, preimage 0x01 || 16 child
	// digests (integrity.abnf S2.1).
	DomainTSInternal byte = 0x01
	// DomainTCLeafRedactable (0x02): T_C redactable leaf, preimage
	// 0x02 || salt(32) || canon(subtree) (integrity.abnf S2.2).
	DomainTCLeafRedactable byte = 0x02
	// DomainSignedObject (0x04): signed_object preimage (integrity.abnf S3.2).
	DomainSignedObject byte = 0x04
	// DomainTCLeafNonredactable (0x07): T_C non-redactable leaf, preimage
	// 0x07 || canon(subtree) (integrity.abnf S2.2).
	DomainTCLeafNonredactable byte = 0x07
	// DomainTCInternal (0x08): T_C internal node, preimage 0x08 || 16 child
	// digests (integrity.abnf S2.2).
	DomainTCInternal byte = 0x08
	// DomainSeveranceCommitment (0x09): severance_commitment (integrity.abnf
	// S7.2).
	DomainSeveranceCommitment byte = 0x09
	// DomainStructureDigest (0x0A): structure_digest preimage (integrity.abnf
	// S3.1).
	DomainStructureDigest byte = 0x0A
)

// Digest is a 32-octet SHA-256 output (a node, leaf, or top-level digest).
type Digest [32]byte

// AbsentChildDigest is SHA-256(0x00): the fixed filler value placed at every
// T_S or T_C internal-node child position that has no corresponding real
// child, at every level of the tree (integrity.abnf S2.3). It is computed at
// package init from its one-octet preimage, and pinned/asserted in the test.
var AbsentChildDigest = func() Digest {
	var d Digest
	sum := sha256.Sum256([]byte{DomainAbsentChild})
	copy(d[:], sum[:])
	return d
}()

// realPreimageTags is the set of domain tags that begin a genuine
// (non-filler) hash preimage in this format. ABSENT_CHILD_DIGEST's tag
// (0x00) is deliberately excluded, which is what makes the filler value
// uncollidable with any real digest.
var realPreimageTags = []byte{
	DomainTSInternal,          // 0x01
	DomainTCLeafRedactable,    // 0x02
	DomainSignedObject,        // 0x04
	DomainTCLeafNonredactable, // 0x07
	DomainTCInternal,          // 0x08
	DomainSeveranceCommitment, // 0x09
	DomainStructureDigest,     // 0x0A
}
