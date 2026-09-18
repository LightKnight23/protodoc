// PageDirectory non-normativity and staleness refusal (T-0244; FR-087, FR-088).
// The PageDirectory is a DERIVED artefact and is declared non-normative
// relative to the authoritative CONTENT it is derived from (FR-087). On open,
// if its stored input digest does not match a fresh recomputation over the
// current inputs, the stored directory is refused and re-derived from the
// authoritative content BEFORE any page renders (FR-088) -- serving a stale
// derived artefact is both a correctness failure and a disclosure channel.
package render

import (
	"Protodoc/pkg/pdlfmt"
)

// Normativity classifies an artefact relative to authoritative content. It is a
// closed 2-value enum: exactly one representation of a capability is
// authoritative, and every derived artefact is non-normative (FR-087).
type Normativity uint8

const (
	// Normative: authoritative content; the source of truth.
	Normative Normativity = iota
	// NonNormative: a derived, reader-side artefact (e.g. PageDirectory) that
	// competes with no authoritative content and is rebuilt when stale.
	NonNormative
)

// Normativity reports the PageDirectory's classification: it is always
// non-normative relative to the CONTENT it was derived from (FR-087).
func (*PageDirectory) Normativity() Normativity { return NonNormative }

// StoredPageDirectory is a PageDirectory as it sits in a file on open: the
// entries plus the input digest recorded when it was last derived. The stored
// digest is never trusted without a fresh recomputation (FR-088).
type StoredPageDirectory struct {
	Entries      []PageEntry
	StoredDigest pdlfmt.Digest256
}

// Deriver rebuilds a PageDirectory from the authoritative content inputs. In a
// full reader this walks the CONTENT segments; here it is the pluggable
// re-derivation used when the stored directory is stale or absent.
type Deriver func(inputs PageDirectoryInputs) *PageDirectory

// OpenPageDirectory returns the PageDirectory a reader must use, enforcing
// FR-088: it recomputes the input digest over the CURRENT inputs and compares
// it to the stored digest. If they match, the stored directory is served as
// is. If they differ (stale) OR the stored directory is nil (absent), the
// stored directory is DISCARDED and rebuilt from the authoritative content via
// derive, before any page render. The bool result reports whether a
// re-derivation happened. A nil derive with a stale/absent directory yields a
// nil directory (the reader falls back to deriving on demand).
func OpenPageDirectory(stored *StoredPageDirectory, current PageDirectoryInputs, derive Deriver) (*PageDirectory, bool) {
	fresh := current.Digest()
	if stored != nil && stored.StoredDigest == fresh {
		// Current: serve the stored directory.
		d := &PageDirectory{}
		for _, e := range stored.Entries {
			d.Insert(e)
		}
		return d, false
	}
	// Stale or absent: refuse and re-derive from authoritative content.
	if derive == nil {
		return nil, true
	}
	return derive(current), true
}
