package governance

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// licensePath and governancePath locate this repository's root-level
// CON-026 governance artifacts relative to this package directory
// (pkg/governance), two levels up from the module root, following the
// same "read the live document off disk" convention as
// secondimplprotocol_test.go and featureregister.go.
const (
	licensePath    = "../../LICENSE"
	governancePath = "../../GOVERNANCE.md"
)

// requiredGovernanceSections are the four headed sections CON-026 and
// T-0355's DoD require GOVERNANCE.md to carry, each as its own Markdown
// heading.
var requiredGovernanceSections = []string{
	"## Licence Grant",
	"## Named Steward",
	"## Succession Process",
	"## Deprecation Window Policy",
}

// governanceSectionPattern splits governance.md into (heading, body) pairs
// for every "## " heading, so a section's body can be checked for
// non-emptiness independent of section order.
var governanceHeadingPattern = regexp.MustCompile(`(?m)^## .+$`)

// sectionBodies splits text into a map from heading line to the
// (trimmed) body text following it, up to the next "## " heading or end of
// document.
func sectionBodies(text string) map[string]string {
	headings := governanceHeadingPattern.FindAllStringIndex(text, -1)
	bodies := make(map[string]string, len(headings))
	for i, h := range headings {
		heading := strings.TrimSpace(text[h[0]:h[1]])
		bodyStart := h[1]
		bodyEnd := len(text)
		if i+1 < len(headings) {
			bodyEnd = headings[i+1][0]
		}
		bodies[heading] = strings.TrimSpace(text[bodyStart:bodyEnd])
	}
	return bodies
}

// checkGovernanceDoc runs every T-0355 structural/honesty assertion
// against text and returns one message per violation found (nil if text
// passes every check). Shared by the production check against the real
// document and a synthetic-violation check, so the two can never drift
// into checking different things -- the same pattern
// secondimplprotocol_test.go's checkProtocolDoc already establishes for
// NFR-028.
func checkGovernanceDoc(text string) []string {
	var problems []string

	bodies := sectionBodies(text)
	for _, section := range requiredGovernanceSections {
		body, ok := bodies[section]
		if !ok {
			problems = append(problems, "missing required section "+section)
			continue
		}
		if body == "" {
			problems = append(problems, "required section "+section+" has an empty body")
		}
	}

	// The document must state exactly one of two honest states: still DRAFT
	// (no approval claimed), or APPROVED with the decision-maker named in the
	// same breath -- never an approval claim with no named decision-maker
	// (which is what "must not fabricate approval" actually guards against).
	// Once approval genuinely happens, cross-checking that claim against a
	// real specs/CHANGES.md entry is TestCON_026_GovernanceGateApprovedAndOnRecord's
	// job (T-0356), not this test's.
	isDraft := strings.Contains(text, "Status: DRAFT")
	isApproved := regexp.MustCompile(`Status:\s*APPROVED by Eyvar Garc[ií]a`).MatchString(text)
	if !isDraft && !isApproved {
		problems = append(problems, "must state either 'Status: DRAFT ...' or 'Status: APPROVED by Eyvar Garcia ...', found neither")
	}
	if isDraft && isApproved {
		problems = append(problems, "must not claim both DRAFT and APPROVED status simultaneously")
	}

	return problems
}

// TestCON_026_LicenceDraftPublished is T-0355's named test. It asserts:
//
//  1. LICENSE exists at the repository root and is non-empty;
//  2. GOVERNANCE.md exists and carries all four CON-026-required headed
//     sections (Licence Grant, Named Steward, Succession Process,
//     Deprecation Window Policy), each with a non-empty body;
//  3. GOVERNANCE.md is honest about its own status: it does not claim
//     Eyvar has already approved it, since T-0356 (a separate, real
//     external action) has not recorded any such approval.
func TestCON_026_LicenceDraftPublished(t *testing.T) {
	licenseRaw, err := os.ReadFile(licensePath)
	if err != nil {
		t.Fatalf("reading %s: %v", licensePath, err)
	}
	if strings.TrimSpace(string(licenseRaw)) == "" {
		t.Fatalf("%s exists but is empty", licensePath)
	}

	govRaw, err := os.ReadFile(governancePath)
	if err != nil {
		t.Fatalf("reading %s: %v", governancePath, err)
	}

	for _, problem := range checkGovernanceDoc(string(govRaw)) {
		t.Errorf("%s: %s", governancePath, problem)
	}
}

// TestCON_026_LicenceDraftPublished_DetectsSyntheticViolation proves
// checkGovernanceDoc fails closed: a synthetic document missing sections,
// with an empty section body, and falsely claiming approval must be
// reported as failing on every one of those points, and a compliant
// document built from the same required pieces must pass with zero
// problems.
func TestCON_026_LicenceDraftPublished_DetectsSyntheticViolation(t *testing.T) {
	const badDoc = `# Fake governance doc

Status: APPROVED, approved by Eyvar.

## Licence Grant

See LICENSE.

## Named Steward
`
	problems := checkGovernanceDoc(badDoc)
	if len(problems) == 0 {
		t.Fatal("checkGovernanceDoc: synthetic non-compliant document reported zero problems, want several")
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		"must state either 'Status: DRAFT",
		"missing required section ## Succession Process",
		"missing required section ## Deprecation Window Policy",
		"required section ## Named Steward has an empty body",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkGovernanceDoc(synthetic bad doc): missing expected problem containing %q; got:\n%s", want, joined)
		}
	}

	var compliantDraft strings.Builder
	compliantDraft.WriteString("Status: DRAFT\n\n")
	for _, section := range requiredGovernanceSections {
		compliantDraft.WriteString(section + "\n\nSome non-empty body text.\n\n")
	}
	if problems := checkGovernanceDoc(compliantDraft.String()); len(problems) != 0 {
		t.Errorf("checkGovernanceDoc(synthetic compliant draft doc) = %v, want zero problems", problems)
	}

	var compliantApproved strings.Builder
	compliantApproved.WriteString("Status: APPROVED by Eyvar Garcia, 2026-09-19.\n\n")
	for _, section := range requiredGovernanceSections {
		compliantApproved.WriteString(section + "\n\nSome non-empty body text.\n\n")
	}
	if problems := checkGovernanceDoc(compliantApproved.String()); len(problems) != 0 {
		t.Errorf("checkGovernanceDoc(synthetic compliant approved doc) = %v, want zero problems", problems)
	}
}
