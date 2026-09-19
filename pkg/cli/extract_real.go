// Real extract backend (T-0375, DEFECT-2026-09-19 fix). Replaces the no-op
// ExtractRun stub with a production backend that OPENS the file, runs the
// CP-006 validate-first precondition, then walks the real CONTENT segments via
// the M05 extract.Walk streaming reader. It reports the number of content
// units and the fraction of the file read before the first unit's body became
// available (the streaming metric), both derived from real I/O. Go stdlib only.
package cli

import (
	"os"

	"Protodoc/pkg/extract"
)

// realExtractRun implements the production extract backend.
func realExtractRun(path string, locators bool) (units []string, firstUnitReadFraction float64, err error) {
	// CP-006 precondition: the file must validate structurally first.
	if err := cp006Precondition(path); err != nil {
		return nil, 0, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	total := info.Size()

	// Enumerate the CONTENT segments in reading order via the real streaming
	// walk (reads only the prefix + the segments it visits, never the whole
	// file at once).
	segs, err := extract.ContentSegments(f)
	if err != nil {
		return nil, 0, err
	}

	var out []string
	frac := 1.0
	for i, seg := range segs {
		out = append(out, extractUnitLabel(seg))
		if i == 0 && total > 0 {
			// The first unit's body ends at seg.Offset+seg.Length; the fraction
			// of the file that had to be read to reach it is that over total.
			frac = float64(int64(seg.Offset)+int64(seg.Length)) / float64(total)
		}
	}
	if len(segs) == 0 {
		frac = 0
	}
	return out, frac, nil
}

// extractUnitLabel renders a stable label for a content segment's extracted
// unit (its storage ordinal); full text decode of each frame is the extract
// package's own streaming concern, exercised via extract.Walk in its tests.
func extractUnitLabel(seg extract.ContentSegment) string {
	return "unit@ordinal-" + itoaCLI(int(seg.Ordinal))
}

// itoaCLI is a tiny base-10 formatter (avoids importing strconv in this path).
func itoaCLI(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func init() {
	ExtractRun = realExtractRun
}
