package canon

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// recoverCanonicalFromText reconstructs the canonical octet sequence from a
// text projection. It is TEST-ONLY (defined in a _test.go file) and is never
// wired into any consumer, per TR-004 — it exists solely to prove round-trip
// fidelity.
func recoverCanonicalFromText(t *testing.T, text string) []byte {
	t.Helper()
	var out []byte
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) != 2 {
			t.Fatalf("bad projection line %q", line)
		}
		frame, err := hex.DecodeString(cols[1])
		if err != nil {
			t.Fatalf("decode frame hex: %v", err)
		}
		out = append(out, frame...)
	}
	return out
}

// recoverCanonicalFromHTML reconstructs the canonical octet sequence from an
// html projection. TEST-ONLY.
func recoverCanonicalFromHTML(t *testing.T, html string) []byte {
	t.Helper()
	var out []byte
	for _, seg := range strings.Split(html, `<div data-unit="`)[1:] {
		close := strings.Index(seg, `">`)
		end := strings.Index(seg, `</div>`)
		if close < 0 || end < 0 {
			t.Fatalf("malformed div in html")
		}
		frame, err := hex.DecodeString(seg[close+2 : end])
		if err != nil {
			t.Fatalf("decode frame hex: %v", err)
		}
		out = append(out, frame...)
	}
	return out
}

// TestTR_004_ProjectionRoundTripsToIdenticalCanonicalOctets is T-0321's named
// unit test (TR-004). Recovering canonical octets from a text or html
// projection reproduces the exact streaming-Canonicalize output.
func TestTR_004_ProjectionRoundTripsToIdenticalCanonicalOctets(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	state := &Document{Subtrees: []ContentSubtree{
		{UnitID: u(0x30), Frame: []byte("gamma")},
		{UnitID: u(0x10), Frame: []byte("alpha")},
		{UnitID: u(0x20), Frame: []byte("beta")},
	}}

	var canonBuf bytes.Buffer
	if err := Canonicalize(state, &canonBuf); err != nil {
		t.Fatalf("Canonicalize: %v", err)
	}
	want := canonBuf.Bytes()

	if got := recoverCanonicalFromText(t, ProjectText(state)); !bytes.Equal(got, want) {
		t.Errorf("text round-trip: got %q, want %q", got, want)
	}
	if got := recoverCanonicalFromHTML(t, ProjectHTML(state)); !bytes.Equal(got, want) {
		t.Errorf("html round-trip: got %q, want %q", got, want)
	}
}
