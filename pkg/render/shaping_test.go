package render

import (
	"reflect"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestNFR_021_ShapingDeterministicGivenPinnedOracle is T-0252's named unit test
// (NFR-021). Given a fixed input tuple and pinned oracle version, shaping
// output is byte-identical across repeated invocations; changing any element of
// the tuple (font digest, axes, scalars, features, language) or the pinned
// profile changes the output; and feature-set order does not matter (fixed
// application order).
func TestNFR_021_ShapingDeterministicGivenPinnedOracle(t *testing.T) {
	const profile ShapingProfileID = 1
	in := ShapingInput{
		FontDigest:    pdlfmt.Digest256{0x11, 0x22},
		VariationAxes: []int32{400, -100},
		Scalars:       []rune("fi office"), // includes a ligature-prone pair
		Features:      []string{"liga", "kern"},
		Language:      "en-US",
	}

	// Repeated invocations are identical.
	a := Shape(profile, in)
	b := Shape(profile, in)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("shaping not deterministic across repeated invocations")
	}
	if len(a) != len([]rune(string(in.Scalars))) {
		t.Fatalf("shaped %d glyphs, want %d", len(a), len(in.Scalars))
	}

	// Feature-set ORDER does not matter (fixed application order).
	reordered := in
	reordered.Features = []string{"kern", "liga"}
	if !reflect.DeepEqual(Shape(profile, reordered), a) {
		t.Error("feature-set order must not change shaping output")
	}

	// Each tuple element is a real input: changing it changes the output.
	changeChecks := []struct {
		name string
		mod  func(*ShapingInput)
	}{
		{"font digest", func(x *ShapingInput) { x.FontDigest[0] ^= 0xFF }},
		{"variation axes", func(x *ShapingInput) { x.VariationAxes = []int32{401, -100} }},
		{"scalars", func(x *ShapingInput) { x.Scalars = []rune("fj office") }},
		{"features", func(x *ShapingInput) { x.Features = []string{"liga"} }},
		{"language", func(x *ShapingInput) { x.Language = "de-DE" }},
	}
	for _, c := range changeChecks {
		mod := in
		mod.FontDigest = in.FontDigest // copy array
		mod.VariationAxes = append([]int32(nil), in.VariationAxes...)
		mod.Scalars = append([]rune(nil), in.Scalars...)
		mod.Features = append([]string(nil), in.Features...)
		c.mod(&mod)
		if reflect.DeepEqual(Shape(profile, mod), a) {
			t.Errorf("changing %s must change shaping output", c.name)
		}
	}

	// A different pinned oracle version yields different output.
	if reflect.DeepEqual(Shape(profile+1, in), a) {
		t.Error("changing the pinned oracle version must change shaping output")
	}
}
