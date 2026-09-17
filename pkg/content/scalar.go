package content

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Scalar admissibility (T-0124, CON-004, rule PD-SCALAR-001). CON-004 admits
// in document text and identifier strings ONLY Unicode scalar values that are
// assigned in the bound Unicode version (15.0.0, matching unicode.Version and
// the NFC tables' nfcUnicodeVersion), EXCLUDING:
//
//   - bidirectional and deprecated formatting characters,
//   - noncharacters,
//   - surrogate scalars,
//   - private-use scalars,
//   - control characters outside an enumerated whitelist.
//
// The check exists because, without a closed permitted set, in-band
// directional controls reappear as ordinary text (the source-hiding attack
// the structural-direction rule closes), unassigned scalars make language
// resolution and raster determinism undefined, and invisible or confusable
// scalars let content evade the octet-level redaction scan. Per CON-004's
// verify clause, rejection occurs at the FIRST offending scalar, before any
// digest is computed; ScalarAdmissibilityError carries that scalar's byte
// offset and code point so a validator can report it precisely.

// ScalarUnicodeVersion is the bound Unicode version whose assigned set and
// property tables define admissibility. It must agree with unicode.Version
// and the NFC tables, so the whole core path decides scalar identity against
// one Unicode version.
const ScalarUnicodeVersion = "15.0.0"

// ScalarRejectReason enumerates why a scalar is inadmissible under CON-004.
type ScalarRejectReason int

const (
	ScalarOK ScalarRejectReason = iota
	ScalarInvalidUTF8
	ScalarSurrogate
	ScalarNoncharacter
	ScalarPrivateUse
	ScalarControlNotWhitelisted
	ScalarBidiOrDeprecatedFormat
	ScalarUnassigned
)

func (r ScalarRejectReason) String() string {
	switch r {
	case ScalarOK:
		return "admissible"
	case ScalarInvalidUTF8:
		return "invalid UTF-8"
	case ScalarSurrogate:
		return "surrogate scalar"
	case ScalarNoncharacter:
		return "noncharacter"
	case ScalarPrivateUse:
		return "private-use scalar"
	case ScalarControlNotWhitelisted:
		return "control character outside whitelist"
	case ScalarBidiOrDeprecatedFormat:
		return "bidirectional or deprecated formatting character"
	case ScalarUnassigned:
		return "unassigned scalar in the bound Unicode version"
	default:
		return "unknown"
	}
}

// RulePDScalar001 is the validator rule id CON-004's admissibility check
// reports on.
const RulePDScalar001 = "PD-SCALAR-001"

// ScalarAdmissibilityError localises the first inadmissible scalar in a
// string: its byte offset, the offending code point, and the reason. It is
// the value CheckScalarAdmissibility returns; a validator maps it to a
// PD-SCALAR-001 diagnostic.
type ScalarAdmissibilityError struct {
	ByteOffset int
	CodePoint  rune
	Reason     ScalarRejectReason
}

func (e *ScalarAdmissibilityError) Error() string {
	return fmt.Sprintf("content: %s: inadmissible scalar U+%04X at byte offset %d (%s)",
		RulePDScalar001, e.CodePoint, e.ByteOffset, e.Reason)
}

// controlWhitelist is the enumerated set of control characters CON-004
// admits: TAB, LF, and CR. Every other control (C0/C1) is rejected.
var controlWhitelist = map[rune]bool{
	0x0009: true, // CHARACTER TABULATION (TAB)
	0x000A: true, // LINE FEED (LF)
	0x000D: true, // CARRIAGE RETURN (CR)
}

