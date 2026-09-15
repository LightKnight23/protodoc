package traceability

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// Paths relative to this package directory (pkg/traceability), two levels
// up to the module root, matching the relative-path convention this
// module's other drift-check tests already use (pkg/ceilings/
// ceilings_test.go's dataModelPath, pkg/governance/
// secondimplprotocol_test.go's secondImplProtocolPath).
const (
	moduleRoot        = "../.."
	specPath          = moduleRoot + "/specs/001-protodoc-format-core/spec.md"
	gapReportDocPath  = moduleRoot + "/docs/nfr-029-traceability-gap-report.md"
	releaseGateEnvVar = "PROTODOC_RELEASE_GATE"
)

// contractPaths is the frozen wire-grammar contract set T-0352/T-0353
// hold every PD-*-NNN validator-rule id accountable against, per the
// task's own instruction to parse ids "out of spec.md and
// contracts/*.abnf".
var contractPaths = []string{
	moduleRoot + "/specs/001-protodoc-format-core/contracts/container.abnf",
	moduleRoot + "/specs/001-protodoc-format-core/contracts/document.abnf",
	moduleRoot + "/specs/001-protodoc-format-core/contracts/integrity.abnf",
}

// extractSection returns the text strictly between the first line
// containing startMarker and the first later line containing endMarker
// (or the end of text, if endMarker is empty or never found). Used to
// pull just the "### Requirement ids ..." / "### Rule ids ..." list
// sections out of docs/nfr-029-traceability-gap-report.md, so ids merely
// mentioned in that document's explanatory prose (e.g. "CON-002" inside a
// sentence about CON-002's own history) are never mistaken for a member
// of the checked-in enumerated gap list itself.
func extractSection(text, startMarker, endMarker string) string {
	startIdx := strings.Index(text, startMarker)
	if startIdx == -1 {
		return ""
	}
	rest := text[startIdx+len(startMarker):]
	if endMarker == "" {
		return rest
	}
	endIdx := strings.Index(rest, endMarker)
	if endIdx == -1 {
		return rest
	}
	return rest[:endIdx]
}

// checkedInGapIDs reads docs/nfr-029-traceability-gap-report.md and
// returns the requirement-id gap list and rule-id gap list it currently
// enumerates, parsed from its own "### Requirement ids ..." and "### Rule
// ids ..." sections.
func checkedInGapIDs(t *testing.T) (requirementGaps, ruleGaps []string) {
	t.Helper()
	raw, err := os.ReadFile(gapReportDocPath)
	if err != nil {
		t.Fatalf("traceability: reading %s: %v", gapReportDocPath, err)
	}
	text := string(raw)

	reqSection := extractSection(text, "### Requirement ids with zero conformance cases", "### Rule ids with zero conformance cases")
	ruleSection := extractSection(text, "### Rule ids with zero conformance cases", "## Release-gate mechanism")

	if reqSection == "" {
		t.Fatalf("traceability: %s: missing '### Requirement ids with zero conformance cases' section", gapReportDocPath)
	}
	if ruleSection == "" {
		t.Fatalf("traceability: %s: missing '### Rule ids with zero conformance cases' section", gapReportDocPath)
	}

	return ExtractRequirementIDs(reqSection), ExtractRuleIDs(ruleSection)
}

