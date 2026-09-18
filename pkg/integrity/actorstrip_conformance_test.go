package integrity

import (
	"bytes"
	"testing"
)

// TestRedactActorStripConformance001 is T-0202's named conformance test
// (corpus redact-actor-strip-conformance-001, FR-080). A corpus of documents
// carrying actor-identity values in various positions (leading, trailing,
// embedded, repeated) publishes with every actor-identity value stripped and
// no value from the identity inventory surviving, while surrounding content
// is retained.
func TestRedactActorStripConformance001(t *testing.T) {
	actor := []byte("actor:carol")
	cases := []struct {
		name  string
		frame []byte
	}{
		{"leading actor value", append(append([]byte(nil), actor...), []byte(" body")...)},
		{"trailing actor value", append([]byte("body "), actor...)},
		{"embedded actor value", append(append([]byte("pre "), actor...), []byte(" post")...)},
		{"repeated actor value", bytes.Join([][]byte{actor, []byte("mid"), actor}, nil)},
		{"no actor value", []byte("clean content only")},
	}

	for _, c := range cases {
		out := Publish(PublishInput{
			Retained:            []ContentRecord{{UnitID: redUnitID(0x01), Frame: c.frame}},
			ActorIdentityValues: [][]byte{actor},
		})
		if bytes.Contains(out.Emitted, actor) {
			t.Errorf("%s: actor-identity value survived publish", c.name)
		}
	}

	// Multiple distinct actor values, all stripped in one publish.
	a1, a2, a3 := []byte("actor:1"), []byte("device:2"), []byte("attrib:3")
	frame := bytes.Join([][]byte{[]byte("x"), a1, []byte("y"), a2, []byte("z"), a3}, nil)
	out := Publish(PublishInput{
		Retained:            []ContentRecord{{UnitID: redUnitID(0x02), Frame: frame}},
		ActorIdentityValues: [][]byte{a1, a2, a3},
	})
	for _, a := range [][]byte{a1, a2, a3} {
		if bytes.Contains(out.Emitted, a) {
			t.Errorf("actor value %q survived multi-value strip", a)
		}
	}
	// The non-identity separators remain.
	for _, keep := range [][]byte{[]byte("x"), []byte("y"), []byte("z")} {
		if !bytes.Contains(out.Emitted, keep) {
			t.Errorf("non-identity separator %q was wrongly stripped", keep)
		}
	}
}
