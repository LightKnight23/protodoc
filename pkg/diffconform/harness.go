// Package diffconform is the M19/NFR-028 cross-implementation
// differential conformance harness (T-0350): it runs two implementations
// of container/validate/canon against the shared conformance corpus
// (docs/nfr-028-second-implementation-protocol.md's authoritative-corpus
// definition) and diffs their canonical octet output and reported
// verdicts, per corpus case, producing one pass/fail with a readable diff
// on mismatch.
//
// This package defines the harness mechanism only. It does not itself
// contain a second, independently-authored implementation -- none exists
// in this repository as of this task (that is a separate, disclosed
// governance gap; see docs/nfr-028-second-implementation-protocol.md's
// Non-goals section and the CQ-012 funding decision it points at). A
// Runner here can be this repository's own reference implementation, a
// future second implementation's compiled binary via BinaryRunner, or any
// other value satisfying the Runner interface (a synthetic fixture used
// to prove the harness itself catches a real disagreement, the same
// pattern pkg/benchconfig's audit_test.go already uses for its own
// drift check).
package diffconform

import (
	"bytes"
	"fmt"
)

// CorpusVersion is the authoritative conformance corpus version this
// harness's built-in Corpus() implements, per docs/nfr-028-second-
// implementation-protocol.md's "Authoritative corpus" section. Any
// addition, removal or edit to a corpus case in this package must bump
// this identifier so two reports naming different versions are never
// mistaken for comparable runs.
const CorpusVersion = "PDL-CONFCORPUS-M19-V1"

// Verdict is one implementation's reported accept-or-reject outcome for a
// corpus case, per contracts/cli.md Section 1's exit-code table: a
// human-readable status name plus the numeric exit code. Two Verdicts are
// equal (Go ==) only when both fields match exactly -- this package never
// treats "close enough" statuses as a match, per the acceptance
// protocol's Match definition Section 2.
type Verdict struct {
	Status   string
	ExitCode int
}

// CaseResult is one implementation's complete output for one corpus case:
// the verdict it reported, and the canonical octets it produced (nil when
// the case exercises verdict-only behaviour, e.g. a negative/malformed
// input that produces no canonical output at all).
type CaseResult struct {
	Verdict   Verdict
	Canonical []byte
}

// Runner is one implementation under comparison. Run is given one corpus
// case's raw input octets and must return that implementation's complete
// result for it. Run must not mutate caseInput.
type Runner interface {
	Run(caseInput []byte) (CaseResult, error)
}

// CorpusCase is one entry of the shared conformance corpus: a name (used
// in reports and t.Run subtests) and the raw input octets fed identically
// to both Runners under comparison.
type CorpusCase struct {
	Name  string
	Input []byte
}

// Diff is one corpus case's comparison outcome between two Runners named
// A and B in the Report that produced it. A zero-value Diff (both
// mismatch flags false) never appears in a Report.Diffs slice; Report
// only records cases that actually disagree.
type Diff struct {
	Case string

	OctetMismatch bool
	OctetDetail   string // human-readable: first differing offset, or a length mismatch

	VerdictMismatch bool
	VerdictA        Verdict
	VerdictB        Verdict

	ErrA error // non-nil if Runner A failed to produce a result at all
	ErrB error // non-nil if Runner B failed to produce a result at all
}

// String renders d as a single readable line, suitable for CI log output
// or a CLI's stderr on mismatch.
func (d Diff) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "case %q: ", d.Case)
	if d.ErrA != nil || d.ErrB != nil {
		fmt.Fprintf(&b, "runner error(s): A=%v B=%v", d.ErrA, d.ErrB)
		return b.String()
	}
	parts := 0
	if d.VerdictMismatch {
		fmt.Fprintf(&b, "verdict mismatch: A={%s,%d} B={%s,%d}", d.VerdictA.Status, d.VerdictA.ExitCode, d.VerdictB.Status, d.VerdictB.ExitCode)
		parts++
	}
	if d.OctetMismatch {
		if parts > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, "canonical octet mismatch: %s", d.OctetDetail)
		parts++
	}
	if parts == 0 {
		b.WriteString("no mismatch (Diff should not have been recorded)")
	}
	return b.String()
}

// Report is one full corpus comparison run's outcome.
type Report struct {
	CorpusVersion string
	Total         int
	Diffs         []Diff
}

// AllMatch reports whether every corpus case produced identical canonical
// octets and identical verdicts across both Runners. Per the acceptance
// protocol's Match definition Section 3, this is an absolute 100% bar:
// even one disagreeing case fails the run.
func (r Report) AllMatch() bool {
	return len(r.Diffs) == 0
}

// String renders a full human-readable summary, one line per mismatched
// case, suitable for a CI job's failure output.
func (r Report) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "differential conformance report (corpus %s): %d/%d cases matched", r.CorpusVersion, r.Total-len(r.Diffs), r.Total)
	if r.AllMatch() {
		return b.String()
	}
	b.WriteString("\n")
	for _, d := range r.Diffs {
		b.WriteString(d.String())
		b.WriteString("\n")
	}
	return b.String()
}

// octetDiffDetail returns a human-readable description of the first way
// a and b differ: a length mismatch, or the offset and octet values of
// the first differing byte. Returns "" if a and b are byte-equal.
func octetDiffDetail(a, b []byte) string {
	if bytes.Equal(a, b) {
		return ""
	}
	if len(a) != len(b) {
		return fmt.Sprintf("length mismatch: A=%d octets, B=%d octets", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			return fmt.Sprintf("first differing octet at offset %d: A=0x%02x B=0x%02x (of %d octets total)", i, a[i], b[i], len(a))
		}
	}
	// Unreachable: bytes.Equal already returned false for equal length,
	// equal-content slices only if lengths differ (handled above) or some
	// byte differs (handled in the loop).
	return "octets differ (unable to localize)"
}

// Run executes every case in cases against both a and b, diffs each
// pair's canonical octets and verdict, and returns the accumulated
// Report. Run itself only returns a non-nil error for a harness-level
// failure unrelated to any single case's comparison (there is none in
// this implementation; the return signature is kept for future runners
// that may need to report setup failures); a and b returning per-case
// errors from their own Run methods is captured in the corresponding
// Diff instead of aborting the whole comparison, so one broken case does
// not hide every other case's result.
func Run(corpusVersion string, cases []CorpusCase, a, b Runner) (Report, error) {
	report := Report{CorpusVersion: corpusVersion, Total: len(cases)}

	for _, c := range cases {
		resA, errA := a.Run(c.Input)
		resB, errB := b.Run(c.Input)

		if errA != nil || errB != nil {
			report.Diffs = append(report.Diffs, Diff{Case: c.Name, ErrA: errA, ErrB: errB})
			continue
		}

		var diff Diff
		diff.Case = c.Name

		if resA.Verdict != resB.Verdict {
			diff.VerdictMismatch = true
			diff.VerdictA = resA.Verdict
			diff.VerdictB = resB.Verdict
		}

		if detail := octetDiffDetail(resA.Canonical, resB.Canonical); detail != "" {
			diff.OctetMismatch = true
			diff.OctetDetail = detail
		}

		if diff.VerdictMismatch || diff.OctetMismatch {
			report.Diffs = append(report.Diffs, diff)
		}
	}

	return report, nil
}
