// Unicode-version binding (NFR-020, contracts/container.abnf S1
// unicode-version-id NORMATIVE comment): the format binds each format
// major version to exactly one Unicode version, and a document must
// record the Unicode version it was written against. This is a separate,
// explicit validation step from Header structural decode (DecodeHeader):
// Header.UnicodeVersionID is a registry-issued per-document value in its
// own right (contracts/container.abnf S1), so a caller opting into
// NFR-020's binding rule calls ValidateUnicodeVersionBinding once a
// Header has already decoded successfully, rather than DecodeHeader
// rejecting an otherwise-well-formed header on this ground itself.
package container

import "fmt"

// UnicodeVersionByFormatMajor is the closed registry-issued binding NFR-020
// requires: the single unicode-version-id a document declaring the given
// format-major must record. Format-major 1 is the only value defined for
// this format version.
var UnicodeVersionByFormatMajor = map[uint16]uint16{
	1: 1,
}

// UnicodeVersionBindingError is NFR-020's rejection: a document's declared
// unicode-version-id does not match the single Unicode version pinned for
// its declared format-major. It names both values, plus the format-major
// they were checked against.
type UnicodeVersionBindingError struct {
	FormatMajor uint16
	Declared    uint16
	Expected    uint16
}

func (e *UnicodeVersionBindingError) Error() string {
	return fmt.Sprintf("container: format-major %d requires unicode-version-id %d, document declares %d (NFR-020)", e.FormatMajor, e.Expected, e.Declared)
}

// ErrUnicodeVersionUnknownFormatMajor is returned when h.FormatMajor has no
// entry in UnicodeVersionByFormatMajor at all -- distinct from a mismatched
// but registered format-major, since there is then no expected value to
// name in a UnicodeVersionBindingError.
var ErrUnicodeVersionUnknownFormatMajor = fmt.Errorf("container: no unicode-version-id is bound to this format-major")

// ValidateUnicodeVersionBinding implements NFR-020: it rejects h if its
// declared UnicodeVersionID does not match the single Unicode version
// bound to its declared FormatMajor, naming both in the returned
// *UnicodeVersionBindingError.
func ValidateUnicodeVersionBinding(h *Header) error {
	expected, ok := UnicodeVersionByFormatMajor[h.FormatMajor]
	if !ok {
		return fmt.Errorf("%w: format-major %d", ErrUnicodeVersionUnknownFormatMajor, h.FormatMajor)
	}
	if h.UnicodeVersionID != expected {
		return &UnicodeVersionBindingError{FormatMajor: h.FormatMajor, Declared: h.UnicodeVersionID, Expected: expected}
	}
	return nil
}
