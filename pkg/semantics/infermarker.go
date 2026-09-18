// Inferred-marker consistency (FR-118; T-0293). Validator rule PD-INFER-001:
// on any inferrable field (T-0292), the inferred-marker + basis pair MUST be
// present and coherent — inferred=1 with a non-none basis when the value was
// NOT explicitly authored, and inferred=0 with basis=none when it was. Any
// incoherent combination is a structural reject naming the field.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

// RuleInferMarker is the validator rule id for inferred-marker consistency.
const RuleInferMarker = "PD-INFER-001"

// FieldValue pairs an inferrable field with its marker and whether its value
// was actually authored by a human (the ground truth the marker must reflect).
type FieldValue struct {
	Field       InferrableField
	Marker      InferredMarker
	WasAuthored bool
}

// InferMarkerFinding is a PD-INFER-001 rejection naming the offending field.
type InferMarkerFinding struct {
	Rule   string
	Field  InferrableField
	Reason string
}

// CheckInferMarker applies PD-INFER-001 to a set of inferrable field values. It
// rejects any field whose marker is incoherent with whether the value was
// authored:
//   - authored value  => marker MUST be {inferred=0, basis=none}
//   - inferred value   => marker MUST be {inferred=1, basis!=none}
//
// An empty result means every marker is consistent.
func CheckInferMarker(values []FieldValue) []InferMarkerFinding {
	var findings []InferMarkerFinding
	for _, v := range values {
		if v.WasAuthored {
			if v.Marker.Inferred || v.Marker.Basis != BasisNone {
				findings = append(findings, InferMarkerFinding{
					Rule: RuleInferMarker, Field: v.Field,
					Reason: "authored value carries an inferred marker/basis",
				})
			}
			continue
		}
		// Not authored => must be marked inferred with a real basis.
		if !v.Marker.Inferred {
			findings = append(findings, InferMarkerFinding{
				Rule: RuleInferMarker, Field: v.Field,
				Reason: "inferred value is not marked inferred",
			})
			continue
		}
		if v.Marker.Basis == BasisNone {
			findings = append(findings, InferMarkerFinding{
				Rule: RuleInferMarker, Field: v.Field,
				Reason: "inferred value carries no basis-of-inference",
			})
		}
	}
	return findings
}
