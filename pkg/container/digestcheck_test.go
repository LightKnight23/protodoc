package container

import (
	"crypto/sha256"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_104_DigestMismatchAbortsBeforeDecode fixtures a segment body
// whose declared digest was computed over the original octets, then
// corrupts one octet of the body actually presented for decode. It
// asserts DecodeSegmentChecked (a) reports a *DigestMismatchError naming
// the segment ordinal and both the expected and actual digests, and (b)
// never invokes decode. It also confirms the positive path: a matching
// digest does invoke decode exactly once.
func TestFR_104_DigestMismatchAbortsBeforeDecode(t *testing.T) {
	const ordinal = 3
	original := []byte("segment octets used as the fixture body for FR-104's digest check")
	declared := pdlfmt.Digest256(sha256.Sum256(original))
	slot := SegmentTableSlot{SegmentType: SegmentTypeContent, Digest: declared}

	corrupted := append([]byte(nil), original...)
	corrupted[0] ^= 0xFF

	var decodeCalled bool
	decode := func([]byte) error {
		decodeCalled = true
		return nil
	}

	err := DecodeSegmentChecked(ordinal, slot, corrupted, decode)
	if err == nil {
		t.Fatal("expected a digest mismatch error, got nil")
	}
	if decodeCalled {
		t.Error("decode was invoked despite a digest mismatch; FR-104 requires abort before decode")
	}

	var mismatch *DigestMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("got error of type %T (%v), want *DigestMismatchError", err, err)
	}
	if mismatch.Ordinal != ordinal {
		t.Errorf("Ordinal = %d, want %d", mismatch.Ordinal, ordinal)
	}
	if mismatch.Expected != declared {
		t.Errorf("Expected = %x, want %x (the slot's declared digest)", mismatch.Expected, declared)
	}
	wantActual := pdlfmt.Digest256(sha256.Sum256(corrupted))
	if mismatch.Actual != wantActual {
		t.Errorf("Actual = %x, want %x (recomputed over the corrupted octets)", mismatch.Actual, wantActual)
	}

	decodeCalled = false
	if err := DecodeSegmentChecked(ordinal, slot, original, decode); err != nil {
		t.Fatalf("unexpected error for a segment whose digest matches: %v", err)
	}
	if !decodeCalled {
		t.Error("decode was not invoked once the digest verified")
	}
}
