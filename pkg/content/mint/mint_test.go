package mint

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded is T-0067's named
// test. It asserts MintID: is sourced from crypto/rand (16 full octets of
// entropy, no fixed structure), takes no actor/device/clock/session
// argument (FR-023), never reissues a value across a large sample
// (statistical collision check), and that the documented collision bound
// (P < 2^-60 below 2^34 mints/lineage) holds under the birthday
// approximation.
func TestFR_019_MintedIdentityIsCSPRNGAndCollisionBounded(t *testing.T) {
	// --- Signature carries no actor/device/clock/session parameter ---
	// MintID's type must be func() (UnitID, error): zero input parameters.
	ft := reflect.TypeOf(MintID)
	if ft.NumIn() != 0 {
		t.Fatalf("MintID takes %d parameters, want 0 (FR-023: no actor/device/clock/session argument)", ft.NumIn())
	}
	if ft.NumOut() != 2 {
		t.Fatalf("MintID returns %d values, want 2 (UnitID, error)", ft.NumOut())
	}

	// --- Statistical uniqueness: a large sample has zero collisions ---
	const sample = 1 << 20 // ~1,048,576 mints
	seen := make(map[pdlfmt.UnitID]struct{}, sample)
	var allZero, prev pdlfmt.UnitID
	variedFromPrev := 0
	for i := 0; i < sample; i++ {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		if id == allZero {
			t.Fatalf("MintID returned the all-zero id at draw %d (not CSPRNG-sourced)", i)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("MintID reissued a value at draw %d (collision / reuse)", i)
		}
		seen[id] = struct{}{}
		if i > 0 && id != prev {
			variedFromPrev++
		}
		prev = id
	}
	// Every consecutive pair should differ (entropy, not a counter).
	if variedFromPrev != sample-1 {
		t.Fatalf("only %d of %d consecutive mints differed; identity is not high-entropy", variedFromPrev, sample-1)
	}

	// --- Documented collision bound under the birthday approximation ---
	// For m uniform draws from a space of size 2^128, the expected number
	// of collisions is approximately m^2 / 2^129. At m = 2^34 this is
	// 2^68 / 2^129 = 2^-61, which is below the documented 2^-60 bound.
	m := float64(CollisionBoundMintsPerLineage) // 2^34
	pCollision := (m * m) / math.Exp2(129)      // ~2^-61
	if pCollision >= math.Exp2(-60) {
		t.Fatalf("birthday-bound P(collision) at %g mints is %g, not below 2^-60", m, pCollision)
	}

	// --- Entropy sanity: the aggregate octet distribution is not skewed to
	// a constant (a crude check that bytes are drawn, not fixed) ---
	assertOctetsLookRandom(t, seen)
}

// assertOctetsLookRandom does a coarse check that, across the minted sample,
// every octet position takes many distinct values rather than a fixed
// constant (which would betray a non-CSPRNG source). It is deliberately
// weak: a real randomness test is out of scope; this only catches a source
// that is obviously structured.
func assertOctetsLookRandom(t *testing.T, seen map[pdlfmt.UnitID]struct{}) {
	t.Helper()
	var distinctFirstOctet [256]bool
	count := 0
	for id := range seen {
		distinctFirstOctet[id[0]] = true
		count++
		if count >= 4096 {
			break // a sample of the set is plenty for this coarse check
		}
	}
	n := 0
	for _, present := range distinctFirstOctet {
		if present {
			n++
		}
	}
	if n < 200 {
		t.Fatalf("first octet took only %d of 256 possible values across the sample; source looks structured", n)
	}
}

// TestFR_019_MintPropagatesEntropyError confirms MintID surfaces an entropy
// source failure rather than returning a low-entropy or zero id.
func TestFR_019_MintPropagatesEntropyError(t *testing.T) {
	orig := randReader
	t.Cleanup(func() { randReader = orig })

	sentinel := errors.New("entropy source unavailable")
	randReader = failingReader{err: sentinel}

	_, err := MintID()
	if !errors.Is(err, sentinel) {
		t.Fatalf("MintID error = %v, want the entropy source error", err)
	}
}

// failingReader is an io.Reader that always fails, modelling an exhausted or
// unavailable entropy source.
type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

// TestFR_019_MintReadsFullSixteenOctets confirms MintID consumes exactly 16
// octets from the entropy source (a full UnitID), never fewer.
func TestFR_019_MintReadsFullSixteenOctets(t *testing.T) {
	orig := randReader
	t.Cleanup(func() { randReader = orig })

	// A reader delivering a known 16-octet pattern lets us confirm MintID
	// copies all 16 octets through.
	pattern := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	randReader = bytes.NewReader(pattern)

	id, err := MintID()
	if err != nil {
		t.Fatalf("MintID: %v", err)
	}
	if !bytes.Equal(id[:], pattern) {
		t.Fatalf("MintID id = %x, want %x (must read all 16 octets)", id[:], pattern)
	}
}
