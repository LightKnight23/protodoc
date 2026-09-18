// Canon-to-html projection (TR-004; T-0320). ProjectHTML maps C(S) to a
// deterministic markup representation, sharing the canonical traversal with the
// text projector. It backs the `project` CLI verb's --format=html mode. The
// projection is a pure function of the logical state.
package canon

import (
	"encoding/hex"
	"strings"
)

// ProjectHTML renders the document state as deterministic HTML: a fixed
// skeleton with one <div data-unit="..."> per content subtree in canonical
// order, its frame hex in the body. It never consults host state, wall-clock,
// or map iteration order. Frame/unit-id are hex-encoded so the markup is always
// well-formed regardless of frame bytes (no escaping ambiguity).
func ProjectHTML(state *Document) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html>\n<body>\n")
	for _, s := range Traverse(state) {
		b.WriteString(`<div data-unit="`)
		b.WriteString(hex.EncodeToString(s.UnitID[:]))
		b.WriteString(`">`)
		b.WriteString(hex.EncodeToString(s.Frame))
		b.WriteString("</div>\n")
	}
	b.WriteString("</body>\n</html>\n")
	return b.String()
}
