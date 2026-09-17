// NFC validation (T-0073, CON-002) using only the checked-in Unicode tables
// (nfctables.go) and the Go standard library -- no third-party dependency in
// this validation core path (CP-010). Every addressable text segment (a
// Run's text) and every identifier string must be in Unicode Normalization
// Form C INDEPENDENTLY, with no normalisation applied across segment
// boundaries: IsNFC validates one string in isolation and IsNFCScoped
// applies it to a single run.
//
// The check is decide-by-normalise: normaliseNFC(s) computes the NFC of s
// (canonical decomposition, canonical reordering by combining class, then
// canonical composition including algorithmic Hangul), and IsNFC reports
// whether s already equals its own NFC. This is the definition of "s is in
// NFC" and its verdict matches a reference normaliser exactly (cross-checked
// against golang.org/x/text/unicode/norm over millions of random Unicode
// strings when the tables were generated). Normalisation here is used ONLY
// to decide; the writer never stores a converted value (that write-side
// rejection is T-0074's own task).
package content

import (
	"strings"
	"unicode/utf8"
)

// Hangul composition constants (UAX #15 "Hangul Syllable Composition").
const (
	hangulLBase  = 0x1100
	hangulVBase  = 0x1161
	hangulTBase  = 0x11A7
	hangulLCount = 19
	hangulVCount = 21
	hangulTCount = 28
	hangulSBase  = 0xAC00
	hangulNCount = hangulVCount * hangulTCount // 588
	hangulSCount = hangulLCount * hangulNCount // 11172
)

// ccc returns r's canonical combining class (0 if r is a starter).
func ccc(r rune) uint8 { return canonicalCombiningClass[r] }

// decomposeInto appends the full canonical decomposition of r to dst,
// expanding Hangul syllables algorithmically and precomposed characters via
// the canonicalDecomp table (already fully expanded at generation time).
func decomposeInto(dst []rune, r rune) []rune {
	// Hangul syllable -> jamo (algorithmic).
	if r >= hangulSBase && r < hangulSBase+hangulSCount {
		si := r - hangulSBase
		l := hangulLBase + si/hangulNCount
		v := hangulVBase + (si%hangulNCount)/hangulTCount
		t := hangulTBase + si%hangulTCount
		dst = append(dst, l, v)
		if t != hangulTBase {
			dst = append(dst, t)
		}
		return dst
	}
	if d, ok := canonicalDecomp[r]; ok {
		return append(dst, d...)
	}
	return append(dst, r)
}

// canonicalOrder stably sorts runs of combining marks by ccc (a stable
// insertion sort respecting the canonical-ordering algorithm: only swap
// adjacent marks when the earlier has a strictly greater ccc, and never
// reorder across a starter).
func canonicalOrder(rs []rune) {
	for i := 1; i < len(rs); i++ {
		c := ccc(rs[i])
		if c == 0 {
			continue // starter: fixed position
		}
		j := i
		for j > 0 {
			pc := ccc(rs[j-1])
			if pc == 0 || pc <= c {
				break
			}
			rs[j-1], rs[j] = rs[j], rs[j-1]
			j--
		}
	}
}

// compose applies canonical composition (UAX #15 D117 / the Unicode
// composition algorithm) to a canonically-ordered decomposition, including
// algorithmic Hangul. It walks the sequence tracking the index of the last
// starter (lastStarter) into out; for each following character C it attempts
// to compose C with out[lastStarter] unless C is BLOCKED, i.e. some
// character between the starter and C has ccc 0 or ccc >= ccc(C). A composed
// character replaces the starter in place; a non-composed starter becomes
// the new lastStarter; a non-composed combining mark is appended.
func compose(rs []rune) []rune {
	if len(rs) == 0 {
		return rs
	}
	out := make([]rune, len(rs))
	copy(out, rs)

	lastStarter := -1 // index into out of the last starter, or -1
	if ccc(out[0]) == 0 {
		lastStarter = 0
	}
	// prevCC is the ccc of the previous character in the output run since the
	// last starter (used for the blocking test); reset at each new starter.
	prevCC := -1 // -1 means "immediately after the starter, nothing between"

	writeIdx := 1
	for i := 1; i < len(rs); i++ {
		c := rs[i]
		cc := int(ccc(c))

		composed := false
		if lastStarter >= 0 {
			blocked := prevCC != -1 && (prevCC == 0 || prevCC >= cc)
			if !blocked {
				if comp, ok := composeTwo(out[lastStarter], c); ok {
					out[lastStarter] = comp
					composed = true
				}
			}
		}
		if composed {
			// c absorbed into the starter; prevCC unchanged.
			continue
		}

		out[writeIdx] = c
		if cc == 0 {
			lastStarter = writeIdx
			prevCC = -1
		} else {
			prevCC = cc
		}
		writeIdx++
	}
	return out[:writeIdx]
}

// composeTwo returns the canonical composite of starter+combiner, if any,
// handling algorithmic Hangul (L+V, LV+T) and the tabulated pairs.
func composeTwo(starter, c rune) (rune, bool) {
	if isHangulL(starter) && isHangulV(c) {
		li := starter - hangulLBase
		vi := c - hangulVBase
		return hangulSBase + (li*hangulVCount+vi)*hangulTCount, true
	}
	if isHangulLV(starter) && c > hangulTBase && c < hangulTBase+hangulTCount {
		return starter + (c - hangulTBase), true
	}
	if comp, ok := nfcCompositionPairs[uint64(starter)<<21|uint64(c)]; ok {
		return comp, true
	}
	return 0, false
}

func isHangulL(r rune) bool { return r >= hangulLBase && r < hangulLBase+hangulLCount }
func isHangulV(r rune) bool { return r >= hangulVBase && r < hangulVBase+hangulVCount }
func isHangulLV(r rune) bool {
	if r < hangulSBase || r >= hangulSBase+hangulSCount {
		return false
	}
	return (r-hangulSBase)%hangulTCount == 0
}

// normaliseNFC returns the NFC form of s. It is used only to decide whether s
// is already NFC (IsNFC); the writer never stores its output (the write-side
// no-conversion rule is enforced separately by T-0074).
func normaliseNFC(s string) string {
	dec := make([]rune, 0, len(s))
	for _, r := range s {
		dec = decomposeInto(dec, r)
	}
	canonicalOrder(dec)
	comp := compose(dec)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range comp {
		b.WriteRune(r)
	}
	return b.String()
}

// IsNFC reports whether s is already in Unicode Normalization Form C,
// validating s in isolation (CON-002). Invalid UTF-8 is never NFC.
func IsNFC(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	return normaliseNFC(s) == s
}

// IsNFCScoped reports whether run r's text is in NFC, validated as a single
// segment in isolation (CON-002): no normalisation is considered across a
// run boundary.
func IsNFCScoped(r Run) bool { return IsNFC(r.Text) }
