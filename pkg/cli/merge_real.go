// Real merge backend (T-0378, DEFECT-2026-09-19 fix; T-0391, DEFECT-2026-09-19c
// fix; T-0393, CON-024/FR-096 guards). Replaces the no-op MergeRun stub with a
// production backend that OPENS all three files (base, a, b), runs the CP-006
// validate-first precondition on each, and classifies the merge from real
// decoded state: a history-mode mismatch between the sides is a CON-025
// REFUSED; an incoming branch change whose CONTENT segment ordinal predates
// the base's declared retention point is a CON-024 REFUSED; an incoming
// change that would re-derive content the base's HISTORY segments record as
// erased is an FR-096 REFUSED; otherwise a genuine per-construct three-way
// merge (pkg/merge.ThreeWayMerge, keyed by the AUTHORED unit-id decoded from
// each CONTENT frame, the same identity project/redact/publish already key by
// post GAP-VERIFY-CONTENT-REBUILD) resolves a construct changed on only one
// side, agrees an identical change on both sides, and reports a genuine
// CONFLICT (naming both real values) for a divergent change -- never
// selecting a side, concatenating, or interleaving (TR-003). A clean result is
// re-canonicalized and its real bytes returned as Output. Go stdlib only.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/history"
	"Protodoc/pkg/merge"
	"Protodoc/pkg/pdlfmt"
)

// errMergeInputUnreadable is a fallback used only if loadMergeState somehow
// reports !ok with no error (should not happen; kept as a safety net so
// realMergeRun never silently drops the failure).
var errMergeInputUnreadable = errors.New("merge input could not be opened, validated, or decoded")

// mergeState is the real decoded state a merge needs from one input.
type mergeState struct {
	historyMode    container.HistoryMode
	retentionPoint uint16                   // meaningful only when historyMode == HistoryRetainedFromPoint
	content        map[pdlfmt.UnitID][]byte // keyed by authored unit-id
	ordinals       map[pdlfmt.UnitID]uint64 // authored unit-id -> its CONTENT segment's storage ordinal
	erasures       []history.ErasureRecord  // this document's own recorded erasures (FR-061)
	ok             bool
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
	info, err := f.Stat()
	if err != nil {
		return mergeState{}, err
	}
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return mergeState{}, err
	}
	h, err := container.DecodeHeader(prefix[:container.HeaderSize])
	if err != nil {
		return mergeState{}, err
	}
	ringRegion := prefix[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
	winner, _, err := container.SelectWinner(ringRegion, uint64(info.Size()))
	if err != nil {
		return mergeState{}, err
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return mergeState{}, err
	}
	records, err := extract.LoadContentRecords(f)
	if err != nil {
		return mergeState{}, err
	}
	content := make(map[pdlfmt.UnitID][]byte, len(records))
	ordinals := make(map[pdlfmt.UnitID]uint64, len(records))
	for _, rec := range records {
		content[rec.UnitID] = rec.Frame
		ordinals[rec.UnitID] = rec.Ordinal
	}
	erasures, err := DiscoverErasureRecords(f, table)
	if err != nil {
		return mergeState{}, err
	}
	return mergeState{
		historyMode:    h.HistoryMode,
		retentionPoint: winner.RetentionPoint,
		content:        content,
		ordinals:       ordinals,
		erasures:       erasures,
		ok:             true,
	}, nil
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

	// CON-024/FR-096 (T-0393): guard each incoming branch's real changes
	// against the base's declared retention point and its own recorded
	// erasures, using the real content-model and HISTORY-segment decode
	// already loaded above -- no fabricated evidence, no skipped check.
	var baseDecl history.Declaration
	var declErr error
	if bs.historyMode == container.HistoryRetainedFromPoint {
		baseDecl, declErr = history.NewDeclarationWithRetentionPoint(bs.historyMode, bs.retentionPoint)
	} else {
		baseDecl, declErr = history.NewDeclaration(bs.historyMode)
	}
	if declErr != nil {
		return MergeOutcome{Err: declErr}
	}
	erasureIdx := merge.NewErasureIndex(bs.erasures)
	for _, branch := range []mergeState{as, bbs} {
		changes := merge.DiffConstructs(bs.content, branch.content)
		var incomingOrdinals []uint16
		for _, c := range changes {
			if c.Kind == merge.Removed {
				continue // nothing incoming from this branch to check
			}
			if ord, ok := branch.ordinals[c.Construct]; ok {
				incomingOrdinals = append(incomingOrdinals, uint16(ord))
			}
			digest := sha256.Sum256(c.Right)
			if err := erasureIdx.CheckReplayOfErasedUnit(merge.IncomingChange{Identity: c.Construct, Digest: digest}); err != nil {
				return MergeOutcome{Kind: MergeRefused, RefusedCondition: "FR-096"}
			}
		}
		if err := merge.CheckRetentionPoint(baseDecl, incomingOrdinals); err != nil {
			return MergeOutcome{Kind: MergeRefused, RefusedCondition: "CON-024"}
		}
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
