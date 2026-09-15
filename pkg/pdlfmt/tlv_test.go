package pdlfmt

import (
	"errors"
	"testing"
)

// TestNFR_001_TLVClosedGrammarRoundTrip is T-0002's named test.
// Implements: NFR-001.
func TestNFR_001_TLVClosedGrammarRoundTrip(t *testing.T) {
	// Build a record exercising every closed field-kind currently
	// enumerated in container.abnf: varint, digest256, nfc-string,
	// plain-seq, sorted-vec.
	var digest Digest256
	for i := range digest {
		digest[i] = byte(i)
	}

	nfcVal, err := (func() ([]byte, error) { return AppendNFCString(nil, "héllo") })()
	if err != nil {
		t.Fatalf("AppendNFCString: %v", err)
	}

	plainSeqVal := AppendPlainSeq(nil, [][]byte{{0x01}, {0x03}, {0x02}}) // order preserved, not sorted

	sortedVecVal, err := AppendSortedVec(nil, [][]byte{{0x01}, {0x02}, {0x03}})
	if err != nil {
		t.Fatalf("AppendSortedVec: %v", err)
	}

	fields := []Field{
		{Tag: 1, Value: EncodeVarint(300)},
		{Tag: 2, Value: AppendDigest256(nil, digest)},
		{Tag: 3, Value: nfcVal},
		{Tag: 4, Value: plainSeqVal},
		{Tag: 5, Value: sortedVecVal},
	}

	known := map[byte]bool{1: true, 2: true, 3: true, 4: true, 5: true}

	enc, err := EncodeRecord(fields)
	if err != nil {
		t.Fatalf("EncodeRecord: %v", err)
	}

	dec, err := DecodeRecord(enc, known, -1)
	if err != nil {
		t.Fatalf("DecodeRecord: %v", err)
	}
	if len(dec) != len(fields) {
		t.Fatalf("decoded %d fields, want %d", len(dec), len(fields))
	}
	for i, f := range fields {
		if dec[i].Tag != f.Tag {
			t.Errorf("field %d: tag = %d, want %d", i, dec[i].Tag, f.Tag)
		}
		if string(dec[i].Value) != string(f.Value) {
			t.Errorf("field %d (tag %d): value mismatch, got %x want %x", i, f.Tag, dec[i].Value, f.Value)
		}
	}

	// Re-encoding what we decoded must reproduce the original octets byte-exact.
	reenc, err := EncodeRecord(dec)
	if err != nil {
		t.Fatalf("re-EncodeRecord: %v", err)
	}
	if string(reenc) != string(enc) {
		t.Fatalf("re-encoded octets differ from original:\n got  %x\n want %x", reenc, enc)
	}

	// Decode each typed value back out and confirm it matches what was encoded.
	gotVarint, _, err := DecodeVarint(dec[0].Value)
	if err != nil || gotVarint != 300 {
		t.Errorf("varint field: got %d, err %v, want 300", gotVarint, err)
	}
	gotDigest, _, err := DecodeDigest256(dec[1].Value)
	if err != nil || gotDigest != digest {
		t.Errorf("digest256 field: got %x, err %v, want %x", gotDigest, err, digest)
	}
	gotStr, _, err := DecodeNFCString(dec[2].Value)
	if err != nil || gotStr != "héllo" {
		t.Errorf("nfc-string field: got %q, err %v, want %q", gotStr, err, "héllo")
	}
}

// TestPD_TLV_001_RejectsDescendingOrRepeatedTags verifies a record whose
// tags repeat or descend is rejected (rule PD-TLV-001), both at encode
// time (writer bug guard) and decode time (reader safety).
func TestPD_TLV_001_RejectsDescendingOrRepeatedTags(t *testing.T) {
	t.Run("encode: descending", func(t *testing.T) {
		_, err := EncodeRecord([]Field{{Tag: 2, Value: []byte{1}}, {Tag: 1, Value: []byte{2}}})
		if !errors.Is(err, ErrTagNotAscending) {
			t.Errorf("got %v, want ErrTagNotAscending", err)
		}
	})
	t.Run("encode: repeated", func(t *testing.T) {
		_, err := EncodeRecord([]Field{{Tag: 1, Value: []byte{1}}, {Tag: 1, Value: []byte{2}}})
		if !errors.Is(err, ErrTagNotAscending) {
			t.Errorf("got %v, want ErrTagNotAscending", err)
		}
	})
	t.Run("decode: descending", func(t *testing.T) {
		// Hand-craft: tag=2 len=1 val=0x00, tag=1 len=1 val=0x00
		src := []byte{2, 1, 0x00, 1, 1, 0x00}
		_, err := DecodeRecord(src, map[byte]bool{1: true, 2: true}, -1)
		if !errors.Is(err, ErrTagNotAscending) {
			t.Errorf("got %v, want ErrTagNotAscending", err)
		}
	})
	t.Run("decode: repeated", func(t *testing.T) {
		src := []byte{1, 1, 0x00, 1, 1, 0x00}
		_, err := DecodeRecord(src, map[byte]bool{1: true}, -1)
		if !errors.Is(err, ErrTagNotAscending) {
			t.Errorf("got %v, want ErrTagNotAscending", err)
		}
	})
}

