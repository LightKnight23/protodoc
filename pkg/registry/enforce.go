package registry

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// Disposition runtime enforcement (T-0059, FR-107; document.abnf S6). IF a
// conforming reader encounters a construct it does not implement, THEN it must
// APPLY that construct's declared disposition and must NOT silently omit,
// approximate, or render around it. This models the single defined behaviour a
// previous-generation reader must exhibit, so its observable action equals the
// declared disposition in 100% of cases.
//
// The enforcement runs the checks in their required order: (1) the
// ext-payload-digest integrity gate (T-0053), which every reader performs
// before anything else; (2) ext-disposition validity (T-0054); (3) for
// ignore/degrade, ext-fallback-ref presence (T-0055). Only after all pass does
// the reader enact the disposition. A failure at any gate is a structural
// rejection, never a silent render-around.

// RenderAction is the single observable behaviour a reader must exhibit for an
// unimplemented construct, determined solely by its declared disposition.
type RenderAction int

const (
	// ActionRenderAbsent: render as if the construct were absent (ignore).
	ActionRenderAbsent RenderAction = iota
	// ActionRenderFallback: render the ext-fallback-ref construct (degrade).
	ActionRenderFallback
	// ActionRefuseDocument: refuse to render the whole document (refuse).
	ActionRefuseDocument
)

func (a RenderAction) String() string {
	switch a {
	case ActionRenderAbsent:
		return "render-absent"
	case ActionRenderFallback:
		return "render-fallback"
	case ActionRefuseDocument:
		return "refuse-document"
	default:
		return "unknown"
	}
}

// EnforceDisposition returns the single RenderAction a reader that does NOT
// implement e.Tok must exhibit, after running the ordered gates. resolve
// supplies the referenced segment's slot digest for the payload-digest gate.
// It returns an error (and no meaningful action) if any gate fails:
// a payload-digest mismatch/unresolved ref (T-0053), an out-of-range
// disposition (T-0054), or a missing mandatory fallback (T-0055). The reader
// must never fall back to a silent omission on any of these failures -- the
// document is rejected.
func EnforceDisposition(e ExtEnvelope, resolve SlotDigestResolver) (RenderAction, error) {
	// (1) integrity gate first: the digest is verified before any disposition.
	if err := VerifyPayloadDigest(e, resolve); err != nil {
		return 0, err
	}
	// (2) disposition must be one of the closed three.
	if err := ValidateDisposition(e); err != nil {
		return 0, err
	}
	// (3) ignore/degrade require a present fallback.
	if err := ValidateFallbackPresence(e); err != nil {
		return 0, err
	}
	// Enact the declared disposition -- the single defined behaviour.
	switch e.Disposition {
	case DispositionIgnore:
		return ActionRenderAbsent, nil
	case DispositionDegrade:
		return ActionRenderFallback, nil
	case DispositionRefuse:
		return ActionRefuseDocument, nil
	default:
		// Unreachable: ValidateDisposition already rejected any other value.
		return 0, fmt.Errorf("registry: unreachable disposition 0x%02x", uint8(e.Disposition))
	}
}

// EnforcedFallbackTarget returns the unit-id a degrade action renders, for a
// caller that has enacted ActionRenderFallback. It is e.FallbackRef (already
// checked non-zero16 by ValidateFallbackPresence for degrade).
func EnforcedFallbackTarget(e ExtEnvelope) pdlfmt.UnitID { return e.FallbackRef }
