# EdDSA-Protodoc-1 security review checklist (T-0112, argus)

Review of the complete `pkg/eddsa` package (the EdDSA-Protodoc-1 signature
primitive) before any downstream Signature / CoverageDescriptor / sign /
verify / redact / publish / migrate task is allowed to depend on it.

Scope: `paramset.go`, `sign.go`, `smallorder.go`, `checkpoint.go`,
`scalar.go`, `verify.go`, `verifyparamset.go` and their tests, at the commit
that introduces this checklist.

## Findings summary

Zero open HIGH or CRITICAL findings. Zero MEDIUM findings. The mechanical
half of this review is enforced continuously by
`TestCON_015_ArgusSecurityReviewChecklist` (a static grep gate over the
package source); the analytic half is recorded below.

## Checklist

1. **No non-crypto RNG in the signing path.** PASS. `sign.go` calls only
   `crypto/ed25519.Sign`, which is deterministic (RFC 8032): no `math/rand`,
   no `crypto/rand`, no seed/nonce parameter appears anywhere in `Sign` or
   its callers. `math/rand` appears nowhere in the package's own source
   (statically enforced). (`crypto/rand` and `math/rand` show up only as
   transitive dependencies of `crypto/ed25519` and `math/big`, never
   imported or called by this package.)

2. **No private key material logged or returned in error strings.** PASS.
   The only formatted error in the package, `paramset.go`'s
   `ErrParamSetNotAllowlisted`, formats a `param_set_id` (a public u16),
   never key bytes. `Sign` returns only the 64-octet public signature; it
   returns no error and logs nothing. No `fmt`/`log` call in the package
   takes a private key, seed, or `ed25519.PrivateKey` as an argument
   (statically enforced by the gate's `%v`/`%+v` + key-identifier grep).

3. **Short-circuit introduces no new secret-dependent branch beyond the 5
   specified checks.** PASS. `Verify` branches only on `checkPublicKey`,
   `checkR`, `checkScalarS` (Steps 1-5) and then the single stdlib
   delegation (Steps 6-7). A, R and S are all PUBLIC values (the public key
   and the signature); none of the five checks branches on any secret. The
   short-circuit's purpose is exactly to make an early-step rejection
   indistinguishable in later-step cost from a Step-7 rejection, which the
   `TestCON_015_Verify7StepShortCircuit` call-counter test confirms.

4. **Small-order table provenance citations intact.** PASS. `smallorder.go`
   cites RFC 8032 section 5.1.3 and Chalkias/Garillot/Nikolaenko "Taming the
   many EdDSAs" (2020) Table 1, and the 8 entries were independently
   cross-verified from first-principles edwards25519 arithmetic (all 8 are
   genuine torsion points: 8P == identity; the four order-8 entries also
   satisfy 4P != identity). The table's SHA-256 is pinned
   (`TestCON_015_SmallOrderTableDigestPinned`) so an accidental edit is
   caught by CI.

5. **Constants are exact and pinned.** PASS. `p = 2^255-19` (fieldPrimeLE)
   and `L = 2^252 + 27742317777372353535851937790883648493` (groupOrderL)
   are the RFC 8032 exact values, asserted against both their decimal and
   power-of-two forms in `scalar_test.go`.

6. **No curve arithmetic reimplemented.** PASS. Steps 1-5 are byte/integer
   comparisons only; Steps 6-7 delegate verbatim to unmodified stdlib
   `crypto/ed25519.Verify`. The package imports no third-party curve library.

## Downstream gate

Downstream milestones (M08 integrity trees, M09 signature & coverage, M10
attestation, M11 redaction, publish/sign/migrate) MAY depend on `pkg/eddsa`
as of this review. Any change to the package should re-run this checklist
and the static gate before that dependency is relied on again.
