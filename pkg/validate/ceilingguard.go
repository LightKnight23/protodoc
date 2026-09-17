// Generated-ceiling pre-allocation guard (T-0115, FR-106): before allocating
// any buffer sized from a declared count (frame_count, segment_count, run
// count, signature count, ...), the declared value is checked against M01's
// CON-009 generated ceiling constants using overflow-safe arithmetic. On a
// violation the guard aborts and names the ceiling id plus the observed
// value BEFORE any proportional allocation occurs -- CP-006/CP-007's
// "verify before you allocate from a declared length" applied to counts.
package validate

import (
	"errors"
	"fmt"

	"Protodoc/pkg/ceilings"
)

// ErrCeilingExceeded is returned by GuardCount when a declared count exceeds
// its named ceiling. It is a structural resource-budget rejection produced
// before allocation, naming the ceiling and the observed value.
var ErrCeilingExceeded = errors.New("validate: declared count exceeds its CON-009 ceiling")

// CeilingError names the ceiling and the observed over-limit value.
type CeilingError struct {
	CeilingName string
	Ceiling     uint64
	Observed    uint64
}

func (e *CeilingError) Error() string {
	return fmt.Sprintf("validate: declared count %d exceeds ceiling %q (%d)", e.Observed, e.CeilingName, e.Ceiling)
}

func (e *CeilingError) Unwrap() error { return ErrCeilingExceeded }

// GuardCount checks a declared count against the named CON-009 ceiling and
// returns a *CeilingError if it exceeds it, WITHOUT performing any
// allocation. The comparison is a direct integer compare (no multiplication,
// so no overflow risk on the count itself); callers that then multiply the
// count by an element size should call GuardAllocation for the overflow-safe
// product check. ceilingName must be an exact ceilings.Table entry name.
func GuardCount(ceilingName string, declared uint64) error {
	max := ceilings.MustMax(ceilingName)
	if declared > max {
		return &CeilingError{CeilingName: ceilingName, Ceiling: max, Observed: declared}
	}
	return nil
}

// GuardAllocation checks that allocating declared elements of elemSize octets
// each is both within the named count ceiling AND does not overflow uint64
// when multiplied, returning a *CeilingError (or an overflow error) before
// any allocation. It is the overflow-safe form for "allocate count*elemSize
// octets" call sites.
func GuardAllocation(ceilingName string, declared, elemSize uint64) error {
	if err := GuardCount(ceilingName, declared); err != nil {
		return err
	}
	if elemSize != 0 && declared > (^uint64(0))/elemSize {
		return fmt.Errorf("validate: declared count %d * element size %d overflows uint64 (ceiling %q)", declared, elemSize, ceilingName)
	}
	return nil
}
