package registry

import (
	"errors"
	"fmt"
	"unicode"
)

// Fallback minimum-content validation (T-0057, FR-016; document.abnf S6,
// validator rule PD-EXT-003). WHERE an ExtensionEnvelope declares a fallback,
// that fallback MUST yield at least one extractable text unit OR at least one
// mark classified as non-decorative. A mere non-emptiness rule is satisfied by
// a single space or a zero-area object, which does not deliver graceful
// degradation; PD-EXT-003 therefore rejects a whitespace-only or zero-extent
// fallback (corpus N-EXT-THIN). A minimum-content criterion is what a
// validator can actually decide.

// PDEXT003 is the validator rule id for a below-minimum fallback.
const PDEXT003 = "PD-EXT-003"

// ErrFallbackBelowMinimum is returned when a declared fallback yields neither
// an extractable text unit nor a non-decorative mark.
var ErrFallbackBelowMinimum = errors.New("registry: fallback yields no extractable text unit and no non-decorative mark (PD-EXT-003)")

// FallbackContent is the minimal, decidable description of what a fallback
// yields, as a validator sees it: the fallback's extractable text (already
// NFC, already scalar-admissible; here inspected only for non-whitespace
// content) and the count of marks it contributes classified as
// non-decorative. A whitespace-only Text with zero NonDecorativeMarks is
// exactly the N-EXT-THIN case PD-EXT-003 rejects.
type FallbackContent struct {
	// Text is the fallback's extractable text content.
	Text string
	// NonDecorativeMarks is the number of marks the fallback contributes that
	// are classified as non-decorative (a decorative-only mark does not count
	// toward the minimum).
	NonDecorativeMarks int
}

// FallbackMinimumError is the PD-EXT-003 rejection, naming the ext-tok whose
// fallback fell below the minimum.
type FallbackMinimumError struct {
	Tok ExtToken
}

func (e *FallbackMinimumError) Error() string {
	return fmt.Sprintf("%s: ext-tok %x fallback is whitespace-only/zero-extent, yielding no extractable text unit or non-decorative mark (%v)",
		PDEXT003, e.Tok, ErrFallbackBelowMinimum)
}

func (e *FallbackMinimumError) Unwrap() error { return ErrFallbackBelowMinimum }

// hasExtractableTextUnit reports whether s yields at least one extractable
// text unit: at least one non-whitespace Unicode scalar. Whitespace-only text
// (spaces, tabs, newlines) yields no extractable unit for degradation
// purposes.
func hasExtractableTextUnit(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// ValidateFallbackMinimumContent enforces PD-EXT-003 (FR-016): the fallback
// content must yield at least one extractable text unit or at least one
// non-decorative mark. It returns a *FallbackMinimumError naming tok when the
// fallback is whitespace-only with no non-decorative marks. The check applies
// only to a declared fallback; callers skip it for a refuse envelope with no
// fallback.
func ValidateFallbackMinimumContent(tok ExtToken, fc FallbackContent) error {
	if hasExtractableTextUnit(fc.Text) || fc.NonDecorativeMarks > 0 {
		return nil
	}
	return &FallbackMinimumError{Tok: tok}
}
