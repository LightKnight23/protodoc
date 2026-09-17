package content

import (
	"errors"
	"testing"
)

// TestFR_031_TextSpanRequiresLanguageTagField is T-0081's named test.
// TextBlock and Run require a declared (nonzero) language-tag reference at
// construction, and a decoded span lacking the field is rejected before any
// resolution logic (FR-031: the field is mandatory, not optional).
func TestFR_031_TextSpanRequiresLanguageTagField(t *testing.T) {
	rid, _ := testMintID()
	bid, _ := testMintID()

	// NewRun rejects an unset language reference.
	if _, err := NewRun(rid, 0, "hello", LangUnset); !errors.Is(err, ErrMissingLanguageRef) {
		t.Fatalf("NewRun with LangUnset returned %v, want ErrMissingLanguageRef", err)
	}
	// NewRun accepts a declared reference and carries it.
	r, err := NewRun(rid, 0, "hello", LangRef(7))
	if err != nil {
		t.Fatalf("NewRun with a declared reference failed: %v", err)
	}
	if r.LangRef != 7 {
		t.Fatalf("run LangRef = %d, want 7", r.LangRef)
	}

	// NewTextBlock rejects an unset block reference.
	if _, err := NewTextBlock(bid, LangUnset, nil); !errors.Is(err, ErrMissingLanguageRef) {
		t.Fatalf("NewTextBlock with LangUnset block ref returned %v, want ErrMissingLanguageRef", err)
	}
	// NewTextBlock rejects a block whose run lacks a reference, even if the
	// block itself declares one.
	badRun := Run{RunID: rid, Text: "x"} // LangRef defaults to LangUnset
	if _, err := NewTextBlock(bid, LangRef(3), []Run{badRun}); !errors.Is(err, ErrMissingLanguageRef) {
		t.Fatalf("NewTextBlock with an unset-ref run returned %v, want ErrMissingLanguageRef", err)
	}
	// NewTextBlock accepts a fully-declared block.
	goodRun, _ := NewRun(rid, 0, "x", LangRef(3))
	tb, err := NewTextBlock(bid, LangRef(3), []Run{goodRun})
	if err != nil {
		t.Fatalf("NewTextBlock with declared refs failed: %v", err)
	}
	if tb.LangRef != 3 || len(tb.Runs) != 1 {
		t.Fatalf("TextBlock not constructed as expected: %+v", tb)
	}

	// Decode-time validation rejects a missing field before resolution.
	if err := ValidateLanguageRefPresent(rid, LangUnset); !errors.Is(err, ErrMissingLanguageRef) {
		t.Fatalf("ValidateLanguageRefPresent(unset) returned %v, want ErrMissingLanguageRef", err)
	}
	if err := ValidateLanguageRefPresent(rid, LangRef(1)); err != nil {
		t.Fatalf("ValidateLanguageRefPresent(set) returned %v, want nil", err)
	}
}
