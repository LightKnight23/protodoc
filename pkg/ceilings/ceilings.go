// Package ceilings is the generated structural ceiling table required by
// CON-009: every structural limit named in data-model.md Section 5 ("the
// complete v1 ceiling table"), checked in here as an exact decimal integer
// (or, for the two entries whose stated value exceeds 2^64-1, an exact
// decimal big.Int) with its own stable identifier.
//
// This file is generated content in the CP-007 sense: Table's Name,
// ValueText and RequirementIDs columns are verbatim copies of
// specs/001-protodoc-format-core/data-model.md Section 5's markdown table,
// row for row, in row order. Regenerate by re-copying that table if it
// changes; TestCON_009_CeilingTableMatchesSpecText (ceilings_test.go) reads
// the live data-model.md file at test time and fails the build the moment
// this copy and the spec text disagree, so a constant can never be edited
// here without a matching spec.md-tree change (and vice versa).
//
// # Disclosed scope: not every row of that table is here
//
// CON-009 requires "every structural limit ... as an exact decimal
// integer ... in a normative table". data-model.md Section 5's own
// introductory sentence claims the same: "every structural limit is an
// exact decimal integer". Both claims are contradicted by that table's own
// content: several of its 62 rows are not expressible as a single exact
// decimal integer without inventing an unstated unit conversion (a
// percentage of file size, a CPU-time or wall-clock budget in seconds,
// milliseconds or working days, a "4x input length" or "2.0x the size of"
// multiplier formula, or qualitative comparative prose with no single
// ceiling value at all), and a few more are explicitly a current
// *estimate* against a ceiling stated elsewhere in the same table, or a
// *minimum* size for the conformance test suite rather than a maximum
// structural ceiling on document content. This is a genuine, disclosed gap
// between CON-009's requirement and data-model.md's present text, not a
// resolution of it: Table below holds exactly the rows that ARE
// expressible as an exact decimal integer (or an explicit integer range,
// or a computed exact power of two) denoting a count, octet size or
// numeric range of document-structural content, in an octet/count unit
// that cannot be mistaken for anything else. excludedFromTable lists every
// row left out, keyed by its exact Name, with the specific reason; the
// row-by-row test still requires that every excluded row's exact spec text
// still exist verbatim in data-model.md, so this scoping decision itself
// stays under drift detection.
package ceilings

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Ceiling is one row of data-model.md's Section 5 ceiling table.
type Ceiling struct {
	// ID is a stable identifier mechanically derived from Name (see
	// slugify): uppercase, non-alphanumeric runs collapsed to a single
	// underscore. Populated by init below, never hand-maintained, so it
	// cannot drift from Name.
	ID string

	// Name is the exact "Ceiling" column text.
	Name string

	// ValueText is the exact "Value (exact decimal)" column text,
	// unparsed. Call Value to parse it.
	ValueText string

	// RequirementIDs is the exact "Requirement ID" column text.
	RequirementIDs string
}

// Value is a parsed ceiling: either a single upper bound (Min==0, Max the
// bound) or an explicit range (Min and Max both set), as an exact decimal
// integer. MaxBig holds the exact decimal value only for the two entries
// (2^80, 2^256) whose value exceeds math.MaxUint64; for every other entry
// MaxBig is nil and Max holds the value.
type Value struct {
	Min    uint64
	Max    uint64
	MaxBig *big.Int
}

// reRangeWhole matches an explicit "<min> to <max>" range occupying the
// entire value cell (the four PDL-VARINT rows).
var reRangeWhole = regexp.MustCompile(`^(\d+) to (\d+)$`)

// reTimesWhole matches an exact "<a> x <b>" occupying the entire value
// cell (the PLP-1 block size row); a leading integer followed by a
// parenthetical "(<a> x <b>)" annotation (e.g. "786432 (16384 x 48)") does
// NOT match this and falls through to the leading-integer rule instead,
// which is what we want: the leading decimal is already the stated
// ceiling there, the parenthetical is only the arithmetic that produced
// it.
var reTimesWhole = regexp.MustCompile(`^(\d+) x (\d+)$`)

// reLeadingInt matches a leading decimal integer, however it is followed.
var reLeadingInt = regexp.MustCompile(`^(\d+)`)

// rePow matches the first "<base>^<exp>" occurring anywhere in the cell.
// Used only once reLeadingInt has told us the leading token is not itself
// the standalone ceiling (see ParseCeilingValue).
var rePow = regexp.MustCompile(`(\d+)\^(\d+)`)

