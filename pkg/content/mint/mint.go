// Package content implements Protodoc's identity and anchoring layer: the
// per-run durable content-unit identity that compresses to run granularity
// (plan.md "Identity & Anchor", data-model.md S2.7/S2.8, CQ-001/CQ-002).
//
// This file implements the identity-minting primitive (T-0067, FR-019/
// FR-021/FR-023): a 128-bit token minted from crypto/rand for every content
// unit or run. Minting is the sole permitted non-deterministic site #1 in
// the CQ-004 allowlist; its construction carries NO value derived from actor
// identity, device identity, wall-clock time or editing session (FR-023,
// DP-003, and the determinism principle's ambient-value exclusion), so the
// identity-stripping publish operation cannot detach an anchor by removing a
// recoverable actor/session component. No code path re-mints or reissues a
// previously returned value (FR-021).
package mint

import (
	"crypto/rand"
	"io"

	"Protodoc/pkg/pdlfmt"
)

// randReader is the entropy source, crypto/rand.Reader in production. It is
// a package var solely so a test can substitute a deterministic reader to
// exercise error handling; production code never reassigns it, and it is
// never seeded from actor/device/clock/session state (FR-023).
var randReader io.Reader = rand.Reader

// CollisionBoundMintsPerLineage is the mint count below which the collision
// probability bound holds: 2^34 (plan.md Section 5 row 22 / the Identity
// layer decision). Below this many mints in one lineage, P(any collision)
// < 2^-60 for a 128-bit uniform token (birthday bound: m^2 / 2^129 with
// m = 2^34 gives 2^68 / 2^129 = 2^-61 < 2^-60).
const CollisionBoundMintsPerLineage = uint64(1) << 34

// MintID returns a freshly minted 128-bit content-unit identity sourced
// from crypto/rand. Its signature accepts NO actor, device, clock or
// session argument (FR-023): identity is pure entropy, nothing more. Every
// call returns an independent draw; the function holds no state and reissues
// nothing (FR-021). It returns an error only if the entropy source fails,
// which a caller must treat as fatal rather than falling back to a
// lower-entropy or deterministic source.
func MintID() (pdlfmt.UnitID, error) {
	var id pdlfmt.UnitID
	if _, err := io.ReadFull(randReader, id[:]); err != nil {
		return pdlfmt.UnitID{}, err
	}
	return id, nil
}
