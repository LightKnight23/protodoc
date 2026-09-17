package registry

import "testing"

// TestCON_020_OwnerIdPartitionBoundaries is T-0065's named conformance test
// (CON-020). It exercises the exact owner-id boundary values between the four
// partitions (reserved-invalid, registered, owner-scoped, permanently-retired)
// as a corpus, asserting each boundary and its immediate neighbours fall in
// the correct partition -- so the partition edges (container.abnf S5.2) are
// contiguous and exhaustive with no off-by-one gap or overlap, and no
// experimental/unregistered prefix exists.
func TestCON_020_OwnerIdPartitionBoundaries(t *testing.T) {
	corpus := []struct {
		ownerID uint32
		want    TokenPartition
		note    string
	}{
		{0x00000000, PartitionReservedInvalid, "reserved-invalid singleton"},
		{0x00000001, PartitionRegistered, "first registered"},
		{0x7FFFFFFE, PartitionRegistered, "penultimate registered"},
		{0x7FFFFFFF, PartitionRegistered, "last registered"},
		{0x80000000, PartitionOwnerScoped, "first owner-scoped"},
		{0x80000001, PartitionOwnerScoped, "second owner-scoped"},
		{0xFFFFFFFD, PartitionOwnerScoped, "penultimate owner-scoped"},
		{0xFFFFFFFE, PartitionOwnerScoped, "last owner-scoped"},
		{0xFFFFFFFF, PartitionRetired, "permanently-retired singleton"},
	}

	// Every corpus vector lands in its expected partition.
	for _, c := range corpus {
		if got := PartitionOf(c.ownerID); got != c.want {
			t.Errorf("%s: owner-id 0x%08X partition = %v, want %v", c.note, c.ownerID, got, c.want)
		}
	}

	// Boundary adjacency: each partition transition happens at exactly the
	// documented edge (the value below the edge is the lower partition, the
	// value at the edge is the higher).
	transitions := []struct {
		below, at uint32
		lower, up TokenPartition
	}{
		{0x00000000, 0x00000001, PartitionReservedInvalid, PartitionRegistered},
		{0x7FFFFFFF, 0x80000000, PartitionRegistered, PartitionOwnerScoped},
		{0xFFFFFFFE, 0xFFFFFFFF, PartitionOwnerScoped, PartitionRetired},
	}
	for _, tr := range transitions {
		if PartitionOf(tr.below) != tr.lower {
			t.Errorf("owner-id 0x%08X should be %v", tr.below, tr.lower)
		}
		if PartitionOf(tr.at) != tr.up {
			t.Errorf("owner-id 0x%08X should be %v (partition edge)", tr.at, tr.up)
		}
	}

	// Only registered and owner-scoped boundaries name live extensions.
	live := map[TokenPartition]bool{PartitionRegistered: true, PartitionOwnerScoped: true}
	for _, c := range corpus {
		tok := NewExtToken(c.ownerID, 0)
		if got := tok.IsLiveExtension(); got != live[c.want] {
			t.Errorf("%s: IsLiveExtension = %v, want %v", c.note, got, live[c.want])
		}
	}
}
