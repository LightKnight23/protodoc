package container

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"testing"
)

func validHeader() *Header {
	return &Header{
		FormatMajor:        1,
		FormatMinor:        0,
		DocumentClass:      1,
		CapabilityWritten:  1,
		CapabilityRequired: 1,
		DurableClaim:       false,
		HistoryMode:        HistoryComplete,
		UnicodeVersionID:   1,
		ShapingProfileID:   0,
		PrefixLayoutID:     1,
	}
}

// TestFR_007_HeaderContentIndependent is T-0003's named test.
// Implements: FR-006, FR-007, NFR-020.
func TestFR_007_HeaderContentIndependent(t *testing.T) {
	fixtures := []*Header{
		validHeader(),
		{FormatMajor: 1, FormatMinor: 5, DocumentClass: 1, CapabilityWritten: 3, CapabilityRequired: 2, DurableClaim: true, HistoryMode: HistoryRetainedFromPoint, UnicodeVersionID: 2, ShapingProfileID: 7, PrefixLayoutID: 1},
		{FormatMajor: 1, FormatMinor: 0, DocumentClass: 1, CapabilityWritten: 0, CapabilityRequired: 0, DurableClaim: false, HistoryMode: HistoryNone, UnicodeVersionID: 1, ShapingProfileID: 0, PrefixLayoutID: 1},
	}

	for i, h := range fixtures {
		enc := h.Encode(nil)
		if len(enc) != HeaderSize {
			t.Fatalf("fixture %d: encoded length = %d, want %d", i, len(enc), HeaderSize)
		}
		dec, err := DecodeHeader(enc)
		if err != nil {
			t.Fatalf("fixture %d: DecodeHeader: %v", i, err)
		}
		if *dec != *h {
			t.Errorf("fixture %d: round-trip mismatch: got %+v, want %+v", i, *dec, *h)
		}
		reenc := dec.Encode(nil)
		if !bytes.Equal(reenc, enc) {
			t.Errorf("fixture %d: re-encoded octets differ:\n got  %x\n want %x", i, reenc, enc)
		}
	}

	// FR-007: two documents differing only in content body (not header
	// fields) must produce byte-identical Header octets. Header.Encode
	// takes no content parameter at all — content independence holds by
	// construction, since there is no code path by which content could
	// reach the encoding. This asserts the runtime consequence directly:
	// two independently-constructed, field-identical Headers always
	// encode identically, regardless of whatever content the caller's
	// document body happens to hold elsewhere.
	h1 := validHeader()
	h2 := validHeader()
	if !bytes.Equal(h1.Encode(nil), h2.Encode(nil)) {
		t.Fatal("two field-identical Headers encoded differently; Header must never derive from content")
	}
}

// TestFR_104_HeaderDigestMismatchAbortsBeforeDecode verifies a corrupted
// header-digest is rejected before any other field is trusted, and that
// the error names both the expected and actual digest.
func TestFR_104_HeaderDigestMismatchAbortsBeforeDecode(t *testing.T) {
	enc := validHeader().Encode(nil)
	corrupt := make([]byte, len(enc))
	copy(corrupt, enc)
	corrupt[0] = 'X' // corrupt magic without fixing the digest

	_, err := DecodeHeader(corrupt)
	if !errors.Is(err, ErrHeaderDigestMismatch) {
		t.Fatalf("got %v, want ErrHeaderDigestMismatch (digest check must precede magic check)", err)
	}
}

// TestPD_HDRZERO_001_RejectsNonzeroReserved verifies a nonzero
// header-reserved octet is rejected, and confirms it is checked before
// field-level enum validation (magic and digest already verified first).
func TestPD_HDRZERO_001_RejectsNonzeroReserved(t *testing.T) {
	h := validHeader()
	enc := h.Encode(nil)
	enc[offHeaderReservedStart] = 0x01
	// Recompute the digest so this test isolates the reserved-zero check
	// from the digest check (already covered separately).
	fixed := recomputeDigest(enc)

	_, err := DecodeHeader(fixed)
	if !errors.Is(err, ErrHeaderReservedNonzero) {
		t.Fatalf("got %v, want ErrHeaderReservedNonzero", err)
	}
}

