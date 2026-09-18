// Inferred-marker enumeration (FR-118; T-0292). Some mandated fields can be
// system-INFERRED rather than authored (e.g. an inferred alt-text, an inferred
// base direction). Each such field carries an inferred-marker flag plus a
// basis-of-inference reference recording HOW the value was inferred, so a
// downstream reader can distinguish an authored value from a machine-inferred
// one and audit the basis. This file enumerates every inferrable field and
// provides the marker shape.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

// InferrableField names a mandated field that MAY be system-inferred.
type InferrableField uint8

const (
	// FieldAltText: an embedded object's / region's text alternative (FR-040).
	FieldAltText InferrableField = iota
	// FieldDirection: a text block's / table's base writing direction (FR-032).
	FieldDirection
	// FieldNumberingLabel: a rendered ordered-item numbering label (FR-083)
	// (always derived, hence always "inferred" in the marker sense).
	FieldNumberingLabel
	// FieldCrossReferenceText: a cross-reference's rendered text (FR-084).
	FieldCrossReferenceText
)

// InferrableFields is the closed enumeration of every mandated field capable
// of being system-inferred rather than authored (FR-118).
var InferrableFields = []InferrableField{
	FieldAltText, FieldDirection, FieldNumberingLabel, FieldCrossReferenceText,
}

// BasisOfInference records HOW an inferred value was produced.
type BasisOfInference uint8

const (
	// BasisNone: the value was AUTHORED, not inferred (marker absent).
	BasisNone BasisOfInference = iota
	// BasisHeuristic: inferred by a documented heuristic (e.g. first-strong
	// character for direction).
	BasisHeuristic
	// BasisModel: inferred by a model (e.g. generated alt text).
	BasisModel
	// BasisDerivation: deterministically derived from other document data
	// (e.g. a numbering label from a SEQUENCE_DEFINITION).
	BasisDerivation
)

// InferredMarker is the marker every inferrable field carries: whether the
// field's value was inferred, and if so, the basis of that inference.
type InferredMarker struct {
	// Inferred is true when the value was system-inferred, false when authored.
	Inferred bool
	// Basis records how it was inferred; BasisNone iff Inferred is false.
	Basis BasisOfInference
}

// FieldName returns a stable human-readable name for an inferrable field.
func (f InferrableField) FieldName() string {
	switch f {
	case FieldAltText:
		return "alt_text"
	case FieldDirection:
		return "direction"
	case FieldNumberingLabel:
		return "numbering_label"
	case FieldCrossReferenceText:
		return "cross_reference_text"
	default:
		return "unknown"
	}
}
