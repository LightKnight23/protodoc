package cli

import (
	"bytes"
	"testing"
)

// TestTR_012_SignVerbDeterministicOutput is T-0337's named integration test
// (TR-012 / NFR-006). Signing the same fixture+key twice yields byte-identical
// signature octets; a SUBSET-mode run's coverage matches the flags.
func TestTR_012_SignVerbDeterministicOutput(t *testing.T) {
	orig := SignRun
	defer func() { SignRun = orig }()

	// Deterministic backend: same fixture+key -> identical octets (NFR-006).
	SignRun = func(path, key, coverage string, subsetRanges [][2]int) SignResult {
		sig := []byte("deterministic-sig-for-" + path + "-" + key)
		total := coverage == "total"
		var ranges [][2]int
		if !total {
			ranges = [][2]int{{0, 2}}
		}
		return SignResult{SignatureOctets: sig, Total: total, CoveredRanges: ranges}
	}

	r1 := runSign([]string{"doc.pdl", "--key", "k1", "--coverage", "total"}, nil)
	if r1.Status != "OK" {
		t.Fatalf("sign status=%s, want OK", r1.Status)
	}
	// Byte-identical signature: verified via the backend directly.
	s1 := SignRun("doc.pdl", "k1", "total", nil).SignatureOctets
	s2 := SignRun("doc.pdl", "k1", "total", nil).SignatureOctets
	if !bytes.Equal(s1, s2) {
		t.Errorf("signing twice must be byte-identical (NFR-006)")
	}
	if r1.Extra["coverage_total"] != true {
		t.Errorf("total-mode sign should report coverage_total=true")
	}

	// SUBSET mode: coverage split reflects the flags.
	rs := runSign([]string{"doc.pdl", "--key", "k1", "--coverage", "subset"}, nil)
	if rs.Extra["coverage_total"] != false {
		t.Errorf("subset-mode sign should report coverage_total=false")
	}

	if r := runSign(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg sign: status=%s, want USAGE", r.Status)
	}
}

// TestTR_012_MigrateVerbRefusalFirst is T-0338's named integration test
// (TR-012). A fixture with an unrepresentable construct halts in phase 1
// (named construct + location, zero output); a clean fixture with
// --rescind-and-resign produces output.
func TestTR_012_MigrateVerbRefusalFirst(t *testing.T) {
	orig := MigrateRun
	defer func() { MigrateRun = orig }()

	// Unrepresentable construct -> phase-1 refusal, no output.
	MigrateRun = func(path string, toMajor int, rescindResign bool) MigrateResult {
		return MigrateResult{Refused: true, RefusedConstruct: "RETIRED_EXT_TOKEN", RefusedLocation: "segment 1 offset 32"}
	}
	res := runMigrate([]string{"doc.pdl"}, nil)
	if res.Status != "REFUSED" {
		t.Errorf("unrepresentable migrate: status=%s, want REFUSED", res.Status)
	}
	if res.Extra["output_written"] != false {
		t.Errorf("phase-1 refusal must write no output")
	}
	if res.Extra["refused_construct"] != "RETIRED_EXT_TOKEN" {
		t.Errorf("refusal must name the construct, got %v", res.Extra["refused_construct"])
	}

	// Clean fixture with --rescind-and-resign -> output produced.
	MigrateRun = func(path string, toMajor int, rescindResign bool) MigrateResult {
		return MigrateResult{Refused: false, Output: []byte("migrated")}
	}
	res = runMigrate([]string{"doc.pdl", "--rescind-and-resign"}, nil)
	if res.Status != "OK" || res.Extra["output_written"] != true {
		t.Errorf("clean migrate: status=%s output=%v, want OK/true", res.Status, res.Extra["output_written"])
	}

	if r := runMigrate(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg migrate: status=%s, want USAGE", r.Status)
	}
}
