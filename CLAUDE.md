# CLAUDE.md — Protodoc

Project-local instructions. Read this before touching anything in this repo. It overrides nothing in
`~/.claude/CLAUDE.md` (global SDD workflow, standards, git author policy) — it exists to give the current state
of THIS project so any agent picking it up mid-stream does not have to re-derive it.

## What this project is

Protodoc is a **new document format, designed from scratch**, plus its reference tooling. It is not a wrapper
around an existing format and not compatible with OOXML/PDF by design — see
[`.specify/memory/constitution.md`](.specify/memory/constitution.md) § Purpose for why.

Origin: the user sketched a protobuf + zip-of-parts + offset-based-style-span design
(`/Users/Eyvar/Downloads/protodocs_WIP.md`, not in this repo). That sketch was treated as a **hypothesis**, not
a decision, and was pressure-tested during phase 3. Two of its three pillars were rejected on evidence:

- **Offset-based style spans**: rejected. A file at rest has no transform function to repair a stale offset;
  the first concurrent edit relocates every annotation silently. Replaced by identity-based anchoring
  (character identity, minted per run, never reissued — see `data-model.md` and CQ-001/CQ-002 in `clarify.md`).
- **zip-of-parts container**: rejected as framing. Repack-on-save forfeits the octet-stable-diff property that
  is the format's main differentiator against OOXML.
- **Protocol buffers**: not adopted. proto3 has no canonical serialization guaranteed across independent
  implementations, which breaks the "one state, one signable octet sequence" requirement (CQ-003/CQ-004).

The chosen architecture is **Protodoc Ledger (PDL)**: a fixed 1,048,576-octet prefix (header, self-digesting
commit ring, segment table) + an append-only sealed-segment ledger + two integrity trees (`T_S` storage,
`T_C` content-commitment, both feeding a single signed root) + Ed25519 signing (`EdDSA-Protodoc-1`, a 7-step
canonical/small-order-rejecting wrapper around the **unmodified** Go stdlib verifier — no curve-arithmetic
reimplementation) + a bespoke deterministic TLV encoding (`PDL-TLV`) for structured records. Full rationale,
the 5 competing architectures it beat, and every rejected alternative are in `research.md` and `plan.md`.

## Where things stand — READ THIS FIRST

This project follows **Spec-Driven Development (SDD)**, per the global CLAUDE.md. Spec is the source of truth;
code is its output.

| Phase | Artifact | Status |
|---|---|---|
| 0 Constitution | `.specify/memory/constitution.md` | ✅ ACTIVE, v0.2.0 (amended 2026-09-15, see CP-009), approved |
| 1 Specify | `specs/001-protodoc-format-core/spec.md` | ✅ APPROVED — 197 requirements, FR-061 amended 2026-09-15 |
| 2 Clarify | `specs/001-protodoc-format-core/clarify.md` | ✅ CLOSED — CQ-001..CQ-017 resolved (CQ-013..017 added at phase 5) |
| 3 Plan | `specs/001-protodoc-format-core/plan.md`, `research.md`, `data-model.md`, `contracts/` | ✅ APPROVED — architecture "Protodoc Ledger (PDL)" |
| 4 Tasks | `specs/001-protodoc-format-core/tasks.md` | ✅ APPROVED — 19 milestones, 372 tasks, 0 orphan requirements |
| 5 Analyze | `specs/001-protodoc-format-core/analysis.md` | ✅ APPROVED — gate PASSES |
| **6 Implement** | `pkg/`, `cmd/` | 🔶 **IN PROGRESS — 363/372 tasks done. Read the next section before touching code.** |
| 7 Verify | test reports | ⬜ not started (this is a separate gate from "tests pass in CI," see below) |
| 8 Converge / 9 Close | — | ⬜ |

Git: branch `001-protodoc-format-core` off `main`. Every task gets its own commit ending
`Refs: 001-protodoc-format-core/T-NNNN (implements ids)`. **This Refs trailer is the ONLY reliable source of
truth for "is task T-NNNN done" — never trust a prior session's or agent's self-reported completion list.**
Verify with:
```
git log --all --format="%H" | while read sha; do git show -s --format="%B" "$sha" | grep "^Refs:"; done \
  | grep -o "T-[0-9]\{4\}" | sort -u
```

## Phase 6 status (verified against git history, 2026-09-18) — READ BEFORE WRITING ANY CODE

**363 of 372 tasks are actually done.** All of them are traceable to a real commit with a `Refs:` trailer;
nothing here is inferred from a workflow's or agent's self-report. This work was carried from 47/372 to
363/372 by a different tool (Kiro) after the Claude Code handoff — verified fresh against `git log`, not
against any prior session's claim.

