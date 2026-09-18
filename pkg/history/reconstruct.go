// State reconstruction (T-0212/T-0214, FR-059/FR-060). WHERE a document
// declares complete history, every state it has previously been published in
// is reconstructable octet-for-octet from the current file alone (FR-059); in
// retained-from-point mode, every state at or after the retention point is
// (FR-060). Reconstruction replays the recorded operation log deterministically
// from a base state up to a target state_id: because each operation records
// its authoring state (op-order-key) and applies a deterministic mutation, the
// content at any past state is a pure function of the log prefix ending at that
// state.
package history

import (
	"errors"
	"sort"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// StateSnapshot is a reconstructed document state: the content units present,
// keyed by unit-id, with their payload octets. Its canonical serialization
// (Octets) is the byte-for-byte content used to compare reconstructed states.
type StateSnapshot struct {
	units map[pdlfmt.UnitID][]byte
}

// newSnapshot builds an empty snapshot.
func newSnapshot() StateSnapshot {
	return StateSnapshot{units: map[pdlfmt.UnitID][]byte{}}
}

// clone returns a deep copy so replay to one target does not disturb another.
func (s StateSnapshot) clone() StateSnapshot {
	c := newSnapshot()
	for k, v := range s.units {
		c.units[k] = append([]byte(nil), v...)
	}
	return c
}

// apply mutates the snapshot by one operation deterministically.
func (s StateSnapshot) apply(op OperationRecord) {
	switch op.Kind {
	case OpSequencePositionClaim, OpValueClaim:
		// Insert or set the target unit's payload to the op payload.
		s.units[op.Target] = append([]byte(nil), op.Payload...)
	case OpDelete:
		delete(s.units, op.Target)
	}
}

// Octets returns the canonical byte-for-byte serialization of the snapshot:
// units in ascending unit-id order, each as unit-id || varint(len) || payload.
// The ordering is deterministic, so two reconstructions of the same state
// produce identical octets.
func (s StateSnapshot) Octets() []byte {
	ids := make([]pdlfmt.UnitID, 0, len(s.units))
	for id := range s.units {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		for k := range ids[i] {
			if ids[i][k] != ids[j][k] {
				return ids[i][k] < ids[j][k]
			}
		}
		return false
	})
	var out []byte
	for _, id := range ids {
		out = append(out, id[:]...)
		out = pdlfmt.AppendVarint(out, uint64(len(s.units[id])))
		out = append(out, s.units[id]...)
	}
	return out
}

// ErrTargetStateNotInLog is returned when a target state-id is never an
// authoring state (op-order-key) in the log, so it cannot be reconstructed.
var ErrTargetStateNotInLog = errors.New("history: target state-id is not a published state in the operation log")

// Reconstruct replays ops (in stored order) from base up to and INCLUDING the
// operation whose op-order-key is target, returning that state's snapshot. Each
// op applies deterministically; the log must be in a legal linearization
// (stored order preserved). The target must be an authoring state present in
// the log. This reconstructs a past state octet-for-octet from the log alone.
func Reconstruct(base StateSnapshot, ops []OperationRecord, target container.StateID) (StateSnapshot, error) {
	found := false
	for _, op := range ops {
		if op.OrderKey == target {
			found = true
			break
		}
	}
	if !found {
		return StateSnapshot{}, ErrTargetStateNotInLog
	}
	cur := base.clone()
	for _, op := range ops {
		cur.apply(op)
		if op.OrderKey == target {
			return cur, nil
		}
	}
	return cur, nil
}
