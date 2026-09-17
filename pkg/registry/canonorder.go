package registry

import (
	"bytes"
	"sort"

	"Protodoc/pkg/content"
)

// Canonical order among unrecognised constructs at one position (T-0058,
// FR-012; document.abnf S6 ext-position-key note, DP-011). When more than one
// ExtensionEnvelope claims the same position, their canonical order is by the
// pair (ext-tok, ext-position-key): first byte-lexicographically by the
// 8-octet ext-token, then, for equal tokens, by the 22-octet anchor-point
// encoding of the position key. This gives two independent implementations an
// identical, deterministic ordering (a signing prerequisite), and it treats an
// envelope insertion as a position-claim resolved the same way as a text or
// table-row insertion (R1 SEQUENCE-ORDER).

// CompareEnvelopeCanonical returns -1, 0, or +1 ordering a and b by the
// DP-011 canonical key (ext-tok, then ext-position-key), each compared as
// unsigned byte-lexicographic over its canonical encoding.
func CompareEnvelopeCanonical(a, b ExtEnvelope) int {
	if c := bytes.Compare(a.Tok[:], b.Tok[:]); c != 0 {
		return c
	}
	ak := content.EncodeAnchorPoint(nil, a.PositionKey)
	bk := content.EncodeAnchorPoint(nil, b.PositionKey)
	return bytes.Compare(ak, bk)
}

// SortEnvelopesCanonical sorts envelopes into DP-011 canonical order in place,
// using a stable sort so envelopes that compare equal (identical tok and
// position key) keep their input relative order. The sort is total and
// deterministic: the same input multiset always yields the same sequence.
func SortEnvelopesCanonical(envelopes []ExtEnvelope) {
	sort.SliceStable(envelopes, func(i, j int) bool {
		return CompareEnvelopeCanonical(envelopes[i], envelopes[j]) < 0
	})
}
