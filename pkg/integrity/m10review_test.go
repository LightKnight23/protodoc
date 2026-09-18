package integrity

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const m10ReviewDocPath = "../../docs/m10-ltv-security-review.md"

// TestArgus_M10_LtvSecurityReviewChecklist is T-0183's named integration test:
// the M10 LTV milestone-exit security-review sign-off gate. It confirms the
// review artifact exists, records zero unresolved P0/P1 findings, requires no
// Eyvar waiver, covers every task T-0167..T-0184 by ID, and names the
// load-bearing LTV controls (offline/no-network, verdict-stable-across-time,
// the FR-072/FR-073 unverified verdicts, check-before-allocate).
func TestArgus_M10_LtvSecurityReviewChecklist(t *testing.T) {
	raw, err := os.ReadFile(m10ReviewDocPath)
	if err != nil {
		t.Fatalf("reading M10 review %s: %v", m10ReviewDocPath, err)
	}
	doc := string(raw)

	if !strings.Contains(doc, "Zero unresolved P0 or P1 findings") {
		t.Fatalf("M10 review does not record zero unresolved P0/P1 findings")
	}
	if !strings.Contains(doc, "No finding requires an Eyvar waiver") {
		t.Fatalf("M10 review does not confirm no finding requires an Eyvar waiver")
	}

	for n := 167; n <= 184; n++ {
		id := fmt.Sprintf("T-%04d", n)
		if !strings.Contains(doc, id) {
			t.Fatalf("M10 review does not cover task %s by ID", id)
		}
	}

	for _, control := range []string{"Offline, no network", "Verdict stable across time", "FR-072", "FR-073", "check-before-allocate"} {
		if !strings.Contains(doc, control) {
			t.Fatalf("M10 review does not address control %q", control)
		}
	}
}
