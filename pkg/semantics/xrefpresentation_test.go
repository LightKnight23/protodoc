package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConformanceXREF000PresentationFunctionRoundtrip is T-0288's named
// conformance test (vector CONFORMANCE-XREF-000-presentation-function-
// roundtrip; FR-084). The presentation-function field round-trips through
// encode/decode for every value; reserved values are rejected; the rendered
// reference text is a COMPUTED_INLINE derived from the target, not frozen text.
func TestConformanceXREF000PresentationFunctionRoundtrip(t *testing.T) {
	for _, fn := range []PresentationFn{PresentTargetPage, PresentCitationLabel, PresentNumberingLabel, PresentTitleText} {
		enc, err := EncodePresentationFn(nil, fn)
		if err != nil {
			t.Fatalf("encode fn %d: %v", fn, err)
		}
		got, n, err := DecodePresentationFn(enc)
		if err != nil || got != fn || n != 1 {
			t.Errorf("roundtrip fn %d: got %d (n=%d) err=%v", fn, got, n, err)
		}
	}

	// Reserved value (0x04) rejected on encode and decode.
	if _, err := EncodePresentationFn(nil, PresentationFn(0x04)); err == nil {
		t.Errorf("reserved presentation-fn 0x04 must be rejected on encode")
	}
	if _, _, err := DecodePresentationFn([]byte{0x04}); err == nil {
		t.Errorf("reserved presentation-fn 0x04 must be rejected on decode")
	}

	// Rendered text is derived from the target (page number here), as a
	// COMPUTED_INLINE, not frozen authored text.
	xr := CrossReference{XrefID: pdUnitSem(0x01), Target: pdUnitSem(0x40), PresentationFn: PresentTargetPage}
	ci := RenderReference(xr, 42, "cite", "num", "title")
	if ci.Value != "42" {
		t.Errorf("target-page render = %q, want %q", ci.Value, "42")
	}
	// It isolates when materialised into surrounding text.
	segs := ResolveInlineSequence("see p. ", ci, ".")
	if len(segs) != 3 || !segs[1].Isolated {
		t.Errorf("rendered reference must be an isolated computed inline, got %+v", segs)
	}

	// The wire grammar declares the field.
	abnf, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "contracts", "document.abnf"))
	if err != nil {
		t.Fatalf("read document.abnf: %v", err)
	}
	if !strings.Contains(string(abnf), "xref-presentation-fn") {
		t.Errorf("document.abnf must define xref-presentation-fn")
	}
}
