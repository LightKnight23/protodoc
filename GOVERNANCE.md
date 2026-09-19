# Protodoc Governance

Status: DRAFT -- pending Eyvar Garcia's review and approval. This document
records a proposal for CON-026's four required elements (licence, named
steward, succession arrangement, deprecation window) and makes no claim
that any of it has been approved yet. The approval record, once it exists,
lives in specs/CHANGES.md (T-0356) and is cross-referenced from there back
to this file's approved revision.

This document exists to satisfy CON-026 ("The Protodoc specification SHALL
be published under an irrevocable royalty-free licence covering all
necessary patent claims, with a named steward, a documented succession
arrangement and a deprecation window stated in years, before the first
stable release") and release gate G-GOVERN, which refuses to publish
without all four elements attached.

## Licence Grant

The Protodoc specification is published under the licence in LICENSE at
this repository's root: an irrevocable, royalty-free copyright licence over
the specification text, plus an irrevocable, royalty-free patent licence
(subject only to the defensive-termination clause in LICENSE Section 2(a))
covering every necessary patent claim needed to implement it. LICENSE is
the normative text; this section exists only to point at it so a reader of
this file does not have to already know to look for a separate file.

The licence grant does not cover this repository's reference-implementation
source code (pkg/, cmd/), which is separately licensed under LICENSE-CODE:
the Apache License, Version 2.0, decided 2026-09-19. Apache 2.0 was chosen
over MIT/BSD to carry the same patent-grant-with-litigation-termination
posture the specification licence already establishes above, and over a
copyleft licence (e.g. MPL-2.0, AGPL-3.0) because broad, unencumbered
implementability is this project's explicit priority (CP-003's two-
independent-implementation requirement is easier to satisfy the more
permissively the reference implementation itself is licensed).

## Named Steward

The initial steward of the Protodoc specification is Eyvar Garcia
(eyvar.0823@gmail.com), the specification's originator and current sole
maintainer of this repository.

The steward's responsibilities are: ruling on extension-registry
submissions within the turnaround CON-021 requires, approving or rejecting
proposed specification amendments, holding the point of contact for patent
and licensing questions, and initiating the succession process below if the
steward becomes unable or unwilling to continue.

This is a draft proposal. A future steward transition (voluntary handoff,
succession under the process below, or a change to a multi-person steering
body) is itself a governance decision requiring the same approval discipline
as this document's own approval in T-0356 -- it is not self-executing from
this file alone.

## Succession Process

If the named steward becomes unable to act (sustained unresponsiveness of
180 days or more to a registry submission or amendment request, or an
explicit statement of inability or unwillingness to continue), succession
proceeds as follows:

1. Any person who has an approved, merged contribution to the specification
   or reference implementation may open a public succession request stating
   the basis for invoking this process.
2. A 30-day public comment period follows, during which any prior
   contributor may nominate themselves or another contributor as successor
   steward.
3. If exactly one nomination stands unopposed at the end of the comment
   period, that nominee becomes steward.
4. If more than one nomination stands, or any nomination is formally
   opposed, the outstanding contributors (any person with an approved,
   merged contribution) decide by simple majority of those who cast a vote
   within a further 15 days.
5. If no nomination is made at all, the specification is considered
   unstewarded; its licence (LICENSE) remains in force regardless -- the
   irrevocable patent and copyright grants in LICENSE Section 2(b) survive
   an unstewarded specification by design, so the absence of a steward
   never revokes an implementer's existing rights.

This process governs stewardship of the specification's governance
activity only. It has no bearing on LICENSE's grants, which are irrevocable
independent of who holds or whether anyone holds the steward role.

## Deprecation Window Policy

Any future breaking change to a stable (post-v1.0) release of the Protodoc
specification -- a change that would cause a document previously conforming
under a stable version to fail conformance under the new version, or a
change to the wire format's normative encoding -- is subject to a minimum
deprecation window of 3 years from the date the replacement version is
published, before the deprecated version may be marked unsupported in this
repository's own conformance tooling.

During the deprecation window:

- the deprecated version's conformance test suite continues to be
  maintained and run in CI alongside the current version's;
- the deprecated version's specification text remains published and
  reachable at a stable location;
- a security-relevant defect in the deprecated version is still eligible
  for a fix, though not for new features.

3 years is proposed here as a minimum, not a target; the steward may commit
to a longer window for a specific deprecation without amending this policy,
but may not commit to a shorter one without first amending this document
through the same approval process as this document's own adoption.
