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
// carries no relationship to reading order. The proper fix — an additive
// ROOT_SEQUENCE record and a genuine reading-order traversal — now exists as
// OrderRecordsBySequence / TCRootWithSequence (FR-036, T-0275). OrderRecords
// is retained as the deterministic fallback used when no ROOT_SEQUENCE is
// present; callers that have a ROOT_SEQUENCE MUST key the traversal on it.
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
	// Salt is the 32-octet per-subtree CSPRNG salt used when Redactable is
	// true (integrity.abnf S2.2 tc-salt); ignored for non-redactable records.
	Salt [SaltSize]byte
	// Redacted marks a redactable record whose frame and salt have been
	// removed (the act of redaction); its leaf is then the RetainedLeaf
	// commitment digest, so the T_C root is preserved across redaction
	// (FR-076). Only a Redactable record may be Redacted.
	Redacted bool
	// RetainedLeaf is the bare 32-octet commitment digest retained after
	// redaction (the redactable leaf's hash output, salt and frame gone). It
	// is the leaf digest used when Redacted is true.
	RetainedLeaf Digest
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

// OrderRecordsBySequence returns records in T_C subtree ordinal order keyed on
// the document's authored ROOT_SEQUENCE reading order (FR-036, T-0275),
// replacing the FLAGGED CSPRNG-ordered interim rule of OrderRecords. The
// subtree ordinal of each record is its unit-id's position in `order` (the
// rs-order list of the ROOT_SEQUENCE record, document.abnf S5.1). Because the
// traversal now derives purely from ROOT_SEQUENCE, the T_C root changes when —
// and only when — ROOT_SEQUENCE changes: a mere permutation of storage order
// leaves both the ordering and the root unchanged, whereas reordering
// rs-order re-keys the traversal and so changes the root.
//
// It returns a new slice; the input is not mutated. A record whose unit-id is
// absent from `order` sorts after all sequenced records, keyed by ascending
// byte-lexicographic unit-id, so the function is total and deterministic even
// on an incomplete sequence (completeness itself is enforced separately by
// semantics.PD-A11Y-005 / T-0274).
func OrderRecordsBySequence(records []ContentRecord, order []pdlfmt.UnitID) []ContentRecord {
	pos := make(map[pdlfmt.UnitID]int, len(order))
	for i, u := range order {
		if _, dup := pos[u]; !dup {
			pos[u] = i
		}
	}
	out := make([]ContentRecord, len(records))
	copy(out, records)
	sort.SliceStable(out, func(i, j int) bool {
		pi, iok := pos[out[i].UnitID]
		pj, jok := pos[out[j].UnitID]
		switch {
		case iok && jok:
			return pi < pj
		case iok != jok:
			return iok // sequenced records sort before unsequenced ones
		default:
			return bytes.Compare(out[i].UnitID[:], out[j].UnitID[:]) < 0
		}
	})
	return out
}
