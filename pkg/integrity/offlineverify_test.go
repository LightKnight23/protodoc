package integrity

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestFR_070_TrustAnchorsNoNetworkCalls is T-0171's named unit test (FR-070).
// LTV verification uses a FIXED, locally-held trust-anchor list with ALL
// network interfaces disabled. This asserts: (1) an empty trust-anchor pool
// trusts nothing (closed-world, no fallback), (2) building anchors never
// consults the network or system store, and (3) a source audit confirms the
// integrity package's LTV verification imports no networking package -- so a
// verification decades hence never depends on a live query.
func TestFR_070_TrustAnchorsNoNetworkCalls(t *testing.T) {
	// (1) Offline verification with an empty pool refuses (nothing trusted).
	empty := TrustAnchors{Roots: nil}
	if _, err := verifyLeafOffline(nil, OfflineVerifyOptions{Anchors: empty}); !errors.Is(err, ErrNoTrustAnchors) {
		t.Errorf("empty trust-anchor pool: err = %v, want ErrNoTrustAnchors", err)
	}

	// (2) NewTrustAnchors with no roots yields a non-nil but empty pool, built
	// purely from the (here zero) supplied DER roots -- no system store.
	ta, err := NewTrustAnchors(nil)
	if err != nil {
		t.Fatalf("NewTrustAnchors(nil): %v", err)
	}
	if ta.Roots == nil {
		t.Error("NewTrustAnchors should return an initialised (empty) pool")
	}
	// A garbage DER root is rejected (parsed locally, not fetched).
	if _, err := NewTrustAnchors([][]byte{{0x00, 0x01, 0x02}}); err == nil {
		t.Error("NewTrustAnchors accepted an unparseable trust anchor")
	}

	// (3) Source audit: no non-test integrity file imports a networking
	// package. LTV verification must be entirely offline.
	forbiddenNet := map[string]bool{
		"net": true, "net/http": true, "net/url": true, "net/rpc": true,
		"crypto/tls": true, "golang.org/x/net": true,
	}
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
		af, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range af.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if forbiddenNet[path] || strings.HasPrefix(path, "net/") {
				t.Errorf("%s imports networking package %q; LTV verification must be offline (FR-070)", name, path)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no-network audit scanned no integrity-package source files")
	}
}
