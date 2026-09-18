// Knuth-Plass reflow engine (T-0250; FR-100, NFR-022, DP-010). The reflowed
// presentation at a given viewport width is a DETERMINISTIC function of
// document content: an integer-demerits Knuth-Plass dynamic program that reads
// its line-break and hyphenation candidates EXCLUSIVELY from the in-document,
// digest-identified break/hyphenation table -- never a host dictionary or
// locale service. Two conforming implementations therefore produce identical
// line breaks. Integer arithmetic only (no floats).
package render

// BoxKind is the type of a Knuth-Plass item.
type BoxKind uint8

const (
	// KindBox: a set-width piece of content (a word or word-fragment).
	KindBox BoxKind = iota
	// KindGlue: stretchable/shrinkable inter-word space.
	KindGlue
	// KindPenalty: a legal breakpoint with an integer penalty; a hyphenation
	// candidate carries the width of the hyphen it would insert (flagged).
	KindPenalty
)

// Item is one Knuth-Plass item, all measures in integer base units.
type Item struct {
	Kind    BoxKind
	Width   int  // box width, glue natural width, or hyphen width for a penalty
	Stretch int  // glue stretchability
	Shrink  int  // glue shrinkability
	Penalty int  // penalty value at a breakpoint (KindPenalty)
	Flagged bool // a flagged penalty is a hyphenation break (from the table)
}

// BreakTable is the in-document, digest-identified break/hyphenation data: the
// item stream (with its legal breakpoints and hyphenation penalties already
// baked in from the document, not a host dictionary) and the digest that binds
// it. Reflow reads ONLY this -- no host locale service (NFR-022).
type BreakTable struct {
	Items       []Item
	InputDigest [32]byte
}

// infinity is the integer "no break allowed / infinite demerits" sentinel.
const infinity = 1 << 30

// Reflow runs the integer-demerits Knuth-Plass DP at the given viewport width
// (in base units) over the break table's items, returning the breakpoint item
// indices (the positions after which a line ends), in order. It is a pure
// function of (items, width): identical inputs give identical output across
// runs and processes (FR-100). Host state is never consulted.
func Reflow(tbl BreakTable, width int) []int {
	items := tbl.Items
	n := len(items)

	// best[i] = minimum total demerits to break optimally ending a line at a
	// breakpoint before item i; prev[i] = the previous breakpoint chosen.
	best := make([]int, n+1)
	prev := make([]int, n+1)
	for i := range best {
		best[i] = infinity
		prev[i] = -1
	}
	best[0] = 0

	// Prefix sums of natural width, stretch, shrink for O(1) line measures.
	sumW := make([]int, n+1)
	sumY := make([]int, n+1)
	sumZ := make([]int, n+1)
	for i := 0; i < n; i++ {
		sumW[i+1] = sumW[i] + items[i].Width
		if items[i].Kind == KindGlue {
			sumY[i+1] = sumY[i] + items[i].Stretch
			sumZ[i+1] = sumZ[i] + items[i].Shrink
		} else {
			sumY[i+1] = sumY[i]
			sumZ[i+1] = sumZ[i]
		}
	}

	// legalBreak reports whether a line may end at breakpoint index j (the item
	// j is a glue or a non-infinite penalty, or the end of the stream).
	legalBreak := func(j int) bool {
		if j == n {
			return true
		}
		switch items[j].Kind {
		case KindGlue:
			return true
		case KindPenalty:
			return items[j].Penalty < infinity
		default:
			return false
		}
	}

	for j := 1; j <= n; j++ {
		if j != n && !legalBreak(j) {
			continue
		}
		for i := 0; i < j; i++ {
			if best[i] == infinity {
				continue
			}
			// Natural width of the line from breakpoint i to breakpoint j.
			lineW := sumW[j] - sumW[i]
			d := lineDemerits(items, i, j, lineW, width, sumY[j]-sumY[i], sumZ[j]-sumZ[i])
			if d >= infinity {
				continue
			}
			total := best[i] + d
			if total < best[j] {
				best[j] = total
				prev[j] = i
			}
		}
	}

	// Reconstruct the breakpoints from n back to 0.
	if best[n] >= infinity {
		return nil // no feasible layout
	}
	var rev []int
	for j := n; j > 0; j = prev[j] {
		rev = append(rev, j)
		if prev[j] < 0 {
			break
		}
	}
	// Reverse into forward order.
	out := make([]int, len(rev))
	for i := range rev {
		out[i] = rev[len(rev)-1-i]
	}
	return out
}

// lineDemerits returns the integer demerits of setting the items on the line
// [i,j) at the target width, or >= infinity if the line cannot fit even with
// full shrink. Demerits are (badness + penalty-contribution)^2-ish but kept
// linear-integer for exact determinism: |width - lineW| adjusted by available
// stretch/shrink, plus a flagged-penalty surcharge.
func lineDemerits(items []Item, i, j, lineW, width, stretch, shrink int) int {
	diff := width - lineW
	var badness int
	switch {
	case diff >= 0:
		// Need to stretch by diff; if no stretch available and diff>0, penalize.
		if stretch == 0 {
			badness = boundedSquare(diff)
		} else {
			r := (diff * 100) / stretch // integer ratio *100
			badness = boundedSquare(r)
		}
	default:
		over := -diff
		if over > shrink {
			return infinity // cannot shrink enough -> line does not fit
		}
		if shrink == 0 {
			badness = boundedSquare(over)
		} else {
			r := (over * 100) / shrink
			badness = boundedSquare(r)
		}
	}
	// Penalty contribution at breakpoint j (if it is a penalty item).
	pen := 0
	if j < len(items) && items[j].Kind == KindPenalty {
		p := items[j].Penalty
		if items[j].Flagged {
			pen += 3000 // integer surcharge discouraging hyphen breaks
		}
		pen += p
	}
	d := badness + pen + 1 // +1 line penalty to prefer fewer lines on ties
	if d < 0 || d >= infinity {
		// A feasible-but-loose line: cap just below infinity so it is still
		// choosable when nothing tighter is available, but always dominated by
		// a genuinely fitting line.
		return infinity - 1
	}
	return d
}

// boundedSquare returns x*x capped at maxBadness so a large-but-feasible line
// is never mistaken for the infeasible sentinel.
func boundedSquare(x int) int {
	if x >= maxBadnessRoot || x <= -maxBadnessRoot {
		return maxBadness
	}
	return x * x
}

// maxBadness is the largest badness a single line may contribute; maxBadnessRoot
// is its integer square root. Both stay well below infinity so sums of a
// bounded number of lines cannot overflow into or past the sentinel.
const (
	maxBadness     = 1 << 20
	maxBadnessRoot = 1 << 10
)
