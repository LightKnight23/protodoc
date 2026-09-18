// Exact-rational rasterizer core (T-0248; NFR-019, DP-010). Every component
// value of the reference 300-dpi raster is determined by integer or exact-
// rational arithmetic -- there is NO float32/float64 anywhere in this file's
// arithmetic path. Two independent conforming implementations produce bit-
// identical output because the sample-grid origin, subdivision order, coverage
// computation and rounding mode are fully pinned:
//
//   - Sample-grid origin: pixel (x,y) samples the scanline y+1/2 (half-integer
//     scanline centres), integer pixel columns.
//   - Curve flattening: de Casteljau subdivision to a tolerance of 762 base
//     units (DP-010); done in exact rationals.
//   - Coverage: exact-rational span coverage per pixel on each scanline.
//   - Rounding: round-half-to-even on the final 0..255 coverage quantization.
package render

import (
	"math/big"
	"sort"
)

// RatPoint is a 2-D point with exact-rational coordinates (base units).
type RatPoint struct {
	X, Y *big.Rat
}

// edge is an active-edge-table edge with exact-rational endpoints, oriented so
// Y0 <= Y1. dir is +1 for an upward edge and -1 for a downward edge (non-zero
// winding).
type edge struct {
	y0, y1 *big.Rat
	x0, x1 *big.Rat
	dir    int
}

// Raster is a coverage bitmap: width x height pixels, each an 0..255 coverage
// value computed by exact-rational area on the scanline. It is deterministic
// and float-free.
type Raster struct {
	Width, Height int
	Cover         []byte // row-major, len == Width*Height
}

// flattenTolerance is the DP-010 curve-flattening tolerance in base units.
var flattenTolerance = big.NewRat(762, 1)

// half is the exact rational 1/2 (the scanline sample centre offset).
var half = big.NewRat(1, 2)

// FlattenQuadratic flattens a quadratic Bezier (p0,p1,p2) to a polyline by
// exact-rational de Casteljau subdivision until the control point's rational
// deviation from the chord is within flattenTolerance. Subdivision order is
// left-then-right (pinned), so the emitted point sequence is deterministic.
func FlattenQuadratic(p0, p1, p2 RatPoint) []RatPoint {
	var out []RatPoint
	var rec func(a, b, c RatPoint, depth int)
	rec = func(a, b, c RatPoint, depth int) {
		if depth >= 24 || withinTolerance(a, b, c) {
			out = append(out, a)
			return
		}
		ab := midpoint(a, b)
		bc := midpoint(b, c)
		abc := midpoint(ab, bc)
		rec(a, ab, abc, depth+1)
		rec(abc, bc, c, depth+1)
	}
	rec(p0, p1, p2, 0)
	out = append(out, p2)
	return out
}

// withinTolerance reports whether control point b deviates from the chord (a,c)
// by no more than flattenTolerance, using an exact-rational squared-distance
// comparison (no sqrt, no floats).
func withinTolerance(a, b, c RatPoint) bool {
	// Deviation of b from the line a->c, measured by the perpendicular-distance
	// numerator |(c-a) x (b-a)| against tol * |c-a|. Compare squares to avoid
	// sqrt, all in exact rationals.
	dx := new(big.Rat).Sub(c.X, a.X)
	dy := new(big.Rat).Sub(c.Y, a.Y)
	ex := new(big.Rat).Sub(b.X, a.X)
	ey := new(big.Rat).Sub(b.Y, a.Y)
	cross := new(big.Rat).Sub(new(big.Rat).Mul(dx, ey), new(big.Rat).Mul(dy, ex))
	crossSq := new(big.Rat).Mul(cross, cross)
	lenSq := new(big.Rat).Add(new(big.Rat).Mul(dx, dx), new(big.Rat).Mul(dy, dy))
	tolSq := new(big.Rat).Mul(flattenTolerance, flattenTolerance)
	// crossSq <= tolSq * lenSq  <=>  deviation <= tol
	rhs := new(big.Rat).Mul(tolSq, lenSq)
	return crossSq.Cmp(rhs) <= 0
}

func midpoint(a, b RatPoint) RatPoint {
	return RatPoint{
		X: new(big.Rat).Mul(new(big.Rat).Add(a.X, b.X), half),
		Y: new(big.Rat).Mul(new(big.Rat).Add(a.Y, b.Y), half),
	}
}

// FillPolygon rasterizes a closed polygon (a slice of vertices, implicitly
// closed) into a width x height coverage Raster using an exact-rational
// active-edge table and non-zero winding. Coverage per pixel is the exact-
// rational covered fraction on the pixel's scanline (analytic in x), quantized
// to 0..255 with round-half-to-even. Float-free.
func FillPolygon(poly []RatPoint, width, height int) Raster {
	r := Raster{Width: width, Height: height, Cover: make([]byte, width*height)}
	if len(poly) < 3 {
		return r
	}
	edges := buildEdges(poly)

	for y := 0; y < height; y++ {
		scan := new(big.Rat).Add(big.NewRat(int64(y), 1), half) // y + 1/2
		xs := scanlineCrossings(edges, scan)
		if len(xs) == 0 {
			continue
		}
		accumulateSpans(r, y, xs, width)
	}
	return r
}

