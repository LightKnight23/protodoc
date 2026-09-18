// Format-major version gate (FR-123; T-0301). At every entry point that reads a
// document Header, the format-major is checked against the implementation's
// supported set BEFORE any other processing. On an unsupported value the reader
// declines immediately, naming the version, with zero downstream disposition
// (no decode, no migration attempt, no partial output).
package migrate

import "fmt"

// SupportedMajors is the implementation's supported format-major set for v1.
var SupportedMajors = map[uint16]bool{1: true}

// UnsupportedMajorError names the unsupported version and carries nothing else.
type UnsupportedMajorError struct {
	Major uint16
}

func (e *UnsupportedMajorError) Error() string {
	return fmt.Sprintf("migrate: unsupported format-major %d declined before any processing", e.Major)
}

// VersionGate checks a document's format-major against SupportedMajors. It is
// the FIRST thing any Header-reading entry point calls: on an unsupported
// value it returns an *UnsupportedMajorError immediately (the caller MUST NOT
// proceed to any decode/migration); on a supported value it returns nil.
func VersionGate(formatMajor uint16) error {
	if !SupportedMajors[formatMajor] {
		return &UnsupportedMajorError{Major: formatMajor}
	}
	return nil
}
