package merge

import (
	"errors"
	"testing"

	"Protodoc/pkg/container"
)

// TestFR_116_RecordsSupersededStateAndResolution is T-0238's named unit test
// (FR-116). A state that supersedes a state which is NOT among its recorded
// ancestors must record the superseded state's identifier together with a
// resolution value; otherwise the supersession is refused as unrecorded
// (making a silent overwrite detectable from the file alone). Superseding a
// recorded ancestor is ordinary linear progression and needs no record.
func TestFR_116_RecordsSupersededStateAndResolution(t *testing.T) {
	nonAncestor := container.StateID{0xC0, 0x11}
	ancestor := container.StateID{0xA1}
	ancestors := map[container.StateID]bool{ancestor: true}

	// Supersede a non-ancestor WITHOUT a record -> refused, names the state.
	err := RequireSupersessionRecord(nonAncestor, ancestors, map[container.StateID]SupersessionRecord{})
	if !errors.Is(err, ErrUnrecordedSupersession) {
		t.Fatalf("unrecorded: err = %v, want ErrUnrecordedSupersession", err)
	}
	var ue *UnrecordedSupersessionError
	if !errors.As(err, &ue) || ue.Superseded != nonAncestor {
		t.Fatalf("error must name superseded state %x, got %v", nonAncestor, err)
	}

	// Supersede a non-ancestor WITH a record (id + resolution) -> permitted,
	// and the record carries both the superseded id and a resolution value.
	rec := SupersessionRecord{Superseded: nonAncestor, Resolution: ResolutionTiebreak}
	records := map[container.StateID]SupersessionRecord{nonAncestor: rec}
	if err := RequireSupersessionRecord(nonAncestor, ancestors, records); err != nil {
		t.Errorf("recorded supersession should be permitted, got %v", err)
	}
	if records[nonAncestor].Superseded != nonAncestor {
		t.Error("record must carry the superseded state's identifier")
	}
	if records[nonAncestor].Resolution != ResolutionTiebreak {
		t.Error("record must carry the resolution value")
	}

	// Supersede a recorded ancestor -> ordinary progression, no record needed.
	if err := RequireSupersessionRecord(ancestor, ancestors, map[container.StateID]SupersessionRecord{}); err != nil {
		t.Errorf("superseding an ancestor should need no record, got %v", err)
	}
}
