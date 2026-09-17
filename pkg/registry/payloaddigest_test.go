package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition is T-0053's
// named unit test (FR-012/FR-107). It proves ext-payload-digest is verified
// against the referenced RESOURCE segment's slot digest and a mismatch is
// rejected BEFORE any disposition is applied: the verification result depends
// only on the digest comparison and the resolver, never on ext-disposition, so
// no disposition value (not even ignore) can let a tampered or dangling payload
// through the integrity gate.
func TestFR_012_ExtPayloadDigestMismatchRejectedBeforeDisposition(t *testing.T) {
	payloadRef := fixtureUnitID(0x50)
	trueDigest := fixtureDigest(0x60)

	// Resolver mapping the payload ref to the segment's actual slot digest.
	resolve := func(ref pdlfmt.UnitID) (pdlfmt.Digest256, bool) {
		if ref == payloadRef {
			return trueDigest, true
		}
		return pdlfmt.Digest256{}, false
	}

	base := ExtEnvelope{
		Tok:           NewExtToken(0x00000123, 1),
		PayloadLength: 42,
		PayloadRef:    payloadRef,
		PayloadDigest: trueDigest,
		FallbackRef:   Zero16,
	}

	// (1) Matching digest passes, for EVERY disposition value -- the check is
	// disposition-independent.
	for _, disp := range []ExtDisposition{DispositionIgnore, DispositionDegrade, DispositionRefuse} {
		e := base
		e.Disposition = disp
		if err := VerifyPayloadDigest(e, resolve); err != nil {
			t.Errorf("matching digest with disposition %v rejected: %v", disp, err)
		}
	}

	// (2) A mismatched digest is rejected for EVERY disposition, including
	// ignore -- proving the digest gate precedes and is independent of
	// disposition (a reader cannot "ignore" its way past a bad payload).
	for _, disp := range []ExtDisposition{DispositionIgnore, DispositionDegrade, DispositionRefuse} {
		e := base
		e.Disposition = disp
		e.PayloadDigest = fixtureDigest(0x99) // != trueDigest
		if err := VerifyPayloadDigest(e, resolve); !errors.Is(err, ErrExtPayloadDigestMismatch) {
			t.Errorf("mismatched digest with disposition %v: err = %v, want ErrExtPayloadDigestMismatch", disp, err)
		}
	}

	// (3) An unresolved payload ref is rejected distinctly, again regardless
	// of disposition.
	e := base
	e.PayloadRef = fixtureUnitID(0x77) // resolver returns ok=false
	e.Disposition = DispositionIgnore
	if err := VerifyPayloadDigest(e, resolve); !errors.Is(err, ErrExtPayloadRefUnresolved) {
		t.Errorf("unresolved payload ref: err = %v, want ErrExtPayloadRefUnresolved", err)
	}
}
