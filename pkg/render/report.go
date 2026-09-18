// Render report (T-0259; FR-112). IF a required embedded resource is absent or
// fails verification, a conforming reader records the substitution in a
// machine-readable render report -- a visible placeholder tells a human, the
// report tells a pipeline. Every substitution event names the resource id, the
// failure reason, and the substitution kind, so an automated publishing chain
// can detect that it shipped a degraded rendering.
package render

import (
	"encoding/json"

	"Protodoc/pkg/pdlfmt"
)

// SubstitutionReason is why a resource was substituted (a closed enum).
type SubstitutionReason string

const (
	// ReasonAbsent: the resource was not present.
	ReasonAbsent SubstitutionReason = "absent"
	// ReasonDigestMismatch: the resource failed its digest check.
	ReasonDigestMismatch SubstitutionReason = "digest-mismatch"
	// ReasonDecodeFailed: the resource failed to decode.
	ReasonDecodeFailed SubstitutionReason = "decode-failed"
)

// SubstitutionKind is what was rendered in the resource's place.
type SubstitutionKind string

const (
	// KindPlaceholder: a non-decorative placeholder mark (FR-111).
	KindPlaceholder SubstitutionKind = "placeholder"
)

// SubstitutionEvent is one recorded resource substitution during a render run.
type SubstitutionEvent struct {
	ResourceID string             `json:"resource_id"` // hex unit id
	Reason     SubstitutionReason `json:"reason"`
	Kind       SubstitutionKind   `json:"kind"`
}

// RenderReport is the machine-readable report of a render run. It records every
// substitution event; PaginationAuthoritative is the pagination-authority flag
// (set false by the substitution recorder, wired in T-0260 for its own
// requirement).
type RenderReport struct {
	Substitutions           []SubstitutionEvent `json:"substitutions"`
	PaginationAuthoritative bool                `json:"pagination_authoritative"`
}

// NewRenderReport returns a report for a clean run: no substitutions,
// pagination authoritative.
func NewRenderReport() *RenderReport {
	return &RenderReport{PaginationAuthoritative: true}
}

// RecordSubstitution appends a substitution event naming the resource id, the
// failure reason, and the substitution kind (FR-112), and marks the resulting
// pagination NON-authoritative: any render that performed at least one
// substitution has PaginationAuthoritative = false, so a consumer can tell its
// page-valued references were resolved against a degraded rendering (T-0260's
// requirement). A render with zero substitutions leaves it authoritative.
func (r *RenderReport) RecordSubstitution(resource pdlfmt.UnitID, reason SubstitutionReason, kind SubstitutionKind) {
	r.Substitutions = append(r.Substitutions, SubstitutionEvent{
		ResourceID: hexID(resource),
		Reason:     reason,
		Kind:       kind,
	})
	r.PaginationAuthoritative = false
}

// MarshalJSON emits the report as machine-readable JSON.
func (r *RenderReport) JSON() ([]byte, error) {
	return json.Marshal(r)
}
