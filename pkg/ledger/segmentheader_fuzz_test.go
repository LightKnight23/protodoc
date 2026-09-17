package ledger

import (
	"bytes"
	"testing"

	"Protodoc/pkg/container"
)

// FuzzFR_056_SegmentFrameDecode is T-0047's named continuous-fuzzing
// target (CP-012). It feeds arbitrary byte sequences into the ledger
// segment-header decode path and asserts, for every input, that the
// decoder: never panics, either parses successfully or returns a
// structured error, and never mutates any octet of the caller's input
// buffer (a decoder is a read-only trust boundary). It is seeded with a
// well-formed header, at-limit and over-limit frame-count fixtures, and
// truncated/empty inputs, mirroring the M01 conformance-fixture seeding
// pattern the container package's own CP-012 targets use.
func FuzzFR_056_SegmentFrameDecode(f *testing.F) {
	// Well-formed header.
	f.Add(EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeContent, CapabilityGen: 1, FrameCount: 3, PayloadLength: 4096}))

	// At the frame-count ceiling (legal boundary).
	f.Add(EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeResource, FrameCount: container.MaxFramesPerSegment, PayloadLength: 1}))

	// One past the frame-count ceiling (must reject, not panic).
	overLimit := EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeContent, FrameCount: container.MaxFramesPerSegment, PayloadLength: 1})
	over := uint32(container.MaxFramesPerSegment) + 1
	overLimit[8] = byte(over >> 24)
	overLimit[9] = byte(over >> 16)
	overLimit[10] = byte(over >> 8)
	overLimit[11] = byte(over)
	f.Add(overLimit)

	// Bad magic, bad type, nonzero reserved.
	badMagic := EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeContent, FrameCount: 1})
	badMagic[0] = 'X'
	f.Add(badMagic)
	badType := EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeContent, FrameCount: 1})
	badType[4] = 0 // unused type, illegal for a real segment header
	f.Add(badType)
	nonzeroReserved := EncodeSegmentHeader(SegmentHeader{Type: container.SegmentTypeContent, FrameCount: 1})
	nonzeroReserved[7] = 0xFF
	f.Add(nonzeroReserved)

	// Truncated and empty.
	f.Add(make([]byte, SegmentHeaderSize-1))
	f.Add(make([]byte, 4))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Snapshot to detect any mutation of the caller's buffer.
		before := append([]byte(nil), data...)

		// Must not panic; result is ignored (the property is "no panic,
		// structured error or success").
		_, _ = DecodeSegmentHeader(data)

		if !bytes.Equal(data, before) {
			t.Fatalf("DecodeSegmentHeader mutated its input buffer")
		}
	})
}
