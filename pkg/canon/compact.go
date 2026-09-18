// Full compaction (NFR-004; T-0310). Compact performs FULL compaction as
// place(BOTTOM, S) = L*(S): it rebuilds the ledger from the canonical sequence
// C(S), discarding all superseded history. Because full compaction changes the
// physical byte layout of segments a signature may cover, it is REFUSED
// (before any I/O) whenever any Signature record is present — a signature and
// full compaction are mutually exclusive.
package canon

import (
	"bytes"
	"errors"
)

// File is the compacted output: the canonical octet sequence of the state,
// rebuilt from BOTTOM. (A thin wrapper so callers see an explicit compaction
// result type distinct from a raw byte slice.)
type File struct {
	Bytes []byte
}

// ErrCompactionRefusedSignaturePresent is the named refusal returned when a
// full compaction is attempted on a document carrying any Signature record.
var ErrCompactionRefusedSignaturePresent = errors.New("canon: full compaction refused — document carries a signature (NFR-004)")

// CompactInput is the document state plus whether it carries any Signature
// record. SignaturePresent is checked BEFORE any I/O.
type CompactInput struct {
	State            *Document
	SignaturePresent bool
}

// Compact performs full compaction: place(BOTTOM, S) = L*(S). If any Signature
// record is present it returns ErrCompactionRefusedSignaturePresent BEFORE
// producing any output (no I/O, no allocation of the compacted file). Otherwise
// it returns the canonical octet sequence C(S) as the compacted File.
func Compact(in CompactInput) (*File, error) {
	if in.SignaturePresent {
		return nil, ErrCompactionRefusedSignaturePresent
	}
	var buf bytes.Buffer
	if err := Canonicalize(in.State, &buf); err != nil {
		return nil, err
	}
	return &File{Bytes: buf.Bytes()}, nil
}
