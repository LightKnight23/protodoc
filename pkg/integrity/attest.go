// ATTEST-segment exclusion (T-0138, NFR-007/FR-002). A single shared
// predicate, IsAttestTyped, determines whether a segment is ATTEST-typed
// (SIGNATURE, RESCIND_RESIGN, ATTESTATION_EVIDENCE frames). Three guarantees
// route through this one function so they can never silently drift apart:
//  1. T_C's subtree set never includes ATTEST-segment content (state
//     identity excludes signatures);
//  2. the no-op-save octet-identity comparison excludes ATTEST octets
//     (signature/time-attestation/revocation octets never count as "did
//     anything change", NFR-003/NFR-007);
//  3. CoverageDescriptor construction excludes ATTEST ordinals (closing
//     FR-002's self-coverage circularity).
package integrity

import "Protodoc/pkg/container"

// IsAttestTyped reports whether slot is ATTEST-typed
// (SegmentTypeAttest == 4). This is THE single definition of "is this an
// attestation segment" for the whole codebase; the three call sites above
// must invoke it rather than re-checking slot-segment-type locally.
func IsAttestTyped(slot container.SegmentTableSlot) bool {
	return slot.SegmentType == container.SegmentTypeAttest
}

// NonAttestOrdinals returns the storage ordinals (slot indices) of every
// non-ATTEST, populated (non-unused) segment in slots, in ascending order.
// It is the shared basis for T_C's covered set, the no-op-save comparator's
// "which extents count", and CoverageDescriptor's coverable universe -- all
// via the single IsAttestTyped predicate.
func NonAttestOrdinals(slots []container.SegmentTableSlot) []uint16 {
	var out []uint16
	for i, s := range slots {
		if s.SegmentType == container.SegmentTypeUnused {
			continue
		}
		if IsAttestTyped(s) {
			continue
		}
		out = append(out, uint16(i))
	}
	return out
}

// AttestOrdinals returns the storage ordinals of every ATTEST-typed segment
// in slots, in ascending order -- the ordinals that must never appear in a
// CoverageDescriptor and never participate in state identity.
func AttestOrdinals(slots []container.SegmentTableSlot) []uint16 {
	var out []uint16
	for i, s := range slots {
		if IsAttestTyped(s) {
			out = append(out, uint16(i))
		}
	}
	return out
}
