package registry

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ext-payload-digest verification (T-0053, FR-012/FR-107; document.abnf S6).
// ext-payload-digest MUST equal the referenced RESOURCE segment's own
// SegmentTableSlot digest. A reader that does not implement the ext-tok still
// verifies this digest and MUST reject a mismatch BEFORE applying any
// disposition -- the payload octets stay opaque, and the digest check is the
// one thing every reader can always do. Ordering matters: the digest is the
// integrity gate that precedes any disposition-driven behaviour, so a tampered
// or dangling payload can never reach ignore/degrade/refuse handling.

// ErrExtPayloadRefUnresolved is returned when the ext-payload-ref does not
// resolve to a known RESOURCE segment slot digest.
var ErrExtPayloadRefUnresolved = errors.New("registry: ext-payload-ref does not resolve to a RESOURCE segment")

// ErrExtPayloadDigestMismatch is returned when ext-payload-digest does not
// equal the referenced segment's SegmentTableSlot digest.
var ErrExtPayloadDigestMismatch = errors.New("registry: ext-payload-digest does not match the referenced segment's slot digest")

// SlotDigestResolver returns the SegmentTableSlot digest of the RESOURCE
// segment addressed by a unit-id (ext-payload-ref), and ok=false if no such
// segment exists. The pipeline supplies this from the file under validation so
// this package needs no direct segment-table access of its own.
type SlotDigestResolver func(payloadRef pdlfmt.UnitID) (slotDigest pdlfmt.Digest256, ok bool)

// VerifyPayloadDigest checks ext-payload-digest against the referenced
// segment's slot digest, the integrity gate that MUST pass before any
// disposition is applied (FR-012/FR-107). It returns nil on a match,
// ErrExtPayloadRefUnresolved if the ref does not resolve, or
// ErrExtPayloadDigestMismatch on a digest disagreement. It never inspects the
// payload octets themselves (they remain opaque) and never consults
// ext-disposition (the digest check strictly precedes disposition).
func VerifyPayloadDigest(e ExtEnvelope, resolve SlotDigestResolver) error {
	slotDigest, ok := resolve(e.PayloadRef)
	if !ok {
		return fmt.Errorf("%w: %x", ErrExtPayloadRefUnresolved, e.PayloadRef)
	}
	if slotDigest != e.PayloadDigest {
		return fmt.Errorf("%w: envelope declares %x, segment slot has %x", ErrExtPayloadDigestMismatch, e.PayloadDigest, slotDigest)
	}
	return nil
}
