package integrity

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

// TestFR_064_PresentationArtefactRoundTrip is T-0154's named unit test
// (FR-064). The PRESENTATION_ARTEFACT record (document.abnf S7.4) with its
// embedded FontRecords (S7.4.1) encodes and decodes byte-exact with every
// field surviving, including negative page-geometry coordinates and the
// closed 7-field FontRecord shape.
func TestFR_064_PresentationArtefactRoundTrip(t *testing.T) {
	orig := PresentationArtefact{
		ProfileVersion: 1,
		// Authored order, including a negative coordinate (off-page object).
		PageGeometry: []int64{0, 914400, -12700, 609600},
		FontIdentity: []FontRecord{
			{
				Name:       "Test Sans",
				Version:    "1.2.3",
				Digest:     Digest{0xAB},
				Axes:       []int64{400, -100, 0},
				Codepoints: []uint32{0x41, 0x42, 0x1F600, 0x10FFFF},
				Features:   []uint16{1, 2, 300},
				EmbedPerm:  0x03,
			},
			{
				Name:      "Empty Font",
				Version:   "",
				Digest:    Digest{0xCD},
				EmbedPerm: 0x00,
			},
		},
	}

	enc, err := orig.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodePresentationArtefact(enc)
	if err != nil {
		t.Fatalf("DecodePresentationArtefact: %v", err)
	}
	if dec.ProfileVersion != orig.ProfileVersion {
		t.Errorf("profile version mismatch")
	}
	if !reflect.DeepEqual(dec.PageGeometry, orig.PageGeometry) {
		t.Errorf("page geometry did not survive: got %v want %v", dec.PageGeometry, orig.PageGeometry)
	}
	if !reflect.DeepEqual(dec.FontIdentity, orig.FontIdentity) {
		t.Errorf("font identity did not survive:\n got %+v\nwant %+v", dec.FontIdentity, orig.FontIdentity)
	}
	reEnc, err := dec.Encode()
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(enc, reEnc) {
		t.Fatalf("round trip not byte-exact")
	}

	// A codepoint out of range is rejected on encode.
	bad := PresentationArtefact{
		ProfileVersion: 1,
		FontIdentity:   []FontRecord{{Name: "x", Codepoints: []uint32{0x110000}}},
	}
	if _, err := bad.Encode(); !errors.Is(err, ErrCodepointRange) {
		t.Errorf("out-of-range codepoint: err = %v, want ErrCodepointRange", err)
	}

	// Wrong discriminant rejected.
	corrupt := append([]byte(nil), enc...)
	if len(corrupt) >= 3 && corrupt[2] == presentationArtefactDiscriminant {
		corrupt[2] = 0x0B
		if _, err := DecodePresentationArtefact(corrupt); !errors.Is(err, ErrPresentationDiscriminant) {
			t.Errorf("wrong discriminant: err = %v, want ErrPresentationDiscriminant", err)
		}
	}
}
