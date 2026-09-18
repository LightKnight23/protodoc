package integrity

import "testing"

// TestFR_076_RedactPreservesTCRoot is T-0191's named unit test (FR-076).
// Redacting a designated redactable subtree (removing its frame and salt,
// retaining only the commitment leaf) leaves the T_C root UNCHANGED, so a
// signature over that root still verifies -- the redaction is a declared
// omission, not a tampering.
func TestFR_076_RedactPreservesTCRoot(t *testing.T) {
	records := []ContentRecord{
		{UnitID: redUnitID(0x01), Frame: []byte("public block one"), Redactable: false},
		{UnitID: redUnitID(0x02), Frame: []byte("SECRET redactable block"), Redactable: true, Salt: redSalt(0x22)},
		{UnitID: redUnitID(0x03), Frame: []byte("public block three"), Redactable: false},
		{UnitID: redUnitID(0x04), Frame: []byte("another SECRET block"), Redactable: true, Salt: redSalt(0x44)},
	}

	rootBefore, err := TCRoot(records)
	if err != nil {
		t.Fatalf("TCRoot before: %v", err)
	}

	// Redact record index 1 (the first redactable block).
	redacted := make([]ContentRecord, len(records))
	copy(redacted, records)
	r, err := RedactRecord(records[1])
	if err != nil {
		t.Fatalf("RedactRecord: %v", err)
	}
	redacted[1] = r

	rootAfter, err := TCRoot(redacted)
	if err != nil {
		t.Fatalf("TCRoot after: %v", err)
	}
	if rootAfter != rootBefore {
		t.Fatal("redaction changed the T_C root; a signature over it would no longer verify (FR-076 broken)")
	}

	// The redacted record carries no frame or salt.
	if redacted[1].Frame != nil || redacted[1].Salt != zeroSalt {
		t.Error("redacted record still carries frame or salt")
	}
	// Its retained leaf is exactly the original redactable-leaf commitment.
	if redacted[1].RetainedLeaf != TCLeafRedactable(records[1].Salt, records[1].Frame) {
		t.Error("retained leaf is not the original commitment")
	}

	// Redacting BOTH redactable blocks also preserves the root.
	r3, _ := RedactRecord(records[3])
	redacted[3] = r3
	rootBoth, _ := TCRoot(redacted)
	if rootBoth != rootBefore {
		t.Error("redacting both redactable blocks changed the T_C root")
	}

	// Redacting a NON-redactable record is refused.
	if _, err := RedactRecord(records[0]); err == nil {
		t.Error("redacting a non-redactable record should be refused")
	}
}
