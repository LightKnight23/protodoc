package cli

import (
	"encoding/json"
	"testing"

	"Protodoc/pkg/validate"
)

// TestTR_012_ValidateVerbInvocationShape is T-0327's named integration test
// (TR-012). `protodoc validate` emits cli.md S2's checks/findings shape and
// maps validity verdicts through the exit-code engine: OK for an at-limit
// fixture, INVALID for an over-limit one.
func TestTR_012_ValidateVerbInvocationShape(t *testing.T) {
	// Save and restore the injectable step builder.
	orig := ValidateStepsFor
	defer func() { ValidateStepsFor = orig }()

	// At-limit fixture: all checks pass.
	ValidateStepsFor = func(string) ([]validate.Step, int) {
		return []validate.Step{
			{ID: validate.StepMagicHeader, Run: func() *validate.Finding { return nil }},
			{ID: validate.StepCapabilityArith, Run: func() *validate.Finding { return nil }},
		}, 2
	}
	res := runValidate([]string{"atlimit.pdl"}, nil)
	if res.Status != "OK" || res.ExitCode != 0 {
		t.Errorf("at-limit: status=%s exit=%d, want OK/0", res.Status, res.ExitCode)
	}

	// The envelope carries checks + findings keys per cli.md S2.
	env := Envelope("validate", res)
	if _, ok := env["checks"]; !ok {
		t.Errorf("envelope missing 'checks' key")
	}
	raw, _ := json.Marshal(env)
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("envelope not valid JSON: %v", err)
	}
	if decoded["verb"] != "validate" {
		t.Errorf("verb key = %v, want validate", decoded["verb"])
	}
	if _, ok := decoded["findings"]; !ok {
		t.Errorf("envelope missing 'findings' key")
	}

	// Over-limit fixture: a validity failure -> INVALID.
	ValidateStepsFor = func(string) ([]validate.Step, int) {
		return []validate.Step{
			{ID: validate.StepMagicHeader, Run: func() *validate.Finding { return nil }},
			{ID: validate.StepCapabilityArith, Run: func() *validate.Finding {
				return &validate.Finding{Step: validate.StepCapabilityArith, RuleID: "PD-CEILING-001", Message: "over limit"}
			}},
		}, 2
	}
	res = runValidate([]string{"overlimit.pdl"}, nil)
	if res.Status != "INVALID" || res.ExitCode != 1 {
		t.Errorf("over-limit: status=%s exit=%d, want INVALID/1", res.Status, res.ExitCode)
	}
	if len(res.Findings) != 1 || res.Findings[0].RuleID != "PD-CEILING-001" {
		t.Errorf("over-limit findings = %+v, want one PD-CEILING-001", res.Findings)
	}

	// Missing file argument -> USAGE.
	if r := runValidate(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg: status=%s, want USAGE", r.Status)
	}
}
