// Package cli implements TR-012's Protodoc command-line tool: the verb
// dispatch framework and the single stdout JSON envelope shape every verb
// writes (contracts/cli.md S0). The full 8-value exit-code type and its
// S1.1 precedence resolver are a separate task (T-0326) layered on top of
// this framework; this file defines only the two exit codes the dispatch
// framework itself can produce directly (cli.md S1's OK and USAGE).
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// The two exit codes reachable from dispatch itself, before T-0326's full
// exit-code engine exists: cli.md S1's OK (0) and USAGE (7).
const (
	exitOK    = 0
	exitUsage = 7
)

const (
	statusOK    = "OK"
	statusUsage = "USAGE"
)

// Finding is one entry of every verb's "findings" array (cli.md S0).
// Offset and UnitID are present in the encoded JSON only when the finding
// actually names one (FR-102).
type Finding struct {
	RuleID  string  `json:"rule_id"`
	Offset  *uint64 `json:"offset,omitempty"`
	UnitID  string  `json:"unit_id,omitempty"`
	Message string  `json:"message"`
}

// Result is one verb invocation's complete outcome: the status name and
// exit code this invocation resolves to, the findings it produced, and
// that verb's own verb-specific top-level stdout keys in Extra (e.g.
// `validate`'s "checks", `verify`'s "signatures").
type Result struct {
	Status   string
	ExitCode int
	Findings []Finding
	Extra    map[string]any
}

// Envelope renders r as cli.md S0's stdout JSON object for the named verb.
// The four common keys (verb, status, exit_code, findings) always win over
// an Extra entry of the same name, so a verb-specific payload can never
// shadow the shape every caller depends on.
func Envelope(verb string, r Result) map[string]any {
	findings := r.Findings
	if findings == nil {
		findings = []Finding{}
	}

	obj := make(map[string]any, len(r.Extra)+4)
	for k, v := range r.Extra {
		obj[k] = v
	}
	obj["verb"] = verb
	obj["status"] = r.Status
	obj["exit_code"] = r.ExitCode
	obj["findings"] = findings
	return obj
}

// WriteEnvelope writes r's envelope for verb to w as exactly one JSON
// object, per cli.md S0's stdout-shape rule. This is the single writer
// every verb uses; no verb invents its own stdout encoding.
func WriteEnvelope(w io.Writer, verb string, r Result) error {
	return json.NewEncoder(w).Encode(Envelope(verb, r))
}

// RunFunc executes one verb's own business logic against its remaining
// argv (the verb name itself already consumed) and returns its Result.
// A RunFunc must not write to stdout itself: Dispatch owns the single
// envelope write (cli.md S0). stderr is for diagnostics and, under
// --format=text, the human-readable rendering.
type RunFunc func(args []string, stderr io.Writer) Result

// Verb is one TR-012 subcommand's registration: its name, a one-line usage
// synopsis for --help, and its RunFunc.
type Verb struct {
	Name  string
	Usage string
	Run   RunFunc
}

// notImplemented is the placeholder RunFunc for a verb whose own
// verb-specific wiring task has not landed in this package yet. T-0325 is
// the dispatch framework only ("it contains no verb business logic
// itself"): it registers all 11 TR-012 verb names without importing or
// calling any downstream package. Calling such a verb honestly reports
// USAGE with a named finding rather than fabricating a validate/verify/etc.
// result.
func notImplemented(verb string) RunFunc {
	return func(_ []string, stderr io.Writer) Result {
		fmt.Fprintf(stderr, "protodoc %s: not yet implemented\n", verb)
		return Result{
			Status:   statusUsage,
			ExitCode: exitUsage,
			Findings: []Finding{{
				RuleID:  "TR-012",
				Message: verb + " is not yet wired to its verb-specific package",
			}},
		}
	}
}

