// Package fuzzmaturity is T-0357's CP-012 continuous-fuzzing
// harness-maturity aggregator: it discovers every native Go fuzz target
// (func FuzzXxx(f *testing.F)) checked in under pkg/, runs each for a
// short CI smoke duration under go test's own fuzzing engine, measures
// its seed-corpus coverage, and checks the aggregate against a documented
// minimum bar before M19's v1-stable capstone decision (T-0358) may cite
// this gate as passing.
//
// # Disclosed scope
//
// CP-012's continuous-fuzzing concern spans M01/M02/M03/M07/M09/M10/M14's
// decode paths over untrusted bytes. As of this task, only M01 has landed
// fuzz targets (pkg/pdlfmt/fuzz_test.go, pkg/container/fuzz_test.go).
// DiscoverFuzzTargets is directory-driven, not a hardcoded milestone
// list, so a later milestone's fuzz_test.go is picked up automatically
// the moment it lands, with no change needed here.
//
// SmokeFuzzTime is deliberately short: a per-commit CI smoke check, not
// the multi-day continuous campaign CP-012's own text separately
// requires ("a named triage owner and a published maximum of 90 days
// from report to fix or advisory ... before enrolment in any public
// fuzzing service"). That disclosure-clock governance precondition is
// T-0372's concern, not this package's; this package only proves that,
// right now, every registered target survives a short mutation run
// without crashing and ships a real seed corpus.
package fuzzmaturity

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SmokeFuzzTime is the CI smoke-check duration every discovered fuzz
// target must survive crash-free. See package doc for why this is short
// rather than CP-012's full continuous-campaign duration.
const SmokeFuzzTime = 2 * time.Second

// MinSeedCorpusSize is the minimum number of seed corpus entries
// (f.Add(...) calls) CP-012 requires of every fuzz target: at least one
// well-formed input and at least one already-malformed rejection vector,
// matching the convention every existing M01 target's own doc comment
// already documents (e.g. pkg/container/fuzz_test.go's
// FuzzCP_012_DecodeUntrustedBytes_Header).
const MinSeedCorpusSize = 2

// FuzzTarget is one discovered `func FuzzXxx(f *testing.F)` entry point.
type FuzzTarget struct {
	Name           string // e.g. "FuzzCP_012_DecodeUntrustedBytes_Header"
	PkgDir         string // package directory relative to moduleRoot, e.g. "pkg/container"
	File           string // source file relative to moduleRoot
	SeedCorpusSize int    // number of f.Add(...) seed corpus entries in the target's source
}

// DiscoverFuzzTargets walks every *_test.go file under moduleRoot/pkg and
// returns every native Go fuzz target it finds, sorted by Name for a
// deterministic report.
func DiscoverFuzzTargets(moduleRoot string) ([]FuzzTarget, error) {
	pkgRoot := filepath.Join(moduleRoot, "pkg")
	fset := token.NewFileSet()
	var targets []FuzzTarget

	err := filepath.WalkDir(pkgRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return fmt.Errorf("fuzzmaturity: parsing %s: %w", path, perr)
		}
		relDir, rerr := filepath.Rel(moduleRoot, filepath.Dir(path))
		if rerr != nil {
			return fmt.Errorf("fuzzmaturity: relativising %s: %w", path, rerr)
		}
		relFile, rerr := filepath.Rel(moduleRoot, path)
		if rerr != nil {
			return fmt.Errorf("fuzzmaturity: relativising %s: %w", path, rerr)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Fuzz") {
				continue
			}
			paramName, ok := fuzzParamName(fn)
			if !ok {
				continue
			}
			targets = append(targets, FuzzTarget{
				Name:           fn.Name.Name,
				PkgDir:         filepath.ToSlash(relDir),
				File:           filepath.ToSlash(relFile),
				SeedCorpusSize: countSeedAdds(fn, paramName),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets, nil
}

// fuzzParamName reports whether fn has Go native fuzzing's required
// shape (exactly one parameter, of type *testing.F) and, if so, returns
// that parameter's name.
func fuzzParamName(fn *ast.FuncDecl) (string, bool) {
	params := fn.Type.Params.List
	if len(params) != 1 || len(params[0].Names) != 1 {
		return "", false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "testing" || sel.Sel.Name != "F" {
		return "", false
	}
	return params[0].Names[0].Name, true
}

// countSeedAdds counts paramName.Add(...) calls directly inside fn's
// body. It never descends into paramName.Fuzz(...)'s own registered
// function literal, since a call shaped like an Add inside that body
// would be on a different value (the fuzz function's own parameters),
// not the harness's seed corpus.
func countSeedAdds(fn *ast.FuncDecl, paramName string) int {
	count := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != paramName {
			return true
		}
		switch sel.Sel.Name {
		case "Add":
			count++
		case "Fuzz":
			return false
		}
		return true
	})
	return count
}

// MaturityResult is one fuzz target's aggregated CP-012 harness-maturity
// measurement.
type MaturityResult struct {
	Target          FuzzTarget
	RunDuration     time.Duration
	CrashFree       bool
	CoveragePercent float64
	Output          string // captured `go test` output; populated only when CrashFree is false
}

// RunSmoke runs target under go test's native fuzzing engine
// (go test -fuzz) for fuzzTime, rooted at moduleRoot, and reports whether
// it stayed crash-free and how long it actually ran. A non-nil error
// means the invocation itself could not be attempted (e.g. the go tool is
// missing); a crash the fuzzer finds is an expected, handled report
// outcome (CrashFree=false, error nil), not a harness failure.
func RunSmoke(moduleRoot string, target FuzzTarget, fuzzTime time.Duration) (MaturityResult, error) {
	start := time.Now()
	cmd := exec.Command("go", "test",
		"-run=^$",
		"-fuzz=^"+target.Name+"$",
		"-fuzztime="+fuzzTime.String(),
		"./"+target.PkgDir,
	)
	cmd.Dir = moduleRoot
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	result := MaturityResult{Target: target, RunDuration: elapsed}

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		result.CrashFree = true
	case errors.As(err, &exitErr):
		result.CrashFree = false
		result.Output = string(out)
	default:
		return MaturityResult{}, fmt.Errorf("fuzzmaturity: running go test -fuzz for %s: %w (output: %s)", target.Name, err, out)
	}
	return result, nil
}

