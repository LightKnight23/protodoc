// Streaming, early-emission, abandonable extraction (T-0094, FR-047; T-0095,
// FR-048). The extractor is pull-based: the caller receives the complete
// text of unit N -- read from unit N's segment only -- before any octet of
// unit N+1's segment is read from storage. A caller may stop at any point
// (break the loop or cancel), after which the extractor performs zero
// further reads: there is no remainder-read obligation.
package extract

import (
	"fmt"
	"io"
)

// TextUnit is one extracted unit: its CONTENT segment placement and the text
// decoded from that segment. Text is produced by the caller-supplied
// DecodeUnit from the segment's octets.
type TextUnit struct {
	Segment ContentSegment
	Text    string
}

// DecodeUnit turns one CONTENT segment's octets into its extracted text. It
// is supplied by the caller (the frame/run decode is a later milestone's
// concern); the extractor only guarantees WHEN it is invoked -- lazily, one
// segment at a time, in storage order.
type DecodeUnit func(seg ContentSegment, octets []byte) (string, error)

// Extractor is a pull-based streaming extractor. Each call to Next reads and
// decodes exactly one CONTENT segment (the next in storage order) and
// returns its TextUnit; it reads no later segment's octets until the caller
// pulls again. Once Next returns ok=false (exhausted) or the caller stops
// pulling, no further reads occur.
type Extractor struct {
	r        io.ReaderAt
	decode   DecodeUnit
	segs     []ContentSegment
	next     int
	err      error
	finished bool
}

// NewExtractor builds a streaming extractor over r. It reads only the fixed
// prefix up front (to learn the CONTENT segment inventory); it reads NO
// segment payload until Next is pulled.
func NewExtractor(r io.ReaderAt, decode DecodeUnit) (*Extractor, error) {
	segs, err := ContentSegments(r)
	if err != nil {
		return nil, err
	}
	return &Extractor{r: r, decode: decode, segs: segs}, nil
}

// Next pulls the next text unit. It returns the unit and ok=true when one was
// produced, or ok=false when the stream is exhausted or a prior error
// occurred. It reads exactly the one segment it returns -- never a later
// one -- so a caller holding unit N's text is guaranteed no octet of unit
// N+1's segment has been read yet (FR-047).
func (e *Extractor) Next() (TextUnit, bool) {
	if e.finished || e.err != nil || e.next >= len(e.segs) {
		return TextUnit{}, false
	}
	seg := e.segs[e.next]
	e.next++

	octets := make([]byte, seg.Length)
	n, err := e.r.ReadAt(octets, int64(seg.Offset))
	if err != nil && err != io.EOF {
		e.err = fmt.Errorf("extract: reading segment ordinal %d: %w", seg.Ordinal, err)
		return TextUnit{}, false
	}
	if uint64(n) < seg.Length {
		e.err = fmt.Errorf("extract: segment ordinal %d truncated: read %d of %d", seg.Ordinal, n, seg.Length)
		return TextUnit{}, false
	}
	text, derr := e.decode(seg, octets)
	if derr != nil {
		e.err = fmt.Errorf("extract: decoding segment ordinal %d: %w", seg.Ordinal, derr)
		return TextUnit{}, false
	}
	return TextUnit{Segment: seg, Text: text}, true
}

// Stop marks the extractor finished so subsequent Next calls read nothing.
// Calling Stop after consuming a prefix of the units guarantees zero further
// reads (FR-048): abandonment has no remainder-read obligation.
func (e *Extractor) Stop() { e.finished = true }

// Err returns the first error the extractor encountered, if any.
func (e *Extractor) Err() error { return e.err }
