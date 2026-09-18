package integrity

import "testing"

// TestRedactOrphanResidueConformance001 is T-0197's named conformance test
// (corpus redact-orphan-residue-conformance-001, FR-078). It checks that no
// removed-unit octets survive on any "orphan" surface a naive implementation
// might leak them onto -- layout advances, cached renderings, index entries,
// preview payloads -- by scanning a published output assembled from retained
// content plus those auxiliary surfaces, none of which may carry a removed
// unit's octets.
func TestRedactOrphanResidueConformance001(t *testing.T) {
	removed := []byte("REMOVED-CONFIDENTIAL-CLAUSE-9")

	// Retained content and clean auxiliary surfaces (index, preview) that do
	// NOT contain the removed octets.
	retained := []ContentRecord{
		{UnitID: redUnitID(0x01), Frame: []byte("kept clause 1")},
		{UnitID: redUnitID(0x02), Frame: []byte("kept clause 2")},
	}
	cleanIndex := []byte("idx: clause1@0 clause2@13")
	cleanPreview := []byte("preview raster: [kept content only]")

	out := Publish(PublishInput{Retained: retained})
	// The publish re-emission plus clean auxiliary surfaces (a full published
	// artefact includes these orphan surfaces; here they are appended to model
	// the whole artefact being scanned).
	whole := PublishOutput{Emitted: append(append(append([]byte(nil), out.Emitted...), cleanIndex...), cleanPreview...)}

	if findings := ScanResidue(whole, [][]byte{removed}); len(findings) != 0 {
		t.Fatalf("clean artefact wrongly flagged residue: %+v", findings)
	}

	// Negative corpus: each orphan surface leaking the removed octets is
	// caught by the residue scan.
	orphanLeaks := []struct {
		name string
		aux  []byte
	}{
		{"index entry leaks removed octets", append([]byte("idx: "), removed...)},
		{"preview payload leaks removed octets", append([]byte("preview: "), removed...)},
		{"layout-advance cache leaks removed octets", append(removed, []byte(" adv=12")...)},
	}
	for _, c := range orphanLeaks {
		leaky := PublishOutput{Emitted: append(append([]byte(nil), out.Emitted...), c.aux...)}
		if len(ScanResidue(leaky, [][]byte{removed})) == 0 {
			t.Errorf("%s: residue scan missed the leaked removed-unit octets", c.name)
		}
	}
}
