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

### T-0349 — NFR-028 / CQ-012 two-implementation funding decision (RESOLVED)

- **Date:** 2026-09-18 (path-forward revised 2026-09-19)
- **Decision-maker:** Eyvar García
- **Path-forward:** open volunteer call, unfunded. No cash bounty; instead, a public GitHub issue and
  outreach (open-source/PL/crypto communities, CS department capstone-project channels) inviting a
  volunteer to build the second independent implementation of container/validate/canon, per the T-0348
  acceptance protocol. Compensation is non-monetary: named credit as CP-003's second implementer,
  permanently recorded in the project's governance record.
- **Fallback-if-unfunded:** if no volunteer emerges within a reasonable window, T-0351 remains deferred
  indefinitely; T-0358's v1-stable capstone records CP-003/NFR-028 as an open blocking gate rather than
  proceeding without it satisfied.
- Mirrored in plan.md Section 8's risk row for this item.

### plan.md Conflict 5 — TR-010 storage-backend abstraction adopted (RESOLVED)

- **Date:** 2026-09-19
- **Decision-maker:** Eyvar García
- **Decision:** adopted the already-implemented `ledger.ConditionalWriter` interface
  (`pkg/ledger/conditionalwriter.go`, T-0044/T-0045) as `plan.md`'s architecture-level answer to TR-010.
  The frozen plan as originally written described no storage-backend abstraction; M02 implemented one
  anyway and M18's CLI (T-0341) built on top of it before the plan document was updated to match. This
  decision closes that gap by formally recording the mechanism in `plan.md` Section 9 Conflict 5.
- **Rationale:** the code was already correct, tested, and in production use (`TestTR_010_
  ConditionalWriterInterfaceContract`, `TestTR_010_ConditionalWriteRefusalNamesCurrentHolder`,
  `TestTR_010_ConditionalWriteRefusesOnMismatch`); this is a documentation/architecture-record fix, not
  a design or wire-format change.
- Mirrored in `plan.md` Section 9 Conflict 5 and `docs/ledger-conditional-write-design-note.md`.

### T-0356 — CON-026 licence/governance gate approval (RESOLVED)

- **Date:** 2026-09-19
- **Decision-maker:** Eyvar García
- **Decision:** approved
- **Referenced-files:** LICENSE, LICENSE-CODE, GOVERNANCE.md
- All four of CON-026's required elements approved as drafted: the Licence Grant (LICENSE for the
  specification, LICENSE-CODE/Apache 2.0 for the reference implementation), Eyvar García as Named
  Steward, the Succession Process (180-day unresponsiveness trigger, 30-day comment period, majority
  vote if contested), and the Deprecation Window Policy (3-year minimum). `GOVERNANCE.md`'s status
  updated to APPROVED accordingly.

### T-0345 — NFR-026 external extract+validate trial (OPEN, recruiting)

- **Status:** OPEN. Requires an **outside implementer** to build a passing extracting-and-validating
  reader from the spec text alone within 5 working days, with actual elapsed days recorded. No such
  external trial has been run; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#2](https://github.com/LightKnight23/protodoc/issues/2).
- **Unblocks when:** an outside implementer runs the trial and a dated report records elapsed days and
  the pass/fail result against the extract+validate conformance corpus.

### T-0365 — NFR-026 extracting-and-validating trial-miss ruling (OPEN or N/A, awaiting inputs)

- **Status:** OPEN pending the T-0345 trial outcome. If the T-0345 extracting-and-validating reader trial
  passed within its 5-working-day budget, this task is NOT APPLICABLE (to be logged as such once the
  trial result is confirmed). If it missed budget or failed the corpus, an explicit ruling by
  **Eyvar García** is required and is recorded OPEN here.
- **Blocked on:** the confirmed T-0345 result, then (if a miss) Eyvar's ruling.

### T-0346 — NFR-027 external rendering trial (OPEN, recruiting)

- **Status:** OPEN. Requires an **outside implementer** to build a passing rendering reader from the
  spec text alone within 30 working days, with actual elapsed days recorded. No such external trial has
  been run; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#3](https://github.com/LightKnight23/protodoc/issues/3).
- **Unblocks when:** an outside implementer runs the trial and a dated report records elapsed days and
  the pass/fail result against the rendering conformance corpus.

### T-0351 — NFR-028 second-implementation differential trial (OPEN, recruiting)

- **Status:** OPEN. Requires a genuinely **second, independently-authored implementation** of
  container/validate/canon, run through the T-0350 differential harness against the full corpus. No
  second implementation exists; its results must not be fabricated.
