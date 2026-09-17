package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_107_DispositionRuntimeEnforcement is T-0059's named integration test
// (FR-107). It proves a reader that does not implement the ext-tok exhibits
// the single defined behaviour for each disposition (ignore -> render absent,
// degrade -> render fallback, refuse -> refuse document), never a silent
// omit/approximate/render-around, and that the ordered gates (payload digest,
// disposition validity, fallback presence) reject before any action is taken.
func TestFR_107_DispositionRuntimeEnforcement(t *testing.T) {
	payloadRef := fixtureUnitID(0x50)
	trueDigest := fixtureDigest(0x60)
	fallback := fixtureUnitID(0x70)
	resolve := func(ref pdlfmt.UnitID) (pdlfmt.Digest256, bool) {
		if ref == payloadRef {
			return trueDigest, true
		}
		return pdlfmt.Digest256{}, false
	}

	base := ExtEnvelope{
		Tok:           NewExtToken(0x00000107, 1),
		PayloadRef:    payloadRef,
		PayloadDigest: trueDigest,
	}

	// (1) Each disposition yields exactly its single defined action.
	cases := []struct {
		disp     ExtDisposition
		fallback pdlfmt.UnitID
		want     RenderAction
	}{
		{DispositionIgnore, fallback, ActionRenderAbsent},
		{DispositionDegrade, fallback, ActionRenderFallback},
		{DispositionRefuse, Zero16, ActionRefuseDocument},
	}
	for _, c := range cases {
		e := base
		e.Disposition = c.disp
		e.FallbackRef = c.fallback
		got, err := EnforceDisposition(e, resolve)
		if err != nil {
			t.Errorf("%v: enforcement errored: %v", c.disp, err)
			continue
		}
		if got != c.want {
			t.Errorf("%v: action = %v, want %v", c.disp, got, c.want)
		}
	}

	// A degrade action renders the declared fallback target.
	deg := base
	deg.Disposition = DispositionDegrade
	deg.FallbackRef = fallback
	if EnforcedFallbackTarget(deg) != fallback {
		t.Error("degrade must render the declared fallback target")
	}

	// (2) The payload-digest gate runs FIRST: a mismatched digest is rejected
	// regardless of a valid disposition, and NO action is produced.
	badDigest := base
	badDigest.Disposition = DispositionIgnore
	badDigest.FallbackRef = fallback
	badDigest.PayloadDigest = fixtureDigest(0x99)
	if _, err := EnforceDisposition(badDigest, resolve); !errors.Is(err, ErrExtPayloadDigestMismatch) {
		t.Errorf("digest gate: err = %v, want ErrExtPayloadDigestMismatch", err)
	}

	// (3) An out-of-range disposition is rejected (never defaulted to an
	// action), after the digest gate passes.
	badDisp := base
	badDisp.Disposition = ExtDisposition(0x7F)
	badDisp.FallbackRef = fallback
	if _, err := EnforceDisposition(badDisp, resolve); !errors.Is(err, ErrExtDispositionOutOfRange) {
		t.Errorf("disposition gate: err = %v, want ErrExtDispositionOutOfRange", err)
	}

	// (4) ignore/degrade with a missing fallback is rejected -- the reader
	// must not silently render-around a construct it cannot degrade.
	noFallback := base
	noFallback.Disposition = DispositionDegrade
	noFallback.FallbackRef = Zero16
	if _, err := EnforceDisposition(noFallback, resolve); !errors.Is(err, ErrExtFallbackMissing) {
		t.Errorf("fallback gate: err = %v, want ErrExtFallbackMissing", err)
	}
}
