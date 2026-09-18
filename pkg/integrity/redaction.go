// Redaction (T-0185..T-0206, FR-074..081; integrity.abnf S2.2/S7.1). WHEN a
// document is signed, the signatory may designate content subtrees as
// redactable, committing each with a hiding-and-binding commitment whose salt
// is stored INSIDE the designated subtree. Later removal (the act of
// redaction) deletes BOTH the subtree's octets AND its salt in one operation,
// leaving only the bare 32-octet commitment digest -- no salt, no plaintext,
// anywhere in the file. This makes redaction PROVABLY TOTAL: the retained
// values cannot recover the omitted content by exhaustive search over any
// candidate set below 2^80 (FR-075; the actual post-removal search is 2^256).
package integrity

import (
	"bytes"
	"errors"
)

// RedactableSubtree is a content subtree designated redactable at signing time
// (FR-074). While present, it holds its own stored frame octets and its
// 32-octet CSPRNG salt (the salt is stored INSIDE the subtree, so removing the
// subtree removes the salt with it). Its commitment is the salted redactable
// leaf digest.
type RedactableSubtree struct {
	// UnitID identifies the subtree (its content-model record boundary; a
	// redaction never splits a record, RED-ALIGN).
	UnitID UnitID
	// Frame is the subtree's stored PDL-TLV frame octets, verbatim.
	Frame []byte
	// Salt is the per-subtree CSPRNG salt, stored inside the subtree.
	Salt [SaltSize]byte
	// Designated marks the subtree as redactable at signing time. A subtree
	// not designated at signing has no salt commitment and cannot be omitted
	// under a declared-omission verdict (FR-077).
	Designated bool
	// removed is set once Remove has deleted the frame and salt.
	removed bool
	// commitment is the retained bare commitment digest after Remove.
	commitment Digest
}

var (
	// ErrNotDesignated is returned when an operation requires a subtree
	// designated redactable at signing time and it was not.
	ErrNotDesignated = errors.New("integrity: subtree was not designated redactable at signing time")
	// ErrAlreadyRemoved is returned when Remove is called on an
	// already-removed subtree.
	ErrAlreadyRemoved = errors.New("integrity: redactable subtree already removed")
)

// DesignateRedactable builds a RedactableSubtree designated at signing time
// with the given frame and salt. The salt is stored inside the subtree
// (FR-074): it lives with the frame and is deleted together with it on
// removal.
func DesignateRedactable(id UnitID, frame []byte, salt [SaltSize]byte) RedactableSubtree {
	return RedactableSubtree{
		UnitID:     id,
		Frame:      append([]byte(nil), frame...),
		Salt:       salt,
		Designated: true,
	}
}

// IsDesignated reports whether the subtree was designated redactable at
// signing time.
func (s RedactableSubtree) IsDesignated() bool { return s.Designated }

// Commitment returns the subtree's hiding-and-binding commitment: the salted
// redactable-leaf digest SHA-256(0x02 || salt || frame). While present it is
// computed from the live frame and salt; after Remove it is the retained bare
// commitment digest. It errors if the subtree was never designated.
func (s RedactableSubtree) Commitment() (Digest, error) {
	if !s.Designated {
		return Digest{}, ErrNotDesignated
	}
	if s.removed {
		return s.commitment, nil
	}
	return TCLeafRedactable(s.Salt, s.Frame), nil
}

// SaltStoredInside reports whether the salt currently lives inside the subtree
// (true while present, false after removal). This is the property FR-074
// requires: the salt is co-located with the subtree it commits, so removing
// the subtree removes the salt.
func (s RedactableSubtree) SaltStoredInside() bool { return !s.removed }