// RunSeedCoverage runs target's seed corpus only (go test -run=^Name$,
// i.e. every f.Add entry executed once as an ordinary subtest, no
// mutation) under go test -coverprofile and returns the resulting
// statement-coverage percentage for target's own package. This measures
// seed-corpus coverage, not full mutation-fuzzing coverage: the
// continuous campaign's own coverage is a property of its accumulating
// corpus over time, not a single CI invocation, and is out of this
// smoke check's scope (see package doc).
func RunSeedCoverage(moduleRoot string, target FuzzTarget) (float64, error) {
	tmp, err := os.CreateTemp("", "fuzzmaturity-cover-*.out")
	if err != nil {
		return 0, fmt.Errorf("fuzzmaturity: creating coverage temp file: %w", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	cmd := exec.Command("go", "test",
		"-run=^"+target.Name+"$",
		"-coverprofile="+tmp.Name(),
		"./"+target.PkgDir,
	)
	cmd.Dir = moduleRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("fuzzmaturity: running go test -coverprofile for %s: %w (output: %s)", target.Name, err, out)
	}

	coverOut, err := exec.Command("go", "tool", "cover", "-func="+tmp.Name()).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("fuzzmaturity: running go tool cover -func for %s: %w (output: %s)", target.Name, err, coverOut)
	}
	return parseTotalCoverage(string(coverOut))
}

// parseTotalCoverage extracts the percentage from go tool cover -func's
// final "total:\t(statements)\tNN.N%" line.
func parseTotalCoverage(coverOutput string) (float64, error) {
	for _, line := range strings.Split(coverOutput, "\n") {
		if !strings.HasPrefix(line, "total:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			break
		}
		last := fields[len(fields)-1]
		pct := strings.TrimSuffix(last, "%")
		v, err := strconv.ParseFloat(pct, 64)
		if err != nil {
			return 0, fmt.Errorf("fuzzmaturity: parsing coverage percentage from %q: %w", last, err)
		}
		return v, nil
	}
	return 0, fmt.Errorf("fuzzmaturity: no 'total:' line in go tool cover output: %q", coverOutput)
}

// Report is one full aggregation run's outcome across every discovered
// fuzz target.
type Report struct {
	Results []MaturityResult
}

// Violations returns one human-readable message per CP-012 bar failure
// across every result in r: an unresolved crash, a smoke run that ended
// short of minDuration, or a seed corpus under minSeeds. A target failing
// more than one check contributes more than one message. Order matches
// r.Results' order for a stable, readable report.
func (r Report) Violations(minDuration time.Duration, minSeeds int) []string {
	var out []string
	for _, res := range r.Results {
		if !res.CrashFree {
			out = append(out, fmt.Sprintf("%s: unresolved crash/failure during smoke run", res.Target.Name))
		}
		if res.RunDuration < minDuration {
			out = append(out, fmt.Sprintf("%s: smoke run duration %s below minimum %s", res.Target.Name, res.RunDuration, minDuration))
		}
		if res.Target.SeedCorpusSize < minSeeds {
			out = append(out, fmt.Sprintf("%s: seed corpus size %d below minimum %d", res.Target.Name, res.Target.SeedCorpusSize, minSeeds))
		}
	}
	return out
}

// Render renders r as a Markdown table suitable for a CI log or a filed
// maturity report.
func (r Report) Render() string {
	var b strings.Builder
	b.WriteString("| Target | Package | Seed corpus | Run duration | Crash-free | Seed coverage |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, res := range r.Results {
		fmt.Fprintf(&b, "| %s | %s | %d | %s | %t | %.1f%% |\n",
			res.Target.Name, res.Target.PkgDir, res.Target.SeedCorpusSize, res.RunDuration, res.CrashFree, res.CoveragePercent)
	}
	return b.String()
}
