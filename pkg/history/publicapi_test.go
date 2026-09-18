package history

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// TestHistoryPackage_PublicAPIStableAndTraceable is T-0224's named unit test
// (FR-059/FR-060/CON-022). It pins the history package's public API surface:
// the exported symbols downstream milestones depend on must exist with stable
// signatures, and every non-test source file must carry a requirement-id
// citation (traceability). This fails CI if a public symbol is removed/renamed
// or a source file loses its requirement tie.
func TestHistoryPackage_PublicAPIStableAndTraceable(t *testing.T) {
	// (1) The stable public API is present and callable with its expected
	// signatures (a compile-time contract check).
	var (
		_ func(Mode) (Declaration, error)                                                  = NewDeclaration
		_ func(Mode, uint16) (Declaration, error)                                          = NewDeclarationWithRetentionPoint
		_ func([]OperationRecord) ([]byte, error)                                          = EncodeOpBatch
		_ func([]byte) ([]OperationRecord, error)                                          = DecodeOpBatch
		_ func([]OperationRecord) ([]byte, error)                                          = EncodeOpBatchRLE
		_ func([]byte) ([]OperationRecord, error)                                          = DecodeOpBatchRLE
		_ func([]OperationRecord) Lineage                                                  = NewLineage
		_ func(StateSnapshot, []OperationRecord, container.StateID) (StateSnapshot, error) = Reconstruct
		_ func(uint16, uint16) bool                                                        = StateReconstructable
		_ func(pdlfmt.UnitID, pdlfmt.Digest256, [SaltSize]byte) ErasureRecord              = NewErasureRecord
		_ func([SaltSize]byte, pdlfmt.Digest256) pdlfmt.Digest256                          = SeveranceCommitment
	)
	// Methods on the public types.
	var d Declaration
	_ = d.Mode
	_ = d.CheckImmutable
	_ = d.CheckInPlaceRemoval
	var l Lineage
	_ = l.WalkPredecessors
	_ = l.NearestCommonAncestor

	// The closed history-mode set is exactly three.
	if len(AllModes) != 3 {
		t.Errorf("AllModes has %d entries, want 3", len(AllModes))
	}

	// (2) Traceability: every non-test .go file cites at least one requirement
	// id (FR/NFR/CON/DP/CQ), so the package's code stays tied to the spec.
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read history dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if _, err := parser.ParseFile(fset, name, data, parser.SkipObjectResolution); err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		src := string(data)
		if !citesRequirement(src) {
			t.Errorf("%s cites no requirement id; every history source must stay traceable", name)
		}
	}
	if scanned == 0 {
		t.Fatal("public-API audit scanned no history source files")
	}
}

// citesRequirement reports whether s contains at least one requirement/design
// identifier token (FR-/NFR-/CON-/DP-/CQ-).
func citesRequirement(s string) bool {
	for _, prefix := range []string{"FR-", "NFR-", "CON-", "DP-", "CQ-", "CP-"} {
		if strings.Contains(s, prefix) {
			return true
		}
	}
	return false
}
