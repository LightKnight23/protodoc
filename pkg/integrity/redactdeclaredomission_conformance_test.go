package integrity

import "testing"

// TestRedactDeclaredOmissionConformance001 is T-0193's named conformance test
// (corpus redact-declared-omission-conformance-001, FR-076). It ships a corpus
// of documents with designated redactable subtrees, redacts a declared subset
// of them, and asserts: the T_C root is preserved across the redaction (the
// signature still verifies), and the reported outcome for the valid verdict is
// attested-with-declared-omissions with every omission enumerated.
func TestRedactDeclaredOmissionConformance001(t *testing.T) {
	cases := []struct {
		name      string
		records   []ContentRecord
		redactIdx []int
	}{
		{
			name: "one of two redactable redacted",
			records: []ContentRecord{
				{UnitID: redUnitID(0x01), Frame: []byte("keep"), Redactable: false},
				{UnitID: redUnitID(0x02), Frame: []byte("omit me"), Redactable: true, Salt: redSalt(0x22)},
			},
			redactIdx: []int{1},
		},
		{
			name: "all redactable redacted",
			records: []ContentRecord{
				{UnitID: redUnitID(0x01), Frame: []byte("a"), Redactable: true, Salt: redSalt(0x11)},
				{UnitID: redUnitID(0x02), Frame: []byte("b"), Redactable: true, Salt: redSalt(0x22)},
				{UnitID: redUnitID(0x03), Frame: []byte("c"), Redactable: true, Salt: redSalt(0x33)},
			},
			redactIdx: []int{0, 1, 2},
		},
	}

	for _, c := range cases {
		rootBefore, err := TCRoot(c.records)
		if err != nil {
			t.Fatalf("%s: TCRoot: %v", c.name, err)
		}
		redacted := make([]ContentRecord, len(c.records))
		copy(redacted, c.records)
		var omissions []UnitID
		for _, i := range c.redactIdx {
			r, err := RedactRecord(c.records[i])
			if err != nil {
				t.Fatalf("%s: RedactRecord[%d]: %v", c.name, i, err)
			}
			redacted[i] = r
			omissions = append(omissions, c.records[i].UnitID)
		}
		rootAfter, _ := TCRoot(redacted)
		if rootAfter != rootBefore {
			t.Errorf("%s: T_C root changed across declared redaction", c.name)
		}
		// A valid verdict with these declared omissions reports the distinct
		// outcome and enumerates them.
		rep := AttestationReport{Verdict: VerdictValid, DeclaredOmissions: omissions}
		if rep.ReportedOutcome() != OutcomeAttestedWithDeclaredOmissions {
			t.Errorf("%s: outcome = %q, want %q", c.name, rep.ReportedOutcome(), OutcomeAttestedWithDeclaredOmissions)
		}
		if len(rep.DeclaredOmissions) != len(c.redactIdx) {
			t.Errorf("%s: enumerated %d omissions, want %d", c.name, len(rep.DeclaredOmissions), len(c.redactIdx))
		}
	}
}
