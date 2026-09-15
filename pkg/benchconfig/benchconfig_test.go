package benchconfig

import (
	"os"
	"strings"
	"testing"
)

// docPath locates docs/nfr-011-reference-config.md relative to this
// package directory (pkg/benchconfig), two levels up from the module root.
const docPath = "../../docs/nfr-011-reference-config.md"

// requiredFields are the NFR-011 fields the reference-config document must
// name (processor model, core count, memory, storage class and operating
// system, per NFR-011's own text), plus the Go toolchain this repo's own
// benchmark jobs run under.
var requiredFields = []string{
	"Identifier",
	"Processor model",
	"Core count",
	"Memory",
	"Storage class",
	"Operating system",
	"Go toolchain",
}

// TestNFR_011_ReferenceConfigPublished is T-0342's named test. It reads the
// live document off disk (never a hardcoded copy of its text, so the two
// cannot silently drift apart) and asserts:
//
//  1. the document exists and names every field NFR-011 requires;
//  2. its published "Identifier" table row matches ReferenceConfigID
//     exactly, so pkg/benchconfig.Stamp always cites the document that is
//     actually checked in.
func TestNFR_011_ReferenceConfigPublished(t *testing.T) {
	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("benchconfig: reading %s: %v", docPath, err)
	}
	text := string(raw)

	for _, field := range requiredFields {
		if !strings.Contains(text, field) {
			t.Errorf("benchconfig: %s: missing required NFR-011 field %q", docPath, field)
		}
	}

	wantIdentifierRow := "| Identifier | `" + ReferenceConfigID + "` |"
	if !strings.Contains(text, wantIdentifierRow) {
		t.Errorf("benchconfig: %s: does not contain the identifier row %q for ReferenceConfigID %q; the document and the pkg/benchconfig constant have drifted apart", docPath, wantIdentifierRow, ReferenceConfigID)
	}
}
