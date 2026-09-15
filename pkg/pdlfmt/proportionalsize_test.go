package pdlfmt

import (
	"reflect"
	"testing"
)

// TestCON_013_ProportionalSizeRoundingRemainderToFinalShare is T-0028's
// named test. Implements: CON-013.
func TestCON_013_ProportionalSizeRoundingRemainderToFinalShare(t *testing.T) {
	cases := []struct {
		name    string
		total   GeometricValue
		weights []uint64
		want    []GeometricValue
	}{
		// Equal three-way split of a total not divisible by 3 (CON-013's
		// own worked example): the single unit of remainder lands entirely
		// on the final share.
		{"three-way, remainder on final share", 10, []uint64{1, 1, 1}, []GeometricValue{3, 3, 4}},
		// Equal seven-way split (spec.md CON-013 Verify clause names
		// "three-way and seven-way splits" explicitly): 100 does not
		// divide by 7, so the accumulated remainder crosses an integer
		// boundary twice, at share index 4 and at the final share.
		{"seven-way, exact and inexact", 100, []uint64{1, 1, 1, 1, 1, 1, 1}, []GeometricValue{14, 14, 14, 15, 14, 14, 15}},
		// Seven-way split that divides evenly: no rounding needed at all.
		{"seven-way, evenly divisible", 7, []uint64{1, 1, 1, 1, 1, 1, 1}, []GeometricValue{1, 1, 1, 1, 1, 1, 1}},
		// Unequal weights.
		{"unequal weights", 10, []uint64{1, 1, 2}, []GeometricValue{2, 3, 5}},
		// Zero total.
		{"zero total", 0, []uint64{1, 1, 1}, []GeometricValue{0, 0, 0}},
		// Single column always takes the whole total.
		{"single column", 5, []uint64{1}, []GeometricValue{5}},
		// Realistic geometric-unit magnitude (11 inches, three equal columns).
		{"geometric magnitude", 11 * UnitsPerInch, []uint64{1, 1, 1}, []GeometricValue{3352800, 3352800, 3352800}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ProportionalShares(c.total, c.weights)
			if err != nil {
				t.Fatalf("ProportionalShares(%d, %v): %v", c.total, c.weights, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("ProportionalShares(%d, %v) = %v, want %v", c.total, c.weights, got, c.want)
			}

			var sum GeometricValue
			for _, s := range got {
				sum += s
			}
			if sum != c.total {
				t.Fatalf("shares %v sum to %d, want exactly total %d", got, sum, c.total)
			}
		})
	}
}

// TestCON_013_ProportionalSharesRejectsInvalidInput covers the caller-error
// cases ProportionalShares refuses rather than silently misresolving.
func TestCON_013_ProportionalSharesRejectsInvalidInput(t *testing.T) {
	if _, err := ProportionalShares(-1, []uint64{1}); err == nil {
		t.Error("ProportionalShares(-1, ...) succeeded, want error for a negative total")
	}
	if _, err := ProportionalShares(10, nil); err == nil {
		t.Error("ProportionalShares(10, nil) succeeded, want error for zero weights")
	}
	if _, err := ProportionalShares(10, []uint64{1, 0, 1}); err == nil {
		t.Error("ProportionalShares with a zero weight succeeded, want error")
	}
}
