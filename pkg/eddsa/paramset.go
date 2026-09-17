// Package eddsa implements EdDSA-Protodoc-1: the closed 7-step Ed25519
// signature-verification procedure (integrity.abnf S6) wrapped around the
// UNMODIFIED Go standard library crypto/ed25519 verifier, plus the v1
// signature parameter-set allowlist (integrity.abnf S10.2) and a
// deterministic signing wrapper (its determinism requirement is claimed by
// a later task). No curve arithmetic is
// reimplemented here: Steps 1-5 are cheap curve-arithmetic-free integer and
// byte checks that pre-filter exactly the inputs on which EdDSA
// implementations are documented to diverge, and Steps 6-7 delegate
// verbatim to stdlib crypto/ed25519.Verify.
//
// This file implements the parameter-set allowlist (T-0102, CON-015).
package eddsa

import (
	"errors"
	"fmt"
)

// ParamSet is a signature parameter-set identifier (integrity.abnf S10.2
// param_set_id, a u16). It selects the closed set of cryptographic
// primitives a SIGNATURE record uses.
type ParamSet uint16

const (
	// ParamSetReserved (0x0000) is never a valid selection; it marks an
	// invalid or absent allowlist choice (integrity.abnf S10.2).
	ParamSetReserved ParamSet = 0x0000

	// ParamSetV1 (0x0001) is the entire v1 allowlist: Ed25519 per RFC 8032,
	// verified via EdDSA-Protodoc-1 (integrity.abnf S6), paired with SHA-256
	// for signed_object and structure_digest (S3). It is the only value a
	// v1 (format-major 1) document may carry (integrity.abnf S10.2).
	ParamSetV1 ParamSet = 0x0001

	// Values 0x0002-0xFFFF are reserved for future major-version allowlist
	// entries only (integrity.abnf S10.2), reachable exclusively via
	// RESCIND_RESIGN at a major-version migration boundary, never mid-
	// version, so they are rejected by the current major version's Validate.
)

// ErrParamSetNotAllowlisted is returned by Validate for any parameter set
// outside the current major version's closed allowlist. It is a hard
// rejection, never a warning-level accept (CON-015).
var ErrParamSetNotAllowlisted = errors.New("eddsa: param_set_id is not on the v1 allowlist (only 0x0001 is valid; integrity.abnf S10.2)")

// Validate reports whether p is on the current major version's closed
// allowlist. For format-major 1 the sole allowlisted value is ParamSetV1
// (0x0001); ParamSetReserved (0x0000) and every reserved value
// (0x0002-0xFFFF) are rejected with a named error naming the offending
// value. There is no warning-level or lenient accept path.
func (p ParamSet) Validate() error {
	if p != ParamSetV1 {
		return fmt.Errorf("%w: got 0x%04x", ErrParamSetNotAllowlisted, uint16(p))
	}
	return nil
}

// IsAllowlisted reports whether p passes Validate, as a convenience for
// call sites that want a boolean rather than an error.
func (p ParamSet) IsAllowlisted() bool { return p.Validate() == nil }
