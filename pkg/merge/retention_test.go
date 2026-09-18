package merge

import (
	"errors"
	"testing"

	"Protodoc/pkg/history"
)

// TestCON_024_RefuseMergeAcrossRetentionPoint is T-0236's named integration
// test (CON-024). It builds a trimmed (retained-from-point) declaration and
// attempts 100 merges of branches carrying changes that predate the retention
// point; every attempt produces a named refusal (RetentionPointError naming
// the retention point) with zero merged outputs. Branches wholly at or after
// the retention point are permitted.
func TestCON_024_RefuseMergeAcrossRetentionPoint(t *testing.T) {
	const retentionPoint = 50
	decl, err := history.NewDeclarationWithRetentionPoint(history.ModeRetainedFromPoint, retentionPoint)
	if err != nil {
		t.Fatalf("build retained-from-point declaration: %v", err)
	}

	mergedOutputs := 0
	for i := 0; i < 100; i++ {
		// Each attempt offers a change predating the retention point.
		preTrimOrdinal := uint16(i % retentionPoint) // 0..49, all < 50
		err := CheckRetentionPoint(decl, []uint16{uint16(retentionPoint) + 5, preTrimOrdinal})
		if !errors.Is(err, ErrMergeAcrossRetentionPoint) {
			t.Fatalf("attempt %d: err = %v, want ErrMergeAcrossRetentionPoint", i, err)
		}
		var re *RetentionPointError
		if !errors.As(err, &re) {
			t.Fatalf("attempt %d: expected *RetentionPointError, got %T", i, err)
		}
		if re.RetentionPoint != retentionPoint {
			t.Errorf("attempt %d: names retention point %d, want %d", i, re.RetentionPoint, retentionPoint)
		}
		if re.OffendingOrdinal >= retentionPoint {
			t.Errorf("attempt %d: offending ordinal %d not pre-trim", i, re.OffendingOrdinal)
		}
		// A refused merge produces no output.
		if err == nil {
			mergedOutputs++
		}
	}
	if mergedOutputs != 0 {
		t.Errorf("100 pre-trim merges produced %d merged outputs, want 0", mergedOutputs)
	}

	// A branch wholly at or after the retention point is permitted.
	if err := CheckRetentionPoint(decl, []uint16{retentionPoint, retentionPoint + 10}); err != nil {
		t.Errorf("post-retention branch should merge, got %v", err)
	}

	// Complete-history and no-history declarations have no retention point.
	for _, m := range []history.Mode{history.ModeComplete, history.ModeNone} {
		d, _ := history.NewDeclaration(m)
		if err := CheckRetentionPoint(d, []uint16{0, 1, 2}); err != nil {
			t.Errorf("mode %s has no retention point to cross, got %v", history.ModeName(m), err)
		}
	}
}
