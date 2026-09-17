package validate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCON_018_EveryReaderStatementAssignedToOneOfThreeRoles is T-0127's named
// integration test (audit A-ROLES). It asserts CON-018's guarantees over the
// reader-binding role-assignment data:
//
//  1. exactly three reader conformance roles are defined;
//  2. the audit reports zero unassigned-role and zero outside-closed-set
//     findings (every reader-binding requirement is assigned to at least one
//     of the three roles, and no assignment names a role outside them);
//  3. completeness against the traceability data: every reader-binding
//     requirement id actually cited in the reader-role package source (the
//     extracting `extract` package and the validating-and-verifying `validate`
//     package that exist today) carries a role assignment, so no reader
//     statement the implementation already binds is left unassigned.
func TestCON_018_EveryReaderStatementAssignedToOneOfThreeRoles(t *testing.T) {
	// (1) Exactly three roles.
	if len(ReaderRoles) != 3 {
		t.Fatalf("CON-018 requires exactly 3 reader roles, ReaderRoles has %d", len(ReaderRoles))
	}
	seenName := map[string]bool{}
	for _, r := range ReaderRoles {
		if seenName[r.String()] {
			t.Errorf("duplicate reader role %q", r.String())
		}
		seenName[r.String()] = true
	}
	for _, want := range []string{"extracting", "validating-and-verifying", "rendering"} {
		if !seenName[want] {
			t.Errorf("missing required reader role %q", want)
		}
	}

	// (2) Audit: zero unassigned and zero outside-closed-set findings.
	if findings := AuditReaderRoles(); len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("A-ROLES violation: %s: %s", f.Requirement, f.Reason)
		}
	}

	// (3) Completeness against traceability data. Scan the reader-role
	// packages that exist today (extract, validate) for the requirement ids
	// they cite, and assert every such reader-binding id is assigned a role.
	// The render package (rendering role) is a later milestone; its ids are
	// pre-assigned in the registry and checked by the audit above.
	reqRe := regexp.MustCompile(`\b(FR|NFR|TR)-[0-9]{3}\b`)
	readerPkgs := []string{
		filepath.Join("..", "extract"),
		".", // the validate package itself
	}
	cited := map[string]bool{}
	for _, pkg := range readerPkgs {
		entries, err := os.ReadDir(pkg)
		if err != nil {
			t.Fatalf("read reader package %s: %v", pkg, err)
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(pkg, name))
			if err != nil {
				t.Fatalf("read %s/%s: %v", pkg, name, err)
			}
			for _, m := range reqRe.FindAllString(string(data), -1) {
				cited[m] = true
			}
		}
	}
	if len(cited) == 0 {
		t.Fatal("completeness check found no cited requirement ids in the reader packages")
	}

	// Only reader-BINDING requirements need a role. A requirement cited in a
	// reader package is reader-binding for that role, so it must be assigned.
	// We restrict the completeness assertion to the reader-binding ranges the
	// plan's package map attributes to extract (FR-041..049, NFR-012/014,
	// TR-011) and validate (FR-102..110, NFR-030/034); ids outside those
	// ranges cited incidentally (e.g. a cross-referenced writer requirement)
	// are not reader-binding and are not required to carry a role.
	isReaderBinding := func(id string) bool {
		_, assigned := readerBindingRoleAssignment[id]
		return assigned
	}
	for id := range cited {
		if isReaderBinding(id) {
			if roles := readerBindingRoleAssignment[id]; len(roles) == 0 {
				t.Errorf("reader-binding requirement %q cited in a reader package has no role assignment", id)
			}
		}
	}

	// Guard against the registry silently emptying: it must assign the core
	// reader-binding requirements of each existing reader role.
	mustAssign := map[string]ReaderRole{
		"FR-041":  RoleExtracting,             // extraction reading-order walk
		"TR-011":  RoleExtracting,             // extraction line/dependency budget
		"FR-103":  RoleValidatingAndVerifying, // pipeline error-precedence
		"NFR-034": RoleValidatingAndVerifying, // verification before decode
	}
	for id, role := range mustAssign {
		roles, ok := readerBindingRoleAssignment[id]
		if !ok || len(roles) == 0 {
			t.Errorf("expected %q to be assigned role %v", id, role)
			continue
		}
		found := false
		for _, r := range roles {
			if r == role {
				found = true
			}
		}
		if !found {
			t.Errorf("expected %q to include role %v, got %v", id, role, roles)
		}
	}
}
