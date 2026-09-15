package pdlfmt

import (
	"fmt"
	"math/big"
)

// ProportionalShares resolves len(weights) proportional/automatic shares of
// total (CON-013), by the documented rounding rule and accumulation order:
//
//	share_i = floor(total*prefix_i/W) - floor(total*prefix_{i-1}/W)
//
// where prefix_i is the sum of weights[0..i] inclusive (prefix_0 = 0) and W
// is the sum of all weights (plan.md S "CON-012/013" row). This is a
// telescoping sum: summing share_i over every i always yields exactly
// floor(total*W/W) - floor(total*0/W) = total, so the shares sum to total
// exactly for any total and any positive weights, with no separate
// remainder-correction step needed. The accumulated rounding remainder
// still has to land somewhere; per this exact accumulation order it lands
// wherever the running cumulative-floor crosses an integer boundary,
// concentrating entirely on the final share when there is only a single
// unit of remainder to place (an equal-weight split whose total is not a
// multiple of len(weights), CON-013's own example).
//
// total must be non-negative and every weight must be strictly positive
// (a zero weight would make that column's prefix indistinguishable from
// its predecessor's, always resolving to a zero-width share, which is a
// caller error, not a rounding case this function silently accepts).
func ProportionalShares(total GeometricValue, weights []uint64) ([]GeometricValue, error) {
	if total < 0 {
		return nil, fmt.Errorf("pdlfmt: proportional total %d is negative", int64(total))
	}
	if len(weights) == 0 {
		return nil, fmt.Errorf("pdlfmt: proportional shares needs at least one weight")
	}

	var w uint64
	for i, wt := range weights {
		if wt == 0 {
			return nil, fmt.Errorf("pdlfmt: proportional weight %d is zero", i)
		}
		w += wt
	}

	totalBig := big.NewInt(int64(total))
	wBig := new(big.Int).SetUint64(w)

	shares := make([]GeometricValue, len(weights))
	var prefix uint64
	prevFloor := big.NewInt(0)
	for i, wt := range weights {
		prefix += wt
		// total*prefix can exceed uint64/int64 range for large inputs;
		// math/big keeps the intermediate product exact so the result
		// matches the documented formula bit-for-bit rather than only
		// approximating it for small totals.
		curFloor := new(big.Int).Mul(totalBig, new(big.Int).SetUint64(prefix))
		curFloor.Quo(curFloor, wBig) // Quo truncates toward zero; both operands are non-negative, so this is floor.

		share := new(big.Int).Sub(curFloor, prevFloor)
		shares[i] = GeometricValue(share.Int64())
		prevFloor = curFloor
	}
	return shares, nil
}
