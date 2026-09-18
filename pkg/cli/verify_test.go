package cli

import "testing"

// TestTR_012_VerifyVerbVerdictMapping is T-0331's named integration test
// (TR-012). Fixtures covering all four verdict outcomes each produce the
// documented payload shape and correct exit code; no non-Valid fixture's
// stdout carries a signer-identity field.
func TestTR_012_VerifyVerbVerdictMapping(t *testing.T) {
	orig := VerifyRun
	defer func() { VerifyRun = orig }()

	cases := []struct {
		name       string
		verdict    string
		signer     string
		wantStatus string
	}{
		{"valid", "valid", "did:key:zAlice", "OK"},
		{"attested-with-declared-omissions", "attested_with_declared_omissions", "", "OK"},
		{"unverified", "unverified", "", "UNVERIFIED"},
		{"covering-unavailable-state", "covering_unavailable_state", "", "UNAVAILABLE"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			VerifyRun = func(string) []VerifyVerdict {
				return []VerifyVerdict{{Verdict: c.verdict, Signer: c.signer}}
			}
			res := runVerify([]string{"doc.pdl"}, nil)
			if res.Status != c.wantStatus {
				t.Errorf("%s: status=%s, want %s", c.name, res.Status, c.wantStatus)
			}
			sigs := res.Extra["signatures"].([]map[string]any)
			if len(sigs) != 1 || sigs[0]["verdict"] != c.verdict {
				t.Fatalf("%s: signatures payload = %+v", c.name, sigs)
			}
			// Signer identity present ONLY for a Valid verdict.
			_, hasSigner := sigs[0]["signer"]
			if c.verdict == "valid" && !hasSigner {
				t.Errorf("valid verdict must carry signer identity")
			}
			if c.verdict != "valid" && hasSigner {
				t.Errorf("%s: non-Valid verdict must NOT carry a signer field", c.name)
			}
		})
	}

	// Missing file -> USAGE.
	if r := runVerify(nil, nil); r.Status != "USAGE" {
		t.Errorf("no-arg verify: status=%s, want USAGE", r.Status)
	}
}
