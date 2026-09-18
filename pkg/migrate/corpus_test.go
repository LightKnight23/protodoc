package migrate

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

func corpUnit(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = seed
	return id
}

// corpusPairs is the source construct list for each published corpus pair; an
// independent implementation reconstructs its source inputs from this.
func corpusPairs() map[string][]SourceConstruct {
	return map[string][]SourceConstruct{
		"pair1_single_textblock": {
			{Kind: 0x01, Location: Location{SegmentOrdinal: 0, IntraOffset: 0, UnitID: corpUnit(0x11), HasUnitID: true}},
		},
		"pair2_three_ordered": {
			{Kind: 0x0E, Location: Location{SegmentOrdinal: 0, IntraOffset: 200, UnitID: corpUnit(0x33), HasUnitID: true}},
			{Kind: 0x01, Location: Location{SegmentOrdinal: 0, IntraOffset: 0, UnitID: corpUnit(0x11), HasUnitID: true}},
			{Kind: 0x02, Location: Location{SegmentOrdinal: 0, IntraOffset: 100, UnitID: corpUnit(0x22), HasUnitID: true}},
		},
		"pair3_no_unitid": {
			{Kind: 0x01, Location: Location{SegmentOrdinal: 1, IntraOffset: 0}},
			{Kind: 0x02, Location: Location{SegmentOrdinal: 2, IntraOffset: 0, UnitID: corpUnit(0x44), HasUnitID: true}},
		},
	}
}

// TestFR_119_ExternalTrialCorpusPublished is T-0299's named external-trial test
// (FR-119 / CP-003 / NFR-028). It verifies the published golden corpus ships
// >=3 pairs whose expected canonical octets match Transform's output, so an
// independent implementation can consume it. It does NOT itself run the
// two-implementation gate (M19 scope; the corpus README records that as OPEN).
func TestFR_119_ExternalTrialCorpusPublished(t *testing.T) {
	target := TargetProfile{Major: 2, Representable: map[ConstructKind]bool{0x01: true, 0x02: true, 0x0E: true}}
	pairs := corpusPairs()

	manifestPath := filepath.Join("testdata", "migration", "manifest.tsv")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 3 {
		t.Fatalf("corpus must ship >=3 pairs, got %d", len(lines))
	}

	for _, line := range lines {
		cols := strings.Split(line, "\t")
		if len(cols) != 3 {
			t.Fatalf("bad manifest row: %q", line)
		}
		name, wantHex := cols[0], cols[2]
		src, ok := pairs[name]
		if !ok {
			t.Errorf("manifest names unknown pair %q", name)
			continue
		}
		out, _, err := Transform(src, target)
		if err != nil {
			t.Errorf("%s: Transform: %v", name, err)
			continue
		}
		if got := hex.EncodeToString(out); got != wantHex {
			t.Errorf("%s: canonical octets drifted\n got  %s\n want %s", name, got, wantHex)
		}
	}

	// The corpus README documents reuse and records the two-implementation
	// trial honestly as OPEN.
	readme, err := os.ReadFile(filepath.Join("testdata", "migration", "CORPUS.md"))
	if err != nil {
		t.Fatalf("read CORPUS.md: %v", err)
	}
	doc := string(readme)
	if !strings.Contains(doc, "independent") || !strings.Contains(doc, "OPEN") {
		t.Errorf("CORPUS.md must document independent-implementation reuse and the OPEN trial status")
	}
}
