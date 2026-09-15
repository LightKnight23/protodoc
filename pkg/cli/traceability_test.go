package cli

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"
)

// cliContractPath locates contracts/cli.md relative to this package
// directory (pkg/cli), two levels up from the module root.
const cliContractPath = "../../specs/001-protodoc-format-core/contracts/cli.md"

// readSpecTraceabilityTable reads contracts/cli.md's own "## 14.
// Requirement-to-verb traceability table" table directly off disk and
// parses every "| Requirement ID | Verb(s) |" row after the header/
// separator rows, stopping at the first line that is no longer a table
// row. This is the live spec text TestTR_012_TraceabilityTableMatchesDispatch
// diffs the checked-in Go copy in Table against.
func readSpecTraceabilityTable(t *testing.T) []TraceEntry {
	t.Helper()
	f, err := os.Open(cliContractPath)
	if err != nil {
		t.Fatalf("traceability: opening %s: %v", cliContractPath, err)
	}
	defer f.Close()

	const heading = "## 14. Requirement-to-verb traceability table"
	const headerRow = "| Requirement ID | Verb(s) |"
	const separatorRow = "|---|---|"

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	state := "seek-heading" // -> seek-header-row -> seek-separator -> in-table
	var rows []TraceEntry
	for sc.Scan() {
		line := sc.Text()
		switch state {
		case "seek-heading":
			if strings.TrimSpace(line) == heading {
				state = "seek-header-row"
			}
		case "seek-header-row":
			if strings.TrimSpace(line) == headerRow {
				state = "seek-separator"
			}
		case "seek-separator":
			if strings.TrimSpace(line) == separatorRow {
				state = "in-table"
			} else if strings.TrimSpace(line) != "" {
				t.Fatalf("traceability: expected separator row %q right after header row, got %q", separatorRow, line)
			}
		case "in-table":
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "|") {
				if err := sc.Err(); err != nil {
					t.Fatalf("traceability: scanning %s: %v", cliContractPath, err)
				}
				return rows
			}
			cols := strings.Split(trimmed, "|")
			// A well-formed "| a | b |" row splits into ["", " a ", " b ", ""].
			if len(cols) != 4 {
				t.Fatalf("traceability: malformed table row %q (%d columns)", line, len(cols))
			}
			rows = append(rows, TraceEntry{
				RequirementID: strings.TrimSpace(cols[1]),
				Verbs:         strings.TrimSpace(cols[2]),
			})
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("traceability: scanning %s: %v", cliContractPath, err)
	}
	if state != "in-table" {
		t.Fatalf("traceability: never found cli.md's %q table (stopped at state %q)", heading, state)
	}
	return rows
}

// TestTR_012_TraceabilityTableMatchesDispatch is T-0339's named test. It
// fails if the checked-in Go Table drifts from cli.md's live Section 14
// text in either direction, if any table row names a verb absent from the
// T-0325 dispatch registry (pkg/cli.Names()), or if any of the 11
// registered verbs is absent from every row's Verb(s) column.
func TestTR_012_TraceabilityTableMatchesDispatch(t *testing.T) {
	spec := readSpecTraceabilityTable(t)

	if len(spec) != len(Table) {
		t.Fatalf("traceability: cli.md's table has %d rows, Table has %d; regenerate Table from cli.md Section 14", len(spec), len(Table))
	}
	for i, want := range spec {
		got := Table[i]
		if got.RequirementID != want.RequirementID {
			t.Errorf("row %d: Table.RequirementID = %q, cli.md says %q", i, got.RequirementID, want.RequirementID)
		}
		if got.Verbs != want.Verbs {
			t.Errorf("row %d (%s): Table.Verbs = %q, cli.md says %q", i, want.RequirementID, got.Verbs, want.Verbs)
		}
	}
	if t.Failed() {
		return
	}

	dispatched := make(map[string]bool, len(Names()))
	for _, name := range Names() {
		dispatched[name] = true
	}

	seenVerb := make(map[string]bool, len(dispatched))
	seenRequirement := make(map[string]bool, len(Table))
	for _, row := range Table {
		if seenRequirement[row.RequirementID] {
			t.Errorf("requirement %s appears in more than one Table row", row.RequirementID)
		}
		seenRequirement[row.RequirementID] = true

		for _, v := range strings.Split(row.Verbs, ",") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if !dispatched[v] {
				t.Errorf("requirement %s names verb %q, which is not in the dispatch registry (Names(): %v)", row.RequirementID, v, Names())
				continue
			}
			seenVerb[v] = true
		}
	}

	var missing []string
	for _, name := range Names() {
		if !seenVerb[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("verb(s) registered in dispatch but absent from every traceability-table row: %v", missing)
	}
}
