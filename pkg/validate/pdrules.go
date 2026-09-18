// Previously-unimplemented spec-named validator rules (M02-M07 negative-corpus
// rules whose implementing tasks were not in the executed build spine). Each
// rule here is implemented faithfully to its spec.md/contracts normative text
// and paired with a negative-corpus conformance case (pdrules_test.go), closing
// its zero-conformance gap in docs/nfr-029-traceability-gap-report.md.
//
// NOTE on the PD-NORM-001 vs PD-NFC-001/002 naming conflict (analysis.md major
// finding, CP-011 non-negotiable #5): spec.md names the NFC rule PD-NORM-001
// while the frozen contracts name it PD-NFC-001/PD-NFC-002. Which id is
// canonical is a ruling that still requires Eyvar/themis and is NOT resolved
// here; this file provides the validator + conformance coverage under ALL
// THREE live spellings so none remains a zero-conformance gap, and does not
// silently pick a winner or rewrite any frozen artifact.
package validate

import (
	"fmt"

	"Protodoc/pkg/content"
)

// --- PD-DISC-001: reserved frame-discriminant rejection ---------------------

// RulePDDISC001 is the reserved-discriminant rejection rule id.
const RulePDDISC001 = "PD-DISC-001"

// reservedDiscriminant reports whether a frame discriminant is in a v1-reserved
// range (0x10-0x3F and 0x80-0xFE per document.abnf S1). 0x00 is unused/never a
// real type; 0x01-0x0F and 0x40-0x7F are assigned or owned by integrity.abnf.
func reservedDiscriminant(d uint8) bool {
	return (d >= 0x10 && d <= 0x3F) || (d >= 0x80 && d <= 0xFE)
}

// CheckDiscriminant rejects a frame whose discriminant is v1-reserved (PD-DISC-
// 001), returning a *Finding naming the offending value; nil when assigned.
func CheckDiscriminant(discriminant uint8) *Finding {
	if reservedDiscriminant(discriminant) {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDDISC001,
			Message: fmt.Sprintf("frame discriminant 0x%02X is in a v1-reserved range and MUST be rejected", discriminant),
		}
	}
	return nil
}

// --- PD-DUP-001: dual-mechanism (A-DUP) rejection ---------------------------

// RulePDDUP001 is the dual-mechanism rejection rule id.
const RulePDDUP001 = "PD-DUP-001"

// CheckDualMechanism rejects a construct that engages more than one capability
// mechanism for the same effect (A-DUP: every construct maps to exactly one
// capability). capabilities is the set of capability ids the construct
// engaged; more than one is a dual-mechanism document.
func CheckDualMechanism(constructID string, capabilities []string) *Finding {
	if len(capabilities) > 1 {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDDUP001,
			Message: fmt.Sprintf("construct %s engages %d capabilities %v; exactly one is allowed (A-DUP)", constructID, len(capabilities), capabilities),
		}
	}
	return nil
}

// --- PD-EXT-001: non-enveloped extension construct rejection ----------------

// RulePDEXT001 is the non-enveloped-extension rejection rule id.
const RulePDEXT001 = "PD-EXT-001"

// CheckExtensionEnveloped rejects an extension construct that is NOT carried in
// an extension envelope (every future/extension construct MUST be enveloped so
// a previous-generation reader can determine its extent/disposition/fallback,
// corpus C-FUTURE).
func CheckExtensionEnveloped(constructID string, enveloped bool) *Finding {
	if !enveloped {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDEXT001,
			Message: fmt.Sprintf("extension construct %s is not carried in an extension envelope and MUST be rejected", constructID),
		}
	}
	return nil
}

// --- PD-FONT-001: font-reference seven-value completeness -------------------

// RulePDFONT001 is the font-reference completeness rule id.
const RulePDFONT001 = "PD-FONT-001"

// FontReference is a reference to an embedded font: the seven declared values
// FR-089 requires, plus the digest of the referenced FONT_SUBSET.
type FontReference struct {
	// Values are the seven normative font values (name, version, digest, axes,
	// codepoints, features, embed-perm). All seven MUST be present (non-empty).
	Values [7]string
	// ReferencedDigest is the digest the reference carries; a substitution
	// whose actual subset digest differs is detected.
	ReferencedDigest [32]byte
}

// CheckFontReference rejects a font reference missing any of its seven values,
// or one whose referenced digest does not match the actual FONT_SUBSET digest
// (PD-FONT-001), naming the failure.
func CheckFontReference(ref FontReference, actualSubsetDigest [32]byte) *Finding {
	for i, v := range ref.Values {
		if v == "" {
			return &Finding{
				Step:    StepExtEnvelope,
				RuleID:  RulePDFONT001,
				Message: fmt.Sprintf("font reference missing value %d of 7 (FR-089)", i+1),
			}
		}
	}
	if ref.ReferencedDigest != actualSubsetDigest {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDFONT001,
			Message: "font reference digest does not match the referenced FONT_SUBSET (substitution detected)",
		}
	}
	return nil
}

// --- PD-INDEX-001: unit-index leaf collision rejection ----------------------

// RulePDINDEX001 is the index-leaf collision rejection rule id.
const RulePDINDEX001 = "PD-INDEX-001"

