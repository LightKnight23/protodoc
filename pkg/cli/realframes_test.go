package cli

import (
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// realFrame builds a minimal well-formed content-model frame: a discriminant
// octet (tag 0) followed by a tag=1 unit-id field.
func realFrame(discriminant byte, id pdlfmt.UnitID) []byte {
	f := []byte{discriminant}
	f = pdlfmt.AppendField(f, pdlfmt.Field{Tag: 1, Value: id[:]})
	return f
}

// cliUnit builds a distinct unit-id seeded by b.
func cliUnit(b byte) pdlfmt.UnitID {
	var id pdlfmt.UnitID
	id[0] = b
	id[15] = b ^ 0x5a
	return id
}

// writeDocWithRealFrames writes a valid document whose CONTENT segments carry
// genuine content-model frames (each decodable into an authored unit-id).
// Returns the path and the authored unit-ids in storage order.
func writeDocWithRealFrames(t *testing.T, ids ...pdlfmt.UnitID) (string, []pdlfmt.UnitID) {
	t.Helper()
	segs := make([]fixtureSeg, len(ids))
	for i, id := range ids {
		segs[i] = fixtureSeg{segType: container.SegmentTypeContent, body: realFrame(0x01, id)}
	}
	return writeDoc(t, defaultHeader(), segs, nil), append([]pdlfmt.UnitID(nil), ids...)
}
