package render

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFR_114_SubstitutionNoHostResourceNoReflow is T-0261's named unit test
// (FR-114). The substitution path renders at the RECORDED extent without any
// host-system resource lookup and without reflowing: behaviorally the extent is
// the document's recorded value (host metrics never consulted) and pagination
// is unchanged; statically the substitution-path source makes no
// filesystem/network/host-font call and never invokes the reflow engine.
func TestFR_114_SubstitutionNoHostResourceNoReflow(t *testing.T) {
	report := NewRenderReport()
	recorded := Extent{X: 10, Y: 20, Width: 64, Height: 48}

	ph := Substitute(report, pdUnit(0xAB), recorded, ReasonAbsent)

	// Rendered at exactly the RECORDED extent (not a host resource's metrics).
	if ph.Extent != recorded {
		t.Errorf("substitution extent %+v, want recorded %+v", ph.Extent, recorded)
	}
	// The substitution is recorded and pagination is non-authoritative, but the
	// extent (and thus pagination geometry) is unchanged -- no reflow happened.
	if len(report.Substitutions) != 1 {
		t.Errorf("substitution not recorded")
	}

	// Static guarantee: the substitution-path source touches no host resource
	// (os / net / filesystem / font-directory) and never calls the reflow
	// engine (Reflow / FitAtViewport).
	assertSubstitutionPathClean(t, "substitute.go", "placeholder.go")
}

// assertSubstitutionPathClean parses the given substitution-path files and
// fails if they import a host-access package or call the reflow engine.
func assertSubstitutionPathClean(t *testing.T, files ...string) {
	t.Helper()
	bannedImports := map[string]bool{
		"os": true, "net": true, "net/http": true, "io/ioutil": true,
		"os/exec": true, "path/filepath": true, "syscall": true,
	}
	bannedCalls := map[string]bool{
		"Reflow": true, "FitAtViewport": true, "ReadFile": true,
		"Open": true, "Dial": true, "Get": true, "LookupHost": true,
	}
	fset := token.NewFileSet()
	for _, name := range files {
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if bannedImports[path] {
				t.Errorf("%s imports host-access package %q (FR-114 forbids host lookup)", name, path)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				if bannedCalls[fn.Name] {
					t.Errorf("%s calls %q on the substitution path (FR-114)", name, fn.Name)
				}
			case *ast.SelectorExpr:
				if bannedCalls[fn.Sel.Name] {
					t.Errorf("%s calls %q on the substitution path (FR-114)", name, fn.Sel.Name)
				}
			}
			return true
		})
	}
}
