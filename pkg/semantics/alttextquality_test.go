package semantics

import "testing"

// TestConformanceA11Y003AltTextQuality is T-0285's named conformance test
// (vector CONFORMANCE-A11Y-003-alttext-quality; FR-040). Rule PD-A11Y-003
// rejects a non-decorative object whose alt text is empty or merely echoes its
// filename, digest, or dimensions.
func TestConformanceA11Y003AltTextQuality(t *testing.T) {
	md := ObjectMetadata{Filename: "chart_final.png", Digest: []byte{0xde, 0xad, 0xbe, 0xef}, Width: 800, Height: 600}

	// A meaningful alt text passes.
	good := EmbeddedObjectAlt{Decorative: false, AltText: "Quarterly revenue rose from $1M to $4M across the year."}
	if f := CheckAltTextQuality(good, md); len(f) != 0 {
		t.Errorf("meaningful alt text should pass, got %+v", f)
	}

	// Empty alt text on a non-decorative object is rejected.
	if f := CheckAltTextQuality(EmbeddedObjectAlt{Decorative: false, AltText: "   "}, md); len(f) != 1 || f[0].Kind != AltEmpty {
		t.Errorf("empty alt text: findings = %+v, want one AltEmpty", f)
	}

	// Alt text echoing the filename is rejected.
	if f := CheckAltTextQuality(EmbeddedObjectAlt{AltText: "chart_final.png"}, md); len(f) != 1 || f[0].Kind != AltEchoesMetadata {
		t.Errorf("filename echo: findings = %+v, want one AltEchoesMetadata", f)
	}
	// Filename without extension too.
	if f := CheckAltTextQuality(EmbeddedObjectAlt{AltText: "chart_final"}, md); len(f) != 1 {
		t.Errorf("filename-stem echo should be rejected, got %+v", f)
	}

	// Alt text echoing the dimensions is rejected.
	if f := CheckAltTextQuality(EmbeddedObjectAlt{AltText: "800 x 600"}, md); len(f) != 1 || f[0].Kind != AltEchoesMetadata {
		t.Errorf("dimension echo: findings = %+v, want one AltEchoesMetadata", f)
	}

	// Alt text echoing the digest hex is rejected.
	if f := CheckAltTextQuality(EmbeddedObjectAlt{AltText: "DEADBEEF"}, md); len(f) != 1 || f[0].Kind != AltEchoesMetadata {
		t.Errorf("digest echo: findings = %+v, want one AltEchoesMetadata", f)
	}

	// A decorative object is not checked (empty alt required, checked at T-0284).
	if f := CheckAltTextQuality(EmbeddedObjectAlt{Decorative: true, AltText: ""}, md); len(f) != 0 {
		t.Errorf("decorative object should not be quality-checked, got %+v", f)
	}

	if got := CheckAltTextQuality(EmbeddedObjectAlt{AltText: "800x600"}, md); len(got) == 1 && got[0].Rule != RuleAltTextQuality {
		t.Errorf("rule = %q, want %q", got[0].Rule, RuleAltTextQuality)
	}
}
