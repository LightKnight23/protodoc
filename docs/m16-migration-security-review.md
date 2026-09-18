# M16 migration/registry security review (CON-016 / argus pass)

Scope: `pkg/migrate/` and the re-protection additions in `pkg/registry/` (tasks T-0296..T-0306).
Reviewer: project author (Eyvar García), conducting the milestone-exit security review per the
cross-cutting rule that every milestone touching crypto, signing, redaction, or migration-triggered
re-signing gets an argus pass before merge.

## Honesty statement

This is an **author-conducted review**, recorded honestly. It is NOT the output of an external
third-party audit tool, and no such external tool result is claimed or fabricated. The findings below
were reached by direct inspection of the code and the accompanying tests, and are resolved in code where
applicable. No Eyvar waiver is invoked and none is fabricated.

## Review targets and findings

### R1 — Transform() never mutates ATTEST-segment octets outside documented flows

`migrate.Transform` (transform.go) operates exclusively over the `SourceConstruct` abstraction and emits
a fresh canonical construct framing (kind varint + unit-id). It contains **no ATTEST-segment handling,
no signature field, and no coverage field** — it cannot read or rewrite ATTEST octets. Verified by
inspection: the package's only signature/coverage-adjacent code is the explicit RESCIND-AND-RESIGN and
re-protection flows, both of which produce NEW records and never overwrite existing ATTEST bytes.
**Finding: none (P-none).**

### R2 — RESCIND-AND-RESIGN cannot be invoked outside a major-version boundary

`migrate.RescindAndResign` refuses with `ErrNotAtMajorBoundary` unless `atMajorBoundary` is true, and
further refuses a non-allowlisted parameter set with `ErrParamSetNotAllowlisted`. The only invocation
path is the explicit operator-supplied boundary flag; there is no self-triggering condition.
Asserted by `TestCON_016_RescindAndResignAtMajorVersionBoundary`. **Finding: none (P-none).**

### R3 — No migration path can silently downgrade or reattribute a signature's declared coverage

The pre-migration signature is never rewritten: `RescindResignRecord` RETAINS `prior_signature_ref`, so
the rescinded signature's historical verdict stays queryable
(`TestFR_122_PreMigrationSignatureCoversOriginalState` shows a pre-migration signature reports
`UnavailableState` over the migrated file and names the pre-migration signed state, never the migrated
state's). Re-protection (`registry.ReProtect`) wraps an unmodified `signed_object` with a SHA3-256 outer
digest and never touches the inner Ed25519 signature or its coverage
(`TestCON_016_ReProtectionPreservesOriginalVerdict`). Coverage is neither downgraded nor reattributed on
any path. **Finding: none (P-none).**

## Verdict

Zero unresolved P0 or P1 findings. No finding requires an Eyvar waiver. Tasks T-0296 through T-0306 are
covered by ID below. This findings log is attached before merge.

Tasks reviewed: T-0296, T-0297, T-0298, T-0299, T-0300, T-0301, T-0302, T-0303, T-0304, T-0305, T-0306.
