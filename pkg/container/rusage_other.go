//go:build !unix

package container

// peakRSSBytes reports whether peak resident-set-size measurement is
// supported on this platform. Go's standard library exposes getrusage(2)
// only through the unix-family syscall package; a non-unix build (e.g.
// windows) has no stdlib equivalent (CP-010: stdlib only), so this stub
// reports "unsupported" rather than fabricating a value.
func peakRSSBytes() (uint64, bool) {
	return 0, false
}
