// LTV time values (T-0171/T-0178/T-0180; FR-072/FR-073). SigningInstant and
// TimeInterval are the deterministic time values LTV verification reasons
// about: a signature's recorded signing instant, the interval a time
// attestation vouches for, and a credential's recorded compromise time. They
// are explicit inputs -- read from the document's own attestation evidence or
// the signature's recorded instant -- never the reader's wall clock (NFR-005:
// wall-clock time is excluded from the octet stream except at the four named
// sites, and verification-time comparisons use the ATTESTED values, not the
// ambient present).
package integrity

import "time"

// SigningInstant is a point in time as a signed count of seconds since the
// Unix epoch. It carries no wall-clock read of its own; it is always a value
// extracted from document evidence.
type SigningInstant int64

// Time converts the instant to a time.Time (UTC) for use with stdlib verify
// APIs that require a time.Time. This is a pure value conversion, not a
// wall-clock read.
func (s SigningInstant) Time() time.Time { return time.Unix(int64(s), 0).UTC() }

// Before reports whether s is strictly before other.
func (s SigningInstant) Before(other SigningInstant) bool { return s < other }

// AtOrBefore reports whether s is at or before other.
func (s SigningInstant) AtOrBefore(other SigningInstant) bool { return s <= other }

// TimeInterval is the closed interval [NotBefore, NotAfter] a time attestation
// vouches for. NotBefore MUST NOT exceed NotAfter (a well-formed interval).
type TimeInterval struct {
	NotBefore SigningInstant
	NotAfter  SigningInstant
}

// WellFormed reports whether the interval is non-inverted.
func (iv TimeInterval) WellFormed() bool { return iv.NotBefore <= iv.NotAfter }

// Contains reports whether instant lies within the closed interval.
func (iv TimeInterval) Contains(instant SigningInstant) bool {
	return instant >= iv.NotBefore && instant <= iv.NotAfter
}
