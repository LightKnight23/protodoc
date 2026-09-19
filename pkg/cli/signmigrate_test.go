package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestTR_012_SignVerbDeterministicOutput is T-0337's named integration test
// (TR-012 / NFR-006, updated by T-0387 for --key/--coverage/--intent/--out
// enforcement). Signing the same fixture+key twice yields byte-identical
// signature octets; a SUBSET-mode run's coverage matches the flags.
func TestTR_012_SignVerbDeterministicOutput(t *testing.T) {
	orig := SignRun
	defer func() { SignRun = orig }()

	doc := filepath.Join(t.TempDir(), "doc.pdl")
	if err := os.WriteFile(doc, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := func() string { return filepath.Join(t.TempDir(), "signed.pdl") }

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

	r1 := runSign([]string{doc, "--key", "k1", "--coverage", "total", "--intent", "author-approval", "--out", outPath()}, nil)
	if r1.Status != "OK" {
		t.Fatalf("sign status=%s (%+v), want OK", r1.Status, r1.Findings)
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
	rs := runSign([]string{doc, "--key", "k1", "--coverage", "subset", "--intent", "author-approval", "--out", outPath()}, nil)
	if rs.Extra["coverage_total"] != false {
		t.Errorf("subset-mode sign should report coverage_total=false")
	}

	if r := runSign(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg sign: status=%s, want USAGE", r.Status)
	}
	if r := runSign([]string{doc, "--coverage", "total", "--intent", "author-approval", "--out", outPath()}, nil); r.Status != "USAGE" {
		t.Errorf("missing --key: status=%s, want USAGE", r.Status)
	}
	if r := runSign([]string{doc, "--key", "k1", "--intent", "author-approval", "--out", outPath()}, nil); r.Status != "USAGE" {
		t.Errorf("missing --coverage: status=%s, want USAGE", r.Status)
	}
	if r := runSign([]string{doc, "--key", "k1", "--coverage", "total", "--out", outPath()}, nil); r.Status != "USAGE" {
		t.Errorf("missing --intent: status=%s, want USAGE", r.Status)
	}
	if r := runSign([]string{doc, "--key", "k1", "--coverage", "total", "--intent", "author-approval"}, nil); r.Status != "USAGE" {
		t.Errorf("missing --out: status=%s, want USAGE", r.Status)
	}
}

// TestTR_012_MigrateVerbRefusalFirst is T-0338's named integration test
// (TR-012, updated by T-0388 for --to-major/--out enforcement and file-write).
// A fixture with an unrepresentable construct halts in phase 1 (named
// construct + location, zero output); a clean fixture with
// --rescind-and-resign produces output actually written to --out.
func TestTR_012_MigrateVerbRefusalFirst(t *testing.T) {
	orig := MigrateRun
	defer func() { MigrateRun = orig }()
	outPath := func() string { return filepath.Join(t.TempDir(), "migrated.pdl") }

	// Unrepresentable construct -> phase-1 refusal, no output.
	MigrateRun = func(path string, toMajor int, rescindResign bool) MigrateResult {
		return MigrateResult{Refused: true, RefusedConstruct: "RETIRED_EXT_TOKEN", RefusedLocation: "segment 1 offset 32"}
	}
	res := runMigrate([]string{"doc.pdl", "--to-major", "2", "--out", outPath()}, nil)
	if res.Status != "REFUSED" {
		t.Errorf("unrepresentable migrate: status=%s, want REFUSED", res.Status)
	}
	if res.Extra["output_written"] != false {
		t.Errorf("phase-1 refusal must write no output")
	}
	if res.Extra["refused_construct"] != "RETIRED_EXT_TOKEN" {
		t.Errorf("refusal must name the construct, got %v", res.Extra["refused_construct"])
	}

	// Clean fixture with --rescind-and-resign -> output produced and written.
	MigrateRun = func(path string, toMajor int, rescindResign bool) MigrateResult {
		return MigrateResult{Refused: false, Output: []byte("migrated")}
	}
	out2 := outPath()
	res = runMigrate([]string{"doc.pdl", "--to-major", "2", "--rescind-and-resign", "--out", out2}, nil)
	if res.Status != "OK" || res.Extra["output_written"] != true {
		t.Errorf("clean migrate: status=%s output=%v, want OK/true", res.Status, res.Extra["output_written"])
	}
	written, err := os.ReadFile(out2)
	if err != nil || string(written) != "migrated" {
		t.Errorf("migrate reported OK but --out file wrong/missing: content=%q err=%v", written, err)
	}

	if r := runMigrate(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg migrate: status=%s, want USAGE", r.Status)
	}
	if r := runMigrate([]string{"doc.pdl", "--out", outPath()}, nil); r.Status != "USAGE" {
		t.Errorf("missing --to-major: status=%s, want USAGE", r.Status)
	}
	if r := runMigrate([]string{"doc.pdl", "--to-major", "2"}, nil); r.Status != "USAGE" {
		t.Errorf("missing --out: status=%s, want USAGE", r.Status)
	}
}
