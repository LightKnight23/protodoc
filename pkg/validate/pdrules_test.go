package validate

import (
	"strings"
	"testing"
)

// This file is the negative-corpus conformance suite for the eleven
// previously-unimplemented validator rules (pdrules.go). Each test provides at
// least one hostile fixture proving the rule fires (rejects, naming the
// offending value) plus a valid fixture that passes, contributing the
// conformance case that closes each rule's zero-conformance gap.

func TestPD_DISC_001_ReservedDiscriminantRejected(t *testing.T) {
	// Hostile: reserved discriminants 0x10..0x3F and 0x80..0xFE are rejected.
	for _, d := range []uint8{0x10, 0x2A, 0x3F, 0x80, 0xFE} {
		f := CheckDiscriminant(d)
		if f == nil || f.RuleID != RulePDDISC001 {
			t.Errorf("discriminant 0x%02X must be rejected as %s", d, RulePDDISC001)
		}
	}
	// Valid: assigned discriminants pass.
	for _, d := range []uint8{0x01, 0x06, 0x0E, 0x0F} {
		if f := CheckDiscriminant(d); f != nil {
			t.Errorf("assigned discriminant 0x%02X should pass, got %v", d, f)
		}
	}
}

func TestPD_DUP_001_DualMechanismRejected(t *testing.T) {
	if f := CheckDualMechanism("cx-1", []string{"cap-a", "cap-b"}); f == nil || f.RuleID != RulePDDUP001 {
		t.Errorf("dual-mechanism construct must be rejected as %s", RulePDDUP001)
	}
	if f := CheckDualMechanism("cx-1", []string{"cap-a"}); f != nil {
		t.Errorf("single-capability construct should pass, got %v", f)
	}
	if f := CheckDualMechanism("cx-1", nil); f != nil {
		t.Errorf("zero-capability construct should pass, got %v", f)
	}
}

func TestPD_EXT_001_NonEnvelopedExtensionRejected(t *testing.T) {
	if f := CheckExtensionEnveloped("ext-1", false); f == nil || f.RuleID != RulePDEXT001 {
		t.Errorf("non-enveloped extension must be rejected as %s", RulePDEXT001)
	}
	if f := CheckExtensionEnveloped("ext-1", true); f != nil {
		t.Errorf("enveloped extension should pass, got %v", f)
	}
}

func TestPD_FONT_001_FontReferenceCompleteness(t *testing.T) {
	full := FontReference{Values: [7]string{"n", "v", "d", "a", "c", "f", "p"}}
	var dg [32]byte
	dg[0] = 0xAB
	full.ReferencedDigest = dg

	// Missing a value -> rejected.
	missing := full
	missing.Values[3] = ""
	if f := CheckFontReference(missing, dg); f == nil || f.RuleID != RulePDFONT001 {
		t.Errorf("font reference missing a value must be rejected as %s", RulePDFONT001)
	}
	// Digest substitution -> detected.
	var other [32]byte
	other[0] = 0xCD
	if f := CheckFontReference(full, other); f == nil || f.RuleID != RulePDFONT001 {
		t.Errorf("font digest substitution must be detected as %s", RulePDFONT001)
	}
	// Complete + matching digest -> passes.
	if f := CheckFontReference(full, dg); f != nil {
		t.Errorf("complete matching font reference should pass, got %v", f)
	}
}

func TestPD_INDEX_001_LeafCollisionRejected(t *testing.T) {
	if f := CheckIndexLeafKeys([]string{"k1", "k2", "k1"}); f == nil || f.RuleID != RulePDINDEX001 {
		t.Errorf("duplicate index leaf key must be rejected as %s", RulePDINDEX001)
	}
	if f := CheckIndexLeafKeys([]string{"k1", "k2", "k3"}); f != nil {
		t.Errorf("distinct index leaf keys should pass, got %v", f)
	}
}

