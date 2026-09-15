// CON-007 audit: the closed field-kind vocabulary (contracts/
// container.abnf S2, pkg/pdlfmt T-0002) and the container's fixed-prefix
// field set are enumerated here as a single closed list, so a static test
// can assert none of them is executed, interpreted, or resolved as a
// network or filesystem location.
package container

// FieldKindCategory classifies a field-kind's value semantics for the
// CON-007 audit. There is deliberately no "network-location",
// "filesystem-location", "uri", "path" or "executable" category in the
// closed set below: no field-kind in this format resolves to, or is
// executed/interpreted as, one of those (CON-007). Adding one to
// closedFieldKindCategories, or assigning it to a FieldKindVocabulary
// entry, is itself the violation TestCON_007_NoFieldResolvesLocationOrExecutes
// exists to catch.
type FieldKindCategory string

const (
	CategoryNumeric          FieldKindCategory = "numeric"
	CategoryRegistryID       FieldKindCategory = "registry-id"
	CategoryDigest           FieldKindCategory = "digest"
	CategoryOpaqueToken      FieldKindCategory = "opaque-token"
	CategoryText             FieldKindCategory = "text"
	CategorySequence         FieldKindCategory = "sequence"
	CategoryFixedEnum        FieldKindCategory = "fixed-enum"
	CategoryOpaqueBinaryBlob FieldKindCategory = "opaque-binary-blob"
	CategoryFixedArray       FieldKindCategory = "fixed-array"
	CategoryReservedPadding  FieldKindCategory = "reserved-padding"
)

// closedFieldKindCategories is the exhaustive category set CON-007
// permits a field-kind to fall into.
var closedFieldKindCategories = map[FieldKindCategory]bool{
	CategoryNumeric:          true,
	CategoryRegistryID:       true,
	CategoryDigest:           true,
	CategoryOpaqueToken:      true,
	CategoryText:             true,
	CategorySequence:         true,
	CategoryFixedEnum:        true,
	CategoryOpaqueBinaryBlob: true,
	CategoryFixedArray:       true,
	CategoryReservedPadding:  true,
}

// forbiddenFieldKindCategories names the categories CON-007 exists to
// rule out. It is checked independently of closedFieldKindCategories
// (rather than relying solely on that set's own contents) so the audit
// still catches a violation even if one of these were ever added to the
// "closed" set by mistake.
var forbiddenFieldKindCategories = map[FieldKindCategory]bool{
	"network-location":    true,
	"filesystem-location": true,
	"uri":                 true,
	"url":                 true,
	"path":                true,
	"executable":          true,
	"interpreted-code":    true,
}

// FieldKind names one closed field-kind from the PDL-TLV vocabulary
// (pkg/pdlfmt, T-0002: varint, digest256, u48, nfc-string, plain-seq-of-X,
// sorted-vec-of-X) or the container's fixed-prefix field set (Header,
// CommitRingRecord, Frontmatter, SegmentTableSlot), together with its
// semantic category.
type FieldKind struct {
	Name     string
	Category FieldKindCategory
}

// FieldKindVocabulary is the closed, exhaustive field-kind vocabulary
// this package and pkg/pdlfmt currently implement (contracts/
// container.abnf S1-S5). It is the single source of truth
// TestCON_007_NoFieldResolvesLocationOrExecutes walks: give a new
// field-kind an entry here and the audit covers it with no change to the
// test itself.
var FieldKindVocabulary = []FieldKind{
	// pkg/pdlfmt closed primitives (T-0002).
	{"varint", CategoryNumeric},
	{"digest256", CategoryDigest},
	{"u48", CategoryNumeric},
	{"nfc-string", CategoryText},
	{"plain-seq-of-X", CategorySequence},
	{"sorted-vec-of-X", CategorySequence},

	// Header (contracts/container.abnf S1).
	{"magic", CategoryFixedArray},
	{"format-major", CategoryNumeric},
	{"format-minor", CategoryNumeric},
	{"document-class", CategoryFixedEnum},
	{"capability-written", CategoryRegistryID},
	{"capability-required", CategoryRegistryID},
	{"durable-claim", CategoryFixedEnum},
	{"history-mode", CategoryFixedEnum},
	{"unicode-version-id", CategoryRegistryID},
	{"shaping-profile-id", CategoryRegistryID},
	{"prefix-layout-id", CategoryNumeric},
	{"header-reserved", CategoryReservedPadding},
	{"header-digest", CategoryDigest},

	// CommitRingRecord (contracts/container.abnf S3).
	{"ring-magic", CategoryFixedArray},
	{"sequence", CategoryNumeric},
	{"ledger-length", CategoryNumeric},
	{"ledger-root", CategoryDigest},
	{"segment-count", CategoryNumeric},
	{"state-id", CategoryOpaqueToken},
	{"frontmatter-digest", CategoryDigest},
	{"segment-table-digest", CategoryDigest},
	{"t-c-root", CategoryDigest},
	{"parent-state-id", CategoryOpaqueToken},
	{"retention-point", CategoryNumeric},
	{"compaction-generation", CategoryNumeric},
	{"index-route", CategoryFixedArray},
	{"structure-digest", CategoryDigest},
	{"ring-reserved", CategoryReservedPadding},
	{"record-digest", CategoryDigest},

	// Frontmatter (contracts/container.abnf S4).
	{"fm-preview-kind", CategoryFixedEnum},
	{"fm-preview-raster", CategoryOpaqueBinaryBlob},
	{"fm-preview-digest", CategoryDigest},
	{"fm-source-snapshot", CategoryOpaqueBinaryBlob},
	{"fm-title", CategoryText},
	{"fm-page-count", CategoryNumeric},
	{"fm-page-dimensions", CategoryNumeric},
	{"fm-language", CategoryText},
	{"fm-colour-profile-id", CategoryRegistryID},
	{"fm-retired-tokens", CategorySequence},
	{"ext-token", CategoryOpaqueToken},
	{"fm-coverage-summary", CategorySequence},
	{"frontmatter-padding", CategoryReservedPadding},

	// SegmentTable (contracts/container.abnf S5).
	{"slot-segment-type", CategoryFixedEnum},
	{"slot-flags", CategoryFixedEnum},
	{"slot-offset", CategoryNumeric},
	{"slot-length", CategoryNumeric},
	{"slot-frame-count", CategoryNumeric},
	{"slot-digest", CategoryDigest},

	// Shared opaque identity token (CON-008).
	{"unit-id", CategoryOpaqueToken},
}
