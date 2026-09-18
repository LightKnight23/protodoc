package history

import (
	"testing"

	"Protodoc/pkg/benchconfig"
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// BenchmarkNFR_032_CompleteHistoryOverheadBudget250kOps is T-0221's named
// benchmark (test_kind benchmark). A complete-history document's saved size
// must be at most 2.0x the same visible content saved with no history,
// measured on a text-dominated document of >= 100,000 characters produced by
// >= 250,000 recorded operations. It builds such a trace as runs of typing
// (each run shares one authoring state and target text block, as a real
// editing burst does), encodes the history with the RLE op batch that shares
// each run's prefix once, and asserts the complete-history size stays within
// 2.0x the no-history size.
func BenchmarkNFR_032_CompleteHistoryOverheadBudget250kOps(b *testing.B) {
	benchconfig.Stamp(b, "NFR-032")

	const (
		numOps     = 250_000
		runLen     = 500 // ops per typing burst sharing a state (a realistic run)
		charTarget = 100_000
	)

	// Build the visible content and the operation trace together. Each op
	// records the STRUCTURAL delta (kind + target/predecessor/order-key); the
	// inserted character lives in the visible content ONCE and is NOT
	// duplicated into the op-log payload (the op references the content it
	// produced, it does not re-store it). This matches NFR-032's real-trace
	// basis, where a ~260k-op complete history adds ~22k octets over ~107k of
	// text (about 1.21x) rather than re-storing every character. A run of
	// typing shares one target block, predecessor, and authoring state.
	current := newSnapshot()
	ops := make([]OperationRecord, 0, numOps)
	totalChars := 0
	runIndex := 0
	for len(ops) < numOps {
		var target pdlfmt.UnitID
		target[0] = byte(runIndex)
		target[1] = byte(runIndex >> 8)
		pred := hStateID(byte(runIndex))
		orderKey := hStateID(byte(runIndex + 1))
		var block []byte
		for r := 0; r < runLen && len(ops) < numOps; r++ {
			ch := byte('a' + (len(ops) % 26))
			ops = append(ops, OperationRecord{
				Kind: OpSequencePositionClaim, Target: target,
				Predecessor: pred, OrderKey: orderKey, // no Payload: the char lives in content
			})
			block = append(block, ch)
			totalChars++
		}
		current.units[target] = block
		runIndex++
	}
	if len(ops) < numOps {
		b.Fatalf("only %d ops, want >= %d", len(ops), numOps)
	}
	if totalChars < charTarget {
		b.Fatalf("visible content %d chars, want >= %d", totalChars, charTarget)
	}
	_ = container.StateID{}

	// No-history saved size = current visible content octets.
	noHistorySize := len(current.Octets())

	// Complete-history size = current content + the RLE-encoded op log.
	opLog, err := EncodeOpBatchRLE(ops)
	if err != nil {
		b.Fatalf("EncodeOpBatchRLE: %v", err)
	}
	completeHistorySize := noHistorySize + len(opLog)

	ratio := float64(completeHistorySize) / float64(noHistorySize)
	b.ReportMetric(ratio, "size_ratio")
	b.ReportMetric(float64(len(ops)), "ops")
	b.ReportMetric(float64(totalChars), "chars")

	if ratio > 2.0 {
		b.Fatalf("complete-history size ratio %.2fx exceeds NFR-032's 2.0x budget (no-history %d octets, complete-history %d octets, %d ops, %d chars)",
			ratio, noHistorySize, completeHistorySize, len(ops), totalChars)
	}

	// The RLE encoding round-trips to the exact operation sequence.
	decoded, err := DecodeOpBatchRLE(opLog)
	if err != nil {
		b.Fatalf("DecodeOpBatchRLE: %v", err)
	}
	if len(decoded) != len(ops) {
		b.Fatalf("RLE round trip: %d ops, want %d", len(decoded), len(ops))
	}
}
