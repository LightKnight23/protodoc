// RLE op-batch encoding (T-0221, NFR-032; document.abnf S9 hob-operations). A
// text-dominated editing trace produces long runs of operations that share
// their fixed-width prefix fields -- a burst of typing shares one authoring
// state (op-order-key) and predecessor, and often one target text block. The
// straightforward per-op encoding repeats those 80 fixed octets (16 target +
// 32 predecessor + 32 order-key) for every operation, which for a 250,000-op /
// 100,000-char document blows NFR-032's 2.0x budget by orders of magnitude
// (the "naive encoding" the spec's Why clause calls out at >1,300 octets per
// character). This run-length encoding emits each shared prefix field ONCE per
// run and repeats only the per-op payload, recovering the budget. It is a
// size-oriented alternate serialization of the SAME operation sequence; the
// plain EncodeOpBatch remains the canonical wire form.
package history

import (
	"errors"
	"fmt"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// EncodeOpBatchRLE encodes ops as a run-length-shared batch. Consecutive
// operations that share (target, predecessor, order-key) form one run whose
// shared prefix is written once. Within a run, if every op additionally shares
// the same kind AND the same payload length, the kind and payload-length are
// written ONCE and the payloads concatenated (a "uniform run", the typing-burst
// case: many single-character same-kind insertions), so per-op framing
// collapses to just the payload bytes -- matching the visible content's own
// per-character cost and meeting NFR-032's 2.0x budget. A non-uniform run falls
// back to per-op (kind, length, payload). DecodeOpBatchRLE inverts both forms.
func EncodeOpBatchRLE(ops []OperationRecord) ([]byte, error) {
	var out []byte
	out = pdlfmt.AppendVarint(out, uint64(len(ops)))
	i := 0
	for i < len(ops) {
		j := i + 1
		for j < len(ops) && sharesPrefix(ops[i], ops[j]) {
			j++
		}
		runLen := j - i
		if !ops[i].Kind.valid() {
			return nil, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(ops[i].Kind))
		}
		out = append(out, ops[i].Target[:]...)
		out = append(out, ops[i].Predecessor[:]...)
		out = append(out, ops[i].OrderKey[:]...)
		out = pdlfmt.AppendVarint(out, uint64(runLen))

		uniform, plen := uniformRun(ops[i:j])
		if uniform {
			out = append(out, rleUniformFlag)
			out = append(out, byte(ops[i].Kind))
			out = pdlfmt.AppendVarint(out, uint64(plen))
			for k := i; k < j; k++ {
				out = append(out, ops[k].Payload...)
			}
		} else {
			out = append(out, rlePerOpFlag)
			for k := i; k < j; k++ {
				if !ops[k].Kind.valid() {
					return nil, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(ops[k].Kind))
				}
				out = append(out, byte(ops[k].Kind))
				out = pdlfmt.AppendVarint(out, uint64(len(ops[k].Payload)))
				out = append(out, ops[k].Payload...)
			}
		}
		i = j
	}
	return out, nil
}

// rle run-encoding flags.
const (
	rlePerOpFlag   = 0x00
	rleUniformFlag = 0x01
)

// uniformRun reports whether every op in run shares the same kind and payload
// length, and returns that common payload length.
func uniformRun(run []OperationRecord) (bool, int) {
	if len(run) == 0 {
		return false, 0
	}
	k := run[0].Kind
	pl := len(run[0].Payload)
	for _, op := range run {
		if op.Kind != k || len(op.Payload) != pl {
			return false, 0
		}
	}
	return true, pl
}

func sharesPrefix(a, b OperationRecord) bool {
	return a.Target == b.Target && a.Predecessor == b.Predecessor && a.OrderKey == b.OrderKey
}

// ErrRLETruncated is returned for a truncated RLE batch.
var ErrRLETruncated = errors.New("history: RLE op batch truncated")

