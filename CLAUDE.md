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
code is its output. **No code exists in this repo yet, and that is correct — do not write implementation code
until phase 5 (analyze) has produced and you have read `specs/001-protodoc-format-core/analysis.md`, and Eyvar
has approved it.**

| Phase | Artifact | Status |
|---|---|---|
| 0 Constitution | `.specify/memory/constitution.md` | ✅ ACTIVE, v0.1.0, approved |
| 1 Specify | `specs/001-protodoc-format-core/spec.md` | ✅ APPROVED — 197 requirements (125 FR, 34 NFR, 26 CON, 12 TR) |
| 2 Clarify | `specs/001-protodoc-format-core/clarify.md` | ✅ CLOSED — 12/12 resolved, 2 accepted deviations (AD-001, AD-002) |
| 3 Plan | `specs/001-protodoc-format-core/plan.md`, `research.md`, `data-model.md`, `contracts/` | ✅ APPROVED — architecture "Protodoc Ledger (PDL)" |
| 4 Tasks | `specs/001-protodoc-format-core/tasks.md` | ✅ APPROVED — 19 milestones, 365 tasks, 0 orphan requirements |
| **5 Analyze** | `specs/001-protodoc-format-core/analysis.md` | ⬜ **NOT STARTED — do this next** |
| 6 Implement | `internal/`, `pkg/`, `cmd/protodoc/` (do not exist yet) | ⬜ blocked on phase 5 approval |
| 7 Verify | test reports | ⬜ |
| 8 Converge / 9 Close | — | ⬜ |

Git: branch `001-protodoc-format-core` off `main`. `main` holds only the baseline (`go.mod`, `.gitignore`).
Every phase artifact is its own commit; approval of a DRAFT status line is its own follow-up commit. Commit
history under `git log --oneline` is a readable phase-by-phase record — read it before asking "what happened."

## Known, disclosed gaps in the current artifacts

Two real issues are already flagged in `tasks.md` § 5 (Residual gaps) rather than hidden. **A phase-5 analyze
pass should either resolve these or explicitly carry them forward — do not silently patch around them:**

1. **Spine/task mismatch**: `tasks.md`'s own milestone table (§1) lists M03 depending only on M01, M02 — but
   two of M03's tasks (T-0056, T-0064) were patched to depend on M07's T-0118 during phase-4 repair. The
   milestone spine needs correcting to `M03: M01, M02, M07` before task scheduling trusts it.
2. **Stale sibling references in task prose**: the phase-4 id-renumbering pass (mechanical, see `tasks.md` §6
   repair log) remapped structured fields (`id`, `depends_on_tasks`) but NOT free-text mentions inside task
   `description` bodies. Roughly 50 task descriptions still say "see T-1103" or similar using a
   pre-renumbering raw id that no longer exists. 5 of these are confirmed mapped in the repair log (e.g. raw
   T-1103 = T-0187); the rest were deliberately left unresolved rather than guessed. If you need to trace one
   of these, match by subject matter (what the referenced task is described as doing), not by the literal id.

## Working conventions specific to this repo

- **Requirement IDs**: `FR-NNN` functional, `NFR-NNN` non-functional, `CON-NNN` constraint, `TR-NNN` tooling
  (a 4th class, deliberately added beyond the global scheme's usual 3 — see `clarify.md` AD-001 for why).
- **Task IDs**: `T-0NNN`, globally unique across all 19 milestones (zero-padded 4 digits). Every task's
  `implements` field cites the requirement id(s) it satisfies; every task has exactly one named primary test.
- **Constitution principles**: `CP-001`..`CP-014`, in `.specify/memory/constitution.md`. These are load-bearing
  — CP-010 (Go stdlib only in core paths), CP-004 (determinism, 4 permitted non-deterministic sites, no more),
  CP-006 (verify before decode), CP-003 (no stable release without 2 independent implementations) will shape
  every implementation task. Read the constitution before writing any code, not just the spec.
- **Clarification decisions**: `CQ-001`..`CQ-012` in `clarify.md`, all resolved. Do not relitigate one; if you
  think a resolved decision is wrong, say so explicitly to Eyvar rather than silently designing around it —
  same rule the constitution's amendment process states for itself.
- **No em dashes** in any spec-phase artifact (house style carried through every generated document).
- **Git commit authorship**: sole author `Eyvar García <eyvar.0823@gmail.com>`, no `Co-Authored-By` line,
  per the user's global preference — this is stricter than the generic Claude Code default and takes
  precedence in this repo.
- **Reference implementation language**: Go 1.25, standard library only in parse/validate/canonicalise/
  verify/extract paths (CP-010). Cross-compilation targets: darwin, linux, windows on amd64 and arm64.

## How this was built (for context, not re-derivation)

Phases 0-4 were produced by five separate multi-agent Workflow runs (research/synthesis for spec, competing-
architecture judging for plan, inventory/decompose/audit/repair for tasks), each with adversarial verification
built in (skeptic agents trying to refute findings, judges scoring 5 independent architecture proposals,
mechanical traceability audits). Full methodology and every rejected option's reasoning is in `research.md` —
read it before proposing an alternative to something already decided; the tradeoff was probably already
weighed and rejected there, with evidence.

If you are an LLM/agent continuing this work and need current status without re-reading every artifact,
`README.md` in this repo root has the plain-English orientation; this file has the operational rules.
