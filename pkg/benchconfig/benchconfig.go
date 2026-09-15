// Package benchconfig holds the NFR-011 reference measurement configuration
// identifier and the citation helper every NFR-012..019 benchmark job calls
// so its reported value is traceable to the configuration it was measured
// under (Audit A-REFPLAT).
package benchconfig

import "testing"

// ReferenceConfigID is the stable identifier of the NFR-011 reference
// measurement configuration published in docs/nfr-011-reference-config.md.
// It must match that document's "Identifier" table cell exactly;
// TestNFR_011_ReferenceConfigPublished fails the build the moment the two
// disagree.
const ReferenceConfigID = "PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX"

// Stamp records nfrIDs (e.g. "NFR-015, NFR-016") as measured under
// ReferenceConfigID, via b.Logf, so the benchmark's own report line always
// carries the configuration citation alongside its measured value.
// TestNFR_011_BenchmarksCiteReferenceConfig requires every func Benchmark*
// in this module to call it.
func Stamp(b *testing.B, nfrIDs string) {
	b.Helper()
	b.Logf("refcfg: %s measured under reference configuration %s (docs/nfr-011-reference-config.md)", nfrIDs, ReferenceConfigID)
}
