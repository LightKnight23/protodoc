// Referential-integrity check (T-0117, FR-108): every xref-target,
// index-entry, annotation range endpoint (run_id + base_ordinal) and
// identity reference must resolve to exactly one currently-present unit. An
// unresolvable (zero matches) or multiply-resolving (>1 match) reference is
// rejected outright -- never silently dropped, defaulted, or auto-repaired.
package validate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// RefKind names the kind of reference for diagnostics.
type RefKind string

const (
	RefXrefTarget      RefKind = "xref-target"
	RefIndexEntry      RefKind = "index-entry"
	RefAnnotationStart RefKind = "annotation-start"
	RefAnnotationEnd   RefKind = "annotation-end"
	RefIdentity        RefKind = "identity-reference"
)

// Reference is one reference to resolve: its kind (for diagnostics), the
// referrer's own id, and the target unit id it must resolve to.
type Reference struct {
	Kind   RefKind
	From   pdlfmt.UnitID
	Target pdlfmt.UnitID
}

// PresenceCount maps a unit id to how many times it is currently present in
// the document (an id present exactly once resolves; zero is dangling, more
// than one is ambiguous). Built from the document's actual units.
type PresenceCount map[pdlfmt.UnitID]int

// ReferenceResolutionError is FR-108's rejection: it names the offending
// reference kind, referrer, target, and how many units the target resolved
// to (0 = dangling, >1 = ambiguous).
type ReferenceResolutionError struct {
	Kind     RefKind
	From     pdlfmt.UnitID
	Target   pdlfmt.UnitID
	Resolved int
}

func (e *ReferenceResolutionError) Error() string {
	reason := "dangling (resolves to no present unit)"
	if e.Resolved > 1 {
		reason = fmt.Sprintf("ambiguous (resolves to %d present units)", e.Resolved)
	}
	return fmt.Sprintf("validate: %s reference from %x to %x is %s (FR-108)", e.Kind, e.From, e.Target, reason)
}

// CheckReferences resolves every reference against present and returns a
// *ReferenceResolutionError for the FIRST reference that does not resolve to
// exactly one present unit. It never repairs or drops a reference. A nil
// return means every reference resolved to exactly one present unit.
func CheckReferences(refs []Reference, present PresenceCount) error {
	for _, r := range refs {
		n := present[r.Target]
		if n != 1 {
			return &ReferenceResolutionError{Kind: r.Kind, From: r.From, Target: r.Target, Resolved: n}
		}
	}
	return nil
}