- **M01–M18: fully done, every task.** M01 (Core Encoding & Fixed Prefix, 33/33), M02 (Ledger & Placement,
  17/17), M04 (Identity & Anchor, 23/23), M05 (Extraction, 15/15), M06 (Signature Primitive, 11/11), M08
  (Integrity Trees T_S/T_C, 17/17), M07 (Structural Validation Core, 20/20), M03 (Extensibility Envelope,
  17/17), M09 (Signature & Coverage, 23/23), M10 (Attestation/LTV, 18/18), M11 (Redaction, 22/22), M12
  (History & Erasure, 18/18), M13 (Concurrent-Edit/Merge, 18/18), M14 (Rendering & Resource, 25/25), M15
  (Accessibility & Semantic Content-Model, 29/29), M16 (Migration & Registry, 11/11), M17 (Canonicalization,
  18/18), M18 (CLI Surface, 17/17). `go build ./... && go vet ./... && go test ./...` is green across every
  package: `pkg/pdlfmt`, `pkg/container`, `pkg/ledger`, `pkg/content`, `pkg/content/mint`, `pkg/eddsa`,
  `pkg/integrity`, `pkg/validate`, `pkg/extract`, `pkg/history`, `pkg/merge`, `pkg/render`, `pkg/semantics`,
  `pkg/migrate`, `pkg/registry`, `pkg/canon`, `pkg/cli`, `pkg/diffconform`, `pkg/fuzzmaturity`,
  `pkg/benchconfig`, `pkg/ceilings`, `pkg/governance`, `pkg/traceability`, plus `cmd/protodoc`,
  `cmd/protodoc-diffconform`, `cmd/protodoc-traceaudit`.
