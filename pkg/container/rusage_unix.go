//go:build unix

package container

import (
	"runtime"
	"syscall"
)

// peakRSSBytes reports the calling process's peak resident set size in
// octets (getrusage(2) ru_maxrss), and whether this platform supports the
// measurement. ru_maxrss is a monotonic high-water mark for the process's
// entire lifetime: it never decreases, so a caller measuring a specific
// operation's contribution takes two readings (before and after) and uses
// the delta, not either reading alone.
//
// Darwin reports ru_maxrss in octets; Linux reports it in kibioctets
// (1024-octet units) — both per getrusage(2) on their respective
// platforms — so the kernel-reported unit is normalized to octets here.
func peakRSSBytes() (uint64, bool) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0, false
	}
	maxrss := uint64(ru.Maxrss)
	if runtime.GOOS == "linux" {
		maxrss *= 1024
	}
	return maxrss, true
}
