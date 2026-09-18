// Role resolution (FR-038; T-0280). Validator rule PD-A11Y-002: every
// construct instance in a document MUST resolve to exactly one accessibility
// role from the closed role map (rolemap.go / data-model.md 2.28). A construct
// whose kind is unmapped — a reserved discriminant, or a kind a future version
// added that this reader does not know — FAILS CLOSED: it is rejected naming
// the offending construct, never silently assigned a default role.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import "Protodoc/pkg/pdlfmt"

// RuleRoleResolution is the validator rule id for role resolution.
const RuleRoleResolution = "PD-A11Y-002"

// ConstructInstance is one construct occurrence in a document: its identity and
// its kind (frame discriminant).
type ConstructInstance struct {
	UnitID pdlfmt.UnitID
	Kind   ConstructKind
}

// RoleResolutionFinding is a PD-A11Y-002 rejection naming a construct that does
// not resolve to exactly one role.
type RoleResolutionFinding struct {
	Rule   string
	UnitID pdlfmt.UnitID
	Kind   ConstructKind
}

// CheckRoleResolution applies PD-A11Y-002: every construct instance must
// resolve to exactly one role via RoleOf. An instance whose kind is unmapped
// fails closed and is reported, naming the offending construct's unit id and
// kind. An empty result means every construct resolved.
func CheckRoleResolution(instances []ConstructInstance) []RoleResolutionFinding {
	var findings []RoleResolutionFinding
	for _, in := range instances {
		if _, ok := RoleOf(in.Kind); !ok {
			findings = append(findings, RoleResolutionFinding{
				Rule: RuleRoleResolution, UnitID: in.UnitID, Kind: in.Kind,
			})
		}
	}
	return findings
}
