// Real merge backend (T-0378, DEFECT-2026-09-19 fix; T-0391, DEFECT-2026-09-19c
// fix). Replaces the no-op MergeRun stub with a production backend that OPENS
// all three files (base, a, b), runs the CP-006 validate-first precondition on
// each, and classifies the merge from real decoded state: a history-mode
// mismatch between the sides is a CON-025 REFUSED; a genuine per-construct
// three-way merge (pkg/merge.ThreeWayMerge, keyed by the AUTHORED unit-id
// decoded from each CONTENT frame, the same identity project/redact/publish
// already key by post GAP-VERIFY-CONTENT-REBUILD) resolves a construct changed
// on only one side, agrees an identical change on both sides, and reports a
// genuine CONFLICT (naming both real values) for a divergent change -- never
// selecting a side, concatenating, or interleaving (TR-003). A clean result is
// re-canonicalized and its real bytes returned as Output. Go stdlib only.
//
// Honest scope (DEFECT-2026-09-19c): this wires the real per-construct
// three-way merge (ThreeWayMerge) but not the full pkg/merge.Orchestrate
// precondition set -- CON-024 (retention-point crossing) and FR-096
// (erased-unit replay) guards need real History/Erasure segment decoding,
// which no CLI verb performs yet. Only CON-025 (history-mode mismatch) is
// checked, as before. This is a real, disclosed gap, not a silent omission.
package cli

import (
	"bytes"
	"encoding/hex"
	"errors"
	"os"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/merge"
	"Protodoc/pkg/pdlfmt"
)

// errMergeInputUnreadable is a fallback used only if loadMergeState somehow
// reports !ok with no error (should not happen; kept as a safety net so
// realMergeRun never silently drops the failure).
var errMergeInputUnreadable = errors.New("merge input could not be opened, validated, or decoded")

// mergeState is the real decoded state a merge needs from one input.
type mergeState struct {
	historyMode container.HistoryMode
	content     map[pdlfmt.UnitID][]byte // keyed by authored unit-id
	ok          bool
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
	records, err := extract.LoadContentRecords(f)
	if err != nil {
		return mergeState{}, err
	}
	content := make(map[pdlfmt.UnitID][]byte, len(records))
	for _, rec := range records {
		content[rec.UnitID] = rec.Frame
	}
	return mergeState{historyMode: h.HistoryMode, content: content, ok: true}, nil
}

// firstBytes returns up to n leading bytes of b (fewer if b is shorter), for
// the conflict summary's short value preview.
func firstBytes(b []byte, n int) []byte {
	if len(b) < n {
		return b
	}
	return b[:n]
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

	// Real per-construct three-way merge (TR-003), keyed by authored unit-id.
	result := merge.ThreeWayMerge(bs.content, as.content, bbs.content)
	if result.HasConflict() {
		c := result.Conflicts[0]
		return MergeOutcome{
			Kind:   MergeConflict,
			ValueA: hex.EncodeToString(firstBytes(c.Left, 4)),
			ValueB: hex.EncodeToString(firstBytes(c.Right, 4)),
		}
	}

	// Clean: re-canonicalize the merged constructs into real output bytes.
	doc := &canon.Document{}
	for id, frame := range result.Merged {
		doc.Subtrees = append(doc.Subtrees, canon.ContentSubtree{UnitID: id, Frame: frame})
	}
	var buf bytes.Buffer
	if err := canon.Canonicalize(doc, &buf); err != nil {
		return MergeOutcome{Err: err}
	}
	return MergeOutcome{Kind: MergeClean, Output: buf.Bytes()}
}

func init() {
	MergeRun = realMergeRun
}
