package integrity

import (
	"bytes"
	"testing"
)

// TestFR_080_PublishStripsActorIdentityValues is T-0201's named integration
// test (FR-080). WHEN publish runs, the output contains no value from the
// actor-identity inventory (FR-079): every actor-identity value is stripped
// from the emitted octets, while non-identity content is retained.
func TestFR_080_PublishStripsActorIdentityValues(t *testing.T) {
	authorID := []byte("author:alice@example.com") // an Annotation.orphan.author_ref value
	authorID2 := []byte("author:bob@example.com")
	content := []byte("the surviving public content of the document")

	// A frame that embeds an actor-identity value (as an orphaned annotation's
	// author_ref would).
	frameWithActor := append(append([]byte("annotation body "), authorID...), []byte(" trailing")...)

	in := PublishInput{
		Retained: []ContentRecord{
			{UnitID: redUnitID(0x01), Frame: content},
			{UnitID: redUnitID(0x02), Frame: frameWithActor},
		},
		ActorIdentityValues: [][]byte{authorID, authorID2},
	}

	out := Publish(in)

	// No actor-identity value survives.
	for _, v := range in.ActorIdentityValues {
		if bytes.Contains(out.Emitted, v) {
			t.Errorf("actor-identity value %q survived publish", v)
		}
	}

	// Non-identity content is retained.
	if !bytes.Contains(out.Emitted, content) {
		t.Error("non-identity content was lost during publish")
	}
	if !bytes.Contains(out.Emitted, []byte("annotation body ")) || !bytes.Contains(out.Emitted, []byte(" trailing")) {
		t.Error("non-identity parts of the actor-bearing frame were lost")
	}

	// The inventory drives what gets stripped: a value NOT in the inventory
	// stays. (Publish only strips what it is told, and the caller derives the
	// values from ActorIdentityInventory.)
	keep := []byte("KEEP-THIS-CUSTODY-VALUE")
	in2 := PublishInput{
		Retained:            []ContentRecord{{UnitID: redUnitID(0x03), Frame: append(append([]byte(nil), keep...), authorID...)}},
		ActorIdentityValues: [][]byte{authorID},
	}
	out2 := Publish(in2)
	if bytes.Contains(out2.Emitted, authorID) {
		t.Error("actor-identity value survived in the second publish")
	}
	if !bytes.Contains(out2.Emitted, keep) {
		t.Error("a non-inventory value was wrongly stripped")
	}

	// The inventory itself is the source of truth for which paths are actor
	// identity (used by a caller to collect the values).
	if !IsActorIdentityField("Annotation.orphan.author_ref") {
		t.Error("the actor-identity inventory does not recognise the known field")
	}
}
