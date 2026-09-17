// Reference-graph cycle detection (T-0118, FR-109; data-model.md S7 step 8).
// A directed graph is built over the FIVE named edge kinds and cycle
// detection runs strictly BEFORE the structural-ceiling step in pipeline
// order, so a cyclic input can never be misreported as merely over-ceiling.
// On the first cycle found, every edge participating in it is named.
//
// FLAGGED, self-disclosed gap (contracts/README.md, data-model.md S7): a
// CROSS_REFERENCE edge is NOT one of the 5 kinds, even though spec.md's
// FR-109 text names cross-references as one of its 4 abstract graph
// categories. This implements the 5-kind design as currently approved and
// records the gap here rather than unilaterally expanding the list; the
// per-edge resolution check (FR-108, T-0117) still applies to cross-refs.
package validate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// EdgeKind is one of the five named reference-graph edge kinds (and only
// these five; CROSS_REFERENCE is intentionally excluded, see the flag above).
type EdgeKind string

const (
	EdgeStructuralMove   EdgeKind = "structural-move"
	EdgeExtFallbackRef   EdgeKind = "ext-envelope-fallback-ref"
	EdgeAnnotationAnchor EdgeKind = "annotation-anchor-ref"
	EdgeRunSplitMerge    EdgeKind = "run-split-merge-lineage"
	EdgeRescindResign    EdgeKind = "rescind-resign-chain"
)

// validEdgeKinds is the closed set of the 5 kinds.
var validEdgeKinds = map[EdgeKind]bool{
	EdgeStructuralMove: true, EdgeExtFallbackRef: true, EdgeAnnotationAnchor: true,
	EdgeRunSplitMerge: true, EdgeRescindResign: true,
}

// Edge is a directed reference-graph edge From -> To of a named kind.
type Edge struct {
	Kind EdgeKind
	From pdlfmt.UnitID
	To   pdlfmt.UnitID
}

// CycleError is FR-109's rejection: it names every edge participating in the
// detected cycle, in cycle order.
type CycleError struct {
	Edges []Edge
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("validate: reference-graph cycle over %d edges (FR-109)", len(e.Edges))
}

// DetectCycle builds the directed graph from edges (all of whose kinds must
// be one of the 5 named kinds) and returns a *CycleError naming every edge
// in the FIRST cycle found, or nil if the graph is acyclic. It ignores no
// edge kind and includes a single-node self-cycle (From == To). Detection is
// a depth-first search with a recursion stack; the reported cycle is the
// edges along the back-edge's stack path.
func DetectCycle(edges []Edge) error {
	for _, e := range edges {
		if !validEdgeKinds[e.Kind] {
			return fmt.Errorf("validate: edge of unrecognised kind %q (only the 5 named kinds are permitted)", e.Kind)
		}
	}

	// Adjacency: node -> outgoing edges.
	adj := map[pdlfmt.UnitID][]Edge{}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e)
	}

	const (
		white = 0 // unvisited
		gray  = 1 // on the current DFS stack
		black = 2 // fully explored
	)
	color := map[pdlfmt.UnitID]int{}
	// stackEdge[node] is the edge by which node was entered (for cycle
	// reconstruction).
	stackEdge := map[pdlfmt.UnitID]Edge{}

	var found *CycleError
	var dfs func(n pdlfmt.UnitID)
	dfs = func(n pdlfmt.UnitID) {
		if found != nil {
			return
		}
		color[n] = gray
		for _, e := range adj[n] {
			if found != nil {
				return
			}
			if color[e.To] == gray {
				// Back edge -> cycle. Reconstruct the edges from e.To around
				// back to n, then close with e.
				found = reconstructCycle(e, stackEdge, n)
				return
			}
			if color[e.To] == white {
				stackEdge[e.To] = e
				dfs(e.To)
			}
		}
		color[n] = black
	}

	// Deterministic iteration: sort start nodes so the "first" cycle is
	// reproducible. Collect and sort node keys.
	starts := sortedNodes(adj)
	for _, n := range starts {
		if color[n] == white {
			dfs(n)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

// reconstructCycle walks the entry-edge chain from backEdge.From up to the
// cycle start (backEdge.To), collecting the edges that form the cycle, then
// appends the closing back edge.
func reconstructCycle(backEdge Edge, stackEdge map[pdlfmt.UnitID]Edge, cur pdlfmt.UnitID) *CycleError {
	target := backEdge.To
	var chain []Edge
	// Walk from cur back to target via entry edges.
	node := cur
	for {
		e, ok := stackEdge[node]
		if !ok {
			break
		}
		chain = append([]Edge{e}, chain...)
		if e.From == target {
			break
		}
		node = e.From
	}
	chain = append(chain, backEdge)
	return &CycleError{Edges: chain}
}

func sortedNodes(adj map[pdlfmt.UnitID][]Edge) []pdlfmt.UnitID {
	nodes := make([]pdlfmt.UnitID, 0, len(adj))
	for n := range adj {
		nodes = append(nodes, n)
	}
	// byte-lexicographic sort for determinism
	for i := 1; i < len(nodes); i++ {
		for j := i; j > 0 && lessUnitID(nodes[j], nodes[j-1]); j-- {
			nodes[j-1], nodes[j] = nodes[j], nodes[j-1]
		}
	}
	return nodes
}

func lessUnitID(a, b pdlfmt.UnitID) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}
