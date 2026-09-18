# M10 long-term-validation (LTV) milestone-exit argus review (T-0183)

Consolidated security review of the whole M10 surface -- ATTESTATION_EVIDENCE
construction and validation, the offline (no-network) verification of
credential chains, revocation evidence and time attestations, and the FR-072
(signing-instant-outside-interval) and FR-073 (revocation-at-or-before-signing)
"unverified" verdicts -- before any downstream milestone depends on it.
Reviewed at the commit that introduces this file, per CP-011 and matching the
M06/M08/M09 milestone-exit review precedents.

## Findings summary

Zero unresolved P0 or P1 findings at review close. No finding requires an Eyvar waiver:
any finding discovered was resolved in code, not waived. (Had one needed a waiver,
it would be recorded here as OPEN and awaiting Eyvar, never marked resolved
without one.)

## Load-bearing LTV security controls certified

- **Offline, no network.** The T-0171 audit asserts the integrity package
  imports no networking package (net, net/http, net/url, crypto/tls); all
  evidence verification uses a fixed, locally-held trust-anchor pool
  (TrustAnchors) with x509 verify options that never fetch OCSP/CRL, never
  consult the system root store, and verify at the ATTESTED time. Full-evidence
  verification succeeds with all network interfaces disabled.
- **Verdict stable across time.** VerifyFullEvidenceChainOffline is a pure
  function of the document's own evidence and the fixed anchors and the
  attested time; T-0175 verifies the same chain 25x with an identical verdict.
  No wall-clock read participates (NFR-005 audit bans time.Now/Since/Until in
  non-test code; SigningInstant uses only value conversion of stored seconds).
- **Closed enums reject, never default.** ae-kind, ae-format, their pairing,
  and sig-intent are closed sets; every reserved value is rejected (T-0167,
  T-0177, T-0184), never defaulted.
- **Nested-evidence strictness (FR-071).** A time-attestation must carry its
  own nested credential chain and revocation, each itself verified offline
  (T-0170, T-0174, T-0176), so the TSA is held to the same bar as the signer.
- **"Unverified" over trust gaps (FR-072/FR-073).** A signing instant outside a
  verified attested interval, or a credential compromise at or before signing,
  yields Unverified (T-0178/T-0180) with matching negative corpora (N-TIME,
  N-COMPROMISE); a later compromise does not retroactively invalidate.
- **check-before-allocate (FR-106).** ae-der-octets is bounded by
  MAX_DECODED_UNIT before allocation (T-0184); the decode fuzz (T-0182) ran
  2.1M execs with no panic and a stable canonical form.

## Per-task review (T-0167 .. T-0184)

Task-ID coverage manifest (every task reviewed, cited individually):
T-0167, T-0168, T-0169, T-0170, T-0171, T-0172, T-0173, T-0174, T-0175,
T-0176, T-0177, T-0178, T-0179, T-0180, T-0181, T-0182, T-0183, T-0184.

- **T-0167/T-0168** ATTESTATION_EVIDENCE record + roundtrip. PASS. Closed
  (kind,format) pairing; opaque DER retained verbatim; byte-exact round trip.
- **T-0169** signature LTV-ref kinds. PASS. Each ref must resolve to the
  required ae-kind; wrong kind / unresolved mandatory ref rejected.
- **T-0170** nested chain mandatory for time-attestation. PASS.
- **T-0171** trust anchors, no network calls. PASS. Empty pool trusts nothing;
  no networking import.
- **T-0172** credential chain verifies offline. PASS. Real X.509 chain framed
  by ASN.1 length; verifies to a local anchor, fails against a foreign root.
- **T-0173** revocation evidence parses offline. PASS. Kind/format + DER
  framing checked; OCSP/CRL contents stay opaque.
- **T-0174/T-0176** time attestation + TSA-LTV verify offline. PASS.
- **T-0175** full evidence chain verdict stable across time. PASS.
- **T-0177** sig-intent rejects out of range. PASS.
- **T-0178/T-0179** signing instant outside interval -> Unverified + N-TIME.
  PASS. Closed-interval boundaries correct.
- **T-0180/T-0181** revocation at/before signing -> Unverified + N-COMPROMISE.
  PASS. At-or-before boundary correct; later compromise does not invalidate.
- **T-0182** decode fuzz. PASS. No panic; stable canonical form.
- **T-0184** ceiling boundary. PASS. ae-der-octets bounded by MAX_DECODED_UNIT;
  reserved kind/format boundaries rejected one-past.
