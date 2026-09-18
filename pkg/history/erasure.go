// ErasureRecord (T-0217/T-0218, FR-061; integrity.abnf S7.2). A document must
// enumerate the identifier and a salted-commitment digest of every previously
// published state it can no longer reconstruct (FR-061). An ERASURE_RECORD
// (discriminant 0x0D, carried in a HISTORY-typed segment) names the severed
// unit/state's own original identity, its content digest before severance, and
// the severance commitment SHA-256(0x09 || salt || digest). The salt is a
// 32-octet CSPRNG value DESTROYED at the moment of trim: present only briefly
// between minting and trim, absent/all-zero afterwards, so the commitment is
// hiding (2^80 floor per the CQ-015 salted-commitment ruling, shared with the
// redaction commitment).
package history

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/integrity"
	"Protodoc/pkg/pdlfmt"
)

// ERASURE_RECORD discriminant + TLV tags (integrity.abnf S7.2).
const (
	erasureRecordDiscriminant = 0x0D
	erTagDiscriminant         = 0
	erTagIdentity             = 1 // er-identity, unit-id (the severed state's own id)
	erTagDigest               = 2 // er-digest, digest256 (content digest before severance)
	erTagCommitment           = 3 // er-commitment, digest256 = SHA-256(0x09 || salt || digest)
	erTagSalt                 = 4 // er-salt, digest256 (destroyed at trim; zero after)
)

// SaltSize is the ErasureRecord salt width (32 octets), matching the redaction
// commitment salt.
const SaltSize = 32

// ErasureRecord is the decoded ERASURE_RECORD (integrity.abnf S7.2).
type ErasureRecord struct {
	Identity   pdlfmt.UnitID    // the severed unit/state's own original identity
	Digest     pdlfmt.Digest256 // the severed content's digest before severance
	Commitment pdlfmt.Digest256 // = SeveranceCommitment(salt, Digest)
	Salt       [SaltSize]byte   // destroyed (all-zero) after trim
}

var zeroSalt [SaltSize]byte

// SeveranceCommitment computes the FR-061 salted commitment
// SHA-256(0x09 || salt || digest), reusing integrity's severance-commitment
// domain tag (0x09). It binds the severed content's digest under a hiding
// salt; once the salt is destroyed at trim, the commitment reveals nothing
// about the digest without an exhaustive search over the salt (>= 2^80 floor).
func SeveranceCommitment(salt [SaltSize]byte, digest pdlfmt.Digest256) pdlfmt.Digest256 {
	h := sha256.New()
	h.Write([]byte{integrity.DomainSeveranceCommitment})
	h.Write(salt[:])
	h.Write(digest[:])
	var out pdlfmt.Digest256
	copy(out[:], h.Sum(nil))
	return out
}

// NewErasureRecord builds an ErasureRecord with the salt still present (the
// brief window before trim), computing its commitment from the salt and digest.
func NewErasureRecord(identity pdlfmt.UnitID, digest pdlfmt.Digest256, salt [SaltSize]byte) ErasureRecord {
	return ErasureRecord{
		Identity:   identity,
		Digest:     digest,
		Commitment: SeveranceCommitment(salt, digest),
		Salt:       salt,
	}
}

// Trimmed returns a copy of the record with the salt DESTROYED (zeroed), as it
// stands after the trim: the identity, digest, and commitment are retained,
// the salt is gone. This is the record's persistent form (FR-061).
func (r ErasureRecord) Trimmed() ErasureRecord {
	t := r
	t.Salt = zeroSalt
	return t
}

// SaltDestroyed reports whether the salt has been destroyed (all-zero).
func (r ErasureRecord) SaltDestroyed() bool { return r.Salt == zeroSalt }

var (
	// ErrErasureDiscriminant is returned for a wrong er-discriminant.
	ErrErasureDiscriminant = errors.New("history: record discriminant is not ERASURE_RECORD (0x0D)")
	// ErrErasureMissingField is returned for a missing required field.
	ErrErasureMissingField = errors.New("history: ERASURE_RECORD missing a required field")
)

