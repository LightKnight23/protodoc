// Fixed pagination projection (T-0247; FR-097). A document's authored fixed
// pagination is expressed directly from the single document file -- the
// Frontmatter's page count and page dimensions plus the PageDirectory -- with
// NO conversion step: a single read-and-project path, no intermediate document
// representation, no transcoding pass between file read and pagination output.
package render

import (
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// FixedPagination is the authored pagination projected directly from the file:
// the page count and page geometry the writer authored, and the per-page break
// units in ordinal order. It is a projection, not a re-derivation.
type FixedPagination struct {
	PageCount  uint32
	PageWidth  pdlfmt.GeometricValue
	PageHeight pdlfmt.GeometricValue
	// BreakUnits are the page-break unit identities in ordinal order (the
	// PageDirectory's entries projected in rank order).
	BreakUnits []pdlfmt.UnitID
}

// ProjectFixedPagination produces the authored fixed pagination by reading the
// Frontmatter's authored page count/dimensions and projecting the
// PageDirectory's entries in ordinal (rank) order -- a single read-and-project
// path (FR-097). It performs NO conversion: the page count and geometry are
// taken verbatim from the frontmatter, and the break units are the directory's
// entries in order; nothing is transcoded or rebuilt into an intermediate
// representation. dir may be nil (no authored page-break directory), in which
// case only the frontmatter-authored count/geometry are projected.
func ProjectFixedPagination(fm *container.Frontmatter, dir *PageDirectory) FixedPagination {
	fp := FixedPagination{
		PageCount:  fm.PageCount,
		PageWidth:  fm.PageWidth,
		PageHeight: fm.PageHeight,
	}
	if dir != nil {
		for _, e := range dir.Entries() { // rank order == ordinal order
			fp.BreakUnits = append(fp.BreakUnits, e.BreakUnit)
		}
	}
	return fp
}
