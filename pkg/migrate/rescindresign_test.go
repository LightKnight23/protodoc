package migrate_test

import (
	"crypto/ed25519"
	"reflect"
	"testing"

	"Protodoc/pkg/eddsa"
	"Protodoc/pkg/migrate"
	"Protodoc/pkg/pdlfmt"
)

// TestCON_016_RescindAndResignAtMajorVersionBoundary is T-0304's named
// integration test (CON-016 / DP-017). A signature under a simulated
// broken scheme, migrated with rescind-and-resign at a major-version boundary,
// produces a fresh signature that verifies Valid under the current parameter
// set; the RescindResignRecord has exactly the five data-model.md 2.18 fields;
// the prior signature is retained (queryable); invoking outside a boundary is
// refused.
func TestCON_016_RescindAndResignAtMajorVersionBoundary(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	var priorSig pdlfmt.UnitID
	priorSig[0] = 0x9
	var newSignedObject, migState pdlfmt.Digest256
	newSignedObject[0] = 0x5
	migState[0] = 0x6

	// The injected signer produces a real Ed25519 signature over the migrated
	// signed_object under the (allowlisted) new parameter set.
	sign := func(paramSet uint16, so pdlfmt.Digest256) [64]byte {
		return eddsa.Sign(priv, [32]byte(so))
	}

	// Refused outside a major-version boundary.
	if _, err := migrate.RescindAndResign(false, priorSig, 1, newSignedObject, migState, sign); err != migrate.ErrNotAtMajorBoundary {
		t.Errorf("outside boundary: err = %v, want ErrNotAtMajorBoundary", err)
	}

	// At a major-version boundary: succeeds.
	rec, err := migrate.RescindAndResign(true, priorSig, 1, newSignedObject, migState, sign)
	if err != nil {
		t.Fatalf("RescindAndResign at boundary: %v", err)
	}

	// The fresh signature verifies Valid under the current parameter set.
	if !ed25519.Verify(pub, newSignedObject[:], rec.NewSignatureValue[:]) {
		t.Errorf("fresh signature must verify Valid under the current parameter set")
	}

	// The prior signature ref is retained (queryable), not erased.
	if rec.PriorSignatureRef != priorSig {
		t.Errorf("prior signature ref must be retained, got %x", rec.PriorSignatureRef[0])
	}
	// The migration state id and param set are recorded.
	if rec.MigrationStateID != migState || rec.NewParamSetID != 1 || rec.NewSignedObject != newSignedObject {
		t.Errorf("record fields not populated as expected: %+v", rec)
	}

	// Exactly five fields, matching data-model.md 2.18, no additional field.
	if n := reflect.TypeOf(rec).NumField(); n != 5 {
		t.Errorf("RescindResignRecord has %d fields, want exactly 5 (data-model.md 2.18)", n)
	}

	// A non-allowlisted parameter set is refused even at a boundary.
	if _, err := migrate.RescindAndResign(true, priorSig, 99, newSignedObject, migState, sign); err != migrate.ErrParamSetNotAllowlisted {
		t.Errorf("non-allowlisted param set: err = %v, want ErrParamSetNotAllowlisted", err)
	}
}
