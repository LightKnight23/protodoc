// T_C content-commitment / redaction tree (integrity.abnf S2.2): the primary
// signed structural commitment. T_C_root is signed directly inside
// signed_object, never routed through T_S. Leaves commit to individual
// content-model records' stored frame bytes VERBATIM (never a re-derived or
// re-normalised copy); redactable leaves carry a per-subtree CSPRNG salt so
// removing the subtree destroys the salt with it, making post-redaction
// search cost the full 2^256 preimage space.
//
// This file implements the T_C redactable leaf (0x02, T-0132) and
// non-redactable leaf (0x07, T-0133) encodings.
package integrity

import "crypto/sha256"

// SaltSize is the length of a T_C redactable-leaf salt (32 octets, reusing
// the digest width for its size; integrity.abnf S2.2 tc-salt).
const SaltSize = 32

// TCLeafRedactable returns the redactable-leaf digest
// SHA-256(0x02 || salt || frame), where frame is the content-model record's
// stored PDL-TLV frame bytes taken byte-for-byte from storage (no re-encode)
// and salt is the 32-octet CSPRNG value minted for this subtree at signing
// time (integrity.abnf S2.2 t-c-leaf-redactable). Changing any octet of
// frame or salt changes the digest.
func TCLeafRedactable(salt [SaltSize]byte, frame []byte) Digest {
	h := sha256.New()
	h.Write([]byte{DomainTCLeafRedactable})
	h.Write(salt[:])
	h.Write(frame)
	var d Digest
	copy(d[:], h.Sum(nil))
	return d
}
