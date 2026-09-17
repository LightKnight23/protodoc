// Reader conformance-role audit (T-0127, CON-018, audit A-ROLES). CON-018
// requires the format define EXACTLY THREE reader conformance roles --
// extracting, validating-and-verifying, and rendering -- and assign every
// normative statement that binds a conforming reader to at least one of them.
// The roles correspond to plan.md Section 4's package map: the extracting
// role is the `extract` package (reading-order text walk over CONTENT only),
// the validating-and-verifying role is the `validate` (+ `integrity`,
// `migrate`) pipeline, and the rendering role is the `render` package. Roles
// are what deliver the promise that an indexer or scanner need not implement
// a renderer.
//
// This file holds the A-ROLES data: the closed set of three roles and the
// assignment of every reader-binding requirement (any FR-*/NFR-* whose plan
// mechanism cites extract, validate, or render) to at least one role. The
// T-0127 audit asserts (1) exactly three roles, (2) every reader-binding
// requirement carries a non-empty role set, and (3) no assignment names a
// role outside the closed set of three.
package validate

// ReaderRole is one of CON-018's exactly three reader conformance roles.
type ReaderRole int

const (
	RoleExtracting ReaderRole = iota
	RoleValidatingAndVerifying
	RoleRendering
)

func (r ReaderRole) String() string {
	switch r {
	case RoleExtracting:
		return "extracting"
	case RoleValidatingAndVerifying:
		return "validating-and-verifying"
	case RoleRendering:
		return "rendering"
	default:
		return "unknown"
	}
}

// ReaderRoles is the closed set of exactly three reader conformance roles
// CON-018 defines. The audit fails if any assignment names a role outside it.
var ReaderRoles = []ReaderRole{RoleExtracting, RoleValidatingAndVerifying, RoleRendering}

// readerBindingRoleAssignment maps every reader-binding requirement id to the
// reader role(s) it binds, per plan.md Section 4's package map. A requirement
// may bind more than one role (e.g. a size budget stated against both
// extracting and validating). The assignment is derived from the plan's
// package responsibilities:
//
//   - extracting  -> the `extract` package: reading-order text extraction,
//     locator emission, streaming/abandonable, the TR-011 line budget and the
//     NFR-012/014 read/memory bounds.
//   - validating-and-verifying -> the `validate`/`integrity`/`migrate`
//     packages: the 13-step pipeline, tree/signature verification, ceilings,
//     the validator memory and decode-ordering bounds.
//   - rendering -> the `render` package: presentation-artefact resolution,
//     the PLP-1 and restricted-PNG codecs, rasterizer, reflow, pagination.
//
// The requirement ids are BUILT programmatically (prefix + number) rather
// than written as literal tokens, so this reference-data file does not
// register spurious traceability "coverage" for requirements whose
// conformance cases live in their own (in several cases not-yet-built)
// milestones; the only requirement this A-ROLES audit genuinely satisfies is
// CON-018 itself.
var readerBindingRoleAssignment = buildReaderBindingAssignment()

// req builds a requirement id from a prefix and number, e.g. req("FR", 41) ->
// the extraction reading-order requirement id. Building ids this way keeps
// whole requirement-id tokens out of this file's source so the traceability
// scanner does not mistake this role-assignment data for conformance coverage
// of those requirements.
func req(prefix string, n int) string {
	d := []byte{byte('0' + (n/100)%10), byte('0' + (n/10)%10), byte('0' + n%10)}
	return prefix + "-" + string(d)
}

func addRange(m map[string][]ReaderRole, prefix string, lo, hi int, role ReaderRole) {
	for n := lo; n <= hi; n++ {
		m[req(prefix, n)] = []ReaderRole{role}
	}
}

func buildReaderBindingAssignment() map[string][]ReaderRole {
	m := map[string][]ReaderRole{}

	// Extracting role (extract package).
	addRange(m, "FR", 41, 49, RoleExtracting)
	m[req("NFR", 12)] = []ReaderRole{RoleExtracting}
	m[req("NFR", 14)] = []ReaderRole{RoleExtracting}
	m[req("TR", 11)] = []ReaderRole{RoleExtracting}

	// Validating-and-verifying role (validate/integrity/migrate).
	addRange(m, "FR", 102, 110, RoleValidatingAndVerifying)
	m[req("NFR", 30)] = []ReaderRole{RoleValidatingAndVerifying}
	m[req("NFR", 34)] = []ReaderRole{RoleValidatingAndVerifying}

	// Rendering role (render package).
	addRange(m, "FR", 89, 101, RoleRendering)
	addRange(m, "FR", 111, 114, RoleRendering)
	addRange(m, "CON", 12, 14, RoleRendering)

	return m
}

// RoleAuditFinding describes a single A-ROLES violation.
type RoleAuditFinding struct {
	Requirement string
	Reason      string
}

// AuditReaderRoles runs the A-ROLES audit over the reader-binding role
// assignment and returns every violation: a reader-binding requirement with
// an empty role set (unassigned) or an assignment naming a role outside the
// closed set of three. An empty result means the audit passes.
func AuditReaderRoles() []RoleAuditFinding {
	closed := map[ReaderRole]bool{}
	for _, r := range ReaderRoles {
		closed[r] = true
	}
	var out []RoleAuditFinding
	for req, roles := range readerBindingRoleAssignment {
		if len(roles) == 0 {
			out = append(out, RoleAuditFinding{Requirement: req, Reason: "no reader role assigned"})
			continue
		}
		for _, role := range roles {
			if !closed[role] {
				out = append(out, RoleAuditFinding{Requirement: req, Reason: "assigned a role outside the closed set of three"})
			}
		}
	}
	return out
}
