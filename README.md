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

**Implementation nearly done: 363 of 372 tasks done, verified against git history.** All specs (phases 0-5) are
approved and frozen. Phase 6 (implement): milestones M01 through M18 are complete (every task, `go build/vet/
test ./...` green); M19 (Conformance, Fuzzing & Governance Convergence) has 11 of 20 tasks done. The 9 remaining
M19 tasks are real-world actions no agent can complete by writing code — external-implementer trials, Eyvar's
personal governance sign-offs, a funding decision, commissioning a second independent implementation, and an
actual IANA registration filing. See [`CLAUDE.md`](CLAUDE.md) § "Phase 6 status" for the exact, git-verified
breakdown and the discipline for continuing safely — **read it before writing any code.**

| Artifact | What it is |
|---|---|
| [`.specify/memory/constitution.md`](.specify/memory/constitution.md) | 14 non-negotiable project principles (v0.2.0) |
| [`specs/001-protodoc-format-core/spec.md`](specs/001-protodoc-format-core/spec.md) | 197 requirements, EARS notation, WHAT/WHY only |
| [`specs/001-protodoc-format-core/clarify.md`](specs/001-protodoc-format-core/clarify.md) | 17 resolved design-fork decisions |
| [`specs/001-protodoc-format-core/plan.md`](specs/001-protodoc-format-core/plan.md) | Chosen architecture, HOW |
| [`specs/001-protodoc-format-core/research.md`](specs/001-protodoc-format-core/research.md) | 5 competing architectures, why one won |
| [`specs/001-protodoc-format-core/data-model.md`](specs/001-protodoc-format-core/data-model.md) | Entities, fields, ceilings, ordering rules |
| [`specs/001-protodoc-format-core/contracts/`](specs/001-protodoc-format-core/contracts/) | Normative wire grammars (ABNF) + CLI contract |
| [`specs/001-protodoc-format-core/tasks.md`](specs/001-protodoc-format-core/tasks.md) | 19 milestones, 372 implementation tasks |
| [`specs/001-protodoc-format-core/analysis.md`](specs/001-protodoc-format-core/analysis.md) | Phase 5 cross-artifact consistency, gate PASSES |
| [`scripts/extract_milestone_tasks.py`](scripts/extract_milestone_tasks.py) | Regenerates the per-milestone task breakdown + real topological build order from `tasks.md` |

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
specs/001-protodoc-format-core/   The full spec-to-tasks-to-analysis pipeline (see table above)
scripts/extract_milestone_tasks.py  Regenerates .impl_tasks/ (gitignored) from tasks.md
pkg/pdlfmt/                       PDL-VARINT, PDL-TLV, and value-kind wire primitives — done (M01)
pkg/container/                    Header, CommitRingRecord, Frontmatter, SegmentTableSlot — done (M01)
pkg/ledger/                       Ledger & Placement — done (M02)
pkg/content/, pkg/content/mint/   Identity & Anchor — done (M04)
pkg/extract/                      Extraction — done (M05)
pkg/eddsa/                        Signature Primitive (EdDSA-Protodoc-1) — done (M06)
pkg/integrity/                    Integrity Trees (T_S/T_C) — done (M08)
pkg/validate/                     Structural Validation Core — done (M07)
                                  Extensibility Envelope — done (M03)
                                  Signature & Coverage — done (M09)
                                  Attestation Evidence & LTV — done (M10)
pkg/history/                      Redaction, History & Erasure — done (M11, M12)
pkg/merge/                        Concurrent-Edit / Merge — done (M13)
pkg/render/                       Rendering & Resource — done (M14)
pkg/semantics/                    Accessibility & Semantic Content-Model Extensions — done (M15)
pkg/migrate/, pkg/registry/       Evolution / Migration & Registry — done (M16)
pkg/canon/                        Canonicalization — done (M17)
pkg/cli/                          CLI Surface — done (M18)
pkg/governance/, pkg/traceability/, pkg/diffconform/, pkg/fuzzmaturity/, pkg/benchconfig/, pkg/ceilings/
                                  Conformance, Fuzzing & Governance Convergence — 11/20 (M19); remaining 9
                                  tasks are real-world actions (external trials, sign-offs, IANA filing)
cmd/                              CLI entry points for the packages above
go.mod                            Go 1.25 module
```

## What's left

All implementation work is done — M01 through M18, every task, and no code-writing tasks remain anywhere in
the project. The only open items are M19's 9 real-world actions that no amount of code can satisfy: two
external-implementer trials, Eyvar's personal governance sign-offs, a funding decision for a second
independent implementation, commissioning that implementation, and the IANA/PRONOM format registration
filings (in progress as of this writing). See [`CLAUDE.md`](CLAUDE.md)'s "Phase 6 status" section for the
exact, git-verified breakdown of which of those 9 remain and why each one specifically requires a human,
not an agent.
