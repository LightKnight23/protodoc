// Self-digesting commit-ring record validation (contracts/container.abnf
// S3 record-digest NORMATIVE comment: "Verified before any other
// ring-slot field is trusted").
package container

import (
	"bytes"
	"crypto/sha256"
)

// VerifyRingSlotDigest reports whether slot's stored record-digest
// (octets [480,512) of one 512-octet ring slot) matches SHA-256
// recomputed over that same slot's own preceding octets [0,480). It
// reads nothing outside slot[:RingSlotSize]: a torn or partial write to
// this slot is detectable without consulting any other slot or any
// other region of the file (FR-117). A slot shorter than RingSlotSize
// cannot be a complete slot and is reported invalid.
//
// A slot that has never been written is all-zero; SHA-256 of 480 zero
// octets is not itself all-zero, so an all-zero slot's stored digest
// (also all-zero) fails to verify and is unambiguously distinguishable
// from a genuinely written slot (contracts/container.abnf S3 comment).
func VerifyRingSlotDigest(slot []byte) bool {
	if len(slot) < RingSlotSize {
		return false
	}
	slot = slot[:RingSlotSize]
	got := sha256.Sum256(slot[:roRecordDigest])
	return bytes.Equal(got[:], slot[roRecordDigest:RingSlotSize])
}
