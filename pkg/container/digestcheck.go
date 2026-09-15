// Digest-mismatch-aborts-before-decode (FR-104; contracts/container.abnf
// S6.3 segment-self-digest NORMATIVE comment: "these are the two
// independent inventories FR-104/105 require to agree"): a stored unit
// whose declared digest disagrees with its own octets must be rejected
// before any attempt to decode its body, naming the offending unit and
// both digest values.
package container

import (
	"crypto/sha256"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// DigestMismatchError is returned when a stored unit's declared digest
// (its SegmentTableSlot.Digest) disagrees with the digest recomputed
// fresh over its own octets. Ordinal identifies the offending
// SegmentTableSlot (its index in the segment table; CON-008's opaque
// unit-id type is a later milestone's task, not yet available to name it
// by).
type DigestMismatchError struct {
	Ordinal  int
	Expected pdlfmt.Digest256
	Actual   pdlfmt.Digest256
}

func (e *DigestMismatchError) Error() string {
	return fmt.Sprintf("container: segment ordinal %d digest mismatch: expected %x, actual %x", e.Ordinal, e.Expected, e.Actual)
}

// VerifySegmentDigest recomputes SHA-256 over segmentOctets (the
// segment's complete octets, container.abnf S5 slot-digest NORMATIVE
// comment) and compares it against declared, the digest the segment's
// SegmentTableSlot carries. It performs no decode of segmentOctets: the
// comparison is a fixed-size digest equality, independent of whatever
// record structure segmentOctets holds.
func VerifySegmentDigest(ordinal int, declared pdlfmt.Digest256, segmentOctets []byte) (pdlfmt.Digest256, error) {
	actual := pdlfmt.Digest256(sha256.Sum256(segmentOctets))
	if actual != declared {
		return actual, &DigestMismatchError{Ordinal: ordinal, Expected: declared, Actual: actual}
	}
	return actual, nil
}

// DecodeSegmentChecked verifies segmentOctets against slot's declared
// digest and, only if that succeeds, invokes decode. FR-104's "abort
// before decoding that unit" is enforced structurally here: decode is
// never reachable on a digest-mismatched segment, because
// VerifySegmentDigest's error returns before decode is ever called.
func DecodeSegmentChecked(ordinal int, slot SegmentTableSlot, segmentOctets []byte, decode func([]byte) error) error {
	if _, err := VerifySegmentDigest(ordinal, slot.Digest, segmentOctets); err != nil {
		return err
	}
	return decode(segmentOctets)
}
