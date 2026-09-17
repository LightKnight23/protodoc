package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
	"Protodoc/pkg/validate"
)

// TestFR_015_FR_109_ExtFallbackCycleRejected is T-0064's named conformance
// test (FR-015 / FR-109). Extension-envelope fallback references are one of
// the reference-graph's five edge kinds; a cycle formed through
// ext-envelope-fallback-ref edges (envelope A degrades to a unit that
// degrades back to A, directly or transitively) is rejected by the same
// cycle detector that rejects any reference-graph cycle, naming the
// participating edges. This closes the case where a fallback chain never
// terminates in real core-feature-set content.
func TestFR_015_FR_109_ExtFallbackCycleRejected(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID { return fixtureUnitID(b) }
	fb := func(from, to byte) validate.Edge {
		return validate.Edge{Kind: validate.EdgeExtFallbackRef, From: u(from), To: u(to)}
	}

	// Acyclic corpus: a straight fallback chain A -> B -> C terminating.
	acyclic := []validate.Edge{fb(0x01, 0x02), fb(0x02, 0x03)}
	if err := validate.DetectCycle(acyclic); err != nil {
		t.Fatalf("acyclic fallback chain rejected: %v", err)
	}

	// Cyclic corpus 1: a direct 2-cycle A -> B -> A via fallback edges.
	twoCycle := []validate.Edge{fb(0x01, 0x02), fb(0x02, 0x01)}
	err := validate.DetectCycle(twoCycle)
	if err == nil {
		t.Fatal("a 2-node fallback cycle was accepted")
	}
	var ce *validate.CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *validate.CycleError, got %T", err)
	}
	for _, e := range ce.Edges {
		if e.Kind != validate.EdgeExtFallbackRef {
			t.Errorf("cycle edge of unexpected kind %q", e.Kind)
		}
	}

	// Cyclic corpus 2: a self-loop A -> A (a fallback to itself).
	if err := validate.DetectCycle([]validate.Edge{fb(0x05, 0x05)}); err == nil {
		t.Fatal("a self-loop fallback cycle was accepted")
	}

	// Cyclic corpus 3: a longer transitive cycle A -> B -> C -> A.
	threeCycle := []validate.Edge{fb(0x01, 0x02), fb(0x02, 0x03), fb(0x03, 0x01)}
	if err := validate.DetectCycle(threeCycle); err == nil {
		t.Fatal("a 3-node transitive fallback cycle was accepted")
	}
}
