// Real project backend (T-0376, DEFECT-2026-09-19 fix). Replaces the no-op
// ProjectStateFor stub with a production backend that OPENS the file, runs the
// CP-006 validate-first precondition, reads the real CONTENT segment bodies,
// and assembles a canon.Document from them so the real M17 canon.ProjectText /
// canon.ProjectHTML projectors run over actual content. Go stdlib only.
//
// Scope (honest): each content subtree's identity is keyed by the segment's
// own content-addressed slot digest (a real, deterministic 32-octet value from
// the prefix), NOT by decoding the authored unit-id out of each frame — that
// authored-unit-id decode is the same missing capability recorded as
// GAP-VERIFY-CONTENT-REBUILD in specs/CHANGES.md. The projection is therefore
// genuinely content-derived and deterministic, but keyed by slot digest rather
// than authored unit-id until that decode path exists.
package cli

import (
	"errors"
	"io"
	"os"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
	"Protodoc/pkg/validate"
)

// realProjectStateFor implements the production project backend.
func realProjectStateFor(path string) (*canon.Document, error) {
	// CP-006 precondition.
	if steps, _ := realValidateStepsFor(path); validate.Run(steps).Validity != nil {
		return nil, errors.New("project: document failed structural validation (CP-006)")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the prefix to get the segment table (for slot digests as identities).
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return nil, err
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return nil, err
	}
	digestByOrdinal := map[uint64][32]byte{}
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeContent {
			digestByOrdinal[uint64(i)] = slot.Digest
		}
	}

	// Walk the real CONTENT segment bodies (frames) via the M05 streaming walk.
	doc := &canon.Document{}
	err = extract.Walk(f, func(seg extract.ContentSegment, r io.Reader) error {
		frame := make([]byte, seg.Length)
		if _, rerr := io.ReadFull(r, frame); rerr != nil {
			return rerr
		}
		var id pdlfmt.UnitID
		d := digestByOrdinal[seg.Ordinal]
		copy(id[:], d[:16]) // content-addressed identity from the slot digest
		doc.Subtrees = append(doc.Subtrees, canon.ContentSubtree{UnitID: id, Frame: frame})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func init() {
	ProjectStateFor = realProjectStateFor
}