// TestNFR_029_AllNormativeStatementsHaveConformanceCase is T-0353's named
// test. It runs the live T-0352 audit against the frozen spec/contracts
// and this module's actual source tree, and enforces two independent
// things:
//
//  1. Always (every `go test ./...` run): the checked-in
//     docs/nfr-029-traceability-gap-report.md gap lists must exactly
//     match what a live audit finds right now. This is a drift check,
//     the same pattern pkg/ceilings and pkg/benchconfig already use for
//     their own checked-in-copy-of-a-computed-fact tests: the filed gap
//     report can never silently go stale, in either direction (a gap
//     that closes must be removed from the doc; a new gap must be added
//     to it), without this test catching it. This assertion is expected
//     to be red the day someone lands a task that closes a gap without
//     touching this file, and green again the moment they do -- that is
//     the report doing its job, not a flaky test.
//  2. Only when PROTODOC_RELEASE_GATE is set to a non-empty value: every
//     remaining gap fails the test individually, by id. This is the
//     literal enforcement of NFR-029's "a release SHALL be blocked while
//     any normative statement has zero mapped cases" and CP-011's "a
//     release is blocked while any normative statement has zero mapped
//     cases" -- a release process that does not set this variable and
//     see this test pass has not satisfied either requirement, regardless
//     of what else it checked. It is off by default so ordinary
//     development test runs stay green while this repository is
//     mid-implementation (specs/001-protodoc-format-core/tasks.md
//     schedules 372 tasks across 19 milestones; most of the 114 gaps
//     currently on file close as their owning milestone's tasks land, not
//     by this test being satisfied early).
func TestNFR_029_AllNormativeStatementsHaveConformanceCase(t *testing.T) {
	report, err := Audit(specPath, contractPaths, moduleRoot)
	if err != nil {
		t.Fatalf("Audit(%s, %v, %s): %v", specPath, contractPaths, moduleRoot, err)
	}

	var liveReqGaps, liveRuleGaps []string
	for _, g := range report.Gaps {
		switch g.Kind {
		case "requirement":
			liveReqGaps = append(liveReqGaps, g.ID)
		case "rule":
			liveRuleGaps = append(liveRuleGaps, g.ID)
		default:
			t.Fatalf("Gap %+v has unrecognised Kind %q", g, g.Kind)
		}
	}

	docReqGaps, docRuleGaps := checkedInGapIDs(t)

	if !reflect.DeepEqual(liveReqGaps, docReqGaps) {
		t.Errorf("traceability: %s requirement-id gap list has drifted from the live audit.\nlive:  %v\ndoc:   %v\nRegenerate with `go run ./cmd/protodoc-traceaudit` and update the checked-in report.",
			gapReportDocPath, liveReqGaps, docReqGaps)
	}
	if !reflect.DeepEqual(liveRuleGaps, docRuleGaps) {
		t.Errorf("traceability: %s rule-id gap list has drifted from the live audit.\nlive:  %v\ndoc:   %v\nRegenerate with `go run ./cmd/protodoc-traceaudit` and update the checked-in report.",
			gapReportDocPath, liveRuleGaps, docRuleGaps)
	}

	if !releaseGateBlocks(os.Getenv(releaseGateEnvVar), report.Gaps) {
		t.Logf("%s", report.String())
		t.Logf("traceability: release gate not active (set %s=1 to enforce NFR-029/CP-011 as a hard release block); %d gap(s) currently on file, see %s", releaseGateEnvVar, len(report.Gaps), gapReportDocPath)
		return
	}

	for _, g := range report.Gaps {
		t.Errorf("NFR-029/CP-011 release gate: %s %s has zero conformance cases; release blocked. See %s.", g.Kind, g.ID, gapReportDocPath)
	}
}

// releaseGateBlocks reports whether the NFR-029/CP-011 release gate must
// fail the build given the PROTODOC_RELEASE_GATE environment variable's
// raw value and the current audit's gaps: only when the gate is armed
// (envVal non-empty) AND at least one gap remains. Factored out as a pure
// function so TestNFR_029_AllNormativeStatementsHaveConformanceCase_
// ReleaseGateFailsClosed can prove the gate's actual truth table without
// needing to fake a *testing.T.
func releaseGateBlocks(envVal string, gaps []Gap) bool {
	return envVal != "" && len(gaps) > 0
}

// TestNFR_029_AllNormativeStatementsHaveConformanceCase_ReleaseGateFailsClosed
// proves the PROTODOC_RELEASE_GATE mechanism actually enforces NFR-029/
// CP-011 rather than being dead code, across its full truth table: gate
// off never blocks regardless of gaps; gate on blocks if and only if at
// least one gap remains.
func TestNFR_029_AllNormativeStatementsHaveConformanceCase_ReleaseGateFailsClosed(t *testing.T) {
	oneGap := []Gap{{ID: "FR-999", Kind: "requirement"}}

	cases := []struct {
		name   string
		envVal string
		gaps   []Gap
		want   bool
	}{
		{"gate off, gaps present", "", oneGap, false},
		{"gate off, no gaps", "", nil, false},
		{"gate on, gaps present", "1", oneGap, true},
		{"gate on, no gaps", "1", nil, false},
	}
	for _, c := range cases {
		if got := releaseGateBlocks(c.envVal, c.gaps); got != c.want {
			t.Errorf("releaseGateBlocks(%q, %v) = %v, want %v (%s)", c.envVal, c.gaps, got, c.want, c.name)
		}
	}
}
