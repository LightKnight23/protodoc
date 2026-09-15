package governance

import (
	"os"
	"strings"
	"testing"
)

// securityPolicyPath locates this repository's root-level CP-012
// disclosure-policy artifact relative to this package directory
// (pkg/governance), two levels up from the module root, following the
// same "read the live document off disk" convention as
// licencedraft_test.go and secondimplprotocol_test.go.
const securityPolicyPath = "../../SECURITY.md"

// requiredSecuritySections are the headed sections CP-012 and T-0372's
// DoD require SECURITY.md to carry, each as its own Markdown heading.
var requiredSecuritySections = []string{
	"## Triage Owner",
	"## Disclosure Clock",
}

// checkSecurityDoc runs every T-0372 structural/honesty assertion against
// text and returns one message per violation found (nil if text passes
// every check). Shared by the production check against the real document
// and the synthetic-violation test below, so the two can never drift into
// checking different things -- the same pattern checkGovernanceDoc and
// checkProtocolDoc already establish for CON-026 and NFR-028.
func checkSecurityDoc(text string) []string {
	var problems []string

	bodies := sectionBodies(text)
	for _, section := range requiredSecuritySections {
		body, ok := bodies[section]
		if !ok {
			problems = append(problems, "missing required section "+section)
			continue
		}
		if body == "" {
			problems = append(problems, "required section "+section+" has an empty body")
		}
	}

	triageBody := bodies["## Triage Owner"]
	if triageBody != "" && !strings.Contains(triageBody, "@") {
		problems = append(problems, "Triage Owner section must name an owner reachable by a contact (e.g. an email address)")
	}

	if !strings.Contains(text, "90") {
		problems = append(problems, "must publish a 90-day report-to-fix-or-advisory disclosure clock")
	}
	if !strings.Contains(text, "report") || !strings.Contains(text, "fix") || !strings.Contains(text, "advisory") {
		problems = append(problems, "disclosure clock must describe the report-to-fix-or-advisory path CP-012 requires")
	}

	if strings.Contains(text, "Status: APPROVED") || strings.Contains(text, "approved by Eyvar") {
		problems = append(problems, "must not claim Eyvar approval; no such approval has been recorded")
	}
	if !strings.Contains(text, "Status: DRAFT") {
		problems = append(problems, "must honestly state DRAFT status pending Eyvar's review")
	}

	return problems
}

// TestCP_012_TriageOwnerAndDisclosureSLAPublished is T-0372's named test.
// It reads the live SECURITY.md document off disk and asserts:
//
//  1. it names a triage owner (by role, with a reachable contact) in its
//     own headed section;
//  2. it publishes a 90-day report-to-fix-or-advisory disclosure clock in
//     its own headed section, with both fields non-empty;
//  3. the document is honest about its own status: it does not claim
//     Eyvar has already approved it, since no such approval has actually
//     happened.
func TestCP_012_TriageOwnerAndDisclosureSLAPublished(t *testing.T) {
	raw, err := os.ReadFile(securityPolicyPath)
	if err != nil {
		t.Fatalf("governance: reading %s: %v", securityPolicyPath, err)
	}

	for _, problem := range checkSecurityDoc(string(raw)) {
		t.Errorf("governance: %s: %s", securityPolicyPath, problem)
	}
}

// TestCP_012_TriageOwnerAndDisclosureSLAPublished_DetectsSyntheticViolation
// proves checkSecurityDoc fails closed: a synthetic document missing the
// disclosure clock, with an unreachable/anonymous triage owner, and
// falsely claiming approval must be reported as failing on every one of
// those points, and a compliant document built from the same required
// pieces must pass with zero problems.
func TestCP_012_TriageOwnerAndDisclosureSLAPublished_DetectsSyntheticViolation(t *testing.T) {
	const badDoc = `# Fake security policy

Status: APPROVED, approved by Eyvar.

## Triage Owner

Someone will handle it.
`
	problems := checkSecurityDoc(badDoc)
	if len(problems) == 0 {
		t.Fatal("checkSecurityDoc: synthetic non-compliant document reported zero problems, want several")
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		"must not claim Eyvar approval",
		"must honestly state DRAFT status",
		"missing required section ## Disclosure Clock",
		"reachable by a contact",
		"90-day report-to-fix-or-advisory",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkSecurityDoc(synthetic bad doc): missing expected problem containing %q; got:\n%s", want, joined)
		}
	}

	compliant := "Status: DRAFT\n\n" +
		"## Triage Owner\n\nheld by owner@example.com.\n\n" +
		"## Disclosure Clock\n\n90 days from report to fix or advisory.\n"
	if problems := checkSecurityDoc(compliant); len(problems) != 0 {
		t.Errorf("checkSecurityDoc(synthetic compliant doc) = %v, want zero problems", problems)
	}
}
