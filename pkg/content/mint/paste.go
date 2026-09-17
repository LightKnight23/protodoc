// Fresh identity on duplicate/paste (T-0072, FR-022). A duplicate or paste
// operation produces new content and therefore must mint a FRESH run_id for
// every produced run rather than cloning the source run's identifier
// (DP-003 burst-boundary rule): a paste is treated as a new contiguous
// typing burst. Cloning identifiers would make an anchor resolve to two
// places at once, the commonest way identity models break in practice.
//
// This lives in the mint subpackage (not package content) because it needs
// crypto/rand-backed minting; keeping it out of package content keeps that
// package crypto-free so the extraction view can import content's types
// without transitively pulling in crypto (TR-011).
package mint

import (
	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

// DuplicateRange returns a copy of the source runs with a freshly minted
// run_id for every produced run: text, base_ordinal and other content ride
// through unchanged, but no produced run shares a run_id with any source run
// (FR-022). It returns an error only if the entropy source fails during
// minting. The source slice is not mutated.
func DuplicateRange(source []content.Run) ([]content.Run, error) {
	out := make([]content.Run, len(source))
	for i, r := range source {
		fresh, err := MintID()
		if err != nil {
			return nil, err
		}
		out[i] = content.Run{
			RunID:       fresh,
			BaseOrdinal: r.BaseOrdinal,
			Text:        r.Text,
			LangRef:     r.LangRef,
		}
	}
	return out, nil
}

// PasteContent produces runs to insert from a paste payload: it is
// DuplicateRange applied to the payload runs, so every pasted run carries a
// freshly minted run_id distinct from its source (FR-022).
func PasteContent(payload []content.Run) ([]content.Run, error) {
	return DuplicateRange(payload)
}

// sourceRunIDSet returns the set of run_ids present in runs, used by tests
// to check no produced id collides with a source id.
func sourceRunIDSet(runs []content.Run) map[pdlfmt.UnitID]struct{} {
	set := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		set[r.RunID] = struct{}{}
	}
	return set
}
