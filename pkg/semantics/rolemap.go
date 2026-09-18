// Accessibility role map (FR-038; T-0279). A CLOSED enum of accessibility
// roles, and a total map from every construct kind defined in document.abnf's
// frame-discriminant registry (0x01..0x0E) onto EXACTLY ONE role. The map is
// the machine-readable form of data-model.md's role table; a companion unit
// test (TestFR_038_RoleMapCompleteness) proves it is total and single-valued.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

// ConstructKind is a document.abnf frame discriminant (the construct's kind).
type ConstructKind uint8

// The assigned construct kinds (document.abnf S1 registry, 0x01..0x0E).
const (
	KindTextBlock            ConstructKind = 0x01
	KindAnnotation           ConstructKind = 0x02
	KindTable                ConstructKind = 0x03
	KindNote                 ConstructKind = 0x04
	KindCrossReference       ConstructKind = 0x05
	KindExtEnvelope          ConstructKind = 0x06
	KindRasterImage          ConstructKind = 0x07
	KindFontSubset           ConstructKind = 0x08
	KindRegistryExcerpt      ConstructKind = 0x09
	KindUnitIndexLeaf        ConstructKind = 0x0A
	KindHistoryOpBatch       ConstructKind = 0x0B
	KindPresentationArtefact ConstructKind = 0x0C
	KindErasureRecord        ConstructKind = 0x0D
	KindRootSequence         ConstructKind = 0x0E
)

// AssignedConstructKinds is the closed set of construct kinds v1 assigns.
var AssignedConstructKinds = []ConstructKind{
	KindTextBlock, KindAnnotation, KindTable, KindNote, KindCrossReference,
	KindExtEnvelope, KindRasterImage, KindFontSubset, KindRegistryExcerpt,
	KindUnitIndexLeaf, KindHistoryOpBatch, KindPresentationArtefact,
	KindErasureRecord, KindRootSequence,
}

// Role is a CLOSED accessibility role. Every construct kind maps to exactly
// one of these; the set is fixed for v1.
type Role uint8

const (
	// RoleProse: readable running text.
	RoleProse Role = iota
	// RoleAnnotation: a durable comment/markup anchored to content.
	RoleAnnotation
	// RoleTable: tabular structure.
	RoleTable
	// RoleNote: foot/endnote content.
	RoleNote
	// RoleReference: a cross-reference / link.
	RoleReference
	// RoleGraphic: a raster or vector image.
	RoleGraphic
	// RoleStructuralNavigation: reading-order / navigation structure exposed
	// to assistive technology (ROOT_SEQUENCE).
	RoleStructuralNavigation
	// RoleNonSemantic: carries no reading-order meaning to a reader/AT
	// (resource-only or bookkeeping constructs: fonts, registry excerpts,
	// index leaves, history batches, presentation artefacts, erasure records,
	// extension envelopes whose semantics are opaque to the base role model).
	RoleNonSemantic
)

// roleMap is the total, single-valued construct-kind -> role assignment. It is
// the machine-readable mirror of data-model.md's role table.
var roleMap = map[ConstructKind]Role{
	KindTextBlock:            RoleProse,
	KindAnnotation:           RoleAnnotation,
	KindTable:                RoleTable,
	KindNote:                 RoleNote,
	KindCrossReference:       RoleReference,
	KindExtEnvelope:          RoleNonSemantic,
	KindRasterImage:          RoleGraphic,
	KindFontSubset:           RoleNonSemantic,
	KindRegistryExcerpt:      RoleNonSemantic,
	KindUnitIndexLeaf:        RoleNonSemantic,
	KindHistoryOpBatch:       RoleNonSemantic,
	KindPresentationArtefact: RoleNonSemantic,
	KindErasureRecord:        RoleNonSemantic,
	KindRootSequence:         RoleStructuralNavigation,
}

// RoleOf returns the accessibility role of a construct kind and whether the
// kind is assigned (mapped). An unassigned/unknown kind returns ok=false — the
// fail-closed behavior the T-0280 role-resolution rule relies on.
func RoleOf(kind ConstructKind) (Role, bool) {
	r, ok := roleMap[kind]
	return r, ok
}
