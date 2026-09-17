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
| **6 Implement** | `pkg/`, `cmd/` | 🔶 **IN PROGRESS — 47/372 tasks done. Read the next section before touching code.** |
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

## Phase 6 status (verified against git history, 2026-09-16) — READ BEFORE WRITING ANY CODE

**47 of 372 tasks are actually done.** All of them are traceable to a real commit with a `Refs:` trailer;
nothing here is inferred from a workflow's self-report (see the cautionary tale two paragraphs down).

- **M01 (Core Encoding & Fixed Prefix): all 33 tasks done** (T-0001..T-0032, T-0359). `go build ./... && go vet
  ./... && go test ./...` is green. Packages: `pkg/pdlfmt` (PDL-VARINT, PDL-TLV, digest256/nfc-string/unit-id/
  u48 value-kind helpers), `pkg/container` (Header, CommitRingRecord, Frontmatter, SegmentTableSlot, plus every
  M01 audit/conformance task).
- **M18 (CLI Surface): only 4 of 17 tasks done** (T-0325, T-0326, T-0329, T-0339 — dispatch framework, exit-code
  engine, `inspect` verb, requirement-to-verb traceability table). 13 tasks remain: T-0327, T-0328, T-0330..338,
  T-0340, T-0341.
- **M19 (Conformance, Fuzzing & Governance): only 10 of 20 tasks done** (T-0342, T-0343, T-0348, T-0350, T-0352,
  T-0353, T-0354, T-0355, T-0357, T-0372). 10 remain: T-0344, T-0345, T-0346, T-0347, T-0349, T-0351, T-0356,
  T-0358, T-0365, T-0371 — several of which (T-0345/T-0346 external-implementer trials, T-0349 funding decision,
  T-0351 commissioning a second implementation, T-0356/T-0358 Eyvar sign-offs, T-0371 an actual IANA filing)
  genuinely cannot be completed by an agent writing code; see the honesty rule below.
- **M02–M17 (everything else — Ledger, Extensibility Envelope, Identity & Anchor, Extraction, Signature
  Primitive, Structural Validation, Integrity Trees, Signature & Coverage, Attestation/LTV, Redaction,
  History & Erasure, Concurrent-Edit/Merge, Rendering, Accessibility, Migration/Registry, Canonicalization):
  ZERO tasks done.** No packages exist for identity/anchoring, Ed25519 signing, the T_S/T_C integrity trees,
  redaction, merge, or rendering. This is the actual reference-implementation work; almost none of it exists.

**Why M18/M19 have any commits at all despite M02–M17 being empty, and a cautionary tale about trusting
self-reports:** an automated batch-implementation run processed milestones in real topological order (M01 →
M02 → M04 → ... → M17 → M18 → M19 — see "Real build order" below, NOT `tasks.md` §1's table row order) via
many sequential subagent calls. A usage-limit error hit partway through M02 and, due to a bug in the
orchestrating script (since fixed — see `git log` history around 2026-09-15/16 if curious), the run did not
stop: it kept advancing through the batch list, so every M02–M16 and early-M17 batch failed identically on the
same limit, while enough real time passed mid-run for the limit to reset before the run reached late-M17/M18/
M19, which then genuinely succeeded. **Separately, and worse: the M18/M19 batches that DID run self-reported
completing all their assigned tasks, but only a fraction actually produced a real commit with a `Refs:`
trailer** (4/17 and 10/20 respectively) — the self-report cannot be trusted at face value even for batches that
"succeeded." **Lesson for whoever continues this: after ANY automated or agent-driven implementation pass,
re-verify against `git log`'s `Refs:` trailers before believing anything is done.** Regenerate the task/status
picture with `python3 scripts/extract_milestone_tasks.py` (writes to `.impl_tasks/`, gitignored, regenerated
from `tasks.md` — never hand-edit its output) rather than trusting any prior conversation's task list.

**M18's and M19's 14 completed tasks were built without their real prerequisites** (M02–M17 don't exist).
They almost certainly need a review/rework pass once the real dependencies land — e.g. the CLI's `inspect` verb
and the traceability table can't yet do anything with identity, signatures, or redaction, because nothing to
inspect exists. Do not assume M18/M19's existing code is correct or complete just because it builds and its own
tests pass; its tests can only exercise what's actually there yet.

### Real build order (topological, NOT tasks.md §1's table row order)

```
M01 → M02 → M04 → M05 → M06 → M08 → M07 → M03 → M09 → M10 → M11 → M12 → M13 → M14 → M15 → M16 → M17 → M18 → M19
```

`tasks.md` §1's table lists milestones in a DIFFERENT order (M01, M02, M03, M04, ...) that is NOT a valid build
order — M03 depends on M07 (a phase-5 fix, see `analysis.md` GAP A), which the table lists after it. Regenerate
this order yourself with `scripts/extract_milestone_tasks.py` rather than reading the table's row order as a
sequence; the script's `_index.json` carries the real order under `topo_order`.

**Next task to implement, in order: T-0033 (first task of M02, Ledger & Placement).** Read its full detail in
`tasks.md` §3 (`### M02: Ledger & Placement`) or in `.impl_tasks/M02.json` after regenerating.

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
