package cli

import "io"

// verify verb wiring (T-0331, TR-012). `protodoc verify` maps M09's per-state
// verdict machinery onto the cli exit codes and stdout payload, enforcing that
// a signer identity is displayed ONLY alongside a Valid verdict (cli.md
// non-negotiable #4 / FR-065).

// VerifyVerdict is the verify backend's per-signature outcome: the verdict name
// and, only when Valid, the signer identity.
type VerifyVerdict struct {
	Verdict string
	Signer  string // non-empty ONLY when Verdict == "valid"
}

// VerifyRun is the injectable verify backend. Production wiring replaces it with
// M09's Verify()/Verdict machinery and M10's LTV path.
var VerifyRun = func(path string) []VerifyVerdict { return nil }

// verifyStatus maps an M09 verdict name to the cli Status.
func verifyStatus(verdict string) Status {
	switch verdict {
	case "valid", "attested_with_declared_omissions":
		return StatusOK
	case "unavailable_state", "covering_unavailable_state":
		return StatusUnavailable
	default: // unverified, unattested
		return StatusUnverified
	}
}

func runVerify(args []string, _ io.Writer) Result {
	if len(args) < 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "verify requires a <file> argument"}},
		})
	}
	verdicts := VerifyRun(args[0])

	active := map[Status]bool{}
	sigs := make([]map[string]any, 0, len(verdicts))
	for _, v := range verdicts {
		active[verifyStatus(v.Verdict)] = true
		entry := map[string]any{"verdict": v.Verdict}
		// Signer identity ONLY for Valid (never on a non-Valid verdict).
		if v.Verdict == "valid" && v.Signer != "" {
			entry["signer"] = v.Signer
		}
		sigs = append(sigs, entry)
	}
	status := Resolve(active)
	return status.ToResult(Result{Extra: map[string]any{"signatures": sigs}})
}
