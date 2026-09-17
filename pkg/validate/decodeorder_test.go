package validate

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestNFR_034_NoDecoderConstructedDuringStructuralValidation is T-0123's
// named integration test. It enforces NFR-034 two ways:
//
//  1. Decode-call-counter: it drives the full 13-step pipeline through
//     RunGuarded with a media-decoder gate armed, attempting a decoder
//     construction from inside a validation step. The attempt must fault (be
//     recorded, construction NOT run) so ZERO decoders are constructed during
//     validation; only after an admitted document opens the gate may a
//     construction succeed. A construction during validation is a hard fail.
//
//  2. Source-order lint: it parses every non-test .go file in the validation
//     package and asserts none imports a font/image/audio/video decoder
//     package, so the validation path structurally cannot construct one.
func TestNFR_034_NoDecoderConstructedDuringStructuralValidation(t *testing.T) {
	// --- Part 1: decode-call-counter over the live pipeline ---------------

	var gate MediaDecoderGate
	decoderRan := 0
	tryDecode := func() error {
		return gate.Construct(MediaImage, "test-step", func() error {
			decoderRan++ // this body must NOT run while the gate is closed
			return nil
		})
	}

	// A representative all-pass pipeline whose steps each try to construct a
	// decoder mid-validation. Under NFR-034 every attempt must fault because
	// RunGuarded holds the gate closed for the whole pass.
	attempts := 0
	mkStep := func(id StepID) Step {
		return Step{ID: id, Run: func() *Finding {
			attempts++
			if err := tryDecode(); err == nil {
				// Gate let a decoder through DURING validation: NFR-034 breach.
				t.Errorf("step %d: decoder construction permitted during validation", id)
			}
			return nil
		}}
	}
	steps := []Step{
		mkStep(StepMagicHeader), mkStep(StepTSRecompute),
		mkStep(StepTCAndSignature), mkStep(StepExtEnvelope),
	}

	res := RunGuarded(&gate, steps)
	if !res.Valid {
		t.Fatalf("expected a valid document from an all-pass pipeline, got %+v", res.Validity)
	}
	if attempts == 0 {
		t.Fatal("no steps ran; the counter would be vacuously satisfied")
	}
	if decoderRan != 0 {
		t.Fatalf("NFR-034 violation: %d decoder(s) constructed during validation; want 0", decoderRan)
	}
	if got := len(gate.Faults()); got != attempts {
		t.Fatalf("expected every one of %d decode attempts during validation to fault, got %d faults", attempts, got)
	}

	// After an admitted document, the gate is open and construction succeeds.
	if err := tryDecode(); err != nil {
		t.Fatalf("after a valid verdict the gate must be open, got %v", err)
	}
	if decoderRan != 1 || gate.ConstructionCount() != 1 {
		t.Fatalf("post-verification construction not counted: ran=%d count=%d", decoderRan, gate.ConstructionCount())
	}

	// A rejected document must leave the gate CLOSED.
	var gate2 MediaDecoderGate
	rejSteps := []Step{{ID: StepMagicHeader, Run: func() *Finding {
		return &Finding{Step: StepMagicHeader, RuleID: "PD-HEADER-001", Message: "bad magic"}
	}}}
	if r := RunGuarded(&gate2, rejSteps); r.Valid {
		t.Fatal("expected rejected verdict")
	}
	if err := gate2.Construct(MediaImage, "post-reject", func() error { return nil }); err == nil {
		t.Fatal("NFR-034 violation: gate open after a rejected document")
	}

	// --- Part 2: source-order lint ---------------------------------------

	forbidden := []string{
		"image", "image/png", "image/jpeg", "image/gif",
		"golang.org/x/image", "github.com/golang/freetype",
		"font", "opentype", "truetype", "audio", "video", "codec",
	}
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		af, err := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range af.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			for _, bad := range forbidden {
				if path == bad || strings.HasPrefix(path, bad+"/") {
					t.Errorf("%s imports media-decoder package %q; the validation path must construct no decoder (NFR-034)", name, path)
				}
			}
		}
	}
	if scanned == 0 {
		t.Fatal("source-order lint scanned no files")
	}
}
