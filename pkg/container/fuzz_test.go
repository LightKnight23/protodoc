// CP-012 continuous fuzzing (T-0359): four of the six untrusted-byte
// decode entry points the constitution's continuous-fuzzing cross-cutting
// concern names M01 as owning (Header, CommitRingRecord, Frontmatter,
// SegmentTable — PDL-VARINT and PDL-TLV's fuzz targets live in
// pkg/pdlfmt, the package that owns those two primitives). Every target
// here asserts exactly one property regardless of input: the decoder
// returns an ordinary error on malformed input and never panics — decode
// is a trust boundary (CP-006), and CP-012 exists to catch a panic there
// before it ever reaches a caller.
package container

import "testing"

// FuzzCP_012_DecodeUntrustedBytes_Header seeds from header_test.go's
// validHeader fixture (well-formed) and its digest-mismatch/truncated
// rejection vectors (malformed).
func FuzzCP_012_DecodeUntrustedBytes_Header(f *testing.F) {
	valid := validHeader().Encode(nil)
	f.Add(valid)

	corruptMagic := append([]byte(nil), valid...)
	corruptMagic[0] = 'X' // digest now stale too: exercises the digest-mismatch path
	f.Add(corruptMagic)

	f.Add(make([]byte, HeaderSize)) // all-zero: digest mismatches (sha256 of zeros is not zero)
	f.Add(make([]byte, 100))        // truncated
	f.Add([]byte{})                 // empty

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeHeader(data)
	})
}

// FuzzCP_012_DecodeUntrustedBytes_CommitRingRecord seeds from
// commitring_test.go's fixtureRingRecord (well-formed, self-digesting)
// alongside an all-zero never-written slot and a truncated one — both
// legitimate inputs DecodeCommitRingRecord must decode without error
// (self-digest verification is VerifyRingSlotDigest's separate concern,
// per this package's own decode/verify split) or reject cleanly.
func FuzzCP_012_DecodeUntrustedBytes_CommitRingRecord(f *testing.F) {
	rec := fixtureRingRecord(1, StateID{})
	f.Add(rec.Encode(nil))
	f.Add(make([]byte, RingSlotSize))   // all-zero, never-written slot
	f.Add(make([]byte, RingSlotSize-1)) // truncated
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeCommitRingRecord(data)
	})
}

// FuzzCP_012_DecodeUntrustedBytes_Frontmatter seeds from
// frontmatter_test.go's fixtureFrontmatter (with and without the preview
// payload) and from its truncated/nonzero-padding rejection vectors.
func FuzzCP_012_DecodeUntrustedBytes_Frontmatter(f *testing.F) {
	withPreview := fixtureFrontmatter()
	withPreview.PreviewKind = PreviewKindPLP1
	withPreview.PreviewRaster = []byte{1, 2, 3}
	withPreview.PreviewDigest = ComputeFrontmatterPreviewDigest(withPreview)
	withPreview.ColourProfileID = ColourProfileSRGBD65
	if enc, err := withPreview.Encode(nil); err != nil {
		f.Fatalf("Encode: %v", err)
	} else {
		f.Add(enc)
	}

	empty := &Frontmatter{}
	if enc, err := empty.Encode(nil); err != nil {
		f.Fatalf("Encode: %v", err)
	} else {
		f.Add(enc)
		nonzeroPadding := append([]byte(nil), enc...)
		nonzeroPadding[FrontmatterRegionSize-1] = 0xFF
		f.Add(nonzeroPadding)
	}

	f.Add(make([]byte, FrontmatterRegionSize-1)) // truncated
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeFrontmatter(data)
	})
}

// FuzzCP_012_DecodeUntrustedBytes_SegmentTable seeds from
// segmenttable_test.go's fixtureSlot fixture, the all-unused (all-zero)
// table, and a truncated region. Every mutated corpus entry still lands
// on DecodeSegmentTable's actual trust boundary, the exported entry point
// bounded-prefix readers call, rather than the unexported per-slot helper.
func FuzzCP_012_DecodeUntrustedBytes_SegmentTable(f *testing.F) {
	if enc, err := EncodeSegmentTable([]SegmentTableSlot{fixtureSlot(1), fixtureSlot(2)}, nil); err != nil {
		f.Fatalf("EncodeSegmentTable: %v", err)
	} else {
		f.Add(enc)
	}
	if allUnused, err := EncodeSegmentTable(nil, nil); err != nil {
		f.Fatalf("EncodeSegmentTable: %v", err)
	} else {
		f.Add(allUnused)
	}
	f.Add(make([]byte, SegmentTableRegionSize-1)) // truncated
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeSegmentTable(data)
	})
}
