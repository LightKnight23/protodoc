package fuzzmaturity

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const moduleRoot = "../.."

// knownM01Targets is the baseline set of fuzz targets T-0359 already
// registered for M01's untrusted-byte decode entry points
// (pkg/pdlfmt/fuzz_test.go, pkg/container/fuzz_test.go). This is a
// drift check, not an exhaustive list: DiscoverFuzzTargets is
// directory-driven (see its own doc comment), so M02/M03/M07/M09/M10/
// M14's fuzz targets are picked up automatically the moment they land,
// with no change needed here or in DiscoverFuzzTargets itself -- this
// list only proves discovery still finds what is known to exist today.
var knownM01Targets = []string{
	"FuzzCP_012_DecodeUntrustedBytes_CommitRingRecord",
	"FuzzCP_012_DecodeUntrustedBytes_Frontmatter",
	"FuzzCP_012_DecodeUntrustedBytes_Header",
	"FuzzCP_012_DecodeUntrustedBytes_SegmentTable",
	"FuzzCP_012_DecodeUntrustedBytes_TLV",
	"FuzzCP_012_DecodeUntrustedBytes_Varint",
}

func TestFuzzMaturity_DiscoverFuzzTargetsFindsKnownM01Targets(t *testing.T) {
	targets, err := DiscoverFuzzTargets(moduleRoot)
	if err != nil {
		t.Fatalf("DiscoverFuzzTargets(%s): %v", moduleRoot, err)
	}

	found := make(map[string]FuzzTarget, len(targets))
	for _, tg := range targets {
		found[tg.Name] = tg
	}
	for _, name := range knownM01Targets {
		tg, ok := found[name]
		if !ok {
			t.Errorf("DiscoverFuzzTargets: missing known target %s", name)
			continue
		}
		if tg.SeedCorpusSize < MinSeedCorpusSize {
			t.Errorf("target %s: SeedCorpusSize = %d, want >= %d (source: %s)", name, tg.SeedCorpusSize, MinSeedCorpusSize, tg.File)
		}
	}
}

func TestFuzzMaturity_ViolationsTruthTable(t *testing.T) {
	base := FuzzTarget{Name: "X", SeedCorpusSize: 5}
	cases := []struct {
		name   string
		result MaturityResult
		want   int
	}{
		{"clean", MaturityResult{Target: base, RunDuration: 2 * time.Second, CrashFree: true}, 0},
		{"crash", MaturityResult{Target: base, RunDuration: 2 * time.Second, CrashFree: false}, 1},
		{"short run", MaturityResult{Target: base, RunDuration: time.Second, CrashFree: true}, 1},
		{"thin corpus", MaturityResult{Target: FuzzTarget{Name: "X", SeedCorpusSize: 1}, RunDuration: 2 * time.Second, CrashFree: true}, 1},
		{"all three", MaturityResult{Target: FuzzTarget{Name: "X", SeedCorpusSize: 1}, RunDuration: time.Second, CrashFree: false}, 3},
	}
	for _, c := range cases {
		report := Report{Results: []MaturityResult{c.result}}
		got := report.Violations(2*time.Second, 2)
		if len(got) != c.want {
			t.Errorf("%s: Violations() = %v (len %d), want len %d", c.name, got, len(got), c.want)
		}
	}
}

// copyModuleTree copies moduleRoot into a fresh t.TempDir(), excluding
// .git and .idea, and returns the copy's path. TestNFR_029_
// FuzzHarnessMaturityMeetsCP012Bar runs against this copy rather than the
// real working tree: go test -fuzz persists any newly-interesting input
// it discovers to testdata/fuzz/<Name>/ as a side effect, and this
// package's own smoke run must never leave untracked files in the actual
// repository just from having been executed.
func copyModuleTree(t *testing.T) string {
	t.Helper()
	src, err := filepath.Abs(moduleRoot)
	if err != nil {
		t.Fatalf("filepath.Abs(%s): %v", moduleRoot, err)
	}
	dst := t.TempDir()

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == ".idea") {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copying %s to %s: %v", src, dst, err)
	}
	return dst
}

// TestNFR_029_FuzzHarnessMaturityMeetsCP012Bar is T-0357's named test. It
// runs every currently-registered fuzz target for real (go test -fuzz)
// for SmokeFuzzTime, measures its seed-corpus coverage, aggregates the
// results into a Report, and fails if any target has an unresolved
// crash, ran short of the smoke bar, or ships too thin a seed corpus --
// the literal DoD: "CI fails the report generation if any target has an
// unresolved crash or falls below the documented minimum crash-free
// duration." Skipped under `go test -short` since it shells out to `go
// test` twice per discovered target (once to fuzz, once for the
// seed-corpus coverage pass) plus a full module copy, and is not needed
// for an ordinary fast development loop.
func TestNFR_029_FuzzHarnessMaturityMeetsCP012Bar(t *testing.T) {
	if testing.Short() {
		t.Skip("fuzzmaturity: skipping CP-012 harness-maturity smoke run under -short")
	}

	targets, err := DiscoverFuzzTargets(moduleRoot)
	if err != nil {
		t.Fatalf("DiscoverFuzzTargets(%s): %v", moduleRoot, err)
	}
	if len(targets) == 0 {
		t.Fatalf("DiscoverFuzzTargets(%s): found zero fuzz targets", moduleRoot)
	}

	isolatedRoot := copyModuleTree(t)

	var report Report
	for _, target := range targets {
		result, err := RunSmoke(isolatedRoot, target, SmokeFuzzTime)
		if err != nil {
			t.Fatalf("RunSmoke(%s): %v", target.Name, err)
		}
		if !result.CrashFree {
			t.Logf("%s: smoke run output:\n%s", target.Name, result.Output)
		}

		coverage, err := RunSeedCoverage(isolatedRoot, target)
		if err != nil {
			t.Fatalf("RunSeedCoverage(%s): %v", target.Name, err)
		}
		result.CoveragePercent = coverage

		report.Results = append(report.Results, result)
	}

	t.Logf("CP-012 fuzz harness maturity report:\n%s", report.Render())

	for _, v := range report.Violations(SmokeFuzzTime, MinSeedCorpusSize) {
		t.Errorf("CP-012 harness-maturity bar: %s", v)
	}
}
