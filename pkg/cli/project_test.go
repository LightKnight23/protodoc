package cli

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/pdlfmt"
)

// recoverFromText reconstructs canonical octets from the text projection
// (test-only reverse mapping, used solely for round-trip verification).
func recoverFromText(t *testing.T, text string) []byte {
	t.Helper()
	var out []byte
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		frame, err := hex.DecodeString(cols[len(cols)-1])
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		out = append(out, frame...)
	}
	return out
}

// TestTR_004_ProjectRoundTripsExactly is T-0334's named integration test
// (TR-004). `protodoc project` output round-trips (via the test's reverse
// mapping) to the identical canonical octet sequence for 3 fixtures of
// increasing size; the projection is rejected as a document by other verbs.
func TestTR_004_ProjectRoundTripsExactly(t *testing.T) {
	orig := ProjectStateFor
	defer func() { ProjectStateFor = orig }()

	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	fixtures := []*canon.Document{
		{Subtrees: []canon.ContentSubtree{{UnitID: u(0x10), Frame: []byte("a")}}},
		{Subtrees: []canon.ContentSubtree{{UnitID: u(0x20), Frame: []byte("bb")}, {UnitID: u(0x10), Frame: []byte("a")}}},
		{Subtrees: []canon.ContentSubtree{{UnitID: u(0x30), Frame: []byte("ccc")}, {UnitID: u(0x10), Frame: []byte("a")}, {UnitID: u(0x20), Frame: []byte("bb")}}},
	}

	for i, state := range fixtures {
		ProjectStateFor = func(string) (*canon.Document, error) { return state, nil }

		res := runProject([]string{"doc.pdl", "--format=text"}, nil)
		if res.Status != "OK" {
			t.Fatalf("fixture %d: status=%s, want OK", i, res.Status)
		}
		projection := res.Extra["projection"].(string)

		var canonBuf bytes.Buffer
		if err := canon.Canonicalize(state, &canonBuf); err != nil {
			t.Fatalf("fixture %d: Canonicalize: %v", i, err)
		}
		if got := recoverFromText(t, projection); !bytes.Equal(got, canonBuf.Bytes()) {
			t.Errorf("fixture %d: round-trip mismatch\n got  %q\n want %q", i, got, canonBuf.Bytes())
		}

		// The projection carries no Protodoc magic (rejected as a document).
		if len(projection) >= 4 && projection[:4] == "PDL1" {
			t.Errorf("fixture %d: projection must not begin with the Protodoc magic", i)
		}
	}

	// Missing file -> USAGE.
	if r := runProject(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg project: status=%s, want USAGE", r.Status)
	}
}
