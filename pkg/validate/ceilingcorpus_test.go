package validate

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

const ceilingCorpusPath = "testdata/ceilings/ceiling_boundary_corpus.tsv"

// TestFR_106_CON010_CeilingFixtureCorpus is T-0116's named conformance test.
// It iterates the checked-in at-limit/over-limit fixture corpus (one accept
// and one abort fixture per M07-owned ceiling, per CON-010) and asserts
// GuardCount's verdict matches each fixture, and that a second run agrees
// (both "implementations" -- the validator run twice -- produce the same
// pass/abort verdict).
func TestFR_106_CON010_CeilingFixtureCorpus(t *testing.T) {
	run := func() (accepts, aborts int) {
		f, err := os.Open(ceilingCorpusPath)
		if err != nil {
			t.Fatalf("opening ceiling corpus: %v", err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			fields := strings.Split(line, "\t")
			if len(fields) != 3 {
				t.Fatalf("malformed fixture line: %q", line)
			}
			verdict, name := fields[0], fields[1]
			declared, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				t.Fatalf("bad declared value in %q: %v", line, err)
			}
			gErr := GuardCount(name, declared)
			switch verdict {
			case "accept":
				if gErr != nil {
					t.Fatalf("%s @ %d: expected accept, got %v", name, declared, gErr)
				}
				accepts++
			case "abort":
				if !errors.Is(gErr, ErrCeilingExceeded) {
					t.Fatalf("%s @ %d: expected abort, got %v", name, declared, gErr)
				}
				aborts++
			default:
				t.Fatalf("unknown verdict %q", verdict)
			}
		}
		if err := sc.Err(); err != nil {
			t.Fatalf("scanning corpus: %v", err)
		}
		return accepts, aborts
	}

	a1, b1 := run()
	a2, b2 := run() // second run -- must agree (deterministic verdicts)
	if a1 != a2 || b1 != b2 {
		t.Fatalf("two runs disagreed: (%d,%d) vs (%d,%d)", a1, b1, a2, b2)
	}
	// One accept and one abort per M07-owned ceiling (6 ceilings).
	if a1 != 6 || b1 != 6 {
		t.Fatalf("corpus has %d accepts / %d aborts, want 6 each (one at-limit + one over-limit per ceiling)", a1, b1)
	}
}
