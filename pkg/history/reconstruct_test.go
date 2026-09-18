package history

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_059_CompleteHistoryReconstructsEveryState is T-0212's named
// integration test (FR-059). In complete-history mode, every previously
// published state is reconstructable octet-for-octet from the current file's
// operation log alone. This replays the log to each published state and
// asserts the reconstruction is deterministic (identical octets on repeat) and
// reflects that state's content (not the final content).
func TestFR_059_CompleteHistoryReconstructsEveryState(t *testing.T) {
	// Confirm complete-history mode is the declared mode.
	d, err := NewDeclaration(ModeComplete)
	if err != nil || d.Mode() != ModeComplete {
		t.Fatalf("NewDeclaration(ModeComplete): %v", err)
	}

	u1, u2 := hUnitID(0x01), hUnitID(0x02)
	s1, s2, s3, s4 := hStateID(0x01), hStateID(0x02), hStateID(0x03), hStateID(0x04)

	// s1: set u1="alpha"; s2: set u2="beta"; s3: set u1="ALPHA2"; s4: delete u2.
	ops := []OperationRecord{
		{Kind: OpValueClaim, Target: u1, OrderKey: s1, Predecessor: container.StateID{}, Payload: []byte("alpha")},
		{Kind: OpValueClaim, Target: u2, OrderKey: s2, Predecessor: s1, Payload: []byte("beta")},
		{Kind: OpValueClaim, Target: u1, OrderKey: s3, Predecessor: s2, Payload: []byte("ALPHA2")},
		{Kind: OpDelete, Target: u2, OrderKey: s4, Predecessor: s3},
	}
	base := newSnapshot()

	// Expected content at each published state.
	expect := map[container.StateID]map[string]string{
		s1: {string(u1[:]): "alpha"},
		s2: {string(u1[:]): "alpha", string(u2[:]): "beta"},
		s3: {string(u1[:]): "ALPHA2", string(u2[:]): "beta"},
		s4: {string(u1[:]): "ALPHA2"},
	}

	prevOctets := map[container.StateID][]byte{}
	for _, target := range []container.StateID{s1, s2, s3, s4} {
		snap, err := Reconstruct(base, ops, target)
		if err != nil {
			t.Fatalf("Reconstruct(%x): %v", target, err)
		}
		// Content matches the expected set at that state.
		want := expect[target]
		if len(snap.units) != len(want) {
			t.Errorf("state %x: %d units, want %d", target, len(snap.units), len(want))
		}
		for k, v := range want {
			got, ok := snap.units[unitFromKey(k)]
			if !ok || string(got) != v {
				t.Errorf("state %x: unit payload = %q (present=%v), want %q", target, got, ok, v)
			}
		}
		// Deterministic: reconstruct again, octets identical.
		snap2, _ := Reconstruct(base, ops, target)
		if !bytes.Equal(snap.Octets(), snap2.Octets()) {
			t.Errorf("state %x: reconstruction is not deterministic", target)
		}
		prevOctets[target] = snap.Octets()
	}

	// Distinct states produce distinct octets (a past state is not the final
	// state): s2 (u1=alpha,u2=beta) differs from s3 (u1=ALPHA2,u2=beta).
	if bytes.Equal(prevOctets[s2], prevOctets[s3]) {
		t.Error("reconstructed s2 and s3 are identical; a past state was not reconstructed distinctly")
	}
	// s4 has u2 deleted, differs from s3.
	if bytes.Equal(prevOctets[s3], prevOctets[s4]) {
		t.Error("reconstructed s3 and s4 are identical despite the delete")
	}

	// A target that is not a published state errors.
	if _, err := Reconstruct(base, ops, hStateID(0x99)); !errors.Is(err, ErrTargetStateNotInLog) {
		t.Errorf("unknown target: err = %v, want ErrTargetStateNotInLog", err)
	}
}

// unitFromKey rebuilds a UnitID from its string(bytes) map key.
func unitFromKey(k string) (id pdlfmt.UnitID) {
	copy(id[:], k)
	return id
}