- **M19 (Conformance, Fuzzing & Governance Convergence): 11 of 20 tasks done.** 9 remain, and all 9 are
  real-world actions no agent can perform by writing code: T-0345/T-0346 (external-implementer trials, 5-day
  and 30-day budgets), T-0347/T-0356/T-0358/T-0365 (Eyvar's personal governance rulings/sign-offs), T-0349
  (second-implementation funding/sponsor decision), T-0351 (commissioning and running the actual second
  independent implementation conformance run), T-0371 (real IANA media-type/format-identification filing).
  See the honesty rule below — none of these should ever be faked to close them out.

**Lesson baked in from the earlier handoff, still load-bearing:** an earlier automated Claude Code batch run
over-reported completion by ~3x on M18/M19 before this jump (self-reported all tasks done, only a fraction had
real commits) — full forensic detail is in git history around 2026-09-15/16 if curious. **The rule stands for
whoever continues this: after ANY automated or agent-driven implementation pass, re-verify against `git log`'s
`Refs:` trailers before believing anything is done.** Regenerate the task/status picture with
`python3 scripts/extract_milestone_tasks.py` (writes to `.impl_tasks/`, gitignored, regenerated from
`tasks.md` — never hand-edit its output) rather than trusting any prior conversation's task list. Cross-check
with:
```
git log --all --format="%H" | while read sha; do git show -s --format="%B" "$sha" | grep "^Refs:"; done \
  | grep -o "T-[0-9]\{4\}" | sort -u
```

**M02–M18 were built by a different tool than the one that did M01** — they have not had a Claude-Code-side
review pass. Nothing here indicates a problem (build/vet/test all green), but if you're doing a Verify-phase
(phase 7) audit, do not assume clean CI means the code matches every contract/data-model detail; a real review
pass against `contracts/*.abnf` and `data-model.md` has not happened yet for M02–M18.

### Real build order (topological, NOT tasks.md §1's table row order)

```
M01 → M02 → M04 → M05 → M06 → M08 → M07 → M03 → M09 → M10 → M11 → M12 → M13 → M14 → M15 → M16 → M17 → M18 → M19
```

`tasks.md` §1's table lists milestones in a DIFFERENT order (M01, M02, M03, M04, ...) that is NOT a valid build
order — M03 depends on M07 (a phase-5 fix, see `analysis.md` GAP A), which the table lists after it. Regenerate
this order yourself with `scripts/extract_milestone_tasks.py` rather than reading the table's row order as a
sequence; the script's `_index.json` carries the real order under `topo_order`.

**Next tasks to implement: the 9 remaining M19 tasks, all real-world (not code) actions — see the list above.**
No code-writing tasks remain in M01–M18. Once M19's 9 items are resolved (or explicitly deferred by Eyvar),
the project moves to Phase 7 (Verify).

### How to continue: the discipline that must not slip

1. Regenerate `.impl_tasks/` (`python3 scripts/extract_milestone_tasks.py`) and re-verify done-task status
   against `git log`'s `Refs:` trailers (command above) before doing anything else. Do not trust this file's
   task counts without that verification if any time has passed or another agent may have worked since.
2. Work through milestones in the topological order above, tasks within a milestone in `tasks.md`'s listed
   order (that intra-milestone order is real dependency order).
3. Before writing anything: `cd /Users/Eyvar/GolandProjects/Protodoc && go build ./... && go vet ./... && go
   test ./...`. If this baseline is broken, fix that first — do not build on top of a broken baseline.
4. Read the relevant `contracts/*.abnf` section and `data-model.md` entity before implementing any wire
   struct — a task's one-paragraph description is a summary, not the spec.
5. Implement the task's code AND its exact named test (`test_name`/`test_kind` in tasks.md). Re-run build/vet/
   test; every existing test must stay green.
6. **Commit ONE task at a time**, never batched, the moment it's green: subject + body + blank line +
   `Refs: 001-protodoc-format-core/T-NNNN (implements ids)`. `git add` only the files that task touched — never
   `git add -A`/`git add .`. No `--author` needed; local git config in this repo is already
   `Eyvar García <eyvar.0823@gmail.com>`.
7. **Honesty rule, non-negotiable:** some tasks require a real external action no agent can perform by writing
   code — Eyvar's personal sign-off, a genuine third-party implementer's trial, a funding/sponsorship decision,
   an actual registry filing, commissioning a real second independent implementation. For these: do not write a
   test pretending the action happened, do not fabricate a decision document or sign-off record, do not commit
   anything for that task id. Leave it undone and say so plainly. A fabricated governance record in this
   codebase is exactly the kind of silent defect this project's whole process exists to prevent.
8. Constitution principles CP-001..CP-014 (`.specify/memory/constitution.md`) bind every line of code, most
   load-bearing: CP-010 (Go stdlib only in core paths — no third-party deps without a recorded exception),
   CP-004 (determinism: exactly 4 permitted non-deterministic sites, no others), CP-006 (verify before
   decoding, never allocate from an unvalidated declared length), CP-007 (every numeric ceiling is a named
   constant from `data-model.md`'s ceiling table, never invented ad hoc), CP-009 (no behavior defined by a
   named application, except the narrow v0.2.0 exception for a pinned external determinism oracle).

## Working conventions specific to this repo

- **Requirement IDs**: `FR-NNN` functional, `NFR-NNN` non-functional, `CON-NNN` constraint, `TR-NNN` tooling
  (a 4th class, deliberately added beyond the global scheme's usual 3 — see `clarify.md` AD-001 for why).
- **Task IDs**: `T-0NNN`, globally unique across all 19 milestones (zero-padded 4 digits). Every task's
  `implements` field cites the requirement id(s) it satisfies; every task has exactly one named primary test.
- **Constitution principles**: `CP-001`..`CP-014`, in `.specify/memory/constitution.md`. These are load-bearing
  — CP-010 (Go stdlib only in core paths), CP-004 (determinism, 4 permitted non-deterministic sites, no more),
  CP-006 (verify before decode), CP-003 (no stable release without 2 independent implementations) will shape
  every implementation task. Read the constitution before writing any code, not just the spec.
- **Clarification decisions**: `CQ-001`..`CQ-017` in `clarify.md`, all resolved (CQ-013..017 were added at
  phase 5 to close judgment-call findings; see `analysis.md` §9-10). Do not relitigate one; if you think a
  resolved decision is wrong, say so explicitly to Eyvar rather than silently designing around it — same rule
  the constitution's amendment process states for itself.
- **No em dashes** in any spec-phase artifact (house style carried through every generated document).
- **Git commit authorship**: sole author `Eyvar García <eyvar.0823@gmail.com>`, no `Co-Authored-By` line,
  per the user's global preference — this is stricter than the generic Claude Code default and takes
  precedence in this repo.
- **Reference implementation language**: Go 1.25, standard library only in parse/validate/canonicalise/
  verify/extract paths (CP-010). Cross-compilation targets: darwin, linux, windows on amd64 and arm64.

## How this was built (for context, not re-derivation)

Phases 0-5 were produced by multi-agent Workflow runs (research/synthesis for spec, competing-architecture
judging for plan, inventory/decompose/audit/repair for tasks, cross-artifact consistency checking for analyze),
each with adversarial verification built in (skeptic agents trying to refute findings, judges scoring 5
independent architecture proposals, mechanical traceability audits). Full methodology and every rejected
option's reasoning is in `research.md` — read it before proposing an alternative to something already decided;
the tradeoff was probably already weighed and rejected there, with evidence.

Phase 6 (implement) is a genuinely different kind of work from phases 0-5: those produced markdown, which is
mergeable and cheap to verify by re-reading; phase 6 produces Go code, which has real compile-time coupling and
cannot be safely authored in parallel across shared packages the way spec documents can. Whatever tool
continues this work should implement tasks strictly sequentially in the topological order above, verify with
a real `go build`/`go vet`/`go test`, and never trust a self-reported completion list over `git log`.

If you are an LLM/agent continuing this work and need current status without re-reading every artifact,
`README.md` in this repo root has the plain-English orientation; this file (§ "Phase 6 status" above) has the
current, verified implementation state and the operational rules for continuing it.
