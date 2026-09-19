package cli

import (
	"errors"
	"fmt"

	"Protodoc/pkg/validate"
)

// ErrFileUnreadable is the sentinel every real backend wraps its os.Open/
// os.Stat failure in (DEFECT-2026-09-19b/T-0390). cli.md S1's own USAGE
// definition explicitly lists "an unreadable path" as a USAGE case (exit 7)
// -- distinct from INVALID, reserved for a file that opens fine but is
// structurally malformed. inspect.go has always drawn this line correctly;
// wrapping every other verb's open-failure in this sentinel, and having each
// dispatcher check errors.Is against it, makes every verb match inspect's
// behaviour instead of variously reporting INVALID or UNVERIFIED for a path
// that was never readable in the first place.
var ErrFileUnreadable = errors.New("protodoc: file could not be opened or read")

// unreadableFileRuleID is the RuleID realValidateStepsFor (which reports
// through validate.Step/Finding, not a plain error) uses for its own
// open-failure step, so runValidate can distinguish it from a genuine
// structural-decode failure the same way the error-based backends do via
// ErrFileUnreadable.
const unreadableFileRuleID = "TR-012-UNREADABLE"

// cp006Precondition runs every real backend's shared CP-006 validate-first
// precondition and returns a single classified error: nil if the document
// validates, an ErrFileUnreadable-wrapped error if the file could not be
// opened at all, or a generic structural-invalidity error otherwise
// (DEFECT-2026-09-19b/T-0390 -- this is what lets every backend's dispatcher
// tell a USAGE condition apart from a genuine INVALID one via one shared
// errors.Is check, instead of each duplicating the classification).
func cp006Precondition(path string) error {
	steps, _ := realValidateStepsFor(path)
	res := validate.Run(steps)
	if res.Validity == nil {
		return nil
	}
	if res.Validity.RuleID == unreadableFileRuleID {
		// res.Validity.Message already carries readBoundedPrefix's own
		// ErrFileUnreadable-prefixed text (validate_real.go); returning a
		// fresh %w-wrap of just the sentinel (rather than re-embedding the
		// message, which would duplicate that prefix) keeps errors.Is
		// working for every dispatcher's classification check.
		return fmt.Errorf("%w", ErrFileUnreadable)
	}
	return fmt.Errorf("document failed structural validation (CP-006): %s", res.Validity.Message)
}
