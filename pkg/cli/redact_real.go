// Real redact backend (T-0379, DEFECT-2026-09-19 fix). Replaces the no-op
// RedactRun stub with a production backend that OPENS the file, runs the
// CP-006 validate-first precondition, reads the real CONTENT segments, and
// produces output in which every designated subtree's body octets are OMITTED
// (replaced by nothing — its plaintext never appears in the output), declaring
// each omission. Designation is by content-addressed slot digest (hex),
// matching how project/diff/merge identify constructs; authored-unit-id
// designation reuses the frame-decode capability tracked as
// GAP-VERIFY-CONTENT-REBUILD. Go stdlib only.
package cli

import (
	"encoding/hex"
	"io"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/extract"
	"Protodoc/pkg/validate"
)

// realRedactRun implements the production redact backend.
func realRedactRun(path string, subtrees []string) RedactResult {
	if steps, _ := realValidateStepsFor(path); validate.Run(steps).Validity != nil {
		return RedactResult{Err: os.ErrInvalid}
	}
	f, err := os.Open(path)
	if err != nil {
		return RedactResult{Err: err}
	}
	defer f.Close()

	prefix := make([]byte, prefixSize)
	if _, err := f.ReadAt(prefix, 0); err != nil {
		return RedactResult{Err: err}
	}
	table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
	if err != nil {
		return RedactResult{Err: err}
	}
	// Map each requested subtree id (hex of a slot digest prefix) to omit.
	omit := map[string]bool{}
	for _, s := range subtrees {
		omit[s] = true
	}
	digestKeyByOrdinal := map[uint64]string{}
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeContent {
			digestKeyByOrdinal[uint64(i)] = hex.EncodeToString(slot.Digest[:4])
		}
	}

	var out []byte
	var declared []string
	err = extract.Walk(f, func(seg extract.ContentSegment, r io.Reader) error {
		frame := make([]byte, seg.Length)
		if _, rerr := io.ReadFull(r, frame); rerr != nil {
			return rerr
		}
		key := digestKeyByOrdinal[seg.Ordinal]
		if omit[key] {
			// Redacted: the plaintext body is dropped entirely; declare it.
			declared = append(declared, key)
			return nil
		}
		out = append(out, frame...)
		return nil
	})
	if err != nil {
		return RedactResult{Err: err}
	}
	return RedactResult{DeclaredOmissions: declared, Output: out}
}

func init() {
	RedactRun = realRedactRun
}
