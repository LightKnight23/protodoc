package integrity

import "testing"

// TestFR_079_ActorFieldInventoryEnumeratesKnownFields is T-0200's named unit
// test (FR-079). The actor-identity inventory is a single enumerable list of
// every field carrying actor identity / device identity / per-actor
// attribution in the current frozen data-model. It enumerates the known field
// (Annotation.orphan.author_ref), recognises it via IsActorIdentityField, does
// not claim a field that carries no actor value (run_id, per FR-023/CQ-004),
// and returns a defensively-copied list.
func TestFR_079_ActorFieldInventoryEnumeratesKnownFields(t *testing.T) {
	inv := ActorIdentityInventory()
	if len(inv) == 0 {
		t.Fatal("actor-identity inventory is empty; it must enumerate the known actor field(s)")
	}

	// The known actor-identity field is present and classified as identity.
	found := false
	for _, f := range inv {
		if f.Path == "Annotation.orphan.author_ref" {
			found = true
			if f.Kind != ActorIdentity {
				t.Errorf("orphan.author_ref classified %v, want actor-identity", f.Kind)
			}
		}
	}
	if !found {
		t.Error("inventory does not enumerate Annotation.orphan.author_ref")
	}

	// IsActorIdentityField agrees with the inventory.
	if !IsActorIdentityField("Annotation.orphan.author_ref") {
		t.Error("IsActorIdentityField missed a listed field")
	}

	// A field that carries NO actor value is not in the inventory. run_id
	// explicitly carries no actor-derived value (FR-023/CQ-004), so it must
	// not appear.
	for _, notActor := range []string{"Run.run_id", "Run.base_ordinal", "TextBlock.block_id", "not.a.field"} {
		if IsActorIdentityField(notActor) {
			t.Errorf("IsActorIdentityField wrongly claimed %q as an actor-identity field", notActor)
		}
	}

	// The returned list is a copy: mutating it does not change the canonical
	// inventory.
	inv[0].Path = "MUTATED"
	if ActorIdentityInventory()[0].Path == "MUTATED" {
		t.Error("ActorIdentityInventory returned a mutable reference to the canonical list")
	}
}
