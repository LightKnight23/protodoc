# Protodoc

A document format designed from scratch, plus its reference implementation. Not compatible with OOXML/PDF and
not intended to be — it exists to make properties those formats cannot guarantee into checkable facts about the
file itself.

## Why

Word-family formats (OOXML) repack the whole file on every save and let signatures cover an enumerated subset
the signer chooses; PDF stores rendered layout instead of content and lets appended, unsigned octets render
anyway. Both leak deleted content and identity through provenance surfaces nobody can fully enumerate, and both
have essentially one complete implementation each after decades. Protodoc's bet: a format where

- **saving costs octets proportional to the edit**, not the file size (real diffs, real git history),
- **a signature covers everything a reader can act on**, or says exactly what it doesn't,
- **redaction is provably total**, not "mostly removed,"
- **annotations survive concurrent editing** because they're anchored to identity, not character offsets,
- **two independent implementations produce byte-identical output** on the same input, always,

are structural guarantees, not aspirations. Full case: [`specs/001-protodoc-format-core/spec.md`](specs/001-protodoc-format-core/spec.md) § 1-2.

## Status

**Pre-code.** This project follows Spec-Driven Development: every requirement traces to a task, every task
traces to a test, and none of that exists yet as code. See [`CLAUDE.md`](CLAUDE.md) for the exact phase-by-phase
status and what's blocking implementation from starting.

| Artifact | What it is |
|---|---|
| [`.specify/memory/constitution.md`](.specify/memory/constitution.md) | 14 non-negotiable project principles |
| [`specs/001-protodoc-format-core/spec.md`](specs/001-protodoc-format-core/spec.md) | 197 requirements, EARS notation, WHAT/WHY only |
| [`specs/001-protodoc-format-core/clarify.md`](specs/001-protodoc-format-core/clarify.md) | 12 resolved design-fork decisions |
| [`specs/001-protodoc-format-core/plan.md`](specs/001-protodoc-format-core/plan.md) | Chosen architecture, HOW |
| [`specs/001-protodoc-format-core/research.md`](specs/001-protodoc-format-core/research.md) | 5 competing architectures, why one won |
| [`specs/001-protodoc-format-core/data-model.md`](specs/001-protodoc-format-core/data-model.md) | Entities, fields, ceilings, ordering rules |
| [`specs/001-protodoc-format-core/contracts/`](specs/001-protodoc-format-core/contracts/) | Normative wire grammars (ABNF) + CLI contract |
| [`specs/001-protodoc-format-core/tasks.md`](specs/001-protodoc-format-core/tasks.md) | 19 milestones, 365 implementation tasks |

## The architecture in one paragraph

**Protodoc Ledger (PDL)**: a fixed 1 MiB prefix (magic header, a self-digesting 7-slot commit ring, a
16,384-slot segment table) followed by an append-only ledger of sealed segments. Two Merkle trees over the
content — a storage-integrity tree and a content-commitment tree — feed one signed root, so signature coverage
is a structural property of the tree rather than a list the signer has to get right. Text has per-character
identity (minted once, never reissued, run-merged for size) instead of offsets, so annotations and edits from
different collaborators reconcile deterministically instead of racing. Everything structured is encoded with a
bespoke deterministic TLV grammar (byte-lexicographic canonical order, no floats, no ambiguous forms) rather
than protobuf, because proto3 does not guarantee identical canonical bytes across independent implementations
and this format's signing story depends on that guarantee. Full detail, every rejected alternative, and why:
[`plan.md`](specs/001-protodoc-format-core/plan.md) and [`research.md`](specs/001-protodoc-format-core/research.md).

## Repo layout

```
.specify/memory/constitution.md   Governing principles (read before any design or code decision)
specs/001-protodoc-format-core/   The full spec-to-tasks pipeline for the core format (see table above)
go.mod                            Bare Go 1.25 module — no source yet, intentionally
```

Implementation code (`internal/`, `pkg/`, `cmd/protodoc/`) does not exist yet. It appears once phase 5
(analyze) has run and been approved — see `CLAUDE.md` for what that means concretely.

## For another LLM or agent picking this up

Read [`CLAUDE.md`](CLAUDE.md) first — it has the exact current phase, two disclosed gaps in the current
`tasks.md` that phase 5 needs to resolve, and the working conventions (ID schemes, commit author policy, house
style) specific to this repo. This README is the orientation; `CLAUDE.md` is the operating manual.

Do not start writing implementation code on your own initiative. The project's own constitution (CP-001, "spec
first, always") makes untraceable code a defect subject to removal, not a head start.
