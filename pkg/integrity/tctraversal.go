// T_C subtree traversal order (T-0134, FR-003; integrity.abnf S2.2.1).
//
// FLAGGED INTERIM RULE (carried forward verbatim from integrity.abnf
// S2.2.1's self-disclosed gap): the subtree ordinal order is ascending
// unsigned byte-lexicographic order of each content-model record's own
// unit-id. This satisfies the stated NEGATIVE constraints exactly -- a
// unit-id is neither a storage ordinal nor a digest, it is a CSPRNG-minted
// token independent of both -- and is fully deterministic and reproducible.
// It does NOT satisfy the aspirational POSITIVE requirement ("fixed by the
// document's logical/reading-order structure"): a CSPRNG-ordered traversal
// carries no relationship to reading order. This is deliberately NOT
// "fixed" here by inventing a reading-order traversal: the proper fix is an
// additive ROOT_SEQUENCE record and a genuine pre-order walk, future work
// out of this task's scope. Do not replace this interim rule silently.
package integrity

import (
	"bytes"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// ContentRecord is one content-model record as T_C sees it: its unit-id
// (which fixes its subtree ordinal under the interim rule), its stored
// PDL-TLV frame bytes (committed verbatim by the leaf), and whether it was
// designated redactable at signing time.
type ContentRecord struct {
	UnitID     pdlfmt.UnitID
	Frame      []byte
	Redactable bool
}

// OrderRecords returns records in T_C subtree ordinal order: ascending
// unsigned byte-lexicographic order of each record's unit-id (the FLAGGED
// interim rule). It returns a new slice; the input is not mutated. The
// ordering is a pure, deterministic function of the unit-ids, so two
// invocations on any permutation of the same records produce the identical
// order.
func OrderRecords(records []ContentRecord) []ContentRecord {
	out := make([]ContentRecord, len(records))
	copy(out, records)
	sort.SliceStable(out, func(i, j int) bool {
		return bytes.Compare(out[i].UnitID[:], out[j].UnitID[:]) < 0
	})
	return out
}
