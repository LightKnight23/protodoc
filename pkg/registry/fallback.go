package registry

import (
	"errors"
	"fmt"
)

// ext-fallback-ref presence validation (T-0055, FR-015; document.abnf S6,
// validator rule PD-EXT-002). WHERE an ExtensionEnvelope declares the ignore
// or degrade disposition, an accompanying fallback expressible entirely in the
// core feature set is MANDATORY: ext-fallback-ref MUST be a non-zero16 unit-id.
// PD-EXT-002 rejects a zero16 fallback under either of those two dispositions.
// The refuse disposition requires no fallback (the reader declines to render
// the whole document, so there is nothing to degrade to), and a fallback under
// refuse is not required.

// PDEXT002 is the validator rule id for a missing mandatory fallback.
const PDEXT002 = "PD-EXT-002"

// ErrExtFallbackMissing is returned when an ignore/degrade envelope has a
// zero16 (absent) ext-fallback-ref.
var ErrExtFallbackMissing = errors.New("registry: ext-fallback-ref is mandatory (non-zero16) when disposition is ignore or degrade")

// FallbackPresenceError is the PD-EXT-002 structural rejection: an ignore- or
// degrade-disposition envelope with no fallback. It names the ext-tok and the
// disposition that made the fallback mandatory.
type FallbackPresenceError struct {
	Tok         ExtToken
	Disposition ExtDisposition
}

func (e *FallbackPresenceError) Error() string {
	return fmt.Sprintf("%s: ext-tok %x declares %v disposition but has a zero16 (absent) ext-fallback-ref (%v)",
		PDEXT002, e.Tok, e.Disposition, ErrExtFallbackMissing)
}

func (e *FallbackPresenceError) Unwrap() error { return ErrExtFallbackMissing }

// ValidateFallbackPresence enforces PD-EXT-002 (FR-015): under the ignore or
// degrade disposition, ext-fallback-ref must be non-zero16; a zero16 fallback
// there is rejected with a *FallbackPresenceError naming the ext-tok. Under
// refuse, no fallback is required and the field is not checked here. It
// assumes the disposition is already known-valid (T-0054's ValidateDisposition
// runs first); an out-of-range disposition is not this check's concern.
func ValidateFallbackPresence(e ExtEnvelope) error {
	switch e.Disposition {
	case DispositionIgnore, DispositionDegrade:
		if e.FallbackRef == Zero16 {
			return &FallbackPresenceError{Tok: e.Tok, Disposition: e.Disposition}
		}
		return nil
	default: // refuse (or any other value, which ValidateDisposition catches)
		return nil
	}
}
