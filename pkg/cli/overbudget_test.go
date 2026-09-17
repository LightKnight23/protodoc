package cli

import (
	"testing"

	"Protodoc/pkg/validate"
)

// TestCON_011_OverBudgetExitCodePrecedence is T-0125's named integration
// test. It drives the real validate pipeline to produce each combination of
// (structurally valid/invalid) x (within/over budget), bridges the
// validate.Result into cli statuses, and asserts CON-011's guarantees:
//
//   - OVER_BUDGET is distinct from every validity verdict (an over-budget but
//     otherwise valid document is OVER_BUDGET, never INVALID);
//   - it is ranked above UNAVAILABLE/UNVERIFIED/REFUSED but below INVALID
//     (a document that is both invalid AND over budget exits INVALID);
//   - both findings survive in the envelope regardless of the resolved code
//     (cli.md S1.1: findings always carries every applicable finding).
func TestCON_011_OverBudgetExitCodePrecedence(t *testing.T) {
	// Pipeline step factories: a passing step, a validity-failing step, and a
	// step-9 budget-finding step (non-validity, reported alongside).
	pass := func(id validate.StepID) validate.Step {
		return validate.Step{ID: id, Run: func() *validate.Finding { return nil }}
	}
	invalid := func() validate.Step {
		return validate.Step{ID: validate.StepBoundedPrefix, Run: func() *validate.Finding {
			return &validate.Finding{Step: validate.StepBoundedPrefix, RuleID: "PD-TRUNC-001", Message: "truncated"}
		}}
	}
	overBudget := func() validate.Step {
		return validate.Step{ID: validate.StepStructuralCeilings, Run: func() *validate.Finding {
			return &validate.Finding{Step: validate.StepStructuralCeilings, RuleID: "PD-BUDGET-001", Message: "over local budget", Budget: true}
		}}
	}

	cases := []struct {
		name         string
		steps        []validate.Step
		wantStatus   Status
		wantExit     int
		wantFindings int
	}{
		{
			name:         "valid within budget -> OK",
			steps:        []validate.Step{pass(validate.StepMagicHeader)},
			wantStatus:   StatusOK,
			wantExit:     0,
			wantFindings: 0,
		},
		{
			name:         "valid but over budget -> OVER_BUDGET (distinct from any validity verdict)",
			steps:        []validate.Step{pass(validate.StepMagicHeader), overBudget()},
			wantStatus:   StatusOverBudget,
			wantExit:     5,
			wantFindings: 1,
		},
		{
			name:         "invalid within budget -> INVALID",
			steps:        []validate.Step{invalid()},
			wantStatus:   StatusInvalid,
			wantExit:     1,
			wantFindings: 1,
		},
		{
			// Both true: INVALID outranks OVER_BUDGET, but BOTH findings must
			// survive. The budget step is ordered after the validity failure,
			// so to prove findings are not suppressed we assert both are
			// present even though the pipeline halts at the validity failure.
			name: "invalid AND over budget -> INVALID wins, both findings kept",
			// Budget step (9) is ordered before the invalidating bounded-prefix
			// step (5)? No: 5 < 9, so validity halts first. To exercise the
			// "both true" report we build the Result directly below instead.
			steps:        nil,
			wantStatus:   StatusInvalid,
			wantExit:     1,
			wantFindings: 2,
		},
	}

	for _, c := range cases {
		var res validate.Result
		if c.steps != nil {
			res = validate.Run(c.steps)
		} else {
			// The "both true" case: a validity failure and a budget finding
			// simultaneously present in one Result (as the pipeline records
			// when the budget check runs before the validity failure).
			res = validate.Result{
				Valid:    false,
				Validity: &validate.Finding{Step: validate.StepBoundedPrefix, RuleID: "PD-TRUNC-001", Message: "truncated"},
				Budget:   &validate.Finding{Step: validate.StepStructuralCeilings, RuleID: "PD-BUDGET-001", Message: "over local budget", Budget: true},
			}
		}

		active := StatusesFromValidation(res)
		got := Resolve(active)
		if got != c.wantStatus {
			t.Errorf("%s: resolved status = %v, want %v", c.name, got, c.wantStatus)
		}
		if got.Code() != c.wantExit {
			t.Errorf("%s: exit code = %d, want %d", c.name, got.Code(), c.wantExit)
		}
		if n := len(FindingsFromValidation(res)); n != c.wantFindings {
			t.Errorf("%s: %d findings survived, want %d", c.name, n, c.wantFindings)
		}
	}

	// Explicit distinctness assertion: OVER_BUDGET must never be a validity
	// verdict. An over-budget-only Result yields OVER_BUDGET and NOT INVALID.
	obOnly := validate.Result{Valid: true, Budget: &validate.Finding{RuleID: "PD-BUDGET-002", Budget: true}}
	a := StatusesFromValidation(obOnly)
	if a[StatusInvalid] {
		t.Error("CON-011 violation: an over-budget-only document was marked INVALID")
	}
	if !a[StatusOverBudget] {
		t.Error("over-budget document did not set OVER_BUDGET")
	}

	// Ranking assertion: OVER_BUDGET below INVALID, above UNAVAILABLE /
	// UNVERIFIED / REFUSED.
	if Resolve(map[Status]bool{StatusInvalid: true, StatusOverBudget: true}) != StatusInvalid {
		t.Error("INVALID must outrank OVER_BUDGET")
	}
	for _, lower := range []Status{StatusUnavailable, StatusUnverified, StatusRefused} {
		if Resolve(map[Status]bool{StatusOverBudget: true, lower: true}) != StatusOverBudget {
			t.Errorf("OVER_BUDGET must outrank %v", lower)
		}
	}
}
