package integrity

import (
	"reflect"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestTR_009_TSRootAlwaysRecomputedNeverCached is T-0131's named test.
// TSRoot takes no cached-root parameter; tampering with a SegmentTableSlot's
// stored slot-digest changes TSRoot's output (it reflects live octets, not a
// cached root); and CompareLedgerRoot reports divergence for a stored
// ledger_root that no longer matches, never silently accepting it.
func TestTR_009_TSRootAlwaysRecomputedNeverCached(t *testing.T) {
	// TSRoot's signature accepts only []SegmentTableSlot -- no cached-root
	// parameter.
	ft := reflect.TypeOf(TSRoot)
	if ft.NumIn() != 1 {
		t.Fatalf("TSRoot takes %d parameters, want 1 (no cached-root parameter)", ft.NumIn())
	}

	slots := make([]container.SegmentTableSlot, 3)
	for i := range slots {
		var dig pdlfmt.Digest256
		dig[0] = byte(i + 1)
		slots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeContent, Digest: dig}
	}

	root1, err := TSRoot(slots)
	if err != nil {
		t.Fatalf("TSRoot: %v", err)
	}

	// The stored ledger_root matching the fresh root => Matches.
	if v, fresh, err := CompareLedgerRoot(slots, root1); err != nil || v != LedgerRootMatches || fresh != root1 {
		t.Fatalf("CompareLedgerRoot(matching) = (%v,%x,%v), want Matches", v, fresh, err)
	}

	// Tamper with one slot's stored digest: TSRoot reflects the tampered
	// value (recomputes from live octets), so it differs from root1.
	slots[1].Digest[5] ^= 0xFF
	root2, err := TSRoot(slots)
	if err != nil {
		t.Fatalf("TSRoot after tamper: %v", err)
	}
	if root2 == root1 {
		t.Fatalf("TSRoot did not reflect the tampered slot-digest (returned a cached root?)")
	}

	// A stale stored ledger_root (root1) no longer matches the tampered
	// state: CompareLedgerRoot reports divergence, never silently accepts,
	// and the fresh value is the tampered root2.
	v, fresh, err := CompareLedgerRoot(slots, root1)
	if err != nil {
		t.Fatalf("CompareLedgerRoot(stale): %v", err)
	}
	if v != LedgerRootDiverges {
		t.Fatalf("stale ledger_root reported %v, want LedgerRootDiverges", v)
	}
	if fresh != root2 {
		t.Fatalf("CompareLedgerRoot returned fresh %x, want the recomputed tampered root %x", fresh, root2)
	}
}
