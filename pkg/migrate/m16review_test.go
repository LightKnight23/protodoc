package migrate_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const m16ReviewDocPath = "../../docs/m16-migration-security-review.md"

// TestCON_016_ArgusReviewNoP0P1Findings is T-0306's named integration test
// (CON-016). It confirms the M16 migration/registry security-review artifact
// exists, is honestly recorded as an author-conducted review (no fabricated
// external result, no Eyvar waiver), records zero unresolved P0/P1 findings,
// addresses the three required review targets, and covers T-0296..T-0306 by ID.
func TestCON_016_ArgusReviewNoP0P1Findings(t *testing.T) {
	raw, err := os.ReadFile(m16ReviewDocPath)
	if err != nil {
		t.Fatalf("reading M16 review %s: %v", m16ReviewDocPath, err)
	}
	doc := string(raw)

	if !strings.Contains(doc, "Zero unresolved P0 or P1 findings") {
		t.Fatalf("M16 review does not record zero unresolved P0/P1 findings")
	}
	if !strings.Contains(doc, "No finding requires an Eyvar waiver") {
		t.Fatalf("M16 review does not confirm no finding requires an Eyvar waiver")
	}
	// Honesty: it is recorded as an author review, not a fabricated external result.
	if !strings.Contains(doc, "author-conducted review") {
		t.Fatalf("M16 review must be honestly recorded as an author-conducted review")
	}

	// The three required review targets.
	for _, target := range []string{
		"never mutates ATTEST",
		"outside a major-version boundary",
		"downgrade or reattribute",
	} {
		if !strings.Contains(doc, target) {
			t.Fatalf("M16 review does not address required target %q", target)
		}
	}

	// Covers every M16 task by ID.
	for n := 296; n <= 306; n++ {
		id := fmt.Sprintf("T-%04d", n)
		if !strings.Contains(doc, id) {
			t.Fatalf("M16 review does not cover task %s by ID", id)
		}
	}
}
