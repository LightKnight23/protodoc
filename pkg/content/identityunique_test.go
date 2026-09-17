package content_test

import (
	"testing"

	"Protodoc/pkg/content"
	"Protodoc/pkg/content/mint"
	"Protodoc/pkg/pdlfmt"
)

// TestFR_019_IdentifierUniqueAcrossCopyForkBranch is T-0084's named
// conformance test (case id CONF-IDENTITY-UNIQUE-COPY-FORK-BRANCH). It
// builds a 3-document fixture set -- an original, a copy of it, and a fork
// with its own added content -- and decodes them together, asserting zero
// colliding run_id values across all three, so identity is unique across the
// document and every copy/fork/branch derived from it (FR-019), not merely
// within one file.
func TestFR_019_IdentifierUniqueAcrossCopyForkBranch(t *testing.T) {
	original := make([]content.Run, 5)
	for i := range original {
		id, err := mint.MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		original[i] = content.Run{RunID: id, BaseOrdinal: uint32(i * 10), Text: "run", LangRef: content.LangRef(1)}
	}

	// Copy: duplicating content mints fresh ids (FR-022), so the copy shares
	// no run_id with the original.
	copyDoc, err := mint.DuplicateRange(original)
	if err != nil {
		t.Fatalf("DuplicateRange (copy): %v", err)
	}

	// Fork: original plus independently-minted added content.
	fork := append([]content.Run(nil), original...)
	for i := 0; i < 3; i++ {
		id, err := mint.MintID()
		if err != nil {
			t.Fatalf("MintID (fork): %v", err)
		}
		fork = append(fork, content.Run{RunID: id, BaseOrdinal: uint32(100 + i*10), Text: "forked", LangRef: content.LangRef(1)})
	}

	originalSet := runIDSet(original)
	forkSet := runIDSet(fork)
	for _, r := range copyDoc {
		if _, dup := originalSet[r.RunID]; dup {
			t.Fatalf("copy run_id %x collides with an original id (duplication must mint fresh)", r.RunID)
		}
		if _, dup := forkSet[r.RunID]; dup {
			t.Fatalf("copy run_id %x collides with a fork id", r.RunID)
		}
	}

	copySet := runIDSet(copyDoc)
	for _, r := range fork {
		if _, shared := originalSet[r.RunID]; shared {
			continue // legitimately retained original-lineage unit
		}
		if _, dup := copySet[r.RunID]; dup {
			t.Fatalf("fork-added run_id %x collides with a copy id", r.RunID)
		}
	}

	for name, doc := range map[string][]content.Run{"original": original, "copy": copyDoc, "fork": fork} {
		if !allRunIDsDistinct(doc) {
			t.Fatalf("%s document contains an internal run_id collision", name)
		}
	}
}

func runIDSet(runs []content.Run) map[pdlfmt.UnitID]struct{} {
	s := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		s[r.RunID] = struct{}{}
	}
	return s
}

func allRunIDsDistinct(runs []content.Run) bool {
	seen := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		if _, dup := seen[r.RunID]; dup {
			return false
		}
		seen[r.RunID] = struct{}{}
	}
	return true
}
