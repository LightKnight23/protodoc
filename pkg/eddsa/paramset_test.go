package eddsa

import (
	"errors"
	"testing"
)

// TestCON_015_ParamSetAllowlist is T-0102's named test. It confirms only
// 0x0001 is accepted by ParamSet.Validate and that 0x0000, 0x0002 and
// 0xFFFF (and a sweep of other reserved values) are rejected with the named
// allowlist error -- never a warning-level accept (CON-015, integrity.abnf
// S10.2).
func TestCON_015_ParamSetAllowlist(t *testing.T) {
	// The one allowlisted value.
	if err := ParamSetV1.Validate(); err != nil {
		t.Fatalf("ParamSetV1 (0x0001) rejected: %v", err)
	}
	if !ParamSetV1.IsAllowlisted() {
		t.Fatalf("ParamSetV1 not reported allowlisted")
	}

	// Named rejections.
	for _, v := range []ParamSet{ParamSetReserved, 0x0002, 0xFFFF} {
		err := v.Validate()
		if err == nil {
			t.Fatalf("param_set 0x%04x was accepted, want rejected", uint16(v))
		}
		if !errors.Is(err, ErrParamSetNotAllowlisted) {
			t.Fatalf("param_set 0x%04x error = %v, want ErrParamSetNotAllowlisted", uint16(v), err)
		}
		if v.IsAllowlisted() {
			t.Fatalf("param_set 0x%04x wrongly reported allowlisted", uint16(v))
		}
	}

	// Exhaustive sweep: exactly one value in the whole u16 space is valid.
	valid := 0
	for i := 0; i <= 0xFFFF; i++ {
		if ParamSet(i).Validate() == nil {
			valid++
			if ParamSet(i) != ParamSetV1 {
				t.Fatalf("unexpected allowlisted value 0x%04x", i)
			}
		}
	}
	if valid != 1 {
		t.Fatalf("%d param_set values are allowlisted, want exactly 1 (0x0001)", valid)
	}
}
