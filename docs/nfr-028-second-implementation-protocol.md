# NFR-028 / CP-003 Second-Implementation Acceptance Protocol

Status: DRAFT, pending Eyvar's review and approval. This document defines the
acceptance protocol only: what scope, independence, corpus and match mean,
and how a disagreement is triaged. It does not itself satisfy NFR-028,
CON-019 or CP-003 (those require an actual second implementation to exist
and pass against this protocol), and it does not record the funding or
sponsorship decision plan.md Section 8's risk row still asks Eyvar to make
before that second implementation can be recruited (tracked separately, see
`specs/001-protodoc-format-core/plan.md` Section 8 and `clarify.md` CQ-012).

## Purpose

NFR-028 requires:

> The Protodoc format SHALL NOT be declared stable until two implementations
> written independently in different languages by different authors produce
> identical canonical octets on the conformance corpus and identical
> accept-or-reject verdicts on a negative corpus of at least 200 hostile
> cases.

CON-019 requires the same two-independent-implementation bar per feature
before that feature is marked normative. Both are release gates with no
executable meaning until "independent," "match" and "corpus" are each
defined precisely enough for a mechanical pass/fail. This document is that
definition.

## Scope

In scope, exactly:

- `container` -- the fixed 1,048,576-octet prefix (Header, CommitRing,
  Frontmatter, SegmentTable) and its decode/encode/verify surface.
- `validate` -- the structural validation pipeline (data-model.md Section 7
  steps 1-9 and 11-13; step 10, signature verification, belongs to `verify`
  and is out of this gate's scope).
- `canon` -- the `L*(state)` canonical-serialiser traversal used by
  `publish`, `compact`, `migrate` and fresh writes.

Out of scope, explicitly exempt from this specific gate per T-0348's own
cross-cutting note (rendering/merge/history/etc. are not required to clear
a second-implementation bar before v1): `render`, `merge`, `history`,
`extract`, `migrate`'s non-canon-reuse portions, `registry`, and every other
package not named above. A second, independently-authored implementation of
those packages is not a precondition for declaring v1 stable.

This scoping is not invented here: it is CQ-012's own approved resolution
("Option A, scoped to container, validator and canonical serialiser. Those
are exactly the layers the byte-identity and verdict-equality gates test,
and the layers where divergence is unrecoverable once files exist") and
plan.md's dependency-spine item 10, both already approved. This document
only makes that scope executable.

## Independence criteria

An implementation counts as "independently authored" (NFR-028's own words:
"written independently in different languages by different authors") only
if every one of the following holds:

1. **Different author.** The primary author or authors have no employment,
   contract, or equivalent working relationship with this reference
   implementation's authorship during the second implementation's
   development.
2. **Different language.** NFR-028's text requires this explicitly, not
   merely prefers it: a second Go implementation, however independently
   typed, does not clear this gate.
3. **No shared source access.** The second implementer has never had access
   to this repository's `pkg/container`, `pkg/pdlfmt`, or any future
   `pkg/validate`/`pkg/canon` source, at any point before or during their
   own implementation work. Reading a compiled binary's behavior (black-box
   observation against the public CLI contract) is not source access;
   reading this repository's `.go` files is.
4. **Permitted inputs, exhaustively.** The second implementer may read only:
   the frozen normative artifacts (`spec.md`, `clarify.md`, `plan.md`,
   `research.md`, `data-model.md`, everything under `contracts/`), and the
   shared conformance corpus (input case files and their recorded expected
   verdicts/octets, never this reference implementation's own Go test
   helper functions or internal fixture-construction code).
5. **Written attestation.** Before any comparison run counts toward the
   gate, the second implementer signs and dates a short attestation naming:
   any prior familiarity with this reference implementation's source (none,
   or disclosed in full), the exact list of artifacts consulted, and
   confirmation of points 1-4 above.

## Authoritative corpus

