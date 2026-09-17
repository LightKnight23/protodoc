package content_test

import (
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/ledger"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_027_BoundaryBehaviourRoundTripsUnchanged is T-0077's named
// integration test. It builds annotation anchor points carrying all four
// boundary-behaviour values at both endpoints, encodes them into a ledger
// segment, "saves" the document by appending that segment via the M02
// append-only place() primitive, then reads the segment back out of the
// resulting image and decodes the anchor points, asserting every
// boundary-behaviour value survives the save/load cycle byte-for-byte with
// zero drift (FR-027, resting on M02's never-rewrite-a-sealed-segment
// guarantee).
func TestFR_027_BoundaryBehaviourRoundTripsUnchanged(t *testing.T) {
	boundaries := []content.AnchorBoundary{
		content.BoundaryInside,
		content.BoundaryOutside,
		content.BoundaryInsideIfInsertedBefore,
		content.BoundaryInsideIfInsertedAfter,
	}

	// Build one anchor-point pair (start, end) for every combination of
	// start-boundary x end-boundary, so all four values appear at BOTH
	// endpoints.
	type annPair struct{ start, end content.AnchorPoint }
	var pairs []annPair
	var rid pdlfmt.UnitID
	for i := range rid {
		rid[i] = byte(i + 1)
	}
	for si, sb := range boundaries {
		for ei, eb := range boundaries {
			pairs = append(pairs, annPair{
				start: content.AnchorPoint{RunID: rid, BirthOrdinal: uint32(si * 10), Side: content.SideBefore, Boundary: sb},
				end:   content.AnchorPoint{RunID: rid, BirthOrdinal: uint32(ei*10 + 5), Side: content.SideAfter, Boundary: eb},
			})
		}
	}

	// Encode all anchor points into one segment payload.
	var payload []byte
	for _, p := range pairs {
		payload = content.EncodeAnchorPoint(payload, p.start)
		payload = content.EncodeAnchorPoint(payload, p.end)
	}

	// A minimal prior image: exactly the fixed prefix (opaque zero bytes are
	// fine; place() treats the prior range as immutable octets and this test
	// only exercises that the appended segment survives).
	prior := make([]byte, ledger.PrefixLength)

	// "Save": append the anchor segment via the append-only placement API.
	// The logical edit size is the payload length; the write stays well
	// within the NFR-008 budget.
	res, err := ledger.PlaceWithinBudget(prior, 0, ledger.EditDelta{
		NewSegments: []ledger.SegmentPayload{{Octets: payload}},
	}, uint64(len(payload)))
	if err != nil {
		t.Fatalf("place (save): %v", err)
	}

	// "Load": read the appended segment back out of the resulting image.
	got := res.Image[ledger.PrefixLength:]
	if len(got) != len(payload) {
		t.Fatalf("saved anchor segment is %d octets, want %d", len(got), len(payload))
	}

	// Decode every anchor point back and compare to what was written.
	pos := 0
	for i, p := range pairs {
		start, n, err := content.DecodeAnchorPoint(got[pos:])
		if err != nil {
			t.Fatalf("pair %d: decode start: %v", i, err)
		}
		pos += n
		end, n, err := content.DecodeAnchorPoint(got[pos:])
		if err != nil {
			t.Fatalf("pair %d: decode end: %v", i, err)
		}
		pos += n

		if start.Boundary != p.start.Boundary {
			t.Fatalf("pair %d: start boundary drifted: got %v, want %v", i, start.Boundary, p.start.Boundary)
		}
		if end.Boundary != p.end.Boundary {
			t.Fatalf("pair %d: end boundary drifted: got %v, want %v", i, end.Boundary, p.end.Boundary)
		}
		// Full anchor-point equality (run-id, ordinal, side, boundary).
		if start != p.start || end != p.end {
			t.Fatalf("pair %d: anchor point drifted across save/load", i)
		}
	}
}
