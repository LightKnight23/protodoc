package integrity

import (
	"bytes"
	"errors"
	"testing"
)

// TestFR_002_StructureDigestPreimageMatchesBitmaskExactly is T-0140's named
// test. For each of the 32 combinations of the 5 real bitmask bits, the
// assembled preimage contains exactly the items that combination specifies
// and omits the rest, with items 1/4/6 always present; a reserved bit (5-7)
// is rejected before assembly.
func TestFR_002_StructureDigestPreimageMatchesBitmaskExactly(t *testing.T) {
	// Distinctive marker payloads so presence/absence is detectable.
	hdr := bytes.Repeat([]byte{0xA1}, 480)
	ring := bytes.Repeat([]byte{0xB2}, 480)
	fm := bytes.Repeat([]byte{0xC3}, 40)
	var tsRoot Digest
	for i := range tsRoot {
		tsRoot[i] = 0xD4
	}
	summaries := []CoveredSegmentSummary{
		{Ordinal: 7, Type: 1, Length: 100, Digest: Digest{0xE5}},
	}

	for bm := 0; bm < 32; bm++ {
		in := StructureDigestInput{
			Bitmask:                  byte(bm),
			HeaderBytes:              hdr,
			RingWinnerBytes:          ring,
			LedgerLength:             0x1122334455667788,
			FrontmatterMetadataBytes: fm,
			CoveredSummaries:         summaries,
			TSRoot:                   tsRoot,
		}
		pre, err := AssembleStructureDigestPreimage(in)
		if err != nil {
			t.Fatalf("bitmask 0x%02x: assemble: %v", bm, err)
		}

		// Item 1 always: first octet is 0x0A.
		if pre[0] != DomainStructureDigest {
			t.Fatalf("bitmask 0x%02x: preimage does not start with 0x0A", bm)
		}
		// Item 4 always: the ledger-length octets appear.
		if !bytes.Contains(pre, []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}) {
			t.Fatalf("bitmask 0x%02x: ledger-length missing", bm)
		}
		// Item 6 always: the bitmask octet appears.
		if !bytes.Contains(pre, []byte{byte(bm)}) {
			t.Fatalf("bitmask 0x%02x: bitmask octet missing", bm)
		}

		checkPresence := func(bit int, marker byte, name string) {
			present := bytes.IndexByte(pre, marker) >= 0
			want := bm&bit != 0
			if present != want {
				t.Fatalf("bitmask 0x%02x: item %s present=%v, want %v", bm, name, present, want)
			}
		}
		checkPresence(CoverageBitHeader, 0xA1, "header")
		checkPresence(CoverageBitRingWinner, 0xB2, "ring-winner")
		checkPresence(CoverageBitFrontmatter, 0xC3, "frontmatter")
		checkPresence(CoverageBitIntegrity, 0xD4, "T_S root")
		// Segment summary marker 0xE5 appears iff SEGMENT_TABLE.
		checkPresence(CoverageBitSegmentTable, 0xE5, "segment-table")

		// StructureDigest hashes without error and is deterministic.
		d1, err := StructureDigest(in)
		if err != nil {
			t.Fatalf("bitmask 0x%02x: StructureDigest: %v", bm, err)
		}
		d2, _ := StructureDigest(in)
		if d1 != d2 {
			t.Fatalf("bitmask 0x%02x: StructureDigest not deterministic", bm)
		}
	}

	// A reserved bit set is rejected before assembly.
	if _, err := AssembleStructureDigestPreimage(StructureDigestInput{Bitmask: 0x20}); !errors.Is(err, ErrStructureReservedBits) {
		t.Fatalf("reserved bit: got %v, want ErrStructureReservedBits", err)
	}
}