// SeveredState describes a previously-published state that a retention advance
// renders unreconstructable: its identity, its content digest as it stood, and
// its segment ordinal.
type SeveredState struct {
	Identity pdlfmt.UnitID
	Digest   pdlfmt.Digest256
	Ordinal  uint16
	Salt     [SaltSize]byte
}

// EnumerateErasuresOnRetentionAdvance returns an ErasureRecord for every
// previously-published state that becomes unreconstructable when the retention
// point advances from oldPoint to newPoint (newPoint > oldPoint): a state whose
// ordinal is in [oldPoint, newPoint) was reconstructable before and is not
// after, so FR-061 requires it be enumerated with its identity and salted
// commitment. Each returned record is TRIMMED (salt destroyed), the persistent
// form. States outside that window are not enumerated. It errors if newPoint
// does not advance past oldPoint.
func EnumerateErasuresOnRetentionAdvance(oldPoint, newPoint uint16, severed []SeveredState) ([]ErasureRecord, error) {
	if newPoint <= oldPoint {
		return nil, ErrRetentionNotAdvanced
	}
	var out []ErasureRecord
	for _, s := range severed {
		if s.Ordinal >= oldPoint && s.Ordinal < newPoint {
			out = append(out, NewErasureRecord(s.Identity, s.Digest, s.Salt).Trimmed())
		}
	}
	return out, nil
}

// ErrRetentionNotAdvanced is returned when a retention advance does not move
// the point forward.
var ErrRetentionNotAdvanced = errors.New("history: retention point did not advance")

// Encode encodes the ErasureRecord to its byte-exact PDL-TLV wire form
// (fields 0..4 in ascending tag order).
func (r ErasureRecord) Encode() ([]byte, error) {
	fields := []pdlfmt.Field{
		{Tag: erTagDiscriminant, Value: []byte{erasureRecordDiscriminant}},
		{Tag: erTagIdentity, Value: pdlfmt.AppendUnitID(nil, r.Identity)},
		{Tag: erTagDigest, Value: pdlfmt.AppendDigest256(nil, r.Digest)},
		{Tag: erTagCommitment, Value: pdlfmt.AppendDigest256(nil, r.Commitment)},
		{Tag: erTagSalt, Value: append([]byte(nil), r.Salt[:]...)},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeErasureRecord decodes an ERASURE_RECORD, the byte-exact inverse of
// Encode. It rejects a wrong discriminant, a missing field, and a wrong field
// width.
func DecodeErasureRecord(src []byte) (ErasureRecord, error) {
	var r ErasureRecord
	known := map[byte]bool{erTagDiscriminant: true, erTagIdentity: true, erTagDigest: true, erTagCommitment: true, erTagSalt: true}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return r, err
	}
	seen := map[byte]bool{}
	for _, f := range fields {
		seen[f.Tag] = true
		switch f.Tag {
		case erTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != erasureRecordDiscriminant {
				return r, ErrErasureDiscriminant
			}
		case erTagIdentity:
			id, _, err := pdlfmt.DecodeUnitID(f.Value)
			if err != nil || len(f.Value) != 16 {
				return r, fmt.Errorf("history: er-identity unit-id: %v", err)
			}
			r.Identity = id
		case erTagDigest:
			d, _, err := pdlfmt.DecodeDigest256(f.Value)
			if err != nil || len(f.Value) != 32 {
				return r, fmt.Errorf("history: er-digest: %v", err)
			}
			r.Digest = d
		case erTagCommitment:
			d, _, err := pdlfmt.DecodeDigest256(f.Value)
			if err != nil || len(f.Value) != 32 {
				return r, fmt.Errorf("history: er-commitment: %v", err)
			}
			r.Commitment = d
		case erTagSalt:
			if len(f.Value) != SaltSize {
				return r, fmt.Errorf("history: er-salt is %d octets, want %d", len(f.Value), SaltSize)
			}
			copy(r.Salt[:], f.Value)
		}
	}
	for tag := byte(erTagDiscriminant); tag <= erTagSalt; tag++ {
		if !seen[tag] {
			return r, fmt.Errorf("%w: tag %d", ErrErasureMissingField, tag)
		}
	}
	return r, nil
}
