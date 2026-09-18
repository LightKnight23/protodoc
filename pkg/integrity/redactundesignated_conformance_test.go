package integrity

import "testing"

// TestRedactUndesignatedOmissionConformance001 is T-0195's named conformance
// test (corpus redact-undesignated-omission-conformance-001, FR-077). Negative
// corpus: every document that omits a non-redactable (undesignated) unit that
// was present at signing must report Unverified, naming the omitted unit;
// paired with all-declared cases that stay Valid.
func TestRedactUndesignatedOmissionConformance001(t *testing.T) {
	u := func(b byte) UnitID { return redUnitID(b) }
	signedUnits := map[UnitID]bool{u(1): true, u(2): true, u(3): true, u(4): true}
	signedRedactable := map[UnitID]bool{u(2): true, u(4): true}

	full := []ContentRecord{
		{UnitID: u(1), Frame: []byte("1"), Redactable: false},
		{UnitID: u(2), Frame: []byte("2"), Redactable: true, Salt: redSalt(0x22)},
		{UnitID: u(3), Frame: []byte("3"), Redactable: false},
		{UnitID: u(4), Frame: []byte("4"), Redactable: true, Salt: redSalt(0x44)},
	}
	r2, _ := RedactRecord(full[1])
	r4, _ := RedactRecord(full[3])

	// Negative corpus: undesignated omissions -> Unverified.
	negatives := []struct {
		name    string
		current []ContentRecord
		omitted UnitID
	}{
		{"drop non-redactable u1", []ContentRecord{full[1], full[2], full[3]}, u(1)},
		{"drop non-redactable u3", []ContentRecord{full[0], full[1], full[3]}, u(3)},
		{"drop u3 even while u2/u4 redacted", []ContentRecord{full[0], r2, r4}, u(3)},
	}
	for _, c := range negatives {
		if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, c.current); v != VerdictUnverified {
			t.Errorf("%s: verdict = %v, want Unverified", c.name, v)
		}
		omitted := UndesignatedOmission(signedUnits, signedRedactable, c.current)
		found := false
		for _, o := range omitted {
			if o == c.omitted {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: expected %x named as omitted, got %v", c.name, c.omitted, omitted)
		}
	}

	// Positive: both redactable units redacted in place, all non-redactable
	// present -> Valid.
	allDeclared := []ContentRecord{full[0], r2, full[2], r4}
	if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, allDeclared); v != VerdictValid {
		t.Errorf("all-declared redaction: verdict = %v, want Valid", v)
	}
}
