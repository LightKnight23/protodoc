// Command protodoc-diffconform is T-0350's runnable form of the M19/
// NFR-028 differential conformance harness (pkg/diffconform): it runs two
// implementation binaries against the harness's built-in corpus
// (pkg/diffconform.Corpus, version pkg/diffconform.CorpusVersion) and
// exits non-zero the moment either implementation disagrees with the
// other on any case's canonical octets or reported verdict.
//
// Usage:
//
//	protodoc-diffconform --verb=publish <implA-binary> <implB-binary>
//
// Each implementation binary must follow pkg/diffconform.BinaryRunner's
// invocation protocol.
package main

import (
	"flag"
	"fmt"
	"os"

	"Protodoc/pkg/diffconform"
)

func main() {
	verb := flag.String("verb", "publish", "TR-012 verb to invoke on both implementation binaries")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: protodoc-diffconform [--verb=<verb>] <implA-binary> <implB-binary>")
		os.Exit(2)
	}

	a := diffconform.BinaryRunner{Path: args[0], Verb: *verb}
	b := diffconform.BinaryRunner{Path: args[1], Verb: *verb}

	report, err := diffconform.Run(diffconform.CorpusVersion, diffconform.Corpus(), a, b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "protodoc-diffconform: %v\n", err)
		os.Exit(2)
	}

	fmt.Println(report.String())
	if !report.AllMatch() {
		os.Exit(1)
	}
}
