// Fresh identity on duplicate/paste (T-0072, FR-022). A duplicate or paste
// operation produces new content and therefore must mint a FRESH run_id for
// every produced run rather than cloning the source run's identifier
// (DP-003 burst-boundary rule): a paste is treated as a new contiguous
// typing burst. Cloning identifiers would make an anchor resolve to two
// places at once (a comment appearing on the wrong copy), the commonest way
// identity models break in practice.
package content

import "Protodoc/pkg/pdlfmt"

// DuplicateRange returns a copy of the source runs with a freshly minted
// run_id for every produced run: text, base_ordinal and other content ride
// through unchanged, but no produced run shares a run_id with any source run
// (FR-022). It returns an error only if the entropy source fails during
// minting. The source slice is not mutated.
func DuplicateRange(source []Run) ([]Run, error) {
	out := make([]Run, len(source))
	for i, r := range source {
		fresh, err := MintID()
		if err != nil {
			return nil, err
		}
		out[i] = Run{
			RunID:       fresh,
			BaseOrdinal: r.BaseOrdinal,
			Text:        r.Text,
		}
	}
	return out, nil
}

// PasteContent produces runs to insert from a paste payload: it is
// DuplicateRange applied to the payload runs, so every pasted run carries a
// freshly minted run_id distinct from its source (FR-022). Paste and
// duplicate share one implementation because they share one rule: produced
// content is new content and gets new identity.
func PasteContent(payload []Run) ([]Run, error) {
	return DuplicateRange(payload)
}

// sourceRunIDSet returns the set of run_ids present in runs, used by tests
// (and callers wanting to assert freshness) to check no produced id
// collides with a source id.
func sourceRunIDSet(runs []Run) map[pdlfmt.UnitID]struct{} {
	set := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		set[r.RunID] = struct{}{}
	}
	return set
}
