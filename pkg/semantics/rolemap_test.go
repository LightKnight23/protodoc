package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFR_038_RoleMapCompleteness is T-0279's named unit test (FR-038). It
// proves the accessibility-role map is (a) total — every assigned construct
// kind in document.abnf's registry maps to a role — and (b) single-valued —
// each kind maps to exactly one role — and (c) that data-model.md publishes
// the same closed role table.
func TestFR_038_RoleMapCompleteness(t *testing.T) {
	// (a) Totality: every assigned construct kind resolves to a role.
	for _, k := range AssignedConstructKinds {
		if _, ok := RoleOf(k); !ok {
			t.Errorf("construct kind 0x%02X has no role mapping", byte(k))
		}
	}

	// (b) Single-valued: roleMap keys are exactly AssignedConstructKinds, with
	// no kind mapped more than once (a Go map guarantees one value per key; we
	// assert the key set matches the closed construct set).
	if len(roleMap) != len(AssignedConstructKinds) {
		t.Errorf("roleMap has %d entries, want %d (one per assigned construct kind)", len(roleMap), len(AssignedConstructKinds))
	}
	for _, k := range AssignedConstructKinds {
		if _, ok := roleMap[k]; !ok {
			t.Errorf("roleMap missing assigned kind 0x%02X", byte(k))
		}
	}

	// (c) data-model.md publishes the role table naming every construct kind.
	dm, err := os.ReadFile(filepath.Join("..", "..", "specs", "001-protodoc-format-core", "data-model.md"))
	if err != nil {
		t.Fatalf("read data-model.md: %v", err)
	}
	doc := string(dm)
	for _, name := range []string{
		"TEXT_BLOCK", "ANNOTATION", "TABLE", "NOTE", "CROSS_REFERENCE",
		"EXT_ENVELOPE", "RASTER_IMAGE", "FONT_SUBSET", "REGISTRY_EXCERPT",
		"UNIT_INDEX_LEAF", "HISTORY_OP_BATCH", "PRESENTATION_ARTEFACT",
		"ERASURE_RECORD", "ROOT_SEQUENCE",
	} {
		if !strings.Contains(doc, name) {
			t.Errorf("data-model.md role table does not name construct kind %s", name)
		}
	}
	if !strings.Contains(doc, "accessibility role") && !strings.Contains(doc, "Accessibility role") {
		t.Errorf("data-model.md does not publish an accessibility role table")
	}
}
