package pdlfmt

import "testing"

func TestUint48RoundTrip(t *testing.T) {
	for _, v := range []uint64{0, 1, 255, 65536, maxUint48} {
		enc, err := AppendUint48(nil, v)
		if err != nil {
			t.Fatalf("AppendUint48(%d): %v", v, err)
		}
		if len(enc) != 6 {
			t.Fatalf("AppendUint48(%d) encoded length = %d, want 6", v, len(enc))
		}
		got, n, err := DecodeUint48(enc)
		if err != nil {
			t.Fatalf("DecodeUint48(%d): %v", v, err)
		}
		if n != 6 || got != v {
			t.Fatalf("DecodeUint48 round trip = (%d,%d), want (%d,6)", got, n, v)
		}
	}
}

func TestUint48RejectsOverflow(t *testing.T) {
	if _, err := AppendUint48(nil, maxUint48+1); err == nil {
		t.Fatal("AppendUint48: want error for value exceeding 48-bit range, got nil")
	}
}

func TestUint48RejectsTruncated(t *testing.T) {
	if _, _, err := DecodeUint48(make([]byte, 5)); err == nil {
		t.Fatal("DecodeUint48: want error for 5-octet input, got nil")
	}
}
