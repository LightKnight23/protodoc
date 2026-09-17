package registry

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
	"Protodoc/pkg/validate"
)

// Fallback reachability exclusion (T-0056, FR-015; document.abnf S6). An
// ext-fallback-ref's target must have NO other live path into the document
// than as this envelope's fallback: the fallback/primary exclusion is a
// genuine, checked relationship, not a declared-but-unverified role. If the
// fallback unit were also reachable by any other live reference -- a cross
// reference, an annotation anchor, a structural move, or another envelope's
// fallback -- then the "fallback" content would in fact be live primary
// content, defeating the point of degrading to it. So the audit requires the
// UNIQUE inbound reference to the fallback unit to be exactly this envelope's
// ext-envelope-fallback-ref edge.

// FallbackReachabilityError is the FR-015 rejection: the fallback unit has an
// inbound reference other than this envelope's own fallback edge (or has none
// at all, which means the fallback edge itself is missing from the graph).
type FallbackReachabilityError struct {
	Tok           ExtToken
	Fallback      pdlfmt.UnitID
	OtherInbound  []validate.Edge // the disqualifying non-fallback inbound edges
	FallbackEdges int             // how many ext-envelope-fallback-ref edges point at Fallback
}

func (e *FallbackReachabilityError) Error() string {
	if e.FallbackEdges == 0 {
		return fmt.Sprintf("registry: ext-tok %x fallback %x has no ext-envelope-fallback-ref edge into the graph (FR-015)", e.Tok, e.Fallback)
	}
	return fmt.Sprintf("registry: ext-tok %x fallback %x is reachable by %d non-fallback inbound reference(s); a fallback must have no other live path (FR-015)",
		e.Tok, e.Fallback, len(e.OtherInbound))
}

// AuditFallbackReachability checks that the envelope e's fallback unit is
// reachable ONLY as this envelope's fallback. inbound is the complete set of
// reference-graph edges in the document whose target is e.FallbackRef (the
// caller collects these; every edge here has To == e.FallbackRef). The audit
// passes iff exactly one inbound edge exists and it is an
// ext-envelope-fallback-ref edge originating from this envelope. It applies
// only when a fallback is present (ignore/degrade with non-zero16 ref); a
// refuse envelope with no fallback has nothing to audit and passes.
//
// envelopeUnit is this envelope's own unit id (the From of its fallback edge),
// so a fallback edge from a DIFFERENT envelope is treated as a disqualifying
// other-inbound path, not as this envelope's own.
func AuditFallbackReachability(e ExtEnvelope, envelopeUnit pdlfmt.UnitID, inbound []validate.Edge) error {
	if e.FallbackRef == Zero16 {
		return nil // no fallback to audit (refuse, or an absent-but-that-is-T-0055's-concern)
	}

	var ownFallbackEdges int
	var fallbackEdgeTotal int
	var other []validate.Edge
	for _, edge := range inbound {
		if edge.To != e.FallbackRef {
			// Defensive: caller should pre-filter to edges targeting the
			// fallback, but ignore any stray non-targeting edge.
			continue
		}
		if edge.Kind == validate.EdgeExtFallbackRef {
			fallbackEdgeTotal++
			if edge.From == envelopeUnit {
				ownFallbackEdges++
				continue
			}
			// A fallback edge from another envelope is another live path.
			other = append(other, edge)
			continue
		}
		// Any non-fallback edge into the unit is a disqualifying live path.
		other = append(other, edge)
	}

	if ownFallbackEdges == 0 || len(other) > 0 {
		return &FallbackReachabilityError{
			Tok:           e.Tok,
			Fallback:      e.FallbackRef,
			OtherInbound:  other,
			FallbackEdges: fallbackEdgeTotal,
		}
	}
	return nil
}
