// Real redact backend (T-0379, updated to close GAP-VERIFY-CONTENT-REBUILD).
// Opens the file, runs the CP-006 validate-first precondition, decodes each
// CONTENT frame into its AUTHORED unit-id + canonical bytes via
// extract.LoadContentRecords, and produces output in which every designated
// subtree's frame octets are OMITTED (its plaintext never appears), declaring
// each omission by AUTHORED unit-id (hex). Go stdlib only.
package cli

import (
	"encoding/hex"
	"os"

	"Protodoc/pkg/extract"
)

// realRedactRun implements the production redact backend.
func realRedactRun(path string, subtrees []string) RedactResult {
	if err := cp006Precondition(path); err != nil {
		return RedactResult{Err: err}
	}
	f, err := os.Open(path)
	if err != nil {
		return RedactResult{Err: err}
	}
	defer f.Close()

	records, err := extract.LoadContentRecords(f)
	if err != nil {
		return RedactResult{Err: err}
	}

	// Designations are authored unit-ids as full 32-hex-char strings.
	omit := map[string]bool{}
	for _, s := range subtrees {
		omit[s] = true
	}

	var out []byte
	var declared []string
	for _, rec := range records {
		key := hex.EncodeToString(rec.UnitID[:])
		if omit[key] {
			// Redacted: the plaintext frame is dropped entirely; declare it by
			// its authored unit-id.
			declared = append(declared, key)
			continue
		}
		out = append(out, rec.Frame...)
	}
	return RedactResult{DeclaredOmissions: declared, Output: out}
}

func init() {
	RedactRun = realRedactRun
}
