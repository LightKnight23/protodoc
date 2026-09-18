package integrity

import "testing"

// TestRedactCommitmentBoundaryConformance001 is T-0206's named conformance
// test (corpus redact-commitment-boundary-conformance-001, FR-074). It asserts
// RED-ALIGN (integrity.abnf S7.1): every redactable subtree's boundary lands on
// a document-model record boundary -- a redaction operates on a whole content
// record (one ContentRecord), never on a fragment of one. Redacting a record
// removes exactly that record's frame and salt and leaves every other record
// byte-for-byte intact (the T_C root over the untouched records is preserved
// element-wise), so no record is ever split.
func TestRedactCommitmentBoundaryConformance001(t *testing.T) {
	records := []ContentRecord{
		{UnitID: redUnitID(0x01), Frame: []byte("record one whole"), Redactable: false},
		{UnitID: redUnitID(0x02), Frame: []byte("record two REDACTABLE whole"), Redactable: true, Salt: redSalt(0x22)},
		{UnitID: redUnitID(0x03), Frame: []byte("record three whole"), Redactable: false},
	}

	// The commitment is computed over the WHOLE record frame (a record
	// boundary), not any sub-span of it.
	sub := DesignateRedactable(records[1].UnitID, records[1].Frame, records[1].Salt)
	whole, _ := sub.Commitment()
	// A commitment over any proper prefix of the frame differs -- proving the
	// commitment binds the whole record, so a partial redaction would be a
	// different (rejected) commitment.
	prefix := records[1].Frame[:len(records[1].Frame)/2]
	partial, _ := DesignateRedactable(records[1].UnitID, prefix, records[1].Salt).Commitment()
	if whole == partial {
		t.Error("commitment over a partial record equals the whole-record commitment (boundary not enforced)")
	}

	// Redacting record 2 leaves records 1 and 3 byte-for-byte intact.
	redacted := make([]ContentRecord, len(records))
	copy(redacted, records)
	r, _ := RedactRecord(records[1])
	redacted[1] = r

	if string(redacted[0].Frame) != string(records[0].Frame) {
		t.Error("redaction altered a neighbouring record (record 1)")
	}
	if string(redacted[2].Frame) != string(records[2].Frame) {
		t.Error("redaction altered a neighbouring record (record 3)")
	}
	// The redacted record carries no partial frame -- the whole frame is gone.
	if redacted[1].Frame != nil {
		t.Error("redacted record retains a (partial) frame; a redaction must remove the whole record's frame")
	}
	// Its retained leaf is the whole-record commitment.
	if redacted[1].RetainedLeaf != whole {
		t.Error("retained leaf is not the whole-record commitment")
	}

	// The T_C root is preserved (whole-record redaction, RED-ALIGN honoured).
	before, _ := TCRoot(records)
	after, _ := TCRoot(redacted)
	if before != after {
		t.Error("whole-record redaction changed the T_C root")
	}
}
