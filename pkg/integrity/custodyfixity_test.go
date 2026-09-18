package integrity

import (
	"bytes"
	"testing"
)

// TestFR_081_PublishPreservesCustodyFixityValues is T-0203's named integration
// test (FR-081). WHEN publish runs, every custody and fixity value in the
// input (signatures, time attestations) is present UNCHANGED in the output --
// even alongside actor-identity stripping, which must not touch them.
func TestFR_081_PublishPreservesCustodyFixityValues(t *testing.T) {
	signature := []byte("SIG:R||S:deadbeefcafef00d")
	timeAttestation := []byte("TSA:2030-01-01T00:00:00Z:token")
	content := []byte("public content")
	actor := []byte("actor:dave")

	in := PublishInput{
		Retained:            []ContentRecord{{UnitID: redUnitID(0x01), Frame: append(append([]byte(nil), content...), actor...)}},
		ActorIdentityValues: [][]byte{actor},
		CustodyFixityValues: [][]byte{signature, timeAttestation},
	}

	out := Publish(in)

	// Every custody/fixity value is present unchanged.
	for _, v := range in.CustodyFixityValues {
		if !bytes.Contains(out.Emitted, v) {
			t.Errorf("custody/fixity value %q was not preserved unchanged in the output", v)
		}
	}
	// The actor-identity value is still stripped (FR-080 unaffected).
	if bytes.Contains(out.Emitted, actor) {
		t.Error("actor-identity value survived; stripping must still apply")
	}
	// Content is retained.
	if !bytes.Contains(out.Emitted, content) {
		t.Error("public content was lost")
	}

	// A custody value that happens to CONTAIN an actor-identity substring is
	// still preserved unchanged (custody/fixity preservation takes precedence
	// for the explicitly-preserved values, since it is appended after the
	// strip).
	custodyWithActorSubstr := append(append([]byte("SIG-by-"), actor...), []byte("-end")...)
	in2 := PublishInput{
		Retained:            []ContentRecord{{UnitID: redUnitID(0x02), Frame: append(append([]byte(nil), content...), actor...)}},
		ActorIdentityValues: [][]byte{actor},
		CustodyFixityValues: [][]byte{custodyWithActorSubstr},
	}
	out2 := Publish(in2)
	if !bytes.Contains(out2.Emitted, custodyWithActorSubstr) {
		t.Error("a custody value containing an actor substring was not preserved unchanged")
	}
	// The bare actor value (from content) is still stripped from the content
	// portion (the content prefix has no standalone actor value left).
	if bytes.Count(out2.Emitted, actor) != 1 {
		// Exactly one occurrence: the one inside the preserved custody value.
		t.Errorf("expected the actor value only within the preserved custody value, found %d occurrences", bytes.Count(out2.Emitted, actor))
	}
}