// ParseCeilingValue derives the exact decimal integer(s) a data-model.md
// Section 5 value cell states, by the following precedence, matched
// against the rows actually present in Table (see ceilings_test.go for a
// row-by-row worked check):
//
//  1. An exact "<min> to <max>" cell (the PDL-VARINT ranges) yields that
//     Min/Max pair.
//  2. An exact "<a> x <b>" cell (PLP-1 block size, "8 x 8") yields
//     Max = a*b.
//  3. A cell beginning with a decimal integer NOT immediately followed by
//     "^" (every plain-value row, and every row whose real ceiling is a
//     leading integer followed by a "(...)" derivation annotation, e.g.
//     "5 (16^5 >= 1048576)") yields that leading integer as Max.
//  4. Otherwise, the first "<base>^<exp>" appearing anywhere in the cell
//     (a cell that is itself a bare power, "2^80"/"2^256", or one
//     introduced by qualifying prose, "below 2^34 ...") is computed
//     exactly via math/big and yields that as Max (or MaxBig if it
//     exceeds math.MaxUint64).
//
// One row (PDL-TLV field tag width, "1 octet, max 256 fields per record
// type") packs two distinct quantities into one cell; rule 3 yields only
// the leading one (1, the tag width in octets). This is a disclosed
// limitation of this row's own text, not a parsing defect: see that
// entry's comment in Table.
func ParseCeilingValue(s string) (Value, error) {
	s = strings.TrimSpace(s)

	if m := reRangeWhole.FindStringSubmatch(s); m != nil {
		min, err := strconv.ParseUint(m[1], 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("ceilings: range min %q: %w", m[1], err)
		}
		max, err := strconv.ParseUint(m[2], 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("ceilings: range max %q: %w", m[2], err)
		}
		return Value{Min: min, Max: max}, nil
	}

	if m := reTimesWhole.FindStringSubmatch(s); m != nil {
		a, err := strconv.ParseUint(m[1], 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("ceilings: product term %q: %w", m[1], err)
		}
		b, err := strconv.ParseUint(m[2], 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("ceilings: product term %q: %w", m[2], err)
		}
		return Value{Max: a * b}, nil
	}

	if loc := reLeadingInt.FindStringIndex(s); loc != nil {
		end := loc[1]
		if end >= len(s) || s[end] != '^' {
			max, err := strconv.ParseUint(s[loc[0]:end], 10, 64)
			if err != nil {
				return Value{}, fmt.Errorf("ceilings: leading integer %q: %w", s[loc[0]:end], err)
			}
			return Value{Max: max}, nil
		}
	}

	if m := rePow.FindStringSubmatch(s); m != nil {
		base, ok := new(big.Int).SetString(m[1], 10)
		if !ok {
			return Value{}, fmt.Errorf("ceilings: power base %q unparseable", m[1])
		}
		exp, ok := new(big.Int).SetString(m[2], 10)
		if !ok {
			return Value{}, fmt.Errorf("ceilings: power exponent %q unparseable", m[2])
		}
		result := new(big.Int).Exp(base, exp, nil)
		if result.IsUint64() {
			return Value{Max: result.Uint64()}, nil
		}
		return Value{MaxBig: result}, nil
	}

	return Value{}, fmt.Errorf("ceilings: value cell %q matches no known ceiling shape", s)
}

// Value parses c.ValueText via ParseCeilingValue.
func (c Ceiling) Value() (Value, error) {
	return ParseCeilingValue(c.ValueText)
}

// slugify mechanically derives a stable ID from a Name: uppercase,
// non-alphanumeric runs collapsed to one underscore, leading/trailing
// underscores trimmed. Two distinct Table rows never collapse to the same
// ID (TestCON_009_CeilingTableMatchesSpecText asserts this), so ID is
// stable and unique without being hand-maintained.
func slugify(name string) string {
	var b strings.Builder
	prevUnderscore := true // suppresses a leading underscore
	for _, r := range strings.ToUpper(name) {
		switch {
		case r >= 'A' && r <= 'Z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevUnderscore = false
		case !prevUnderscore:
			b.WriteByte('_')
			prevUnderscore = true
		}
	}
	return strings.TrimRight(b.String(), "_")
}

