package content

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_036_NoteBodyBlockResolvesToTextBlock is T-0368's named test. A NOTE
// round-trips byte-exact for both placement values; a note_body_block not
// resolving to a present TextBlock is rejected; a note_placement value
// outside {0x00, 0x01} is rejected.
func TestFR_036_NoteBodyBlockResolvesToTextBlock(t *testing.T) {
	noteID, _ := MintID()
	bodyBlock, _ := MintID()
	anchorRun, _ := MintID()

	for _, placement := range []NotePlacement{PlacementFootnote, PlacementEndnote} {
		n := Note{
			ID:        noteID,
			Anchor:    AnchorPoint{RunID: anchorRun, BirthOrdinal: 7, Side: SideBefore, Boundary: BoundaryInside},
			BodyBlock: bodyBlock,
			Placement: placement,
		}
		enc := EncodeNote(nil, n)
		got, adv, err := DecodeNote(enc)
		if err != nil {
			t.Fatalf("placement %d: DecodeNote: %v", placement, err)
		}
		if adv != len(enc) {
			t.Fatalf("placement %d: consumed %d of %d", placement, adv, len(enc))
		}
		if got != n {
			t.Fatalf("placement %d: note did not round-trip: got %+v want %+v", placement, got, n)
		}
		if !bytes.Equal(EncodeNote(nil, got), enc) {
			t.Fatalf("placement %d: note not byte-exact on re-encode", placement)
		}

		// Body-block resolves against a document containing it.
		present := map[pdlfmt.UnitID]struct{}{bodyBlock: {}}
		if err := ResolveNoteBodyBlock(got, present); err != nil {
			t.Fatalf("placement %d: body-block should resolve: %v", placement, err)
		}
		// Body-block not present is rejected naming the note.
		if err := ResolveNoteBodyBlock(got, map[pdlfmt.UnitID]struct{}{}); !errors.Is(err, ErrNoteBodyBlockUnresolved) {
			t.Fatalf("placement %d: unresolved body-block returned %v, want ErrNoteBodyBlockUnresolved", placement, err)
		}
	}

	// A placement value outside {0x00, 0x01} is rejected on decode.
	n := Note{ID: noteID, Anchor: AnchorPoint{RunID: anchorRun, Side: SideAfter, Boundary: BoundaryOutside}, BodyBlock: bodyBlock, Placement: PlacementFootnote}
	enc := EncodeNote(nil, n)
	enc[len(enc)-1] = 0x02 // corrupt placement to a reserved value
	if _, _, err := DecodeNote(enc); !errors.Is(err, ErrInvalidNotePlacement) {
		t.Fatalf("reserved placement 0x02 returned %v, want ErrInvalidNotePlacement", err)
	}
	// Truncated encoding is rejected.
	if _, _, err := DecodeNote(enc[:noteFixedLen-1]); !errors.Is(err, ErrNoteTruncated) {
		t.Fatalf("truncated note returned %v, want ErrNoteTruncated", err)
	}
}
