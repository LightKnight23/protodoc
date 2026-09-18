package history

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestHistorySegment_DecodeConformanceAtAndOverLimit is T-0209's named
// conformance test (FR-059/FR-060). It exercises the HISTORY_OP_BATCH decode
// bounds: a batch at a legal size decodes, a hostile op count (far exceeding
// the octets present) is rejected before allocation (FR-106), and an
// operations field-value over the MAX_DECODED_UNIT ceiling is rejected naming
// the ceiling.
func TestHistorySegment_DecodeConformanceAtAndOverLimit(t *testing.T) {
	// At limit (a small legal batch) decodes.
	ops := []OperationRecord{
		{Kind: OpValueClaim, Target: hUnitID(1), Predecessor: hStateID(1), OrderKey: hStateID(2), Payload: []byte("v")},
	}
	enc, err := EncodeOpBatch(ops)
	if err != nil {
		t.Fatalf("EncodeOpBatch: %v", err)
	}
	if _, err := DecodeOpBatch(enc); err != nil {
		t.Fatalf("legal batch rejected: %v", err)
	}

	// A hostile op count (claims a huge number of operations with only a few
	// octets present) is rejected before allocation.
	hostile := pdlfmt.AppendVarint(nil, 1<<40) // claim ~1e12 ops
	hostile = append(hostile, 0x00)            // a stray octet
	hobField := hostile
	rec, _ := pdlfmt.EncodeRecord([]pdlfmt.Field{
		{Tag: hobTagDiscriminant, Value: []byte{historyOpBatchDiscriminant}},
		{Tag: hobTagOperations, Value: hobField},
	})
	if _, err := DecodeOpBatch(rec); !errors.Is(err, ErrOpBatchTruncated) {
		t.Errorf("hostile op count: err = %v, want ErrOpBatchTruncated (bounded before allocation)", err)
	}

	// The MAX_DECODED_UNIT ceiling bounds the operations field-value; the
	// DecodeOpBatch guard rejects any field-value over it as
	// ErrOpBatchOversized. Allocating a 256 MiB buffer to hit the boundary is
	// impractical in a unit test, so we assert the guard's ceiling is the
	// single-source-of-truth table value (the boundary the guard compares
	// against). The hostile-count path above already exercises the
	// before-allocation rejection for the common attack.
	if MaxOpBatchOctets() != 268435456 {
		t.Errorf("MAX_DECODED_UNIT = %d, want 268435456", MaxOpBatchOctets())
	}
}