// Table is the complete v1 ceiling table (data-model.md Section 5, 44
// rows), row for row, in the source table's order. Name, ValueText and
// RequirementIDs are verbatim copies of that table's three columns.
var Table = []Ceiling{
	{Name: "File prefix total size", ValueText: "1048576", RequirementIDs: "TR-006, TR-007, TR-008"},
	{Name: "Header region size", ValueText: "512", RequirementIDs: "FR-006, FR-007, FR-008, FR-009, FR-011"},
	{Name: "Header frozen-forever window", ValueText: "32 (octets `[0,32)`)", RequirementIDs: "CP-008, FR-123"},
	{Name: "CommitRing slot count", ValueText: "7", RequirementIDs: "HC-027, FR-117"},
	{Name: "CommitRing slot size", ValueText: "512", RequirementIDs: "HC-027"},
	{Name: "CommitRing region size", ValueText: "3584 (7 x 512)", RequirementIDs: "HC-027"},
	{Name: "Frontmatter region size", ValueText: "258048", RequirementIDs: "FR-051, FR-053, FR-054"},
	{Name: "Frontmatter preview_raster max", ValueText: "131072", RequirementIDs: "FR-051"},
	{Name: "Frontmatter preview_source_snapshot max", ValueText: "4096", RequirementIDs: "FR-053"},
	{Name: "Frontmatter document_metadata max", ValueText: "16384", RequirementIDs: "FR-054"},
	{Name: "Frontmatter retired_token_list max", ValueText: "16384", RequirementIDs: "DP-011"},
	{Name: "SegmentTable slot count (MAX_SEGMENTS)", ValueText: "16384", RequirementIDs: "TR-006, NFR-004"},
	{Name: "SegmentTable slot size", ValueText: "48", RequirementIDs: "TR-006"},
	{Name: "SegmentTable region size", ValueText: "786432 (16384 x 48)", RequirementIDs: "TR-006"},
	{Name: "MAX_FRAMES_PER_SEGMENT", ValueText: "8192", RequirementIDs: "FR-102"},
	{Name: "MAX_DECODED_UNIT", ValueText: "268435456", RequirementIDs: "NFR-030"},
	{Name: "MAX_TEXT_UNIT_OCTETS", ValueText: "65536", RequirementIDs: "DP-009"},
	{Name: "PDL-VARINT 1-octet range", ValueText: "0 to 252", RequirementIDs: "DP-002"},
	{Name: "PDL-VARINT 2-octet-extension range", ValueText: "253 to 65535", RequirementIDs: "DP-002"},
	{Name: "PDL-VARINT 4-octet-extension range", ValueText: "65536 to 4294967295", RequirementIDs: "DP-002"},
	{Name: "PDL-VARINT 8-octet-extension range", ValueText: "4294967296 to 18446744073709551615", RequirementIDs: "DP-002"},
	// This cell packs two distinct quantities (tag width in octets, and
	// max field count); ParseCeilingValue's rule 3 yields only the
	// leading one (1, the tag width). See ParseCeilingValue's doc comment.
	{Name: "PDL-TLV field tag width", ValueText: "1 octet, max 256 fields per record type", RequirementIDs: "DP-002"},
	// "below 2^34" is stated as an exclusive probabilistic bound, not a
	// hard structural ceiling; Value().Max is the literal 2^34 named in
	// the cell, not 2^34-1.
	{Name: "Run collision bound scale", ValueText: "below 2^34 mints per lineage at P < 2^-60", RequirementIDs: "FR-023"},
	{Name: "MAX_REFERENCES (annotations/font records)", ValueText: "4194304", RequirementIDs: "FR-028, FR-089"},
	{Name: "Annotation boundary_behaviour values", ValueText: "4 (exactly)", RequirementIDs: "FR-026, FR-027"},
	{Name: "T_C / T_S tree arity", ValueText: "16 (always, absent children filled)", RequirementIDs: "DP-006"},
	{Name: "T_C max depth", ValueText: "5 (16^5 >= 1048576)", RequirementIDs: "MAX_CONTENT_UNITS"},
	{Name: "T_S depth", ValueText: "4 (16^4 = 65536 capacity)", RequirementIDs: "headroom over MAX_SEGMENTS"},
	{Name: "MAX_CONTENT_UNITS", ValueText: "1048576", RequirementIDs: "T_C sizing"},
	{Name: "Redaction search-cost floor", ValueText: "2^80", RequirementIDs: "FR-075"},
	{Name: "Redaction actual search cost", ValueText: "2^256", RequirementIDs: "FR-075 (margin)"},
	{Name: "MAX_SIGNATURES", ValueText: "64", RequirementIDs: "FR-063"},
	{Name: "de Casteljau flattening tolerance", ValueText: "762 base units (914400/300/4)", RequirementIDs: "CON-012, HC-020"},
	{Name: "de Casteljau max recursion depth", ValueText: "16", RequirementIDs: "HC-020"},
	{Name: "PLP-1 block size", ValueText: "8 x 8", RequirementIDs: "CQ-010"},
	{Name: "PLP-1 fixed quantization matrix count", ValueText: "8", RequirementIDs: "CQ-010"},
	{Name: "Font recorded values", ValueText: "7 (exactly)", RequirementIDs: "FR-089"},
	{Name: "MAX_PAGES", ValueText: "131072", RequirementIDs: "PageDirectory"},
	{Name: "MAX_REGISTRY_EXCERPT_OCTETS", ValueText: "1048576", RequirementIDs: "DP-017"},
	{Name: "Extraction role statement ceiling", ValueText: "150", RequirementIDs: "CP-002 (current estimate ~65)"},
	{Name: "Validating-and-verifying role statement ceiling (cumulative)", ValueText: "400", RequirementIDs: "CP-002 (current estimate ~220)"},
	{Name: "Rendering role statement ceiling (cumulative)", ValueText: "500", RequirementIDs: "CP-002 (current estimate ~321)"},
	{Name: "Extraction reference-implementation line budget", ValueText: "1000 (source lines, no font/graphics dependency)", RequirementIDs: "CP-010"},
	{Name: "FR-057 combined index+integrity delta per edit", ValueText: "262144 octets", RequirementIDs: "FR-057"},
	{Name: "FR-055 dependent reads to locate and read a unit", ValueText: "3 (inclusive of the unit read itself)", RequirementIDs: "FR-055"},
	{Name: "NFR-014 extraction peak memory", ValueText: "33554432 octets", RequirementIDs: "NFR-014"},
	{Name: "NFR-016 preview peak memory", ValueText: "67108864 octets", RequirementIDs: "NFR-016"},
	{Name: "NFR-017 open-and-render peak memory", ValueText: "209715200 octets", RequirementIDs: "NFR-017"},
	{Name: "NFR-018 per-page-plus-resources read budget", ValueText: "8388608 octets", RequirementIDs: "NFR-018"},
}

