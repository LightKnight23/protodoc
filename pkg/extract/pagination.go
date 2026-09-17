// Page adjunct emission (T-0092, FR-045); the stale-omission half is T-0093. A PageDirectory is
// a derived pagination artefact bound to the content it was computed from by
// an input digest. When that digest matches the current content (the
// directory is CURRENT), the extraction view emits a page-number adjunct
// alongside each locator; when it is stale or absent, adjuncts are omitted
// and PaginationStale is set rather than a page number being guessed
// (T-0093). Library-level only: wiring page adjuncts / the stale flag into
// the extract CLI verb's stdout payload is M18's concern (the plan.md
// review's disclosed gap) -- flagged forward to the M18 CLI task.
package extract

import (
	"Protodoc/pkg/pdlfmt"
)

// PageDirectory maps content units to page numbers and records the input
// digest of the content it was computed from. It is a derived artefact: a
// reader rebuilds or ignores it, and never trusts a page number whose input
// digest does not match current content.
type PageDirectory struct {
	// InputDigest is the digest of the content the directory was computed
	// from; the directory is current iff it equals the current content
	// digest.
	InputDigest pdlfmt.Digest256
	// Pages maps a unit identity to its 1-based page number.
	Pages map[pdlfmt.UnitID]uint32
}

// IsCurrent reports whether the directory's input digest matches the given
// current content digest.
func (d PageDirectory) IsCurrent(currentContentDigest pdlfmt.Digest256) bool {
	return d.InputDigest == currentContentDigest
}

// PageAdjunct is the optional page-number annotation attached to a locator
// when pagination is current. A nil *PageAdjunct means "no page adjunct"
// (pagination absent or stale).
type PageAdjunct struct {
	Page uint32
}

// PageAdjunctFor returns the page adjunct for unitID from a CURRENT page
// directory: a non-nil adjunct with the unit's page when the directory is
// current and holds the unit, or nil otherwise. currentContentDigest is the
// digest of the content being extracted, used to confirm the directory is
// not stale before any page number is emitted (FR-045).
func PageAdjunctFor(dir *PageDirectory, unitID pdlfmt.UnitID, currentContentDigest pdlfmt.Digest256) *PageAdjunct {
	if dir == nil || !dir.IsCurrent(currentContentDigest) {
		return nil
	}
	page, ok := dir.Pages[unitID]
	if !ok {
		return nil
	}
	return &PageAdjunct{Page: page}
}

// PaginatedLocator pairs a locator with its (possibly nil) page adjunct.
type PaginatedLocator struct {
	Locator Locator
	Page    *PageAdjunct
}

// PaginatedResult is an extraction result carrying per-locator page adjuncts
// and a document-level staleness flag. PaginationStale is true when the page
// directory was absent or its digest mismatched current content, in which
// case every PaginatedLocator.Page is nil (no guessed page numbers; the requirement for that is T-0093's).
type PaginatedResult struct {
	Locators        []PaginatedLocator
	PaginationStale bool
}

// PaginateLocators attaches page adjuncts to locators from dir, given the
// current content digest. When dir is current, each locator that has a page
// gets a non-nil adjunct and PaginationStale is false. When dir is nil or
// stale, every adjunct is nil and PaginationStale is true (FR-045; stale-path requirement is T-0093's).
func PaginateLocators(locs []Locator, dir *PageDirectory, currentContentDigest pdlfmt.Digest256) PaginatedResult {
	stale := dir == nil || !dir.IsCurrent(currentContentDigest)
	out := make([]PaginatedLocator, len(locs))
	for i, l := range locs {
		var adj *PageAdjunct
		if !stale {
			adj = PageAdjunctFor(dir, l.UnitID, currentContentDigest)
		}
		out[i] = PaginatedLocator{Locator: l, Page: adj}
	}
	return PaginatedResult{Locators: out, PaginationStale: stale}
}
