// Whole-document content decode (closes GAP-VERIFY-CONTENT-REBUILD). This file
// reads every CONTENT segment body and decodes each content-model frame into
// its AUTHORED unit-id plus its canonical frame octets. It is the missing
// decode path that lets verify recompute the content-commitment tree over real
// authored ids, and lets project/redact key constructs by authored unit-id
// rather than a slot-digest surrogate.
//
// Every content-model record shares one wire shape at this level of detail
// (document.abnf S2-S5): a one-octet discriminant (field tag 0) followed by a
// PDL-TLV record whose field tag 1 carries the record's own unit-id. This
// decoder reads that shared shape only; it never constructs a font, image,
// shaping, layout, crypto or merge facility (CP-006), staying the small
// dependency-isolated reader the extract module promises.
package extract

import (
	"errors"
	"fmt"
	"io"

	"Protodoc/pkg/pdlfmt"
)

// ContentUnit is one decoded content-model frame: its storage ordinal, its
// AUTHORED unit-id (the field tag=1 value), and its canonical frame octets
// (the exact bytes the segment carried, suitable for a content-tree leaf).
type ContentUnit struct {
	Ordinal uint64
	UnitID  pdlfmt.UnitID
	Frame   []byte
}

// contentDiscriminantTag is the field tag of the leading discriminant octet;
// unitIDTag is the field tag every content-model record uses for its own
// unit-id (document.abnf S2-S5: tb-block-id, tbl-id, note-id, xref-id, ann-id,
// rs-id, r2-id all live at tag 1).
const (
	unitIDTag = 1
)

// ErrNoUnitIDField is returned when a content frame carries no tag=1 unit-id
// field (a structurally malformed content record; not silently tolerated).
var ErrNoUnitIDField = errors.New("extract: content frame has no unit-id (tag=1) field")

// DecodeContentFrameUnitID decodes the authored unit-id from a single content
// frame's octets. The first octet is the discriminant (tag 0); the remainder
// is a PDL-TLV record. It returns the tag=1 field's value as the unit-id.
//
// It accepts any field tag the frame carries (known = nil, reservedFrom = 0):
// this reader only needs the unit-id, so it does not impose a per-kind schema
// here (each record kind's full schema is validated by its own package). It
// still inherits DecodeRecord's structural guarantees (ascending tags, field
// ceiling, no over-long field).
func DecodeContentFrameUnitID(frame []byte) (pdlfmt.UnitID, error) {
	if len(frame) < 1 {
		return pdlfmt.UnitID{}, errors.New("extract: empty content frame")
	}
	// frame[0] is the discriminant (tag 0); the record body follows.
	body := frame[1:]
	fields, err := pdlfmt.DecodeRecord(body, nil, 0)
	if err != nil {
		return pdlfmt.UnitID{}, fmt.Errorf("extract: decoding content frame record: %w", err)
	}
	for _, f := range fields {
		if f.Tag == unitIDTag {
			id, _, derr := pdlfmt.DecodeUnitID(f.Value)
			if derr != nil {
				return pdlfmt.UnitID{}, fmt.Errorf("extract: content frame unit-id: %w", derr)
			}
			return id, nil
		}
	}
	return pdlfmt.UnitID{}, ErrNoUnitIDField
}

// LoadContentRecords reads every CONTENT segment body from r and decodes each
// into a ContentUnit (authored unit-id + canonical frame octets), in the
// document's storage-ordinal walk order. It reads only the prefix plus the
// CONTENT segment bodies (never RESOURCE/HISTORY/ATTEST bodies), reusing the
// bounded streaming Walk. A frame that fails to decode is a hard error (no
// partial/best-effort output, FR-103), returned to the caller rather than
// skipped.
func LoadContentRecords(r io.ReaderAt) ([]ContentUnit, error) {
	var out []ContentUnit
	err := Walk(r, func(seg ContentSegment, sr io.Reader) error {
		frame := make([]byte, seg.Length)
		if _, rerr := io.ReadFull(sr, frame); rerr != nil {
			return fmt.Errorf("extract: reading CONTENT segment %d: %w", seg.Ordinal, rerr)
		}
		id, derr := DecodeContentFrameUnitID(frame)
		if derr != nil {
			return fmt.Errorf("extract: CONTENT segment %d: %w", seg.Ordinal, derr)
		}
		out = append(out, ContentUnit{Ordinal: seg.Ordinal, UnitID: id, Frame: frame})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
