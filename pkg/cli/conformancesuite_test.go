package cli

import (
	"testing"

	"Protodoc/pkg/validate"
)

// TestTR_012_CLIConformanceSuite is T-0340's named conformance test (TR-012).
// It exercises every verb against its documented, fixture-reachable exit codes
// through the injectable backends, asserting each resolves to the expected
// Status. (The dispatch/envelope shape is covered by the dispatch tests; this
// suite is the per-verb exit-code matrix.)
func TestTR_012_CLIConformanceSuite(t *testing.T) {
	// Save/restore every injectable backend.
	oVal, oExt, oVer, oDif, oMer, oRed, oPub, oSig, oMig :=
		ValidateStepsFor, ExtractRun, VerifyRun, DiffRun, MergeRun, RedactRun, PublishRun, SignRun, MigrateRun
	defer func() {
		ValidateStepsFor, ExtractRun, VerifyRun, DiffRun, MergeRun, RedactRun, PublishRun, SignRun, MigrateRun =
			oVal, oExt, oVer, oDif, oMer, oRed, oPub, oSig, oMig
	}()

	check := func(name, want, got string) {
		if got != want {
			t.Errorf("%s: status=%s, want %s", name, got, want)
		}
	}

	// validate: OK (all pass), INVALID (validity failure), USAGE (no arg).
	ValidateStepsFor = func(string) ([]validate.Step, int) {
		return []validate.Step{{ID: validate.StepMagicHeader, Run: func() *validate.Finding { return nil }}}, 1
	}
	check("validate/ok", "OK", runValidate([]string{"f"}, nil).Status)
	ValidateStepsFor = func(string) ([]validate.Step, int) {
		return []validate.Step{{ID: validate.StepMagicHeader, Run: func() *validate.Finding {
			return &validate.Finding{Step: validate.StepMagicHeader, RuleID: "PD-HEADER-001", Message: "bad"}
		}}}, 1
	}
	check("validate/invalid", "INVALID", runValidate([]string{"f"}, nil).Status)
	check("validate/usage", "USAGE", runValidate(nil, nil).Status)

	// extract: OK and USAGE.
	ExtractRun = func(string, bool) ([]string, float64, error) { return []string{"x"}, 0.1, nil }
	check("extract/ok", "OK", runExtract([]string{"f"}, nil).Status)
	check("extract/usage", "USAGE", runExtract(nil, nil).Status)

	// verify: OK, UNVERIFIED, UNAVAILABLE, USAGE.
	VerifyRun = func(string) []VerifyVerdict { return []VerifyVerdict{{Verdict: "valid", Signer: "s"}} }
	check("verify/ok", "OK", runVerify([]string{"f"}, nil).Status)
	VerifyRun = func(string) []VerifyVerdict { return []VerifyVerdict{{Verdict: "unverified"}} }
	check("verify/unverified", "UNVERIFIED", runVerify([]string{"f"}, nil).Status)
	VerifyRun = func(string) []VerifyVerdict { return []VerifyVerdict{{Verdict: "covering_unavailable_state"}} }
	check("verify/unavailable", "UNAVAILABLE", runVerify([]string{"f"}, nil).Status)
	check("verify/usage", "USAGE", runVerify(nil, nil).Status)

	// diff: OK and USAGE.
	DiffRun = func(string, string) ([]string, error) { return nil, nil }
	check("diff/ok", "OK", runDiff([]string{"a", "b"}, nil).Status)
	check("diff/usage", "USAGE", runDiff([]string{"a"}, nil).Status)

	// merge: OK, INVALID (conflict), REFUSED, USAGE.
	MergeRun = func(string, string, string) MergeOutcome { return MergeOutcome{Kind: MergeClean} }
	check("merge/ok", "OK", runMerge([]string{"a", "b", "c"}, nil).Status)
	MergeRun = func(string, string, string) MergeOutcome {
		return MergeOutcome{Kind: MergeConflict, ValueA: "x", ValueB: "y"}
	}
	check("merge/conflict", "INVALID", runMerge([]string{"a", "b", "c"}, nil).Status)
	MergeRun = func(string, string, string) MergeOutcome {
		return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CON-024"}
	}
	check("merge/refused", "REFUSED", runMerge([]string{"a", "b", "c"}, nil).Status)
	check("merge/usage", "USAGE", runMerge([]string{"a"}, nil).Status)

	// project: OK and USAGE. project now reads real files (T-0376), so give it a
	// genuine valid prefix rather than a nonexistent path.
	projValid := writeValidPrefix(t)
	check("project/ok", "OK", runProject([]string{projValid, "--format=text"}, nil).Status)
	check("project/usage", "USAGE", runProject(nil, nil).Status)

	// redact: OK and USAGE.
	RedactRun = func(string, []string) RedactResult { return RedactResult{} }
	check("redact/ok", "OK", runRedact([]string{"f", "--subtree", "u"}, nil).Status)
	check("redact/usage", "USAGE", runRedact(nil, nil).Status)

	// publish: OK, INVALID (residue), USAGE.
	PublishRun = func(string, bool) PublishResult { return PublishResult{CustodyPreserved: true} }
	check("publish/ok", "OK", runPublish([]string{"f"}, nil).Status)
	PublishRun = func(string, bool) PublishResult { return PublishResult{ResidueOctets: 1, CustodyPreserved: true} }
	check("publish/invalid", "INVALID", runPublish([]string{"f"}, nil).Status)
	check("publish/usage", "USAGE", runPublish(nil, nil).Status)

	// sign: OK and USAGE.
	SignRun = func(string, string, string, [][2]int) SignResult { return SignResult{} }
	check("sign/ok", "OK", runSign([]string{"f", "--key", "k"}, nil).Status)
	check("sign/usage", "USAGE", runSign(nil, nil).Status)

	// migrate: OK, REFUSED (phase-1), USAGE.
	MigrateRun = func(string, int, bool) MigrateResult { return MigrateResult{} }
	check("migrate/ok", "OK", runMigrate([]string{"f"}, nil).Status)
	MigrateRun = func(string, int, bool) MigrateResult {
		return MigrateResult{Refused: true, RefusedConstruct: "X", RefusedLocation: "seg 0"}
	}
	check("migrate/refused", "REFUSED", runMigrate([]string{"f"}, nil).Status)
	check("migrate/usage", "USAGE", runMigrate(nil, nil).Status)

	// inspect + all 11 verbs registered.
	if len(Names()) != 11 {
		t.Errorf("expected 11 registered verbs, got %d", len(Names()))
	}
	if _, ok := Lookup("inspect"); !ok {
		t.Errorf("inspect verb must be registered")
	}
}
