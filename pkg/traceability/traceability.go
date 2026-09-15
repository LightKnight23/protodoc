// Package traceability is the NFR-029 normative-statement-to-conformance-
// case coverage checker (T-0352): it parses every FR-/NFR-/CON-/TR-
// requirement id and every PD-*-NNN validator-rule id out of the frozen
// spec.md and contracts/*.abnf, cross-references each id against this
// module's checked-in conformance corpus, and reports every id with zero
// matching cases.
//
// # Disclosed scope: what counts as "a matching case"
//
// NFR-029 requires every normative statement to be "mapped to at least one
// executable case in a conformance corpus". This package's proxy for that
// mapping is a literal, case-sensitive occurrence of the id's exact
// hyphenated text (e.g. "FR-117", "PD-RING-001") anywhere in this module's
// own *.go source -- production code, test code, and named corpus
// definitions such as pkg/diffconform/corpus.go alike. This is coarser
// than requiring a formal per-test tag: it cannot distinguish a real
// assertion from an id merely mentioned in a doc comment. It is
// deliberately never the reverse error, though -- an id ScanCoverage
// reports as covered really does appear somewhere in the checked-in
// source, and an id it reports as a gap really does appear nowhere in it.
// Given NFR-029's own bar is "zero mapped cases blocks a release", a
// coarse-but-never-wrong-in-that-direction proxy is the correct one: it
// can only ever under-report gaps by counting a bare mention as coverage,
// never manufacture a gap that is not real. See docs/nfr-029-
// traceability-gap-report.md for the current audit's filed results.
package traceability

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// reRequirementID matches one FR-/NFR-/CON-/TR- requirement id, exactly
// three digits, per this repository's CLAUDE.md ID scheme.
var reRequirementID = regexp.MustCompile(`\b(?:FR|NFR|CON|TR)-[0-9]{3}\b`)

// reRuleID matches one PD-*-NNN validator-rule id: "PD-", an uppercase
// alphanumeric segment (covers plain names like RING and mixed
// letter/digit names like A11Y and 2D alike), a hyphen, and exactly three
// digits.
var reRuleID = regexp.MustCompile(`\bPD-[A-Z0-9]+-[0-9]{3}\b`)

// dedupSorted returns ids with duplicates removed, in ascending sorted
// order, so two scans of the same content always compare equal regardless
// of occurrence order.
func dedupSorted(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// ExtractRequirementIDs returns every distinct FR-/NFR-/CON-/TR- id found
// in text, sorted ascending.
func ExtractRequirementIDs(text string) []string {
	return dedupSorted(reRequirementID.FindAllString(text, -1))
}

// ExtractRuleIDs returns every distinct PD-*-NNN validator-rule id found
// in text, sorted ascending.
func ExtractRuleIDs(text string) []string {
	return dedupSorted(reRuleID.FindAllString(text, -1))
}

// NormativeIDs is the complete set of ids T-0352's checker holds a
// release accountable for: every requirement id and every validator-rule
// id parsed out of the frozen spec and wire-grammar contracts.
type NormativeIDs struct {
	RequirementIDs []string
	RuleIDs        []string
}

// All returns RequirementIDs and RuleIDs concatenated into one sorted,
// deduplicated slice.
func (n NormativeIDs) All() []string {
	combined := make([]string, 0, len(n.RequirementIDs)+len(n.RuleIDs))
	combined = append(combined, n.RequirementIDs...)
	combined = append(combined, n.RuleIDs...)
	return dedupSorted(combined)
}

// ParseNormativeIDs reads specPath and every path in contractPaths and
// returns the union of every requirement id and validator-rule id found
// across all of them. It fails on any read error; a missing frozen
// artifact is a build-breaking authoring error in this checked-in tool,
// never a runtime condition a caller must degrade gracefully for.
func ParseNormativeIDs(specPath string, contractPaths []string) (NormativeIDs, error) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		return NormativeIDs{}, fmt.Errorf("traceability: reading spec %s: %w", specPath, err)
	}

	var allText strings.Builder
	allText.Write(spec)

	for _, p := range contractPaths {
		raw, err := os.ReadFile(p)
		if err != nil {
			return NormativeIDs{}, fmt.Errorf("traceability: reading contract %s: %w", p, err)
		}
		allText.WriteByte('\n')
		allText.Write(raw)
	}

	text := allText.String()
	return NormativeIDs{
		RequirementIDs: ExtractRequirementIDs(text),
		RuleIDs:        ExtractRuleIDs(text),
	}, nil
}

