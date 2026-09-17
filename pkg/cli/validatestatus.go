package cli

import "Protodoc/pkg/validate"

// validate->cli status bridge (T-0125, CON-011). The structural validation
// pipeline (pkg/validate) produces a Result carrying, independently, a
// validity verdict (Validity, the earliest failing structural check) and a
// step-9 resource-budget finding (Budget, a PD-BUDGET-xxx status). CON-011
// requires the over-budget status be "distinct from every validity verdict":
// a document valid at the format ceilings can still exceed THIS reader's own
// budget, and reporting that as INVALID would apply a stricter validity limit
// than the format defines. This bridge keeps the two orthogonal -- an
// over-budget document that is ALSO structurally invalid contributes BOTH
// active statuses -- and defers the single process exit code to Resolve,
// which applies cli.md S1.1's ranking (INVALID > OVER_BUDGET > ...). Neither
// finding is ever suppressed from the stdout envelope.

// StatusesFromValidation maps a validate.Result to the set of cli Statuses it
// makes true. A failing validity check sets StatusInvalid; a step-9 budget
// finding sets StatusOverBudget; the two are set independently so a document
// that is both invalid and over budget yields both. Resolve then selects the
// single highest-ranked status for the process exit code, but callers must
// still emit every finding (validity and budget) in the envelope.
func StatusesFromValidation(res validate.Result) map[Status]bool {
	active := map[Status]bool{}
	if res.Validity != nil {
		active[StatusInvalid] = true
	}
	if res.Budget != nil {
		active[StatusOverBudget] = true
	}
	return active
}

// FindingsFromValidation flattens a validate.Result's validity and budget
// findings into the cli envelope's Finding list, preserving BOTH regardless
// of which status Resolve picks for the exit code (cli.md S1.1: "findings
// always carries every applicable finding regardless of rank"). Budget
// findings are never dropped in favour of a higher-ranked validity verdict.
func FindingsFromValidation(res validate.Result) []Finding {
	var out []Finding
	add := func(f *validate.Finding) {
		if f == nil {
			return
		}
		cf := Finding{RuleID: f.RuleID, Message: f.Message}
		if f.Diag != nil {
			off := f.Diag.Offset
			cf.Offset = &off
		}
		out = append(out, cf)
	}
	add(res.Validity)
	add(res.Budget)
	return out
}
