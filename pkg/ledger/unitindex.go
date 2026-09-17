// UnitIndexLeaf: the derived, non-normative, rebuilt-on-open index leaf
// that maps a content unit's identity to its absolute offset, length, kind
// and digest (T-0041, FR-055; document.abnf S8 unit-index-leaf,
// data-model.md S2.23 UnitIndex). It is the lookup structure FR-055's
// bounded 3-dependent-read path depends on (wired in T-0042). Being
// derived, it carries no persisted-only field: every entry is recomputable
// from the authoritative content alone, so a reader MAY rebuild it from
// scratch rather than trusting a stored copy, and a rebuild always
// reproduces the same leaf.
package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"Protodoc/pkg/pdlfmt"
)

// IndexRouteBuckets is the number of index-route buckets (container.abnf S3
// index-route = 16u16): a unit is routed to bucket unit-id[0] >> 4.
const IndexRouteBuckets = 16

// AddressableUnit is one content unit as the index sees it: the authoritative
// facts a UnitIndexLeaf entry is derived from. Every field here is read from
// the content itself (identity minted with the unit, offset/length from its
// placement, kind from its record, digest over its octets), so a leaf built
// from these is a pure function of content.
type AddressableUnit struct {
	Identity pdlfmt.UnitID
	Offset   uint64 // absolute file offset (u48 range)
	Length   uint32
	Kind     uint16
	Digest   pdlfmt.Digest256
}

// UnitIndexEntry is one derived leaf entry (document.abnf S8
// unit-index-entry): identity, absolute offset, length, kind and digest.
// It has no field that is not recomputable from content.
type UnitIndexEntry struct {
	Identity pdlfmt.UnitID
	Offset   uint64
	Length   uint32
	Kind     uint16
	Digest   pdlfmt.Digest256
}

// UnitIndexLeaf is a derived index leaf: the entries for one index-route
// bucket, sorted by unit-id octets (document.abnf S8 uil-entries), all
// sharing the bucket their unit-id[0]>>4 selects.
type UnitIndexLeaf struct {
	// Bucket is the index-route bucket [0,16) every entry belongs to.
	Bucket uint8
	// Entries are sorted ascending by Identity's raw octets.
	Entries []UnitIndexEntry
}

var (
	// ErrEmptyUnitIndexInput is returned when BuildUnitIndexLeaves is given
	// no units: there is nothing to index.
	ErrEmptyUnitIndexInput = errors.New("ledger: no addressable units to index")

	// ErrDuplicateUnitIdentity is returned when two addressable units share
	// an identity: a content-unit identity is unique within a document (the
	// document-wide identity-uniqueness rule, enforced in its own right by a
	// later M04 identity task), so a duplicate is a malformed input, not a
	// valid index.
	ErrDuplicateUnitIdentity = errors.New("ledger: duplicate content-unit identity in index input")
)

// bucketOf returns the index-route bucket a unit-id routes to: the high
// nibble of its first octet (container.abnf S3 index-route).
func bucketOf(id pdlfmt.UnitID) uint8 {
	return id[0] >> 4
}

// BuildUnitIndexLeaves constructs the derived UnitIndexLeaf set from the
// authoritative addressable units, one leaf per non-empty index-route
// bucket, each leaf's entries sorted by unit-id octets. The result is a
// pure function of units: the same units in any input order produce the
// same leaves (entries are sorted; leaves are returned in ascending bucket
// order). It carries no persisted-only state, so BuildUnitIndexLeaves is
// exactly the rebuild-from-content operation FR-055's derived index relies
// on. A duplicate identity is rejected (the document-wide
// identity-uniqueness rule).
func BuildUnitIndexLeaves(units []AddressableUnit) ([]UnitIndexLeaf, error) {
	if len(units) == 0 {
		return nil, ErrEmptyUnitIndexInput
	}

	seen := make(map[pdlfmt.UnitID]bool, len(units))
	byBucket := make(map[uint8][]UnitIndexEntry)
	for _, u := range units {
		if seen[u.Identity] {
			return nil, fmt.Errorf("%w: %x", ErrDuplicateUnitIdentity, u.Identity)
		}
		seen[u.Identity] = true

		b := bucketOf(u.Identity)
		byBucket[b] = append(byBucket[b], UnitIndexEntry{
			Identity: u.Identity,
			Offset:   u.Offset,
			Length:   u.Length,
			Kind:     u.Kind,
			Digest:   u.Digest,
		})
	}

	leaves := make([]UnitIndexLeaf, 0, len(byBucket))
	for b := 0; b < IndexRouteBuckets; b++ {
		entries, ok := byBucket[uint8(b)]
		if !ok {
			continue
		}
		sort.Slice(entries, func(i, j int) bool {
			return bytes.Compare(entries[i].Identity[:], entries[j].Identity[:]) < 0
		})
		leaves = append(leaves, UnitIndexLeaf{Bucket: uint8(b), Entries: entries})
	}
	return leaves, nil
}

// Lookup finds the entry for id within this leaf via binary search over its
// sorted entries, returning the entry and whether it was found. It assumes
// id routes to this leaf's bucket (the caller reaches a leaf via
// index-route); an id from another bucket simply will not be found.
func (l UnitIndexLeaf) Lookup(id pdlfmt.UnitID) (UnitIndexEntry, bool) {
	i := sort.Search(len(l.Entries), func(i int) bool {
		return bytes.Compare(l.Entries[i].Identity[:], id[:]) >= 0
	})
	if i < len(l.Entries) && l.Entries[i].Identity.Equal(id) {
		return l.Entries[i], true
	}
	return UnitIndexEntry{}, false
}
