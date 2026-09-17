// Package validate implements Protodoc's structural validation pipeline: the
// 13 ordered checks of data-model.md S7, run in a fixed order with strict
// error precedence (CP-006: verification precedes decoding). A document
// failing an earlier check is rejected on that check's verdict and later
// checks do not run; no partial presentation output and no heuristic
// reconstruction is produced past the first structural failure (FR-103).
//
// This file is the pipeline orchestrator (T-0113). Individual step
// implementations (diagnostics, ceilings, referential integrity, cycle
// detection, tree/signature recomputation, NFC, registry, extension
// envelope) are layered on by later M07 tasks; the orchestrator invokes them
// in the documented order and enforces the precedence rule.
package validate

// StepID identifies one of the 13 ordered validation checks
// (data-model.md S7). Values are the 1-based step numbers, so ordering by
// StepID is exactly the documented precedence order.
type StepID int

const (
	StepMagicHeader        StepID = 1  // magic + frozen header window
	StepCapabilityArith    StepID = 2  // capability_required <= capability_written
	StepRingWinner         StepID = 3  // commit-ring winner selection (PD-RING-001)
	StepTruncationRollback StepID = 4  // ledger_length <= file length
	StepBoundedPrefix      StepID = 5  // SegmentTableSlot bounds-safe arithmetic
	StepTSRecompute        StepID = 6  // T_S recomputation vs ledger_root
	StepSegTypeCoverage    StepID = 7  // segment-type enum + coverage well-formedness
	StepCycleDetection     StepID = 8  // 5-edge-kind cycle detection
	StepStructuralCeilings StepID = 9  // structural ceilings (PD-BUDGET, non-validity)
	StepTCAndSignature     StepID = 10 // T_C + structure_digest + signature verify
	StepNFCAndIdentity     StepID = 11 // NFC quick-check + A-FIELD-ROLE
	StepRegistryExcerpt    StepID = 12 // durable-claim RegistryExcerpt completeness
	StepExtEnvelope        StepID = 13 // extension-envelope reachability + disposition
)

// Finding is one validation result: the step that produced it, whether it is
// a validity failure (rejects the document) versus a non-validity budget
// status (step 9's PD-BUDGET codes, reported alongside a validity verdict),
// the violated rule id, and a human diagnostic. A nil *Finding from a step
// means the step passed.
type Finding struct {
	Step    StepID
	RuleID  string
	Message string
	// Budget marks a step-9 PD-BUDGET-xxx status: a non-validity resource
	// budget finding (CP-007) reported ALONGSIDE any validity verdict, never
	// instead of one.
	Budget bool
	// Diag carries the offset/unit/rule diagnostic for a structural failure
	// (FR-102); nil for findings that do not localise to an octet offset.
	Diag *Diagnostic
}

// Step is one pipeline check. It returns a non-nil *Finding on failure (or a
// budget status), or nil on pass. Steps must not produce presentation output
// or attempt reconstruction; they only inspect and report.
type Step struct {
	ID  StepID
	Run func() *Finding
}

// Result is the pipeline outcome: the validity Finding (the earliest-ordered
// failing validity check, or nil if every validity check passed) and,
// separately, any step-9 budget Finding (reported alongside, per the S7
// error-precedence exception).
type Result struct {
	// Validity is the earliest-ordered failing validity check (steps other
	// than a pure step-9 budget), or nil if all validity checks passed.
	Validity *Finding
	// Budget is a step-9 PD-BUDGET status if one was produced, independent
	// of Validity.
	Budget *Finding
	// Valid reports whether no validity check failed (Budget does not affect
	// this: a budget-exceeded document can still be structurally valid).
	Valid bool
}

// Run executes the steps in ascending StepID order and applies the S7 error-
// precedence rule: the earliest-numbered failing VALIDITY check determines
// the reported validity verdict and NO later check is evaluated once a
// validity failure is found (no partial output, no reconstruction, FR-103).
// Step 9's PD-BUDGET status is the one documented exception: a budget finding
// is captured and reported alongside the validity verdict rather than
// halting the pipeline, but only budget findings have that exemption -- a
// validity failure at any step still halts.
//
// Steps are sorted by ID defensively so a caller supplying them out of order
// cannot subvert the precedence guarantee.
func Run(steps []Step) Result {
	ordered := append([]Step(nil), steps...)
	sortByID(ordered)

	var res Result
	res.Valid = true
	for _, s := range ordered {
		f := s.Run()
		if f == nil {
			continue
		}
		if f.Budget {
			// Non-validity budget status: record it once and keep going;
			// it never halts the pipeline and never flips Valid.
			if res.Budget == nil {
				res.Budget = f
			}
			continue
		}
		// A validity failure: this is the reported verdict; halt. No later
		// check runs, no partial output is produced.
		res.Validity = f
		res.Valid = false
		return res
	}
	return res
}

// sortByID sorts steps ascending by StepID (insertion sort; the slice is
// always tiny -- at most 13 entries).
func sortByID(steps []Step) {
	for i := 1; i < len(steps); i++ {
		for j := i; j > 0 && steps[j-1].ID > steps[j].ID; j-- {
			steps[j-1], steps[j] = steps[j], steps[j-1]
		}
	}
}
