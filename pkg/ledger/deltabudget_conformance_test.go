package ledger

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

// deltaBudgetVector is one FR-057 combined-delta conformance fixture: a
// CombinedDelta sized to a target total, with the verdict a conforming
// implementation must produce for it.
type deltaBudgetVector struct {
	name         string
	delta        CombinedDelta
	total        uint64
	wantCommit   bool // true: the commit is accepted; false: refused naming the ceiling
	verdictLabel string
}

// deltaBudgetBoundaryCorpus returns the CON-010/CP-011 at-limit and
// one-past-limit conformance pair for FR-057's 262,144-octet combined
// index+integrity delta ceiling: one fixture whose delta totals exactly
// 262,144 (accepted) and a sibling totalling 262,145 (refused naming the
// ceiling). Both are built from real CombinedDelta component fields so the
// vector exercises the actual accounting, not a placeholder.
func deltaBudgetBoundaryCorpus() []deltaBudgetVector {
	// Build a delta totalling exactly the ceiling: put the whole budget in
	// the index-leaf component, with the other components zero, so Total()
	// == CombinedDeltaBudget exactly.
	atLimit := CombinedDelta{IndexLeafOctets: CombinedDeltaBudget}
	overLimit := CombinedDelta{IndexLeafOctets: CombinedDeltaBudget + 1}

	return []deltaBudgetVector{
		{
			name:         "CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY/at-limit",
			delta:        atLimit,
			total:        CombinedDeltaBudget,
			wantCommit:   true,
			verdictLabel: "commit",
		},
		{
			name:         "CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY/over-limit",
			delta:        overLimit,
			total:        CombinedDeltaBudget + 1,
			wantCommit:   false,
			verdictLabel: "refused-naming-ceiling",
		},
	}
}

// TestCONF_LEDGER_DELTA_BUDGET_262144_BOUNDARY is T-0048's named
// conformance test (vector id CONF-LEDGER-DELTA-BUDGET-262144-BOUNDARY). It
// runs the at-limit and one-past-limit FR-057 combined-delta fixtures
// through the delta-budget enforcement and asserts each produces its
// recorded verdict: the 262,144-octet delta commits, and the 262,145-octet
// delta is refused with an error naming the ceiling (CON-010/CP-011: the
// at-limit and over-limit fixtures ship alongside the mechanism).
func TestCONF_LEDGER_DELTA_BUDGET_262144_BOUNDARY(t *testing.T) {
	for _, v := range deltaBudgetBoundaryCorpus() {
		v := v
		t.Run(v.name, func(t *testing.T) {
			// The fixture's Total must equal its recorded target, proving
			// the vector really sits at the boundary it claims.
			if got := v.delta.Total(); got != v.total {
				t.Fatalf("fixture %s: delta total %d, recorded %d", v.name, got, v.total)
			}

			err := EnforceCombinedDeltaBudget(v.delta)
			if v.wantCommit {
				if err != nil {
					t.Fatalf("%s (total %d): expected verdict %q, got refusal %v", v.name, v.total, v.verdictLabel, err)
				}
				return
			}

			// Over-limit: must be refused, naming the ceiling.
			if err == nil {
				t.Fatalf("%s (total %d): expected verdict %q, but the commit was accepted", v.name, v.total, v.verdictLabel)
			}
			if !errors.Is(err, ErrCombinedDeltaExceeded) {
				t.Fatalf("%s: refusal %v does not match ErrCombinedDeltaExceeded", v.name, err)
			}
			// The refusal message must name the 262,144 ceiling.
			if !strings.Contains(err.Error(), strconv.FormatUint(CombinedDeltaBudget, 10)) {
				t.Fatalf("%s: refusal %q does not name the ceiling %d", v.name, err.Error(), CombinedDeltaBudget)
			}
		})
	}
}