The corpus a comparison run is scored against must be a single, named,
version-tagged artifact, never "whatever files happen to be on disk" at run
time. This protocol names the corpus version `PDL-CONFCORPUS-M19-V1`: the
corpus assembled and versioned by the differential conformance harness
(`specs/001-protodoc-format-core/tasks.md` T-0350), covering at minimum:

- Every CON-010 at-limit and one-past-limit ceiling pair shipped to date.
- PD-RING-001's equal-sequence tie-rejection vector.
- The `frame_count*48+32 <= segment_length-64` at-limit/over-limit
  boundary vector (CON-010's frame-count case).
- The negative corpus of at least 200 hostile cases NFR-028's own text
  requires, once assembled (not yet complete in this repository as of this
  document; tracked as a residual gap, not silently assumed done).

Any addition, removal or edit to a corpus case bumps the version identifier
(`...-V2`, `...-V3`, ...); a comparison run must record which corpus version
it ran against, and two runs against different corpus versions are not
comparable.

## Match definition

A comparison run passes only if, for every case in the authoritative
corpus:

1. **Canonical octets match byte-for-byte.** Exact equality of the full
   output byte sequence for every corpus case that exercises `container`
   or `canon`, never a digest-only comparison or a normalized/whitespace-
   tolerant comparison. A SHA-256 digest may be used as a cheap first-pass
   filter, but a digest match alone is never sufficient to record a case as
   passing; the full byte sequence must actually have been compared.
2. **Verdicts match exactly.** Identical `status` name and identical
   `exit_code`, per `contracts/cli.md` Section 1's closed 8-value table, for
   every corpus case that exercises `validate`. A case is not "close
   enough" if the exit codes agree but the `findings` rule IDs differ, or
   vice versa; both are part of the recorded verdict this protocol scores.
3. **100%, not a threshold.** NFR-028's text states an absolute bar
   ("identical canonical octets," "identical accept-or-reject verdicts"),
   not a percentage. A single disagreeing case fails the run; there is no
   passing score below 100% of the corpus on both dimensions.

## Disagreement triage

When a comparison run finds a disagreement on any case, triage proceeds in
this fixed order, and the outcome of each step is recorded before moving to
the next:

1. **Reproduce.** Re-run both implementations against the exact same corpus
   case file at least twice each. A disagreement that does not reproduce
   deterministically is itself a CP-004 determinism defect in whichever
   implementation varies, filed as such, before any spec-versus-
   implementation question is even asked.
2. **Check against the corpus's own recorded expectation.** Every corpus
   case carries a recorded expected verdict and, where applicable, expected
   canonical octets. If exactly one implementation matches the recorded
   expectation, the diverging implementation has an implementation bug,
   filed against that implementation, not against the spec.
3. **If both implementations agree with each other but not with the
   recorded expectation**, suspect the corpus fixture first: re-derive the
   expected verdict/octets by hand against the normative ABNF/spec text. If
   the fixture's recorded expectation was wrong, fix the fixture (a patch-
   tier change per this project's own tiering rules) and re-run; this is
   not a spec bug.
4. **If both implementations agree with each other, the fixture's recorded
   expectation is confirmed correct by hand-derivation, and the
   disagreement still stands** (this step is only reached when a
   disagreement persists after step 3 has already fixed the fixture, so it
   should be rare in practice), the normative text is ambiguous. This is
   escalated to Eyvar through the project's existing `clarify.md` process
   as a new, numbered clarification question; it is never resolved
   unilaterally by whichever agent found it, and it never silently amends
   an already-resolved CQ.
5. **Record the outcome.** Every triaged disagreement, its category
   (implementation bug / fixture bug / spec ambiguity), and its resolution
   is logged as one line in `specs/001-protodoc-format-core/CHANGES.md`,
   per this project's patch-tier logging convention, regardless of which
   category it fell into.

## Non-goals

This document does not fund, recruit, or commit to a second implementation.
The CQ-012 / plan.md Section 8 funding-and-sponsorship decision remains open
and is tracked as its own, separate governance item; this protocol defines
only what "pass" means once a second implementation exists to run it
against.
