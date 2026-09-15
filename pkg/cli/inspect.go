// inspect verb (TR-012, contracts/cli.md S3): reads only the fixed
// 1,048,576-octet prefix (container.abnf S1) and reports what TR-006,
// TR-007 and TR-008 require be determinable from it with zero segment-body
// decode: per-unit type/length/digest, a non-authoritative coverage HINT,
// the header fields and the winning commit-ring slot's fields.
package cli

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"Protodoc/pkg/container"
)

// prefixSize is the fixed bounded-prefix width inspect never reads past
// (container.abnf S1): Header + CommitRing + Frontmatter + SegmentTable.
const prefixSize = container.SegmentTableOffset + container.SegmentTableRegionSize

// segmentTypeName maps a closed SegmentTableSlot.SegmentType octet to its
// contracts/container.abnf S5 name for stdout readability. This is a
// display mapping only, never decoded back into a SegmentType: the wire
// value itself stays the sole source of truth (CP-008).
var segmentTypeName = map[byte]string{
	container.SegmentTypeContent:  "CONTENT",
	container.SegmentTypeResource: "RESOURCE",
	container.SegmentTypeHistory:  "HISTORY",
	container.SegmentTypeAttest:   "ATTEST",
}

// inspectPrefix implements cli.md S3's bounded-prefix read against r: it
// reads EXACTLY prefixSize octets (io.ReadFull never asks r for more than
// len(buf)) and never constructs a segment/font/image/audio/video decoder
// (CP-006) — only the fixed-offset Header/CommitRing/Frontmatter/
// SegmentTable structs container already defines. fileLength is the
// document's total octet length from a stat, never a read of ledger
// content past the prefix; it is needed only for PD-RING-001's
// ledger-length eligibility check.
func inspectPrefix(r io.Reader, fileLength uint64) Result {
	buf := make([]byte, prefixSize)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{
				RuleID:  "TR-006",
				Message: fmt.Sprintf("bounded prefix truncated: read %d of %d required octets: %v", n, prefixSize, err),
			}},
		})
	}

	header, err := container.DecodeHeader(buf[:container.HeaderSize])
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-006-HEADER", Message: err.Error()}},
		})
	}

	if _, ok := container.UnicodeVersionByFormatMajor[header.FormatMajor]; !ok {
		return StatusUnsupported.ToResult(Result{
			Findings: []Finding{{
				RuleID:  "FR-123",
				Message: fmt.Sprintf("format-major %d is not implemented by this tool", header.FormatMajor),
			}},
		})
	}

	ringRegion := buf[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
	winner, winnerIndex, err := container.SelectWinner(ringRegion, fileLength)
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "PD-RING-001", Message: err.Error()}},
		})
	}

	fm, err := container.DecodeFrontmatter(buf[container.FrontmatterOffset : container.FrontmatterOffset+container.FrontmatterRegionSize])
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-006-FRONTMATTER", Message: err.Error()}},
		})
	}

	table, err := container.DecodeSegmentTable(buf[container.SegmentTableOffset:])
	if err != nil {
		return StatusInvalid.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-006-SEGMENTTABLE", Message: err.Error()}},
		})
	}

	segments := make([]map[string]any, 0, len(table))
	for i, slot := range table {
		if slot.SegmentType == container.SegmentTypeUnused {
			continue
		}
		typeName, ok := segmentTypeName[slot.SegmentType]
		if !ok {
			typeName = fmt.Sprintf("UNKNOWN(%d)", slot.SegmentType)
		}
		segments = append(segments, map[string]any{
			"ordinal":       i,
			"type":          typeName,
			"offset":        slot.Offset,
			"length":        slot.Length,
			"digest":        hex.EncodeToString(slot.Digest[:]),
			"coverage_hint": slot.Flags&container.SlotFlagCoverageHint != 0,
		})
	}

	payload := map[string]any{
		"header": map[string]any{
			"format_major":        header.FormatMajor,
			"format_minor":        header.FormatMinor,
			"document_class":      header.DocumentClass,
			"capability_written":  header.CapabilityWritten,
			"capability_required": header.CapabilityRequired,
			"durable_claim":       header.DurableClaim,
			"history_mode":        int(header.HistoryMode),
			"unicode_version_id":  header.UnicodeVersionID,
			"shaping_profile_id":  header.ShapingProfileID,
			"prefix_layout_id":    header.PrefixLayoutID,
		},
		"ring_winner": map[string]any{
			"slot_index":            winnerIndex,
			"sequence":              winner.Sequence,
			"ledger_length":         winner.LedgerLength,
			"ledger_root":           hex.EncodeToString(winner.LedgerRoot[:]),
			"segment_count":         winner.SegmentCount,
			"state_id":              hex.EncodeToString(winner.StateID[:]),
			"frontmatter_digest":    hex.EncodeToString(winner.FrontmatterDigest[:]),
			"segment_table_digest":  hex.EncodeToString(winner.SegmentTableDigest[:]),
			"t_c_root":              hex.EncodeToString(winner.TCRoot[:]),
			"parent_state_id":       hex.EncodeToString(winner.ParentStateID[:]),
			"retention_point":       winner.RetentionPoint,
			"compaction_generation": winner.CompactionGeneration,
			"structure_digest":      hex.EncodeToString(winner.StructureDigest[:]),
		},
		"frontmatter": map[string]any{
			"title":             fm.Title,
			"page_count":        fm.PageCount,
			"language":          fm.Language,
			"colour_profile_id": uint16(fm.ColourProfileID),
		},
		"segments": segments,
	}

	return StatusOK.ToResult(Result{
		Extra: map[string]any{"prefix": payload},
	})
}

// runInspect is inspect's RunFunc (contracts/cli.md S3). It opens the one
// positional input file, stats it for fileLength (no ledger octet read),
// and delegates the bounded-prefix read itself to inspectPrefix, so that
// function's own I/O boundedness is independently testable against a plain
// io.Reader with no filesystem involved.
func runInspect(args []string, stderr io.Writer) Result {
	if len(args) != 1 {
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: "inspect requires exactly one input file"}},
		})
	}
	path := args[0]

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(stderr, "protodoc inspect: %v\n", err)
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: fmt.Sprintf("cannot open %q: %v", path, err)}},
		})
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		fmt.Fprintf(stderr, "protodoc inspect: %v\n", err)
		return StatusUsage.ToResult(Result{
			Findings: []Finding{{RuleID: "TR-012", Message: fmt.Sprintf("cannot stat %q: %v", path, err)}},
		})
	}

	return inspectPrefix(f, uint64(info.Size()))
}
