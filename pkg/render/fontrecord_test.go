package render

import (
	"bytes"
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFontRecordSevenFields is T-0245's named conformance test (vector id
// font_record_seven_fields; FR-089). It holds the FontRecord at EXACTLY seven
// declared fields, round-trips a record through its wire form with all seven
// present, and rejects a conformance fixture with a missing or extra field.
func TestFontRecordSevenFields(t *testing.T) {
	// The struct declares exactly seven fields (FR-089), and the pinned
	// constant agrees.
	if got := fontRecordFieldCount(); got != FontRecordFieldCount {
		t.Fatalf("FontRecord has %d fields, want exactly %d (FR-089)", got, FontRecordFieldCount)
	}
	if FontRecordFieldCount != 7 {
		t.Fatalf("FontRecordFieldCount = %d, want 7", FontRecordFieldCount)
	}

	// A full record round-trips with all seven fields present.
	rec := FontRecord{
		Name:                 "Protodoc Sans",
		Version:              "1.2.3",
		ContentDigest:        pdlfmt.Digest256{0xAB, 0xCD},
		VariationAxes:        []int32{400, -100},
		CodePoints:           []rune{'A', 'B', 0x1F600},
		LayoutFeatures:       []string{"liga", "kern"},
		EmbeddingPermissions: []uint16{0x0000, 0x0008},
	}
	wire := rec.Encode()
	fields, err := DecodeFontRecordFieldCount(wire)
	if err != nil {
		t.Fatalf("valid seven-field record rejected: %v", err)
	}
	if len(fields) != 7 {
		t.Fatalf("decoded %d fields, want 7", len(fields))
	}
	// Field presence round-trips: name and version octets survive verbatim.
	if !bytes.Equal(fields[0], []byte("Protodoc Sans")) || !bytes.Equal(fields[1], []byte("1.2.3")) {
		t.Errorf("field payloads did not round-trip: %q %q", fields[0], fields[1])
	}
	if !bytes.Equal(fields[2], rec.ContentDigest[:]) {
		t.Error("content-digest field did not round-trip")
	}

	// Conformance fixture: a MISSING field (six fields declared) is rejected.
	missing := pdlfmt.AppendVarint(nil, 6)
	for i := 0; i < 6; i++ {
		missing = pdlfmt.AppendVarint(missing, 0)
	}
	if _, err := DecodeFontRecordFieldCount(missing); !errors.Is(err, ErrFontRecordFieldCount) {
		t.Errorf("six-field fixture: err = %v, want ErrFontRecordFieldCount", err)
	}

	// Conformance fixture: an EXTRA field (eight fields declared) is rejected.
	extra := pdlfmt.AppendVarint(nil, 8)
	for i := 0; i < 8; i++ {
		extra = pdlfmt.AppendVarint(extra, 0)
	}
	if _, err := DecodeFontRecordFieldCount(extra); !errors.Is(err, ErrFontRecordFieldCount) {
		t.Errorf("eight-field fixture: err = %v, want ErrFontRecordFieldCount", err)
	}
}
