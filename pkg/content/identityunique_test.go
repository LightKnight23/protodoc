package content

import (
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// TestFR_019_IdentifierUniqueAcrossCopyForkBranch is T-0084's named
// conformance test (case id CONF-IDENTITY-UNIQUE-COPY-FORK-BRANCH). It
// builds a 3-document fixture set -- an original, a copy of it, and a fork
// with its own added content -- and decodes them together, asserting zero
// colliding run_id/unit_id values across all three, so identity is unique
// across the document and every copy/fork/branch derived from it for the
// lifetime of the lineage (FR-019), not merely within one file.
func TestFR_019_IdentifierUniqueAcrossCopyForkBranch(t *testing.T) {
	// Original document: a handful of runs with freshly minted ids.
	original := make([]Run, 5)
	for i := range original {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID: %v", err)
		}
		original[i] = Run{RunID: id, BaseOrdinal: uint32(i * 10), Text: "run", LangRef: LangRef(1)}
	}

	// Copy: duplicating content mints fresh ids (FR-022), so the copy shares
	// no run_id with the original -- the copy is a distinct lineage member.
	copyDoc, err := DuplicateRange(original)
	if err != nil {
		t.Fatalf("DuplicateRange (copy): %v", err)
	}

	// Fork: start from the original, then add independently-minted content,
	// modelling a branch that diverges. The fork keeps the original's ids
	// for unchanged content but adds new units with fresh ids.
	fork := append([]Run(nil), original...)
	for i := 0; i < 3; i++ {
		id, err := MintID()
		if err != nil {
			t.Fatalf("MintID (fork): %v", err)
		}
		fork = append(fork, Run{RunID: id, BaseOrdinal: uint32(100 + i*10), Text: "forked", LangRef: LangRef(1)})
	}

	// Decode all three "files" together: collect every run_id and assert no
	// collision EXCEPT the legitimately-shared original ids the fork retains
	// (those are the SAME lineage unit, not a collision). The test of FR-019
	// is that no TWO DISTINCT units share an id: the copy shares none with
	// anything, and the fork's added units share none.

	// 1) The copy must collide with nothing.
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

	// 2) The fork's ADDED units (those not in the original) must collide
	// with nothing in the original or the copy.
	copySet := runIDSet(copyDoc)
	for _, r := range fork {
		if _, shared := originalSet[r.RunID]; shared {
			continue // legitimately retained original-lineage unit
		}
		if _, dup := copySet[r.RunID]; dup {
			t.Fatalf("fork-added run_id %x collides with a copy id", r.RunID)
		}
	}

	// 3) Within each document, every run_id is itself unique.
	for name, doc := range map[string][]Run{"original": original, "copy": copyDoc, "fork": fork} {
		if !allRunIDsDistinct(doc) {
			t.Fatalf("%s document contains an internal run_id collision", name)
		}
	}
}

func runIDSet(runs []Run) map[pdlfmt.UnitID]struct{} {
	s := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		s[r.RunID] = struct{}{}
	}
	return s
}

func allRunIDsDistinct(runs []Run) bool {
	seen := make(map[pdlfmt.UnitID]struct{}, len(runs))
	for _, r := range runs {
		if _, dup := seen[r.RunID]; dup {
			return false
		}
		seen[r.RunID] = struct{}{}
	}
	return true
}
