// signed_object (T-0152, FR-063; integrity.abnf S3.2): the value a signature
// authenticates. It is SHA-256 over a fixed 129-octet preimage,
//
//	%x04 || t-c-root || structure-digest || presentation-artefact-digest || coverage-descriptor-digest
//	  1  +    32     +        32         +             32                +              32              = 129
//
// where the leading octet is DomainSignedObject (0x04). t-c-root is the
// CURRENT T_C_root recomputed fresh at verify time; structure-digest is the
// fresh structure_digest; presentation-artefact-digest is the referenced
// PRESENTATION_ARTEFACT segment's own SegmentTableSlot.slot-digest
// (content-addressed, never separately stored); coverage-descriptor-digest is
// CoverageDescriptor.Digest() over the descriptor carried in the signature.
// None of these four inputs is ever trusted from a stored slot -- each is
// recomputed/read fresh -- so signed_object is a fact about the current file.
package integrity

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// SignedObjectPreimageLen is the fixed preimage length: 1 domain octet + four
// 32-octet digests.
const SignedObjectPreimageLen = 1 + 32 + 32 + 32 + 32 // 129

// ErrSignedObjectPreimageLen is a defensive internal error if the assembled
// preimage is not exactly 129 octets (it always is by construction).
var ErrSignedObjectPreimageLen = errors.New("integrity: signed_object preimage is not 129 octets")

// SignedObjectInput carries the four fresh digests the signed_object preimage
// commits to. Every field is recomputed/read fresh at verify time, never
// trusted from a stored slot.
type SignedObjectInput struct {
	TCRoot                     Digest // current T_C_root (S2.2)
	StructureDigest            Digest // fresh structure_digest (S3.1)
	PresentationArtefactDigest Digest // referenced PRESENTATION_ARTEFACT slot digest (S7.4)
	CoverageDescriptorDigest   Digest // CoverageDescriptor.Digest() (S4)
}

// AssembleSignedObjectPreimage builds the fixed 129-octet preimage in the
// normative field order, prefixed with DomainSignedObject (0x04).
func AssembleSignedObjectPreimage(in SignedObjectInput) []byte {
	buf := make([]byte, 0, SignedObjectPreimageLen)
	buf = append(buf, DomainSignedObject)
	buf = append(buf, in.TCRoot[:]...)
	buf = append(buf, in.StructureDigest[:]...)
	buf = append(buf, in.PresentationArtefactDigest[:]...)
	buf = append(buf, in.CoverageDescriptorDigest[:]...)
	return buf
}

// SlotDigestResolver returns the SegmentTableSlot.slot-digest of the segment
// addressed by a unit-id (e.g. a SIGNATURE's sig-presentation-ref -> a
// PRESENTATION_ARTEFACT segment), and ok=false if no such segment exists. The
// caller supplies this from the current file; this package never trusts an
// inline copy of a digest that is defined to be content-addressed.
type SlotDigestResolver func(ref UnitID) (slotDigest Digest, ok bool)

// UnitID is the 16-octet unit identifier (aliased from pdlfmt so callers of
// this package need not import pdlfmt just to name the resolver's key).
type UnitID = pdlfmt.UnitID

// ErrPresentationRefUnresolved is returned when a signature's
// sig-presentation-ref does not resolve to a present PRESENTATION_ARTEFACT
// segment.
var ErrPresentationRefUnresolved = errors.New("integrity: sig-presentation-ref does not resolve to a present PRESENTATION_ARTEFACT segment")

// SignedObjectForSignature assembles the signed_object for a SIGNATURE record,
// sourcing presentation_artefact_digest STRICTLY from the referenced
// PRESENTATION_ARTEFACT segment's own SegmentTableSlot.slot-digest (via
// resolve), never from any inline copy in the signature (integrity.abnf S3.2,
// document.abnf S7.4: the digest is content-addressed and never separately
// stored). tcRoot, structureDigest, and coverageDescriptorDigest are the other
// three fresh inputs; the coverage digest is taken from the signature's own
// carried descriptor. A signature with a zero16 presentation ref, or a ref
// that does not resolve, yields ErrPresentationRefUnresolved.
func SignedObjectForSignature(sig SignatureRecord, tcRoot, structureDigest Digest, resolve SlotDigestResolver) (Digest, error) {
	presDigest, ok := resolve(sig.PresentationRef)
	if !ok {
		return Digest{}, fmt.Errorf("%w: %x", ErrPresentationRefUnresolved, sig.PresentationRef)
	}
	covDigest, err := sig.Coverage.Digest()
	if err != nil {
		return Digest{}, fmt.Errorf("integrity: coverage digest: %w", err)
	}
	return SignedObject(SignedObjectInput{
		TCRoot:                     tcRoot,
		StructureDigest:            structureDigest,
		PresentationArtefactDigest: presDigest,
		CoverageDescriptorDigest:   covDigest,
	})
}

// SignedObject assembles the preimage and hashes it once with SHA-256 to
// produce the 32-octet signed_object (integrity.abnf S3.2), the message a
// signature's value is verified against.
func SignedObject(in SignedObjectInput) (Digest, error) {
	preimage := AssembleSignedObjectPreimage(in)
	if len(preimage) != SignedObjectPreimageLen {
		return Digest{}, ErrSignedObjectPreimageLen
	}
	var out Digest
	sum := sha256.Sum256(preimage)
	copy(out[:], sum[:])
	return out, nil
}
