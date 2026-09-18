package canon_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNFR_004_CompactionExplicitOnlyNotInvokedBySave is T-0314's named
// integration test (NFR-004). Neither Compact nor PartialCompact is reachable
// from the ledger place() NoOp/Edit save paths: the ledger package must not
// reference the canon compaction API at all, so an ordinary save can never
// internally compact. Compaction is callable ONLY via canon's explicit
// top-level API.
func TestNFR_004_CompactionExplicitOnlyNotInvokedBySave(t *testing.T) {
	ledgerDir := filepath.Join("..", "ledger")
	entries, err := os.ReadDir(ledgerDir)
	if err != nil {
		t.Fatalf("read ledger dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(ledgerDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(raw)
		// The ledger save path must not import or call the canon compaction API.
		if strings.Contains(src, "Protodoc/pkg/canon") {
			t.Errorf("%s imports pkg/canon; the save path must not reach compaction", name)
		}
		for _, sym := range []string{"canon.Compact", "canon.PartialCompact"} {
			if strings.Contains(src, sym) {
				t.Errorf("%s references %s; compaction must be explicit-only, never from save", name, sym)
			}
		}
	}
}
