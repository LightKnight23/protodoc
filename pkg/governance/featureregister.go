// Feature register for CON-019/A-FEATURE (T-0354): the register that tracks,
// per named feature/mechanism in data-model.md, whether two source-
// independent implementations have passed every conformance case for it.
// Every feature starts Provisional; RecordTwoImplPass is the only path to
// Normative, per CON-019's Verify clause ("the feature register records two
// independent passing implementations before any feature is marked
// normative").
package governance

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"Protodoc/pkg/diffconform"
)

// dataModelPath locates data-model.md relative to this package directory,
// following the same "read the live document off disk, never a hardcoded
// copy" convention this package already uses for the NFR-028 protocol
// drift check in secondimplprotocol_test.go.
const dataModelPath = "../../specs/001-protodoc-format-core/data-model.md"

// entityHeadingPattern matches a data-model.md Section 2 entity heading,
// e.g. "### 2.9 Annotation / Range" or "### 2.22 PageDirectory entry
// (derived, non-normative)".
var entityHeadingPattern = regexp.MustCompile(`^### 2\.\d+ (.+)$`)

// FeatureStatus is CON-019's two-value gate. There is no "regressed" or
// "revoked" value: a Register never moves a feature back to Provisional
// once Normative, since CON-019 only ever restricts marking normative, it
// does not define an un-marking process.
type FeatureStatus string

const (
	StatusProvisional FeatureStatus = "provisional"
	StatusNormative   FeatureStatus = "normative"
)

// LoadFeatureNames reads path (data-model.md) and returns the name of every
// Section 2 entity heading that is not itself declared non-normative in its
// own heading parenthetical (e.g. "(derived, non-normative)", "(encoding
// primitive, not a stored entity)"). Those entities are already excluded
// from the normative track by data-model.md itself, so CON-019's two-
// implementation gate -- which exists to police the *promotion* of a
// feature to normative status -- has nothing to gate for them.
func LoadFeatureNames(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, line := range strings.Split(string(raw), "\n") {
		m := entityHeadingPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		heading := strings.TrimSpace(m[1])
		if isDeclaredNonNormative(heading) {
			continue
		}
		names = append(names, stripParenthetical(heading))
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("governance: no Section 2 entity headings found in %s", path)
	}
	return names, nil
}

func isDeclaredNonNormative(heading string) bool {
	return strings.Contains(heading, "non-normative") || strings.Contains(heading, "not a stored entity")
}

func stripParenthetical(heading string) string {
	if i := strings.Index(heading, " ("); i != -1 {
		return heading[:i]
	}
	return heading
}

// Register is CON-019's feature register: one status per named feature. A
// freshly built Register never contains a Normative entry -- the only way a
// feature reaches Normative is RecordTwoImplPass.
type Register map[string]FeatureStatus

// NewRegister builds a Register with every name in names starting
// Provisional (A-FEATURE's required starting state).
func NewRegister(names []string) Register {
	r := make(Register, len(names))
	for _, n := range names {
		r[n] = StatusProvisional
	}
	return r
}

// ImplementationID identifies one implementation for the mechanical half of
// source-independence checking. This package cannot verify "shares no
// source code" from a string alone -- CON-019's full independence bar is
// adjudicated by the NFR-028 protocol's human review (docs/nfr-028-second-
// implementation-protocol.md's Independence criteria section).
// RecordTwoImplPass only enforces the always-checkable half: the two IDs
// must differ literally, so a single implementation's run can never satisfy
// the gate by itself.
type ImplementationID string

// TwoImplPassEvidence is one differential-harness run (pkg/diffconform)
// offered as evidence that a named feature may be promoted to Normative.
type TwoImplPassEvidence struct {
	Feature string
	ImplA   ImplementationID
	ImplB   ImplementationID
	Report  diffconform.Report
}

// RecordTwoImplPass applies ev to r. On success it flips ev.Feature to
// Normative and returns nil; on any of the following it returns an error
// and leaves r completely unchanged:
//
//   - ev.Feature is not a registered feature;
//   - either implementation ID is unset;
//   - ImplA equals ImplB (single-implementation coverage -- CON-019's exact
//     failure case: "a feature specified but implemented once");
//   - the report covers zero corpus cases;
//   - the report records any mismatch at all -- per CON-019's Verify
//     clause this is a 100% pass bar, not a threshold below it.
func (r Register) RecordTwoImplPass(ev TwoImplPassEvidence) error {
	if _, ok := r[ev.Feature]; !ok {
		return fmt.Errorf("governance: unknown feature %q", ev.Feature)
	}
	if ev.ImplA == "" || ev.ImplB == "" {
		return fmt.Errorf("governance: feature %q: both implementation IDs must be set", ev.Feature)
	}
	if ev.ImplA == ev.ImplB {
		return fmt.Errorf("governance: feature %q: ImplA and ImplB are the same implementation (%q); CON-019 requires two source-independent implementations", ev.Feature, ev.ImplA)
	}
	if ev.Report.Total == 0 {
		return fmt.Errorf("governance: feature %q: report covers zero corpus cases", ev.Feature)
	}
	if !ev.Report.AllMatch() {
		return fmt.Errorf("governance: feature %q: differential report has %d mismatched case(s); CON-019 requires a 100%% pass", ev.Feature, len(ev.Report.Diffs))
	}
	r[ev.Feature] = StatusNormative
	return nil
}
