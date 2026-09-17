package validate

import (
	"bytes"
	"testing"
)

// TestFR_011_RegistryExcerptCompletenessStep12 is T-0370's named unit test. It
// covers FR-011 / data-model.md S7 step 12:
//
//   - RegistryExcerpt encodes and decodes byte-exact (round trip);
//   - a durable_claim = 1 document missing a RegistryExcerpt entry for a
//     referenced external id fails validation step 12, naming the missing id,
//     as a distinct step-12 finding;
//   - a complete durable-profile document passes;
//   - a durable_claim = 0 document passes regardless (the excerpt is optional).
func TestFR_011_RegistryExcerptCompletenessStep12(t *testing.T) {
	// --- byte-exact round trip ---
	orig := RegistryExcerpt{Entries: []RegistryEntry{
		{Kind: KindUnicodeVersionID, Value: 15, SpecPointer: "Unicode 15.0.0", SnapshotDigest: digestFill(1)},
		{Kind: KindShapingProfileID, Value: 7, SpecPointer: "shaping profile 7", SnapshotDigest: digestFill(2)},
		{Kind: KindExtensionToken, Value: 4242, SpecPointer: "ext token spec pointer", SnapshotDigest: digestFill(3)},
	}}
	enc, err := orig.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dec, err := DecodeRegistryExcerpt(enc)
	if err != nil {
		t.Fatalf("DecodeRegistryExcerpt: %v", err)
	}
	reenc, err := dec.Encode()
	if err != nil {
		t.Fatalf("re-Encode: %v", err)
	}
	if !bytes.Equal(enc, reenc) {
		t.Fatalf("round trip not byte-exact:\n first=%x\nsecond=%x", enc, reenc)
	}
	if len(dec.Entries) != len(orig.Entries) {
		t.Fatalf("decoded %d entries, want %d", len(dec.Entries), len(orig.Entries))
	}
	for i := range orig.Entries {
		if dec.Entries[i] != orig.Entries[i] {
			t.Errorf("entry %d: decoded %+v, want %+v", i, dec.Entries[i], orig.Entries[i])
		}
	}

	// A wrong discriminant is rejected.
	bad := append([]byte(nil), enc...)
	// The discriminant field is tag 0, len 1, value 0x09 at bytes [0,1,2].
	if len(bad) >= 3 && bad[2] == registryExcerptDiscriminant {
		bad[2] = 0x08
		if _, err := DecodeRegistryExcerpt(bad); err == nil {
			t.Error("decode accepted a record with a non-0x09 discriminant")
		}
	}

	// --- step-12 completeness ---
	referenced := []ExternalID{
		{Kind: KindUnicodeVersionID, Value: 15},
		{Kind: KindShapingProfileID, Value: 7},
		{Kind: KindExtensionToken, Value: 4242},
	}

	// Complete durable-profile document passes.
	if res := CheckRegistryExcerptCompleteness(true, orig, referenced); !res.Passed {
		t.Errorf("complete durable document failed step 12: missing %v", res.Missing)
	}

	// Missing one referenced id fails, naming the missing id.
	incomplete := RegistryExcerpt{Entries: orig.Entries[:2]} // drops the ext-token entry
	res := CheckRegistryExcerptCompleteness(true, incomplete, referenced)
	if res.Passed {
		t.Fatal("durable document missing an excerpt entry passed step 12, want failure")
	}
	if len(res.Missing) != 1 || res.Missing[0] != (ExternalID{Kind: KindExtensionToken, Value: 4242}) {
		t.Fatalf("missing ids = %v, want [ext-token 4242]", res.Missing)
	}
	f := RegistryCompletenessFinding(res)
	if f == nil || f.Step != StepRegistryExcerpt || f.RuleID != RegistryExcerptCompletenessCheckKey {
		t.Fatalf("expected a step-12 %s finding, got %+v", RegistryExcerptCompletenessCheckKey, f)
	}

	// durable_claim = 0 passes regardless (excerpt optional), even with an
	// empty excerpt and referenced ids present.
	if res := CheckRegistryExcerptCompleteness(false, RegistryExcerpt{}, referenced); !res.Passed {
		t.Errorf("non-durable document failed step 12, but the excerpt is optional for it")
	}
}

func digestFill(b byte) [32]byte {
	var d [32]byte
	for i := range d {
		d[i] = b
	}
	return d
}
