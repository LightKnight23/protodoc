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
	"errors"
	"os"

	"Protodoc/pkg/container"
)

// errMergeInputUnreadable is a fallback used only if loadMergeState somehow
// reports !ok with no error (should not happen; kept as a safety net so
// realMergeRun never silently drops the failure).
var errMergeInputUnreadable = errors.New("merge input could not be opened, validated, or decoded")

// mergeState is the real decoded state a merge needs from one input's prefix.
type mergeState struct {
	historyMode   container.HistoryMode
	contentDigest map[uint64][32]byte
	ok            bool
}

func loadMergeState(path string) (mergeState, error) {
	if err := cp006Precondition(path); err != nil {
		return mergeState{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return mergeState{}, err
	}
	defer f.Close()
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return mergeState{}, err
	}
	h, err := container.DecodeHeader(prefix[:container.HeaderSize])
	if err != nil {
		return mergeState{}, err
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return mergeState{}, err
	}
	cd := map[uint64][32]byte{}
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeContent {
			cd[uint64(i)] = slot.Digest
		}
	}
	return mergeState{historyMode: h.HistoryMode, contentDigest: cd, ok: true}, nil
}

// realMergeRun implements the production merge classifier.
func realMergeRun(base, a, b string) MergeOutcome {
	bs, errB := loadMergeState(base)
	as, errA := loadMergeState(a)
	bbs, errBB := loadMergeState(b)
	if !bs.ok || !as.ok || !bbs.ok {
		// A missing/unreadable/malformed input is INVALID (or USAGE if it
		// could not be opened at all), never REFUSED (DEFECT-2026-09-19b/
		// T-0384): REFUSED is reserved for a genuine policy refusal on an
		// otherwise-fine input (CON-024/CON-025/FR-024), and cli.md does not
		// even list exit code 6 among merge's used codes.
		for _, err := range []error{errB, errA, errBB} {
			if err != nil {
				return MergeOutcome{Err: err}
			}
		}
		return MergeOutcome{Err: errMergeInputUnreadable}
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
