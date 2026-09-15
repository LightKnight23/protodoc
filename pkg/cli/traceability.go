package cli

// TraceEntry is one row of contracts/cli.md Section 14's requirement-to-verb
// traceability table: a single FR/NFR/CON/TR requirement identifier and the
// comma-separated verb name(s) that satisfy it, exactly as that column
// reads in the markdown table.
type TraceEntry struct {
	// RequirementID is the exact "Requirement ID" column text.
	RequirementID string

	// Verbs is the exact "Verb(s)" column text (e.g. "sign, verify").
	Verbs string
}

// Table is a verbatim copy of contracts/cli.md Section 14's table, row for
// row, in that table's own order. Regenerate by re-copying that table if it
// changes; TestTR_012_TraceabilityTableMatchesDispatch (traceability_test.go)
// reads the live cli.md file at test time and fails the build the moment
// this copy and the spec text disagree.
var Table = []TraceEntry{
	{RequirementID: "FR-024", Verbs: "merge"},
	{RequirementID: "FR-035", Verbs: "extract"},
	{RequirementID: "FR-041", Verbs: "extract"},
	{RequirementID: "FR-042", Verbs: "extract"},
	{RequirementID: "FR-047", Verbs: "extract"},
	{RequirementID: "FR-048", Verbs: "extract"},
	{RequirementID: "FR-062", Verbs: "verify"},
	{RequirementID: "FR-063", Verbs: "sign, verify"},
	{RequirementID: "FR-064", Verbs: "sign, verify"},
	{RequirementID: "FR-065", Verbs: "sign, verify"},
	{RequirementID: "FR-066", Verbs: "sign, verify"},
	{RequirementID: "FR-067", Verbs: "sign, verify"},
	{RequirementID: "FR-068", Verbs: "diff, sign, verify"},
	{RequirementID: "FR-069", Verbs: "sign, verify"},
	{RequirementID: "FR-070", Verbs: "sign, verify"},
	{RequirementID: "FR-071", Verbs: "sign, verify"},
	{RequirementID: "FR-072", Verbs: "verify"},
	{RequirementID: "FR-073", Verbs: "verify"},
	{RequirementID: "FR-074", Verbs: "redact, verify"},
	{RequirementID: "FR-075", Verbs: "redact, verify"},
	{RequirementID: "FR-076", Verbs: "redact, verify"},
	{RequirementID: "FR-077", Verbs: "redact, verify"},
	{RequirementID: "FR-078", Verbs: "redact, verify"},
	{RequirementID: "FR-080", Verbs: "redact"},
	{RequirementID: "FR-081", Verbs: "redact"},
	{RequirementID: "FR-089", Verbs: "sign"},
	{RequirementID: "FR-092", Verbs: "merge"},
	{RequirementID: "FR-093", Verbs: "merge"},
	{RequirementID: "FR-094", Verbs: "merge"},
	{RequirementID: "FR-095", Verbs: "merge"},
	{RequirementID: "FR-096", Verbs: "merge"},
	{RequirementID: "FR-102", Verbs: "validate"},
	{RequirementID: "FR-103", Verbs: "validate"},
	{RequirementID: "FR-104", Verbs: "validate, verify"},
	{RequirementID: "FR-105", Verbs: "validate, verify"},
	{RequirementID: "FR-106", Verbs: "validate"},
	{RequirementID: "FR-107", Verbs: "validate"},
	{RequirementID: "FR-108", Verbs: "validate"},
	{RequirementID: "FR-109", Verbs: "validate"},
	{RequirementID: "FR-110", Verbs: "validate"},
	{RequirementID: "FR-116", Verbs: "merge"},
	{RequirementID: "FR-117", Verbs: "validate"},
	{RequirementID: "FR-119", Verbs: "migrate"},
	{RequirementID: "FR-120", Verbs: "migrate"},
	{RequirementID: "FR-121", Verbs: "migrate"},
	{RequirementID: "FR-122", Verbs: "migrate"},
	{RequirementID: "FR-123", Verbs: "migrate"},
	{RequirementID: "NFR-004", Verbs: "publish"},
	{RequirementID: "NFR-012", Verbs: "extract"},
	{RequirementID: "NFR-013", Verbs: "extract"},
	{RequirementID: "NFR-014", Verbs: "extract"},
	{RequirementID: "CON-003", Verbs: "validate"},
	{RequirementID: "CON-009", Verbs: "validate"},
	{RequirementID: "CON-010", Verbs: "validate"},
	{RequirementID: "CON-011", Verbs: "validate, verify"},
	{RequirementID: "TR-004", Verbs: "project"},
	{RequirementID: "TR-005", Verbs: "project"},
	{RequirementID: "TR-006", Verbs: "inspect, validate"},
	{RequirementID: "TR-007", Verbs: "inspect, validate"},
	{RequirementID: "TR-008", Verbs: "inspect, validate"},
	{RequirementID: "TR-011", Verbs: "extract"},
}
