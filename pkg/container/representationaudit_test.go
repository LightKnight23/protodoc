package container

import "testing"

// TestCON_005_ExactlyOneRepresentationPerCapability is T-0026's named test.
// Implements: CON-005.
//
// It walks auditRepresentations' output -- every field of every audited
// pkg/container/pkg/pdlfmt struct type, classified into the numeric,
// repeated-value or geometric capability category -- and asserts each
// category is expressed by exactly the one representation form
// permittedForm names for it: a field expressing that category any other
// way (a float alongside the fixed-point/integer form, a map alongside the
// sequence form) fails the audit by name, not just by a mismatched count.
func TestCON_005_ExactlyOneRepresentationPerCapability(t *testing.T) {
	observations := auditRepresentations()
	if len(observations) == 0 {
		t.Fatal("auditRepresentations found no fields to audit, audit has nothing to check")
	}

	seen := make(map[Capability]map[representationForm]bool)
	for _, o := range observations {
		want, ok := permittedForm[o.Category]
		if !ok {
			t.Fatalf("field %s.%s classified into capability %q, which has no entry in permittedForm", o.TypeName, o.Field, o.Category)
		}
		if o.Form != want {
			t.Fatalf("field %s.%s expresses capability %q via representation form %q -- a second mechanism alongside the one normative form %q CON-005 permits for that capability", o.TypeName, o.Field, o.Category, o.Form, want)
		}
		if seen[o.Category] == nil {
			seen[o.Category] = make(map[representationForm]bool)
		}
		seen[o.Category][o.Form] = true
	}

	for cat, want := range permittedForm {
		forms := seen[cat]
		if len(forms) == 0 {
			t.Fatalf("capability %q has zero observed fields in the audited types; the audit cannot confirm it has exactly one representation form", cat)
		}
		if len(forms) != 1 || !forms[want] {
			t.Fatalf("capability %q has representation forms %v, want exactly {%q}", cat, forms, want)
		}
	}
}
