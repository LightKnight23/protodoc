package canon

import (
	"bytes"
	"io"
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/pdlfmt"
)

// countingWriter counts bytes written without retaining them, modelling a sink
// that does not force materialization of the full canonical sequence.
type countingWriter struct{ n int }

func (c *countingWriter) Write(p []byte) (int, error) { c.n += len(p); return len(p), nil }

func benchDoc(nSubtrees, subtreeBytes int) *Document {
	subs := make([]ContentSubtree, nSubtrees)
	for i := range subs {
		var id pdlfmt.UnitID
		id[0] = byte(i)
		id[1] = byte(i >> 8)
		subs[i] = ContentSubtree{UnitID: id, Frame: bytes.Repeat([]byte{byte(i)}, subtreeBytes)}
	}
	return &Document{Subtrees: subs}
}

// BenchmarkNFR_002_CanonicalizeStreamsWithoutMaterializing is T-0308's
// benchmark (NFR-002; the named test is a Benchmark per its benchmark kind). It
// canonicalizes a large document to a counting sink; peak allocation should
// scale with the largest single subtree, not total document size, because the
// stream never buffers the full canonical sequence.
func BenchmarkNFR_002_CanonicalizeStreamsWithoutMaterializing(b *testing.B) {
	benchconfig.Stamp(b, "NFR-002")
	doc := benchDoc(4096, 4096) // ~16 MiB total, 4 KiB largest subtree
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var w countingWriter
		if err := Canonicalize(doc, &w); err != nil {
			b.Fatalf("Canonicalize: %v", err)
		}
	}
}

// TestNFR_002_CanonicalizeMatchesReferenceBuffer checks streaming output is
// byte-identical to a reference in-memory serialization over the canonical
// order (the correctness half of T-0308's DoD).
func TestNFR_002_CanonicalizeMatchesReferenceBuffer(t *testing.T) {
	doc := benchDoc(64, 128)

	var streamed bytes.Buffer
	if err := Canonicalize(doc, &streamed); err != nil {
		t.Fatalf("Canonicalize: %v", err)
	}

	// Reference: buffer the whole canonical sequence explicitly.
	var ref bytes.Buffer
	for _, s := range Traverse(doc) {
		ref.Write(s.Frame)
	}
	if !bytes.Equal(streamed.Bytes(), ref.Bytes()) {
		t.Errorf("streamed output differs from reference in-memory serialization")
	}

	// BOTTOM state writes nothing.
	var empty bytes.Buffer
	if err := Canonicalize(&Document{}, &empty); err != nil || empty.Len() != 0 {
		t.Errorf("BOTTOM state must write zero bytes, wrote %d (err %v)", empty.Len(), err)
	}

	// A failing writer surfaces its error.
	if err := Canonicalize(doc, failWriter{}); err == nil {
		t.Errorf("Canonicalize must surface a writer error")
	}
}

type failWriter struct{}

func (failWriter) Write(p []byte) (int, error) { return 0, io.ErrShortWrite }
