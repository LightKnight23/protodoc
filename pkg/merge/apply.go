// Merge application (T-0226..; FR-092). Applying a set of concurrent
// operations must be order-independent for disjoint (commuting) operations:
// two conforming implementations given the same changes in ANY delivery order
// produce identical document state. This models a document as a map of target
// unit-id -> value, and applies operations classified by DP-015; disjoint
// operations touch distinct targets and commute.
package merge

import (
	"sort"

	"Protodoc/pkg/history"
	"Protodoc/pkg/pdlfmt"
)

// Op is a merge operation: its kind, target, authoring state-id (the R2
// tiebreak key), and value payload.
type Op struct {
	Kind    history.OpKind
	Target  pdlfmt.UnitID
	StateID [32]byte // authoring state id, unsigned big-endian tiebreak key
	Value   []byte
}

// Doc is a merged document state: target unit-id -> value. A deleted target is
// absent from the map.
type Doc map[pdlfmt.UnitID][]byte

// canonical returns a deterministic serialization of the document for
// order-independence comparison: targets in ascending unit-id order.
func (d Doc) canonical() []byte {
	ids := make([]pdlfmt.UnitID, 0, len(d))
	for id := range d {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		for k := range ids[i] {
			if ids[i][k] != ids[j][k] {
				return ids[i][k] < ids[j][k]
			}
		}
		return false
	})
	var out []byte
	for _, id := range ids {
		out = append(out, id[:]...)
		out = append(out, byte(len(d[id])>>8), byte(len(d[id])))
		out = append(out, d[id]...)
	}
	return out
}

// ApplyDisjoint applies a set of operations that are pairwise disjoint (each
// targets a distinct unit) to a base document, returning the merged state. For
// disjoint operations the delivery order is irrelevant: each op independently
// sets or deletes its own target. A value/position claim sets the target's
// value; a delete removes it.
func ApplyDisjoint(base Doc, ops []Op) Doc {
	out := Doc{}
	for id, v := range base {
		out[id] = append([]byte(nil), v...)
	}
	for _, op := range ops {
		switch op.Kind {
		case history.OpDelete:
			delete(out, op.Target)
		default:
			out[op.Target] = append([]byte(nil), op.Value...)
		}
	}
	return out
}
