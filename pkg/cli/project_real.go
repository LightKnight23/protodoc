// Real project backend (T-0376, updated to close GAP-VERIFY-CONTENT-REBUILD).
// Opens the file, runs the CP-006 validate-first precondition, decodes each
// CONTENT frame into its AUTHORED unit-id + canonical bytes via
// extract.LoadContentRecords, and assembles a canon.Document so the real M17
// canon.ProjectText / canon.ProjectHTML projectors run over actual content
// keyed by the authored unit-id (no slot-digest surrogate). Go stdlib only.
package cli

import (
	"os"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/extract"
)

// realProjectStateFor implements the production project backend.
func realProjectStateFor(path string) (*canon.Document, error) {
	// CP-006 precondition.
	if err := cp006Precondition(path); err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode each CONTENT frame into its authored unit-id + canonical bytes.
	records, err := extract.LoadContentRecords(f)
	if err != nil {
		return nil, err
	}
	doc := &canon.Document{}
	for _, rec := range records {
		doc.Subtrees = append(doc.Subtrees, canon.ContentSubtree{UnitID: rec.UnitID, Frame: rec.Frame})
	}
	return doc, nil
}

func init() {
	ProjectStateFor = realProjectStateFor
}
