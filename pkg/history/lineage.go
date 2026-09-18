// State lineage (T-0210/T-0211, FR-005; document.abnf S9 op-predecessor). The
// edit-operation log's op-predecessor fields form a verifiable causal chain of
// states, independent of any client clock. A state's predecessor is the
// causally preceding state's state-id; walking those links back yields the
// state's ancestry, and the nearest common ancestor of two divergent copies is
// computable from the two files' own operation logs alone -- no coordinating
// service, no network (FR-005).
package history

import (
	"errors"

	"Protodoc/pkg/container"
)

// Lineage is a state graph built from operation predecessors: it maps each
// state-id to its immediate predecessor (parent) state-id. The root state has
// no predecessor (mapped to the zero state-id, or absent).
type Lineage struct {
	parent map[container.StateID]container.StateID
	known  map[container.StateID]bool
}

var zeroStateID container.StateID

// NewLineage builds a Lineage from an operation log: every op contributes the
// edge OrderKey(authoring state) -> Predecessor(preceding state). A state
// present only as a predecessor (never an authoring state) is still known as a
// root-side node. The construction is a pure function of the log, so two
// implementations build the identical graph.
func NewLineage(ops []OperationRecord) Lineage {
	l := Lineage{parent: map[container.StateID]container.StateID{}, known: map[container.StateID]bool{}}
	for _, op := range ops {
		state := op.OrderKey
		pred := op.Predecessor
		l.known[state] = true
		l.known[pred] = true
		// Record the parent edge. If a state already has a parent, keep the
		// first (a well-formed log gives each state one predecessor).
		if _, ok := l.parent[state]; !ok {
			l.parent[state] = pred
		}
	}
	return l
}

// AddState records a state's parent explicitly (used when assembling a lineage
// from state metadata rather than an op log). A zero parent marks a root.
func (l *Lineage) AddState(state, parent container.StateID) {
	if l.parent == nil {
		l.parent = map[container.StateID]container.StateID{}
		l.known = map[container.StateID]bool{}
	}
	l.known[state] = true
	l.parent[state] = parent
	if parent != zeroStateID {
		l.known[parent] = true
	}
}

// ErrUnknownState is returned when a walk starts from a state not in the graph.
var ErrUnknownState = errors.New("history: state is not present in the lineage graph")

// WalkPredecessors returns the chain of states from `state` back to the root,
// inclusive of `state` and ending at the root (a state whose predecessor is
// the zero state-id or is not itself recorded). It detects and stops on a
// cycle (a malformed log) rather than looping forever, returning the chain up
// to the repeat. It errors if `state` is unknown.
func (l Lineage) WalkPredecessors(state container.StateID) ([]container.StateID, error) {
	if !l.known[state] {
		return nil, ErrUnknownState
	}
	var chain []container.StateID
	seen := map[container.StateID]bool{}
	cur := state
	for {
		if seen[cur] {
			break // cycle guard
		}
		seen[cur] = true
		chain = append(chain, cur)
		parent, ok := l.parent[cur]
		if !ok || parent == zeroStateID {
			break // reached a root
		}
		cur = parent
	}
	return chain, nil
}

// Ancestors returns the set of ancestor state-ids of `state` (including
// `state` itself), for common-ancestor computation.
func (l Lineage) Ancestors(state container.StateID) (map[container.StateID]bool, error) {
	chain, err := l.WalkPredecessors(state)
	if err != nil {
		return nil, err
	}
	set := make(map[container.StateID]bool, len(chain))
	for _, s := range chain {
		set[s] = true
	}
	return set, nil
}

// ErrNoCommonAncestor is returned when two states share no common ancestor in
// the combined lineage (they are unrelated).
var ErrNoCommonAncestor = errors.New("history: the two states share no common ancestor")

// NearestCommonAncestor returns the nearest common ancestor state of a and b,
// computed from this lineage alone (FR-005): no coordinating service, no
// network. It walks a's ancestry to a set, then walks b's ancestry until it
// hits a state in that set -- the first hit is the nearest common ancestor
// (nearest to b along its own chain, which for a tree lineage is the deepest
// shared state). Both a and b must be known to the graph. It errors if the two
// share no common ancestor.
//
// For divergent copies, the caller MERGES the two files' lineages (each file
// carries its own operation log with the shared prefix) into one Lineage, then
// calls this: the shared prefix's states appear in both logs with identical
// state-ids, so the common ancestry is discoverable from the two files alone.
func (l Lineage) NearestCommonAncestor(a, b container.StateID) (container.StateID, error) {
	aSet, err := l.Ancestors(a)
	if err != nil {
		return zeroStateID, err
	}
	if !l.known[b] {
		return zeroStateID, ErrUnknownState
	}
	bChain, err := l.WalkPredecessors(b)
	if err != nil {
		return zeroStateID, err
	}
	for _, s := range bChain {
		if aSet[s] {
			return s, nil
		}
	}
	return zeroStateID, ErrNoCommonAncestor
}

// Merge folds another lineage's edges into this one (used to combine two
// divergent copies' operation-log lineages for common-ancestor computation).
// A shared state keeps its existing parent (the shared prefix is identical in
// both, so no conflict arises for a well-formed pair).
func (l *Lineage) Merge(other Lineage) {
	if l.parent == nil {
		l.parent = map[container.StateID]container.StateID{}
		l.known = map[container.StateID]bool{}
	}
	for s := range other.known {
		l.known[s] = true
	}
	for s, p := range other.parent {
		if _, ok := l.parent[s]; !ok {
			l.parent[s] = p
		}
	}
}
