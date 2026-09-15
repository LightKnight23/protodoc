package cli

// Status is one of contracts/cli.md S1's exactly 8 documented process-exit
// categories. Its zero value is StatusOK.
type Status int

const (
	StatusOK Status = iota
	StatusInvalid
	StatusUnsupported
	StatusUnverified
	StatusUnavailable
	StatusOverBudget
	StatusRefused
	StatusUsage
)

// allStatuses lists every documented Status exactly once, in cli.md S1's
// table order (the process exit-code integer order), independent of
// S1.1's precedence order.
var allStatuses = [...]Status{
	StatusOK, StatusInvalid, StatusUnsupported, StatusUnverified,
	StatusUnavailable, StatusOverBudget, StatusRefused, StatusUsage,
}

// statusExitCode is cli.md S1's process exit-code integer for each Status.
var statusExitCode = map[Status]int{
	StatusOK:          exitOK,
	StatusInvalid:     1,
	StatusUnsupported: 2,
	StatusUnverified:  3,
	StatusUnavailable: 4,
	StatusOverBudget:  5,
	StatusRefused:     6,
	StatusUsage:       exitUsage,
}

// statusDisplayName is the stdout envelope's "status" field spelling for
// each Status (cli.md S1's Name column).
var statusDisplayName = map[Status]string{
	StatusOK:          statusOK,
	StatusInvalid:     "INVALID",
	StatusUnsupported: "UNSUPPORTED",
	StatusUnverified:  "UNVERIFIED",
	StatusUnavailable: "UNAVAILABLE",
	StatusOverBudget:  "OVER_BUDGET",
	StatusRefused:     "REFUSED",
	StatusUsage:       statusUsage,
}

// Code returns the process exit-code integer for s (cli.md S1).
func (s Status) Code() int { return statusExitCode[s] }

// String returns the stdout envelope's "status" field spelling for s
// (cli.md S1).
func (s Status) String() string { return statusDisplayName[s] }

// precedenceOrder is cli.md S1.1's ranking, highest to lowest:
// USAGE > INVALID > UNSUPPORTED > OVER_BUDGET > UNAVAILABLE > UNVERIFIED > REFUSED > OK.
var precedenceOrder = [...]Status{
	StatusUsage, StatusInvalid, StatusUnsupported, StatusOverBudget,
	StatusUnavailable, StatusUnverified, StatusRefused, StatusOK,
}

// Resolve implements cli.md S1.1: given the set of Statuses simultaneously
// true for one invocation, it returns the single highest-ranked one, which
// is the process's exit code. The caller is still responsible for
// reporting every applicable finding in the stdout envelope regardless of
// which single Status this resolves to (S1.1: "findings always carries
// every applicable finding regardless of rank"). An empty set resolves to
// StatusOK, since establishing no failing condition at all is OK by
// definition.
func Resolve(active map[Status]bool) Status {
	for _, s := range precedenceOrder {
		if active[s] {
			return s
		}
	}
	return StatusOK
}

// ToResult sets r's Status and ExitCode fields from s, leaving Findings
// and Extra untouched. Verb RunFuncs build a Result's Findings/Extra and
// call this once they know which Status (from Resolve, or directly) the
// invocation resolves to, so the exit code integer is never hand-typed
// at a verb call site.
func (s Status) ToResult(r Result) Result {
	r.Status = s.String()
	r.ExitCode = s.Code()
	return r
}
