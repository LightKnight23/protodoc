package validate

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_109_ExtensionEnvelopeFallbackCycleVector is T-0119's named
// conformance vector (plan.md Section 10 first-class task g): an
// extension-envelope fallback-reference (ext-fallback-ref) forming a cycle
// -- one of FR-109's 5 edge kinds -- must be rejected by T-0118's detector,
// naming all participating edges. This is authored as its own dedicated
// vector, not folded into general conformance work.
func TestFR_109_ExtensionEnvelopeFallbackCycleVector(t *testing.T) {
	id := func(b byte) pdlfmt.UnitID { var u pdlfmt.UnitID; u[0] = b; return u }

	// Two EXT_ENVELOPE records whose ext-fallback-ref point at each other:
	// env A falls back to env B, env B falls back to env A -- a 2-edge cycle
	// entirely within the ext-envelope-fallback edge kind.
	vector := []Edge{
		{Kind: EdgeExtFallbackRef, From: id(0xE1), To: id(0xE2)},
		{Kind: EdgeExtFallbackRef, From: id(0xE2), To: id(0xE1)},
	}

	err := DetectCycle(vector)
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("ext-envelope fallback cycle not detected: %v", err)
	}
	if len(ce.Edges) != 2 {
		t.Fatalf("named %d edges, want both fallback edges in the cycle", len(ce.Edges))
	}
	for _, e := range ce.Edges {
		if e.Kind != EdgeExtFallbackRef {
			t.Fatalf("cycle edge of kind %q, want ext-envelope-fallback-ref", e.Kind)
		}
	}
	// Both envelopes appear as cycle participants.
	seen := map[pdlfmt.UnitID]bool{}
	for _, e := range ce.Edges {
		seen[e.From] = true
	}
	if !seen[id(0xE1)] || !seen[id(0xE2)] {
		t.Fatalf("cycle does not name both ext-envelope records")
	}

	// A single ext-envelope with a self-fallback is also a cycle.
	selfVec := []Edge{{Kind: EdgeExtFallbackRef, From: id(0xE3), To: id(0xE3)}}
	if err := DetectCycle(selfVec); !errors.As(err, &ce) {
		t.Fatalf("self-referential ext-envelope fallback not detected: %v", err)
	}

	// A non-cyclic fallback chain (A -> B -> C, terminating) passes.
	chain := []Edge{
		{Kind: EdgeExtFallbackRef, From: id(0xE4), To: id(0xE5)},
		{Kind: EdgeExtFallbackRef, From: id(0xE5), To: id(0xE6)},
	}
	if err := DetectCycle(chain); err != nil {
		t.Fatalf("terminating fallback chain rejected: %v", err)
	}
}
