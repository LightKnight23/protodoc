// Presentation staleness exemption (T-0163, FR-069; document.abnf S7.4). A
// presentation artefact's general contract is that it must reflect the current
// state; a normally-derived artefact whose inputs no longer match the current
// state is STALE and refused. But a presentation artefact BOUND to a signed
// state (named by a signature's sig-presentation-ref and committed into
// signed_object) is exempt from that staleness refusal relative to the current
// state: its inputs are the SIGNED state, not the current one, so the general
// stale-artefact refusal would destroy the attested rendering the moment
// anyone edits the document.
package integrity

import "Protodoc/pkg/pdlfmt"

// PresentationStalenessInput describes what a reader knows when deciding
// whether a presentation artefact is stale.
type PresentationStalenessInput struct {
	// Presentation is the presentation artefact's own slot digest.
	Presentation pdlfmt.UnitID
	// DerivedFromCurrentState is true iff the artefact was derived from (and
	// still matches) the CURRENT document state.
	DerivedFromCurrentState bool
	// BoundBySignature is true iff a signature binds this presentation to a
	// signed state (its sig-presentation-ref names it).
	BoundBySignature bool
}

// PresentationStaleRefused reports whether a reader must refuse a presentation
// artefact as stale (FR-069). A presentation bound by a signature is NEVER
// refused as stale relative to the current state -- its inputs are the signed
// state, and refusing it would destroy the attested rendering. An unbound
// presentation is refused as stale exactly when it no longer derives from the
// current state.
func PresentationStaleRefused(in PresentationStalenessInput) bool {
	if in.BoundBySignature {
		return false // exempt from staleness refusal (FR-069)
	}
	return !in.DerivedFromCurrentState
}