// CheckIndexLeafKeys rejects a set of unit-index leaf keys containing a
// duplicate (which would let two different leaves collide), naming the
// offending key (PD-INDEX-001).
func CheckIndexLeafKeys(keys []string) *Finding {
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		if seen[k] {
			return &Finding{
				Step:    StepExtEnvelope,
				RuleID:  RulePDINDEX001,
				Message: fmt.Sprintf("unit-index leaf key %q appears more than once; two leaves would collide", k),
			}
		}
		seen[k] = true
	}
	return nil
}

// --- PD-MODE-001: exactly-one-mode / no-mode-change rejection ---------------

// RulePDMODE001 is the mode-declaration rule id.
const RulePDMODE001 = "PD-MODE-001"

// CheckMode rejects a document that does not declare exactly one mode, or that
// attempts to change an already-declared mode (PD-MODE-001).
func CheckMode(declaredModes []string, priorMode string) *Finding {
	if len(declaredModes) != 1 {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDMODE001,
			Message: fmt.Sprintf("document declares %d modes; exactly one is required", len(declaredModes)),
		}
	}
	if priorMode != "" && declaredModes[0] != priorMode {
		return &Finding{
			Step:    StepExtEnvelope,
			RuleID:  RulePDMODE001,
			Message: fmt.Sprintf("attempt to change declared mode from %q to %q is rejected", priorMode, declaredModes[0]),
		}
	}
	return nil
}

// --- PD-NFC-001 / PD-NFC-002 / PD-NORM-001: NFC normalization ---------------

// The NFC rule under all three live spellings (see the package doc note on the
// unresolved canonical-id ruling). PD-NFC-001 is the run-level rule; PD-NFC-002
// its cross-run/boundary counterpart; PD-NORM-001 is spec.md's spelling of the
// same rule.
const (
	RulePDNFC001  = "PD-NFC-001"
	RulePDNFC002  = "PD-NFC-002"
	RulePDNORM001 = "PD-NORM-001"
)

// CheckNFCRun rejects a run whose text is not NFC-normalized, under the given
// rule id spelling (PD-NFC-001 / PD-NORM-001), naming the run.
func CheckNFCRun(ruleID, runID, text string) *Finding {
	if !content.IsNFC(text) {
		return &Finding{
			Step:    StepNFCAndIdentity,
			RuleID:  ruleID,
			Message: fmt.Sprintf("run %s text is not NFC-normalized", runID),
		}
	}
	return nil
}

// CheckNFCBoundary rejects text that is NFC per-run but forms a non-NFC
// sequence across a run boundary (PD-NFC-002): concatenating two adjacent runs
// must itself remain NFC.
func CheckNFCBoundary(leftRunID, leftText, rightText string) *Finding {
	if content.IsNFC(leftText) && content.IsNFC(rightText) && !content.IsNFC(leftText+rightText) {
		return &Finding{
			Step:    StepNFCAndIdentity,
			RuleID:  RulePDNFC002,
			Message: fmt.Sprintf("runs adjacent at %s form a non-NFC sequence across the boundary", leftRunID),
		}
	}
	return nil
}

// --- PD-PREV-001: bounded-prefix payload presence, no network ---------------

// RulePDPREV001 is the bounded-prefix payload rule id.
const RulePDPREV001 = "PD-PREV-001"

// BoundedPrefixLen is the octet within which the preview/metadata payload must
// be present (container.abnf: the header+ring+frontmatter region ends at
// 262144, before the segment table).
const BoundedPrefixLen = 262144

// CheckBoundedPrefixPayload rejects a document whose bounded-prefix payload
// does not fit within the first BoundedPrefixLen octets, or that would require
// a network fetch to obtain it (PD-PREV-001). payloadEndOffset is the octet
// offset past which the payload extends; networkRequired flags a socket call.
func CheckBoundedPrefixPayload(payloadEndOffset int, networkRequired bool) *Finding {
	if networkRequired {
		return &Finding{
			Step:    StepMagicHeader,
			RuleID:  RulePDPREV001,
			Message: "bounded-prefix payload requires a network fetch; it MUST be present in the prefix",
		}
	}
	if payloadEndOffset > BoundedPrefixLen {
		return &Finding{
			Step:    StepMagicHeader,
			RuleID:  RulePDPREV001,
			Message: fmt.Sprintf("bounded-prefix payload extends to octet %d, past the %d-octet prefix bound", payloadEndOffset, BoundedPrefixLen),
		}
	}
	return nil
}

// --- PD-SEGTYPE-001: reserved slot-segment-type rejection -------------------

// RulePDSEGTYPE001 is the reserved-segment-type rejection rule id.
const RulePDSEGTYPE001 = "PD-SEGTYPE-001"

// reservedSegmentType reports whether a slot-segment-type is v1-reserved.
// Assigned: 0 (unused), 1 (CONTENT), 2 (RESOURCE), 3 (HISTORY), 4 (ATTEST).
// Everything else (5-255) is reserved and MUST be rejected (not skipped).
func reservedSegmentType(t uint8) bool { return t > 4 }

// CheckSegmentType rejects a segment whose slot-segment-type is v1-reserved
// (PD-SEGTYPE-001), naming the offending value; nil when assigned.
func CheckSegmentType(segType uint8) *Finding {
	if reservedSegmentType(segType) {
		return &Finding{
			Step:    StepSegTypeCoverage,
			RuleID:  RulePDSEGTYPE001,
			Message: fmt.Sprintf("slot-segment-type %d is v1-reserved and MUST be rejected, not skipped", segType),
		}
	}
	return nil
}
