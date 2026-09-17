package validate

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_109_CycleDetectionNamesAllEdgesBeforeCeilingChecks is T-0118's
// named conformance test. A single-edge self-cycle and a multi-edge cycle
// each produce a rejection naming every edge in the cycle; an acyclic graph
// passes; and a pipeline ordering test asserts cycle detection runs and
// rejects before any ceiling-check counter is incremented.
func TestFR_109_CycleDetectionNamesAllEdgesBeforeCeilingChecks(t *testing.T) {
	id := func(b byte) pdlfmt.UnitID { var u pdlfmt.UnitID; u[0] = b; return u }

	// Single-edge self-cycle A -> A.
	self := []Edge{{Kind: EdgeStructuralMove, From: id(1), To: id(1)}}
	err := DetectCycle(self)
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("self-cycle: got %v, want CycleError", err)
	}
	if len(ce.Edges) != 1 || ce.Edges[0].From != id(1) || ce.Edges[0].To != id(1) {
		t.Fatalf("self-cycle: edges = %+v, want the single self edge", ce.Edges)
	}

	// Multi-edge cycle A -> B -> C -> A (three edges).
	multi := []Edge{
		{Kind: EdgeAnnotationAnchor, From: id(1), To: id(2)},
		{Kind: EdgeRunSplitMerge, From: id(2), To: id(3)},
		{Kind: EdgeRescindResign, From: id(3), To: id(1)},
	}
	err = DetectCycle(multi)
	if !errors.As(err, &ce) {
		t.Fatalf("multi-cycle: got %v, want CycleError", err)
	}
	if len(ce.Edges) != 3 {
		t.Fatalf("multi-cycle: named %d edges, want all 3", len(ce.Edges))
	}
	// Every node in the cycle appears; the named edges form a closed loop.
	seen := map[pdlfmt.UnitID]bool{}
	for _, e := range ce.Edges {
		seen[e.From] = true
	}
	for _, n := range []byte{1, 2, 3} {
		if !seen[id(n)] {
			t.Fatalf("multi-cycle: node %d not named among cycle edges", n)
		}
	}

	// Acyclic graph (a DAG) passes.
	dag := []Edge{
		{Kind: EdgeStructuralMove, From: id(1), To: id(2)},
		{Kind: EdgeStructuralMove, From: id(1), To: id(3)},
		{Kind: EdgeExtFallbackRef, From: id(2), To: id(4)},
	}
	if err := DetectCycle(dag); err != nil {
		t.Fatalf("acyclic graph rejected: %v", err)
	}

	// Ordering: cycle detection rejects BEFORE any ceiling-check counter is
	// incremented. Model a pipeline where step 8 (cycle) precedes step 9
	// (ceiling), and assert the ceiling counter stays zero when a cycle is
	// present.
	ceilingChecks := 0
	steps := []Step{
		{ID: StepCycleDetection, Run: func() *Finding {
			if err := DetectCycle(multi); err != nil {
				return &Finding{Step: StepCycleDetection, RuleID: "FR-109", Message: err.Error()}
			}
			return nil
		}},
		{ID: StepStructuralCeilings, Run: func() *Finding {
			ceilingChecks++ // must never run when a cycle failed step 8
			return nil
		}},
	}
	res := Run(steps)
	if res.Valid {
		t.Fatalf("cyclic document reported valid")
	}
	if res.Validity == nil || res.Validity.Step != StepCycleDetection {
		t.Fatalf("verdict step = %v, want cycle-detection (step 8)", res.Validity)
	}
	if ceilingChecks != 0 {
		t.Fatalf("ceiling check ran %d times despite a cycle at step 8 (want 0, cycle must reject first)", ceilingChecks)
	}
}
