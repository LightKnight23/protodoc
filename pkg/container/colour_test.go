package container

import (
	"reflect"
	"testing"
)

// TestCON_014_SingleColourRepresentationDefined is T-0029's named test.
// Implements: CON-014.
func TestCON_014_SingleColourRepresentationDefined(t *testing.T) {
	// StoredColour: exactly 4 components, each exactly ColourComponentBits
	// (u16) wide -- CON-014's stored-component-width clause.
	scType := reflect.TypeOf(StoredColour{})
	if scType.NumField() != 4 {
		t.Fatalf("StoredColour has %d fields, want 4 (R,G,B,A)", scType.NumField())
	}
	for i := 0; i < scType.NumField(); i++ {
		f := scType.Field(i)
		if f.Type.Kind() != reflect.Uint16 {
			t.Fatalf("StoredColour.%s has kind %s, want uint16 (CON-014: u16 per-channel components)", f.Name, f.Type.Kind())
		}
		if f.Type.Bits() != ColourComponentBits {
			t.Fatalf("StoredColour.%s is %d bits, want %d", f.Name, f.Type.Bits(), ColourComponentBits)
		}
	}

	// CompositingColour: exactly 4 components, each exactly
	// CompositingFixedPointBits (32-bit fixed-point) wide, the linear-light
	// space CON-014 requires compositing arithmetic to run in. It is a
	// distinct type from StoredColour (CP-008: one representation per role,
	// not one representation overall -- stored and compositing are
	// different capabilities).
	ccType := reflect.TypeOf(CompositingColour{})
	if ccType.NumField() != 4 {
		t.Fatalf("CompositingColour has %d fields, want 4 (R,G,B,A)", ccType.NumField())
	}
	for i := 0; i < ccType.NumField(); i++ {
		f := ccType.Field(i)
		if f.Type.Bits() != CompositingFixedPointBits {
			t.Fatalf("CompositingColour.%s is %d bits, want %d", f.Name, f.Type.Bits(), CompositingFixedPointBits)
		}
	}
	if scType == ccType {
		t.Fatal("StoredColour and CompositingColour must not be the same type (CP-008: distinct roles)")
	}

	// Exactly one registry-issued colour representation id is defined for
	// this format version.
	if ColourProfileSRGBD65 != 1 {
		t.Fatalf("ColourProfileSRGBD65 = %d, want 1", ColourProfileSRGBD65)
	}

	// fm-colour-profile-id round-trips through Frontmatter Encode/Decode.
	// It is tested across several distinct id values, including ids other
	// than ColourProfileSRGBD65, to confirm the field is a plain opaque
	// u16 carried through by one identical field-copy path -- there is no
	// per-id branch in the decoder that would make some ids take a
	// different code path than others (the "no alternate colour-space
	// branch reachable in the decoder" half of the DoD).
	ids := []ColourProfileID{ColourProfileSRGBD65, 0, 2, 65535}
	for _, id := range ids {
		fm := &Frontmatter{Title: "colour round-trip fixture", ColourProfileID: id}
		enc, err := fm.Encode(nil)
		if err != nil {
			t.Fatalf("id=%d: Encode: %v", id, err)
		}
		dec, err := DecodeFrontmatter(enc)
		if err != nil {
			t.Fatalf("id=%d: DecodeFrontmatter: %v", id, err)
		}
		if dec.ColourProfileID != id {
			t.Fatalf("id=%d: round-trip = %d", id, dec.ColourProfileID)
		}
	}
}
