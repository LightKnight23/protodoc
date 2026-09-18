// Reflow viewport fit (T-0251; FR-098). A re-laid-out presentation succeeds at
// any viewport width down to 320 reference pixels with no two-dimensional
// scrolling, except within regions declared to require 2-D presentation. The
// 2-D-region declaration construct does not yet exist in the data model (an M15
// gap; its requirement is owned by a later task); this treats "no declared
// regions present" as the default case and does not block on that gap.
package render

// MinViewportPixels is the FR-098 floor: reflow must succeed down to this width.
const MinViewportPixels = 320

// pixelBaseUnits is the number of base units per reference pixel used to map a
// viewport width in reference pixels to the reflow width in base units. It is a
// fixed integer scale (no floats).
const pixelBaseUnits = 1000

// ReflowFit is the result of fitting a paragraph at a viewport width: whether
// reflow succeeded and whether any line overflows the viewport (which would
// force horizontal scrolling). TwoDRegions counts declared 2-D regions that are
// legitimately exempt from the no-scroll rule (0 in the default case).
type ReflowFit struct {
	Succeeded          bool
	MaxLineWidth       int // widest set line, in base units
	RequiresHScroll    bool
	TwoDRegionsPresent int
}

// FitAtViewport reflows the break table at the given viewport width (in
// reference pixels) and reports whether it fits without horizontal scrolling.
// With no declared 2-D regions (the default), a successful reflow whose widest
// line does not exceed the viewport requires no scrolling. A width below
// MinViewportPixels is out of contract and reported as not succeeded.
func FitAtViewport(tbl BreakTable, viewportPixels int) ReflowFit {
	if viewportPixels < MinViewportPixels {
		return ReflowFit{Succeeded: false}
	}
	width := viewportPixels * pixelBaseUnits
	breaks := Reflow(tbl, width)
	if len(breaks) == 0 {
		return ReflowFit{Succeeded: false}
	}

	maxLine := widestSetLine(tbl.Items, breaks, width)
	return ReflowFit{
		Succeeded:       true,
		MaxLineWidth:    maxLine,
		RequiresHScroll: maxLine > width,
	}
}

// widestSetLine returns the widest SET line width across the chosen
// breakpoints (in base units). A line's set width is its natural width brought
// toward the target by its available shrink (a line never sets wider than its
// natural width minus its total shrink), so the measure reflects what the
// reader actually shows -- overflow means the line cannot shrink to fit.
func widestSetLine(items []Item, breaks []int, target int) int {
	sumW := make([]int, len(items)+1)
	sumZ := make([]int, len(items)+1)
	for i := 0; i < len(items); i++ {
		sumW[i+1] = sumW[i] + items[i].Width
		if items[i].Kind == KindGlue {
			sumZ[i+1] = sumZ[i] + items[i].Shrink
		} else {
			sumZ[i+1] = sumZ[i]
		}
	}
	maxLine := 0
	prev := 0
	for _, b := range breaks {
		if b > len(items) {
			b = len(items)
		}
		natural := sumW[b] - sumW[prev]
		shrink := sumZ[b] - sumZ[prev]
		// Minimum achievable width for this line (fully shrunk).
		minW := natural - shrink
		set := natural
		if set > target {
			set = natural - shrink // shrink toward the target
			if set < target {
				set = target
			}
		}
		if minW > set {
			set = minW
		}
		if set > maxLine {
			maxLine = set
		}
		prev = b
	}
	return maxLine
}
