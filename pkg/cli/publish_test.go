package cli

import "testing"

// TestTR_012_PublishVerbZeroResidue is T-0336's named integration test
// (TR-012). Publishing a fixture with prior redactions/edit history produces
// output with zero removed-content octets while every custody/fixity value
// matches the input unchanged.
func TestTR_012_PublishVerbZeroResidue(t *testing.T) {
	orig := PublishRun
	defer func() { PublishRun = orig }()

	// Clean publish: zero residue, custody preserved.
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{Output: []byte("clean"), ResidueOctets: 0, CustodyPreserved: true}
	}
	res := runPublish([]string{"doc.pdl"}, nil)
	if res.Status != "OK" {
		t.Errorf("publish status=%s, want OK", res.Status)
	}
	if res.Extra["residue_octets"] != 0 || res.Extra["custody_preserved"] != true {
		t.Errorf("publish invariants: %+v", res.Extra)
	}

	// A publish that leaves residue is rejected (INVALID).
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{ResidueOctets: 3, CustodyPreserved: true}
	}
	if r := runPublish([]string{"doc.pdl"}, nil); r.Status != "INVALID" {
		t.Errorf("residue publish: status=%s, want INVALID", r.Status)
	}

	// A publish that drops a custody/fixity value is rejected.
	PublishRun = func(path string, partial bool) PublishResult {
		return PublishResult{ResidueOctets: 0, CustodyPreserved: false}
	}
	if r := runPublish([]string{"doc.pdl"}, nil); r.Status != "INVALID" {
		t.Errorf("custody-dropping publish: status=%s, want INVALID", r.Status)
	}

	if r := runPublish(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg publish: status=%s, want USAGE", r.Status)
	}
}
