package validate

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_108_UnresolvedOrAmbiguousReferenceRejected is T-0117's named
// conformance test. A dangling xref, a dangling annotation anchor, and a
// reference resolving to two units are all rejected citing FR-108; a
// fully-resolved valid document passes.
func TestFR_108_UnresolvedOrAmbiguousReferenceRejected(t *testing.T) {
	id := func(b byte) pdlfmt.UnitID { var u pdlfmt.UnitID; u[0] = b; return u }
	from := id(0xF0)
	present := PresenceCount{
		id(1): 1,
		id(2): 1,
		id(9): 2, // an id that (illegally) appears twice
	}

	// Fully resolved: every target present exactly once.
	ok := []Reference{
		{Kind: RefXrefTarget, From: from, Target: id(1)},
		{Kind: RefAnnotationStart, From: from, Target: id(2)},
	}
	if err := CheckReferences(ok, present); err != nil {
		t.Fatalf("fully-resolved references rejected: %v", err)
	}

	// Dangling xref target (resolves to zero).
	danglingXref := []Reference{{Kind: RefXrefTarget, From: from, Target: id(3)}}
	err := CheckReferences(danglingXref, present)
	var rerr *ReferenceResolutionError
	if !errors.As(err, &rerr) || rerr.Resolved != 0 || rerr.Kind != RefXrefTarget {
		t.Fatalf("dangling xref: got %v, want dangling xref-target error", err)
	}

	// Dangling annotation anchor.
	danglingAnchor := []Reference{{Kind: RefAnnotationEnd, From: from, Target: id(4)}}
	if err := CheckReferences(danglingAnchor, present); !isRefErr(err, RefAnnotationEnd, 0) {
		t.Fatalf("dangling anchor: got %v, want dangling annotation-end error", err)
	}

	// Ambiguous reference (resolves to two units).
	ambiguous := []Reference{{Kind: RefIdentity, From: from, Target: id(9)}}
	if err := CheckReferences(ambiguous, present); !isRefErr(err, RefIdentity, 2) {
		t.Fatalf("ambiguous reference: got %v, want ambiguous identity error", err)
	}

	// The check reports the FIRST failing reference and does not repair.
	mixed := []Reference{
		{Kind: RefXrefTarget, From: from, Target: id(1)}, // ok
		{Kind: RefXrefTarget, From: from, Target: id(3)}, // dangling -> reported
		{Kind: RefIdentity, From: from, Target: id(9)},   // ambiguous, not reached
	}
	if err := CheckReferences(mixed, present); !isRefErr(err, RefXrefTarget, 0) {
		t.Fatalf("mixed: got %v, want first (dangling xref) error", err)
	}
}

func isRefErr(err error, kind RefKind, resolved int) bool {
	var e *ReferenceResolutionError
	return errors.As(err, &e) && e.Kind == kind && e.Resolved == resolved
}