func TestPD_MODE_001_ExactlyOneModeNoChange(t *testing.T) {
	// Zero or multiple declared modes -> rejected.
	if f := CheckMode(nil, ""); f == nil || f.RuleID != RulePDMODE001 {
		t.Errorf("zero declared modes must be rejected as %s", RulePDMODE001)
	}
	if f := CheckMode([]string{"a", "b"}, ""); f == nil || f.RuleID != RulePDMODE001 {
		t.Errorf("multiple declared modes must be rejected as %s", RulePDMODE001)
	}
	// Mode change -> rejected.
	if f := CheckMode([]string{"complete"}, "no-history"); f == nil || f.RuleID != RulePDMODE001 {
		t.Errorf("mode change must be rejected as %s", RulePDMODE001)
	}
	// Exactly one mode, unchanged -> passes.
	if f := CheckMode([]string{"complete"}, "complete"); f != nil {
		t.Errorf("stable single mode should pass, got %v", f)
	}
	if f := CheckMode([]string{"complete"}, ""); f != nil {
		t.Errorf("first single-mode declaration should pass, got %v", f)
	}
}

func TestPD_NFC_001_And_NORM_001_NonNFCRunRejected(t *testing.T) {
	// A decomposed sequence (e + combining acute) is not NFC.
	nonNFC := "e\u0301"
	nfc := "\u00e9" // precomposed é
	// PD-NFC-001 spelling.
	if f := CheckNFCRun(RulePDNFC001, "run-1", nonNFC); f == nil || f.RuleID != RulePDNFC001 {
		t.Errorf("non-NFC run must be rejected as %s", RulePDNFC001)
	}
	// PD-NORM-001 spelling (spec.md's id for the same rule).
	if f := CheckNFCRun(RulePDNORM001, "run-1", nonNFC); f == nil || f.RuleID != RulePDNORM001 {
		t.Errorf("non-NFC run must also be rejectable under %s", RulePDNORM001)
	}
	if f := CheckNFCRun(RulePDNFC001, "run-1", nfc); f != nil {
		t.Errorf("NFC run should pass, got %v", f)
	}
}

func TestPD_NFC_002_CrossBoundaryNonNFCRejected(t *testing.T) {
	// Left ends with a base letter, right starts with a combining mark: each is
	// NFC alone, but the concatenation is not.
	left := "e"
	right := "\u0301x"
	if f := CheckNFCBoundary("run-1", left, right); f == nil || f.RuleID != RulePDNFC002 {
		t.Errorf("cross-boundary non-NFC sequence must be rejected as %s", RulePDNFC002)
	}
	// Two independently-NFC runs that stay NFC when joined -> pass.
	if f := CheckNFCBoundary("run-1", "abc", "def"); f != nil {
		t.Errorf("NFC-safe boundary should pass, got %v", f)
	}
}

func TestPD_PREV_001_BoundedPrefixPayload(t *testing.T) {
	// Network required -> rejected.
	if f := CheckBoundedPrefixPayload(1000, true); f == nil || f.RuleID != RulePDPREV001 {
		t.Errorf("network-required payload must be rejected as %s", RulePDPREV001)
	}
	// Payload past the prefix bound -> rejected.
	if f := CheckBoundedPrefixPayload(BoundedPrefixLen+1, false); f == nil || f.RuleID != RulePDPREV001 {
		t.Errorf("out-of-prefix payload must be rejected as %s", RulePDPREV001)
	}
	// Payload within bound, no network -> passes.
	if f := CheckBoundedPrefixPayload(BoundedPrefixLen, false); f != nil {
		t.Errorf("in-prefix payload should pass, got %v", f)
	}
}

func TestPD_SEGTYPE_001_ReservedSegmentTypeRejected(t *testing.T) {
	for _, tp := range []uint8{5, 42, 255} {
		f := CheckSegmentType(tp)
		if f == nil || f.RuleID != RulePDSEGTYPE001 {
			t.Errorf("reserved segment-type %d must be rejected as %s", tp, RulePDSEGTYPE001)
		}
		// The message must direct rejection, not skipping.
		if f != nil && !strings.Contains(f.Message, "rejected") {
			t.Errorf("segment-type %d finding must reject, got %q", tp, f.Message)
		}
	}
	for _, tp := range []uint8{0, 1, 2, 3, 4} {
		if f := CheckSegmentType(tp); f != nil {
			t.Errorf("assigned segment-type %d should pass, got %v", tp, f)
		}
	}
}
