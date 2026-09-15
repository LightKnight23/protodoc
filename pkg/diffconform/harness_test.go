package diffconform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureSourceGood is a minimal, stdlib-only TR-012-style CLI fixture:
// it echoes its input file back unchanged as its canonical output and
// always reports {"status":"OK","exit_code":0}. It stands in for a
// correctly-agreeing implementation in the tests below; it is not, and
// does not claim to be, a real second Protodoc implementation (none
// exists in this repository -- see docs/nfr-028-second-implementation-
// protocol.md's Non-goals section).
const fixtureSourceGood = `package main

import (
	"encoding/json"
	"os"
)

func main() {
	in, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[4], in, 0o600); err != nil {
		panic(err)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"verb": os.Args[1], "status": "OK", "exit_code": 0, "findings": []any{},
	})
}
`

// fixtureSourceMutatedBoth deliberately disagrees with fixtureSourceGood
// on BOTH dimensions the harness scores: it flips the first octet of its
// canonical output (when the input is non-empty) and reports a different
// verdict ({"status":"INVALID","exit_code":1}). Used to prove Run catches
// a genuinely mismatched implementation pair on both axes at once.
const fixtureSourceMutatedBoth = `package main

import (
	"encoding/json"
	"os"
)

func main() {
	in, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	if len(in) > 0 {
		in[0] ^= 0xFF
	}
	if err := os.WriteFile(os.Args[4], in, 0o600); err != nil {
		panic(err)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"verb": os.Args[1], "status": "INVALID", "exit_code": 1, "findings": []any{},
	})
}
`

// fixtureSourceMutatedVerdictOnly agrees with fixtureSourceGood's
// canonical octets exactly but disagrees only on verdict. Used to prove
// Run's verdict comparison is independent of its octet comparison: a
// case with byte-identical canonical output can still be flagged as a
// mismatch on verdict alone.
const fixtureSourceMutatedVerdictOnly = `package main

import (
	"encoding/json"
	"os"
)

func main() {
	in, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[4], in, 0o600); err != nil {
		panic(err)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"verb": os.Args[1], "status": "INVALID", "exit_code": 1, "findings": []any{},
	})
}
`

// buildFixtureBinary compiles src (a complete, stdlib-only main package)
// into a fresh executable under t.TempDir() and returns its path. It
// fails the test via t.Fatalf on any compile error.
func buildFixtureBinary(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0o600); err != nil {
		t.Fatalf("writing fixture source %s: %v", name, err)
	}
	exePath := filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", exePath, srcPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building fixture binary %s: %v\n%s", name, err, out)
	}
	return exePath
}

// TestNFR_028_DifferentialHarnessDiffsCanonicalOctets is T-0350's named
// test. It runs the harness's built-in corpus (Corpus(), CorpusVersion)
// through BinaryRunner against real compiled fixture binaries -- not
// in-process mocks -- proving the harness's actual subprocess-invocation,
// stdout-envelope-parsing and output-file-reading machinery works end to
// end, and that a deliberately-mismatched implementation pair is caught
// on both the octet dimension and the verdict dimension independently.
func TestNFR_028_DifferentialHarnessDiffsCanonicalOctets(t *testing.T) {
	corpus := Corpus()
	if len(corpus) == 0 {
		t.Fatal("Corpus() returned zero cases")
	}
	for _, c := range corpus {
		if len(c.Input) == 0 {
			t.Fatalf("corpus case %q has empty Input", c.Name)
		}
	}

	goodExe := buildFixtureBinary(t, "good", fixtureSourceGood)
	mutatedBothExe := buildFixtureBinary(t, "mutated-both", fixtureSourceMutatedBoth)
	mutatedVerdictExe := buildFixtureBinary(t, "mutated-verdict", fixtureSourceMutatedVerdictOnly)

	good := BinaryRunner{Path: goodExe, Verb: "publish"}

	t.Run("two agreeing binaries produce a clean report", func(t *testing.T) {
		other := BinaryRunner{Path: goodExe, Verb: "publish"}
		report, err := Run(CorpusVersion, corpus, good, other)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !report.AllMatch() {
			t.Fatalf("Run(good, good) = %+v, want AllMatch() true; report:\n%s", report, report)
		}
		if report.Total != len(corpus) {
			t.Errorf("report.Total = %d, want %d", report.Total, len(corpus))
		}
	})

	t.Run("byte-for-byte octet mismatch is caught on every case", func(t *testing.T) {
		mutated := BinaryRunner{Path: mutatedBothExe, Verb: "publish"}
		report, err := Run(CorpusVersion, corpus, good, mutated)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if report.AllMatch() {
			t.Fatal("Run(good, mutatedBoth) reported AllMatch() true, want disagreement caught")
		}
		if len(report.Diffs) != len(corpus) {
			t.Fatalf("len(report.Diffs) = %d, want %d (every case should disagree)", len(report.Diffs), len(corpus))
		}
		for _, d := range report.Diffs {
			if !d.OctetMismatch {
				t.Errorf("case %q: OctetMismatch = false, want true; diff: %s", d.Case, d)
			}
			if !strings.Contains(d.OctetDetail, "offset 0") {
				t.Errorf("case %q: OctetDetail = %q, want it to localize the mismatch to offset 0", d.Case, d.OctetDetail)
			}
			if !d.VerdictMismatch {
				t.Errorf("case %q: VerdictMismatch = false, want true (mutatedBoth also disagrees on verdict)", d.Case)
			}
		}
	})

	t.Run("verdict-only mismatch is caught independently of octet equality", func(t *testing.T) {
		mutated := BinaryRunner{Path: mutatedVerdictExe, Verb: "publish"}
		report, err := Run(CorpusVersion, corpus, good, mutated)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if report.AllMatch() {
			t.Fatal("Run(good, mutatedVerdictOnly) reported AllMatch() true, want disagreement caught")
		}
		if len(report.Diffs) != len(corpus) {
			t.Fatalf("len(report.Diffs) = %d, want %d (every case should disagree on verdict)", len(report.Diffs), len(corpus))
		}
		for _, d := range report.Diffs {
			if d.OctetMismatch {
				t.Errorf("case %q: OctetMismatch = true, want false (this fixture only mutates the verdict)", d.Case)
			}
			if !d.VerdictMismatch {
				t.Errorf("case %q: VerdictMismatch = false, want true", d.Case)
			}
			if d.VerdictA != (Verdict{Status: "OK", ExitCode: 0}) {
				t.Errorf("case %q: VerdictA = %+v, want {OK 0}", d.Case, d.VerdictA)
			}
			if d.VerdictB != (Verdict{Status: "INVALID", ExitCode: 1}) {
				t.Errorf("case %q: VerdictB = %+v, want {INVALID 1}", d.Case, d.VerdictB)
			}
		}
	})

	t.Run("Report.String names every mismatched case", func(t *testing.T) {
		mutated := BinaryRunner{Path: mutatedBothExe, Verb: "publish"}
		report, err := Run(CorpusVersion, corpus, good, mutated)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		summary := report.String()
		if !strings.Contains(summary, CorpusVersion) {
			t.Errorf("Report.String() = %q, want it to cite corpus version %q", summary, CorpusVersion)
		}
		for _, c := range corpus {
			if !strings.Contains(summary, c.Name) {
				t.Errorf("Report.String() does not mention mismatched case %q:\n%s", c.Name, summary)
			}
		}
	})
}
