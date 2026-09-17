// Per-unit language tag emission (T-0090, FR-043): the extraction view
// resolves each emitted text unit's language reference to exactly one
// language tag via LanguageOf. A unit whose reference is unresolvable is
// reported through a distinct per-unit error value in the result rather than
// being silently omitted, so a consumer sees the gap.
package extract

import (
	"Protodoc/pkg/content"
	"Protodoc/pkg/pdlfmt"
)

// UnitLanguage is the per-unit language result: the unit's identity, its
// resolved language tag (empty when Err != nil), and a per-unit error set
// when the reference is unset or unresolvable. Emitting an error value
// (rather than dropping the unit) is what keeps every unit accounted for.
type UnitLanguage struct {
	UnitID pdlfmt.UnitID
	Tag    string
	Err    error
}

// LanguageOf resolves the language tag for a text unit identified by unitID
// with language reference ref, against reg. It returns a UnitLanguage whose
// Tag is the single resolved tag on success, or whose Err names the failure
// (unset reference, or unresolvable) on failure -- never both, and never a
// silently-omitted unit.
func LanguageOf(unitID pdlfmt.UnitID, ref content.LangRef, reg content.LanguageRegistry) UnitLanguage {
	if err := content.ValidateLanguageResolves(unitID, ref, reg); err != nil {
		return UnitLanguage{UnitID: unitID, Err: err}
	}
	tag, _ := reg.ResolveLang(ref)
	return UnitLanguage{UnitID: unitID, Tag: tag}
}

// LanguagesForRuns resolves the language of each run in reading order,
// returning one UnitLanguage per run (each with exactly one resolved tag or
// one error). No run is omitted.
func LanguagesForRuns(runs []content.Run, reg content.LanguageRegistry) []UnitLanguage {
	out := make([]UnitLanguage, 0, len(runs))
	for _, r := range runs {
		out = append(out, LanguageOf(r.RunID, r.LangRef, reg))
	}
	return out
}
