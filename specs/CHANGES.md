# Protodoc spec — governance & release-gate change log

This file records structured, dated governance decisions that the frozen `spec.md` cannot carry
(spec.md is frozen; these are `plan.md`/release-gate decisions that reference it). Each entry names its
decision-maker honestly. Entries requiring **Eyvar García's** explicit ruling are recorded as **OPEN**
until that ruling genuinely occurs — no decision-maker is fabricated.

## Status of the M19 release-gate governance items

The following M19 tasks each require an action that CANNOT be performed by the implementing agent and
have therefore NOT been marked complete. They are recorded here honestly as OPEN, with what is blocked
and what would unblock each. This mirrors the project's standing precedent (T-0267 recorded a design
ruling OPEN rather than fabricate an Eyvar sign-off; T-0299 recorded the two-implementation trial OPEN;
T-0306 recorded an author-conducted review without fabricating an external audit).

### T-0347 — NFR-027 rendering schedule-risk ruling (OPEN, awaiting Eyvar García)

- **Status:** OPEN. plan.md Section 9 Conflict 4 flags NFR-027's 30-day rendering budget as at-risk
  (revised estimate 28–49 days). Resolving it requires **Eyvar García's** explicit ruling (accept the
  revised bound / descope / accept the miss).
- **Blocked on:** an explicit decision by Eyvar García. The task's test asserts
  `decision-maker == 'Eyvar García'`; that name must not be written on a decision the author invented.
- **Unblocks when:** Eyvar records the ruling (date, decision-maker, decision, rationale) here and in
  plan.md Section 9.

### T-0349 — NFR-028 / CQ-012 two-implementation funding decision (OPEN, awaiting Eyvar García)

- **Status:** OPEN. plan.md Section 8 discloses (high severity) that the two-independent-implementation
  gate has no funding/sponsor/recruitment strategy on record and asks Eyvar to decide.
- **Blocked on:** an explicit funding/sponsor decision by Eyvar García.
- **Unblocks when:** Eyvar records the decision (date, decision-maker, path-forward,
  fallback-if-unfunded) here and in plan.md Section 8.

### T-0356 — CON-026 licence/governance gate approval (OPEN, awaiting Eyvar García)

- **Status:** OPEN. The draft `LICENSE` and `GOVERNANCE.md` (T-0355) exist, but CON-026's "on record
  before v1" bar requires **Eyvar García's** explicit approval.
- **Blocked on:** Eyvar's approval (or requested changes) of LICENSE + GOVERNANCE.md.
- **Unblocks when:** Eyvar records the approval (date, decision-maker, decision=approved,
  referenced-files=[LICENSE, GOVERNANCE.md]) here.

### T-0365 — NFR-026 extracting-and-validating trial-miss ruling (OPEN or N/A, awaiting inputs)

- **Status:** OPEN pending the T-0345 trial outcome. If the T-0345 extracting-and-validating reader trial
  passed within its 5-working-day budget, this task is NOT APPLICABLE (to be logged as such once the
  trial result is confirmed). If it missed budget or failed the corpus, an explicit ruling by
  **Eyvar García** is required and is recorded OPEN here.
- **Blocked on:** the confirmed T-0345 result, then (if a miss) Eyvar's ruling.

### T-0346 — NFR-027 external rendering trial (OPEN, awaiting an outside implementer)

- **Status:** OPEN. Requires an **outside implementer** to build a passing rendering reader from the
  spec text alone within 30 working days, with actual elapsed days recorded. No such external trial has
  been run; its results must not be fabricated.
- **Unblocks when:** an outside implementer runs the trial and a dated report records elapsed days and
  the pass/fail result against the rendering conformance corpus.

### T-0351 — NFR-028 second-implementation differential trial (OPEN, awaiting a second implementer)

- **Status:** OPEN. Requires a genuinely **second, independently-authored implementation** of
  container/validate/canon (blocked on the T-0349 funding decision), run through the T-0350 differential
  harness against the full corpus. No second implementation exists; its results must not be fabricated.
- **Unblocks when:** the second implementation is commissioned and passes (or has every mismatch triaged)
  through the T-0350 harness.

## Honesty note

No entry above names Eyvar García as a decision-maker, because none of these decisions has actually been
made by him. Recording them as OPEN is the correct, honest state; the corresponding tests will remain
red-by-design until the real ruling/trial occurs, which is the release gate doing its job rather than a
defect to paper over.
