// Per-role normative-statement ceiling check (NFR-025; T-0344). CON-018's
// three reader roles each have a stated maximum normative-statement count in
// data-model.md's ceiling table (Extraction<=150, Validating-and-verifying<=400
// cumulative, Rendering<=500 cumulative; CP-002/CON-009). This file counts the
// statements assigned to each role (from the A-ROLES reader-binding map) and
// exposes the counts + ceiling names so CI can fail on any overage.
package validate

import "Protodoc/pkg/ceilings"

// roleCeilingName maps each reader role to its ceiling-table row name.
var roleCeilingName = map[ReaderRole]string{
	RoleExtracting:             "Extraction role statement ceiling",
	RoleValidatingAndVerifying: "Validating-and-verifying role statement ceiling (cumulative)",
	RoleRendering:              "Rendering role statement ceiling (cumulative)",
}

// RoleCeilingName returns the ceiling-table row name for a reader role.
func RoleCeilingName(r ReaderRole) string { return roleCeilingName[r] }

// CountStatementsByRole returns the number of distinct normative statements
// assigned to each reader role by the A-ROLES reader-binding map. A statement
// bound to several roles counts once per role it binds. Roles are cumulative
// per CP-002: validating-and-verifying includes extracting, and rendering
// includes both, mirroring data-model.md's "(cumulative)" ceiling labels.
func CountStatementsByRole() map[ReaderRole]int {
	direct := map[ReaderRole]int{}
	for _, roles := range readerBindingRoleAssignment {
		for _, r := range roles {
			direct[r]++
		}
	}
	// Cumulative rollup (CP-002): extracting is the base; validating-and-
	// verifying adds to it; rendering adds to that.
	cumulative := map[ReaderRole]int{
		RoleExtracting:             direct[RoleExtracting],
		RoleValidatingAndVerifying: direct[RoleExtracting] + direct[RoleValidatingAndVerifying],
		RoleRendering:              direct[RoleExtracting] + direct[RoleValidatingAndVerifying] + direct[RoleRendering],
	}
	return cumulative
}

// RoleCeiling returns the enforced statement ceiling for a role from the
// checked-in ceilings table.
func RoleCeiling(r ReaderRole) uint64 {
	return ceilings.MustMax(roleCeilingName[r])
}

// RoleCeilingOverage returns, for each role, how far its statement count
// exceeds its ceiling (0 when within budget). A non-zero value for any role is
// a build failure (NFR-025).
func RoleCeilingOverage() map[ReaderRole]int {
	counts := CountStatementsByRole()
	over := map[ReaderRole]int{}
	for _, r := range ReaderRoles {
		ceil := int(RoleCeiling(r))
		if counts[r] > ceil {
			over[r] = counts[r] - ceil
		}
	}
	return over
}
