package eddsa

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// smallOrderTableDigestPath is the checked-in pinned SHA-256 of the
// small-order point table. If the table is ever accidentally edited, the
// recomputed digest diverges from this pinned value and the test fails,
// catching the change without re-deriving the 8 curve points by hand
// (mirrors the CON-009 ceiling-table pin pattern).
const smallOrderTableDigestPath = "testdata/eddsa-protodoc-1/small_order_table.sha256"

// tableDigest computes SHA-256 over the 8 x 32 = 256 concatenated octets of
// the small-order point table, in table order.
func tableDigest() [32]byte {
	h := sha256.New()
	for _, pt := range smallOrderPoints {
		h.Write(pt[:])
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// TestCON_015_SmallOrderTableDigestPinned is T-0104's named test. It
// recomputes the small-order table's SHA-256 and asserts it equals the
// pinned value in testdata, and sanity-checks the table's shape (8 distinct
// 32-octet entries).
func TestCON_015_SmallOrderTableDigestPinned(t *testing.T) {
	if len(smallOrderPoints) != SmallOrderPointCount || SmallOrderPointCount != 8 {
		t.Fatalf("table has %d entries, want exactly 8 (the edwards25519 torsion subgroup)", len(smallOrderPoints))
	}
	// Entries must be distinct.
	seen := map[[32]byte]bool{}
	for i, pt := range smallOrderPoints {
		if seen[pt] {
			t.Fatalf("small-order entry %d duplicates an earlier entry", i)
		}
		seen[pt] = true
	}

	got := tableDigest()
	gotHex := hex.EncodeToString(got[:])

	raw, err := os.ReadFile(smallOrderTableDigestPath)
	if err != nil {
		t.Fatalf("reading pinned digest %s: %v", smallOrderTableDigestPath, err)
	}
	want := strings.TrimSpace(string(raw))
	if gotHex != want {
		t.Fatalf("small-order table digest drifted:\n computed %s\n pinned   %s\n(if the table was intentionally changed, update the pin AND re-verify the points against RFC 8032 / Taming the many EdDSAs)", gotHex, want)
	}
}
