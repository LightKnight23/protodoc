package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/validate"
)

// TestFR_015_FallbackReachabilityExclusion is T-0056's named integration test
// (FR-015). It exercises the fallback reachability-exclusion audit over the
// validate-package reference-graph edge model: a fallback unit must be
// reachable ONLY as this envelope's own ext-envelope-fallback-ref edge, with
// no other live inbound path.
func TestFR_015_FallbackReachabilityExclusion(t *testing.T) {
	envelopeUnit := fixtureUnitID(0x01)
	fallback := fixtureUnitID(0x02)
	other := fixtureUnitID(0x03)

	e := ExtEnvelope{
		Tok:         NewExtToken(0x00000055, 1),
		Disposition: DispositionDegrade,
		FallbackRef: fallback,
	}

	ownEdge := validate.Edge{Kind: validate.EdgeExtFallbackRef, From: envelopeUnit, To: fallback}

	// (1) Only the envelope's own fallback edge targets the fallback: passes.
	if err := AuditFallbackReachability(e, envelopeUnit, []validate.Edge{ownEdge}); err != nil {
		t.Fatalf("exclusive fallback rejected: %v", err)
	}

	// (2) An additional annotation-anchor edge into the fallback is a
	// disqualifying live path.
	annEdge := validate.Edge{Kind: validate.EdgeAnnotationAnchor, From: other, To: fallback}
	err := AuditFallbackReachability(e, envelopeUnit, []validate.Edge{ownEdge, annEdge})
	if err == nil {
		t.Fatal("fallback with an extra annotation-anchor path was accepted")
	}
	var re *FallbackReachabilityError
	if !errors.As(err, &re) {
		t.Fatalf("expected *FallbackReachabilityError, got %T", err)
	}
	if len(re.OtherInbound) != 1 || re.OtherInbound[0].Kind != validate.EdgeAnnotationAnchor {
		t.Errorf("expected the annotation edge listed as other-inbound, got %+v", re.OtherInbound)
	}

	// (3) A fallback edge from a DIFFERENT envelope is also a disqualifying
	// live path (the unit is shared as two envelopes' fallback).
	otherEnvEdge := validate.Edge{Kind: validate.EdgeExtFallbackRef, From: other, To: fallback}
	if err := AuditFallbackReachability(e, envelopeUnit, []validate.Edge{ownEdge, otherEnvEdge}); err == nil {
		t.Fatal("fallback shared with another envelope was accepted")
	}

	// (4) The envelope's own fallback edge missing entirely is a rejection
	// (the fallback is declared but not wired as a fallback edge).
	if err := AuditFallbackReachability(e, envelopeUnit, []validate.Edge{annEdge}); err == nil {
		t.Fatal("fallback with no own fallback edge was accepted")
	}

	// (5) A refuse envelope with no fallback has nothing to audit: passes.
	eRefuse := ExtEnvelope{Tok: e.Tok, Disposition: DispositionRefuse, FallbackRef: Zero16}
	if err := AuditFallbackReachability(eRefuse, envelopeUnit, nil); err != nil {
		t.Fatalf("refuse envelope with no fallback rejected: %v", err)
	}

	// (6) Stray edges not targeting the fallback are ignored.
	stray := validate.Edge{Kind: validate.EdgeStructuralMove, From: other, To: fixtureUnitID(0x09)}
	if err := AuditFallbackReachability(e, envelopeUnit, []validate.Edge{ownEdge, stray}); err != nil {
		t.Fatalf("stray non-targeting edge caused a false rejection: %v", err)
	}
}
