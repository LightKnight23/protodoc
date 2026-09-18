package render

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestFR_112_RenderReportRecordsSubstitution is T-0259's named integration test
// (FR-112). Every resource-substitution event during a render run is recorded
// in the machine-readable render report, naming the resource id, failure
// reason, and substitution kind; a clean run reports zero substitutions.
func TestFR_112_RenderReportRecordsSubstitution(t *testing.T) {
	rep := NewRenderReport()

	// Clean run so far: no substitutions.
	if len(rep.Substitutions) != 0 {
		t.Fatal("fresh report must have zero substitutions")
	}

	// Record substitutions for each failure reason.
	resA := pdUnit(0x11)
	resB := pdUnit(0x22)
	resC := pdUnit(0x33)
	rep.RecordSubstitution(resA, ReasonAbsent, KindPlaceholder)
	rep.RecordSubstitution(resB, ReasonDigestMismatch, KindPlaceholder)
	rep.RecordSubstitution(resC, ReasonDecodeFailed, KindPlaceholder)

	if len(rep.Substitutions) != 3 {
		t.Fatalf("recorded %d substitutions, want 3", len(rep.Substitutions))
	}

	// The report is machine-readable JSON naming each resource, reason, kind.
	raw, err := rep.JSON()
	if err != nil {
		t.Fatalf("render report JSON: %v", err)
	}
	var decoded RenderReport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("render report not machine-readable JSON: %v", err)
	}
	if len(decoded.Substitutions) != 3 {
		t.Fatalf("decoded %d substitutions, want 3", len(decoded.Substitutions))
	}

	// Every substitution names its resource, reason and kind.
	wantReasons := map[string]SubstitutionReason{
		hexID(resA): ReasonAbsent,
		hexID(resB): ReasonDigestMismatch,
		hexID(resC): ReasonDecodeFailed,
	}
	seen := map[string]bool{}
	for _, ev := range decoded.Substitutions {
		wr, ok := wantReasons[ev.ResourceID]
		if !ok {
			t.Errorf("unexpected resource id %q in report", ev.ResourceID)
			continue
		}
		if ev.Reason != wr {
			t.Errorf("resource %s: reason %q, want %q", ev.ResourceID, ev.Reason, wr)
		}
		if ev.Kind != KindPlaceholder {
			t.Errorf("resource %s: kind %q, want placeholder", ev.ResourceID, ev.Kind)
		}
		seen[ev.ResourceID] = true
	}
	if len(seen) != 3 {
		t.Errorf("not all substitutions reported: %v", seen)
	}

	// The JSON is a keyed object a pipeline can parse.
	if !strings.Contains(string(raw), "\"substitutions\"") {
		t.Error("render report JSON must carry a substitutions field")
	}
}