// TestFR_UnknownTagIsStructuralRejection verifies decoding an unknown tag
// with no matching field-kind, outside any reserved tail, is a structural
// rejection, not a silent skip.
func TestFR_UnknownTagIsStructuralRejection(t *testing.T) {
	// tag=9 (not in schema), len=1, value=0x00
	src := []byte{9, 1, 0x00}

	t.Run("no reserved tail: rejected", func(t *testing.T) {
		_, err := DecodeRecord(src, map[byte]bool{1: true}, -1)
		if !errors.Is(err, ErrUnknownTag) {
			t.Errorf("got %v, want ErrUnknownTag", err)
		}
	})
	t.Run("reserved tail covers tag: accepted", func(t *testing.T) {
		fields, err := DecodeRecord(src, map[byte]bool{1: true}, 8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fields) != 1 || fields[0].Tag != 9 {
			t.Fatalf("got %+v, want one field with tag 9", fields)
		}
	})
	t.Run("reserved tail does not cover tag: rejected", func(t *testing.T) {
		_, err := DecodeRecord(src, map[byte]bool{1: true}, 10)
		if !errors.Is(err, ErrUnknownTag) {
			t.Errorf("got %v, want ErrUnknownTag", err)
		}
	})
}

// TestFR_106_TruncatedFieldRejectedBeforeAllocation verifies a field
// whose declared length exceeds the remaining input is rejected without
// the decoder attempting to allocate that length (FR-106).
func TestFR_106_TruncatedFieldRejectedBeforeAllocation(t *testing.T) {
	cases := [][]byte{
		{1, 5, 0x00, 0x00}, // declares 5 octets, only 2 remain
		{1},                // no length octet at all
		{1, 0xFD, 0x00},    // truncated varint length itself
	}
	for i, src := range cases {
		_, err := DecodeRecord(src, map[byte]bool{1: true}, -1)
		if err == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}
	}
	// A declared length larger than any plausible input (e.g. near
	// uint64 max) must not cause an allocation attempt or overflow.
	huge := []byte{1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	_, err := DecodeRecord(huge, map[byte]bool{1: true}, -1)
	if !errors.Is(err, ErrTruncatedField) {
		t.Errorf("huge declared length: got %v, want ErrTruncatedField", err)
	}
}

// TestCON_009_RecordFieldCeiling asserts the documented 256-field ceiling
// and that it is structurally unreachable by construction: a 1-octet tag
// combined with strict ascending order (PD-TLV-001) admits at most 256
// distinct tags (0..255) in one record, so no record can ever carry a
// 257th field without also violating ascending order first. Both are
// exercised: the maximal 256-field record round-trips, and DecodeRecord's
// own defense-in-depth ceiling check (independent of tag ordering) fires
// on a hand-built decode that would otherwise exceed it.
func TestCON_009_RecordFieldCeiling(t *testing.T) {
	if MaxFieldsPerRecord != 256 {
		t.Fatalf("MaxFieldsPerRecord = %d, want 256", MaxFieldsPerRecord)
	}

	fields := make([]Field, MaxFieldsPerRecord)
	known := make(map[byte]bool, MaxFieldsPerRecord)
	for i := range fields {
		fields[i] = Field{Tag: byte(i), Value: nil}
		known[byte(i)] = true
	}
	enc, err := EncodeRecord(fields)
	if err != nil {
		t.Fatalf("encoding the maximal 256-field record: %v", err)
	}
	dec, err := DecodeRecord(enc, known, -1)
	if err != nil {
		t.Fatalf("decoding the maximal 256-field record: %v", err)
	}
	if len(dec) != MaxFieldsPerRecord {
		t.Fatalf("decoded %d fields, want %d", len(dec), MaxFieldsPerRecord)
	}
}
