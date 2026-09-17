package registry

import (
	"errors"
	"fmt"
)

// Retired-token reissue prevention (T-0060, CON-020; container.abnf S4/S5.2).
// A retired extension token is NEVER reissued: reusing a retired token in one
// major encoding would produce silent data corruption rather than an error.
// A token is retired if it is either (a) listed in the document's
// fm-retired-tokens set, or (b) in the permanently-retired owner-id namespace
// (owner-id 0xFFFFFFFF), which is only ever a tombstone, never a live
// extension. Using such a token as a live extension is rejected.

// ErrTokenRetired is returned when a token that is retired (in the retired set
// or the retired owner-id namespace) is used as a live extension.
var ErrTokenRetired = errors.New("registry: extension token is retired and must never be reissued (CON-020)")

// RetiredSet is the set of permanently-retired extension tokens for a
// document (its fm-retired-tokens). Membership is exact-octet.
type RetiredSet map[ExtToken]bool

// NewRetiredSet builds a RetiredSet from a token slice (e.g. the decoded
// fm-retired-tokens).
func NewRetiredSet(tokens []ExtToken) RetiredSet {
	s := make(RetiredSet, len(tokens))
	for _, t := range tokens {
		s[t] = true
	}
	return s
}

// IsRetired reports whether tok is retired: it is in the retired-owner-id
// namespace (0xFFFFFFFF) OR listed in this set. Both are permanent: a token
// that is retired can never be a live extension again.
func (s RetiredSet) IsRetired(tok ExtToken) bool {
	if tok.Partition() == PartitionRetired {
		return true
	}
	return s[tok]
}

// TokenRetiredError names the token rejected for being retired.
type TokenRetiredError struct {
	Tok ExtToken
}

func (e *TokenRetiredError) Error() string {
	return fmt.Sprintf("%v: ext-tok %x", ErrTokenRetired, e.Tok)
}

func (e *TokenRetiredError) Unwrap() error { return ErrTokenRetired }

// CheckNotRetired returns a *TokenRetiredError if tok is retired (per
// IsRetired), or nil if tok may be used as a live extension. This is the guard
// a writer applies before (re)issuing or using a token, and a reader/validator
// applies to any live ext-tok it encounters.
func (s RetiredSet) CheckNotRetired(tok ExtToken) error {
	if s.IsRetired(tok) {
		return &TokenRetiredError{Tok: tok}
	}
	return nil
}
