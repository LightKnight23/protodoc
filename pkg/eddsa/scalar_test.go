package eddsa

import (
	"errors"
	"math/big"
	"testing"
)

// bigToLE encodes n as 32 little-endian octets (n must fit in 256 bits).
func bigToLE(n *big.Int) [32]byte {
	var out [32]byte
	be := n.Bytes() // big-endian, minimal
	for i := 0; i < len(be); i++ {
		out[i] = be[len(be)-1-i]
	}
	return out
}

// TestCON_015_Step5_ScalarRangeCheck is T-0107's named test. checkScalarS
// rejects S==L, S==L+1 and S==2^256-1, and accepts S==L-1 and S==0, using
// exact integer comparison (no float).
func TestCON_015_Step5_ScalarRangeCheck(t *testing.T) {
	L := new(big.Int).Set(groupOrderL)
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	reject := map[string][32]byte{
		"S==L":       bigToLE(L),
		"S==L+1":     bigToLE(new(big.Int).Add(L, big.NewInt(1))),
		"S==2^256-1": bigToLE(max256),
	}
	for name, S := range reject {
		if err := checkScalarS(S); !errors.Is(err, ErrScalarOutOfRange) {
			t.Fatalf("%s: got %v, want ErrScalarOutOfRange", name, err)
		}
	}

	accept := map[string][32]byte{
		"S==L-1": bigToLE(new(big.Int).Sub(L, big.NewInt(1))),
		"S==0":   bigToLE(big.NewInt(0)),
	}
	for name, S := range accept {
		if err := checkScalarS(S); err != nil {
			t.Fatalf("%s: got %v, want accept", name, err)
		}
	}

	// Sanity: L is the documented exact value.
	wantL, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	if groupOrderL.Cmp(wantL) != 0 {
		t.Fatalf("groupOrderL != documented L")
	}
	// And L == 2^252 + 27742317777372353535851937790883648493.
	twoTo252 := new(big.Int).Lsh(big.NewInt(1), 252)
	addend, _ := new(big.Int).SetString("27742317777372353535851937790883648493", 10)
	if new(big.Int).Add(twoTo252, addend).Cmp(groupOrderL) != 0 {
		t.Fatalf("L != 2^252 + 27742317777372353535851937790883648493")
	}
}