// verbs is TR-012's exactly 11 registered subcommands, in TR-012's own
// stated order: "validate, inspect, extract, verify, diff, merge, project,
// redact, publish, sign and migrate".
var verbs = []Verb{
	{Name: "validate", Usage: "protodoc validate <file>", Run: notImplemented("validate")},
	{Name: "inspect", Usage: "protodoc inspect <file>", Run: notImplemented("inspect")},
	{Name: "extract", Usage: "protodoc extract <file> [--to <path>] [--locators]", Run: notImplemented("extract")},
	{Name: "verify", Usage: "protodoc verify <file> [--signature <id>] [--offline]", Run: notImplemented("verify")},
	{Name: "diff", Usage: "protodoc diff <fileA> <fileB> [--format=json|text]", Run: notImplemented("diff")},
	{Name: "merge", Usage: "protodoc merge <base> <a> <b> --out <path>", Run: notImplemented("merge")},
	{Name: "project", Usage: "protodoc project <file> --to <path> [--format=text|html]", Run: notImplemented("project")},
	{Name: "redact", Usage: "protodoc redact <file> --subtree <unit-id> [--subtree <unit-id>...] --out <path>", Run: notImplemented("redact")},
	{Name: "publish", Usage: "protodoc publish <file> --out <path> [--partial]", Run: notImplemented("publish")},
	{Name: "sign", Usage: "protodoc sign <file> --key <ref> --coverage total|subset [--subset-range <start>:<end> ...] --intent <value> --out <path>", Run: notImplemented("sign")},
	{Name: "migrate", Usage: "protodoc migrate <file> --to-major <N> --out <path> [--rescind-and-resign --new-key <ref> --new-param-set <id>]", Run: notImplemented("migrate")},
}

// Names returns the registered verb names, in registration order.
func Names() []string {
	names := make([]string, len(verbs))
	for i, v := range verbs {
		names[i] = v.Name
	}
	return names
}

// Lookup returns the registered Verb named name, if any.
func Lookup(name string) (Verb, bool) {
	for _, v := range verbs {
		if v.Name == name {
			return v, true
		}
	}
	return Verb{}, false
}

// Version is the protodoc tool's own version string, reported by
// `protodoc --version`. It is not one of TR-012's 11 verbs and carries no
// stdout JSON envelope of its own.
const Version = "0.0.0-dev"

// Dispatch parses args (os.Args[1:]) and routes it to the named verb,
// writing that verb's single stdout JSON envelope (cli.md S0) and
// returning the process exit code the caller should use.
//
// Dispatch contains no verb business logic itself: an unrecognized verb
// name is rejected here, before any downstream package is imported or
// called (T-0325's own DoD).
func Dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printGlobalUsage(stderr)
		return exitUsage
	}

	switch args[0] {
	case "--help", "-h":
		printGlobalUsage(stdout)
		return exitOK
	case "--version":
		fmt.Fprintln(stdout, Version)
		return exitOK
	}

	name := args[0]
	v, ok := Lookup(name)
	if !ok {
		fmt.Fprintf(stderr, "protodoc: unrecognized verb %q\n\n", name)
		printGlobalUsage(stderr)
		return exitUsage
	}

	rest := stripGlobalFlags(args[1:])
	for _, a := range rest {
		if a == "--help" || a == "-h" {
			fmt.Fprintln(stdout, v.Usage)
			return exitOK
		}
	}

	result := v.Run(rest, stderr)
	if err := WriteEnvelope(stdout, v.Name, result); err != nil {
		fmt.Fprintf(stderr, "protodoc: writing stdout envelope: %v\n", err)
		return exitUsage
	}
	return result.ExitCode
}

// stripGlobalFlags removes the shared global flags common to every verb
// (--json is accepted everywhere; the JSON envelope is unconditional per
// cli.md S0, so --json is a no-op kept only so a script can pass it
// explicitly) before handing the remaining args to a verb's own RunFunc.
func stripGlobalFlags(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--json" {
			continue
		}
		out = append(out, a)
	}
	return out
}

func printGlobalUsage(w io.Writer) {
	fmt.Fprintln(w, "protodoc <verb> <file> [<file2>] [flags...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "verbs:")
	for _, v := range verbs {
		fmt.Fprintf(w, "  %s\n", v.Usage)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "global flags: --json, --help, --version")
}
