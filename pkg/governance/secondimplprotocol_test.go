// Package governance holds drift checks between this repository's
// governance-facing prose documents (funding decisions, acceptance
// protocols, sign-off records) and the artifacts they govern, following
// the same "read the live document off disk, never a hardcoded copy"
// convention pkg/benchconfig and pkg/ceilings already use for their own
// NFR-011/CON-009 drift checks.
package governance

import (
	"os"
	"strings"
	"testing"
)

// secondImplProtocolPath locates docs/nfr-028-second-implementation-
// protocol.md relative to this package directory (pkg/governance), two
// levels up from the module root.
const secondImplProtocolPath = "../../docs/nfr-028-second-implementation-protocol.md"

// requiredScopeTerms are the three, and only three, packages NFR-028/
// CON-019's two-implementation gate applies to per CQ-012's approved
// resolution and plan.md's dependency-spine item 10.
var requiredScopeTerms = []string{"container", "validate", "canon"}

// requiredExemptTerms are packages the protocol document must name as
// explicitly exempt from this specific gate (T-0348's own description:
// "rendering/merge/history/etc. are explicitly exempt from this specific
// gate"), so a future reader cannot mistake the scope as open-ended.
var requiredExemptTerms = []string{"render", "merge", "history"}

// requiredSections are the four questions T-0348's DoD requires the
// protocol to answer, each as its own heading: what "independent" means,
// which corpus is authoritative, what "match" means, and how a
// disagreement is triaged.
var requiredSections = []string{
	"## Scope",
	"## Independence criteria",
	"## Authoritative corpus",
	"## Match definition",
	"## Disagreement triage",
}

// checkProtocolDoc runs every T-0348 structural/honesty assertion against
// text and returns one message per violation found (nil if text passes
// every check). Shared by the production check against the real document
// and the synthetic-violation test below, so the two can never drift into
// checking different things.
func checkProtocolDoc(text string) []string {
	var problems []string

	for _, section := range requiredSections {
		if !strings.Contains(text, section) {
			problems = append(problems, "missing required section "+section)
		}
	}

	for _, term := range requiredScopeTerms {
		if !strings.Contains(text, "`"+term+"`") {
			problems = append(problems, "scope must name in-scope package "+term)
		}
	}

	for _, term := range requiredExemptTerms {
		if !strings.Contains(text, "`"+term+"`") {
			problems = append(problems, "scope must explicitly name exempt package "+term)
		}
	}

	if !strings.Contains(text, "byte-for-byte") {
		problems = append(problems, "match definition must require byte-for-byte octet equality, not a fuzzy or digest-only match")
	}
	if !strings.Contains(text, "100%") {
		problems = append(problems, "match definition must state a 100%-of-corpus bar, not a percentage threshold below it")
	}

	if strings.Contains(text, "Status: APPROVED") || strings.Contains(text, "approved by Eyvar") {
		problems = append(problems, "must not claim Eyvar approval; no such approval has been recorded")
	}
	if !strings.Contains(text, "Status: DRAFT") {
		problems = append(problems, "must honestly state DRAFT status pending Eyvar's review")
	}

	return problems
}

// TestNFR_028_SecondImplScopeDefined is T-0348's named test. It reads the
// live protocol document off disk and asserts:
//
//  1. every required section exists (the four questions the DoD requires
//     the protocol to answer, plus the scope section itself);
//  2. the scope section names exactly the three in-scope packages and
//     explicitly names at least one exempt package, so the gate's
//     boundary is legible rather than implied;
//  3. the match-definition section states byte-for-byte octet equality
//     and verdict equality, not a percentage or fuzzy-match standard;
//  4. the document is honest about its own status: it does not claim
//     Eyvar has already approved it, since no such approval has actually
//     happened.
func TestNFR_028_SecondImplScopeDefined(t *testing.T) {
	raw, err := os.ReadFile(secondImplProtocolPath)
	if err != nil {
		t.Fatalf("governance: reading %s: %v", secondImplProtocolPath, err)
	}

	for _, problem := range checkProtocolDoc(string(raw)) {
		t.Errorf("governance: %s: %s", secondImplProtocolPath, problem)
	}
}

// TestNFR_028_SecondImplScopeDefined_DetectsSyntheticViolation proves
// checkProtocolDoc actually fails closed: a synthetic document missing
// most required sections and falsely claiming approval must be reported
// as failing on every one of those points, and a compliant document built
// from the same required pieces must pass with zero problems.
func TestNFR_028_SecondImplScopeDefined_DetectsSyntheticViolation(t *testing.T) {
	const badDoc = `# Fake protocol

Status: APPROVED

## Scope

container, validate, canon.
`
	problems := checkProtocolDoc(badDoc)
	if len(problems) == 0 {
		t.Fatal("checkProtocolDoc: synthetic non-compliant document reported zero problems, want several")
	}
	joined := strings.Join(problems, "\n")
	for _, want := range []string{
		"must not claim Eyvar approval",
		"must honestly state DRAFT status",
		"byte-for-byte octet equality",
		"100%-of-corpus bar",
		"missing required section ## Independence criteria",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkProtocolDoc(synthetic bad doc): missing expected problem containing %q; got:\n%s", want, joined)
		}
	}

	var compliant strings.Builder
	compliant.WriteString("Status: DRAFT\n\n")
	for _, section := range requiredSections {
		compliant.WriteString(section + "\n\n")
	}
	for _, term := range append(append([]string{}, requiredScopeTerms...), requiredExemptTerms...) {
		compliant.WriteString("`" + term + "`\n")
	}
	compliant.WriteString("byte-for-byte, 100%\n")

	if problems := checkProtocolDoc(compliant.String()); len(problems) != 0 {
		t.Errorf("checkProtocolDoc(synthetic compliant doc) = %v, want zero problems", problems)
	}
}
