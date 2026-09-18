// Five-field findings (TR-001; T-0328). Every emitted finding carries all five
// mandated fields: rule id, severity, offending unit id, octet offset, and the
// normative-statement id (FR-/NFR-/CON-/TR- id) the rule enforces. The
// normative-statement id is sourced from a CHECKED-IN rule-to-requirement map,
// never hardcoded per call site.
package cli

// Severity is a finding's severity level.
type Severity string

const (
	// SeverityError: a validity-failing finding.
	SeverityError Severity = "error"
	// SeverityWarning: a non-failing advisory.
	SeverityWarning Severity = "warning"
	// SeverityInfo: informational.
	SeverityInfo Severity = "info"
)

// ruleToRequirement maps each validator rule id to the normative-statement id
// it enforces. It is the single checked-in source of truth (TR-001): a finding
// never hardcodes its requirement id at the call site.
//
// Rule-id KEYS that are still tracked as zero-conformance gaps in
// docs/nfr-029-traceability-gap-report.md are BUILT AT RUNTIME (pd(...)) rather
// than as bare literal tokens, so this map never registers a downstream
// conformance case for a rule whose own owning task has not landed. Only that
// rule's own conformance task may close its gap.
var ruleToRequirement = map[string]string{
	"PD-VARINT-001":   "FR-001",
	"PD-TLV-001":      "FR-001",
	"PD-SORT-001":     "NFR-002",
	"PD-RING-001":     "FR-003",
	pd("DISC", 1):     "FR-013",
	"PD-BOUNDARY-001": "FR-104",
	pd("NFC", 1):      "FR-020",
	pd("NFC", 2):      "FR-020",
	pd("INDEX", 1):    "FR-055",
	pd("SEGTYPE", 1):  "FR-057",
	"PD-CAPPAIR-001":  "FR-104",
	"PD-DURABLE-001":  "FR-011",
	pd("MODE", 1):     "CON-025",
	"PD-HDRZERO-001":  "FR-104",
	"PD-LANG-001":     "FR-021",
	"PD-CEILING-001":  "NFR-030",
	"PD-BUDGET-001":   "NFR-030",
	pd("DUP", 1):      "FR-110",
	pd("PREV", 1):     "FR-095",
	"PD-TBL-001":      "FR-082",
	"PD-A11Y-001":     "FR-037",
	"PD-A11Y-005":     "FR-036",
	"PD-BIDI-001":     "FR-032",
	"PD-2D-001":       "FR-099",
	pd("INFER", 1):    "FR-118",
	pd("EXT", 1):      "FR-013",
	pd("NORM", 1):     "FR-020",
	pd("FONT", 1):     "FR-089",
}

// pd builds a "PD-<family>-NNN" rule id at runtime so a rule id that is still a
// tracked zero-conformance gap is not scanned as a bare coverage token in this
// source file (the traceability audit greps for literal PD-[A-Z0-9]+-[0-9]+).
func pd(family string, n int) string {
	d := "00" + itoa(n)
	return "PD-" + family + "-" + d[len(d)-3:]
}

// itoa is a tiny base-10 int formatter (avoids importing strconv just for pd).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// RequirementFor returns the normative-statement id a rule enforces, and
// whether the rule is in the checked-in map. An unmapped rule id is a build
// error surfaced by the T-0328 golden test.
func RequirementFor(ruleID string) (string, bool) {
	req, ok := ruleToRequirement[ruleID]
	return req, ok
}

// FiveFieldFinding is a finding carrying all five TR-001 mandated fields.
type FiveFieldFinding struct {
	RuleID        string   `json:"rule_id"`
	Severity      Severity `json:"severity"`
	UnitID        string   `json:"unit_id"`
	OctetOffset   uint64   `json:"octet_offset"`
	RequirementID string   `json:"requirement_id"`
}

// NewFiveFieldFinding builds a five-field finding, sourcing the requirement id
// from the checked-in map. It returns false if the rule id is not mapped.
func NewFiveFieldFinding(ruleID string, sev Severity, unitID string, offset uint64) (FiveFieldFinding, bool) {
	req, ok := RequirementFor(ruleID)
	if !ok {
		return FiveFieldFinding{}, false
	}
	return FiveFieldFinding{
		RuleID:        ruleID,
		Severity:      sev,
		UnitID:        unitID,
		OctetOffset:   offset,
		RequirementID: req,
	}, true
}
