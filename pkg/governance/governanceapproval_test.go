package governance

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// changesEntryPattern locates T-0356's structured CHANGES.md entry and its
// required fields, mirroring fundingdecision_test.go's pattern for T-0349.
var (
	governanceHeadingResolvedPattern  = regexp.MustCompile(`### T-0356 .+\(RESOLVED\)`)
	governanceDatePattern             = regexp.MustCompile(`\*\*Date:\*\*\s*\S+`)
	governanceDeciderPattern          = regexp.MustCompile(`\*\*Decision-maker:\*\*\s*Eyvar Garc[ií]a`)
	governanceDecisionApprovedPattern = regexp.MustCompile(`\*\*Decision:\*\*\s*approved`)
	governanceFilesPattern            = regexp.MustCompile(`\*\*Referenced-files:\*\*.*LICENSE.*GOVERNANCE\.md`)
	governanceDocApprovedPattern      = regexp.MustCompile(`Status:\s*APPROVED by Eyvar Garc[ií]a`)
)

// checkGovernanceApproval runs every T-0356 structural assertion against the
// CHANGES.md and GOVERNANCE.md text and returns one message per violation
// found (nil if both pass).
func checkGovernanceApproval(changesText, governanceText string) []string {
	var problems []string

	if !governanceHeadingResolvedPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing T-0356 RESOLVED heading")
	}
	if !governanceDatePattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Date field on T-0356 entry")
	}
	if !governanceDeciderPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Decision-maker == 'Eyvar García' on T-0356 entry")
	}
	if !governanceDecisionApprovedPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Decision: approved on T-0356 entry")
	}
	if !governanceFilesPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Referenced-files naming LICENSE and GOVERNANCE.md on T-0356 entry")
	}

	if !governanceDocApprovedPattern.MatchString(governanceText) {
		problems = append(problems, "GOVERNANCE.md: Status line does not read APPROVED by Eyvar García")
	}
	if strings.Contains(governanceText, "Status: DRAFT") {
		problems = append(problems, "GOVERNANCE.md: still declares itself DRAFT after approval")
	}

	return problems
}

// TestCON_026_GovernanceGateApprovedAndOnRecord is T-0356's named test. It
// asserts Eyvar García's approval of CON-026's four governance elements
// (licence, named steward, succession, deprecation window) is recorded as a
// structured, dated entry in specs/CHANGES.md, and that GOVERNANCE.md itself
// reflects the approved status rather than remaining a draft.
func TestCON_026_GovernanceGateApprovedAndOnRecord(t *testing.T) {
	changesRaw, err := os.ReadFile(changesPath)
	if err != nil {
		t.Fatalf("reading %s: %v", changesPath, err)
	}
	govRaw, err := os.ReadFile(governancePath)
	if err != nil {
		t.Fatalf("reading %s: %v", governancePath, err)
	}

	for _, problem := range checkGovernanceApproval(string(changesRaw), string(govRaw)) {
		t.Errorf("%s", problem)
	}
}

// TestCON_026_GovernanceGateApprovedAndOnRecord_DetectsSyntheticViolation
// proves checkGovernanceApproval fails closed: a synthetic OPEN/DRAFT state
// must be reported as missing every approval field, and a compliant
// synthetic state built from the same required pieces must pass.
func TestCON_026_GovernanceGateApprovedAndOnRecord_DetectsSyntheticViolation(t *testing.T) {
	const openChanges = `### T-0356 — CON-026 licence/governance gate approval (OPEN, awaiting Eyvar García)

- **Status:** OPEN.
`
	const draftGovernance = `Status: DRAFT -- pending Eyvar Garcia's review and approval.`

	problems := checkGovernanceApproval(openChanges, draftGovernance)
	if len(problems) == 0 {
		t.Fatal("checkGovernanceApproval: synthetic OPEN/DRAFT state reported zero problems, want several")
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		"missing T-0356 RESOLVED heading",
		"missing Date field",
		"missing Decision-maker == 'Eyvar García'",
		"missing Decision: approved",
		"missing Referenced-files",
		"Status line does not read APPROVED",
		"still declares itself DRAFT",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkGovernanceApproval(synthetic OPEN/DRAFT): missing expected problem containing %q; got:\n%s", want, joined)
		}
	}

	const resolvedChanges = `### T-0356 — CON-026 licence/governance gate approval (RESOLVED)

- **Date:** 2026-09-19
- **Decision-maker:** Eyvar García
- **Decision:** approved
- **Referenced-files:** LICENSE, LICENSE-CODE, GOVERNANCE.md
`
	const approvedGovernance = `Status: APPROVED by Eyvar Garcia, 2026-09-19.`

	if problems := checkGovernanceApproval(resolvedChanges, approvedGovernance); len(problems) != 0 {
		t.Errorf("checkGovernanceApproval(synthetic resolved) = %v, want zero problems", problems)
	}
}
