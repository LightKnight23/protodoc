// Real validate backend (T-0373, DEFECT-2026-09-19 fix). Replaces the no-op
// ValidateStepsFor stub with a production step builder that OPENS the file,
// reads the fixed bounded prefix (the same CP-006 verify-before-decode read
// inspect.go performs), and constructs real validate.Step checks over the
// decoded Header / commit-ring / Frontmatter / SegmentTable. A file that is
// absent, truncated, or structurally malformed yields a failing validity step
// (FR-103: the earliest failure halts the pipeline; no partial output). Go
// stdlib only in this path (CP-010).
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"Protodoc/pkg/container"
	"Protodoc/pkg/validate"
)

// realValidateStepsFor builds the real validation step set for a document at
// path. It reads only the bounded prefix (never a segment body, CP-006) and
// returns steps whose Run closures inspect the decoded prefix. The returned
// step count is the number of checks assembled.
//
// If the file cannot be opened or the prefix cannot be read, it returns a
// single always-failing step reporting that, so validate.Run yields INVALID
// rather than a spurious OK.
func realValidateStepsFor(path string) ([]validate.Step, int) {
	prefix, fileLen, openErr := readBoundedPrefix(path)
	if openErr != nil {
		// A file we cannot even read is structurally invalid at step 1.
		return []validate.Step{{
			ID: validate.StepMagicHeader,
			Run: func() *validate.Finding {
				return &validate.Finding{
					Step:    validate.StepMagicHeader,
					RuleID:  "TR-006",
					Message: openErr.Error(),
				}
			},
		}}, 1
	}

	// Decode the prefix regions once; each step closes over the results.
	steps := []validate.Step{
		// Step 1: magic + header decode.
		{ID: validate.StepMagicHeader, Run: func() *validate.Finding {
			if _, err := container.DecodeHeader(prefix[:container.HeaderSize]); err != nil {
				return &validate.Finding{Step: validate.StepMagicHeader, RuleID: "PD-HEADER-001", Message: err.Error()}
			}
			return nil
		}},
		// Step 2: capability arithmetic (required <= written).
		{ID: validate.StepCapabilityArith, Run: func() *validate.Finding {
			h, err := container.DecodeHeader(prefix[:container.HeaderSize])
			if err != nil {
				return nil // step 1 already reports the decode failure
			}
			if h.CapabilityRequired > h.CapabilityWritten {
				return &validate.Finding{Step: validate.StepCapabilityArith, RuleID: "FR-104",
					Message: fmt.Sprintf("capability_required %d > capability_written %d", h.CapabilityRequired, h.CapabilityWritten)}
			}
			if _, ok := container.UnicodeVersionByFormatMajor[h.FormatMajor]; !ok {
				return &validate.Finding{Step: validate.StepCapabilityArith, RuleID: "FR-123",
					Message: fmt.Sprintf("format-major %d is not implemented by this tool", h.FormatMajor)}
			}
			return nil
		}},
		// Step 3: commit-ring winner selection (PD-RING-001).
		{ID: validate.StepRingWinner, Run: func() *validate.Finding {
			ring := prefix[container.HeaderSize : container.HeaderSize+container.CommitRingSize]
			if _, _, err := container.SelectWinner(ring, fileLen); err != nil {
				return &validate.Finding{Step: validate.StepRingWinner, RuleID: "PD-RING-001", Message: err.Error()}
			}
			return nil
		}},
		// Step 5/7: frontmatter + segment-table decode and segment-type validity.
		{ID: validate.StepBoundedPrefix, Run: func() *validate.Finding {
			if _, err := container.DecodeFrontmatter(prefix[container.FrontmatterOffset : container.FrontmatterOffset+container.FrontmatterRegionSize]); err != nil {
				return &validate.Finding{Step: validate.StepBoundedPrefix, RuleID: "TR-006-FRONTMATTER", Message: err.Error()}
			}
			table, err := container.DecodeSegmentTable(prefix[container.SegmentTableOffset:])
			if err != nil {
				return &validate.Finding{Step: validate.StepBoundedPrefix, RuleID: "TR-006-SEGMENTTABLE", Message: err.Error()}
			}
			for _, slot := range table {
				if f := validate.CheckSegmentType(slot.SegmentType); f != nil {
					return f
				}
			}
			return nil
		}},
	}
	return steps, len(steps)
}

// readBoundedPrefix opens path, stats it, and reads exactly the fixed prefix.
// It returns the prefix bytes and the total file length. An absent file, a
// stat error, or a short read is an error (never a silent empty prefix).
func readBoundedPrefix(path string) (prefix []byte, fileLen uint64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("cannot stat %q: %w", path, err)
	}
	fileLen = uint64(info.Size())

	buf := make([]byte, prefixSize)
	n, err := io.ReadFull(f, buf)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return nil, 0, fmt.Errorf("bounded prefix truncated: read %d of %d required octets", n, prefixSize)
		}
		return nil, 0, fmt.Errorf("reading bounded prefix of %q: %w", path, err)
	}
	return buf, fileLen, nil
}

func init() {
	// Replace the test-double stub with the real production backend. Tests that
	// need a fixture reassign ValidateStepsFor and restore it afterward.
	ValidateStepsFor = realValidateStepsFor
}
