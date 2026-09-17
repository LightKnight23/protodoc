package registry

import (
	"errors"
	"testing"
)

// TestFR_016_PD_EXT_003_FallbackMinimumContent is T-0057's named unit test for
// validator rule PD-EXT-003 (FR-016). A declared fallback must yield at least
// one extractable text unit OR at least one non-decorative mark; whitespace-
// only and zero-extent (no non-decorative mark) fallbacks are rejected -- the
// N-EXT-THIN corpus -- with zero false accepts.
func TestFR_016_PD_EXT_003_FallbackMinimumContent(t *testing.T) {
	tok := NewExtToken(0x00000016, 3)

	// Accepted: real text, or a non-decorative mark, or both.
	accepted := []FallbackContent{
		{Text: "hello"},
		{Text: "  x  "},                   // surrounded by spaces but has a non-space scalar
		{Text: "", NonDecorativeMarks: 1}, // a non-decorative mark, no text
		{Text: "\t\n word", NonDecorativeMarks: 0},
		{Text: "     ", NonDecorativeMarks: 2}, // whitespace text but marks carry it
	}
	for i, fc := range accepted {
		if err := ValidateFallbackMinimumContent(tok, fc); err != nil {
			t.Errorf("accepted[%d] %+v rejected: %v", i, fc, err)
		}
	}

	// Rejected (N-EXT-THIN): whitespace-only or empty text with zero
	// non-decorative marks.
	rejected := []FallbackContent{
		{Text: "", NonDecorativeMarks: 0},          // empty, zero-extent
		{Text: " ", NonDecorativeMarks: 0},         // single space
		{Text: "\t\n\r   ", NonDecorativeMarks: 0}, // assorted whitespace
	}
	for i, fc := range rejected {
		err := ValidateFallbackMinimumContent(tok, fc)
		if !errors.Is(err, ErrFallbackBelowMinimum) {
			t.Fatalf("rejected[%d] %+v: err = %v, want ErrFallbackBelowMinimum", i, fc, err)
		}
		var me *FallbackMinimumError
		if !errors.As(err, &me) {
			t.Fatalf("rejected[%d]: expected *FallbackMinimumError, got %T", i, err)
		}
		if me.Tok != tok {
			t.Errorf("rejected[%d]: error names tok %x, want %x", i, me.Tok, tok)
		}
	}
}
