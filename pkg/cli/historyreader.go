// HISTORY-segment record reader (T-0393; CON-024/FR-096). Mirrors
// attestreader.go's ATTEST-segment convention: each HISTORY segment carries
// exactly one ERASURE_RECORD (integrity.abnf S7.2), optionally framed behind a
// ledger segment header. This is the real decode path merge needs to check the
// retention-point and erased-unit-replay guards; no other verb touches HISTORY
// payload bytes yet. Go stdlib only.
package cli

import (
	"fmt"
	"io"

	"Protodoc/pkg/container"
	"Protodoc/pkg/history"
	"Protodoc/pkg/ledger"
)

// historySegmentBody strips the leading ledger segment header (if present)
// from a HISTORY segment's raw octets, returning the record body — the same
// framing rule attestSegmentBody applies to ATTEST segments.
func historySegmentBody(raw []byte) []byte {
	if len(raw) >= 4 && raw[0] == 'P' && raw[1] == 'D' && raw[2] == 'S' && raw[3] == '1' {
		if len(raw) >= ledger.SegmentHeaderSize {
			return raw[ledger.SegmentHeaderSize:]
		}
	}
	return raw
}

// DiscoverErasureRecords scans every HISTORY segment of the document at r and
// decodes its ERASURE_RECORD (FR-061), in segment-table order.
func DiscoverErasureRecords(r io.ReaderAt, table [container.MaxSegments]container.SegmentTableSlot) ([]history.ErasureRecord, error) {
	var out []history.ErasureRecord
	for _, slot := range table {
		if slot.SegmentType != container.SegmentTypeHistory {
			continue
		}
		raw := make([]byte, slot.Length)
		if _, err := r.ReadAt(raw, int64(slot.Offset)); err != nil {
			return nil, fmt.Errorf("cli: reading HISTORY segment at %d: %w", slot.Offset, err)
		}
		rec, err := history.DecodeErasureRecord(historySegmentBody(raw))
		if err != nil {
			return nil, fmt.Errorf("cli: decoding ERASURE_RECORD: %w", err)
		}
		out = append(out, rec)
	}
	return out, nil
}
