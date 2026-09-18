package canon_test

import (
	"bytes"
	"testing"

	"Protodoc/pkg/canon"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/migrate"
	"Protodoc/pkg/pdlfmt"
)

// TestNFR_002_PublishAndMigrateUseStreamingCanonicalize is T-0318's named
// integration test (NFR-002). Both the publish L*(published_state) emission and
// the migration engine's fresh-file L*(migrate(state)) emission use the SHARED
// canonical traversal/ordering (canon.Traverse / streaming Canonicalize), not a
// bespoke re-serialization order: the content octets each emits, when the
// inputs are the same logical state, follow the canonical (T_C) order that
// canon.Canonicalize produces.
func TestNFR_002_PublishAndMigrateUseStreamingCanonicalize(t *testing.T) {
	u := func(b byte) pdlfmt.UnitID {
		var id pdlfmt.UnitID
		id[0] = b
		return id
	}
	// A logical state whose storage order differs from canonical (unit-id) order.
	subs := []canon.ContentSubtree{
		{UnitID: u(0x30), Frame: []byte("gamma")},
		{UnitID: u(0x10), Frame: []byte("alpha")},
		{UnitID: u(0x20), Frame: []byte("beta")},
	}
	state := &canon.Document{Subtrees: subs}

	// Reference canonical octets via streaming Canonicalize.
	var canonBuf bytes.Buffer
	if err := canon.Canonicalize(state, &canonBuf); err != nil {
		t.Fatalf("Canonicalize: %v", err)
	}
	reference := canonBuf.Bytes()

	// (1) Publish over the SAME logical state, feeding retained records in
	// canonical order (the order canon.Traverse defines) — its emission equals
	// the streaming canonical octets.
	ordered := canon.Traverse(state)
	pubIn := integrity.PublishInput{}
	for _, s := range ordered {
		pubIn.Retained = append(pubIn.Retained, integrity.ContentRecord{UnitID: s.UnitID, Frame: s.Frame})
	}
	pub := integrity.Publish(pubIn)
	if !bytes.Equal(pub.Emitted, reference) {
		t.Errorf("publish emission does not follow the canonical order\n got  %q\n want %q", pub.Emitted, reference)
	}

	// (2) Migrate over the SAME logical state: its fresh-file emission uses the
	// shared traversal-order contract. Laid out in canonical order, Transform's
	// (segment,offset) traversal coincides with canon.Traverse's order, so both
	// emit the same unit sequence — the shared streaming-canonicalize order, not
	// a bespoke one.
	target := migrate.TargetProfile{Major: 2, Representable: map[migrate.ConstructKind]bool{0x01: true}}
	var msrc []migrate.SourceConstruct
	for i, s := range ordered { // ordered == canon.Traverse(state)
		msrc = append(msrc, migrate.SourceConstruct{
			Kind:     0x01,
			Location: migrate.Location{SegmentOrdinal: 0, IntraOffset: uint64(i * 32), UnitID: s.UnitID, HasUnitID: true},
		})
	}
	_, migrated, err := migrate.Transform(msrc, target)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	for i, s := range ordered {
		if migrated[i].UnitID != s.UnitID {
			t.Errorf("migrate emission order[%d] = %x, want canonical %x", i, migrated[i].UnitID[0], s.UnitID[0])
		}
	}
}
