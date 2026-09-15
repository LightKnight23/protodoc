package cli

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/pdlfmt"
)

// trackingReader records the total octet count actually delivered to its
// caller via Read, so a test can prove a decoder never asked its source
// for more than a bounded window's worth of octets.
type trackingReader struct {
	r     io.Reader
	total int
}

func (t *trackingReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	t.total += n
	return n, err
}

// buildValidPrefix assembles a complete, decodable prefixSize-octet fixed
// prefix (Header, CommitRing, Frontmatter, SegmentTable, in that order)
// from the given segment slots, exactly as a real writer would lay them
// out (contracts/container.abnf S1-S5).
func buildValidPrefix(t *testing.T, slots []container.SegmentTableSlot) []byte {
	t.Helper()

	header := &container.Header{
		FormatMajor:        1,
		FormatMinor:        0,
		DocumentClass:      1,
		CapabilityWritten:  1,
		CapabilityRequired: 1,
		HistoryMode:        container.HistoryComplete,
		UnicodeVersionID:   1,
		PrefixLayoutID:     1,
	}

	var ring [container.CommitRingSlots]container.CommitRingRecord
	for i := range ring {
		ring[i] = container.CommitRingRecord{
			Sequence:     uint64(i + 1), // strictly increasing, no PD-RING-001 tie
			LedgerLength: uint64(container.CommitRingSize),
		}
	}

	fm := &container.Frontmatter{
		Title:           "Quarterly Report",
		PageCount:       12,
		Language:        "en-US",
		ColourProfileID: container.ColourProfileSRGBD65,
	}

	prefix := make([]byte, 0, prefixSize)
	prefix = append(prefix, header.Encode(nil)...)
	prefix = append(prefix, container.EncodeCommitRing(&ring, nil)...)

	fmEnc, err := fm.Encode(nil)
	if err != nil {
		t.Fatalf("Frontmatter.Encode: %v", err)
	}
	prefix = append(prefix, fmEnc...)

	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	prefix = append(prefix, stEnc...)

	if len(prefix) != prefixSize {
		t.Fatalf("assembled prefix length = %d, want %d", len(prefix), prefixSize)
	}
	return prefix
}

