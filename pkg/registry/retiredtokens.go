package registry

import (
	"errors"
	"fmt"

	"Protodoc/pkg/ceilings"
	"Protodoc/pkg/pdlfmt"
)

// fm-retired-tokens (container.abnf S4, tag=10): the plain-seq-of(ext-token)
// set of extension tokens a document has permanently retired, never reissued
// (CON-020). The field-value is capped at a total octet length; a document
// declaring more is a structural rejection, so the retired-token list can
// never grow without bound.

// retiredTokensCeilingName is the ceiling-table entry naming the fm-
// retired-tokens field-value octet cap (currently 16384).
const retiredTokensCeilingName = "Frontmatter retired_token_list max"

// ExtTokenOctets is the fixed on-wire width of one ext-token.
const ExtTokenOctets = 8

// ErrRetiredTokensOversized is returned when a fm-retired-tokens field-value
// exceeds the ceiling-table octet cap (CON-020).
var ErrRetiredTokensOversized = errors.New("registry: fm-retired-tokens field-value exceeds the retired-token-list octet cap (CON-020)")

// MaxRetiredTokensOctets returns the fm-retired-tokens octet cap from the
// ceiling table (a single source of truth, checked against spec text by
// pkg/ceilings' own CI equality test).
func MaxRetiredTokensOctets() uint64 { return ceilings.MustMax(retiredTokensCeilingName) }

// DecodeRetiredTokens decodes a fm-retired-tokens field-value: a plain-seq-of
// (ext-token). It enforces the CON-020 octet cap on the WHOLE field-value
// before decoding elements (a bound checked before allocation), and that the
// declared count matches the octets present exactly. Returns the retired
// tokens in authored order.
func DecodeRetiredTokens(fieldValue []byte) ([]ExtToken, error) {
	if uint64(len(fieldValue)) > MaxRetiredTokensOctets() {
		return nil, fmt.Errorf("%w: %d octets exceeds cap %d", ErrRetiredTokensOversized, len(fieldValue), MaxRetiredTokensOctets())
	}
	count, n, err := pdlfmt.DecodeSeqCount(fieldValue)
	if err != nil {
		return nil, fmt.Errorf("registry: fm-retired-tokens count: %w", err)
	}
	rest := fieldValue[n:]
	if uint64(len(rest)) != count*ExtTokenOctets {
		return nil, fmt.Errorf("registry: fm-retired-tokens declares %d tokens (%d octets) but has %d element octets", count, count*ExtTokenOctets, len(rest))
	}
	tokens := make([]ExtToken, count)
	for i := uint64(0); i < count; i++ {
		copy(tokens[i][:], rest[i*ExtTokenOctets:(i+1)*ExtTokenOctets])
	}
	return tokens, nil
}

// EncodeRetiredTokens encodes tokens as a fm-retired-tokens field-value
// (plain-seq-of(ext-token)), enforcing the CON-020 octet cap so a writer
// cannot emit an over-cap list.
func EncodeRetiredTokens(tokens []ExtToken) ([]byte, error) {
	elems := make([][]byte, len(tokens))
	for i := range tokens {
		elems[i] = tokens[i][:]
	}
	value := pdlfmt.AppendPlainSeq(nil, elems)
	if uint64(len(value)) > MaxRetiredTokensOctets() {
		return nil, fmt.Errorf("%w: %d octets exceeds cap %d", ErrRetiredTokensOversized, len(value), MaxRetiredTokensOctets())
	}
	return value, nil
}
