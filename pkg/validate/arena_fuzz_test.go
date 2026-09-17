package validate

import (
	"bytes"
	"runtime"
	"testing"
)

// FuzzValidate_NFR030_MemoryBoundAndNoCrash is T-0122's native fuzz target
// (feeding CP-012's continuous fuzz). It drives arbitrary byte inputs through
// the validator entrypoint ValidateBytes and asserts two properties for every
// input, valid or malformed: (1) the process never panics, and (2) the memory
// the call allocates never exceeds T-0121's adopted NFR-030 bound
// max(ArenaFloor, 4*inputLen). A panic or bound violation fails the fuzz.
func FuzzValidate_NFR030_MemoryBoundAndNoCrash(f *testing.F) {
	// Seed corpus (>=2 seeds so fuzz maturity discovery treats this as a
	// real target): empty, a short/malformed prefix, and a full-prefix-sized
	// zero image so the segment-table walk path is exercised.
	f.Add([]byte{})
	f.Add([]byte("PDL\x00not-a-valid-prefix"))
	f.Add(make([]byte, 1<<20+4096))

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		bound := AdoptedMemoryBound(len(data))

		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		// Must not panic on any input.
		_ = ValidateBytes(r, int64(len(data)))
		runtime.ReadMemStats(&after)

		allocated := after.TotalAlloc - before.TotalAlloc
		if allocated > bound {
			t.Fatalf("ValidateBytes allocated %d octets for %d-octet input, exceeds adopted NFR-030 bound %d", allocated, len(data), bound)
		}
	})
}
