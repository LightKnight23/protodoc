# Extension Token Registry Governance

This document is the governance/process policy for the Protodoc extension-token registry, satisfying
**CON-021** (publish a maximum extension-registry review turnaround in business days and track the
observed turnaround against it). It is a process obligation, not a wire-format mechanism; nothing here
changes any octet on disk.

## Scope

The registry issues **registered-tier** extension tokens: those with owner-id in
`0x00000001..0x7FFFFFFF` (container.abnf S5.2). Owner-scoped tokens (`0x80000000..0xFFFFFFFE`) are
self-issued and need no registry round trip, so they are out of scope for the review SLA. The
reserved (`0x00000000`) and permanently-retired (`0xFFFFFFFF`) namespaces are never issued at all.

## Published maximum review turnaround (the SLA)

**The maximum review turnaround for a registered-tier token issuance request is 15 business days**,
measured from the day a complete request is received (all required fields present) to the day a
decision (issue, decline, or request-for-changes) is published back to the requester. A request that
is incomplete on receipt does not start the clock until it is completed; the "request-for-changes"
decision stops the clock and a resubmission starts a fresh 15-business-day window.

This figure is the published ceiling, not a target: the registry aims to decide faster, but 15
business days is the number against which conformance to CON-021 is measured.

## Tracking mechanism (observed turnaround)

Observed turnaround is recorded **per request** and reported **per release**:

1. **Per request.** Every registered-tier issuance request is filed as an issue in the registry
   issue tracker carrying two dated labels: `registry:received:<YYYY-MM-DD>` (set when the complete
   request is received) and `registry:decided:<YYYY-MM-DD>` (set when the decision is published). The
   observed turnaround for that request is the number of business days between the two dates.
2. **Per release.** The release pipeline collects every request whose `registry:decided` date falls
   in the release window and reports the observed turnaround distribution (count, min, median, max,
   and the number exceeding the 15-business-day SLA) in the release notes under a
   `Registry turnaround (G-REG)` heading. Any request that exceeded the SLA is listed individually
   with its overage, so a drift is visible rather than averaged away.

Governance metric **G-REG** is exactly this published-ceiling-plus-observed-distribution pair: the
SLA figure above and the per-release observed distribution from the tracking mechanism.

## Amendment

The SLA figure may only be changed by a governance amendment (the same process that governs the
constitution), and any change is recorded with its effective date so historical turnaround
measurements remain interpretable against the SLA that was in force when each request was decided.
