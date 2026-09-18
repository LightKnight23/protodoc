// Package semantics implements Protodoc's M15 accessibility & semantic content-
// model extensions. The constructs are built PROVISIONALLY against the T-0267
// design-ruling memo (specs/001-protodoc-format-core/clarify-002.md), which is
// OPEN awaiting Eyvar's ruling; they are reversible until that ruling.
//
// Constructs (added by their own tasks): base-writing-direction, the
// ROOT_SEQUENCE reading-order record, computed-inline isolation, the
// accessibility role map, table header scope + cell tiling, alt-text, ordered-
// sequence numbering, cross-reference presentation-function + staleness, 2-D
// regions, inferred-value markers, and the rule registry (its own requirement
// is owned by its later tasks).
package semantics
