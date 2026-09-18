// Conditional write (TR-010; T-0341). Every mutating verb (sign, redact,
// publish, merge, migrate) accepts --if-match=<expected-state-id>. Before
// committing a write, the verb checks the target's CURRENT state id against the
// expected one via an ETag-capable backend; on mismatch it REFUSES, naming the
// state id it actually found, so a concurrent write cannot silently clobber.
package cli

import "fmt"

// StateIDReader returns the current state id of the target the writer is about
// to overwrite (an ETag-capable backend read). ok is false if the target does
// not exist yet (no conflict possible).
type StateIDReader func(target string) (currentStateID string, ok bool)

// IfMatchResult reports a conditional-write decision.
type IfMatchResult struct {
	// Proceed is true when the write may proceed (no --if-match given, or the
	// expected id matches the current one).
	Proceed bool
	// FoundStateID is the current state id found when a mismatch occurred.
	FoundStateID string
}

// CheckIfMatch implements TR-010's conditional write. If expected is empty
// (--if-match not supplied) the write proceeds unconditionally. Otherwise it
// reads the target's current state id and proceeds only if it equals expected;
// on mismatch it refuses, reporting the state id it found.
func CheckIfMatch(expected, target string, read StateIDReader) IfMatchResult {
	if expected == "" {
		return IfMatchResult{Proceed: true}
	}
	current, ok := read(target)
	if !ok {
		// Target absent: no concurrent state to conflict with.
		return IfMatchResult{Proceed: true}
	}
	if current != expected {
		return IfMatchResult{Proceed: false, FoundStateID: current}
	}
	return IfMatchResult{Proceed: true}
}

// IfMatchRefusal builds the REFUSED Result for a conditional-write mismatch,
// naming the state id actually found (the first writer's resulting state).
func IfMatchRefusal(found string) Result {
	return StatusRefused.ToResult(Result{
		Extra: map[string]any{"if_match_conflict": true, "found_state_id": found},
		Findings: []Finding{{
			RuleID:  "TR-010",
			Message: fmt.Sprintf("conditional write refused: target is now at state %s", found),
		}},
	})
}

// parseIfMatch extracts the --if-match=<id> value from args, if present.
func parseIfMatch(args []string) string {
	const pfx = "--if-match="
	for _, a := range args {
		if len(a) > len(pfx) && a[:len(pfx)] == pfx {
			return a[len(pfx):]
		}
	}
	return ""
}
