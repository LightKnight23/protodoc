package cli

import "io"

// sign verb wiring (T-0337, TR-012). `protodoc sign` produces a deterministic
// EdDSA-Protodoc-1 signature (NFR-006: same fixture+key -> identical octets)
// with a CoverageDescriptor matching the coverage flags.

// SignResult is the sign backend's result.
type SignResult struct {
	SignatureOctets []byte
	CoveredRanges   [][2]int // covered segment ordinal ranges (SUBSET mode)
	Total           bool     // TOTAL coverage mode
	Err             error
}

// SignRun is the injectable signing backend (M06/M09 EdDSA-Protodoc-1).
var SignRun = func(path, key, coverage string, subsetRanges [][2]int) SignResult { return SignResult{} }

func runSign(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires a <file> argument"}}})
	}
	key, coverage := "", "total"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--key":
			if i+1 < len(args) {
				key = args[i+1]
			}
		case "--coverage":
			if i+1 < len(args) {
				coverage = args[i+1]
			}
		}
	}
	out := SignRun(args[0], key, coverage, nil)
	if out.Err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{
			"signature_len":  len(out.SignatureOctets),
			"coverage_total": out.Total,
			"covered_ranges": out.CoveredRanges,
		},
	})
}

// migrate verb wiring (T-0338, TR-012). `protodoc migrate` runs M16's
// refusal-first two-phase migration: phase 1 (Scan) halts on an unrepresentable
// construct with no output; phase 2 (Transform) emits the migrated file, and
// --rescind-and-resign preserves the pre-migration signature's coverage report.

// MigrateResult is the migrate backend's result.
type MigrateResult struct {
	Refused          bool
	RefusedConstruct string
	RefusedLocation  string
	Output           []byte
	Err              error
}

// MigrateRun is the injectable migration backend (M16).
var MigrateRun = func(path string, toMajor int, rescindResign bool) MigrateResult { return MigrateResult{} }

func runMigrate(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate requires a <file> argument"}}})
	}
	rescind := false
	for _, a := range args[1:] {
		if a == "--rescind-and-resign" {
			rescind = true
		}
	}
	out := MigrateRun(args[0], 0, rescind)
	if out.Err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	if out.Refused {
		// Phase-1 refusal: no output file, name construct + location.
		return StatusRefused.ToResult(Result{
			Extra:    map[string]any{"output_written": false, "refused_construct": out.RefusedConstruct, "refused_location": out.RefusedLocation},
			Findings: []Finding{{RuleID: "FR-121", Message: "unrepresentable construct " + out.RefusedConstruct + " at " + out.RefusedLocation}},
		})
	}
	return StatusOK.ToResult(Result{Extra: map[string]any{"output_written": true, "output_len": len(out.Output)}})
}
