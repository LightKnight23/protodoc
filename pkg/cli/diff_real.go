// Real diff backend (T-0377, DEFECT-2026-09-19 fix). Replaces the no-op
// DiffRun stub with a production backend that OPENS both files, runs the
// CP-006 validate-first precondition on each, reads their real CONTENT segment
// inventories (slot digest per ordinal), and reports the constructs that
// differ between the two — by content-addressed slot digest, so a changed or
// moved construct is detected from real bytes. Agreements are omitted (TR-002).
// Go stdlib only.
package cli

import (
	"encoding/hex"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/validate"
)

// contentDigests returns the per-ordinal CONTENT slot digests of a document,
// or an error if the file fails to open/validate/decode. This is a real read
// of the file's segment table (content-addressed identities), not a stub.
func contentDigests(path string) (map[uint64][32]byte, error) {
	if steps, _ := realValidateStepsFor(path); validate.Run(steps).Validity != nil {
		return nil, os.ErrInvalid
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return nil, err
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return nil, err
	}
	out := map[uint64][32]byte{}
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeContent {
			out[uint64(i)] = slot.Digest
		}
	}
	return out, nil
}

// realDiffRun implements the production diff backend. It reports one changed-
// construct description per CONTENT ordinal whose slot digest differs (or is
// present in only one document); identical constructs are omitted. It returns
// an error if either input fails to validate/decode.
func realDiffRun(pathA, pathB string) ([]string, error) {
	da, errA := contentDigests(pathA)
	db, errB := contentDigests(pathB)
	if errA != nil {
		return nil, errA
	}
	if errB != nil {
		return nil, errB
	}

	seen := map[uint64]bool{}
	var changed []string
	// Ordinals present in A: changed if absent from B or digest differs.
	for ord, dgA := range da {
		seen[ord] = true
		dgB, ok := db[ord]
		if !ok {
			changed = append(changed, "removed@ordinal-"+itoaCLI(int(ord)))
			continue
		}
		if dgA != dgB {
			changed = append(changed, "changed@ordinal-"+itoaCLI(int(ord))+" ("+hex.EncodeToString(dgA[:4])+"->"+hex.EncodeToString(dgB[:4])+")")
		}
	}
	// Ordinals present only in B: added.
	for ord := range db {
		if !seen[ord] {
			changed = append(changed, "added@ordinal-"+itoaCLI(int(ord)))
		}
	}
	return changed, nil
}

func init() {
	DiffRun = realDiffRun
}
