package ceilings

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// dataModelPath locates data-model.md relative to this package directory
// (pkg/ceilings), two levels up from the module root.
const dataModelPath = "../../specs/001-protodoc-format-core/data-model.md"

// specCeilingRow is one parsed markdown row from data-model.md Section 5's
// ceiling table, read fresh from disk at test time.
type specCeilingRow struct {
	name, valueText, requirementIDs string
}

// readSpecCeilingTable reads data-model.md's own "## 5. Ceilings" table
// directly off disk and parses every "| Name | Value | Requirement ID |"
// row after the header/separator rows, stopping at the first line that is
// no longer a table row. This is the live spec text T-0017's DoD requires
// the checked-in Go table be diffed against.
func readSpecCeilingTable(t *testing.T) []specCeilingRow {
	t.Helper()
	f, err := os.Open(dataModelPath)
	if err != nil {
		t.Fatalf("ceilings: opening %s: %v", dataModelPath, err)
	}
	defer f.Close()

	const heading = "## 5. Ceilings"
	const headerRow = "| Ceiling | Value (exact decimal) | Requirement ID |"
	const separatorRow = "|---|---|---|"

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	state := "seek-heading" // -> seek-header-row -> seek-separator -> in-table
	var rows []specCeilingRow
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
				t.Fatalf("ceilings: expected separator row %q right after header row, got %q", separatorRow, line)
			}
		case "in-table":
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "|") {
				// Table ended.
				if err := sc.Err(); err != nil {
					t.Fatalf("ceilings: scanning %s: %v", dataModelPath, err)
				}
				return rows
			}
			cols := strings.Split(trimmed, "|")
			// A well-formed "| a | b | c |" row splits into
			// ["", " a ", " b ", " c ", ""].
			if len(cols) != 5 {
				t.Fatalf("ceilings: malformed ceiling table row %q (%d columns)", line, len(cols))
			}
			rows = append(rows, specCeilingRow{
				name:           strings.TrimSpace(cols[1]),
				valueText:      strings.TrimSpace(cols[2]),
				requirementIDs: strings.TrimSpace(cols[3]),
			})
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("ceilings: scanning %s: %v", dataModelPath, err)
	}
	if state != "in-table" {
		t.Fatalf("ceilings: never found data-model.md's %q table (stopped at state %q)", heading, state)
	}
	return rows
}

// TestCON_009_CeilingTableMatchesSpecText is the CI drift check CON-009
// and T-0017's DoD require: it re-reads data-model.md's own ceiling table
// at test time and fails the moment it disagrees, in either direction,
// with the checked-in Go copy in Table (after setting aside the rows
// excludedFromTable deliberately leaves out, per the package doc comment's
// "Disclosed scope" section). Editing a constant in Table without a
// matching data-model.md edit fails here; editing data-model.md's table
// without updating Table fails here too; and removing an excluded row
// from data-model.md without updating excludedFromTable fails here as
// well.
func TestCON_009_CeilingTableMatchesSpecText(t *testing.T) {
	all := readSpecCeilingTable(t)

	seenExcluded := make(map[string]bool, len(excludedFromTable))
	var spec []specCeilingRow
	for _, row := range all {
		if _, excluded := excludedFromTable[row.name]; excluded {
			seenExcluded[row.name] = true
			continue
		}
		spec = append(spec, row)
	}
	for name := range excludedFromTable {
		if !seenExcluded[name] {
			t.Errorf("excludedFromTable names %q but data-model.md's ceiling table no longer has that row; update excludedFromTable", name)
		}
	}

	if len(spec) != len(Table) {
		t.Fatalf("ceilings: data-model.md's table has %d non-excluded rows, Table has %d; regenerate Table from data-model.md Section 5", len(spec), len(Table))
	}
	for i, want := range spec {
		got := Table[i]
		if got.Name != want.name {
			t.Errorf("row %d: Table.Name = %q, data-model.md says %q", i, got.Name, want.name)
		}
		if got.ValueText != want.valueText {
			t.Errorf("row %d (%s): Table.ValueText = %q, data-model.md says %q", i, want.name, got.ValueText, want.valueText)
		}
		if got.RequirementIDs != want.requirementIDs {
			t.Errorf("row %d (%s): Table.RequirementIDs = %q, data-model.md says %q", i, want.name, got.RequirementIDs, want.requirementIDs)
		}
	}
}

