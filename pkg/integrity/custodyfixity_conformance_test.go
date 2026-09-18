package integrity

import (
	"bytes"
	"testing"
)

// TestRedactCustodyFixityConformance001 is T-0204's named conformance test
// (corpus redact-custody-fixity-conformance-001, FR-081). A corpus of publish
// operations, each combining actor-identity stripping with custody/fixity
// preservation, asserts every custody and fixity value is present unchanged in
// the output while actor-identity values are stripped.
func TestRedactCustodyFixityConformance001(t *testing.T) {
	sig1 := []byte("SIG-1-bytes")
	sig2 := []byte("SIG-2-bytes")
	tsa := []byte("TSA-token-bytes")
	actor := []byte("actor:eve")
	content := []byte("body")

	cases := []struct {
		name    string
		custody [][]byte
	}{
		{"single signature", [][]byte{sig1}},
		{"two signatures + time attestation", [][]byte{sig1, sig2, tsa}},
		{"no custody values", nil},
	}

	for _, c := range cases {
		out := Publish(PublishInput{
			Retained:            []ContentRecord{{UnitID: redUnitID(0x01), Frame: append(append([]byte(nil), content...), actor...)}},
			ActorIdentityValues: [][]byte{actor},
			CustodyFixityValues: c.custody,
		})
		// Every custody/fixity value present unchanged.
		for _, v := range c.custody {
			if !bytes.Contains(out.Emitted, v) {
				t.Errorf("%s: custody/fixity value %q not preserved", c.name, v)
			}
		}
		// Actor identity stripped from content.
		if c.name != "no custody values" || len(c.custody) == 0 {
			// content portion has no standalone actor value.
			contentPortion := out.Emitted
			for _, v := range c.custody {
				contentPortion = bytes.Replace(contentPortion, v, nil, 1)
			}
			if bytes.Contains(contentPortion, actor) {
				t.Errorf("%s: actor-identity value survived in content", c.name)
			}
		}
		// Content retained.
		if !bytes.Contains(out.Emitted, content) {
			t.Errorf("%s: content lost", c.name)
		}
	}
}
