package benchconfig

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot is this package's directory, two levels up (pkg/benchconfig ->
// module root), matching the relative-path convention this module's other
// drift-check tests already use (see pkg/ceilings/ceilings_test.go's
// dataModelPath).
const moduleRoot = "../.."

// isTestingBFunc reports whether fn's sole parameter is exactly *testing.B,
// i.e. fn has the signature every func Benchmark* must have to be run by
// `go test -bench`.
func isTestingBFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	star, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	return ok && pkgIdent.Name == "testing" && sel.Sel.Name == "B"
}

// findBenchmarksMissingCitation parses every *_test.go file under root and
// returns "path:FuncName" for every func Benchmark*(b *testing.B) whose
// body does not call benchconfig.Stamp. Used both against this module's
// real source tree (TestNFR_011_BenchmarksCiteReferenceConfig) and against
// a synthetic fixture that proves the check actually fails closed
// (TestNFR_011_BenchmarksCiteReferenceConfig_DetectsSyntheticViolation).
func findBenchmarksMissingCitation(root string) ([]string, error) {
	var missing []string
	fset := token.NewFileSet()

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Recv != nil {
				continue
			}
			if !strings.HasPrefix(fn.Name.Name, "Benchmark") || !isTestingBFunc(fn) {
				continue
			}
			start := fset.Position(fn.Body.Pos()).Offset
			end := fset.Position(fn.Body.End()).Offset
			body := string(src[start:end])
			if !strings.Contains(body, "benchconfig.Stamp(") {
				missing = append(missing, path+":"+fn.Name.Name)
			}
		}
		return nil
	})
	return missing, walkErr
}

// TestNFR_011_BenchmarksCiteReferenceConfig is T-0343's named test (NFR-011
// Audit A-REFPLAT, enforced at build time rather than at benchmark run
// time: a benchmark job that never runs in a given CI invocation, e.g.
// because -bench was not passed, must still fail the build if its source
// omits the citation). It walks this entire module's *_test.go files and
// fails if any func Benchmark*(b *testing.B) does not call
// benchconfig.Stamp somewhere in its body.
func TestNFR_011_BenchmarksCiteReferenceConfig(t *testing.T) {
	missing, err := findBenchmarksMissingCitation(moduleRoot)
	if err != nil {
		t.Fatalf("benchconfig: walking %s for benchmark funcs: %v", moduleRoot, err)
	}
	for _, m := range missing {
		t.Errorf("benchconfig: %s does not call benchconfig.Stamp; every NFR-012..019 benchmark job must cite %s (NFR-011 Audit A-REFPLAT)", m, ReferenceConfigID)
	}
}

// TestNFR_011_BenchmarksCiteReferenceConfig_DetectsSyntheticViolation proves
// the audit above actually fails closed: a synthetic benchmark func with no
// Stamp call, written to a scratch directory (never the real source tree),
// must be reported missing, and a sibling compliant one must not be.
func TestNFR_011_BenchmarksCiteReferenceConfig_DetectsSyntheticViolation(t *testing.T) {
	dir := t.TempDir()
	const src = `package fixture

import "testing"

func BenchmarkNoncompliant(b *testing.B) {
	b.Log("measured with no reference-config citation")
}

func BenchmarkCompliant(b *testing.B) {
	benchconfig.Stamp(b, "NFR-999")
}
`
	if err := os.WriteFile(filepath.Join(dir, "synthetic_bench_test.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing synthetic fixture: %v", err)
	}

	missing, err := findBenchmarksMissingCitation(dir)
	if err != nil {
		t.Fatalf("findBenchmarksMissingCitation(%s): %v", dir, err)
	}
	if len(missing) != 1 || !strings.HasSuffix(missing[0], ":BenchmarkNoncompliant") {
		t.Fatalf("findBenchmarksMissingCitation(%s) = %v, want exactly one entry naming BenchmarkNoncompliant", dir, missing)
	}
}
