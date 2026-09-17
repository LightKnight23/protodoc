// Media-decoder ordering guard (T-0123, NFR-034). A conforming reader must
// complete structural validation AND full integrity/signature verification
// before it constructs ANY font, image, audio, or video decoder -- because
// the decoder is where remote code execution actually happens, and parser
// state is established at construction, not merely at invocation. Embedded
// media is integrity-checked as opaque octet sequences during validation;
// only after a document is admitted may a caller construct a decoder for it.
//
// This file provides MediaDecoderGate, the single chokepoint every media
// decoder construction in the reference implementation must pass through. The
// pipeline keeps the gate CLOSED for the whole duration of Run, so any
// decoder construction attempted during validation faults deterministically
// rather than silently establishing parser state. The T-0123 integration
// test drives the full pipeline with the gate armed and asserts zero
// constructions, and a source-order lint asserts the validation package
// imports no media-decoder package.
package validate

import (
	"fmt"
	"sync"
)

// MediaKind enumerates the four decoder classes NFR-034 forbids constructing
// before verification completes.
type MediaKind int

const (
	MediaFont MediaKind = iota
	MediaImage
	MediaAudio
	MediaVideo
)

func (k MediaKind) String() string {
	switch k {
	case MediaFont:
		return "font"
	case MediaImage:
		return "image"
	case MediaAudio:
		return "audio"
	case MediaVideo:
		return "video"
	default:
		return "unknown"
	}
}

// DecoderConstructedBeforeVerificationError is raised (as the value a faulting
// gate reports) when a media decoder construction is attempted while the gate
// is closed, i.e. before structural validation and integrity/signature
// verification have completed. Its presence is a hard NFR-034 violation.
type DecoderConstructedBeforeVerificationError struct {
	Kind MediaKind
	Site string
}

func (e *DecoderConstructedBeforeVerificationError) Error() string {
	return fmt.Sprintf("validate: NFR-034 violation: %s decoder constructed at %q before verification completed", e.Kind, e.Site)
}

// MediaDecoderGate is the single chokepoint media decoder construction passes
// through. While Closed, Construct faults; the pipeline closes the gate for
// the whole of Run and only an admitted document's caller opens it. The
// zero value is a closed gate (safe default: nothing may decode until
// verification explicitly opens it).
type MediaDecoderGate struct {
	mu     sync.Mutex
	open   bool
	faults []*DecoderConstructedBeforeVerificationError
	count  int
}

// Open marks verification complete: decoder construction is now permitted.
func (g *MediaDecoderGate) Open() {
	g.mu.Lock()
	g.open = true
	g.mu.Unlock()
}

// Close re-arms the gate (decoder construction forbidden). Used to bracket a
// validation pass.
func (g *MediaDecoderGate) Close() {
	g.mu.Lock()
	g.open = false
	g.mu.Unlock()
}

// Construct is the guarded entry point every media decoder construction must
// call. If the gate is closed it records an NFR-034 violation and returns the
// error WITHOUT running construct (parser state is never established early);
// if open it runs construct. site is a human label for the construction site.
func (g *MediaDecoderGate) Construct(kind MediaKind, site string, construct func() error) error {
	g.mu.Lock()
	if !g.open {
		e := &DecoderConstructedBeforeVerificationError{Kind: kind, Site: site}
		g.faults = append(g.faults, e)
		g.mu.Unlock()
		return e
	}
	g.count++
	g.mu.Unlock()
	if construct == nil {
		return nil
	}
	return construct()
}

// Faults returns the NFR-034 violations recorded so far (decoder
// constructions attempted while the gate was closed). Empty means compliant.
func (g *MediaDecoderGate) Faults() []*DecoderConstructedBeforeVerificationError {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]*DecoderConstructedBeforeVerificationError, len(g.faults))
	copy(out, g.faults)
	return out
}

// ConstructionCount returns how many decoders were constructed through the
// gate while it was open.
func (g *MediaDecoderGate) ConstructionCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.count
}

// RunGuarded runs the validation pipeline with the media-decoder gate held
// CLOSED for the whole pass, then leaves it open iff the document is valid
// (an admitted document may have decoders constructed for it; a rejected one
// may not). This is the entry point that makes NFR-034 a structural property
// of the pipeline rather than a convention: no step can construct a decoder
// because the gate faults for the entire duration of validation.
func RunGuarded(gate *MediaDecoderGate, steps []Step) Result {
	gate.Close()
	res := Run(steps)
	if res.Valid {
		gate.Open()
	}
	return res
}