- **Recruitment:** volunteer call posted 2026-09-19 as
  [LightKnight23/protodoc#1](https://github.com/LightKnight23/protodoc/issues/1), per the T-0349
  unfunded-open-volunteer-call decision.
- **Unblocks when:** a volunteer is found, the implementation is built, and it passes (or has every
  mismatch triaged) through the T-0350 harness.

### T-0371 — CP-014 IANA media-type / PRONOM format registration (IN PROGRESS)

- **Status:** IN PROGRESS. Repository made public at
  https://github.com/LightKnight23/protodoc (default branch `001-protodoc-format-core`), with GitHub
  Pages enabled and verified serving `spec.md` at a stable public URL, so both filings below have a
  real "published specification" link to cite.
- **IANA media-type registration:** submitted 2026-09-19 by Eyvar García to media-types@iana.org
  (subtype `application/vnd.protodoc`, vendor tree per RFC 6838 Section 5.3.4). Awaiting IANA Designated
  Expert review; no tracking ID assigned yet by IANA as of this entry.
- **PRONOM format registration:** submitted 2026-09-19 by Eyvar García to PRONOM@nationalarchives.gov.uk,
  via the official PRONOM Submission template (docx), with a real, decoder-valid, minimal `protodoc-sample.pdl`
  attached for signature testing. Signature independently verified beforehand with The National Archives'
  own DROID tooling: `sigtool` byte-match (1/1 hit on the sample, 0/1 on a control file) and a full DROID
  identification run using a locally patched copy of the real DROID_SignatureFile_V119.xml (correctly
  identified the sample, correctly left the control file unidentified as Protodoc). Awaiting PRONOM team
  review; no PUID assigned yet as of this entry.
- **Unblocks when:** IANA approves and publishes the registration (tracking: the published registry
  entry) AND PRONOM assigns a PUID. Both dates/reference IDs get recorded here, and
  TestFR_125_MediaTypeRegistrationOnRecord is written and made to pass, before this task closes.

## DEFECT-2026-09-19 — CLI verbs never wired to real decode chain (found post-M18)

**Severity:** high (correctness). **Found:** 2026-09-19, external review. **Status:** IN PROGRESS.

M18 (CLI Surface) was marked 17/17 with all tests green, but the compiled `protodoc` binary does not
process real files for 10 of its 11 verbs. Reproduction (current tree, before fix):

```
$ go build -o /tmp/protodoc ./cmd/protodoc
$ /tmp/protodoc validate /tmp/this-file-does-not-exist-at-all.pdl
{"checks":0,"exit_code":0,"findings":[],"status":"OK","verb":"validate"}
```

A file that does not exist reports OK. **Root cause:** every verb except `inspect` routes through a
package-level "injectable backend" variable (`ValidateStepsFor`, `ProjectStateFor`, `ExtractRun`,
`VerifyRun`, `DiffRun`, `MergeRun`, `RedactRun`, `PublishRun`, `SignRun`, `MigrateRun`) that DEFAULTS to
a no-op stub. The stubs are legitimate test-doubles, but the production wiring that replaces them with
the real M01–M17 decode chain was never written, and `cmd/protodoc/main.go` imports only `pkg/cli`.
`go test ./...` stayed green because the verb tests exercise the stubs directly, never real file I/O.
Only `pkg/cli/inspect.go` genuinely opens the file and decodes real bytes.

This defect is NOT covered by any T-NNN task in the original spine (all 17 M18 tasks tested the stub
seam, not the production seam). It is fixed under new numbered tasks **T-0373..T-0382** (one per stub
verb) added to `tasks.md`'s M18 section, one verb per commit, each wiring the stub to the real decode
chain (following `inspect.go`'s pattern) plus a real test proving it rejects a garbage/truncated/absent
file and accepts a genuine sample. `validate` is done first because CP-006 requires every other verb to
run validation as a precondition. No heuristic recovery, no partial output past the first structural
failure (CP-006/FR-103).

## Honesty note

T-0349, T-0356, and plan.md's Conflict 5 above name Eyvar García as decision-maker because those
decisions were actually made by him, on the dates given. Every remaining OPEN entry names no
decision-maker, because none of those decisions has actually been made yet. Recording them as OPEN is
the correct, honest state; the corresponding tests will remain red-by-design until the real
ruling/trial occurs, which is the release gate doing its job rather than a defect to paper over.
