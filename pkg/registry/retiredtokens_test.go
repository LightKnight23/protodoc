package registry

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestCON_020_FmRetiredTokensSizeCap is T-0051's named conformance test for
// the CON-020 / container.abnf S4 octet cap on the fm-retired-tokens field-
// value (a plain-seq-of(ext-token)). A field-value at the cap decodes; one a
// single octet over the cap is rejected before any element allocation, so the
// permanently-retired token list can never grow without bound.
func TestCON_020_FmRetiredTokensSizeCap(t *testing.T) {
	cap := MaxRetiredTokensOctets()
	if cap == 0 {
		t.Fatal("retired-token octet cap is 0")
	}

	// Build a plain-seq whose field-value is exactly `cap` octets: choose a
	// token count so varintLen(count) + 8*count == cap where possible; if cap
	// is not exactly reachable, pad by choosing the largest count that fits
	// and asserting <= cap.
	fill := func(nTokens int) []byte {
		elems := make([][]byte, nTokens)
		for i := range elems {
			tok := NewExtToken(uint32(i%0x7FFFFFFF)+1, uint32(i))
			elems[i] = tok[:]
		}
		return pdlfmt.AppendPlainSeq(nil, elems)
	}

	// Largest count whose encoding is <= cap.
	atLimitCount := 0
	for n := 0; ; n++ {
		if uint64(len(fill(n))) > cap {
			atLimitCount = n - 1
			break
		}
	}
	atLimit := fill(atLimitCount)
	if uint64(len(atLimit)) > cap {
		t.Fatalf("at-limit corpus %d octets exceeds cap %d", len(atLimit), cap)
	}

	// (1) At-limit decodes and round-trips.
	toks, err := DecodeRetiredTokens(atLimit)
	if err != nil {
		t.Fatalf("at-limit (%d octets, %d tokens) rejected: %v", len(atLimit), atLimitCount, err)
	}
	if len(toks) != atLimitCount {
		t.Fatalf("decoded %d tokens, want %d", len(toks), atLimitCount)
	}
	reEnc, err := EncodeRetiredTokens(toks)
	if err != nil {
		t.Fatalf("re-encode at-limit: %v", err)
	}
	if uint64(len(reEnc)) > cap {
		t.Fatalf("re-encoded at-limit %d octets exceeds cap %d", len(reEnc), cap)
	}

	// (2) Over-limit is rejected as oversized, before element decode.
	over := make([]byte, cap+1)
	// Give it a plausible-looking count prefix so the rejection is the cap,
	// not a malformed count.
	copy(over, pdlfmt.AppendVarint(nil, uint64((cap+1)/ExtTokenOctets)))
	if _, err := DecodeRetiredTokens(over); !errors.Is(err, ErrRetiredTokensOversized) {
		t.Fatalf("over-limit (%d octets) decode error = %v, want ErrRetiredTokensOversized", len(over), err)
	}

	// (3) A writer cannot emit an over-cap list either.
	tooMany := make([]ExtToken, atLimitCount+64)
	for i := range tooMany {
		tooMany[i] = NewExtToken(uint32(i)+1, uint32(i))
	}
	if _, err := EncodeRetiredTokens(tooMany); !errors.Is(err, ErrRetiredTokensOversized) {
		t.Fatalf("encoding an over-cap list error = %v, want ErrRetiredTokensOversized", err)
	}

	// (4) A count/octet mismatch (declared count does not match element
	// octets present) is rejected distinctly from the cap.
	mismatch := pdlfmt.AppendVarint(nil, 3)         // claims 3 tokens...
	mismatch = append(mismatch, make([]byte, 8)...) // ...but only 1 token of octets
	if _, err := DecodeRetiredTokens(mismatch); err == nil {
		t.Fatal("count/octet mismatch was accepted, want rejection")
	}
}