// DecodeOpBatchRLE inverts EncodeOpBatchRLE, reproducing the exact operation
// sequence.
func DecodeOpBatchRLE(src []byte) ([]OperationRecord, error) {
	total, n, err := pdlfmt.DecodeVarint(src)
	if err != nil {
		return nil, fmt.Errorf("history: RLE total count: %w", err)
	}
	pos := n
	// `total` is an operation COUNT, not a byte length. A uniform run can
	// encode many ops in few octets (shared payload), so total may exceed
	// len(src); do NOT pre-allocate from the untrusted total. Bound total
	// against the decoded-unit ceiling (a document cannot declare more ops than
	// that many octets could ever describe even at zero payload), then let the
	// decode loop's per-run truncation checks bound the actual work (FR-106).
	if total > MaxOpBatchOctets() {
		return nil, fmt.Errorf("%w: declared op count %d exceeds the decoded-unit ceiling", ErrRLETruncated, total)
	}
	ops := make([]OperationRecord, 0)
	for uint64(len(ops)) < total {
		if pos+16+32+32 > len(src) {
			return nil, ErrRLETruncated
		}
		var target pdlfmt.UnitID
		var pred, ok container.StateID
		copy(target[:], src[pos:pos+16])
		pos += 16
		copy(pred[:], src[pos:pos+32])
		pos += 32
		copy(ok[:], src[pos:pos+32])
		pos += 32
		runLen, rn, err := pdlfmt.DecodeVarint(src[pos:])
		if err != nil {
			return nil, fmt.Errorf("history: RLE run length: %w", err)
		}
		pos += rn
		// Bound runLen so the cumulative op count never exceeds the declared
		// total: a run cannot claim more ops than remain to be produced. This
		// caps a hostile huge runLen (especially with a zero-length uniform
		// payload) against the declared total, which is itself bounded below.
		if runLen > total-uint64(len(ops)) {
			return nil, fmt.Errorf("%w: run length %d exceeds remaining op budget", ErrRLETruncated, runLen)
		}
		if pos >= len(src) {
			return nil, ErrRLETruncated
		}
		flag := src[pos]
		pos++
		switch flag {
		case rleUniformFlag:
			if pos >= len(src) {
				return nil, ErrRLETruncated
			}
			kind := OpKind(src[pos])
			if !kind.valid() {
				return nil, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(kind))
			}
			pos++
			plen, pn, err := pdlfmt.DecodeVarint(src[pos:])
			if err != nil {
				return nil, fmt.Errorf("history: RLE uniform payload length: %w", err)
			}
			pos += pn
			need := int(runLen) * int(plen)
			if len(src)-pos < need {
				return nil, ErrRLETruncated
			}
			for r := uint64(0); r < runLen; r++ {
				op := OperationRecord{Kind: kind, Target: target, Predecessor: pred, OrderKey: ok}
				if plen > 0 {
					op.Payload = append([]byte(nil), src[pos:pos+int(plen)]...)
					pos += int(plen)
				}
				ops = append(ops, op)
			}
		case rlePerOpFlag:
			for r := uint64(0); r < runLen; r++ {
				if pos >= len(src) {
					return nil, ErrRLETruncated
				}
				kind := OpKind(src[pos])
				if !kind.valid() {
					return nil, fmt.Errorf("%w: 0x%02x", ErrOpKindReserved, uint8(kind))
				}
				pos++
				plen, pn, err := pdlfmt.DecodeVarint(src[pos:])
				if err != nil {
					return nil, fmt.Errorf("history: RLE payload length: %w", err)
				}
				pos += pn
				if uint64(len(src)-pos) < plen {
					return nil, ErrRLETruncated
				}
				op := OperationRecord{Kind: kind, Target: target, Predecessor: pred, OrderKey: ok}
				if plen > 0 {
					op.Payload = append([]byte(nil), src[pos:pos+int(plen)]...)
				}
				pos += int(plen)
				ops = append(ops, op)
			}
		default:
			return nil, fmt.Errorf("history: RLE run flag 0x%02x invalid", flag)
		}
	}
	return ops, nil
}
