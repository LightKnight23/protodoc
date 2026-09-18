package history

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_061_ErasureEnumeratedOnRetentionAdvance is T-0218's named integration
// test (FR-061). When the retention point advances, every previously published
// state that becomes unreconstructable (its ordinal now falls below the new
// point) is enumerated as an ErasureRecord with its identity and salted
// commitment, its salt destroyed.
func TestFR_061_ErasureEnumeratedOnRetentionAdvance(t *testing.T) {
	dg := func(b byte) pdlfmt.Digest256 { var d pdlfmt.Digest256; d[0] = b; return d }
	sl := func(b byte) [SaltSize]byte { var s [SaltSize]byte; s[0] = b; return s }

	// States at ordinals 1..5.
	severed := []SeveredState{
		{Identity: hUnitID(0x01), Digest: dg(0x11), Ordinal: 1, Salt: sl(0xA1)},
		{Identity: hUnitID(0x02), Digest: dg(0x22), Ordinal: 2, Salt: sl(0xA2)},
		{Identity: hUnitID(0x03), Digest: dg(0x33), Ordinal: 3, Salt: sl(0xA3)},
		{Identity: hUnitID(0x04), Digest: dg(0x44), Ordinal: 4, Salt: sl(0xA4)},
		{Identity: hUnitID(0x05), Digest: dg(0x55), Ordinal: 5, Salt: sl(0xA5)},
	}

	// Retention advances from 1 to 4: states at ordinals 1,2,3 (in [1,4))
	// become unreconstructable and must be enumerated.
	records, err := EnumerateErasuresOnRetentionAdvance(1, 4, severed)
	if err != nil {
		t.Fatalf("EnumerateErasuresOnRetentionAdvance: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("enumerated %d erasures, want 3 (ordinals 1,2,3)", len(records))
	}
	gotIDs := map[pdlfmt.UnitID]bool{}
	for _, r := range records {
		gotIDs[r.Identity] = true
		// Each record's salt is destroyed and its commitment matches the
		// severance commitment of its digest under the (now-destroyed) salt.
		if !r.SaltDestroyed() {
			t.Errorf("erasure record %x salt not destroyed", r.Identity)
		}
	}
	for _, ord := range []byte{0x01, 0x02, 0x03} {
		if !gotIDs[hUnitID(ord)] {
			t.Errorf("state at ordinal %d not enumerated as erased", ord)
		}
	}
	// States 4,5 (at or after the new point) are NOT enumerated.
	if gotIDs[hUnitID(0x04)] || gotIDs[hUnitID(0x05)] {
		t.Error("a retained state was wrongly enumerated as erased")
	}

	// The commitment for the ordinal-1 state binds its digest under its salt.
	if records[0].Commitment != SeveranceCommitment(sl(0xA1), dg(0x11)) {
		// records order follows `severed` order; ordinal 1 is first.
		t.Error("erasure commitment does not match the severed content's salted commitment")
	}

	// A non-advancing retention "advance" is refused.
	if _, err := EnumerateErasuresOnRetentionAdvance(4, 4, severed); !errors.Is(err, ErrRetentionNotAdvanced) {
		t.Errorf("non-advancing retention: err = %v, want ErrRetentionNotAdvanced", err)
	}
}
