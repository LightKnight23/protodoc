package registry

import (
	"errors"
	"testing"
)

// TestFR_015_PD_EXT_002_FallbackMandatoryWhenIgnoreOrDegrade is T-0055's named
// unit test for validator rule PD-EXT-002 (FR-015). Under the ignore or
// degrade disposition, ext-fallback-ref is mandatory (non-zero16); a zero16
// fallback there is rejected naming the ext-tok. Under refuse, no fallback is
// required.
func TestFR_015_PD_EXT_002_FallbackMandatoryWhenIgnoreOrDegrade(t *testing.T) {
	tok := NewExtToken(0x00000042, 0x00000001)
	present := fixtureUnitID(0x30)

	base := ExtEnvelope{
		Tok:           tok,
		PayloadLength: 1,
		PayloadRef:    fixtureUnitID(0x10),
		PayloadDigest: fixtureDigest(0x20),
	}

	// (1) ignore/degrade WITH a non-zero16 fallback: accepted.
	for _, disp := range []ExtDisposition{DispositionIgnore, DispositionDegrade} {
		e := base
		e.Disposition = disp
		e.FallbackRef = present
		if err := ValidateFallbackPresence(e); err != nil {
			t.Errorf("%v with a present fallback rejected: %v", disp, err)
		}
	}

	// (2) ignore/degrade WITH a zero16 fallback: PD-EXT-002 reject naming tok.
	for _, disp := range []ExtDisposition{DispositionIgnore, DispositionDegrade} {
		e := base
		e.Disposition = disp
		e.FallbackRef = Zero16
		err := ValidateFallbackPresence(e)
		if !errors.Is(err, ErrExtFallbackMissing) {
			t.Fatalf("%v with zero16 fallback: err = %v, want ErrExtFallbackMissing", disp, err)
		}
		var fe *FallbackPresenceError
		if !errors.As(err, &fe) {
			t.Fatalf("%v: expected a *FallbackPresenceError, got %T", disp, err)
		}
		if fe.Tok != tok {
			t.Errorf("%v: error names tok %x, want %x", disp, fe.Tok, tok)
		}
		if fe.Disposition != disp {
			t.Errorf("%v: error carries disposition %v", disp, fe.Disposition)
		}
	}

	// (3) refuse requires no fallback: a zero16 fallback under refuse is
	// accepted (nothing to degrade to).
	e := base
	e.Disposition = DispositionRefuse
	e.FallbackRef = Zero16
	if err := ValidateFallbackPresence(e); err != nil {
		t.Errorf("refuse with no fallback rejected: %v", err)
	}
	// A present fallback under refuse is also not an error here (not required,
	// not forbidden by PD-EXT-002).
	e.FallbackRef = present
	if err := ValidateFallbackPresence(e); err != nil {
		t.Errorf("refuse with a present fallback rejected by PD-EXT-002: %v", err)
	}
}
