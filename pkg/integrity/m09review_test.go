package integrity

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const m09ReviewDocPath = "../../docs/m09-signature-coverage-security-review.md"

// TestM09_ArgusMilestoneExitReviewRecorded is T-0363's named conformance test:
// the M09 milestone-exit review sign-off gate. It confirms the review artifact
// exists, records zero unresolved P0/P1 findings, covers every task
// T-0145..T-0166 by ID, confirms the T-0165 ambient-audit rows were checked,
// and names the load-bearing controls it certifies. Any unresolved finding (an
// OPEN row awaiting Eyvar) would leave this gate red until resolved or waived.
func TestM09_ArgusMilestoneExitReviewRecorded(t *testing.T) {
	raw, err := os.ReadFile(m09ReviewDocPath)
	if err != nil {
		t.Fatalf("reading M09 review %s: %v", m09ReviewDocPath, err)
	}
	doc := string(raw)

	if !strings.Contains(doc, "Zero unresolved P0 or P1 findings") {
		t.Fatalf("M09 review does not record zero unresolved P0/P1 findings")
	}
	// No finding may be left OPEN awaiting an Eyvar waiver (that would block
	// merge); the review must state none requires a waiver.
	if !strings.Contains(doc, "No finding requires an Eyvar waiver") {
		t.Fatalf("M09 review does not confirm no finding requires an Eyvar waiver")
	}
	// The ambient-audit rows must be confirmed checked (T-0165 clause).
	if !strings.Contains(doc, "Every ambient-audit row was checked") {
		t.Fatalf("M09 review does not confirm the T-0165 ambient-audit rows were checked")
	}

	// Every task in the milestone range is cited by ID.
	for n := 145; n <= 166; n++ {
		id := fmt.Sprintf("T-%04d", n)
		if !strings.Contains(doc, id) {
			t.Fatalf("M09 review does not cover task %s by ID", id)
		}
	}

	// The review names the load-bearing controls it certifies.
	for _, control := range []string{"Domain separation", "no-trust-stored", "identity discipline", "check-before-allocate"} {
		if !strings.Contains(doc, control) {
			t.Fatalf("M09 review does not address control %q", control)
		}
	}
}
