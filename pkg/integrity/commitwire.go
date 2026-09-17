// Commit-time state-identity wiring (T-0137, FR-001/FR-003): at every
// commit, CommitRingRecord.state-id-field (and the record's own t-c-root
// field) must be set to the FRESHLY computed T_C_root of the state being
// committed -- never a stale or separately-derived value. This is the
// concrete mechanism plan.md cites for both FR-001 (state identity exists)
// and FR-003 (identity differs whenever any value differs). The same fresh
// T_C_root computed here is the value that feeds the signed_object preimage
// (integrity.abnf S3.2); it is never a value cached from an earlier stage.
package integrity

import (
	"Protodoc/pkg/container"
)

// SetCommitRingStateID computes T_C_root fresh from records and writes it
// into rec.StateID (state-id-field, FR-003) and rec.TCRoot (the record's own
// t-c-root field), both of which are the document state identity. It
// recomputes T_C_root from the records' current stored frame bytes; it never
// copies a cached root. It returns the computed T_C_root and any capacity
// error from the tree builder.
func SetCommitRingStateID(rec *container.CommitRingRecord, records []ContentRecord) (Digest, error) {
	root, err := TCRoot(records)
	if err != nil {
		return Digest{}, err
	}
	rec.StateID = container.StateID(root)
	rec.TCRoot = [32]byte(root)
	return root, nil
}
