// Deterministic migration transform (FR-119; T-0298). Transform migrates a
// Phase-1-clean source (empty RefusalReport) into the target major version's
// canonical octet sequence deterministically: the output is a pure function of
// (source, targetMajor) — no map-iteration order, no wall-clock, no
// goroutine/process-order dependency. The traversal-ordering contract is the
// same defined order Scan uses (segment ordinal, then intra-segment offset).
package migrate

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// ErrSourceNotClean is returned when Transform is called on a source that
// Scan would refuse; migration must not proceed past a refusal.
var ErrSourceNotClean = errors.New("migrate: source is not Phase-1-clean (Scan refuses it)")

// MigratedConstruct is one construct in the migrated output: its kind, its
// preserved unit id (threaded unchanged from source, FR-120), and its canonical
// payload octets.
type MigratedConstruct struct {
	Kind    ConstructKind
	UnitID  pdlfmt.UnitID
	Payload []byte
}

// sortedByTraversal returns source constructs in the defined deterministic
// traversal order without mutating the input. It is a stable insertion sort on
// (segment ordinal, intra-offset) so the ordering never depends on the input
// slice order or any map iteration.
func sortedByTraversal(source []SourceConstruct) []SourceConstruct {
	out := make([]SourceConstruct, len(source))
	copy(out, source)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && TraversalLess(out[j], out[j-1]); j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// Transform migrates a Phase-1-clean source into the target's canonical octet
// sequence. It first re-runs Scan; if the source is not clean it returns
// ErrSourceNotClean and no output. Otherwise it emits each construct in the
// defined traversal order, threading every unit id unchanged (FR-120), and
// returns the canonical concatenation of a fixed per-construct framing. The
// result is a deterministic function of (source, target): identical inputs
// yield byte-identical output across invocations, processes and machines.
func Transform(source []SourceConstruct, target TargetProfile) ([]byte, []MigratedConstruct, error) {
	if r := Scan(source, target); r.Refused {
		return nil, nil, ErrSourceNotClean
	}
	ordered := sortedByTraversal(source)

	var out []byte
	migrated := make([]MigratedConstruct, 0, len(ordered))
	for _, c := range ordered {
		mc := MigratedConstruct{Kind: c.Kind, UnitID: c.Location.UnitID}
		// Canonical per-construct framing: kind (varint), a presence byte for
		// the unit id, then the 16-octet unit id when present. No timestamps,
		// no map iteration, no host state.
		out = pdlfmt.AppendVarint(out, uint64(c.Kind))
		if c.Location.HasUnitID {
			out = append(out, 0x01)
			out = append(out, c.Location.UnitID[:]...)
			mc.UnitID = c.Location.UnitID
		} else {
			out = append(out, 0x00)
		}
		migrated = append(migrated, mc)
	}
	return out, migrated, nil
}
