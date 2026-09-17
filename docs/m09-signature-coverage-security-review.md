# M09 signature-&-coverage milestone-exit argus review (T-0363)

Consolidated security review of the whole M09 surface — CoverageDescriptor
well-formedness, signed_object / SIGNATURE / PRESENTATION_ARTEFACT
construction, the Verify() orchestration and its failure classification, and
the NFR-005 ambient-value audit — before any downstream milestone (M10 redaction
commitment, M11 redaction, publish/sign/migrate) depends on it. Reviewed at the
commit that introduces this file, per CP-011 and matching the M06 (T-0112) and
M08 (T-0362) milestone-exit review precedents.

## Findings summary

Zero unresolved P0 or P1 findings at review close.

One P1 finding WAS discovered and RESOLVED during the milestone: T-0150's
coverage-canonicalisation fuzz found an allocation-DoS in `decodeRanges`
(coverage.go sized a slice from an untrusted varint count before bounds-checking
it against the remaining input, panicking on a hostile count). It was fixed in
the T-0150 commit (count bounded by `len(remaining)/2` before any allocation,
FR-106 check-before-allocate) and the crashing input was committed as a
regression seed. No P0/P1 remains open; any future finding is filed as a new
task, never fixed silently inside this artifact.

No finding requires an Eyvar waiver: the sole P1 was resolved in code, not
waived. (Had any finding needed a waiver, it would be recorded here as OPEN and
awaiting Eyvar, never marked resolved without one.)

## Files reviewed (every file touched by T-0145..T-0166)

- `pkg/integrity/coverage.go` (CoverageDescriptor + T-0150 allocation fix)
- `pkg/integrity/signedobject.go` (signed_object preimage + slot-sourced digest)
- `pkg/integrity/signature.go` (SIGNATURE record)
- `pkg/integrity/presentation.go` (PRESENTATION_ARTEFACT + FontRecord)
- `pkg/integrity/presentationattest.go` (FR-065 attestation)
- `pkg/integrity/verdict.go` (closed 4-value verdict enum)
- `pkg/integrity/verify.go` (per-state Verify orchestration + coverage guard)
- `pkg/integrity/diffunits.go` (FR-068 differing-unit enumeration)
- `pkg/integrity/staleness.go` (FR-069 exemption)
- `pkg/integrity/ambient.go` (NFR-005 allowlist)

## Per-task review (T-0145 .. T-0166)

Task-ID coverage manifest (every task reviewed, cited individually):
T-0145, T-0146, T-0147, T-0148, T-0149, T-0150, T-0151, T-0152, T-0153,
T-0154, T-0155, T-0156, T-0157, T-0158, T-0159, T-0160, T-0161, T-0162,
T-0163, T-0164, T-0165, T-0166.


- **T-0145..T-0149 CoverageDescriptor round trip + PD-COVER-001..004.** PASS.
  Single canonical implementation reused (no second encoder); all four
  structural rules (zero-length, unmerged/unsorted, gap/overlap, ATTEST-named)
  reject with distinct errors.
- **T-0150 canonicalisation fuzz.** PASS after the P1 fix above; 4.3M execs
  clean, decode/encode/decode stable, one-encoding-per-validating-set.
- **T-0151 coverage-descriptor-digest.** PASS. SHA-256 over the canonical
  encoding; deterministic and sensitive to every range/bitmask/mode change.
- **T-0152 signed_object preimage.** PASS. Fixed 129-octet
  `%x04||t_c_root||structure_digest||presentation_digest||coverage_digest`,
  domain-tagged, sensitive to each input.
- **T-0153 SIGNATURE record.** PASS. Byte-exact round trip; non-zero16
  cred-chain and time-attestation refs enforced; MAX_SIGNATURES = 64.
- **T-0154 PresentationArtefact + FontRecord.** PASS. Closed 7-field font
  record; count-prefixed lists all bounds-checked before allocation;
  codepoints range-checked to 0..0x10FFFF; i64 page geometry (signed).
- **T-0155 presentation digest sourced from slot.** PASS. `signed_object`'s
  presentation input is the referenced segment's own slot digest, never an
  inline copy; an unresolved ref is rejected (no-trust-inline control).
- **T-0156/T-0158/T-0159/T-0161 verify orchestration + verdict classification.**
  PASS. Verdict order is UnavailableState (FR-062) > Unattested (FR-065) >
  cryptographic Unverified/Valid; every failure path returns a non-Valid
  verdict and NO signer identity.
- **T-0157 closed four-value verdict enum.** PASS. Exactly {valid, unverified,
  unavailable_state, unattested}; out-of-range values are undefined and
  identity-free.
- **T-0160 signature preserved across edit.** PASS. A preserved signature
  round-trips byte-identically, still names its state and presentation, still
  verifies Valid against its own state, and does not carry to the edited state.
- **T-0162 differing-unit enumeration.** PASS. Exhaustive, deterministic
  added/removed/changed classification.
- **T-0163 staleness exemption.** PASS. A signature-bound presentation is never
  refused as stale relative to the current state (FR-069).
- **T-0164 unverified never carries identity.** PASS. All three FR-115 triggers
  (bad signature, disallowed param set, acted-upon octet outside coverage)
  yield a non-Valid verdict with no signer identity or positive badge.
- **T-0165 ambient-value audit (NFR-005).** PASS. The allowlist is exactly the
  four named sites (identifier minting, redaction salts, signature values, time
  attestations); the source audit confirms the signature/coverage layer imports
  no ambient source (time, os, os/user, crypto/rand, math/rand) in non-test
  code. Every ambient-audit row was checked; none failed.
- **T-0166 at/over-limit fixtures.** PASS. MAX_SIGNATURES and coverage-partition
  boundaries pinned by paired accept/reject fixtures.

## Load-bearing controls certified

- **Domain separation.** signed_object is prefixed with DomainSignedObject
  (0x04); the coverage digest is a distinct SHA-256, not raw bytes reused.
- **no-trust-stored / no-trust-inline.** t_c_root, structure_digest, the
  presentation digest, and the coverage digest are all recomputed or read fresh
  at verify time; no inline or cached value is trusted.
- **identity discipline.** A signer identity is surfaced only alongside a Valid
  verdict; every other verdict is identity-free (FR-115).
- **check-before-allocate.** Every count-prefixed decode in the milestone
  (coverage ranges, font-record axes/codepoints/features, page geometry) bounds
  the declared count against remaining input before allocating (FR-106).
