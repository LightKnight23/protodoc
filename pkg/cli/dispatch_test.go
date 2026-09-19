package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// wantVerbs is TR-012's exact text: "validate, inspect, extract, verify,
// diff, merge, project, redact, publish, sign and migrate", in that order.
var wantVerbs = []string{
	"validate", "inspect", "extract", "verify", "diff",
	"merge", "project", "redact", "publish", "sign", "migrate",
}

// TestTR_012_DispatchRegistersAllElevenVerbs is T-0325's named test.
func TestTR_012_DispatchRegistersAllElevenVerbs(t *testing.T) {
	if len(wantVerbs) != 11 {
		t.Fatalf("test fixture itself lists %d verbs, want 11", len(wantVerbs))
	}

	got := Names()
	if len(got) != 11 {
		t.Fatalf("Names() returned %d verbs, want exactly 11: %v", len(got), got)
	}
	for i, name := range wantVerbs {
		if got[i] != name {
			t.Errorf("verb %d = %q, want %q (registration order must match TR-012's own text)", i, got[i], name)
		}
	}

	// `protodoc <verb> --help` must print per-verb usage naming that verb,
	// for every one of the 11 registered verbs, with none missing.
	for _, name := range wantVerbs {
		var stdout, stderr bytes.Buffer
		code := Dispatch([]string{name, "--help"}, &stdout, &stderr)
		if code != exitOK {
			t.Errorf("%s --help: exit code = %d, want %d (OK)", name, code, exitOK)
		}
		if stderr.Len() != 0 {
			t.Errorf("%s --help: unexpected stderr: %q", name, stderr.String())
		}
		if !strings.Contains(stdout.String(), name) {
			t.Errorf("%s --help: usage output %q does not name the verb", name, stdout.String())
		}
	}
}

// TestTR_012_DispatchUnregisteredVerbIsUsage confirms an unrecognized verb
// name exits USAGE without producing a stdout envelope. T-0325's DoD
// requires this happen "without importing or calling any downstream
// package" — trivially satisfied here since package cli imports nothing
// beyond the Go standard library and Lookup fails before any Verb.Run is
// ever invoked.
func TestTR_012_DispatchUnregisteredVerbIsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Dispatch([]string{"frobnicate"}, &stdout, &stderr)
	if code != exitUsage {
		t.Errorf("exit code = %d, want %d (USAGE)", code, exitUsage)
	}
	if stdout.Len() != 0 {
		t.Errorf("unexpected stdout for an unrecognized verb: %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Error("expected a diagnostic on stderr for an unrecognized verb")
	}
}

// TestTR_012_DispatchNoArgsIsUsage confirms invoking protodoc with no verb
// at all is also USAGE, not a crash or a silent success.
func TestTR_012_DispatchNoArgsIsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Dispatch(nil, &stdout, &stderr); code != exitUsage {
		t.Errorf("exit code = %d, want %d (USAGE)", code, exitUsage)
	}
}

// TestTR_012_DispatchGlobalHelpAndVersion covers the shared --help and
// --version global flags.
func TestTR_012_DispatchGlobalHelpAndVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Dispatch([]string{"--help"}, &stdout, &stderr); code != exitOK {
		t.Errorf("--help: exit code = %d, want %d (OK)", code, exitOK)
	}
	for _, name := range wantVerbs {
		if !strings.Contains(stdout.String(), name) {
			t.Errorf("--help output missing verb %q", name)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if code := Dispatch([]string{"--version"}, &stdout, &stderr); code != exitOK {
		t.Errorf("--version: exit code = %d, want %d (OK)", code, exitOK)
	}
	if stdout.Len() == 0 {
		t.Error("--version produced no output")
	}
}

// TestTR_012_DispatchEnvelopeIsSingleJSONObject confirms a registered
// verb's stdout carries exactly the cli.md S0 envelope shape (verb,
// status, exit_code, findings) and nothing else on stdout.
func TestTR_012_DispatchEnvelopeIsSingleJSONObject(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Dispatch([]string{"validate", "somefile.pdl", "--json"}, &stdout, &stderr)
	// validate is now wired to the real decode chain (T-0373): an absent file
	// resolves to USAGE (T-0390: an unreadable path is USAGE, not INVALID)
	// with a real finding. The point of THIS test is the single-JSON-envelope
	// shape, which holds regardless of the resolved status, so we assert the
	// shape and a non-crash exit code rather than a specific verdict.
	if code != int(StatusUsage.Code()) {
		t.Fatalf("exit code = %d, want %d (USAGE: absent file, real decode)", code, StatusUsage.Code())
	}

	var obj map[string]any
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	if err := dec.Decode(&obj); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if dec.More() {
		t.Error("stdout carries more than one JSON value")
	}
	for _, key := range []string{"verb", "status", "exit_code", "findings"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("envelope missing required key %q", key)
		}
	}
	if obj["verb"] != "validate" {
		t.Errorf("verb = %v, want validate", obj["verb"])
	}
}

// TestTR_012_LookupUnknownVerb confirms Lookup itself reports absence
// rather than a zero-value Verb that could be mistaken for a real one.
func TestTR_012_LookupUnknownVerb(t *testing.T) {
	if _, ok := Lookup("frobnicate"); ok {
		t.Error("Lookup(\"frobnicate\") reported found, want not found")
	}
	if v, ok := Lookup("validate"); !ok || v.Name != "validate" {
		t.Errorf("Lookup(\"validate\") = %+v, %v, want the registered validate Verb", v, ok)
	}
}
