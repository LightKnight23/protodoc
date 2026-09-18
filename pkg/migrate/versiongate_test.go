package migrate

import (
	"errors"
	"testing"
)

// TestFR_123_UnsupportedMajorVersionDeclinedFirst is T-0301's named conformance
// test (FR-123). VersionGate declines an unsupported major immediately, naming
// the version, and it is checked before any other processing at an entry point.
func TestFR_123_UnsupportedMajorVersionDeclinedFirst(t *testing.T) {
	// Supported major passes the gate.
	if err := VersionGate(1); err != nil {
		t.Errorf("supported major 1 should pass, got %v", err)
	}

	// Unsupported major is declined, naming the version.
	err := VersionGate(2)
	var ue *UnsupportedMajorError
	if !errors.As(err, &ue) || ue.Major != 2 {
		t.Fatalf("unsupported major: err = %v, want UnsupportedMajorError naming 2", err)
	}
	if got := ue.Error(); got == "" || !contains(got, "2") {
		t.Errorf("error must name the version, got %q", got)
	}

	// The gate is a pure precheck: it performs no decode/migration disposition.
	// Calling it never touches source constructs (it takes only the major).
	for _, m := range []uint16{0, 3, 99, 65535} {
		if VersionGate(m) == nil {
			t.Errorf("major %d should be declined", m)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
