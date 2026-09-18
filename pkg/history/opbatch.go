// HISTORY_OP_BATCH operation log (T-0208, FR-059/FR-060; document.abnf S9). A
// HISTORY-typed segment carries the edit-operation log as a HISTORY_OP_BATCH
// (discriminant 0x0B): a plain sequence of operation-records serving all three
// history modes (CQ-008). Each operation-record is
// op-kind || op-target || op-predecessor || op-order-key || op-payload; the
// batch is stored in DP-015 canonical order (not re-sorted by byte value). The
// RLE-friendly encoding keeps the whole-history overhead within NFR-032's
// budget by not repeating the fixed-width prefix fields when a run of
// operations shares them; the wire form here is the straightforward plain-seq,
// with a run-length helper for the size-sensitive path.
package history

import (
	"errors"
	"fmt"

	"Protodoc/pkg/ceilings"
	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// minOpOctets is the smallest possible encoded operation-record: op-kind(1) +
// op-target(16) + op-predecessor(32) + op-order-key(32) + a 1-octet varint
// zero-length payload = 82 octets. Used to bound a declared op count against
// the remaining input before allocation.
const minOpOctets = 1 + 16 + 32 + 32 + 1

// ErrOpBatchOversized is returned when a HISTORY_OP_BATCH operations field-
// value exceeds the MAX_DECODED_UNIT ceiling.
var ErrOpBatchOversized = errors.New("history: HISTORY_OP_BATCH exceeds the MAX_DECODED_UNIT ceiling")

// MaxOpBatchOctets returns the decoded-unit octet ceiling (MAX_DECODED_UNIT)
// that bounds a HISTORY_OP_BATCH, from the single-source-of-truth ceiling
// table.
func MaxOpBatchOctets() uint64 { return ceilings.MustMax("MAX_DECODED_UNIT") }

// HISTORY_OP_BATCH discriminant + TLV tags (document.abnf S9).
const (
	historyOpBatchDiscriminant = 0x0B
	hobTagDiscriminant         = 0
	hobTagOperations           = 1 // plain-seq-of(operation-record)
)

// OpKind is op-kind (document.abnf S9, DP-015): the closed 3-value operation
// classification.
type OpKind uint8

const (
	OpSequencePositionClaim OpKind = 0x00 // text/table/envelope insertion, annotation creation
	OpValueClaim            OpKind = 0x01 // property set, rename, non-cycle move, resize
	OpDelete                OpKind = 0x02 // delete
)

func (k OpKind) valid() bool { return k <= OpDelete }

// OperationRecord is one edit operation in the log (document.abnf S9).
type OperationRecord struct {
	Kind        OpKind
	Target      pdlfmt.UnitID
	Predecessor container.StateID // the causally preceding state's id
	OrderKey    container.StateID // the authoring state_id (R2 tiebreak)
	Payload     []byte            // kind-specific payload octets
}

var (
	// ErrOpKindReserved is returned for an op-kind outside {0,1,2}.
	ErrOpKindReserved = errors.New("history: op-kind outside the closed set {sequence-position-claim, value-claim, delete}")
	// ErrOpBatchDiscriminant is returned for a wrong hob-discriminant.
	ErrOpBatchDiscriminant = errors.New("history: record discriminant is not HISTORY_OP_BATCH (0x0B)")
	// ErrOpBatchTruncated is returned for a truncated operation record.
	ErrOpBatchTruncated = errors.New("history: operation record truncated")
)

// encodeOperation encodes one operation-record: op-kind(1) op-target(16)
// op-predecessor(32) op-order-key(32) op-payload(varint len + octets).
func encodeOperation(op OperationRecord) ([]byte, error) {
	if !op.Kind.valid() {
		return nil, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(op.Kind))
	}
	var buf []byte
	buf = append(buf, byte(op.Kind))
	buf = append(buf, op.Target[:]...)
	buf = append(buf, op.Predecessor[:]...)
	buf = append(buf, op.OrderKey[:]...)
	buf = pdlfmt.AppendVarint(buf, uint64(len(op.Payload)))
	buf = append(buf, op.Payload...)
	return buf, nil
}

