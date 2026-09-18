// Package migrate implements Protodoc document migration between format major
// versions (the FR-119 migration family). This file provides Scan (FR-121): a
// non-mutating, deterministic pre-migration walk that halts at the FIRST
// construct not representable in the target major version's construct set and
// returns a RefusalReport naming that construct's kind and exact location. No
// approximation or partial migration is ever attempted.
package migrate

import (
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ConstructKind identifies a source construct's kind for representability. It
// mirrors document.abnf's frame-discriminant registry plus the cross-cutting
// non-structural kinds (retired extension token, out-of-allowlist crypto
// parameter) that migration must also refuse.
type ConstructKind uint16

// Location is a construct's physical position: segment ordinal + intra-segment
// octet offset (matching validate.Location's shape), plus the unit id where
// one applies.
type Location struct {
	SegmentOrdinal uint16
	IntraOffset    uint64
	UnitID         pdlfmt.UnitID
	HasUnitID      bool
}

func (l Location) String() string {
	if l.HasUnitID {
		return fmt.Sprintf("segment %d offset %d unit %x", l.SegmentOrdinal, l.IntraOffset, l.UnitID[:4])
	}
	return fmt.Sprintf("segment %d offset %d", l.SegmentOrdinal, l.IntraOffset)
}

// SourceConstruct is one construct encountered in the source document's
// deterministic traversal order.
type SourceConstruct struct {
	Kind     ConstructKind
	KindName string
	Location Location
}

// RefusalReport is Scan's result. Refused is true when an unrepresentable
// construct was found; the report then names that construct's kind and
// location. An empty (Refused=false) report is the proceed signal.
type RefusalReport struct {
	Refused  bool
	Kind     ConstructKind
	KindName string
	Location Location
}

// TargetProfile describes a target major version's construct set: the set of
// construct kinds representable in it. A construct kind absent from
// Representable is unrepresentable and triggers refusal.
type TargetProfile struct {
	Major         uint16
	Representable map[ConstructKind]bool
}

// Scan walks source in the given deterministic traversal order and halts at the
// FIRST construct not representable in target's construct set, returning a
// RefusalReport naming it. It performs NO mutation and attempts NO
// approximation. With zero unrepresentable constructs it returns an empty
// report (proceed). With several, it names ONLY the first in traversal order.
//
// The caller supplies source constructs already in the defined traversal order
// (the same ordering validate uses for error precedence, FR-103): ascending
// segment ordinal, then ascending intra-segment offset.
func Scan(source []SourceConstruct, target TargetProfile) RefusalReport {
	for _, c := range source {
		if !target.Representable[c.Kind] {
			return RefusalReport{
				Refused:  true,
				Kind:     c.Kind,
				KindName: c.KindName,
				Location: c.Location,
			}
		}
	}
	return RefusalReport{}
}

// TraversalLess reports whether construct a precedes b in the defined
// deterministic traversal order (segment ordinal, then intra-offset). Callers
// use it to sort source constructs before Scan when they are not already
// ordered.
func TraversalLess(a, b SourceConstruct) bool {
	if a.Location.SegmentOrdinal != b.Location.SegmentOrdinal {
		return a.Location.SegmentOrdinal < b.Location.SegmentOrdinal
	}
	return a.Location.IntraOffset < b.Location.IntraOffset
}
