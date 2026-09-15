// Package benchconfig holds the NFR-011 reference measurement configuration
// identifier every NFR-012..019 benchmark job cites, so its reported value
// is traceable to the configuration it was measured under (Audit
// A-REFPLAT).
package benchconfig

// ReferenceConfigID is the stable identifier of the NFR-011 reference
// measurement configuration published in docs/nfr-011-reference-config.md.
// It must match that document's "Identifier" table cell exactly;
// TestNFR_011_ReferenceConfigPublished fails the build the moment the two
// disagree.
const ReferenceConfigID = "PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX"
