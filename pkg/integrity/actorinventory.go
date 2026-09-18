// Actor-identity inventory (T-0200, FR-079). A SINGLE enumerable inventory of
// every field carrying actor identity, actor-device identity, or per-actor
// edit attribution -- so the publish operation can strip exactly those fields
// (FR-080) and a reviewer can audit the complete set. The inventory is scoped
// to the fields that exist in the current frozen data-model; the open
// clarification RR-FR-079 (clarify.md) tracks the disputed live-annotation
// authorship field, which must be added here if a ruling introduces it.
package integrity

// ActorFieldKind classifies why a field is in the actor-identity inventory.
type ActorFieldKind int

const (
	// ActorIdentity: a field naming who an actor is (an author/authorship id).
	ActorIdentity ActorFieldKind = iota
	// ActorDeviceIdentity: a field naming an actor's device.
	ActorDeviceIdentity
	// PerActorAttribution: a field attributing an edit to a specific actor.
	PerActorAttribution
)

func (k ActorFieldKind) String() string {
	switch k {
	case ActorIdentity:
		return "actor-identity"
	case ActorDeviceIdentity:
		return "actor-device-identity"
	case PerActorAttribution:
		return "per-actor-attribution"
	default:
		return "unknown"
	}
}

// ActorField is one entry in the inventory: the record.field path and its kind.
type ActorField struct {
	Path string // e.g. "Annotation.orphan.author_ref"
	Kind ActorFieldKind
}

// actorIdentityInventory is the single, complete inventory of actor-identity-
// carrying fields in the current frozen data-model (FR-079). It is built by
// ActorIdentityInventory so callers see one authoritative list.
//
// KNOWN ENTRY: Annotation.orphan.author_ref (OrphanRecord.Author) -- the
// authorship id captured when an annotation is orphaned (data-model.md 2.13,
// document.abnf). This is the only actor-identity-carrying field in the frozen
// model; run_id explicitly carries NO actor-derived value (FR-023/CQ-004), and
// the disputed live-annotation authorship field is tracked by RR-FR-079.
func actorIdentityInventory() []ActorField {
	return []ActorField{
		{Path: "Annotation.orphan.author_ref", Kind: ActorIdentity},
	}
}

// ActorIdentityInventory returns the complete actor-identity field inventory
// (FR-079). The returned slice is a fresh copy the caller may not mutate to
// affect the canonical list.
func ActorIdentityInventory() []ActorField {
	src := actorIdentityInventory()
	out := make([]ActorField, len(src))
	copy(out, src)
	return out
}

// IsActorIdentityField reports whether the given record.field path is in the
// actor-identity inventory. The publish operation (FR-080) uses this to decide
// which values to strip.
func IsActorIdentityField(path string) bool {
	for _, f := range actorIdentityInventory() {
		if f.Path == path {
			return true
		}
	}
	return false
}