// TestTR_012_InspectVerbBoundedRead is T-0329's named test: inspect on a
// valid fixture lists every SegmentTableSlot's type/length/digest and
// coverage-hint field in the documented JSON shape, with an I/O trace
// confirming reads never extend past octet 1,048,576.
func TestTR_012_InspectVerbBoundedRead(t *testing.T) {
	slots := []container.SegmentTableSlot{
		{
			SegmentType: container.SegmentTypeContent,
			Flags:       container.SlotFlagCoverageHint,
			Offset:      uint64(prefixSize),
			Length:      4096,
			FrameCount:  1,
			Digest:      pdlfmt.Digest256{0xAA, 0xBB},
		},
		{
			SegmentType: container.SegmentTypeResource,
			Flags:       0,
			Offset:      uint64(prefixSize) + 4096,
			Length:      2048,
			FrameCount:  1,
			Digest:      pdlfmt.Digest256{0xCC, 0xDD},
		},
	}
	prefix := buildValidPrefix(t, slots)

	// Simulate a much larger on-disk file: real ledger octets past the
	// fixed prefix that inspect must never read (contracts/cli.md S3:
	// "Never reads a single octet of the ledger past the prefix").
	ledgerTail := bytes.Repeat([]byte{0xEE}, 8192)
	full := append(append([]byte{}, prefix...), ledgerTail...)

	tr := &trackingReader{r: bytes.NewReader(full)}
	result := inspectPrefix(tr, uint64(len(full)))

	if tr.total != prefixSize {
		t.Fatalf("inspectPrefix read %d octets from its source, want exactly %d (never past the fixed prefix)", tr.total, prefixSize)
	}
	if result.Status != statusOK {
		t.Fatalf("status = %q, exit_code = %d, want OK; findings = %+v", result.Status, result.ExitCode, result.Findings)
	}
	if result.ExitCode != exitOK {
		t.Fatalf("exit_code = %d, want %d", result.ExitCode, exitOK)
	}

	env := Envelope("inspect", result)
	encoded, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal(envelope): %v", err)
	}

	var decoded struct {
		Prefix struct {
			Segments []struct {
				Ordinal      int    `json:"ordinal"`
				Type         string `json:"type"`
				Offset       uint64 `json:"offset"`
				Length       uint64 `json:"length"`
				Digest       string `json:"digest"`
				CoverageHint bool   `json:"coverage_hint"`
			} `json:"segments"`
			Frontmatter struct {
				Title           string `json:"title"`
				PageCount       uint32 `json:"page_count"`
				Language        string `json:"language"`
				ColourProfileID uint16 `json:"colour_profile_id"`
			} `json:"frontmatter"`
			Header struct {
				FormatMajor uint16 `json:"format_major"`
			} `json:"header"`
		} `json:"prefix"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(decoded.Prefix.Segments) != len(slots) {
		t.Fatalf("segments count = %d, want %d", len(decoded.Prefix.Segments), len(slots))
	}
	for i, slot := range slots {
		got := decoded.Prefix.Segments[i]
		if got.Ordinal != i {
			t.Errorf("segment %d: ordinal = %d, want %d", i, got.Ordinal, i)
		}
		wantType := segmentTypeName[slot.SegmentType]
		if got.Type != wantType {
			t.Errorf("segment %d: type = %q, want %q", i, got.Type, wantType)
		}
		if got.Offset != slot.Offset {
			t.Errorf("segment %d: offset = %d, want %d", i, got.Offset, slot.Offset)
		}
		if got.Length != slot.Length {
			t.Errorf("segment %d: length = %d, want %d", i, got.Length, slot.Length)
		}
		wantDigest := hex.EncodeToString(slot.Digest[:])
		if got.Digest != wantDigest {
			t.Errorf("segment %d: digest = %q, want %q", i, got.Digest, wantDigest)
		}
		wantHint := slot.Flags&container.SlotFlagCoverageHint != 0
		if got.CoverageHint != wantHint {
			t.Errorf("segment %d: coverage_hint = %v, want %v", i, got.CoverageHint, wantHint)
		}
	}

	if decoded.Prefix.Frontmatter.Title != "Quarterly Report" {
		t.Errorf("frontmatter.title = %q, want %q", decoded.Prefix.Frontmatter.Title, "Quarterly Report")
	}
	if decoded.Prefix.Frontmatter.PageCount != 12 {
		t.Errorf("frontmatter.page_count = %d, want 12", decoded.Prefix.Frontmatter.PageCount)
	}
	if decoded.Prefix.Frontmatter.Language != "en-US" {
		t.Errorf("frontmatter.language = %q, want %q", decoded.Prefix.Frontmatter.Language, "en-US")
	}
	if decoded.Prefix.Frontmatter.ColourProfileID != uint16(container.ColourProfileSRGBD65) {
		t.Errorf("frontmatter.colour_profile_id = %d, want %d", decoded.Prefix.Frontmatter.ColourProfileID, container.ColourProfileSRGBD65)
	}
	if decoded.Prefix.Header.FormatMajor != 1 {
		t.Errorf("header.format_major = %d, want 1", decoded.Prefix.Header.FormatMajor)
	}
}

// TestTR_012_InspectVerbTruncatedPrefixIsInvalid confirms a file shorter
// than the fixed 1,048,576-octet prefix reports INVALID rather than a
// partial result.
func TestTR_012_InspectVerbTruncatedPrefixIsInvalid(t *testing.T) {
	short := bytes.NewReader(make([]byte, container.HeaderSize))
	result := inspectPrefix(short, container.HeaderSize)
	if result.Status != "INVALID" {
		t.Fatalf("status = %q, want INVALID", result.Status)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected at least one finding for a truncated prefix")
	}
}

// TestTR_012_InspectVerbUnsupportedFormatMajor confirms a document
// declaring a format-major this tool does not implement reports
// UNSUPPORTED (FR-123) rather than attempting to interpret its constructs.
func TestTR_012_InspectVerbUnsupportedFormatMajor(t *testing.T) {
	header := &container.Header{
		FormatMajor:        99,
		DocumentClass:      1,
		CapabilityWritten:  1,
		CapabilityRequired: 1,
		HistoryMode:        container.HistoryComplete,
		UnicodeVersionID:   1,
		PrefixLayoutID:     1,
	}
	buf := make([]byte, prefixSize)
	copy(buf, header.Encode(nil))

	result := inspectPrefix(bytes.NewReader(buf), prefixSize)
	if result.Status != "UNSUPPORTED" {
		t.Fatalf("status = %q, want UNSUPPORTED", result.Status)
	}
	if result.ExitCode != 2 {
		t.Fatalf("exit_code = %d, want 2", result.ExitCode)
	}
}
