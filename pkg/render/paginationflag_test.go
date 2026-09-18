package render

import (
	"encoding/json"
	"testing"
)

// TestFR_113_SubstitutionMarksPaginationNonAuthoritative is T-0260's named
// integration test (FR-113). A render that performed at least one resource
// substitution marks the resulting pagination non-authoritative; a render with
// zero substitutions leaves fixed pagination authoritative.
func TestFR_113_SubstitutionMarksPaginationNonAuthoritative(t *testing.T) {
	// Clean render: zero substitutions -> pagination authoritative.
	clean := NewRenderReport()
	if !clean.PaginationAuthoritative {
		t.Error("a clean render must leave pagination authoritative")
	}

	// Degraded render: one substitution -> pagination non-authoritative.
	degraded := NewRenderReport()
	degraded.RecordSubstitution(pdUnit(0x01), ReasonAbsent, KindPlaceholder)
	if degraded.PaginationAuthoritative {
		t.Error("a render with a substitution must mark pagination non-authoritative")
	}

	// More substitutions keep it non-authoritative.
	degraded.RecordSubstitution(pdUnit(0x02), ReasonDecodeFailed, KindPlaceholder)
	if degraded.PaginationAuthoritative {
		t.Error("pagination must stay non-authoritative after further substitutions")
	}

	// The flag survives JSON round-trip in both directions.
	for _, rep := range []*RenderReport{clean, degraded} {
		raw, _ := rep.JSON()
		var dec RenderReport
		if err := json.Unmarshal(raw, &dec); err != nil {
			t.Fatalf("report JSON: %v", err)
		}
		if dec.PaginationAuthoritative != rep.PaginationAuthoritative {
			t.Errorf("pagination_authoritative round-trip mismatch: got %v want %v",
				dec.PaginationAuthoritative, rep.PaginationAuthoritative)
		}
	}
}