// bidiAndDeprecatedFormat is the enumerated set of bidirectional control and
// deprecated formatting characters CON-004 excludes even though they are
// assigned format (Cf) characters. Rejecting them in-band is what closes the
// source-hiding attack class.
var bidiAndDeprecatedFormat = map[rune]bool{
	0x061C: true, // ARABIC LETTER MARK
	0x200E: true, // LEFT-TO-RIGHT MARK
	0x200F: true, // RIGHT-TO-LEFT MARK
	0x202A: true, // LEFT-TO-RIGHT EMBEDDING
	0x202B: true, // RIGHT-TO-LEFT EMBEDDING
	0x202C: true, // POP DIRECTIONAL FORMATTING
	0x202D: true, // LEFT-TO-RIGHT OVERRIDE
	0x202E: true, // RIGHT-TO-LEFT OVERRIDE
	0x2066: true, // LEFT-TO-RIGHT ISOLATE
	0x2067: true, // RIGHT-TO-LEFT ISOLATE
	0x2068: true, // FIRST STRONG ISOLATE
	0x2069: true, // POP DIRECTIONAL ISOLATE
	0x206A: true, // INHIBIT SYMMETRIC SWAPPING (deprecated)
	0x206B: true, // ACTIVATE SYMMETRIC SWAPPING (deprecated)
	0x206C: true, // INHIBIT ARABIC FORM SHAPING (deprecated)
	0x206D: true, // ACTIVATE ARABIC FORM SHAPING (deprecated)
	0x206E: true, // NATIONAL DIGIT SHAPES (deprecated)
	0x206F: true, // NOMINAL DIGIT SHAPES (deprecated)
}

// isNoncharacter reports whether r is a Unicode noncharacter: the 32 in
// U+FDD0..U+FDEF and the two per plane U+xFFFE/U+xFFFF (planes 0..16).
func isNoncharacter(r rune) bool {
	if r >= 0xFDD0 && r <= 0xFDEF {
		return true
	}
	low := r & 0xFFFF
	return low == 0xFFFE || low == 0xFFFF
}

// isAssigned reports whether r is assigned in the bound Unicode version. A
// scalar is assigned iff it belongs to at least one general category other
// than Cn (Cn is exactly the unassigned category). Go's unicode.Categories
// map (bundled Unicode unicode.Version == ScalarUnicodeVersion) DOES expose a
// Cn table, so we scan every two-letter subcategory except Cn.
func isAssigned(r rune) bool {
	for name, tab := range unicode.Categories {
		if len(name) == 1 {
			// Skip the single-letter super-categories (L, M, N, ...) to
			// avoid double-scanning; the two-letter subcategories cover the
			// same runes.
			continue
		}
		if name == "Cn" {
			// Cn is the unassigned category; membership means NOT assigned.
			continue
		}
		if unicode.Is(tab, r) {
			return true
		}
	}
	return false
}

// classifyScalar returns why r is inadmissible, or ScalarOK. The order of the
// tests is chosen so the most specific structural reasons (surrogate,
// noncharacter, private-use) are reported before the general assigned check.
func classifyScalar(r rune) ScalarRejectReason {
	if r > utf8.MaxRune {
		return ScalarInvalidUTF8
	}
	if r >= 0xD800 && r <= 0xDFFF {
		return ScalarSurrogate
	}
	if isNoncharacter(r) {
		return ScalarNoncharacter
	}
	if unicode.In(r, unicode.Co) { // private use (BMP + planes 15,16)
		return ScalarPrivateUse
	}
	if bidiAndDeprecatedFormat[r] {
		return ScalarBidiOrDeprecatedFormat
	}
	if unicode.IsControl(r) { // category Cc (C0 + C1)
		if controlWhitelist[r] {
			return ScalarOK
		}
		return ScalarControlNotWhitelisted
	}
	if !isAssigned(r) {
		return ScalarUnassigned
	}
	return ScalarOK
}

// CheckScalarAdmissibility scans s left to right and returns a
// *ScalarAdmissibilityError for the FIRST inadmissible scalar (CON-004's
// "rejection occurs at the first offending scalar"), or nil if every scalar
// is admissible. Invalid UTF-8 is itself inadmissible (utf8.RuneError with a
// one-byte advance), reported at its byte offset. Callers must run this
// before computing any digest over the text.
func CheckScalarAdmissibility(s string) *ScalarAdmissibilityError {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return &ScalarAdmissibilityError{ByteOffset: i, CodePoint: r, Reason: ScalarInvalidUTF8}
		}
		if reason := classifyScalar(r); reason != ScalarOK {
			return &ScalarAdmissibilityError{ByteOffset: i, CodePoint: r, Reason: reason}
		}
		i += size
	}
	return nil
}
