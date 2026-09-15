// Command protodoc is the reference command-line tool for the Protodoc
// document format (TR-012, contracts/cli.md).
package main

import (
	"os"

	"Protodoc/pkg/cli"
)

func main() {
	os.Exit(cli.Dispatch(os.Args[1:], os.Stdout, os.Stderr))
}
