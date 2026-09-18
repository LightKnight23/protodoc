// Substitution never uses a host resource and never reflows (T-0261; FR-114).
// IF a required embedded resource is absent or fails verification, a conforming
// reader SHALL NOT substitute a host-system resource and SHALL NOT reflow the
// content. The substitution path renders a placeholder at the recorded extent
// (T-0258) from self-contained data only -- no filesystem, network, or host
// font-directory lookup, and no call into the reflow engine -- so pagination is
// octet-identical to the reference under a font-less/network-less harness
// (CP-005/CP-010).
package render

import "Protodoc/pkg/pdlfmt"

// Substitute performs the FR-114-compliant substitution for a failed/absent
// resource: it renders the placeholder at the resource's RECORDED extent
// (never a host resource's metrics) and records the event in the report,
// WITHOUT any host lookup and WITHOUT reflowing. It returns the placeholder;
// the caller composites it at the recorded extent, leaving pagination
// unchanged.
func Substitute(report *RenderReport, resource pdlfmt.UnitID, recordedExtent Extent, reason SubstitutionReason) Placeholder {
	// The placeholder is derived solely from the resource id and the RECORDED
	// extent -- no host font/resource is consulted, and the extent is the
	// document's own recorded value, not a host resource's metrics.
	ph := RenderPlaceholder(resource, recordedExtent)
	report.RecordSubstitution(resource, reason, KindPlaceholder)
	// Note: Reflow / FitAtViewport are deliberately NOT called here -- the
	// content is not re-laid-out on substitution (FR-114).
	return ph
}
