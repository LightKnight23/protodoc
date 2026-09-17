package integrity

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// determinismFixture reconstructs the fixed raw inputs whose golden roots are
// pinned in expected.json. It is deterministic (no ambient input), so a
// fresh reconstruction in any process yields byte-identical roots.
func determinismFixture() ([]container.SegmentTableSlot, []ContentRecord) {
	var slots []container.SegmentTableSlot
	for i := 0; i < 10; i++ {
		var d pdlfmt.Digest256
		d[0] = byte(i + 1)
		d[1] = byte(i * 7)
		slots = append(slots, container.SegmentTableSlot{
			SegmentType: container.SegmentTypeContent,
			Offset:      uint64(1<<20) + uint64(i)*64,
			Length:      uint64(64 + i),
			Digest:      d,
		})
	}
	var recs []ContentRecord
	for i := 0; i < 8; i++ {
		recs = append(recs, mkRec(i, i%2 == 0))
	}
	return slots, recs
}

// TestFR_003_TSAndTCRootsDeterministicAcrossTwoRuns is T-0142's named
// conformance test. It recomputes TSRoot and TCRoot from the fixture's raw
// inputs and asserts an exact match against the checked-in golden values in
// expected.json -- the two integrity trees recompute identically (closing
// NFR-001's canonical-octet-determinism concern for the trees). It also
// recomputes twice in-process to confirm intra-run stability.
func TestFR_003_TSAndTCRootsDeterministicAcrossTwoRuns(t *testing.T) {
	raw, err := os.ReadFile("testdata/conformance/integrity/determinism/expected.json")
	if err != nil {
		t.Fatalf("reading expected.json: %v", err)
	}
	var expected struct {
		TSRoot string `json:"ts_root"`
		TCRoot string `json:"tc_root"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("parsing expected.json: %v", err)
	}

	slots, recs := determinismFixture()

	ts, err := TSRoot(slots)
	if err != nil {
		t.Fatalf("TSRoot: %v", err)
	}
	tc, err := TCRoot(recs)
	if err != nil {
		t.Fatalf("TCRoot: %v", err)
	}

	if got := hex.EncodeToString(ts[:]); got != expected.TSRoot {
		t.Fatalf("TSRoot = %s, want pinned %s", got, expected.TSRoot)
	}
	if got := hex.EncodeToString(tc[:]); got != expected.TCRoot {
		t.Fatalf("TCRoot = %s, want pinned %s", got, expected.TCRoot)
	}

	// Second in-process recomputation from freshly reconstructed inputs is
	// byte-identical (a fresh process would compute the same, since the
	// fixture and both trees carry no ambient input).
	slots2, recs2 := determinismFixture()
	ts2, _ := TSRoot(slots2)
	tc2, _ := TCRoot(recs2)
	if ts2 != ts || tc2 != tc {
		t.Fatalf("roots differ across two recomputations from identical raw inputs")
	}
}
