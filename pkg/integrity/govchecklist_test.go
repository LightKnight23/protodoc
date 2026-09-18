package integrity

import (
	"os"
	"strings"
	"testing"
)

// TestGovChecklistFR061FR075RulingRecorded is T-0190's named integration test
// (gov-checklist-FR061-FR075-ruling-recorded, FR-075). the severed-state-digest requirement's original frozen
// text required a bare unsalted severed-state digest, which cannot provide
// FR-075's 2^80 hiding floor; the governance ruling CQ-015 amended the severed-state-digest requirement to
// require the salted-commitment form. This test confirms that ruling is
// recorded in clarify.md (the salted-commitment form the M11 code is built
// against is the ruled-upon one, not a provisional guess).
func TestGovChecklistFR061FR075RulingRecorded(t *testing.T) {
	clarifyPath := "../../specs/001-protodoc-format-core/clarify.md"
	raw, err := os.ReadFile(clarifyPath)
	if err != nil {
		t.Fatalf("reading clarify.md: %v", err)
	}
	doc := string(raw)

	// Note: the severed-state-digest requirement references below are constructed at runtime so this test's
	// source does not register false traceability coverage for the severed-state-digest requirement (this
	// task covers FR-075's governance ruling; the severed-state digest that
	// the severed-state-digest requirement governs is implemented in a later milestone).
	fr061 := "FR-" + "061"
	idx := strings.Index(doc, "CQ-015: "+fr061+" vs FR-075")
	if idx < 0 {
		t.Fatal("clarify.md has no CQ-015 (" + fr061 + " vs FR-075) ruling section")
	}
	section := doc[idx:]

	required := []struct {
		desc   string
		needle string
	}{
		{"names the severed-state-digest requirement", fr061},
		{"names FR-075", "FR-075"},
		{"records a resolution", "Resolution:"},
		{"rules for the salted-commitment form", "salted-commitment"},
		{"cites the 2^80 hiding floor rationale", "hiding floor"},
	}
	for _, r := range required {
		if !strings.Contains(section, r.needle) {
			t.Errorf("CQ-015 ruling does not %s (missing %q)", r.desc, r.needle)
		}
	}

	// The M11 commitment IS the salted form (matching the ruling): the
	// commitment of designated content depends on the salt.
	frame := []byte("x")
	a, _ := DesignateRedactable(redUnitID(1), frame, redSalt(1)).Commitment()
	b, _ := DesignateRedactable(redUnitID(1), frame, redSalt(2)).Commitment()
	if a == b {
		t.Error("the M11 commitment is not salted, contradicting the CQ-015 ruling")
	}
}
