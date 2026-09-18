package integrity

import "testing"

// TestFR_077_UndesignatedOmissionUnverified is T-0194's named unit test
// (FR-077). IF a signed document omits content that was NOT designated
// redactable at signing time, a conforming reader presents it as UNVERIFIED --
// never valid, never attested-with-declared-omissions. Omitting a designated
// redactable subtree (redacted in place, preserving the T_C root) is a
// declared omission and stays Valid; omitting anything else is undesignated.
func TestFR_077_UndesignatedOmissionUnverified(t *testing.T) {
	u1, u2, u3 := redUnitID(0x01), redUnitID(0x02), redUnitID(0x03)
	// At signing: u1 non-redactable, u2 redactable, u3 non-redactable.
	signedUnits := map[UnitID]bool{u1: true, u2: true, u3: true}
	signedRedactable := map[UnitID]bool{u2: true}

	full := []ContentRecord{
		{UnitID: u1, Frame: []byte("one"), Redactable: false},
		{UnitID: u2, Frame: []byte("two"), Redactable: true, Salt: redSalt(0x22)},
		{UnitID: u3, Frame: []byte("three"), Redactable: false},
	}

	// (1) No omission -> Valid.
	if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, full); v != VerdictValid {
		t.Errorf("no omission: verdict = %v, want Valid", v)
	}

	// (2) The designated redactable subtree redacted in place -> Valid
	// (declared omission; still present as its retained commitment).
	r2, _ := RedactRecord(full[1])
	declared := []ContentRecord{full[0], r2, full[2]}
	if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, declared); v != VerdictValid {
		t.Errorf("declared redaction: verdict = %v, want Valid", v)
	}

	// (3) A NON-redactable record (u3) dropped entirely -> Unverified.
	undesignated := []ContentRecord{full[0], full[1]} // u3 missing
	if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, undesignated); v != VerdictUnverified {
		t.Errorf("undesignated omission (u3 dropped): verdict = %v, want Unverified", v)
	}
	omitted := UndesignatedOmission(signedUnits, signedRedactable, undesignated)
	if len(omitted) != 1 || omitted[0] != u3 {
		t.Errorf("expected u3 named as the undesignated omission, got %v", omitted)
	}

	// (4) A redactable record dropped ENTIRELY (not redacted-in-place, so its
	// retained commitment is absent) is still an omission of its unit; since
	// u2 IS designated redactable, dropping it entirely is NOT undesignated --
	// but it would change the T_C root (no retained leaf). FR-077 targets
	// UNDESIGNATED omissions specifically; u2 being redactable means it is a
	// (mal-formed) declared case, so it is not flagged here.
	dropRedactable := []ContentRecord{full[0], full[2]} // u2 missing entirely
	if v := UndesignatedOmissionVerdict(signedUnits, signedRedactable, dropRedactable); v != VerdictValid {
		t.Errorf("dropping a designated-redactable unit is not an UNDESIGNATED omission: verdict = %v, want Valid", v)
	}
}
