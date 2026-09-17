package content

import (
	"testing"
	"unicode"
)

// TestCON_004_ExcludedScalarClassesRejected is T-0124's named conformance
// test for rule PD-SCALAR-001. It runs a negative corpus with one fixture per
// excluded scalar class plus admissible controls, asserting (a) every
// excluded class is rejected, (b) rejection localises to the FIRST offending
// scalar's byte offset with the correct reason (CON-004: "rejection occurs at
// the first offending scalar before any digest is computed"), and (c)
// admissible text (including the whitelisted TAB/LF/CR controls) is accepted.
func TestCON_004_ExcludedScalarClassesRejected(t *testing.T) {
	if unicode.Version != ScalarUnicodeVersion {
		t.Fatalf("bound Unicode version drift: unicode.Version=%s want %s", unicode.Version, ScalarUnicodeVersion)
	}

	// Accepted: ASCII, a combining mark, CJK, an emoji, and the three
	// whitelisted controls all interleaved with ordinary text.
	accepted := []string{
		"Hello, world.",
		"cafe\u0301 re\u0301sume\u0301", // combining acute
		"\u4e2d\u6587\u6587\u6863",      // CJK
		"emoji \U0001F600 ok",           // assigned emoji
		"line1\nline2\tcol\r\n",         // whitelisted TAB/LF/CR
	}
	for _, s := range accepted {
		if err := CheckScalarAdmissibility(s); err != nil {
			t.Errorf("admissible string %q rejected: %v", s, err)
		}
	}

	// Rejected: one fixture per excluded class. Prefix "OK " (3 bytes) makes
	// the offending scalar start at byte offset 3, so we can assert the
	// first-offense localisation precisely.
	const prefix = "OK "
	cases := []struct {
		name       string
		bad        string
		wantReason ScalarRejectReason
	}{
		// A raw surrogate cannot appear in well-formed UTF-8; the encoded
		// form U+D800 is the invalid 3-byte sequence 0xED 0xA0 0x80, so via
		// string decoding it surfaces as invalid UTF-8 (which is itself
		// inadmissible). The ScalarSurrogate branch is covered directly below.
		{"surrogate bytes (invalid utf8)", prefix + "\xed\xa0\x80", ScalarInvalidUTF8},
		{"noncharacter FDD0", prefix + string([]rune{0xFDD0}), ScalarNoncharacter},
		{"noncharacter FFFE", prefix + string([]rune{0xFFFE}), ScalarNoncharacter},
		{"noncharacter 10FFFF", prefix + string([]rune{0x10FFFF}), ScalarNoncharacter},
		{"private use E000", prefix + string([]rune{0xE000}), ScalarPrivateUse},
		{"private use plane15", prefix + string([]rune{0xF0000}), ScalarPrivateUse},
		{"control NUL", prefix + string([]rune{0x0000}), ScalarControlNotWhitelisted},
		{"control C1", prefix + string([]rune{0x0085}), ScalarControlNotWhitelisted},
		{"bidi RLO", prefix + string([]rune{0x202E}), ScalarBidiOrDeprecatedFormat},
		{"bidi isolate", prefix + string([]rune{0x2066}), ScalarBidiOrDeprecatedFormat},
		{"deprecated format", prefix + string([]rune{0x206A}), ScalarBidiOrDeprecatedFormat},
		{"arabic letter mark", prefix + string([]rune{0x061C}), ScalarBidiOrDeprecatedFormat},
		{"unassigned", prefix + string([]rune{0x0378}), ScalarUnassigned}, // U+0378 unassigned in 15.0.0
	}
	for _, c := range cases {
		err := CheckScalarAdmissibility(c.bad)
		if err == nil {
			t.Errorf("%s: expected rejection, got accept", c.name)
			continue
		}
		if err.Reason != c.wantReason {
			t.Errorf("%s: reason = %v, want %v", c.name, err.Reason, c.wantReason)
		}
		if err.ByteOffset != len(prefix) {
			t.Errorf("%s: byte offset = %d, want %d (first offending scalar)", c.name, err.ByteOffset, len(prefix))
		}
	}

	// The ScalarSurrogate classification branch, reachable by any caller that
	// passes a surrogate rune directly (e.g. decoding a non-UTF-8 source),
	// must classify it as a surrogate rather than accept it.
	if got := classifyScalar(0xD800); got != ScalarSurrogate {
		t.Errorf("classifyScalar(U+D800) = %v, want ScalarSurrogate", got)
	}
	if got := classifyScalar(0xDFFF); got != ScalarSurrogate {
		t.Errorf("classifyScalar(U+DFFF) = %v, want ScalarSurrogate", got)
	}

	// First-offense wins: a string with two different violations reports the
	// EARLIER one, before any later scalar (and before any digest).
	two := prefix + string([]rune{0x202E}) + "x" + string([]rune{0xE000})
	err := CheckScalarAdmissibility(two)
	if err == nil || err.Reason != ScalarBidiOrDeprecatedFormat || err.ByteOffset != len(prefix) {
		t.Fatalf("first-offense precedence failed: got %+v", err)
	}
}
