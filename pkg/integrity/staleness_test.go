package integrity

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_069_SignedPresentationExemptFromStaleness is T-0163's named unit test
// (FR-069). A presentation artefact bound to a signed state is exempt from
// staleness refusal relative to the current state, even after the document is
// edited so the artefact no longer derives from the current state. An unbound
// presentation is refused as stale exactly when it no longer matches the
// current state.
func TestFR_069_SignedPresentationExemptFromStaleness(t *testing.T) {
	p := pdlfmt.UnitID{0x01}

	// Bound-by-signature presentation: exempt regardless of whether it still
	// derives from the current state (an edit does not stale it).
	if PresentationStaleRefused(PresentationStalenessInput{Presentation: p, DerivedFromCurrentState: false, BoundBySignature: true}) {
		t.Error("a signature-bound presentation must be exempt from staleness refusal after an edit (FR-069)")
	}
	if PresentationStaleRefused(PresentationStalenessInput{Presentation: p, DerivedFromCurrentState: true, BoundBySignature: true}) {
		t.Error("a signature-bound presentation must never be refused as stale")
	}

	// Unbound presentation: refused as stale iff it no longer derives from the
	// current state.
	if !PresentationStaleRefused(PresentationStalenessInput{Presentation: p, DerivedFromCurrentState: false, BoundBySignature: false}) {
		t.Error("an unbound presentation not matching the current state must be refused as stale")
	}
	if PresentationStaleRefused(PresentationStalenessInput{Presentation: p, DerivedFromCurrentState: true, BoundBySignature: false}) {
		t.Error("an unbound presentation matching the current state must not be refused as stale")
	}
}