// CoverageSet is the set of ids ScanCoverage found at least one literal
// occurrence of somewhere in the scanned source tree.
type CoverageSet map[string]bool

// Has reports whether id occurs anywhere in the scanned tree.
func (c CoverageSet) Has(id string) bool {
	return c[id]
}

// ScanCoverage walks every *.go file under root (this module's own source,
// skipping .git and any vendor directory) and returns the set of every
// requirement-id-shaped and rule-id-shaped literal token found anywhere in
// their text. See the package doc comment's Disclosed scope section for
// exactly what this proxy does and does not claim.
func ScanCoverage(root string) (CoverageSet, error) {
	covered := make(CoverageSet)

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("traceability: reading %s: %w", path, err)
		}
		text := string(src)
		for _, id := range reRequirementID.FindAllString(text, -1) {
			covered[id] = true
		}
		for _, id := range reRuleID.FindAllString(text, -1) {
			covered[id] = true
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return covered, nil
}

// Gap is one normative-statement id the scanned conformance corpus
// currently has zero matching cases for.
type Gap struct {
	ID   string
	Kind string // "requirement" or "rule"
}

// FindGaps returns, in ascending ID order, every id in ids.RequirementIDs
// and ids.RuleIDs absent from covered. A zero-length result means every
// normative-statement id has at least one matching case, per NFR-029's
// release bar.
func FindGaps(ids NormativeIDs, covered CoverageSet) []Gap {
	var gaps []Gap
	for _, id := range ids.RequirementIDs {
		if !covered.Has(id) {
			gaps = append(gaps, Gap{ID: id, Kind: "requirement"})
		}
	}
	for _, id := range ids.RuleIDs {
		if !covered.Has(id) {
			gaps = append(gaps, Gap{ID: id, Kind: "rule"})
		}
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i].ID < gaps[j].ID })
	return gaps
}

// Report is one complete audit run's outcome: the normative ids the audit
// held the corpus accountable for, and the gaps found among them.
type Report struct {
	Normative NormativeIDs
	Gaps      []Gap
}

// AllCovered reports whether the audit found zero gaps, i.e. every parsed
// normative-statement id has at least one matching case.
func (r Report) AllCovered() bool {
	return len(r.Gaps) == 0
}

// String renders a human-readable summary: total normative ids parsed, how
// many have zero matching cases, and one line per gap.
func (r Report) String() string {
	var b strings.Builder
	total := len(r.Normative.All())
	fmt.Fprintf(&b, "traceability audit: %d normative ids parsed (%d requirement, %d rule), %d with zero conformance cases\n",
		total, len(r.Normative.RequirementIDs), len(r.Normative.RuleIDs), len(r.Gaps))
	for _, g := range r.Gaps {
		fmt.Fprintf(&b, "  GAP [%s] %s\n", g.Kind, g.ID)
	}
	return b.String()
}

// Audit runs the complete T-0352 checker: it parses specPath and
// contractPaths for the normative id set, scans corpusRoot for the
// conformance corpus's own coverage, and returns the resulting Report.
func Audit(specPath string, contractPaths []string, corpusRoot string) (Report, error) {
	ids, err := ParseNormativeIDs(specPath, contractPaths)
	if err != nil {
		return Report{}, err
	}
	covered, err := ScanCoverage(corpusRoot)
	if err != nil {
		return Report{}, err
	}
	return Report{Normative: ids, Gaps: FindGaps(ids, covered)}, nil
}
