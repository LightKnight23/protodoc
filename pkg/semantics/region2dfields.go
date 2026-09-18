// 2D-region field validation (FR-099; T-0291). Validator rule PD-2D-001: a
// 2D-presentation region is rejected when it is missing its linearised-
// reading-order reference, missing its (non-decorative) text alternative, or
// when its reading-order reference does not resolve into ROOT_SEQUENCE.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"strings"

	"Protodoc/pkg/pdlfmt"
)

// RuleRegion2DFields is the validator rule id for 2D-region field presence.
const RuleRegion2DFields = "PD-2D-001"

// Region2DDefectKind classifies a PD-2D-001 defect.
type Region2DDefectKind uint8

const (
	// Region2DMissingLinearRef: the linear-ref is absent (zero unit-id).
	Region2DMissingLinearRef Region2DDefectKind = iota
	// Region2DUnresolvedLinearRef: the linear-ref is not in ROOT_SEQUENCE.
	Region2DUnresolvedLinearRef
	// Region2DMissingAltText: a non-decorative region lacks alt text.
	Region2DMissingAltText
)

// Region2DFinding is a PD-2D-001 rejection naming the offending region.
type Region2DFinding struct {
	Rule     string
	RegionID pdlfmt.UnitID
	Kind     Region2DDefectKind
}

var zeroUnit pdlfmt.UnitID

// CheckRegion2DFields applies PD-2D-001 to a set of regions given the
// ROOT_SEQUENCE order (the set of unit-ids the linear-ref must resolve into).
// It reports a region missing its linear-ref, one whose linear-ref is not in
// ROOT_SEQUENCE, and a non-decorative region with empty alt text, each naming
// the region. An empty result means every region is well-formed.
func CheckRegion2DFields(regions []Region2D, rootSequence []pdlfmt.UnitID) []Region2DFinding {
	inSeq := make(map[pdlfmt.UnitID]bool, len(rootSequence))
	for _, u := range rootSequence {
		inSeq[u] = true
	}
	var findings []Region2DFinding
	for _, r := range regions {
		switch {
		case r.LinearRef == zeroUnit:
			findings = append(findings, Region2DFinding{Rule: RuleRegion2DFields, RegionID: r.RegionID, Kind: Region2DMissingLinearRef})
		case !inSeq[r.LinearRef]:
			findings = append(findings, Region2DFinding{Rule: RuleRegion2DFields, RegionID: r.RegionID, Kind: Region2DUnresolvedLinearRef})
		}
		if !r.Alt.Decorative && strings.TrimSpace(r.Alt.AltText) == "" {
			findings = append(findings, Region2DFinding{Rule: RuleRegion2DFields, RegionID: r.RegionID, Kind: Region2DMissingAltText})
		}
	}
	return findings
}
