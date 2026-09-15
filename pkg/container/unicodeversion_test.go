package container

import (
	"errors"
	"testing"
)

// TestNFR_020_UnicodeVersionBoundToFormatMajor is T-0031's named test.
// Implements: NFR-020.
func TestNFR_020_UnicodeVersionBoundToFormatMajor(t *testing.T) {
	t.Run("matching fixture passes", func(t *testing.T) {
		h := validHeader()
		h.FormatMajor = 1
		h.UnicodeVersionID = UnicodeVersionByFormatMajor[1]
		if err := ValidateUnicodeVersionBinding(h); err != nil {
			t.Fatalf("ValidateUnicodeVersionBinding: %v, want nil", err)
		}
	})

	t.Run("mismatched fixture is rejected naming both versions", func(t *testing.T) {
		h := validHeader()
		h.FormatMajor = 1
		h.UnicodeVersionID = UnicodeVersionByFormatMajor[1] + 1

		err := ValidateUnicodeVersionBinding(h)
		var mismatch *UnicodeVersionBindingError
		if !errors.As(err, &mismatch) {
			t.Fatalf("got %v, want *UnicodeVersionBindingError", err)
		}
		if mismatch.FormatMajor != 1 {
			t.Errorf("FormatMajor = %d, want 1", mismatch.FormatMajor)
		}
		if mismatch.Declared != UnicodeVersionByFormatMajor[1]+1 {
			t.Errorf("Declared = %d, want %d", mismatch.Declared, UnicodeVersionByFormatMajor[1]+1)
		}
		if mismatch.Expected != UnicodeVersionByFormatMajor[1] {
			t.Errorf("Expected = %d, want %d", mismatch.Expected, UnicodeVersionByFormatMajor[1])
		}
	})

	t.Run("unregistered format-major is rejected", func(t *testing.T) {
		h := validHeader()
		h.FormatMajor = 9999
		h.UnicodeVersionID = 1

		err := ValidateUnicodeVersionBinding(h)
		if !errors.Is(err, ErrUnicodeVersionUnknownFormatMajor) {
			t.Fatalf("got %v, want ErrUnicodeVersionUnknownFormatMajor", err)
		}
	})
}
