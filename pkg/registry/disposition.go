package registry

import (
	"errors"
	"fmt"
)

// ext-disposition structural validation (T-0054, FR-013/FR-014; document.abnf
// S6). ext-disposition is a u8 that MUST be exactly one of 0x00 (ignore),
// 0x01 (degrade), 0x02 (refuse). A missing disposition (caught at decode as a
// missing required field) or an out-of-range value (0x03..0xFF) is a
// STRUCTURAL REJECT that names the offending ext-tok -- never a reader-chosen
// default. Reserving a default here is exactly the ambiguity the closed enum
// exists to prevent: two readers must resolve an unknown construct identically.

// ErrExtDispositionOutOfRange is returned when ext-disposition holds a value
// outside the closed set {0x00, 0x01, 0x02}.
var ErrExtDispositionOutOfRange = errors.New("registry: ext-disposition outside the closed set {ignore, degrade, refuse}")

// DispositionError is the structural rejection for an invalid ext-disposition.
// It names the ext-tok (per FR-014: the reject names the token, never a
// reader default) and carries the offending value.
type DispositionError struct {
	Tok   ExtToken
	Value uint8
}

func (e *DispositionError) Error() string {
	return fmt.Sprintf("registry: ext-disposition 0x%02x on ext-tok %x is outside the closed set {0x00,0x01,0x02} (%v)",
		e.Value, e.Tok, ErrExtDispositionOutOfRange)
}

func (e *DispositionError) Unwrap() error { return ErrExtDispositionOutOfRange }

// ValidateDisposition returns a *DispositionError naming e.Tok if e.Disposition
// is out of range, or nil if it is one of the three legal values. A missing
// disposition is already rejected earlier, at decode, as a missing required
// field (ErrExtEnvelopeMissingField); this check adds the value range test the
// decoder deliberately leaves open so the offending token can be named here.
func ValidateDisposition(e ExtEnvelope) error {
	if e.Disposition.IsValid() {
		return nil
	}
	return &DispositionError{Tok: e.Tok, Value: uint8(e.Disposition)}
}
