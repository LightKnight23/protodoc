// Real diff backend (T-0377, DEFECT-2026-09-19 fix; T-0395,
// DEFECT-2026-09-20b fix). Replaces the no-op DiffRun stub with a production
// backend that OPENS both files, runs the CP-006 validate-first precondition
// on each, and reports the constructs that differ by real decoded content,
// keyed by the AUTHORED unit-id (extract.LoadContentRecords) -- the same
// identity merge/project/redact/publish already key by post
// GAP-VERIFY-CONTENT-REBUILD. Agreements are omitted (TR-002). Go stdlib
// only.
//
// T-0395 fixes a real defect: the previous version classified changes by
// comparing SegmentTableSlot.Digest, the stored per-slot digest field --
// but nothing in any writer path populates that field with a real content
// digest for CONTENT segments (it is left zero), so two documents with the
// same segment COUNT but genuinely different CONTENT always compared equal
// (zero == zero) and were reported "identical" regardless of real content.
// This also contradicted diff.go's own documented contract ("construct-level
// changes, not storage units") by keying on storage ordinal in the first
// place. Found by an end-to-end battery run against real fixtures with
// matching segment counts and differing payloads.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"

	"Protodoc/pkg/extract"
	"Protodoc/pkg/pdlfmt"
)

// loadDiffContent decodes a document's real CONTENT into a map keyed by its
// authored unit-id, or returns an error if the file fails to
// open/validate/decode.
func loadDiffContent(path string) (map[pdlfmt.UnitID][]byte, error) {
	if err := cp006Precondition(path); err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	records, err := extract.LoadContentRecords(f)
	if err != nil {
		return nil, err
	}
	out := make(map[pdlfmt.UnitID][]byte, len(records))
	for _, rec := range records {
		out[rec.UnitID] = rec.Frame
	}
	return out, nil
}

// realDiffRun implements the production diff backend. It reports one changed-
// construct description per authored unit-id whose real content differs (or
// is present in only one document), sorted by unit-id for determinism;
// identical constructs are omitted. It returns an error if either input fails
// to validate/decode.
func realDiffRun(pathA, pathB string) ([]string, error) {
	ca, errA := loadDiffContent(pathA)
	if errA != nil {
		return nil, errA
	}
	cb, errB := loadDiffContent(pathB)
	if errB != nil {
		return nil, errB
	}

	type entry struct {
		id   pdlfmt.UnitID
		desc string
	}
	var entries []entry
	seen := make(map[pdlfmt.UnitID]bool, len(ca))
	for id, fa := range ca {
		seen[id] = true
		fb, ok := cb[id]
		if !ok {
			entries = append(entries, entry{id, "removed@" + hex.EncodeToString(id[:])})
			continue
		}
		if !bytes.Equal(fa, fb) {
			da := sha256.Sum256(fa)
			db := sha256.Sum256(fb)
			entries = append(entries, entry{id, "changed@" + hex.EncodeToString(id[:]) + " (" + hex.EncodeToString(da[:4]) + "->" + hex.EncodeToString(db[:4]) + ")"})
		}
	}
	for id := range cb {
		if !seen[id] {
			entries = append(entries, entry{id, "added@" + hex.EncodeToString(id[:])})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return bytes.Compare(entries[i].id[:], entries[j].id[:]) < 0 })

	changed := make([]string, len(entries))
	for i, e := range entries {
		changed[i] = e.desc
	}
	return changed, nil
}

func init() {
	DiffRun = realDiffRun
}
