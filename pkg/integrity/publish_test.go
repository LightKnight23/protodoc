package integrity

import (
	"bytes"
	"testing"
)

// TestFR_078_PublishOmitsRemovedUnitOctets is T-0196's named integration test
// (FR-078). The publish operation's output contains no octet sequence
// belonging to any removed or superseded content unit: a residue scan over the
// output finds zero occurrences of any removed unit's frame octets, while the
// retained content is present.
func TestFR_078_PublishOmitsRemovedUnitOctets(t *testing.T) {
	kept1 := []byte("PUBLIC-INTRO-PARAGRAPH")
	kept2 := []byte("PUBLIC-CONCLUSION")
	removed1 := []byte("SECRET-SALARY-FIGURE-12345")
	removed2 := []byte("PRIVATE-HOME-ADDRESS-742-EVERGREEN")

	// A redacted record retained by commitment (its plaintext is a removed
	// secret that must not appear).
	redactedSecret := []byte("REDACTED-DIAGNOSIS-TEXT")
	salt := redSalt(0x55)
	rr, _ := RedactRecord(ContentRecord{UnitID: redUnitID(0x09), Frame: redactedSecret, Redactable: true, Salt: salt})

	in := PublishInput{
		Retained: []ContentRecord{
			{UnitID: redUnitID(0x01), Frame: kept1},
			rr, // redacted: emits only its commitment leaf, never redactedSecret
			{UnitID: redUnitID(0x02), Frame: kept2},
		},
		RemovedFrames: [][]byte{removed1, removed2, redactedSecret},
	}

	out := Publish(in)

	// Retained content is present.
	if !bytes.Contains(out.Emitted, kept1) || !bytes.Contains(out.Emitted, kept2) {
		t.Error("published output is missing retained content")
	}

	// No removed-unit octets survive anywhere in the output.
	findings := ScanResidue(out, in.RemovedFrames)
	if len(findings) != 0 {
		t.Fatalf("residue scan found %d removed-unit octet sequence(s) in the published output: %+v", len(findings), findings)
	}

	// Direct octet checks for the removed sequences.
	for _, r := range [][]byte{removed1, removed2, redactedSecret} {
		if bytes.Contains(out.Emitted, r) {
			t.Errorf("removed unit octets %q appear in the published output", r)
		}
	}

	// The redacted record contributed its commitment leaf, not plaintext.
	if !bytes.Contains(out.Emitted, rr.RetainedLeaf[:]) {
		t.Error("redacted record's retained commitment leaf is missing from the output")
	}

	// Negative control: a scanner over a leaky output (retaining the secret as
	// plaintext) DOES find the residue -- proving the scan is not vacuous.
	leaky := Publish(PublishInput{Retained: []ContentRecord{{UnitID: redUnitID(0x01), Frame: removed1}}})
	if len(ScanResidue(leaky, [][]byte{removed1})) == 0 {
		t.Error("residue scanner failed to find a genuinely leaked removed unit (scan is vacuous)")
	}
}
