package history

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

func hUnitID(seed byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	for i := range id {
		id[i] = seed + byte(i)
	}
	return id
}

func hStateID(seed byte) container.StateID {
	var s container.StateID
	for i := range s {
		s[i] = seed + byte(i)
	}
	return s
}

// TestHistorySegment_RLEBatchEncoding is T-0208's named unit test (FR-059/
// FR-060). A HISTORY_OP_BATCH of operation-records encodes and decodes
// byte-exact, preserving operation order (never re-sorted) and every field of
// each record, across all three op-kinds and a run of same-prefix operations
// (the RLE-friendly path).
func TestHistorySegment_RLEBatchEncoding(t *testing.T) {
	ops := []OperationRecord{
		{Kind: OpSequencePositionClaim, Target: hUnitID(0x01), Predecessor: hStateID(0x10), OrderKey: hStateID(0x20), Payload: []byte("inserted run element")},
		{Kind: OpValueClaim, Target: hUnitID(0x02), Predecessor: hStateID(0x20), OrderKey: hStateID(0x30), Payload: []byte("new value")},
		{Kind: OpDelete, Target: hUnitID(0x03), Predecessor: hStateID(0x30), OrderKey: hStateID(0x40), Payload: nil},
		// A run of same-target/same-predecessor ops (the RLE-friendly case).
		{Kind: OpSequencePositionClaim, Target: hUnitID(0x05), Predecessor: hStateID(0x40), OrderKey: hStateID(0x50), Payload: []byte("a")},
		{Kind: OpSequencePositionClaim, Target: hUnitID(0x05), Predecessor: hStateID(0x40), OrderKey: hStateID(0x51), Payload: []byte("b")},
	}

	enc, err := EncodeOpBatch(ops)
	if err != nil {
		t.Fatalf("EncodeOpBatch: %v", err)
	}
	dec, err := DecodeOpBatch(enc)
	if err != nil {
		t.Fatalf("DecodeOpBatch: %v", err)
	}
	if !reflect.DeepEqual(dec, ops) {
		t.Errorf("decoded batch differs from original:\n got %+v\nwant %+v", dec, ops)
	}
	// Order preserved (not re-sorted).
	for i := range ops {
		if dec[i].Target != ops[i].Target || dec[i].OrderKey != ops[i].OrderKey {
			t.Errorf("operation %d out of order after round trip", i)
		}
	}
	// Byte-exact re-encode.
	reEnc, err := EncodeOpBatch(dec)
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(enc, reEnc) {
		t.Fatal("op batch round trip not byte-exact")
	}

	// A reserved op-kind is rejected on encode and decode.
	bad := []OperationRecord{{Kind: OpKind(0x03), Target: hUnitID(1), Predecessor: hStateID(1), OrderKey: hStateID(1)}}
	if _, err := EncodeOpBatch(bad); !errors.Is(err, ErrOpKindReserved) {
		t.Errorf("reserved op-kind encode: err = %v, want ErrOpKindReserved", err)
	}

	// Wrong discriminant rejected.
	corrupt := append([]byte(nil), enc...)
	if len(corrupt) >= 3 && corrupt[2] == historyOpBatchDiscriminant {
		corrupt[2] = 0x0A
		if _, err := DecodeOpBatch(corrupt); !errors.Is(err, ErrOpBatchDiscriminant) {
			t.Errorf("wrong discriminant: err = %v, want ErrOpBatchDiscriminant", err)
		}
	}

	// Empty batch round-trips.
	empty, err := EncodeOpBatch(nil)
	if err != nil {
		t.Fatalf("encode empty: %v", err)
	}
	if d, err := DecodeOpBatch(empty); err != nil || len(d) != 0 {
		t.Errorf("empty batch round trip: %d ops, err %v", len(d), err)
	}
}
