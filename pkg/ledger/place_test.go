package ledger

import (
	"bytes"
	"errors"
	"testing"
)

// makePriorImage builds a minimal valid placement base: a prior image of
// exactly PrefixLength octets whose contents are a fixed, non-zero pattern
// so a mutation anywhere in the prior range is detectable by byte compare.
// The bytes need not be a semantically valid prefix for T-0033: place()
// treats the prior range as opaque immutable octets and only guarantees it
// is not disturbed, which is exactly the FR-056 property under test.
func makePriorImage() []byte {
	img := make([]byte, PrefixLength)
	for i := range img {
		img[i] = byte(i*31 + 7)
	}
	return img
}

// makeSegment returns a distinct non-empty synthetic segment payload of
// the given length, tagged by seed so successive appends are individually
// identifiable in the resulting image.
func makeSegment(seed, length int) SegmentPayload {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte((seed+1)*13 + i*7)
	}
	return SegmentPayload{Octets: b}
}

// TestFR_056_PlaceNeverMutatesExistingOctets is T-0033's primary test.
// It writes N successive edits via place(), retaining every intermediate
// image, and after each edit asserts that the whole prior image is
// byte-identical to its earlier self at every offset below the new tail:
// place appended octets only and mutated nothing previously committed
// (FR-056, container.abnf S8, DoD: "a differential test writing N edits
// and diffing intermediate images confirms zero mutation of previously
// committed octets").
func TestFR_056_PlaceNeverMutatesExistingOctets(t *testing.T) {
	const editCount = 12

	// image[k] is the file image after k edits; image[0] is the base.
	images := make([][]byte, 0, editCount+1)
	base := makePriorImage()
	images = append(images, base)

	// Keep an independent, defensive copy of the base to detect any
	// in-place mutation of the caller's slice by place() itself.
	baseCopy := append([]byte(nil), base...)

	for k := 1; k <= editCount; k++ {
		prior := images[k-1]
		priorLen := len(prior)

		// A snapshot of prior taken BEFORE the edit, so the post-edit
		// comparison is against octets place() could not have seen change.
		priorSnapshot := append([]byte(nil), prior...)

		delta := EditDelta{NewSegments: []SegmentPayload{
			makeSegment(k, 64+k),          // varying-length segment
			makeSegment(k*100, 96),        // a second segment same edit
		}}

		next, err := place(prior, delta)
		if err != nil {
			t.Fatalf("edit %d: place returned error: %v", k, err)
		}

		// The new image must be strictly longer (two non-empty segments
		// were appended) and its head must equal prior exactly.
		if len(next) <= priorLen {
			t.Fatalf("edit %d: new image length %d not greater than prior length %d", k, len(next), priorLen)
		}
		if !bytes.Equal(next[:priorLen], priorSnapshot) {
			// Locate the first differing offset for a precise failure.
			for off := 0; off < priorLen; off++ {
				if next[off] != priorSnapshot[off] {
					t.Fatalf("edit %d: place mutated a previously committed octet at offset %d: was 0x%02x, now 0x%02x", k, off, priorSnapshot[off], next[off])
				}
			}
			t.Fatalf("edit %d: head of new image differs from prior but no single offset located (length mismatch?)", k)
		}

		// place must not mutate the caller's prior slice in place.
		if !bytes.Equal(prior, priorSnapshot) {
			t.Fatalf("edit %d: place mutated the caller's prior slice in place", k)
		}

		images = append(images, next)
	}

	// Cross-check every intermediate image against every LATER image:
	// image[i]'s full octet range must survive unchanged as the head of
	// every image[j>i]. This is the "diffing intermediate images" the DoD
	// names, checked pairwise across the whole sequence rather than only
	// step-to-step.
	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			if len(images[j]) < len(images[i]) {
				t.Fatalf("image[%d] (len %d) is shorter than earlier image[%d] (len %d): ledger shrank", j, len(images[j]), i, len(images[i]))
			}
			if !bytes.Equal(images[j][:len(images[i])], images[i]) {
				t.Fatalf("image[%d]'s head is not byte-identical to image[%d]: an earlier committed octet was later mutated", j, i)
			}
		}
	}

	// The original caller-supplied base slice must be untouched throughout.
	if !bytes.Equal(base, baseCopy) {
		t.Fatalf("the original base image was mutated in place over the course of %d edits", editCount)
	}
}

// TestFR_056_PlaceAppendsContiguouslyAtTail confirms the newly-allocated
// octets are exactly the appended payloads, contiguous and in order,
// starting immediately at the prior tail with no gap or padding: this is
// the "no more than newly-allocated octets differ" half of FR-056's
// append-only guarantee.
func TestFR_056_PlaceAppendsContiguouslyAtTail(t *testing.T) {
	prior := makePriorImage()
	s1 := makeSegment(1, 80)
	s2 := makeSegment(2, 200)

	out, err := place(prior, EditDelta{NewSegments: []SegmentPayload{s1, s2}})
	if err != nil {
		t.Fatalf("place: %v", err)
	}

	wantLen := len(prior) + len(s1.Octets) + len(s2.Octets)
	if len(out) != wantLen {
		t.Fatalf("new image length %d, want %d", len(out), wantLen)
	}

	off := len(prior)
	if !bytes.Equal(out[off:off+len(s1.Octets)], s1.Octets) {
		t.Fatalf("first appended segment octets not placed at prior tail offset %d", off)
	}
	off += len(s1.Octets)
	if !bytes.Equal(out[off:off+len(s2.Octets)], s2.Octets) {
		t.Fatalf("second appended segment octets not placed contiguously after the first at offset %d", off)
	}

	// Every appended segment begins at an absolute offset >= PrefixLength
	// (container.abnf S5 slot-offset invariant), trivially true here since
	// the prior image already fills the whole prefix, but asserted so the
	// invariant is guarded, not merely assumed.
	if len(prior) < PrefixLength {
		t.Fatalf("test base shorter than PrefixLength: %d", len(prior))
	}
}

// TestFR_056_PlaceEmptyDeltaIsZeroMutation confirms an edit with no new
// segments produces an image byte-identical to prior (the storage-layer
// no-op; the full open/save round trip is T-0035's own test).
func TestFR_056_PlaceEmptyDeltaIsZeroMutation(t *testing.T) {
	prior := makePriorImage()
	priorSnapshot := append([]byte(nil), prior...)

	out, err := place(prior, EditDelta{})
	if err != nil {
		t.Fatalf("place with empty delta: %v", err)
	}
	if !bytes.Equal(out, priorSnapshot) {
		t.Fatalf("empty-delta place did not return a byte-identical image")
	}
}

// TestFR_056_PlaceRejectsShortPriorAndEmptyPayload confirms place()
// verifies before it allocates (CP-006): a prior shorter than the fixed
// prefix and an empty segment payload are both refused with named errors
// and no image is produced.
func TestFR_056_PlaceRejectsShortPriorAndEmptyPayload(t *testing.T) {
	short := make([]byte, PrefixLength-1)
	if _, err := place(short, EditDelta{}); !errors.Is(err, ErrPriorTooShort) {
		t.Fatalf("short prior: got err %v, want ErrPriorTooShort", err)
	}

	prior := makePriorImage()
	delta := EditDelta{NewSegments: []SegmentPayload{{Octets: nil}}}
	if _, err := place(prior, delta); !errors.Is(err, ErrEmptySegmentPayload) {
		t.Fatalf("empty payload: got err %v, want ErrEmptySegmentPayload", err)
	}
}
