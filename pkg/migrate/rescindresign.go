// RESCIND-AND-RESIGN (CON-016 scheme-break half, DP-017; T-0304). When a
// signature's cryptographic SCHEME (not merely its hash family) is considered
// broken, an operator may — ONLY at a major-version boundary, via an explicit
// --rescind-and-resign migrate flag (not a 12th CLI verb, per TR-012) — rescind
// that signature and attach a FRESH signature under the current allowlisted
// parameter set. The rescission is recorded in a RescindResignRecord
// (data-model.md 2.18 / integrity.abnf S8) with exactly five fields so the
// rescinded signature's historical verdict stays inspectable, never erased.
package migrate

import (
	"errors"

	"Protodoc/pkg/pdlfmt"
)

// RescindResignRecord mirrors integrity.abnf S8 / data-model.md 2.18 exactly:
// five fields, no more.
type RescindResignRecord struct {
	PriorSignatureRef pdlfmt.UnitID    // rr-prior-sig-ref: the retained pre-migration signature
	NewParamSetID     uint16           // rr-new-param-set: the new allowlist entry
	NewSignedObject   pdlfmt.Digest256 // rr-new-signed-object: over the migrated state
	NewSignatureValue [64]byte         // rr-new-signature-value: R||S
	MigrationStateID  pdlfmt.Digest256 // rr-migration-state-id: the migrated document's state
}

// ErrNotAtMajorBoundary is returned when rescind-and-resign is invoked outside
// a major-version boundary.
var ErrNotAtMajorBoundary = errors.New("migrate: --rescind-and-resign is only valid at a major-version boundary")

// AllowlistedParamSets is the current set of allowlisted signature parameter
// set ids a fresh signature may use.
var AllowlistedParamSets = map[uint16]bool{1: true}

// ErrParamSetNotAllowlisted is returned when the new parameter set is not
// currently allowlisted.
var ErrParamSetNotAllowlisted = errors.New("migrate: new parameter set is not allowlisted")

// Signer produces a fresh signature value over a signed_object digest under a
// parameter set. It is injected so the mechanism does not itself embed key
// material; the CLI wires the real EdDSA signer.
type Signer func(paramSet uint16, newSignedObject pdlfmt.Digest256) [64]byte

// RescindAndResign rescinds priorSigRef and attaches a fresh signature under
// newParamSet over the migrated state's newSignedObject, returning the
// RescindResignRecord. It refuses unless atMajorBoundary is true (the only
// invocation path is the explicit operator flag) and unless newParamSet is
// allowlisted. It does NOT delete or rewrite the prior signature — priorSigRef
// remains so the rescinded signature's historical verdict stays queryable.
func RescindAndResign(
	atMajorBoundary bool,
	priorSigRef pdlfmt.UnitID,
	newParamSet uint16,
	newSignedObject pdlfmt.Digest256,
	migrationStateID pdlfmt.Digest256,
	sign Signer,
) (RescindResignRecord, error) {
	if !atMajorBoundary {
		return RescindResignRecord{}, ErrNotAtMajorBoundary
	}
	if !AllowlistedParamSets[newParamSet] {
		return RescindResignRecord{}, ErrParamSetNotAllowlisted
	}
	return RescindResignRecord{
		PriorSignatureRef: priorSigRef,
		NewParamSetID:     newParamSet,
		NewSignedObject:   newSignedObject,
		NewSignatureValue: sign(newParamSet, newSignedObject),
		MigrationStateID:  migrationStateID,
	}, nil
}