// excludedFromTable lists every row of data-model.md Section 5's ceiling
// table deliberately left out of Table, keyed by its exact Name, with the
// specific reason (see the package doc comment's "Disclosed scope"
// section). TestCON_009_CeilingTableMatchesSpecText requires every key
// here still be present verbatim in data-model.md's live table, so an
// exclusion cannot silently go stale either.
var excludedFromTable = map[string]string{
	"PLP-1 decode line-count estimate":                     "explicitly a design estimate (CP-003 feasibility), not an enforced ceiling",
	"Extraction reference-implementation current estimate": "explicitly a current measurement against the line budget stated in the row above it, not itself a ceiling",
	"NFR-008 write amplification per K-octet edit":         "value cell is a formula with a K-multiplier and a multi-condition test matrix (1 MB / 50 MB / 500 MB), not a single exact decimal integer",
	"NFR-012 bounded read fraction":                        "value is a percentage (ratio of file size), not a count/size ceiling",
	"NFR-013 extraction CPU budget":                        "value is a CPU-time budget in seconds, benchmarked directly rather than a decode-time structural ceiling",
	"NFR-015 cold-cache preview budget":                    "value is a wall-clock budget in milliseconds, benchmarked directly by T-0015 rather than a decode-time structural ceiling",
	"CP-012 fuzzing memory-oracle multiplier":              "value is a multiplier formula (\"4x input octet length ..., floor-exempted below 1048577 octets\"), not a fixed ceiling",
	"CP-012 disclosure clock":                              "value is a time budget in days, a governance process property, not a document-structural ceiling",
	"NFR-026 extracting-and-validating implementer budget": "value is a time budget in working days, an implementation-effort estimate, not a document-structural ceiling",
	"NFR-027 rendering implementer budget":                 "value is a time budget in working days, an implementation-effort estimate, not a document-structural ceiling",
	"Negative corpus minimum":                              "states a MINIMUM size for the conformance test suite, not a maximum structural ceiling on document content",
	"NFR-032 complete-history size overhead":               "value is a multiplier (\"2.0x\"), not a fixed decimal integer",
	"NFR-033 open/render time and memory bound":            "value is qualitative comparative prose (\"proportional to ... within 1.5x ... at 10x ...\"), not a single ceiling integer",
}

func init() {
	for i := range Table {
		Table[i].ID = slugify(Table[i].Name)
	}
}

// Lookup returns the Table entry with the given exact Name, and whether
// it was found.
func Lookup(name string) (Ceiling, bool) {
	for _, c := range Table {
		if c.Name == name {
			return c, true
		}
	}
	return Ceiling{}, false
}

// MustMax returns the Max of the Table entry with the given exact Name,
// parsed via Value. It panics if no such entry exists or its value cell
// fails to parse: both are build-time-detectable authoring errors in this
// checked-in file, never a runtime condition a caller can hit from
// document input.
func MustMax(name string) uint64 {
	c, ok := Lookup(name)
	if !ok {
		panic(fmt.Sprintf("ceilings: no table entry named %q", name))
	}
	v, err := c.Value()
	if err != nil {
		panic(fmt.Sprintf("ceilings: entry %q: %v", name, err))
	}
	return v.Max
}