// decodeOperation decodes one operation-record from the start of src, returning
// it and the octets consumed.
func decodeOperation(src []byte) (OperationRecord, int, error) {
	var op OperationRecord
	const fixed = 1 + 16 + 32 + 32
	if len(src) < fixed {
		return op, 0, ErrOpBatchTruncated
	}
	pos := 0
	op.Kind = OpKind(src[pos])
	if !op.Kind.valid() {
		return op, 0, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(op.Kind))
	}
	pos++
	copy(op.Target[:], src[pos:pos+16])
	pos += 16
	copy(op.Predecessor[:], src[pos:pos+32])
	pos += 32
	copy(op.OrderKey[:], src[pos:pos+32])
	pos += 32
	valLen, vn, verr := pdlfmt.DecodeVarint(src[pos:])
	if verr != nil {
		return op, 0, fmt.Errorf("history: op-payload length: %w", verr)
	}
	pos += vn
	if uint64(len(src)-pos) < valLen {
		return op, 0, ErrOpBatchTruncated
	}
	op.Payload = append([]byte(nil), src[pos:pos+int(valLen)]...)
	pos += int(valLen)
	return op, pos, nil
}

// EncodeOpBatch encodes a HISTORY_OP_BATCH: the plain sequence of operation
// records in the given (DP-015-canonical) order, wrapped in the discriminated
// record. Order is preserved verbatim, never re-sorted.
func EncodeOpBatch(ops []OperationRecord) ([]byte, error) {
	elems := make([][]byte, len(ops))
	for i, op := range ops {
		b, err := encodeOperation(op)
		if err != nil {
			return nil, err
		}
		elems[i] = b
	}
	opsValue := pdlfmt.AppendPlainSeq(nil, elems)
	fields := []pdlfmt.Field{
		{Tag: hobTagDiscriminant, Value: []byte{historyOpBatchDiscriminant}},
		{Tag: hobTagOperations, Value: opsValue},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeOpBatch decodes a HISTORY_OP_BATCH, the byte-exact inverse of
// EncodeOpBatch, preserving operation order.
func DecodeOpBatch(src []byte) ([]OperationRecord, error) {
	known := map[byte]bool{hobTagDiscriminant: true, hobTagOperations: true}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return nil, err
	}
	var ops []OperationRecord
	var sawDisc, sawOps bool
	for _, f := range fields {
		switch f.Tag {
		case hobTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != historyOpBatchDiscriminant {
				return nil, ErrOpBatchDiscriminant
			}
			sawDisc = true
		case hobTagOperations:
			// FR-106 / CP-012: the whole operations field-value is bounded by
			// the decoded-unit ceiling, and the declared op count is bounded by
			// the remaining octets (each op is at least minOpOctets), both
			// checked BEFORE any allocation.
			if uint64(len(f.Value)) > MaxOpBatchOctets() {
				return nil, fmt.Errorf("%w: %d octets exceeds MAX_DECODED_UNIT %d", ErrOpBatchOversized, len(f.Value), MaxOpBatchOctets())
			}
			count, n, err := pdlfmt.DecodeSeqCount(f.Value)
			if err != nil {
				return nil, fmt.Errorf("history: op batch count: %w", err)
			}
			pos := n
			if count > uint64((len(f.Value)-pos)/minOpOctets) {
				return nil, fmt.Errorf("%w: op count %d exceeds the %d octets remaining", ErrOpBatchTruncated, count, len(f.Value)-pos)
			}
			for i := uint64(0); i < count; i++ {
				op, consumed, err := decodeOperation(f.Value[pos:])
				if err != nil {
					return nil, err
				}
				ops = append(ops, op)
				pos += consumed
			}
			if pos != len(f.Value) {
				return nil, fmt.Errorf("history: op batch has %d trailing octets", len(f.Value)-pos)
			}
			sawOps = true
		}
	}
	if !sawDisc {
		return nil, ErrOpBatchDiscriminant
	}
	_ = sawOps
	return ops, nil
}
