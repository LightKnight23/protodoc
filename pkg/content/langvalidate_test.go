package content

import (
	"errors"
	"testing"
)

// TestFR_031_PDLANG001RejectsUnresolvableTag is T-0082's named conformance
// test. A text span whose language-tag reference does not resolve to any
// registered tag is rejected by the validator naming rule PD-LANG-001 and
// the offending unit id; a resolvable reference is accepted; an unset
// reference is rejected before resolution (FR-031).
func TestFR_031_PDLANG001RejectsUnresolvableTag(t *testing.T) {
	spanID, _ := testMintID()

	reg := MapLanguageRegistry{
		LangRef(1): "en-US",
		LangRef(2): "ja-JP",
		LangRef(3): "ar-EG",
	}

	// A resolvable reference passes.
	if err := ValidateLanguageResolves(spanID, LangRef(2), reg); err != nil {
		t.Fatalf("resolvable reference rejected: %v", err)
	}

	// An unresolvable reference (not in the registry) is rejected naming
	// PD-LANG-001 and the offending unit id.
	err := ValidateLanguageResolves(spanID, LangRef(99), reg)
	if !errors.Is(err, ErrUnresolvableLanguageTag) {
		t.Fatalf("unresolvable reference returned %v, want ErrUnresolvableLanguageTag", err)
	}
	var ule *UnresolvableLanguageError
	if !errors.As(err, &ule) {
		t.Fatalf("error is %T, want *UnresolvableLanguageError", err)
	}
	if !ule.UnitID.Equal(spanID) {
		t.Fatalf("error names unit %x, want %x", ule.UnitID, spanID)
	}
	if ule.Ref != LangRef(99) {
		t.Fatalf("error names ref %d, want 99", ule.Ref)
	}

	// An unset reference is rejected before resolution logic (FR-031
	// mandatory-field check), distinct from the unresolvable case.
	if err := ValidateLanguageResolves(spanID, LangUnset, reg); !errors.Is(err, ErrMissingLanguageRef) {
		t.Fatalf("unset reference returned %v, want ErrMissingLanguageRef", err)
	}
}
