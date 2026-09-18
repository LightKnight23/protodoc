package semantics

import "testing"

// TestConformanceBIDI002InbandControlRejected is T-0271's named conformance
// test (vector id CONFORMANCE-BIDI-002-inband-control-rejected; FR-033). Text
// carrying any in-band Unicode directional control character is rejected by
// rule PD-BIDI-002 naming the control and its position; plain text with no
// controls passes.
func TestConformanceBIDI002InbandControlRejected(t *testing.T) {
	unit := pdUnitSem(0x77)

	// Each control character is rejected.
	controls := []struct {
		r    rune
		name string
	}{
		{0x202A, "LRE"}, {0x202B, "RLE"}, {0x202D, "LRO"}, {0x202E, "RLO"},
		{0x202C, "PDF"}, {0x2066, "LRI"}, {0x2067, "RLI"}, {0x2068, "FSI"},
		{0x2069, "PDI"}, {0x200E, "LRM"}, {0x200F, "RLM"}, {0x061C, "ALM"},
	}
	for _, c := range controls {
		text := "abc" + string(c.r) + "def"
		f := CheckNoInbandDirectionControl(unit, text)
		if f == nil {
			t.Errorf("%s (%#x): in-band control not rejected", c.name, c.r)
			continue
		}
		if f.Rule != RuleInbandDirectionControl {
			t.Errorf("%s: rule = %q, want %q", c.name, f.Rule, RuleInbandDirectionControl)
		}
		if f.Control != c.name {
			t.Errorf("finding control = %q, want %q", f.Control, c.name)
		}
		if f.RuneIdx != 3 {
			t.Errorf("%s: rune index = %d, want 3", c.name, f.RuneIdx)
		}
	}

	// Plain mixed-direction text WITHOUT in-band controls passes (direction is
	// structural, not in-band).
	if f := CheckNoInbandDirectionControl(unit, "hello שלום world"); f != nil {
		t.Errorf("plain mixed-direction text must pass, got %v", f)
	}
	if f := CheckNoInbandDirectionControl(unit, ""); f != nil {
		t.Errorf("empty text must pass, got %v", f)
	}
}