// Remove performs the act of redaction: it deletes BOTH the subtree's frame
// octets AND its salt in the same operation, retaining only the bare 32-octet
// commitment digest (integrity.abnf S7.1). After Remove, the subtree holds no
// plaintext and no salt anywhere; the commitment digest is unchanged, so the
// signed T_C_root that committed to it still verifies (FR-076). Remove is
// idempotent-safe only in that a second call errors rather than corrupting the
// retained commitment. It errors if the subtree was never designated.
func (s *RedactableSubtree) Remove() error {
	if !s.Designated {
		return ErrNotDesignated
	}
	if s.removed {
		return ErrAlreadyRemoved
	}
	// Compute the commitment from the live frame+salt BEFORE erasing them.
	s.commitment = TCLeafRedactable(s.Salt, s.Frame)
	// Erase both, together, in this one operation.
	for i := range s.Frame {
		s.Frame[i] = 0
	}
	s.Frame = nil
	s.Salt = zeroSalt
	s.removed = true
	return nil
}

// Removed reports whether this subtree has been redacted (frame and salt
// erased, only the bare commitment retained).
func (s RedactableSubtree) Removed() bool { return s.removed }

// UndesignatedOmission reports whether the current record set omits content
// that was present and NOT designated redactable at signing time (FR-077). A
// signed document may lawfully omit only its DESIGNATED redactable subtrees
// (which are redacted in place, preserving the T_C root); omitting anything
// else -- a non-redactable record, or a redactable record removed without
// leaving its retained commitment -- changes the T_C root and is an
// UNDESIGNATED omission. signedUnits is the set of unit-ids present at signing;
// current is the record set now. Returns the unit-ids omitted without
// designation (empty if every omission was a declared redaction).
func UndesignatedOmission(signedUnits map[UnitID]bool, signedRedactable map[UnitID]bool, current []ContentRecord) []UnitID {
	present := make(map[UnitID]bool, len(current))
	for _, r := range current {
		// A redacted record is still "present" as its retained commitment
		// leaf (a declared omission), so it does not count as omitted.
		present[r.UnitID] = true
	}
	var undesignated []UnitID
	for id := range signedUnits {
		if !present[id] && !signedRedactable[id] {
			// Missing entirely AND not designated redactable at signing.
			undesignated = append(undesignated, id)
		}
	}
	return undesignated
}

// UndesignatedOmissionVerdict returns VerdictUnverified when the current record
// set contains an undesignated omission relative to the signed state (FR-077),
// or VerdictValid when every omission was a declared redaction. An undeclared
// omission changes the T_C root, so the signature cannot verify -- the reader
// must present the document as unverified, never valid or
// attested-with-declared-omissions.
func UndesignatedOmissionVerdict(signedUnits, signedRedactable map[UnitID]bool, current []ContentRecord) Verdict {
	if len(UndesignatedOmission(signedUnits, signedRedactable, current)) > 0 {
		return VerdictUnverified
	}
	return VerdictValid
}

// RedactRecord converts a redactable ContentRecord into its redacted form: it
// computes the retained commitment leaf (the redactable-leaf digest over the
// live salt+frame), then returns a record with Redacted set, the retained leaf
// carried, and the frame and salt cleared. The redacted record's leaf digest
// equals the pre-redaction redactable-leaf digest, so the T_C root -- and thus
// any signature over it -- is PRESERVED across redaction (FR-076). It errors
// if the record was not designated redactable.
func RedactRecord(rec ContentRecord) (ContentRecord, error) {
	if !rec.Redactable {
		return ContentRecord{}, ErrNotDesignated
	}
	retained := TCLeafRedactable(rec.Salt, rec.Frame)
	return ContentRecord{
		UnitID:       rec.UnitID,
		Redactable:   true,
		Redacted:     true,
		RetainedLeaf: retained,
		// Frame and Salt deliberately left zero: the act of redaction removes
		// both from the redacted document.
	}, nil
}

// zeroSalt is the all-zero salt, used to prove the salt is gone after removal.
var zeroSalt [SaltSize]byte

// SaltPresent reports whether any non-zero salt remains after an operation
// (used by tests to prove the salt was actually erased on removal).
func (s RedactableSubtree) SaltPresent() bool {
	return !s.removed && !bytes.Equal(s.Salt[:], zeroSalt[:])
}
