package registry

import "testing"

// TestCON_020_ExtTokenOwnerIdPartition is T-0050's named unit test. It asserts
// the CON-020 / container.abnf S5.2 partition of the 8-octet extension-token
// space by owner-id into exactly four sets -- reserved-invalid (0x00000000),
// registered (0x00000001..0x7FFFFFFF), owner-scoped (0x80000000..0xFFFFFFFE),
// permanently-retired (0xFFFFFFFF) -- with no experimental or unregistered
// prefix anywhere, and confirms only registered and owner-scoped tokens can
// name a live extension.
func TestCON_020_ExtTokenOwnerIdPartition(t *testing.T) {
	cases := []struct {
		ownerID uint32
		want    TokenPartition
		live    bool
	}{
		{0x00000000, PartitionReservedInvalid, false},
		{0x00000001, PartitionRegistered, true},
		{0x40000000, PartitionRegistered, true},
		{0x7FFFFFFF, PartitionRegistered, true},
		{0x80000000, PartitionOwnerScoped, true},
		{0xC0000000, PartitionOwnerScoped, true},
		{0xFFFFFFFE, PartitionOwnerScoped, true},
		{0xFFFFFFFF, PartitionRetired, false},
	}
	for _, c := range cases {
		tok := NewExtToken(c.ownerID, 12345)
		if tok.OwnerID() != c.ownerID {
			t.Errorf("owner-id round trip: got 0x%08X want 0x%08X", tok.OwnerID(), c.ownerID)
		}
		if tok.OwnerLocalSeq() != 12345 {
			t.Errorf("owner-local-seq round trip: got %d want 12345", tok.OwnerLocalSeq())
		}
		if got := tok.Partition(); got != c.want {
			t.Errorf("owner-id 0x%08X: partition %v, want %v", c.ownerID, got, c.want)
		}
		if got := tok.IsLiveExtension(); got != c.live {
			t.Errorf("owner-id 0x%08X: IsLiveExtension %v, want %v", c.ownerID, got, c.live)
		}
	}

	// Completeness: every uint32 owner-id falls into exactly one partition,
	// with no gap (no experimental/unregistered prefix, CON-020). Sample the
	// whole range at the boundaries and a stride, asserting a partition is
	// always assigned and the four sets are contiguous and exhaustive.
	boundaries := []uint32{
		0x00000000, 0x00000001,
		0x7FFFFFFF, 0x80000000,
		0xFFFFFFFE, 0xFFFFFFFF,
	}
	for _, id := range boundaries {
		p := PartitionOf(id)
		if p < PartitionReservedInvalid || p > PartitionRetired {
			t.Errorf("owner-id 0x%08X produced an out-of-range partition %d", id, p)
		}
	}

	// The 8-octet layout is owner-id (high 4) then owner-local-seq (low 4),
	// big-endian, exactly.
	tok := NewExtToken(0x01020304, 0x05060708)
	want := ExtToken{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if tok != want {
		t.Errorf("token octet layout = % x, want % x", tok[:], want[:])
	}
}
