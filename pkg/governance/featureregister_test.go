package governance

import (
	"testing"

	"Protodoc/pkg/diffconform"
)

// TestCON_019_FeatureNormativeStatusGatedOnTwoImplPass is T-0354's named
// test. It asserts: every named feature/mechanism from data-model.md
// starts Provisional; a feature with only single-implementation coverage
// (same ImplA/ImplB, or a report the two implementations disagree on)
// remains Provisional; and a feature only flips to Normative once genuine
// two-source-independent-implementation, 100%-pass evidence is recorded
// for it -- and even then, every other feature in the register is
// unaffected.
func TestCON_019_FeatureNormativeStatusGatedOnTwoImplPass(t *testing.T) {
	names, err := LoadFeatureNames(dataModelPath)
	if err != nil {
		t.Fatalf("LoadFeatureNames(%s): %v", dataModelPath, err)
	}
	if len(names) < 2 {
		t.Fatalf("LoadFeatureNames(%s) = %v, want at least 2 features to gate independently", dataModelPath, names)
	}

	reg := NewRegister(names)
	for _, n := range names {
		if reg[n] != StatusProvisional {
			t.Fatalf("freshly built register: feature %q = %q, want %q", n, reg[n], StatusProvisional)
		}
	}

	target := names[0]
	untouched := names[1]

	passingReport := diffconform.Report{
		CorpusVersion: "TEST-CORPUS-V1",
		Total:         3,
	}
	failingReport := diffconform.Report{
		CorpusVersion: "TEST-CORPUS-V1",
		Total:         3,
		Diffs: []diffconform.Diff{
			{Case: "case-2", VerdictMismatch: true},
		},
	}

	t.Run("single implementation coverage stays provisional", func(t *testing.T) {
		err := reg.RecordTwoImplPass(TwoImplPassEvidence{
			Feature: target,
			ImplA:   "reference-go",
			ImplB:   "reference-go", // same implementation, not independent
			Report:  passingReport,
		})
		if err == nil {
			t.Fatal("RecordTwoImplPass with ImplA == ImplB: got nil error, want refusal")
		}
		if reg[target] != StatusProvisional {
			t.Fatalf("feature %q after refused single-impl evidence = %q, want %q", target, reg[target], StatusProvisional)
		}
	})

	t.Run("disagreeing second implementation stays provisional", func(t *testing.T) {
		err := reg.RecordTwoImplPass(TwoImplPassEvidence{
			Feature: target,
			ImplA:   "reference-go",
			ImplB:   "second-impl-rust",
			Report:  failingReport,
		})
		if err == nil {
			t.Fatal("RecordTwoImplPass with a mismatched report: got nil error, want refusal")
		}
		if reg[target] != StatusProvisional {
			t.Fatalf("feature %q after refused mismatched evidence = %q, want %q", target, reg[target], StatusProvisional)
		}
	})

	t.Run("unknown feature is refused", func(t *testing.T) {
		if err := reg.RecordTwoImplPass(TwoImplPassEvidence{
			Feature: "NotARealFeature",
			ImplA:   "reference-go",
			ImplB:   "second-impl-rust",
			Report:  passingReport,
		}); err == nil {
			t.Fatal("RecordTwoImplPass for an unregistered feature: got nil error, want refusal")
		}
	})

	t.Run("genuine two-independent-implementation pass flips to normative", func(t *testing.T) {
		if err := reg.RecordTwoImplPass(TwoImplPassEvidence{
			Feature: target,
			ImplA:   "reference-go",
			ImplB:   "second-impl-rust",
			Report:  passingReport,
		}); err != nil {
			t.Fatalf("RecordTwoImplPass with genuine two-impl pass evidence: %v", err)
		}
		if reg[target] != StatusNormative {
			t.Fatalf("feature %q after genuine two-impl pass = %q, want %q", target, reg[target], StatusNormative)
		}
	})

	t.Run("promoting one feature leaves others provisional", func(t *testing.T) {
		if reg[untouched] != StatusProvisional {
			t.Fatalf("unrelated feature %q after target's promotion = %q, want %q (still single-implementation coverage)", untouched, reg[untouched], StatusProvisional)
		}
	})
}
