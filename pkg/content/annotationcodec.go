// Orphaned-annotation wire encode/decode, sufficient to demonstrate the
// FR-027..FR-030 end-to-end round trip through the append-only ledger
// (T-0086): the boundary-behaviour values of both anchors and all four
// orphan-record fields survive a save/reload with zero drift. This is a
// fixed-layout serialization of exactly the fields the exit-criteria vector
// checks; the full PDL-TLV annotation frame shape (document.abnf S3) is a
// later milestone's concern.
package content

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// ErrOrphanAnnotationTruncated is returned when the encoded orphaned
// annotation is shorter than its fixed leading fields require.
var ErrOrphanAnnotationTruncated = errors.New("content: orphaned annotation encoding truncated")

// EncodeOrphanedAnnotation appends a canonical encoding of an orphaned
// annotation to dst: ann-id(16) start-anchor(22) end-anchor(22)
// body-block(16) orphan-author(2) orphan-prev(16) orphan-next(16)
// quoted-len(varint via u32) quoted-bytes. It requires a.Orphaned; the
// quoted text is length-prefixed so it round-trips exactly.
func EncodeOrphanedAnnotation(dst []byte, a Annotation) []byte {
	dst = pdlfmt.AppendUnitID(dst, a.ID)
	dst = EncodeAnchorPoint(dst, a.Start)
	dst = EncodeAnchorPoint(dst, a.End)
	dst = pdlfmt.AppendUnitID(dst, a.BodyBlock)
	var author [2]byte
	binary.BigEndian.PutUint16(author[:], a.Orphan.Author)
	dst = append(dst, author[:]...)
	dst = pdlfmt.AppendUnitID(dst, a.Orphan.Prev)
	dst = pdlfmt.AppendUnitID(dst, a.Orphan.Next)
	var qlen [4]byte
	binary.BigEndian.PutUint32(qlen[:], uint32(len(a.Orphan.QuotedText)))
	dst = append(dst, qlen[:]...)
	dst = append(dst, a.Orphan.QuotedText...)
	return dst
}

// orphanAnnotationFixedLen is the fixed portion before the quoted text:
// 16 + 22 + 22 + 16 + 2 + 16 + 16 + 4 = 114 octets.
const orphanAnnotationFixedLen = 16 + AnchorPointSize + AnchorPointSize + 16 + 2 + 16 + 16 + 4

// DecodeOrphanedAnnotation decodes an orphaned annotation from the leading
// octets of src, returning it, the number of octets consumed, and an error.
// The returned annotation has Orphaned=true. It validates anchor points
// (side + boundary) via DecodeAnchorPoint.
func DecodeOrphanedAnnotation(src []byte) (Annotation, int, error) {
	if len(src) < orphanAnnotationFixedLen {
		return Annotation{}, 0, ErrOrphanAnnotationTruncated
	}
	var a Annotation
	pos := 0
	id, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: ann-id: %w", err)
	}
	a.ID = id
	pos += 16

	start, n, err := DecodeAnchorPoint(src[pos:])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: start anchor: %w", err)
	}
	a.Start = start
	pos += n

	end, n, err := DecodeAnchorPoint(src[pos:])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: end anchor: %w", err)
	}
	a.End = end
	pos += n

	bb, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: body-block: %w", err)
	}
	a.BodyBlock = bb
	pos += 16

	a.Orphaned = true
	a.Orphan.Author = binary.BigEndian.Uint16(src[pos : pos+2])
	pos += 2
	prev, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: orphan-prev: %w", err)
	}
	a.Orphan.Prev = prev
	pos += 16
	next, _, err := pdlfmt.DecodeUnitID(src[pos : pos+16])
	if err != nil {
		return Annotation{}, 0, fmt.Errorf("content: orphan-next: %w", err)
	}
	a.Orphan.Next = next
	pos += 16

	qlen := int(binary.BigEndian.Uint32(src[pos : pos+4]))
	pos += 4
	if qlen < 0 || pos+qlen > len(src) {
		return Annotation{}, 0, ErrOrphanAnnotationTruncated
	}
	a.Orphan.QuotedText = string(src[pos : pos+qlen])
	pos += qlen

	return a, pos, nil
}
