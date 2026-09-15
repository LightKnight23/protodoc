// Command protodoc-traceaudit is T-0353's runnable form of the NFR-029
// traceability checker (pkg/traceability, T-0352): it parses the frozen
// spec.md and contracts/*.abnf for every normative-statement id, scans
// this module's own source for conformance-case coverage, and prints the
// resulting report to stdout. Used to regenerate
// docs/nfr-029-traceability-gap-report.md; see that file's own generation
// note.
//
// Usage (run from the module root):
//
//	go run ./cmd/protodoc-traceaudit
package main

import (
	"fmt"
	"os"

	"Protodoc/pkg/traceability"
)

func main() {
	report, err := traceability.Audit(
		"specs/001-protodoc-format-core/spec.md",
		[]string{
			"specs/001-protodoc-format-core/contracts/container.abnf",
			"specs/001-protodoc-format-core/contracts/document.abnf",
			"specs/001-protodoc-format-core/contracts/integrity.abnf",
		},
		".",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "protodoc-traceaudit: %v\n", err)
		os.Exit(2)
	}

	fmt.Print(report.String())
	if !report.AllCovered() {
		os.Exit(1)
	}
}
