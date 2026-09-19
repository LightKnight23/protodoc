// Real merge backend (T-0378, DEFECT-2026-09-19 fix). Replaces the no-op
// MergeRun stub with a production backend that OPENS all three files (base, a,
// b), runs the CP-006 validate-first precondition on each, and classifies the
// merge from real decoded state: a history-mode mismatch between the sides is a
// CON-025 REFUSED; a construct that BOTH sides changed divergently from base is
// a genuine R2/R3 CONFLICT naming both sides' values; otherwise the merge is
// clean. Go stdlib only.
package cli

import (
	"encoding/hex"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/validate"
)

// mergeState is the real decoded state a merge needs from one input's prefix.
type mergeState struct {
	historyMode   container.HistoryMode
	contentDigest map[uint64][32]byte
	ok            bool
}

func loadMergeState(path string) mergeState {
	if steps, _ := realValidateStepsFor(path); validate.Run(steps).Validity != nil {
		return mergeState{}
	}
	f, err := os.Open(path)
	if err != nil {
		return mergeState{}
	}
	defer f.Close()
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return mergeState{}
	}
	h, err := container.DecodeHeader(prefix[:container.HeaderSize])
	if err != nil {
		return mergeState{}
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return mergeState{}
	}
	cd := map[uint64][32]byte{}
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeContent {
			cd[uint64(i)] = slot.Digest
		}
	}
	return mergeState{historyMode: h.HistoryMode, contentDigest: cd, ok: true}
}

// realMergeRun implements the production merge classifier.
func realMergeRun(base, a, b string) MergeOutcome {
	bs, as, bbs := loadMergeState(base), loadMergeState(a), loadMergeState(b)
	if !bs.ok || !as.ok || !bbs.ok {
		return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CP-006"}
	}

	// CON-025: history-mode mismatch between the two sides is refused.
	if as.historyMode != bbs.historyMode {
		return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CON-025"}
	}

	// R2/R3 conflict: a content ordinal that BOTH sides changed away from base,
	// to DIFFERENT values, is a genuine divergent-edit conflict.
	for ord, baseDg := range bs.contentDigest {
		aDg, aok := as.contentDigest[ord]
		bDg, bok := bbs.contentDigest[ord]
		if aok && bok && aDg != baseDg && bDg != baseDg && aDg != bDg {
			return MergeOutcome{
				Kind:   MergeConflict,
				ValueA: hex.EncodeToString(aDg[:4]),
				ValueB: hex.EncodeToString(bDg[:4]),
			}
		}
	}
	return MergeOutcome{Kind: MergeClean}
}

func init() {
	MergeRun = realMergeRun
}