func buildEdges(poly []RatPoint) []edge {
	var edges []edge
	n := len(poly)
	for i := 0; i < n; i++ {
		a := poly[i]
		b := poly[(i+1)%n]
		if a.Y.Cmp(b.Y) == 0 {
			continue // horizontal edges contribute no crossings
		}
		e := edge{}
		if a.Y.Cmp(b.Y) < 0 {
			e.y0, e.y1, e.x0, e.x1, e.dir = a.Y, b.Y, a.X, b.X, +1
		} else {
			e.y0, e.y1, e.x0, e.x1, e.dir = b.Y, a.Y, b.X, a.X, -1
		}
		edges = append(edges, e)
	}
	return edges
}

// crossing is an exact-rational x-intercept with a winding direction.
type crossing struct {
	x   *big.Rat
	dir int
}

// scanlineCrossings returns the exact-rational x-intercepts of every edge that
// straddles the scanline y, with winding directions, sorted by x.
func scanlineCrossings(edges []edge, y *big.Rat) []crossing {
	var xs []crossing
	for _, e := range edges {
		// Half-open [y0, y1) so shared vertices are counted once.
		if y.Cmp(e.y0) < 0 || y.Cmp(e.y1) >= 0 {
			continue
		}
		// x = x0 + (x1-x0) * (y-y0)/(y1-y0), all exact rational.
		t := new(big.Rat).Quo(new(big.Rat).Sub(y, e.y0), new(big.Rat).Sub(e.y1, e.y0))
		x := new(big.Rat).Add(e.x0, new(big.Rat).Mul(new(big.Rat).Sub(e.x1, e.x0), t))
		xs = append(xs, crossing{x: x, dir: e.dir})
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i].x.Cmp(xs[j].x) < 0 })
	return xs
}

// accumulateSpans walks the sorted crossings applying non-zero winding, and for
// each covered [xStart,xEnd) span adds exact-rational per-pixel coverage to row
// y, quantized round-half-to-even to 0..255.
func accumulateSpans(r Raster, y int, xs []crossing, width int) {
	winding := 0
	for i := 0; i+1 < len(xs); i++ {
		winding += xs[i].dir
		if winding == 0 {
			continue
		}
		spanStart := xs[i].x
		spanEnd := xs[i+1].x
		addSpanCoverage(r, y, spanStart, spanEnd, width)
	}
}

// addSpanCoverage adds the exact covered fraction of [start,end) to each pixel
// column it overlaps on row y.
func addSpanCoverage(r Raster, y int, start, end *big.Rat, width int) {
	if start.Cmp(end) >= 0 {
		return
	}
	// Clamp to [0,width].
	zero := big.NewRat(0, 1)
	w := big.NewRat(int64(width), 1)
	if start.Cmp(zero) < 0 {
		start = zero
	}
	if end.Cmp(w) > 0 {
		end = w
	}
	if start.Cmp(end) >= 0 {
		return
	}
	firstCol := floorRat(start)
	lastCol := ceilRat(end) - 1
	for col := firstCol; col <= lastCol; col++ {
		if col < 0 || col >= width {
			continue
		}
		left := maxRat(start, big.NewRat(int64(col), 1))
		right := minRat(end, big.NewRat(int64(col+1), 1))
		frac := new(big.Rat).Sub(right, left) // 0..1 exact
		if frac.Sign() <= 0 {
			continue
		}
		cov := quantizeCoverage(frac)
		idx := y*width + col
		total := int(r.Cover[idx]) + cov
		if total > 255 {
			total = 255
		}
		r.Cover[idx] = byte(total)
	}
}

// quantizeCoverage maps an exact-rational fraction in [0,1] to 0..255 with
// round-half-to-even. Float-free: scale by 255, then integer round-half-even.
func quantizeCoverage(frac *big.Rat) int {
	scaled := new(big.Rat).Mul(frac, big.NewRat(255, 1))
	return roundHalfToEven(scaled)
}

// roundHalfToEven rounds an exact rational to the nearest integer, ties to even,
// using only big.Int arithmetic.
func roundHalfToEven(r *big.Rat) int {
	num := new(big.Int).Set(r.Num())
	den := new(big.Int).Set(r.Denom())
	q := new(big.Int)
	rem := new(big.Int)
	q.QuoRem(num, den, rem) // truncated toward zero
	if rem.Sign() == 0 {
		return int(q.Int64())
	}
	// Compare 2*|rem| to den.
	twiceRem := new(big.Int).Abs(rem)
	twiceRem.Lsh(twiceRem, 1)
	cmp := twiceRem.Cmp(den)
	up := func(qq *big.Int) *big.Int {
		if num.Sign() >= 0 {
			return new(big.Int).Add(qq, big.NewInt(1))
		}
		return new(big.Int).Sub(qq, big.NewInt(1))
	}
	switch {
	case cmp < 0:
		return int(q.Int64())
	case cmp > 0:
		return int(up(q).Int64())
	default: // exactly half: round to even
		if q.Bit(0) == 0 {
			return int(q.Int64())
		}
		return int(up(q).Int64())
	}
}

func floorRat(r *big.Rat) int {
	q := new(big.Int).Div(r.Num(), r.Denom()) // Div floors toward -inf
	return int(q.Int64())
}

func ceilRat(r *big.Rat) int {
	q := new(big.Int)
	m := new(big.Int)
	q.DivMod(r.Num(), r.Denom(), m)
	if m.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	return int(q.Int64())
}

func maxRat(a, b *big.Rat) *big.Rat {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

func minRat(a, b *big.Rat) *big.Rat {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}
