package validate

import "testing"

// TestFR_103_ErrorPrecedenceReportsEarliestFailureOnly is T-0113's named
// test. Given steps where 3 and 8 both fail, only step 3's verdict is
// returned and step 8 is never evaluated; no reconstruction/partial output
// path exists. Step 9's budget status is reported alongside a validity
// verdict, not instead of it.
func TestFR_103_ErrorPrecedenceReportsEarliestFailureOnly(t *testing.T) {
	// Track which steps were evaluated.
	evaluated := map[StepID]bool{}
	mkStep := func(id StepID, fail bool) Step {
		return Step{ID: id, Run: func() *Finding {
			evaluated[id] = true
			if fail {
				return &Finding{Step: id, RuleID: "TEST", Message: "fail"}
			}
			return nil
		}}
	}

	// Steps 3 and 8 both fail. Provide them out of order to prove the
	// orchestrator sorts by ID.
	steps := []Step{
		mkStep(StepCycleDetection, true),   // 8, fails
		mkStep(StepRingWinner, true),       // 3, fails
		mkStep(StepMagicHeader, false),     // 1, passes
		mkStep(StepCapabilityArith, false), // 2, passes
	}
	res := Run(steps)

	if res.Valid {
		t.Fatalf("document with failing steps reported Valid")
	}
	if res.Validity == nil || res.Validity.Step != StepRingWinner {
		t.Fatalf("verdict = %+v, want earliest failing step 3 (RingWinner)", res.Validity)
	}
	// Step 8 must never have been evaluated.
	if evaluated[StepCycleDetection] {
		t.Fatalf("step 8 was evaluated despite step 3 failing first")
	}
	// Steps 1 and 2 (before 3) were evaluated and passed.
	if !evaluated[StepMagicHeader] || !evaluated[StepCapabilityArith] {
		t.Fatalf("steps before the failing step were not all evaluated")
	}

	// All-pass pipeline yields Valid with no findings.
	evaluated = map[StepID]bool{}
	allPass := []Step{mkStep(StepMagicHeader, false), mkStep(StepRingWinner, false), mkStep(StepExtEnvelope, false)}
	if r := Run(allPass); !r.Valid || r.Validity != nil {
		t.Fatalf("all-pass pipeline: Valid=%v Validity=%v, want valid/nil", r.Valid, r.Validity)
	}

	// Step-9 budget status is reported ALONGSIDE a valid verdict, not
	// instead: a budget finding does not halt the pipeline or flip Valid.
	evaluated = map[StepID]bool{}
	budgetSteps := []Step{
		mkStep(StepMagicHeader, false),
		{ID: StepStructuralCeilings, Run: func() *Finding {
			evaluated[StepStructuralCeilings] = true
			return &Finding{Step: StepStructuralCeilings, RuleID: "PD-BUDGET-001", Message: "over budget", Budget: true}
		}},
		mkStep(StepExtEnvelope, false), // 13, must still run after a budget finding
	}
	rb := Run(budgetSteps)
	if !rb.Valid {
		t.Fatalf("a pure budget finding wrongly made the document invalid")
	}
	if rb.Budget == nil || rb.Budget.RuleID != "PD-BUDGET-001" {
		t.Fatalf("budget finding not reported alongside: %+v", rb.Budget)
	}
	if !evaluated[StepExtEnvelope] {
		t.Fatalf("a budget finding halted the pipeline; step 13 did not run")
	}
}
