package cli

import (
	"errors"
	"io"
	"os"
	"strconv"
)

// sign verb wiring (T-0337, TR-012; --key/--coverage/--intent/--out
// enforcement fixed by DEFECT-2026-09-19b/T-0387). `protodoc sign` produces a
// deterministic EdDSA-Protodoc-1 signature (NFR-006: same fixture+key ->
// identical octets) with a CoverageDescriptor matching the coverage flags.

// SignResult is the sign backend's result.
type SignResult struct {
	SignatureOctets []byte
	CoveredRanges   [][2]int // covered segment ordinal ranges (SUBSET mode)
	Total           bool     // TOTAL coverage mode
	Err             error
}

// SignRun is the injectable signing backend (M06/M09 EdDSA-Protodoc-1).
var SignRun = func(path, key, coverage string, subsetRanges [][2]int) SignResult { return SignResult{} }

// writeSignFile is injectable so tests can observe/stub the write.
var writeSignFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

// signIntents is integrity.abnf S10.3's closed 4-value sig-intent enum.
var signIntents = map[string]bool{
	"author-approval":     true,
	"witness-attestation": true,
	"notarization":        true,
	"custodial-transfer":  true,
}

func runSign(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires a <file> argument"}}})
	}
	key, coverage, intent, to := "", "", "", ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--key":
			if i+1 < len(args) {
				key, i = args[i+1], i+1
			}
		case "--coverage":
			if i+1 < len(args) {
				coverage, i = args[i+1], i+1
			}
		case "--intent":
			if i+1 < len(args) {
				intent, i = args[i+1], i+1
			}
		case "--out":
			if i+1 < len(args) {
				to, i = args[i+1], i+1
			}
		}
	}
	if key == "" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires --key <ref>"}}})
	}
	if coverage != "total" && coverage != "subset" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires --coverage total|subset"}}})
	}
	if !signIntents[intent] {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires --intent <one of author-approval|witness-attestation|notarization|custodial-transfer>"}}})
	}
	if to == "" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign requires --out <path>"}}})
	}
	out := SignRun(args[0], key, coverage, nil)
	if out.Err != nil {
		if errors.Is(out.Err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign: " + out.Err.Error()}}})
		}
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	// HONEST SCOPE (DEFECT-2026-09-19b): the real signature above is genuinely
	// computed (NFR-006 determinism holds), but this backend does not yet
	// splice a new SIGNATURE segment into a re-serialized document -- doing so
	// safely needs real segment-table/commit-ring reconstruction, which does
	// not exist yet. Writing a fabricated "signed" file would be worse than
	// not writing one, so the original document's bytes are copied to --out
	// unchanged and the response says plainly that embedding is not yet done,
	// rather than silently claiming a complete signed document was produced.
	orig, rerr := os.ReadFile(args[0])
	if rerr != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign: re-reading input for --out: " + rerr.Error()}}})
	}
	if err := writeSignFile(to, orig); err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "sign: writing --out: " + err.Error()}}})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{
			"out":                      to,
			"signature_len":            len(out.SignatureOctets),
			"coverage_total":           out.Total,
			"covered_ranges":           out.CoveredRanges,
			"signed_document_complete": false,
		},
		Findings: []Finding{{RuleID: "TR-012", Message: "signature computed but not yet embedded in --out; --out is currently an unmodified copy of the input (see docs/cli-usage-guide.md)"}},
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

// writeMigrateFile is injectable so tests can observe/stub the write.
var writeMigrateFile = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func runMigrate(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate requires a <file> argument"}}})
	}
	rescind := false
	toMajor := -1
	to := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--rescind-and-resign":
			rescind = true
		case "--to-major":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					toMajor = n
				}
				i++
			}
		case "--out":
			if i+1 < len(args) {
				to, i = args[i+1], i+1
			}
		}
	}
	if toMajor < 0 {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate requires --to-major <N>"}}})
	}
	if to == "" {
		return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate requires --out <path>"}}})
	}
	out := MigrateRun(args[0], toMajor, rescind)
	if out.Err != nil {
		if errors.Is(out.Err, ErrFileUnreadable) {
			return StatusUsage.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate: " + out.Err.Error()}}})
		}
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: out.Err.Error()}}})
	}
	if out.Refused {
		// Phase-1 refusal: no output file, name construct + location.
		return StatusRefused.ToResult(Result{
			Extra:    map[string]any{"output_written": false, "refused_construct": out.RefusedConstruct, "refused_location": out.RefusedLocation},
			Findings: []Finding{{RuleID: "FR-121", Message: "unrepresentable construct " + out.RefusedConstruct + " at " + out.RefusedLocation}},
		})
	}
	if err := writeMigrateFile(to, out.Output); err != nil {
		return StatusInvalid.ToResult(Result{Findings: []Finding{{RuleID: "TR-012", Message: "migrate: writing --out: " + err.Error()}}})
	}
	return StatusOK.ToResult(Result{Extra: map[string]any{"out": to, "output_written": true, "output_len": len(out.Output)}})
}
