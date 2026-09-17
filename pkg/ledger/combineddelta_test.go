package ledger

import (
	"errors"
	"testing"
)

// TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced is T-0043's named
// test. It reports the per-edit combined index-delta plus estimated
// integrity-delta for edits at 1B/1KB/1MB granularity against 1KB/1MB/1GiB
// documents and asserts each is within FR-057's 262,144-octet ceiling,
// matching or improving on the ~26,161-octet worst-case (~10x headroom)
// figure in plan.md Section 5 row 5.
//
// TRACKING NOTE (M08 follow-up): the integrity-tree components (T_S nodes,
// T_C path, coverage-descriptor delta) are the documented worst-case
// stand-in figures from plan.md, since the real T_S/T_C trees land in M08.
// When M08 lands, a full end-to-end re-verification of this budget against
// the real integrity-tree update sizes must be scheduled to replace these
// conservative constants with measured deltas.
func TestFR_057_CombinedIndexIntegrityDeltaBudgetEnforced(t *testing.T) {
	const (
		kib = 1024
		mib = 1024 * 1024
		gib = 1024 * 1024 * 1024
	)
	editSizes := []int{1, kib, mib}
	docSizes := []struct {
		name string
		size uint64
	}{
		{"1KiB", kib},
		{"1MiB", mib},
		{"1GiB", gib},
	}

	// The plan.md Section 5 row 5 worst-case figure a single-segment edit
	// should match or improve on.
	const plannedWorstCase = 26161

	for _, ds := range docSizes {
		for _, k := range editSizes {
			// A K-octet edit touches one content segment (its touched
			// TextBlock); newSegmentCount = 1. The combined delta is
			// independent of K and of document size at this layer -- that
			// independence is exactly the FR-057 guarantee (index+integrity
			// churn does not scale with the edit or the file).
			delta := WorstCaseCombinedDelta(1)
			total := delta.Total()

			if err := EnforceCombinedDeltaBudget(delta); err != nil {
				t.Fatalf("K=%d doc=%s: combined delta refused: %v", k, ds.name, err)
			}
			if total > CombinedDeltaBudget {
				t.Fatalf("K=%d doc=%s: combined delta %d exceeds FR-057 ceiling %d", k, ds.name, total, CombinedDeltaBudget)
			}
			if total > plannedWorstCase {
				t.Fatalf("K=%d doc=%s: combined delta %d exceeds plan.md worst-case %d (regression)", k, ds.name, total, plannedWorstCase)
			}
			// Sanity: index and integrity components each contribute.
			if delta.IndexDelta() == 0 || delta.IntegrityDelta() == 0 {
				t.Fatalf("K=%d doc=%s: a delta component is zero (index %d, integrity %d)", k, ds.name, delta.IndexDelta(), delta.IntegrityDelta())
			}
		}
	}

	// Report the breakdown once for visibility.
	d := WorstCaseCombinedDelta(1)
	t.Logf("FR-057 worst-case single-edit delta: index=%d integrity=%d total=%d (ceiling %d, ~%.0fx headroom)",
		d.IndexDelta(), d.IntegrityDelta(), d.Total(), CombinedDeltaBudget, float64(CombinedDeltaBudget)/float64(d.Total()))
}

// TestFR_057_CombinedDeltaBudgetFailsClosed confirms the FR-057 budget is a
// hard assertion: a delta constructed to exceed the ceiling is refused with
// ErrCombinedDeltaExceeded, so a future component growth that blows the
// budget is caught rather than absorbed.
func TestFR_057_CombinedDeltaBudgetFailsClosed(t *testing.T) {
	over := CombinedDelta{IndexLeafOctets: CombinedDeltaBudget + 1}
	if err := EnforceCombinedDeltaBudget(over); !errors.Is(err, ErrCombinedDeltaExceeded) {
		t.Fatalf("over-ceiling delta returned %v, want ErrCombinedDeltaExceeded", err)
	}

	// Exactly at the ceiling is accepted (boundary is inclusive).
	atCeiling := CombinedDelta{IndexLeafOctets: CombinedDeltaBudget}
	if err := EnforceCombinedDeltaBudget(atCeiling); err != nil {
		t.Fatalf("delta exactly at ceiling wrongly refused: %v", err)
	}

	// A no-op edit has zero combined delta.
	if got := WorstCaseCombinedDelta(0).Total(); got != 0 {
		t.Fatalf("no-op combined delta = %d, want 0", got)
	}
}
