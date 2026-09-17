package integrity

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const m08ReviewDocPath = "../../docs/m08-integrity-trees-security-review.md"

// TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff is T-0362's named
// test: the milestone-exit review sign-off gate. It confirms the M08 review
// artifact exists, records zero unresolved P0/P1 findings, and covers every
// task T-0129..T-0144 by ID.
func TestArgus_M08IntegrityTreesMilestoneExitReviewSignOff(t *testing.T) {
	raw, err := os.ReadFile(m08ReviewDocPath)
	if err != nil {
		t.Fatalf("reading M08 review %s: %v", m08ReviewDocPath, err)
	}
	doc := string(raw)

	if !strings.Contains(doc, "Zero unresolved P0 or P1 findings") {
		t.Fatalf("M08 review does not record zero unresolved P0/P1 findings")
	}

	// Every task in the milestone range must be cited by ID.
	for n := 129; n <= 144; n++ {
		id := fmt.Sprintf("T-%04d", n)
		if !strings.Contains(doc, id) {
			t.Fatalf("M08 review does not cover task %s by ID", id)
		}
	}

	// The review must name the load-bearing controls it certifies.
	for _, control := range []string{"Domain separation", "no-trust-stored", "redactable salts"} {
		if !strings.Contains(doc, control) {
			t.Fatalf("M08 review does not address control %q", control)
		}
	}
}
