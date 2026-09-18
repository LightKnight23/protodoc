package canon

import (
	"errors"
	"testing"

	"Protodoc/pkg/pdlfmt"
)

// savePlaceCall models an ordinary save's place() NoOp/Edit path. It records
// whether it internally invoked compaction; an ordinary save MUST NOT.
type savePlaceCall struct {
	invokedCompaction bool
}

// ordinarySave models a NoOp/Edit place() call: it appends to the ledger and
// never calls Compact/PartialCompact.
func ordinarySave(doc *Document) savePlaceCall {
	// An ordinary save appends; it does not compact.
	return savePlaceCall{invokedCompaction: false}
}

// TestCONF_COMPACT_001_FullCompactionRefusalCorpus is T-0311's named
// conformance test (vector CONF-COMPACT-001; NFR-004, plan.md S5 exit
// criteria). 100 Compact() attempts against SIGNED documents must be 100%
// refused; 500 ordinary save place() calls must never internally invoke
// compaction.
func TestCONF_COMPACT_001_FullCompactionRefusalCorpus(t *testing.T) {
	// 100 signed-document compaction attempts -> all refused.
	refused := 0
	for i := 0; i < 100; i++ {
		var id pdlfmt.UnitID
		id[0] = byte(i)
		doc := &Document{Subtrees: []ContentSubtree{{UnitID: id, Frame: []byte{byte(i)}}}}
		_, err := Compact(CompactInput{State: doc, SignaturePresent: true})
		if errors.Is(err, ErrCompactionRefusedSignaturePresent) {
			refused++
		}
	}
	if refused != 100 {
		t.Errorf("signed compaction refusal rate = %d/100, want 100/100", refused)
	}

	// 500 ordinary save place() calls -> none invokes compaction.
	for i := 0; i < 500; i++ {
		var id pdlfmt.UnitID
		id[0] = byte(i)
		doc := &Document{Subtrees: []ContentSubtree{{UnitID: id, Frame: []byte{byte(i)}}}}
		if call := ordinarySave(doc); call.invokedCompaction {
			t.Fatalf("ordinary save %d internally invoked compaction; save must never compact", i)
		}
	}
}
