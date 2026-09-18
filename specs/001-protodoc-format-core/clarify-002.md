# clarify-002.md — M15 accessibility & semantic content-model design ruling

Status: **OPEN — awaiting Eyvar's ruling** (requested 2026-09-18). This addendum is the T-0267
governance phase-gate memo. Per SDD / CP-011 / AD-002, no `plan.md` / `data-model.md` / `document.abnf`
amendment proceeds without recorded sign-off. This memo enumerates the missing constructs and their
proposed field shapes and **requests** Eyvar's per-construct approval; it does **not** record an approval
that has not been given.

## Why this memo exists

`spec.md`'s FR-032/033/034/036/037/038/039/040/082/083/084/085/099/118 and NFR-031 are frozen and correct
as WHAT/WHY. The gap is in the HOW artifacts: `plan.md` / `data-model.md` / `document.abnf` never added a
mechanism for 16 of the 18 spec-named validator rules those requirements need. Per SDD this is a
plan/data-model **amendment**, not a requirement reopening — but an amendment to frozen phase-3/4 artifacts
still needs Eyvar's sign-off (CP-011/AD-002). This memo is that request.

## Spine dependency-edge gap (flagged for the same ruling)

The spine records M15 as depending on **M04 alone**. In practice T-0275/T-0276 retrofit **M08** (T_C
traversal reading order) and **M05** (extract walk order). This is a spine dependency-edge gap; it is
surfaced here for Eyvar to rule on alongside the construct approvals rather than silently relied upon.

## The 8 construct groups (proposed field shapes)

Each group is **PROPOSED, not ratified**. The M15 implementation tasks (T-0268..T-0293) build these
constructs **provisionally** against the shapes below; the work is **reversible** until Eyvar rules, exactly
as M11's actor-identity inventory proceeded against a provisional reading.

1. **Base-writing-direction field (FR-032/FR-033)** — a mandatory closed-enum `direction` field
   {LTR, RTL} on every text-container entity (TextBlock, Table, top-level document unit); in-band
   direction control characters are rejected (FR-033). Validator rules: PD-BIDI-001 (missing direction),
   PD-BIDI-002 (in-band control rejected).
2. **ROOT_SEQUENCE reading-order record (FR-036)** — a document-level ordered list of content-unit
   identities defining the canonical reading order, walked by T_C traversal and by extraction.
3. **Computed-inline isolation wrapper (FR-034)** — a wrapper marking a computed inline run so its bidi
   isolation is computed deterministically, not host-inferred.
4. **Accessibility role-mapping table (FR-037/FR-038)** — a closed registry mapping every structural
   construct to an accessibility role; PD-A11Y-002 resolves a construct to its role, and PD-A11Y-001
   checks accessibility mark placement (FR-037).
5. **Table header-scope + cell-tiling invariant (FR-039/FR-082)** — a `scope` field on table headers and
   a tiling invariant that cells exactly tile the table grid (no gaps/overlaps).
6. **Alt-text field (FR-040)** — a mandatory text-alternative field on non-text content, with a quality
   check (non-empty, not a filename echo).
7. **Ordered-sequence numbering construct (FR-083)** — a deterministic numbering-label generator; literal
   baked-in numbers are rejected (PD-A11Y-004).
8. **Cross-reference presentation-function + staleness (FR-084/FR-085)** and **2D-region record
   (FR-099)** and **inferred-value marker (FR-118)** — an xref presentation-function field with a
   staleness flag resolvable without layout; a 2D-region record carrying a linearised reading order and
   text alternative; and an inferred-value marker enumerating which values were inferred vs authored.

## Ruling requested

Eyvar to approve (or amend) each of the 8 construct groups' field shapes, and to rule on the M04-only
spine dependency-edge gap. Until then every M15 construct is provisional and reversible.

- **Date requested:** 2026-09-18.
- **Status:** OPEN. No construct is recorded as approved. No Eyvar sign-off is fabricated here.
