package registry

import (
	"errors"
	"testing"
)

// TestCON_020_RetiredTokenNeverReissued is T-0060's named unit test (CON-020).
// A token that is retired -- either listed in fm-retired-tokens or in the
// permanently-retired owner-id namespace (0xFFFFFFFF) -- is never reissued:
// using it as a live extension is rejected, naming the token.
func TestCON_020_RetiredTokenNeverReissued(t *testing.T) {
	retired1 := NewExtToken(0x00000010, 0x00000001)  // registered, but retired by listing
	retired2 := NewExtToken(0x80000000, 0x00000002)  // owner-scoped, retired by listing
	live := NewExtToken(0x00000010, 0x00000099)      // same owner, different seq: still live
	tombstone := NewExtToken(OwnerIDRetired, 0x1234) // retired owner-id namespace

	set := NewRetiredSet([]ExtToken{retired1, retired2})

	// (1) A listed retired token is rejected.
	for _, tok := range []ExtToken{retired1, retired2} {
		if !set.IsRetired(tok) {
			t.Errorf("listed retired token %x reported not retired", tok)
		}
		err := set.CheckNotRetired(tok)
		if !errors.Is(err, ErrTokenRetired) {
			t.Errorf("CheckNotRetired(%x) = %v, want ErrTokenRetired", tok, err)
		}
		var te *TokenRetiredError
		if !errors.As(err, &te) || te.Tok != tok {
			t.Errorf("error should name the retired token %x, got %+v", tok, te)
		}
	}

	// (2) A token in the retired owner-id namespace is retired even if not
	// listed -- the tombstone namespace is never a live extension.
	if !set.IsRetired(tombstone) {
		t.Errorf("tombstone-namespace token %x reported not retired", tombstone)
	}
	if err := set.CheckNotRetired(tombstone); !errors.Is(err, ErrTokenRetired) {
		t.Errorf("tombstone token: err = %v, want ErrTokenRetired", err)
	}
	// And an empty set still treats the tombstone namespace as retired.
	if !(RetiredSet{}).IsRetired(tombstone) {
		t.Error("empty set must still treat the retired owner-id namespace as retired")
	}

	// (3) A live token (not listed, not in the retired namespace) passes,
	// even sharing an owner-id with a retired token: retirement is per exact
	// token, plus the whole retired namespace.
	if set.IsRetired(live) {
		t.Errorf("live token %x wrongly reported retired", live)
	}
	if err := set.CheckNotRetired(live); err != nil {
		t.Errorf("live token rejected: %v", err)
	}
}
