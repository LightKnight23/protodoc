package integrity

import "testing"

// TestFR_050_VerdictIsClosedFourValueEnum is T-0157's named unit test
// (FR-050). Verification status is reported as a value distinct from extracted
// content, taking one of a CLOSED set of defined verdicts. This asserts the
// Verdict enum is exactly the four defined values (valid, unverified,
// unavailable_state, unattested), that each has a distinct canonical name,
// that AllVerdicts enumerates precisely those, and that any other integer is
// not a defined verdict -- so the enum cannot silently grow or admit an
// unrecognised status.
func TestFR_050_VerdictIsClosedFourValueEnum(t *testing.T) {
	want := map[Verdict]string{
		VerdictValid:            "valid",
		VerdictUnverified:       "unverified",
		VerdictUnavailableState: "unavailable_state",
		VerdictUnattested:       "unattested",
	}

	// Exactly four verdicts, enumerated by AllVerdicts in enum order.
	if len(AllVerdicts) != 4 {
		t.Fatalf("AllVerdicts has %d entries, want exactly 4 (FR-050 closed set)", len(AllVerdicts))
	}
	seen := map[Verdict]bool{}
	names := map[string]bool{}
	for _, v := range AllVerdicts {
		if !v.IsDefined() {
			t.Errorf("AllVerdicts contains an undefined verdict %d", v)
		}
		if seen[v] {
			t.Errorf("AllVerdicts lists verdict %v twice", v)
		}
		seen[v] = true
		n := v.String()
		if n != want[v] {
			t.Errorf("verdict %d name = %q, want %q", v, n, want[v])
		}
		if names[n] {
			t.Errorf("two verdicts share the name %q", n)
		}
		names[n] = true
	}
	if len(seen) != 4 {
		t.Fatalf("AllVerdicts covers %d distinct verdicts, want 4", len(seen))
	}

	// Every intended value is present.
	for v := range want {
		if !seen[v] {
			t.Errorf("closed set missing verdict %v", v)
		}
	}

	// A value outside the closed set is not defined and has no real name.
	bogus := Verdict(len(AllVerdicts) + 1)
	if bogus.IsDefined() {
		t.Errorf("out-of-range verdict %d reported as defined", bogus)
	}
	if bogus.String() != "undefined" {
		t.Errorf("out-of-range verdict name = %q, want \"undefined\"", bogus.String())
	}
	if bogus.CarriesSignerIdentity() {
		t.Errorf("out-of-range verdict must not carry a signer identity")
	}
}
