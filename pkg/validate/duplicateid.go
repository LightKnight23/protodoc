// Duplicate-identifier rejection (T-0120, FR-110): any two stored units or
// index entries resolving to the same identifier are rejected, naming BOTH
// physical locations (segment ordinal + intra-segment offset for each) --
// never silently renaming either or applying precedence between them.
package validate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// Location is a physical position of a stored unit or index entry: its
// segment ordinal and the offset within that segment.
type Location struct {
	SegmentOrdinal uint16
	IntraOffset    uint64
}

func (l Location) String() string {
	return fmt.Sprintf("segment %d offset %d", l.SegmentOrdinal, l.IntraOffset)
}

// IdentifiedUnit is a stored unit or index entry with its identifier and
// physical location, the input to duplicate detection.
type IdentifiedUnit struct {
	ID       pdlfmt.UnitID
	Location Location
}

// DuplicateIdentifierError is FR-110's rejection: it names the shared
// identifier and BOTH physical locations that carry it.
type DuplicateIdentifierError struct {
	ID    pdlfmt.UnitID
	First Location
	Later Location
}

func (e *DuplicateIdentifierError) Error() string {
	return fmt.Sprintf("validate: identifier %x appears at two locations (%s and %s); rejected without rename or precedence (FR-110)", e.ID, e.First, e.Later)
}

// CheckDuplicateIdentifiers scans units and returns a
// *DuplicateIdentifierError for the FIRST identifier that appears at two
// distinct locations, naming both. It never renames or picks a winner. A nil
// return means every identifier is unique across the supplied units.
func CheckDuplicateIdentifiers(units []IdentifiedUnit) error {
	first := make(map[pdlfmt.UnitID]Location, len(units))
	for _, u := range units {
		if prev, seen := first[u.ID]; seen {
			return &DuplicateIdentifierError{ID: u.ID, First: prev, Later: u.Location}
		}
		first[u.ID] = u.Location
	}
	return nil
}
