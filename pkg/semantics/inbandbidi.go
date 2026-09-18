// PD-BIDI-002 in-band-directional-control validator rule (T-0271; FR-033).
// Directional scope is expressed structurally (the direction field, FR-032),
// never as in-band Unicode directional control characters. Any text sequence
// carrying an embedding, override, isolate, or pop control is rejected -- these
// leak ordering into neighbouring text and an identity-based edit can split a
// control pair. Provisional per the T-0267 ruling (clarify-002.md, OPEN).
package semantics

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// RuleInbandDirectionControl is the validator rule id for in-band directional
// control characters in a text sequence.
const RuleInbandDirectionControl = "PD-BIDI-002"

// bidiControlRunes is the closed set of Unicode bidirectional control code
// points that MUST NOT appear in-band (deprecated embeddings/overrides plus
// isolates and pops).
var bidiControlRunes = map[rune]string{
	0x202A: "LRE", 0x202B: "RLE", 0x202D: "LRO", 0x202E: "RLO", 0x202C: "PDF",
	0x2066: "LRI", 0x2067: "RLI", 0x2068: "FSI", 0x2069: "PDI",
	0x200E: "LRM", 0x200F: "RLM", 0x061C: "ALM",
}

// InbandControlFinding is a PD-BIDI-002 rejection, naming the offending unit,
// the control character's name, and its rune index within the text.
type InbandControlFinding struct {
	Rule    string
	UnitID  pdlfmt.UnitID
	Control string
	RuneIdx int
}

func (f InbandControlFinding) String() string {
	return fmt.Sprintf("%s: unit %x carries in-band directional control %s at rune %d",
		f.Rule, f.UnitID, f.Control, f.RuneIdx)
}

// CheckNoInbandDirectionControl applies PD-BIDI-002: it scans a text sequence
// for any Unicode bidirectional control character and returns a
// *InbandControlFinding at the first one found (naming the unit, control, and
// rune index), or nil when the text carries none.
func CheckNoInbandDirectionControl(unit pdlfmt.UnitID, text string) *InbandControlFinding {
	for i, r := range []rune(text) {
		if name, ok := bidiControlRunes[r]; ok {
			return &InbandControlFinding{Rule: RuleInbandDirectionControl, UnitID: unit, Control: name, RuneIdx: i}
		}
	}
	return nil
}
