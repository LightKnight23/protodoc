package integrity

import (
	"bytes"
	"testing"
)

// FuzzFR_078_ResidueScannerFindsNoMatches is T-0198's fuzz target (test_kind
// fuzz; named for FR-078's residue property). It generates a published output
// from arbitrary "retained" content and a set of arbitrary "removed" frames,
// and asserts the FR-078 invariant: the published output (built from the
// retained content alone) contains NO removed frame UNLESS that exact byte
// sequence also occurs in the retained content it was built from. Equivalently,
// ScanResidue reports a residue only when the removed octets genuinely appear
// in the emitted output -- never a false negative (a leaked removed unit the
// scanner misses) and never a spurious hit against octets the publish never
// emitted.
//
// The security-relevant direction is no-false-negative: if a removed frame's
// bytes appear in the output, ScanResidue must report it. This target drives
// that with adversarial inputs.
func FuzzFR_078_ResidueScannerFindsNoMatches(f *testing.F) {
	f.Add([]byte("kept"), []byte("removed"))
	f.Add([]byte(""), []byte("x"))
	f.Add([]byte("overlapremovedoverlap"), []byte("removed"))

	f.Fuzz(func(t *testing.T, retainedFrame, removedFrame []byte) {
		in := PublishInput{
			Retained:      []ContentRecord{{UnitID: redUnitID(1), Frame: retainedFrame}},
			RemovedFrames: [][]byte{removedFrame},
		}
		out := Publish(in)

		findings := ScanResidue(out, [][]byte{removedFrame})

		// Ground truth: does the removed frame actually occur in the output?
		occurs := len(removedFrame) > 0 && bytes.Contains(out.Emitted, removedFrame)

		if occurs && len(findings) == 0 {
			t.Fatalf("FALSE NEGATIVE: removed frame occurs in output but ScanResidue missed it")
		}
		if !occurs && len(findings) != 0 {
			t.Fatalf("FALSE POSITIVE: ScanResidue reported residue not present in the output")
		}

		// FR-078 construction guarantee: the publish output equals the
		// retained frame exactly (it is built from the retained set alone), so
		// any removed-frame occurrence is necessarily also in the retained
		// content -- publish never introduces removed octets on its own.
		if !bytes.Equal(out.Emitted, retainedFrame) {
			t.Fatalf("publish emitted octets other than the retained content")
		}
		if occurs && !bytes.Contains(retainedFrame, removedFrame) {
			t.Fatalf("a removed frame appeared in the output without being in the retained content")
		}
	})
}
