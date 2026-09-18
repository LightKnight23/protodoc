package canon_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestNFR_002_CanonicalizerHas100PercentCoverage is T-0324's named unit test
// (NFR-002). It enforces the constitution's 100%-coverage-on-canonicalisation
// quality bar: the canonicalizer core files (traverse.go, canonicalize.go) must
// have 100% line coverage. It runs `go test -coverprofile` on the canon package
// (in a child process, guarded against recursion) and parses per-file coverage.
func TestNFR_002_CanonicalizerHas100PercentCoverage(t *testing.T) {
	if os.Getenv("CANON_COVERAGE_CHILD") == "1" {
		t.Skip("child coverage run: skip the meta-gate to avoid recursion")
	}

	dir := t.TempDir()
	prof := filepath.Join(dir, "cover.out")

	// Run the canon package's own tests with coverage, in a child process.
	cmd := exec.Command("go", "test", "-covermode=set", "-coverprofile="+prof, "Protodoc/pkg/canon")
	cmd.Env = append(os.Environ(), "CANON_COVERAGE_CHILD=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child coverage run failed: %v\n%s", err, out)
	}

	raw, err := os.ReadFile(prof)
	if err != nil {
		t.Fatalf("read coverage profile: %v", err)
	}

	// Aggregate covered/total statements for the two core files.
	type acc struct{ covered, total int }
	core := map[string]*acc{"traverse.go": {}, "canonicalize.go": {}}
	for _, line := range strings.Split(string(raw), "\n") {
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		// Format: <file>:<startLine.col>,<endLine.col> <numStmts> <count>
		sp := strings.Fields(line)
		if len(sp) != 3 {
			continue
		}
		fileField := sp[0]
		for name, a := range core {
			if strings.Contains(fileField, "/pkg/canon/"+name+":") {
				stmts, _ := strconv.Atoi(sp[1])
				count, _ := strconv.Atoi(sp[2])
				a.total += stmts
				if count > 0 {
					a.covered += stmts
				}
			}
		}
	}

	for name, a := range core {
		if a.total == 0 {
			t.Errorf("no coverage data found for canon core file %s", name)
			continue
		}
		if a.covered != a.total {
			t.Errorf("%s coverage %d/%d statements < 100%% (NFR-002 canonicalisation quality bar)", name, a.covered, a.total)
		}
	}
}
