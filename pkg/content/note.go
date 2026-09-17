// NOTE base record (T-0368, FR-036; document.abnf S5, data-model.md S2.25).
// A note anchors at a single point, carries its own text block, and declares
// a footnote/endnote placement. Its anchor-boundary is present for shape
// uniformity but NORMATIVE-ignored (a note anchors at a point, not a range).
package content

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// NotePlacement is the closed 2-value note placement enum (document.abnf S5
// note-placement): footnote or endnote.
type NotePlacement uint8

const (
	// PlacementFootnote renders the note at the bottom of the page (0x00).
	PlacementFootnote NotePlacement = 0x00
	// PlacementEndnote renders the note at the end of the document (0x01).
	PlacementEndnote NotePlacement = 0x01
)

// Note is the decoded NOTE record.
type Note struct {
	ID        pdlfmt.UnitID
	Anchor    AnchorPoint
	BodyBlock pdlfmt.UnitID
	Placement NotePlacement
}

var (
	// ErrInvalidNotePlacement is returned for a note-placement octet outside
	// the closed set {0x00, 0x01}.
	ErrInvalidNotePlacement = errors.New("content: note-placement outside the closed set {0,1}")
	// ErrNoteBodyBlockUnresolved is returned when a note's body-block does
	// not resolve to a present TextBlock.
	ErrNoteBodyBlockUnresolved = errors.New("content: note body-block does not resolve to a present TextBlock")
	// ErrNoteTruncated is returned when the encoding ends prematurely.
	ErrNoteTruncated = errors.New("content: note encoding truncated")
)

// noteFixedLen: id(16) + anchor(22) + body-block(16) + placement(1) = 55.
const noteFixedLen = 16 + AnchorPointSize + 16 + 1

// EncodeNote appends a canonical fixed-layout encoding of n to dst.
func EncodeNote(dst []byte, n Note) []byte {
	dst = pdlfmt.AppendUnitID(dst, n.ID)
	dst = EncodeAnchorPoint(dst, n.Anchor)
	dst = pdlfmt.AppendUnitID(dst, n.BodyBlock)
	dst = append(dst, byte(n.Placement))
	return dst
}

// DecodeNote decodes and validates a Note from the leading noteFixedLen
// octets of src, rejecting an invalid anchor (side/boundary) and a placement
// value outside {0x00, 0x01}.
func DecodeNote(src []byte) (Note, int, error) {
	if len(src) < noteFixedLen {
		return Note{}, 0, ErrNoteTruncated
	}
	var n Note
	pos := 0
	id, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Note{}, 0, fmt.Errorf("content: note-id: %w", err)
	}
	n.ID = id
	pos += 16

	anchor, adv, err := DecodeAnchorPoint(src[pos:])
	if err != nil {
		return Note{}, 0, fmt.Errorf("content: note-anchor: %w", err)
	}
	n.Anchor = anchor
	pos += adv

	bb, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Note{}, 0, fmt.Errorf("content: note-body-block: %w", err)
	}
	n.BodyBlock = bb
	pos += 16

	placement := src[pos]
	if placement != byte(PlacementFootnote) && placement != byte(PlacementEndnote) {
		return Note{}, 0, fmt.Errorf("%w: got 0x%02x", ErrInvalidNotePlacement, placement)
	}
	n.Placement = NotePlacement(placement)
	pos++

	return n, pos, nil
}

// ResolveNoteBodyBlock checks that n's body-block resolves to a present
// TextBlock, given the set of TextBlock ids in the document, returning
// ErrNoteBodyBlockUnresolved naming the note otherwise (FR-036).
func ResolveNoteBodyBlock(n Note, textBlocks map[pdlfmt.UnitID]struct{}) error {
	if _, ok := textBlocks[n.BodyBlock]; !ok {
		return fmt.Errorf("%w: note %x body-block %x", ErrNoteBodyBlockUnresolved, n.ID, n.BodyBlock)
	}
	return nil
}
