package integrity

import (
	"crypto/sha256"
	"testing"
)

// TestFR_063_SignedObjectPreimageAssembly is T-0152's named unit test
// (FR-063). It asserts the signed_object preimage is the exact 129-octet
// concatenation %x04 || t-c-root || structure-digest ||
// presentation-artefact-digest || coverage-descriptor-digest, in that fixed
// field order, and that SignedObject is SHA-256 over it. It also checks the
// digest is sensitive to each of the four inputs (a change to any one changes
// signed_object).
func TestFR_063_SignedObjectPreimageAssembly(t *testing.T) {
	fill := func(b byte) Digest {
		var d Digest
		for i := range d {
			d[i] = b
		}
		return d
	}
	in := SignedObjectInput{
		TCRoot:                     fill(0x11),
		StructureDigest:            fill(0x22),
		PresentationArtefactDigest: fill(0x33),
		CoverageDescriptorDigest:   fill(0x44),
	}

	pre := AssembleSignedObjectPreimage(in)
	if len(pre) != SignedObjectPreimageLen || len(pre) != 129 {
		t.Fatalf("preimage length = %d, want 129", len(pre))
	}
	// Field order: domain octet, then the four digests in order.
	if pre[0] != DomainSignedObject {
		t.Errorf("preimage[0] = 0x%02x, want DomainSignedObject 0x%02x", pre[0], DomainSignedObject)
	}
	segments := []struct {
		off int
		d   Digest
	}{
		{1, in.TCRoot}, {33, in.StructureDigest}, {65, in.PresentationArtefactDigest}, {97, in.CoverageDescriptorDigest},
	}
	for _, s := range segments {
		for i := 0; i < 32; i++ {
			if pre[s.off+i] != s.d[i] {
				t.Fatalf("preimage octet %d != expected digest byte", s.off+i)
			}
		}
	}

	// SignedObject == SHA-256(preimage).
	so, err := SignedObject(in)
	if err != nil {
		t.Fatalf("SignedObject: %v", err)
	}
	want := sha256.Sum256(pre)
	if [32]byte(so) != want {
		t.Fatal("SignedObject is not SHA-256 of the assembled preimage")
	}

	// Sensitivity: changing any one input changes signed_object.
	base, _ := SignedObject(in)
	mutators := []func(*SignedObjectInput){
		func(x *SignedObjectInput) { x.TCRoot = fill(0x99) },
		func(x *SignedObjectInput) { x.StructureDigest = fill(0x99) },
		func(x *SignedObjectInput) { x.PresentationArtefactDigest = fill(0x99) },
		func(x *SignedObjectInput) { x.CoverageDescriptorDigest = fill(0x99) },
	}
	for i, m := range mutators {
		v := in
		m(&v)
		vo, _ := SignedObject(v)
		if vo == base {
			t.Errorf("input %d change did not alter signed_object", i)
		}
	}
}
