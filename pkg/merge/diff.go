// Construct-level diff (T-0239; TR-002). Differences and three-way merges are
// expressed in terms of document constructs (identified units) rather than
// storage units (segments/octets), and a construct on which the two inputs
// agree is not reported at all. This is the semantic layer the CLI diff/merge
// driver reports from. (The conflict-reporting guard is added by T-0240.)
package merge

import (
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// ChangeKind is how a construct differs between two inputs.
type ChangeKind uint8

const (
	// Added: the construct exists only in the right input.
	Added ChangeKind = iota
	// Removed: the construct exists only in the left input.
	Removed
	// Modified: the construct exists in both with differing content.
	Modified
)

// ConstructChange is one reported difference, keyed by the construct's identity
// (a document construct, not a storage unit). A construct on which both inputs
// agree is never emitted.
type ConstructChange struct {
	Construct pdlfmt.UnitID
	Kind      ChangeKind
	Left      []byte // content in the left input (nil if Added)
	Right     []byte // content in the right input (nil if Removed)
}

// DiffConstructs reports the differences between two inputs, each a map from a
// construct's identity to its content, in terms of constructs (TR-002). It
// reports no construct on which the two inputs agree (identical content), and
// keys every reported change by the construct identity -- never by segment,
// octet range, or any other storage unit. The result is sorted by construct id
// for determinism.
func DiffConstructs(left, right map[pdlfmt.UnitID][]byte) []ConstructChange {
	seen := make(map[pdlfmt.UnitID]bool, len(left)+len(right))
	var changes []ConstructChange

	for id, lv := range left {
		seen[id] = true
		rv, inRight := right[id]
		switch {
		case !inRight:
			changes = append(changes, ConstructChange{Construct: id, Kind: Removed, Left: lv})
		case !bytesEqual(lv, rv):
			changes = append(changes, ConstructChange{Construct: id, Kind: Modified, Left: lv, Right: rv})
			// equal content -> agreement -> not reported (TR-002).
		}
	}
	for id, rv := range right {
		if seen[id] {
			continue
		}
		changes = append(changes, ConstructChange{Construct: id, Kind: Added, Right: rv})
	}

	sort.Slice(changes, func(i, j int) bool {
		return unitIDLess(changes[i].Construct, changes[j].Construct)
	})
	return changes
}

// unitIDLess reports whether a sorts before b byte-lexicographically.
func unitIDLess(a, b pdlfmt.UnitID) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
