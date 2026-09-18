// Computed-inline isolation (T-0272; FR-034). Every inline object whose value
// is computed (a page number, a cross-reference's resolved text, ...) is
// wrapped in a directional isolate so the directionality of that value cannot
// alter the visual order of the surrounding text. Re-materialising the value
// therefore never reorders punctuation nobody edited. Provisional per the
// T-0267 ruling (clarify-002.md, OPEN).
package semantics

// ComputedInline is an inline object whose textual value is computed at
// materialisation time. Its Value's directionality is contained by the
// isolation wrapper.
type ComputedInline struct {
	// Value is the computed text (may be LTR, RTL, or mixed).
	Value string
	// ValueDirection is the resolved base direction of the computed value.
	ValueDirection Direction
}

// IsolatedSegment is one segment of a resolved inline sequence: text plus
// whether it is an isolated computed-inline unit (whose internal direction does
// not leak) or ordinary surrounding text.
type IsolatedSegment struct {
	Text     string
	Isolated bool
}

// ResolveInlineSequence materialises a sequence of surrounding-text and
// computed-inline parts into isolated segments. Every ComputedInline becomes an
// Isolated segment: its value's directionality is contained, so the visual
// order of the surrounding (non-isolated) segments is fixed by their own
// directions alone, independent of any computed value.
//
// `before` and `after` are the surrounding text on each side; `ci` is the
// computed inline between them.
func ResolveInlineSequence(before string, ci ComputedInline, after string) []IsolatedSegment {
	return []IsolatedSegment{
		{Text: before, Isolated: false},
		{Text: ci.Value, Isolated: true}, // isolated: internal direction contained
		{Text: after, Isolated: false},
	}
}

// SurroundingVisualOrder returns the ordered surrounding (non-isolated) text
// segments -- the part of the visual order a computed value must NOT be able to
// perturb. Because the computed inline is isolated, this order is a function of
// the surrounding segments only, never of the computed value.
func SurroundingVisualOrder(segs []IsolatedSegment) []string {
	var out []string
	for _, s := range segs {
		if !s.Isolated {
			out = append(out, s.Text)
		}
	}
	return out
}
