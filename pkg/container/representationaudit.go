// CON-005 audit: at most one normative representation per semantic
// capability (spec.md CON-005, audit A-DUP). This file walks the exported
// wire-facing struct types in pkg/container and pkg/pdlfmt and classifies
// every field into one of the capability categories T-0026's DoD names
// (numeric, repeated-value, geometric), then TestCON_005_
// ExactlyOneRepresentationPerCapability asserts each category is expressed
// by exactly one Go-level representation form: an integer for "numeric"
// and "geometric" (never a floating-point field alongside it, CON-012),
// and a sequence for "repeated-value" (never an associative map alongside
// it, no map type exists anywhere in this vocabulary today, contracts/
// container.abnf S2: plain-seq-of-X and sorted-vec-of-X are its only two
// repeated-value forms).
package container

import (
	"reflect"

	"Protodoc/pkg/pdlfmt"
)

// Capability names one of the semantic capability categories the CON-005
// audit enumerates (T-0026 DoD: numeric, repeated-value, geometric).
type Capability string

const (
	CapabilityNumericValue   Capability = "numeric"
	CapabilityRepeatedValue  Capability = "repeated-value"
	CapabilityGeometricValue Capability = "geometric"
)

// representationForm names a Go-level representation form a struct field
// can take. permittedForm below names the single one of these CON-005
// allows per Capability; formFloatingPoint and formAssociativeMap exist
// only so a violation can be named, they are never a permittedForm value.
type representationForm string

const (
	formInteger        representationForm = "integer"
	formFloatingPoint  representationForm = "floating-point" // forbidden: no float alongside a fixed-point/integer form (CON-005, CON-012)
	formSequence       representationForm = "sequence"
	formAssociativeMap representationForm = "associative-map" // forbidden: no map alongside plain-seq-of-X/sorted-vec-of-X (CON-005)
)

// permittedForm is the single normative representation form CON-005 allows
// per capability category.
var permittedForm = map[Capability]representationForm{
	CapabilityNumericValue:   formInteger,
	CapabilityRepeatedValue:  formSequence,
	CapabilityGeometricValue: formInteger,
}

// geometricFields names the specific (struct type name -> field name) pairs
// this codebase currently defines as carrying a CON-012 geometric value:
// Frontmatter.PageWidth/PageHeight (fm-page-dimensions, contracts/
// container.abnf S4). reflect can classify a field's Go Kind but has no way
// to know "this int64 means 1/914400-inch units" on its own, so this list
// is what lets the audit tell CapabilityGeometricValue apart from the
// general CapabilityNumericValue bucket; a field not listed here is only
// ever audited as numeric.
var geometricFields = map[string]map[string]bool{
	"Frontmatter": {"PageWidth": true, "PageHeight": true},
}

// auditedTypes is the closed set of exported struct types in pkg/container
// and pkg/pdlfmt whose fields carry wire-level (or wire-derived) capability
// values. walkType recurses into any nested struct or slice/array-of-struct
// field it finds (e.g. Frontmatter.CoverageSummary reaches
// CoverageSummaryEntry, which in turn reaches SegmentRange), so this list
// only needs each type's outermost entry point, not every type it contains.
var auditedTypes = []reflect.Type{
	reflect.TypeOf(Header{}),
	reflect.TypeOf(CommitRingRecord{}),
	reflect.TypeOf(Frontmatter{}),
	reflect.TypeOf(SegmentTableSlot{}),
	reflect.TypeOf(StoredUnit{}),
	reflect.TypeOf(pdlfmt.Field{}),
}

// representationObservation is one struct field the audit classified into
// a capability and the representation form it actually uses.
type representationObservation struct {
	TypeName string
	Field    string
	Category Capability
	Form     representationForm
}

// classifyKind maps a reflect.Kind to the representationForm it expresses,
// and reports whether the audit has an opinion about that Kind at all: a
// bool, string or nested-struct-typed field is not itself a numeric,
// repeated-value or geometric capability and is skipped, not misclassified.
func classifyKind(k reflect.Kind) (representationForm, bool) {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return formInteger, true
	case reflect.Float32, reflect.Float64:
		return formFloatingPoint, true
	case reflect.Slice, reflect.Array:
		return formSequence, true
	case reflect.Map:
		return formAssociativeMap, true
	default:
		return "", false
	}
}

// walkType recursively classifies every exported field of t (dereferencing
// pointers, and recursing into slice/array element types and nested struct
// types) into representationObservations, appending to out. visited
// prevents infinite recursion and duplicate work on a type reachable more
// than one way (e.g. SegmentRange, reachable via CoverageSummaryEntry).
func walkType(typeName string, t reflect.Type, geo map[string]bool, visited map[reflect.Type]bool, out *[]representationObservation) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct || visited[t] {
		return
	}
	visited[t] = true

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		ft := f.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if form, ok := classifyKind(ft.Kind()); ok {
			category := CapabilityNumericValue
			switch form {
			case formSequence, formAssociativeMap:
				category = CapabilityRepeatedValue
			case formInteger, formFloatingPoint:
				if geo[f.Name] {
					category = CapabilityGeometricValue
				}
			}
			*out = append(*out, representationObservation{TypeName: typeName, Field: f.Name, Category: category, Form: form})
		}

		switch ft.Kind() {
		case reflect.Struct:
			walkType(ft.Name(), ft, geometricFields[ft.Name()], visited, out)
		case reflect.Slice, reflect.Array:
			et := ft.Elem()
			for et.Kind() == reflect.Ptr {
				et = et.Elem()
			}
			if et.Kind() == reflect.Struct {
				walkType(et.Name(), et, geometricFields[et.Name()], visited, out)
			}
		}
	}
}

// auditRepresentations walks auditedTypes and returns every field
// observation classified into one of CON-005's three DoD capability
// categories. TestCON_005_ExactlyOneRepresentationPerCapability is the gate
// that reads this output.
func auditRepresentations() []representationObservation {
	var out []representationObservation
	visited := make(map[reflect.Type]bool)
	for _, t := range auditedTypes {
		walkType(t.Name(), t, geometricFields[t.Name()], visited, &out)
	}
	return out
}
