package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCON_006_ShapingExceptionRecorded is T-0253's named conformance test
// (CON-006). It is an automated fixture check against the committed
// clarify.md: the file must carry a dated entry that names CP-009 v0.2.0's
// pinned-oracle exception clause, cites the constitution amendment-log entry,
// and states that CON-006 is satisfied by the same reasoning for the shaping
// construction. This is the audit trail for a ruling Eyvar already made (the
// CP-009 v0.2.0 amendment) -- the test verifies the record exists, it does not
// fabricate a ruling.
func TestCON_006_ShapingExceptionRecorded(t *testing.T) {
	clarifyPath := filepath.Join("..", "..", "specs", "001-protodoc-format-core", "clarify.md")
	data, err := os.ReadFile(clarifyPath)
	if err != nil {
		t.Fatalf("read clarify.md: %v", err)
	}
	text := string(data)

	// The dated CP-009 v0.2.0 exception clause must be named.
	required := []string{
		"CP-009",
		"v0.2.0",
		"2026-09-15",          // the amendment date
		"determinism oracle",  // the exception's substance
		"amendment-log entry", // cites the amendment log
		"CON-006",             // states CON-006's consistent reading
		"shaping",
	}
	for _, tok := range required {
		if !strings.Contains(text, tok) {
			t.Errorf("clarify.md must record the shaping exception token %q (CON-006 audit trail)", tok)
		}
	}

	// The record must tie CON-006 to CP-009's exception explicitly (both tokens
	// appearing in one CQ-013 section).
	idx := strings.Index(text, "CQ-013: CP-009 vs the pinned shaping oracle")
	if idx < 0 {
		t.Fatal("clarify.md must contain the CQ-013 shaping-oracle ruling section")
	}
	section := text[idx:]
	if end := strings.Index(section, "### CQ-014"); end >= 0 {
		section = section[:end]
	}
	if !strings.Contains(section, "CON-006") || !strings.Contains(section, "v0.2.0") {
		t.Error("the CQ-013 section must state CON-006 is satisfied via CP-009 v0.2.0's exception")
	}

	// Cross-check the constitution actually carries the amendment (the ruling
	// is real, not invented here).
	constPath := filepath.Join("..", "..", ".specify", "memory", "constitution.md")
	cdata, err := os.ReadFile(constPath)
	if err != nil {
		t.Fatalf("read constitution.md: %v", err)
	}
	ctext := string(cdata)
	if !strings.Contains(ctext, "Exception (added 2026-09-15, v0.2.0)") {
		t.Error("constitution.md must carry CP-009's v0.2.0 exception clause the clarify.md entry cites")
	}
}
