// Differing-unit enumeration (T-0162, FR-068; cli.md S6 diff). A reader can
// enumerate every content unit that differs between a signed state and the
// current state. Telling a recipient the current state is unattested is
// useless without telling them WHAT changed; this turns a warning into an
// answer. A unit differs if it was added, removed, or its content digest
// changed between the two states.
package integrity

import (
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// UnitDiffKind classifies how a unit differs between two states.
type UnitDiffKind int

const (
	// UnitAdded: present in the current state, absent in the signed state.
	UnitAdded UnitDiffKind = iota
	// UnitRemoved: present in the signed state, absent in the current state.
	UnitRemoved
	// UnitChanged: present in both, but its content digest differs.
	UnitChanged
)

func (k UnitDiffKind) String() string {
	switch k {
	case UnitAdded:
		return "added"
	case UnitRemoved:
		return "removed"
	case UnitChanged:
		return "changed"
	default:
		return "unknown"
	}
}

// UnitDifference names one differing unit and how it differs.
type UnitDifference struct {
	Unit pdlfmt.UnitID
	Kind UnitDiffKind
}

// StateUnits maps a content unit's id to its content digest in one state. Two
// units are "the same" iff their digests are equal (content-addressed).
type StateUnits map[pdlfmt.UnitID]Digest

// EnumerateUnitDifferences returns every content unit that differs between the
// signed state and the current state (FR-068), in deterministic order (by
// unit-id octets). A unit present in both with an equal digest is not
// returned; added/removed/changed units are. The result is exhaustive: every
// difference is named, none is summarised away.
func EnumerateUnitDifferences(signed, current StateUnits) []UnitDifference {
	var diffs []UnitDifference
	for id, sd := range signed {
		cd, ok := current[id]
		if !ok {
			diffs = append(diffs, UnitDifference{Unit: id, Kind: UnitRemoved})
		} else if cd != sd {
			diffs = append(diffs, UnitDifference{Unit: id, Kind: UnitChanged})
		}
	}
	for id := range current {
		if _, ok := signed[id]; !ok {
			diffs = append(diffs, UnitDifference{Unit: id, Kind: UnitAdded})
		}
	}
	sort.Slice(diffs, func(i, j int) bool {
		c := compareUnitID(diffs[i].Unit, diffs[j].Unit)
		if c != 0 {
			return c < 0
		}
		return diffs[i].Kind < diffs[j].Kind
	})
	return diffs
}

func compareUnitID(a, b pdlfmt.UnitID) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