// TestFR_010_CapabilityRequiredExceedsWritten is T-0004's named test,
// implemented here since T-0004 builds directly on T-0003's Header.
// Implements: FR-008, FR-009, FR-010.
func TestFR_010_CapabilityRequiredExceedsWritten(t *testing.T) {
	h := validHeader()
	h.CapabilityWritten = 2
	h.CapabilityRequired = 5
	enc := recomputeDigest(h.Encode(nil))

	_, err := DecodeHeader(enc)
	var mismatch *CapabilityMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("got %v, want *CapabilityMismatchError", err)
	}
	if mismatch.Required != 5 || mismatch.Written != 2 {
		t.Fatalf("error reports required=%d written=%d, want required=5 written=2", mismatch.Required, mismatch.Written)
	}
}

// TestFR_011_DurableClaimReadFromHeader is T-0005's named test.
// Implements: FR-011.
func TestFR_011_DurableClaimReadFromHeader(t *testing.T) {
	for _, claim := range []bool{false, true} {
		h := validHeader()
		h.DurableClaim = claim
		enc := h.Encode(nil)

		// DoD: the leading 512 octets alone determine durable-claim, with
		// zero further decode — DecodeHeader takes only the HeaderSize
		// window and nothing else.
		dec, err := DecodeHeader(enc[:HeaderSize])
		if err != nil {
			t.Fatalf("claim=%v: DecodeHeader: %v", claim, err)
		}
		if dec.DurableClaim != claim {
			t.Errorf("claim=%v: DurableClaim = %v after round-trip", claim, dec.DurableClaim)
		}
	}
}

func TestPD_DURABLE_001_RejectsInvalidOctet(t *testing.T) {
	h := validHeader()
	enc := h.Encode(nil)
	enc[offDurableClaim] = 2
	fixed := recomputeDigest(enc)
	_, err := DecodeHeader(fixed)
	if !errors.Is(err, ErrInvalidDurableClaim) {
		t.Fatalf("got %v, want ErrInvalidDurableClaim", err)
	}
}

func TestPD_MODE_001_RejectsInvalidHistoryModeOctet(t *testing.T) {
	h := validHeader()
	enc := h.Encode(nil)
	enc[offHistoryMode] = 3
	fixed := recomputeDigest(enc)
	_, err := DecodeHeader(fixed)
	if !errors.Is(err, ErrInvalidHistoryMode) {
		t.Fatalf("got %v, want ErrInvalidHistoryMode", err)
	}
}

func TestDecodeHeader_RejectsInvalidDocumentClassZero(t *testing.T) {
	h := validHeader()
	h.DocumentClass = 0
	enc := h.Encode(nil)
	_, err := DecodeHeader(enc)
	if !errors.Is(err, ErrInvalidDocumentClass) {
		t.Fatalf("got %v, want ErrInvalidDocumentClass", err)
	}
}

func TestDecodeHeader_RejectsTruncatedInput(t *testing.T) {
	_, err := DecodeHeader(make([]byte, 100))
	if !errors.Is(err, ErrHeaderTruncated) {
		t.Fatalf("got %v, want ErrHeaderTruncated", err)
	}
}

// recomputeDigest fixes up header-digest in place, over whatever raw octets
// are currently in enc[:480], WITHOUT reinterpreting them through the
// typed Header struct. This is deliberate: several tests mutate a single
// byte to an invalid raw value (e.g. durable-claim = 2, a value with no
// valid bool/enum meaning) specifically to test that DecodeHeader rejects
// it; round-tripping through Header's typed fields first would silently
// coerce that invalid byte back into a valid one and defeat the test.
func recomputeDigest(enc []byte) []byte {
	out := make([]byte, HeaderSize)
	copy(out, enc[:HeaderSize])
	digest := sha256.Sum256(out[:offHeaderDigest])
	copy(out[offHeaderDigest:HeaderSize], digest[:])
	return out
}
