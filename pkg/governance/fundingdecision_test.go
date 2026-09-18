package governance

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// changesPath and planPath locate this repository's root-level T-0349
// governance artifacts relative to this package directory (pkg/governance),
// two levels up from the module root, following the same "read the live
// document off disk" convention as licencedraft_test.go.
const (
	changesPath = "../../specs/CHANGES.md"
	planPath    = "../../specs/001-protodoc-format-core/plan.md"
)

// fundingEntryPattern locates T-0349's structured CHANGES.md entry and
// captures its four required fields loosely enough to survive minor prose
// changes around them, while still failing if a field is missing outright.
var (
	fundingHeadingPattern   = regexp.MustCompile(`### T-0349 .+\(RESOLVED\)`)
	fundingDatePattern      = regexp.MustCompile(`\*\*Date:\*\*\s*\S+`)
	fundingDeciderPattern   = regexp.MustCompile(`\*\*Decision-maker:\*\*\s*Eyvar Garc[ií]a`)
	fundingPathPattern      = regexp.MustCompile(`\*\*Path-forward:\*\*\s*\S+`)
	fundingFallbackPattern  = regexp.MustCompile(`\*\*Fallback-if-unfunded:\*\*\s*\S+`)
	planResolvedRiskPattern = regexp.MustCompile(`CQ-012/CP-003 two-implementation gate[\s\S]*?RESOLVED[\s\S]*?Eyvar Garc[ií]a`)
)

// checkFundingDecision runs every T-0349 structural assertion against the
// CHANGES.md and plan.md text and returns one message per violation found
// (nil if both pass). Shared by the production check and a synthetic
// check so the two cannot drift into checking different things, mirroring
// checkGovernanceDoc's pattern in licencedraft_test.go.
func checkFundingDecision(changesText, planText string) []string {
	var problems []string

	if !fundingHeadingPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing T-0349 RESOLVED heading")
	}
	if !fundingDatePattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Date field on T-0349 entry")
	}
	if !fundingDeciderPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Decision-maker == 'Eyvar García' on T-0349 entry")
	}
	if !fundingPathPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Path-forward field on T-0349 entry")
	}
	if !fundingFallbackPattern.MatchString(changesText) {
		problems = append(problems, "CHANGES.md: missing Fallback-if-unfunded field on T-0349 entry")
	}

	if !planResolvedRiskPattern.MatchString(planText) {
		problems = append(problems, "plan.md Section 8: CQ-012/CP-003 funding risk row not marked RESOLVED with Eyvar García recorded")
	}

	return problems
}

// TestNFR_028_CQ012FundingDecisionRecorded is T-0349's named test. It
// asserts Eyvar García's community-bounty funding decision for the second
// independent implementation (CP-003/NFR-028's release gate) is recorded
// as a structured, dated entry in both specs/CHANGES.md and plan.md
// Section 8's risk row, with all four required fields present.
func TestNFR_028_CQ012FundingDecisionRecorded(t *testing.T) {
	changesRaw, err := os.ReadFile(changesPath)
	if err != nil {
		t.Fatalf("reading %s: %v", changesPath, err)
	}
	planRaw, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading %s: %v", planPath, err)
	}

	for _, problem := range checkFundingDecision(string(changesRaw), string(planRaw)) {
		t.Errorf("%s", problem)
	}
}

// TestNFR_028_CQ012FundingDecisionRecorded_DetectsSyntheticViolation proves
// checkFundingDecision fails closed: a synthetic OPEN (not yet decided)
// entry must be reported as missing every resolved-decision field, and a
// compliant synthetic entry built from the same required pieces must pass.
func TestNFR_028_CQ012FundingDecisionRecorded_DetectsSyntheticViolation(t *testing.T) {
	const openChanges = `### T-0349 — NFR-028 / CQ-012 two-implementation funding decision (OPEN, awaiting Eyvar García)

- **Status:** OPEN.
`
	const openPlan = `| The CQ-012/CP-003 two-implementation gate may never be cleared | high | Recommend Eyvar decide | Before phase 5 closes |`

	problems := checkFundingDecision(openChanges, openPlan)
	if len(problems) == 0 {
		t.Fatal("checkFundingDecision: synthetic OPEN entries reported zero problems, want several")
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		"missing T-0349 RESOLVED heading",
		"missing Date field",
		"missing Decision-maker == 'Eyvar García'",
		"missing Path-forward field",
		"missing Fallback-if-unfunded field",
		"risk row not marked RESOLVED",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkFundingDecision(synthetic OPEN): missing expected problem containing %q; got:\n%s", want, joined)
		}
	}

	const resolvedChanges = `### T-0349 — NFR-028 / CQ-012 two-implementation funding decision (RESOLVED)

- **Date:** 2026-09-18
- **Decision-maker:** Eyvar García
- **Path-forward:** community bounty.
- **Fallback-if-unfunded:** defer T-0351 indefinitely.
`
	const resolvedPlan = `| The CQ-012/CP-003 two-implementation gate may never be cleared | high (RESOLVED 2026-09-18) | Decision recorded 2026-09-18 by Eyvar García | Before phase 5 closes |`

	if problems := checkFundingDecision(resolvedChanges, resolvedPlan); len(problems) != 0 {
		t.Errorf("checkFundingDecision(synthetic resolved) = %v, want zero problems", problems)
	}
}
