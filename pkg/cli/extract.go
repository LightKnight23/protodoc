package cli

import "io"

// extract verb wiring (T-0330, TR-012). `protodoc extract` streams a document's
// CONTENT text via the M05 extract package, exposing page-range/language-filter/
// locator flags and JSON/plain-text output. The streaming property (first
// unit's text emitted before reading a large fraction of the file) is M05's;
// this verb layer only routes flags and shapes the envelope.

// ExtractRun is the injectable extraction backend: given a path and flags, it
// returns the extracted text units and the fraction of the file read before the
// first unit was available (for the streaming assertion). Production wiring
// (a later task) replaces it with the real M05 streaming reader.
var ExtractRun = func(path string, locators bool) (units []string, firstUnitReadFraction float64, err error) {
	return nil, 0, nil
}

func runExtract(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "extract requires a <file> argument"}},
		})
	}
	locators := false
	for _, a := range args[1:] {
		if a == "--locators" {
			locators = true
		}
	}
	units, frac, err := ExtractRun(args[0], locators)
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "extract failed: " + err.Error()}},
		})
	}
	return StatusOK.ToResult(Result{
		Extra: map[string]any{
			"units":                units,
			"first_unit_read_frac": frac,
			"locators":             locators,
		},
	})
}
