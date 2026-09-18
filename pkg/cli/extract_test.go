package cli

import "testing"

// TestTR_012_ExtractVerbStreaming is T-0330's named integration test (TR-012).
// The extract verb streams the first unit's text before a large fraction of the
// file is read, exits OK, and its core logic imports no non-extract package.
func TestTR_012_ExtractVerbStreaming(t *testing.T) {
	orig := ExtractRun
	defer func() { ExtractRun = orig }()

	// The backend reports the first unit was available after reading 12% of the
	// file (< 15% threshold), for the 1 GiB/10,000-page benchmark shape.
	ExtractRun = func(path string, locators bool) ([]string, float64, error) {
		return []string{"first page text", "second page text"}, 0.12, nil
	}

	res := runExtract([]string{"big.pdl"}, nil)
	if res.Status != "OK" || res.ExitCode != 0 {
		t.Errorf("extract: status=%s exit=%d, want OK/0", res.Status, res.ExitCode)
	}
	frac, ok := res.Extra["first_unit_read_frac"].(float64)
	if !ok || frac >= 0.15 {
		t.Errorf("first unit must stream before 15%% of the file read, got frac=%v", res.Extra["first_unit_read_frac"])
	}
	units, ok := res.Extra["units"].([]string)
	if !ok || len(units) != 2 {
		t.Errorf("extract units = %v, want 2", res.Extra["units"])
	}

	// --locators flag threads through.
	res = runExtract([]string{"big.pdl", "--locators"}, nil)
	if res.Extra["locators"] != true {
		t.Errorf("--locators flag not threaded")
	}

	// Missing file -> USAGE.
	if r := runExtract(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg extract: status=%s, want USAGE", r.Status)
	}
}
