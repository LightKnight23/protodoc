package diffconform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BinaryRunner invokes an external implementation binary as a subprocess,
// per this harness's CLI comparison protocol (deliberately shaped after
// contracts/cli.md's own conventions, so a future TR-012-conformant
// second implementation needs no protocol beyond the one it already must
// build: one positional input file, --out for output, one JSON object on
// stdout carrying at least "status" and "exit_code").
//
// Invocation shape: `<Path> <Verb> <input-file> --out <output-file>`,
// plus any caller-supplied Args appended after those four. Path's process
// must write its canonical output octets (if any) to <output-file> and
// print exactly one JSON object to stdout containing "status" (string)
// and "exit_code" (number), matching contracts/cli.md Section 0's stdout
// shape. A verdict-only case (no canonical output produced) is signalled
// by leaving <output-file> absent; BinaryRunner treats a missing output
// file as CaseResult.Canonical == nil, never as an error.
type BinaryRunner struct {
	Path string   // path to the implementation's executable
	Verb string   // the TR-012 verb to invoke (e.g. "validate", "publish")
	Args []string // extra arguments appended after the fixed positional/--out ones
	Env  []string // extra environment variables (appended to os.Environ()); used by
	// this package's own tests to make one compiled fixture binary behave as
	// two distinguishable "implementations" without duplicating its source.
}

// stdoutEnvelope is the subset of contracts/cli.md's stdout JSON object
// this harness needs: "status" and "exit_code". Every other key ("verb",
// "findings", verb-specific payloads) is accepted and ignored.
type stdoutEnvelope struct {
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code"`
}

// Run implements Runner. It writes caseInput to a fresh temp file,
// invokes the configured binary against it, and returns the parsed
// verdict plus whatever canonical octets (if any) the binary wrote to
// its --out path.
func (r BinaryRunner) Run(caseInput []byte) (CaseResult, error) {
	dir, err := os.MkdirTemp("", "diffconform-*")
	if err != nil {
		return CaseResult{}, fmt.Errorf("diffconform: BinaryRunner: creating temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	inPath := filepath.Join(dir, "input.bin")
	outPath := filepath.Join(dir, "output.bin")
	if err := os.WriteFile(inPath, caseInput, 0o600); err != nil {
		return CaseResult{}, fmt.Errorf("diffconform: BinaryRunner: writing input file: %w", err)
	}

	args := append([]string{r.Verb, inPath, "--out", outPath}, r.Args...)
	cmd := exec.Command(r.Path, args...)
	if len(r.Env) > 0 {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	stdout, err := cmd.Output()
	if err != nil {
		return CaseResult{}, fmt.Errorf("diffconform: BinaryRunner: running %s %v: %w", r.Path, args, err)
	}

	var env stdoutEnvelope
	if err := json.Unmarshal(stdout, &env); err != nil {
		return CaseResult{}, fmt.Errorf("diffconform: BinaryRunner: parsing stdout JSON envelope from %s: %w (stdout was %q)", r.Path, err, stdout)
	}

	var canonical []byte
	if out, err := os.ReadFile(outPath); err == nil {
		canonical = out
	} else if !os.IsNotExist(err) {
		return CaseResult{}, fmt.Errorf("diffconform: BinaryRunner: reading output file from %s: %w", r.Path, err)
	}

	return CaseResult{
		Verdict:   Verdict{Status: env.Status, ExitCode: env.ExitCode},
		Canonical: canonical,
	}, nil
}
