// Canon-to-text projection (TR-004; T-0319). ProjectText maps C(S) to a
// deterministic, human-readable plain-text representation, driven by the shared
// canonical traversal (T-0307). It backs the `project` CLI verb's
// --format=text mode. The projection is a pure function of the logical state:
// identical state yields identical text across runs.
package canon

import (
	"encoding/hex"
	"strings"
)

// ProjectText renders the document state as deterministic plain text: one line
// per content subtree in canonical order, each line "<unit-id-hex>\t<frame-hex>".
// It never consults host state, wall-clock, or map iteration order.
func ProjectText(state *Document) string {
	var b strings.Builder
	for _, s := range Traverse(state) {
		b.WriteString(hex.EncodeToString(s.UnitID[:]))
		b.WriteByte('\t')
		b.WriteString(hex.EncodeToString(s.Frame))
		b.WriteByte('\n')
	}
	return b.String()
}
