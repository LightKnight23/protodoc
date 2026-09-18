package integrity

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const m11ReviewDocPath = "../../docs/m11-redaction-security-review.md"

// TestSEC_M11_ArgusReviewPassed is T-0205's named integration test: the M11
// redaction milestone-exit security-review sign-off gate. It confirms the
// review artifact exists, records zero unresolved P0/P1 findings, requires no
// Eyvar waiver, covers every task T-0185..T-0206 by ID, and names the
// load-bearing redaction controls (provably total removal, T_C root preserved,
// undesignated omission, octet-absent publish, actor-identity stripping,
// RED-ALIGN).
func TestSEC_M11_ArgusReviewPassed(t *testing.T) {
	raw, err := os.ReadFile(m11ReviewDocPath)
	if err != nil {
		t.Fatalf("reading M11 review %s: %v", m11ReviewDocPath, err)
	}
	doc := string(raw)

	if !strings.Contains(doc, "Zero unresolved P0 or P1 findings") {
		t.Fatalf("M11 review does not record zero unresolved P0/P1 findings")
	}
	if !strings.Contains(doc, "No finding requires an Eyvar waiver") {
		t.Fatalf("M11 review does not confirm no finding requires an Eyvar waiver")
	}

	for n := 185; n <= 206; n++ {
		id := fmt.Sprintf("T-%04d", n)
		if !strings.Contains(doc, id) {
			t.Fatalf("M11 review does not cover task %s by ID", id)
		}
	}

	for _, control := range []string{"Provably total removal", "T_C root preserved", "Undesignated omission", "Octet-absent publish", "Actor-identity stripping", "RED-ALIGN"} {
		if !strings.Contains(doc, control) {
			t.Fatalf("M11 review does not address control %q", control)
		}
	}
}
