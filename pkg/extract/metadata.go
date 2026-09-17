// Extraction-view descriptive metadata (T-0091, FR-044): Metadata returns a
// document's title, page count, language and other Frontmatter-sourced
// descriptive fields, computed by decoding only the fixed Frontmatter region
// -- no font, shaping, layout or graphics decoder is constructed, and no
// integrity or render facility is reached. Library-level only: wiring this
// into the `extract` CLI verb's stdout payload is M18's concern (the plan.md
// review's disclosed gap: the CLI currently maps this data only to
// `inspect`); flagged forward to the M18 CLI task.
package extract

import (
	"fmt"
	"io"

	"Protodoc/pkg/container"
)

// DocumentMetadata is the descriptive, decoder-free metadata surfaced by the
// extraction view. Absent Frontmatter fields take their documented zero
// value (empty string, zero count).
type DocumentMetadata struct {
	Title     string
	PageCount uint32
	Language  string
}

// Metadata reads the document's Frontmatter region from r and returns its
// descriptive metadata. It decodes only the fixed prefix's Frontmatter
// region (container.DecodeFrontmatter); it constructs no font/shaping/
// layout/graphics decoder and calls into neither render nor integrity. An
// absent field is returned at its zero value.
func Metadata(r io.ReaderAt) (DocumentMetadata, error) {
	prefix := make([]byte, PrefixLength)
	n, err := r.ReadAt(prefix, 0)
	if err != nil && err != io.EOF {
		return DocumentMetadata{}, fmt.Errorf("extract: reading fixed prefix: %w", err)
	}
	if n < PrefixLength {
		return DocumentMetadata{}, ErrPrefixTooShort
	}
	fmBytes := prefix[container.FrontmatterOffset : container.FrontmatterOffset+container.FrontmatterRegionSize]
	fm, err := container.DecodeFrontmatter(fmBytes)
	if err != nil {
		return DocumentMetadata{}, fmt.Errorf("extract: decoding frontmatter: %w", err)
	}
	return DocumentMetadata{
		Title:     fm.Title,
		PageCount: fm.PageCount,
		Language:  fm.Language,
	}, nil
}
