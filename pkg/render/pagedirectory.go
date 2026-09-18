// Package render implements Protodoc's rendering layer: the derived,
// non-normative PageDirectory (rank-augmented over content identity, not
// absolute page ordinals), font records, fixed pagination, the exact-rational
// rasterizer, Knuth-Plass reflow, deterministic shaping via a pinned oracle,
// the PLP-1 and restricted-PNG codecs, resource substitution, and the render
// report. It builds on pkg/content (unit identity), pkg/container (segment
// table), and pkg/integrity (input-digest binding).
package render

import (
	"Protodoc/pkg/pdlfmt"
)

// MaxPages is the PageDirectory ceiling (data-model.md 2.22).
const MaxPages = 131072

// PageEntry is one PageDirectory entry, KEYED by the content identity of its
// page-break unit -- never by an absolute page ordinal (NFR-010). The page's
// ordinal is DERIVED as the rank of the entry within the directory, so a
// single-page content change touches only that entry, not every later ordinal.
type PageEntry struct {
	// BreakUnit is the content-unit identity of the unit at which this page
	// begins (the page-break unit). This is the entry's key.
	BreakUnit pdlfmt.UnitID
	// Geometry is the page's declared geometry token (opaque here).
	Geometry uint32
}

// pdNode is a rank-augmented binary-search-tree node over BreakUnit identity.
// subtreeSize supports O(log n) rank queries and, critically, an insert/remove
// at a position rewrites only the nodes on the touched root-to-leaf path --
// the other entries' identities and stored fields are untouched (NFR-010).
type pdNode struct {
	entry       PageEntry
	left, right *pdNode
	subtreeSize int
}

// PageDirectory is a derived, non-normative rank-augmented structure over
// content-unit identity (data-model.md 2.22). Absolute page ordinals are never
// stored; a page's ordinal is the rank of its entry.
type PageDirectory struct {
	root *pdNode
	// touched records the BreakUnit identities whose node was written by the
	// most recent mutation, for the NFR-010 "touches only the affected entry"
	// assertion. It is reset at the start of each mutation.
	touched []pdlfmt.UnitID
}

// Len returns the number of pages (entries).
func (d *PageDirectory) Len() int { return size(d.root) }

func size(n *pdNode) int {
	if n == nil {
		return 0
	}
	return n.subtreeSize
}

// less orders entries by BreakUnit identity byte-lexicographically.
func less(a, b pdlfmt.UnitID) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// Insert adds (or updates) a page entry keyed by its break-unit identity,
// writing only the nodes on the touched path (NFR-010). It returns false if the
// directory is at MaxPages and the key is new.
func (d *PageDirectory) Insert(e PageEntry) bool {
	d.touched = d.touched[:0]
	if size(d.root) >= MaxPages {
		if _, exists := d.lookupNode(e.BreakUnit); !exists {
			return false
		}
	}
	d.root = d.insert(d.root, e)
	return true
}

func (d *PageDirectory) insert(n *pdNode, e PageEntry) *pdNode {
	if n == nil {
		d.touched = append(d.touched, e.BreakUnit)
		return &pdNode{entry: e, subtreeSize: 1}
	}
	switch {
	case less(e.BreakUnit, n.entry.BreakUnit):
		n.left = d.insert(n.left, e)
	case less(n.entry.BreakUnit, e.BreakUnit):
		n.right = d.insert(n.right, e)
	default:
		// Same key: update in place (only this node touched).
		n.entry = e
		d.touched = append(d.touched, e.BreakUnit)
		return n
	}
	n.subtreeSize = 1 + size(n.left) + size(n.right)
	return n
}

// Remove deletes the entry with the given break-unit identity, writing only the
// touched path. It returns false if the key is absent.
func (d *PageDirectory) Remove(key pdlfmt.UnitID) bool {
	d.touched = d.touched[:0]
	if _, ok := d.lookupNode(key); !ok {
		return false
	}
	d.root = d.remove(d.root, key)
	return true
}

func (d *PageDirectory) remove(n *pdNode, key pdlfmt.UnitID) *pdNode {
	if n == nil {
		return nil
	}
	switch {
	case less(key, n.entry.BreakUnit):
		n.left = d.remove(n.left, key)
	case less(n.entry.BreakUnit, key):
		n.right = d.remove(n.right, key)
	default:
		d.touched = append(d.touched, key)
		if n.left == nil {
			return n.right
		}
		if n.right == nil {
			return n.left
		}
		// Replace with in-order successor (smallest in right subtree).
		succ := n.right
		for succ.left != nil {
			succ = succ.left
		}
		n.entry = succ.entry
		n.right = d.remove(n.right, succ.entry.BreakUnit)
	}
	n.subtreeSize = 1 + size(n.left) + size(n.right)
	return n
}

// Rank returns the zero-based page ordinal DERIVED for the given break-unit
// identity (the number of entries ordered before it), and whether the key is
// present.
func (d *PageDirectory) Rank(key pdlfmt.UnitID) (int, bool) {
	n := d.root
	rank := 0
	for n != nil {
		switch {
		case less(key, n.entry.BreakUnit):
			n = n.left
		case less(n.entry.BreakUnit, key):
			rank += size(n.left) + 1
			n = n.right
		default:
			return rank + size(n.left), true
		}
	}
	return 0, false
}

func (d *PageDirectory) lookupNode(key pdlfmt.UnitID) (PageEntry, bool) {
	n := d.root
	for n != nil {
		switch {
		case less(key, n.entry.BreakUnit):
			n = n.left
		case less(n.entry.BreakUnit, key):
			n = n.right
		default:
			return n.entry, true
		}
	}
	return PageEntry{}, false
}

// Touched returns the break-unit identities whose node was written by the most
// recent Insert/Remove -- used to prove a single-page change touched only the
// affected entry (NFR-010).
func (d *PageDirectory) Touched() []pdlfmt.UnitID { return d.touched }

// Entries returns all entries in ascending break-unit order (ordinal order).
func (d *PageDirectory) Entries() []PageEntry {
	var out []PageEntry
	var walk func(*pdNode)
	walk = func(n *pdNode) {
		if n == nil {
			return
		}
		walk(n.left)
		out = append(out, n.entry)
		walk(n.right)
	}
	walk(d.root)
	return out
}
