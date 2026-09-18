// Alt-text quality validation (FR-040; T-0285). Validator rule PD-A11Y-003: a
// non-decorative embedded object MUST carry a meaningful text alternative. It
// is rejected when the alt text is absent/empty, or when it merely echoes the
// object's filename, content digest, or pixel dimensions — none of which
// convey the object's meaning to a reader who cannot see it.
//
// Provisional per T-0267 (clarify-002.md, OPEN awaiting Eyvar).
package semantics

import (
	"encoding/hex"
	"strconv"
	"strings"
)

// RuleAltTextQuality is the validator rule id for alt-text quality.
const RuleAltTextQuality = "PD-A11Y-003"

// AltTextQualityKind classifies why an alt text was rejected.
type AltTextQualityKind uint8

const (
	// AltEmpty: a non-decorative object with absent/empty alt text.
	AltEmpty AltTextQualityKind = iota
	// AltEchoesMetadata: alt text that merely restates filename/digest/dims.
	AltEchoesMetadata
)

// ObjectMetadata is the object's non-semantic surface an alt text must not
// merely echo.
type ObjectMetadata struct {
	Filename string
	Digest   []byte // content digest bytes
	Width    int
	Height   int
}

// AltTextQualityFinding is a PD-A11Y-003 rejection.
type AltTextQualityFinding struct {
	Rule string
	Kind AltTextQualityKind
}

// normalizeAlt lowercases and trims for echo comparison.
func normalizeAlt(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// echoesMetadata reports whether alt merely restates the object's filename,
// digest (hex, any case) or dimensions, carrying no descriptive content.
func echoesMetadata(alt string, md ObjectMetadata) bool {
	n := normalizeAlt(alt)
	if n == "" {
		return false
	}
	// Filename (with or without extension).
	if fn := normalizeAlt(md.Filename); fn != "" {
		if n == fn {
			return true
		}
		if dot := strings.LastIndex(fn, "."); dot > 0 && n == fn[:dot] {
			return true
		}
	}
	// Digest hex.
	if len(md.Digest) > 0 {
		if n == hex.EncodeToString(md.Digest) {
			return true
		}
	}
	// Dimensions like "800x600" / "800 x 600" / "800×600".
	if md.Width > 0 && md.Height > 0 {
		w, h := strconv.Itoa(md.Width), strconv.Itoa(md.Height)
		collapsed := strings.NewReplacer(" ", "", "×", "x", "*", "x").Replace(n)
		if collapsed == w+"x"+h {
			return true
		}
	}
	return false
}

// CheckAltTextQuality applies PD-A11Y-003 to a non-decorative object's alt
// text given its metadata. It returns a finding if the alt text is empty, or
// merely echoes the object's filename/digest/dimensions. A decorative object
// is not checked here (its empty alt text is required, checked at T-0284). An
// empty result means the alt text is acceptable.
func CheckAltTextQuality(a EmbeddedObjectAlt, md ObjectMetadata) []AltTextQualityFinding {
	if a.Decorative {
		return nil
	}
	if strings.TrimSpace(a.AltText) == "" {
		return []AltTextQualityFinding{{Rule: RuleAltTextQuality, Kind: AltEmpty}}
	}
	if echoesMetadata(a.AltText, md) {
		return []AltTextQualityFinding{{Rule: RuleAltTextQuality, Kind: AltEchoesMetadata}}
	}
	return nil
}
