package canon_test

import (
	"testing"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestTR_005_ProjectionRejectedAsDocumentInput is T-0323's named integration
// test (TR-005). A text/html projection carries no valid Protodoc magic or
// header, so the document-input path (magic identification / header decode used
// by validate/extract/verify) rejects it as a structural failure — keeping the
// projection strictly outside the document trust boundary.
func TestTR_005_ProjectionRejectedAsDocumentInput(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	state := &canon.Document{Subtrees: []canon.ContentSubtree{
		{UnitID: u(0x10), Frame: []byte("alpha")},
		{UnitID: u(0x20), Frame: []byte("beta")},
	}}

	text := []byte(canon.ProjectText(state))
	html := []byte(canon.ProjectHTML(state))

	for name, proj := range map[string][]byte{"text": text, "html": html} {
		// No valid Protodoc magic.
		if container.IsProtodocMagic(proj) {
			t.Errorf("%s projection must not carry the Protodoc magic", name)
		}
		// The document-input path (header decode) rejects it structurally.
		buf := proj
		if len(buf) < container.HeaderSize {
			padded := make([]byte, container.HeaderSize)
			copy(padded, buf)
			buf = padded
		}
		if _, err := container.DecodeHeader(buf[:container.HeaderSize]); err == nil {
			t.Errorf("%s projection must be rejected by the header decoder (document-input gate)", name)
		}
	}
}
