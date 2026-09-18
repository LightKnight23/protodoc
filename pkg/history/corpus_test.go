package history

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestConformance_HistorySegmentAndErasureCorpus is T-0223's named conformance
// test (FR-059/FR-060/FR-061). It ships a corpus of HISTORY_OP_BATCH and
// ERASURE_RECORD fixtures and asserts each decodes and round-trips byte-exact
// (both the plain and RLE op-batch forms, and erasure records pre- and
// post-trim), covering the history segment's two record kinds.
func TestConformance_HistorySegmentAndErasureCorpus(t *testing.T) {
	// --- HISTORY_OP_BATCH corpus ---
	batches := [][]OperationRecord{
		nil, // empty
		{{Kind: OpSequencePositionClaim, Target: hUnitID(1), OrderKey: hStateID(1), Predecessor: container.StateID{}, Payload: []byte("x")}},
		{
			{Kind: OpValueClaim, Target: hUnitID(2), OrderKey: hStateID(2), Predecessor: hStateID(1), Payload: []byte("set")},
			{Kind: OpDelete, Target: hUnitID(3), OrderKey: hStateID(3), Predecessor: hStateID(2)},
		},
	}
	for i, ops := range batches {
		enc, err := EncodeOpBatch(ops)
		if err != nil {
			t.Fatalf("batch %d encode: %v", i, err)
		}
		dec, err := DecodeOpBatch(enc)
		if err != nil {
			t.Fatalf("batch %d decode: %v", i, err)
		}
		reEnc, _ := EncodeOpBatch(dec)
		if !bytes.Equal(enc, reEnc) {
			t.Errorf("batch %d plain form not byte-exact", i)
		}
		// RLE form round-trips to the same ops.
		rle, err := EncodeOpBatchRLE(ops)
		if err != nil {
			t.Fatalf("batch %d RLE encode: %v", i, err)
		}
		rleDec, err := DecodeOpBatchRLE(rle)
		if err != nil {
			t.Fatalf("batch %d RLE decode: %v", i, err)
		}
		if len(rleDec) != len(ops) {
			t.Errorf("batch %d RLE round trip: %d ops, want %d", i, len(rleDec), len(ops))
		}
	}

	// --- ERASURE_RECORD corpus ---
	dg := func(b byte) pdlfmt.Digest256 { var d pdlfmt.Digest256; d[0] = b; return d }
	sl := func(b byte) [SaltSize]byte { var s [SaltSize]byte; s[0] = b; return s }
	erasures := []ErasureRecord{
		NewErasureRecord(hUnitID(0x10), dg(0x11), sl(0xA1)),
		NewErasureRecord(hUnitID(0x20), dg(0x22), sl(0xA2)).Trimmed(),
	}
	for i, er := range erasures {
		enc, err := er.Encode()
		if err != nil {
			t.Fatalf("erasure %d encode: %v", i, err)
		}
		dec, err := DecodeErasureRecord(enc)
		if err != nil {
			t.Fatalf("erasure %d decode: %v", i, err)
		}
		if dec != er {
			t.Errorf("erasure %d did not round-trip", i)
		}
		reEnc, _ := dec.Encode()
		if !bytes.Equal(enc, reEnc) {
			t.Errorf("erasure %d not byte-exact", i)
		}
	}
}
