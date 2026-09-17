// Writer-side NFC rejection (T-0074, CON-003). A conforming writer presented
// with text that is not already in NFC REFUSES the write and reports the
// offending unit, rather than silently converting it: silent conversion
// relocates every identity anchor (a converted run has different scalar
// counts and boundaries) and lets a document be altered after signing yet
// still verify because signer and verifier normalised different inputs. The
// writer never returns a normalised/mutated value from a non-NFC input.
package content

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ErrNonNFCText is the writer's rejection of non-NFC input (CON-003). It is
// returned instead of a converted value.
var ErrNonNFCText = errors.New("content: text is not in Unicode Normalization Form C; writer refuses to convert (CON-003)")

// NonNFCError names the offending unit whose text was not NFC, so a caller
// sees which run was rejected rather than a silent normalisation.
type NonNFCError struct {
	RunID pdlfmt.UnitID
	Text  string
}

func (e *NonNFCError) Error() string {
	return fmt.Sprintf("content: run %x carries non-NFC text %q; writer refuses to convert it (CON-003)", e.RunID, e.Text)
}

func (e *NonNFCError) Is(target error) bool { return target == ErrNonNFCText }

// WriteRun validates a run for writing: it returns the run unchanged if its
// text is already NFC, or a *NonNFCError (matching ErrNonNFCText) naming the
// run if the text is not NFC. It NEVER converts the text: the returned run,
// on success, is byte-identical to the input, so no anchor is relocated by a
// silent renormalisation (CON-003). Validation is per-run and in isolation
// (CON-002), so a run legally starting with a combining mark at its own
// segment boundary is accepted.
func WriteRun(r Run) (Run, error) {
	if !IsNFCScoped(r) {
		return Run{}, &NonNFCError{RunID: r.RunID, Text: r.Text}
	}
	return r, nil
}
