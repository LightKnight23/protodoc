// Reading-order text fidelity (T-0089, FR-035): the extraction view emits a
// text unit's authored Unicode scalar sequence exactly as stored, by
// concatenating each Run's text in reading order. Text was NFC-scoped at
// write time by the identity layer and is NEVER renormalized at extraction
// time -- renormalising would change scalar counts and relocate anchors, so
// the extractor reproduces the stored octets byte-for-byte.
package extract

import (
	"strings"

	"Protodoc/pkg/content"
)

// ExtractText returns a text unit's authored scalar sequence: the runs' text
// concatenated in reading (slice) order, exactly as stored, with no
// normalisation applied (FR-035). The result is byte-identical to the
// concatenation of the stored run texts.
func ExtractText(runs []content.Run) string {
	var b strings.Builder
	for _, r := range runs {
		b.WriteString(r.Text)
	}
	return b.String()
}
