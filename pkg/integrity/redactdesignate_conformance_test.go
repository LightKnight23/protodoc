package integrity

import "testing"

// TestRedactDesignateConformance001 is T-0188's named conformance test (corpus
// redact-designate-conformance-001, FR-074). It ships a corpus of redactable-
// subtree designations and asserts each behaves per FR-074: a designated
// subtree commits with a salted hiding-and-binding commitment whose salt is
// stored inside, and removal deletes frame+salt while retaining the exact
// commitment digest.
func TestRedactDesignateConformance001(t *testing.T) {
	corpus := []struct {
		name  string
		frame []byte
		salt  [SaltSize]byte
	}{
		{"short text block", []byte("hi"), redSalt(0x01)},
		{"empty frame", []byte{}, redSalt(0x02)},
		{"binary-ish frame", []byte{0x00, 0xFF, 0x30, 0x82, 0x10}, redSalt(0x03)},
		{"longer frame", []byte("a considerably longer redactable content-model record frame with punctuation, digits 0123, and unicode café"), redSalt(0x04)},
	}

	for _, c := range corpus {
		sub := DesignateRedactable(redUnitID(0x10), c.frame, c.salt)
		if !sub.IsDesignated() || !sub.SaltStoredInside() {
			t.Errorf("%s: not designated / salt not inside", c.name)
		}
		commit, err := sub.Commitment()
		if err != nil {
			t.Fatalf("%s: Commitment: %v", c.name, err)
		}
		if commit != TCLeafRedactable(c.salt, c.frame) {
			t.Errorf("%s: commitment != salted redactable leaf", c.name)
		}

		// Removal retains the exact commitment and erases salt+frame.
		if err := sub.Remove(); err != nil {
			t.Fatalf("%s: Remove: %v", c.name, err)
		}
		after, _ := sub.Commitment()
		if after != commit {
			t.Errorf("%s: commitment changed across removal", c.name)
		}
		if sub.SaltPresent() || sub.Frame != nil {
			t.Errorf("%s: salt/frame not erased after removal", c.name)
		}
	}
}
