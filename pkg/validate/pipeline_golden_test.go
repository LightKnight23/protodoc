package validate

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestM07_PipelineStepOrderingGoldenCorpus is T-0128's named conformance test.
// It ships and runs a golden corpus exercising the full 13-step pipeline
// ordering end-to-end (CP-011: ship the corpus with the spec text for this
// milestone's cross-cutting concern). The corpus has 14 fixtures:
//
//   - 13 "single" fixtures, one per StepID, each specifying that exactly that
//     step fails; the pipeline must report that step's verdict and evaluate no
//     later step (FR-102/106/108/109/110 class failures localise to their own
//     step).
//   - 1 "combined" fixture with EVERY step failing at once; the pipeline must
//     report ONLY the earliest-ordered failure (StepMagicHeader = 1) and never
//     evaluate a later step (FR-103: no partial output, no reconstruction past
//     the first failure).
//
// Each fixture is a committed testdata file naming the failing StepID(s) and
// the expected reported step. The test builds a pipeline where the designated
// steps fail (recording how many steps actually ran) and asserts Run's verdict
// and the halt-early guarantee.
func TestM07_PipelineStepOrderingGoldenCorpus(t *testing.T) {
	dir := filepath.Join("testdata", "pipeline_golden")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read golden corpus dir %s: %v", dir, err)
	}

	allSteps := []StepID{
		StepMagicHeader, StepCapabilityArith, StepRingWinner, StepTruncationRollback,
		StepBoundedPrefix, StepTSRecompute, StepSegTypeCoverage, StepCycleDetection,
		StepStructuralCeilings, StepTCAndSignature, StepNFCAndIdentity,
		StepRegistryExcerpt, StepExtEnvelope,
	}

	fixtureCount := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".fixture") {
			continue
		}
		fixtureCount++
		name := e.Name()
		spec := loadGoldenFixture(t, filepath.Join(dir, name))

		// Track which steps actually run, to prove the pipeline halts at the
		// first validity failure (no later step evaluated).
		ran := map[StepID]bool{}
		var steps []Step
		for _, id := range allSteps {
			id := id
			steps = append(steps, Step{ID: id, Run: func() *Finding {
				ran[id] = true
				if spec.failing[id] {
					f := &Finding{Step: id, RuleID: spec.ruleFor(id), Message: "golden fixture " + name}
					// Step 9 (structural ceilings) is the documented budget
					// exception: it produces a non-validity PD-BUDGET status,
					// reported alongside any validity verdict, and never halts.
					if id == StepStructuralCeilings {
						f.Budget = true
					}
					return f
				}
				return nil
			}})
		}

		res := Run(steps)

		if spec.expectValid {
			if !res.Valid {
				t.Errorf("%s: expected valid, got failure at step %d", name, res.Validity.Step)
			}
			// A step-9-only fixture is valid but carries a PD-BUDGET status.
			if spec.expectBudgetStep != 0 {
				if res.Budget == nil {
					t.Errorf("%s: expected a budget finding at step %d, got none", name, spec.expectBudgetStep)
				} else if res.Budget.Step != spec.expectBudgetStep {
					t.Errorf("%s: budget finding at step %d, expected %d", name, res.Budget.Step, spec.expectBudgetStep)
				}
			}
			continue
		}
		if res.Valid || res.Validity == nil {
			t.Errorf("%s: expected a validity failure at step %d, got valid", name, spec.expectStep)
			continue
		}
		if res.Validity.Step != spec.expectStep {
			t.Errorf("%s: reported step %d, expected earliest failing step %d", name, res.Validity.Step, spec.expectStep)
		}
		// Halt-early: no step ordered AFTER the expected failing step ran.
		for _, id := range allSteps {
			if id > spec.expectStep && ran[id] {
				t.Errorf("%s: step %d ran after the failing step %d (no later step may be evaluated, FR-103)", name, id, spec.expectStep)
			}
		}
	}

	if fixtureCount != 14 {
		t.Fatalf("golden corpus has %d fixtures, expected 14 (13 single-step + 1 combined)", fixtureCount)
	}
}

type goldenFixture struct {
	failing          map[StepID]bool
	expectStep       StepID
	expectValid      bool
	expectBudgetStep StepID
}

func (g goldenFixture) ruleFor(id StepID) string {
	return "PD-STEP-" + strconv.Itoa(int(id))
}

// loadGoldenFixture parses a .fixture file. Format (line-based, # comments):
//
//	fail: <stepid>[,<stepid>...]   # steps that fail (empty allowed)
//	expect_step: <stepid>          # the reported failing step (omit if valid)
//	expect_valid: true             # optional; document is valid
func loadGoldenFixture(t *testing.T, path string) goldenFixture {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture %s: %v", path, err)
	}
	defer f.Close()

	g := goldenFixture{failing: map[StepID]bool{}}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("%s: malformed line %q", path, line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "fail":
			if val == "" {
				continue
			}
			for _, tok := range strings.Split(val, ",") {
				n, err := strconv.Atoi(strings.TrimSpace(tok))
				if err != nil {
					t.Fatalf("%s: bad step id %q: %v", path, tok, err)
				}
				g.failing[StepID(n)] = true
			}
		case "expect_step":
			n, err := strconv.Atoi(val)
			if err != nil {
				t.Fatalf("%s: bad expect_step %q: %v", path, val, err)
			}
			g.expectStep = StepID(n)
		case "expect_valid":
			g.expectValid = val == "true"
		case "expect_budget_step":
			n, err := strconv.Atoi(val)
			if err != nil {
				t.Fatalf("%s: bad expect_budget_step %q: %v", path, val, err)
			}
			g.expectBudgetStep = StepID(n)
		default:
			t.Fatalf("%s: unknown key %q", path, key)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("%s: scan: %v", path, err)
	}
	return g
}
