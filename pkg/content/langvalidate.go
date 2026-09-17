// PD-LANG-001 validator rule (T-0082, FR-031): a text span whose
// language-tag reference does not resolve to any language tag registered in
// the document is rejected, naming the rule and the offending unit id. This
// is the resolution check that complements T-0081's mandatory-field check:
// T-0081 ensures a reference is present; PD-LANG-001 ensures it resolves.
package content

import (
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// LanguageRegistry resolves a LangRef to its language tag. It models the
// document's registry excerpt (the set of BCP-47 tags the document
// declares); a reference not present in it is unresolvable. The concrete
// registry wiring lands with the RegistryExcerpt resource; this interface is
// the resolution boundary the validator rule depends on.
type LanguageRegistry interface {
	// ResolveLang returns the language tag for ref and whether ref resolves.
	ResolveLang(ref LangRef) (tag string, ok bool)
}

// MapLanguageRegistry is a simple in-memory LanguageRegistry backed by a
// map, used by tests and by callers assembling a registry from a decoded
// excerpt.
type MapLanguageRegistry map[LangRef]string

// ResolveLang implements LanguageRegistry.
func (m MapLanguageRegistry) ResolveLang(ref LangRef) (string, bool) {
	tag, ok := m[ref]
	return tag, ok
}

// ErrUnresolvableLanguageTag is rule PD-LANG-001: a text span's language-tag
// reference does not resolve to any registered tag.
var ErrUnresolvableLanguageTag = errors.New("content: language-tag reference does not resolve to any registered tag (rule PD-LANG-001)")

// UnresolvableLanguageError names the offending unit id and its unresolvable
// reference, per PD-LANG-001.
type UnresolvableLanguageError struct {
	UnitID pdlfmt.UnitID
	Ref    LangRef
}

func (e *UnresolvableLanguageError) Error() string {
	return fmt.Sprintf("content: unit %x language-tag reference %d does not resolve (rule PD-LANG-001)", e.UnitID, e.Ref)
}

func (e *UnresolvableLanguageError) Is(target error) bool {
	return target == ErrUnresolvableLanguageTag
}

// ValidateLanguageResolves implements PD-LANG-001 for one text span: it
// requires the reference to be present (FR-031, delegating to the T-0081
// check) AND to resolve against reg, returning an *UnresolvableLanguageError
// naming spanID if it does not. It runs at the validation trust boundary,
// after structural decode.
func ValidateLanguageResolves(spanID pdlfmt.UnitID, ref LangRef, reg LanguageRegistry) error {
	if err := ValidateLanguageRefPresent(spanID, ref); err != nil {
		return err
	}
	if _, ok := reg.ResolveLang(ref); !ok {
		return &UnresolvableLanguageError{UnitID: spanID, Ref: ref}
	}
	return nil
}
