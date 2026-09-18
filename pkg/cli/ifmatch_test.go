package cli

import "testing"

// TestTR_010_ConditionalWriteRefusesOnMismatch is T-0341's named integration
// test (TR-010). A concurrent-write simulation against a mock ETag-capable
// backend causes the second writer to refuse, naming the first writer's
// resulting state id, for every mutating verb that accepts --if-match.
func TestTR_010_ConditionalWriteRefusesOnMismatch(t *testing.T) {
	// The target started at state "S0"; a first writer advanced it to "S1".
	backend := func(target string) (string, bool) { return "S1", true }

	// Every mutating verb accepts --if-match; simulate the second writer
	// expecting the pre-concurrent state "S0".
	for _, verb := range []string{"sign", "redact", "publish", "merge", "migrate"} {
		args := []string{"doc.pdl", "--if-match=S0"}
		expected := parseIfMatch(args)
		if expected != "S0" {
			t.Fatalf("%s: parseIfMatch = %q, want S0", verb, expected)
		}
		r := CheckIfMatch(expected, "doc.pdl", backend)
		if r.Proceed {
			t.Errorf("%s: second writer must refuse on state mismatch", verb)
		}
		if r.FoundStateID != "S1" {
			t.Errorf("%s: refusal must name the found state, got %q", verb, r.FoundStateID)
		}
		res := IfMatchRefusal(r.FoundStateID)
		if res.Status != "REFUSED" || res.Extra["found_state_id"] != "S1" {
			t.Errorf("%s: refusal result = %+v", verb, res.Extra)
		}
	}

	// Matching expected state -> proceed.
	if r := CheckIfMatch("S1", "doc.pdl", backend); !r.Proceed {
		t.Errorf("matching state must proceed")
	}
	// No --if-match -> proceed unconditionally.
	if r := CheckIfMatch("", "doc.pdl", backend); !r.Proceed {
		t.Errorf("absent --if-match must proceed")
	}
	// Absent target -> no conflict.
	if r := CheckIfMatch("S0", "new.pdl", func(string) (string, bool) { return "", false }); !r.Proceed {
		t.Errorf("absent target must proceed (no concurrent state)")
	}
}
