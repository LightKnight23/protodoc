# Protodoc Security Disclosure Policy

Status: DRAFT -- pending Eyvar Garcia's review and approval. This document
proposes the triage owner and disclosure clock CP-012 requires before this
repository's fuzz targets (pkg/fuzzmaturity and the corpora it aggregates
maturity reports for) may be enrolled in any public fuzzing service (for
example OSS-Fuzz). It makes no claim that any such enrolment has happened,
and none is authorized by this document alone.

This document exists to satisfy CP-012 ("The reference library is under
continuous coverage-guided fuzzing seeded from the conformance corpus... A
named triage owner and a published maximum of 90 days from report to fix or
advisory exist before enrolment in any public fuzzing service") and its
Consequence clause: "The disclosure clock is staffed before it starts, not
after the first finding arrives."

## Triage Owner

The triage owner role is held by Eyvar Garcia (eyvar.0823@gmail.com), in
the same capacity as GOVERNANCE.md's Named Steward role, until a separate
triage role is delegated by that same approval process. The triage owner
is responsible for: acknowledging an incoming report, assessing its
severity, assigning or personally driving the fix or advisory work, and
tracking the report against the 90-day clock below.

This is a draft proposal, not a staffing commitment made unilaterally by
whichever agent or contributor drafted this file -- it takes effect only
once Eyvar approves it, the same discipline GOVERNANCE.md's Named Steward
section applies to itself.

## Disclosure Clock

From the date a vulnerability report is received, the triage owner
publishes one of the following within 90 calendar days:

1. a released fix, referenced from specs/CHANGES.md with the report's
   tracking reference; or
2. a public advisory explaining why no fix is released within the window
   (for example: disputed severity, upstream dependency, or a fix requiring
   a breaking change subject to the deprecation window policy in
   GOVERNANCE.md).

90 days is a maximum, not a target -- the triage owner may resolve sooner
and is expected to for a high-severity, easily reproduced defect. The
clock starts on receipt of the report, not on the day it is first read or
triaged, so a delay in acknowledgement does not extend the window.

## How to report

Report a suspected vulnerability by emailing eyvar.0823@gmail.com. Do not
open a public issue for an unpatched vulnerability; the triage owner will
coordinate public disclosure once a fix or advisory is ready.

## Precondition for public fuzzing-service enrolment

Per CP-012's Consequence clause, no task that enrolls this repository's
fuzz targets in a public fuzzing service may proceed until this document
carries Eyvar's recorded approval (tracked the same way as GOVERNANCE.md's
approval, in specs/CHANGES.md). Enrolling before that approval exists would
start public intake of reports against an unstaffed disclosure clock, which
is exactly the ordering CP-012 forbids.
