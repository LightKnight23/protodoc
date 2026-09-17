package integrity

import (
	"encoding/hex"
	"testing"
)

// TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed is T-0129's named
// test. ABSENT_CHILD_DIGEST equals SHA-256(0x00) against a hard-coded
// expected hex value, and all 8 registry tag values (0x00 plus the 7 real
// preimage-start octets) are pairwise distinct.
func TestFR_003_DomainTagsDistinctAndAbsentChildDigestFixed(t *testing.T) {
	// ABSENT_CHILD_DIGEST is the fixed SHA-256 of the single octet 0x00.
	const wantHex = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	gotHex := hex.EncodeToString(AbsentChildDigest[:])
	if gotHex != wantHex {
		t.Fatalf("AbsentChildDigest = %s, want SHA-256(0x00) = %s", gotHex, wantHex)
	}

	// All 8 registry tags pairwise distinct.
	tags := map[string]byte{
		"AbsentChild":         DomainAbsentChild,         // 0x00
		"TSInternal":          DomainTSInternal,          // 0x01
		"TCLeafRedactable":    DomainTCLeafRedactable,    // 0x02
		"SignedObject":        DomainSignedObject,        // 0x04
		"TCLeafNonredactable": DomainTCLeafNonredactable, // 0x07
		"TCInternal":          DomainTCInternal,          // 0x08
		"SeveranceCommitment": DomainSeveranceCommitment, // 0x09
		"StructureDigest":     DomainStructureDigest,     // 0x0A
	}
	if len(tags) != 8 {
		t.Fatalf("expected 8 named tags, got %d", len(tags))
	}
	seen := map[byte]string{}
	for name, v := range tags {
		if prev, dup := seen[v]; dup {
			t.Fatalf("tag value 0x%02x used by both %s and %s (not distinct)", v, prev, name)
		}
		seen[v] = name
	}
	// Confirm the exact values match the integrity.abnf S2 registry.
	want := map[byte]bool{0x00: true, 0x01: true, 0x02: true, 0x04: true, 0x07: true, 0x08: true, 0x09: true, 0x0A: true}
	for v := range seen {
		if !want[v] {
			t.Fatalf("registry has unexpected tag value 0x%02x", v)
		}
	}

	// realPreimageTags must exclude 0x00 (the property that makes the filler
	// uncollidable) and contain the other 7.
	for _, tag := range realPreimageTags {
		if tag == DomainAbsentChild {
			t.Fatalf("realPreimageTags must not include the absent-child tag 0x00")
		}
	}
	if len(realPreimageTags) != 7 {
		t.Fatalf("realPreimageTags has %d entries, want 7", len(realPreimageTags))
	}
}