// TestCON_009_EveryCeilingParsesToExactDecimalInteger confirms every row
// (CON-009: "exact decimal integers") parses without error, and spot-checks
// a representative sample of each ParseCeilingValue shape against its
// independently hand-computed expected value.
func TestCON_009_EveryCeilingParsesToExactDecimalInteger(t *testing.T) {
	for _, c := range Table {
		if _, err := c.Value(); err != nil {
			t.Errorf("ceiling %q: ValueText %q: %v", c.Name, c.ValueText, err)
		}
	}

	cases := []struct {
		name       string
		wantMin    uint64
		wantMax    uint64
		wantBigHex string // set only for the two >uint64 rows; hex to keep the literal short
		wantHasBig bool
	}{
		{name: "File prefix total size", wantMax: 1048576},
		{name: "CommitRing region size", wantMax: 3584},     // "3584 (7 x 512)": leading integer, not 7*512 recomputed
		{name: "SegmentTable region size", wantMax: 786432}, // "786432 (16384 x 48)"
		{name: "MAX_FRAMES_PER_SEGMENT", wantMax: 8192},
		{name: "PDL-VARINT 1-octet range", wantMin: 0, wantMax: 252},
		{name: "PDL-VARINT 8-octet-extension range", wantMin: 4294967296, wantMax: 18446744073709551615},
		{name: "T_C max depth", wantMax: 5},                       // "5 (16^5 >= 1048576)": leading integer, not the power
		{name: "T_S depth", wantMax: 4},                           // "4 (16^4 = 65536 capacity)"
		{name: "PLP-1 block size", wantMax: 64},                   // "8 x 8"
		{name: "Run collision bound scale", wantMax: 17179869184}, // "below 2^34 ..." -> 2^34
		{name: "FR-057 combined index+integrity delta per edit", wantMax: 262144},
		{name: "FR-055 dependent reads to locate and read a unit", wantMax: 3},
		{name: "NFR-014 extraction peak memory", wantMax: 33554432},
		{name: "NFR-018 per-page-plus-resources read budget", wantMax: 8388608},
		{name: "Redaction search-cost floor", wantHasBig: true, wantBigHex: "100000000000000000000"}, // 2^80
		{name: "Redaction actual search cost", wantHasBig: true, wantBigHex: strings1e256},           // 2^256
	}
	for _, tc := range cases {
		c, ok := Lookup(tc.name)
		if !ok {
			t.Fatalf("test bug: no such ceiling %q", tc.name)
		}
		v, err := c.Value()
		if err != nil {
			t.Fatalf("ceiling %q: %v", tc.name, err)
		}
		if tc.wantHasBig {
			if v.MaxBig == nil {
				t.Errorf("ceiling %q: expected MaxBig set, got nil (Max=%d)", tc.name, v.Max)
				continue
			}
			if got := v.MaxBig.Text(16); got != tc.wantBigHex {
				t.Errorf("ceiling %q: MaxBig hex = %s, want %s", tc.name, got, tc.wantBigHex)
			}
			continue
		}
		if v.MaxBig != nil {
			t.Errorf("ceiling %q: expected MaxBig nil, got %s", tc.name, v.MaxBig.String())
		}
		if v.Min != tc.wantMin || v.Max != tc.wantMax {
			t.Errorf("ceiling %q: got Min=%d Max=%d, want Min=%d Max=%d", tc.name, v.Min, v.Max, tc.wantMin, tc.wantMax)
		}
	}
}

// strings1e256 is 2^256 in hex, split out as a named constant purely so
// the table above stays legible.
const strings1e256 = "10000000000000000000000000000000000000000000000000000000000000000"

// TestCON_009_StableIDsUniqueAndDerivedFromName confirms every Table
// entry's ID is non-empty, mechanically derived from Name (never
// hand-edited out of sync), and unique.
func TestCON_009_StableIDsUniqueAndDerivedFromName(t *testing.T) {
	seen := make(map[string]string, len(Table))
	for _, c := range Table {
		if c.ID == "" {
			t.Errorf("ceiling %q: empty ID", c.Name)
		}
		if want := slugify(c.Name); c.ID != want {
			t.Errorf("ceiling %q: ID = %q, slugify(Name) = %q", c.Name, c.ID, want)
		}
		if prev, ok := seen[c.ID]; ok {
			t.Errorf("ceiling ID %q used by both %q and %q", c.ID, prev, c.Name)
		}
		seen[c.ID] = c.Name
	}
}
