package integrity

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestNFR_005_NoAmbientValuesOutsideAllowlistedSites is T-0165's named unit
// test (NFR-005, audit A-AMBIENT). It asserts two things:
//
//  1. The ambient-value allowlist is exactly the four named sites
//     (identifier minting, redaction commitment salts, signature values, time
//     attestations) -- the closed set cannot silently grow.
//  2. Source-order audit: no non-test .go file in the integrity package (the
//     signature/coverage layer) imports a package that reads an ambient value
//     -- wall-clock time, machine/user identity, filesystem metadata, or
//     process randomness. The signature value itself is a supplied input
//     (an allowlisted site's product), not produced here; the M09 code path
//     introduces no ambient source of its own.
func TestNFR_005_NoAmbientValuesOutsideAllowlistedSites(t *testing.T) {
	// (1) The allowlist is exactly four sites with distinct names.
	if len(AllowlistedAmbientSites) != 4 {
		t.Fatalf("allowlist has %d sites, want exactly 4 (NFR-005)", len(AllowlistedAmbientSites))
	}
	names := map[string]bool{}
	for _, s := range AllowlistedAmbientSites {
		if !IsAllowlistedAmbientSite(s) {
			t.Errorf("site %v not recognised as allowlisted", s)
		}
		if names[s.String()] {
			t.Errorf("duplicate site name %q", s.String())
		}
		names[s.String()] = true
	}
	for _, want := range []string{"identifier-minting", "redaction-commitment-salt", "signature-value", "time-attestation"} {
		if !names[want] {
			t.Errorf("allowlist missing required site %q", want)
		}
	}
	// Anything outside the range is not allowlisted.
	if IsAllowlistedAmbientSite(AmbientSite(len(AllowlistedAmbientSites))) {
		t.Errorf("an out-of-range site was reported as allowlisted")
	}

	// (2) Source-order audit over the integrity package's non-test sources.
	// The forbidden IMPORTS are ambient-value sources with no legitimate
	// deterministic use here: user/machine identity, filesystem metadata, and
	// process randomness. The "time" package is NOT import-banned, because a
	// value conversion of a STORED instant (time.Unix(storedSeconds, 0)) is
	// not a wall-clock read; instead the wall-clock READS themselves
	// (time.Now / time.Since / time.Until) are banned at the call level below,
	// which is the precise NFR-005 concern ("values derived from wall-clock
	// time").
	forbidden := map[string]string{
		"os/user":      "user identity",
		"os":           "filesystem metadata / environment",
		"crypto/rand":  "process randomness",
		"math/rand":    "process randomness",
		"math/rand/v2": "process randomness",
	}
	// wallClockReads are the ambient time READS (a value derived from the
	// wall clock), banned even though the time package's value types are fine.
	wallClockReads := []string{"time.Now", "time.Since", "time.Until"}
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read integrity package dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		af, err := parser.ParseFile(fset, name, src, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range af.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if reason, bad := forbidden[path]; bad {
				t.Errorf("%s imports %q (%s); the signature/coverage layer must introduce no ambient value outside the four allowlisted sites (NFR-005)", name, path, reason)
			}
		}
		// Ban wall-clock reads at the call level (a value derived from the
		// wall clock), independent of the time package's value-type use.
		for _, read := range wallClockReads {
			if strings.Contains(string(src), read+"(") {
				t.Errorf("%s calls %s (a wall-clock read); NFR-005 excludes values derived from wall-clock time from the octet stream", name, read)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("A-AMBIENT audit scanned no integrity-package source files")
	}
}
