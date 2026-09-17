package content

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// fieldRole is one entry in the A-FIELD-ROLE registry (plan.md Section 10
// item 2): the classification of a content-package struct field into the
// closed 7-tag vocabulary. Crucially there is NO POSITION role: a field can
// be an identity component or a physical offset/length, but never a
// persisted document-wide counted text position (CON-001).
type fieldRole string

const (
	roleIdentityComponent fieldRole = "IDENTITY-COMPONENT" // run_id, base_ordinal, unit ids
	roleOrdinal           fieldRole = "ORDINAL"            // storage/allocation ordinal (not a text position)
	roleDigest            fieldRole = "DIGEST"             // a content digest
	roleEnum              fieldRole = "ENUM"               // a closed enum value
	roleScalarValue       fieldRole = "SCALAR-VALUE"       // literal text/scalar content
	roleCount             fieldRole = "COUNT"              // a cardinality, not a position from a start
	rolePhysicalOffsetLen fieldRole = "PHYSICAL-OFFSET-OR-LENGTH"
	// rolePersistedCountedPosition is the FORBIDDEN role: a field that
	// persists a document-wide counted text position. No field may carry it;
	// the audit fails if any does. It exists only so the audit can name the
	// prohibited class it checks for.
	rolePersistedCountedPosition fieldRole = "PERSISTED-COUNTED-POSITION"
)

// fieldRoleRegistry classifies every exported field of every exported struct
// in the content package, keyed "Type.Field". The audit below fails if a
// field is missing from this registry (untagged) or if any field is tagged
// with the forbidden rolePersistedCountedPosition.
var fieldRoleRegistry = map[string]fieldRole{
	// Anchor: pure content identity.
	"Anchor.RunID":       roleIdentityComponent,
	"Anchor.BaseOrdinal": roleIdentityComponent,

	// AnchorPoint: identity + closed enums; birth-ordinal is IDENTITY-COMPONENT.
	"AnchorPoint.RunID":        roleIdentityComponent,
	"AnchorPoint.BirthOrdinal": roleIdentityComponent,
	"AnchorPoint.Side":         roleEnum,
	"AnchorPoint.Boundary":     roleEnum,

	// OrphanRecord: captured context at orphaning time.
	"OrphanRecord.Author":     roleIdentityComponent, // authorship id, not a position
	"OrphanRecord.QuotedText": roleScalarValue,
	"OrphanRecord.Prev":       roleIdentityComponent,
	"OrphanRecord.Next":       roleIdentityComponent,

	// Annotation: identities + embedded structures + a flag.
	"Annotation.ID":        roleIdentityComponent,
	"Annotation.Start":     roleIdentityComponent, // an AnchorPoint (identity-addressed)
	"Annotation.End":       roleIdentityComponent,
	"Annotation.BodyBlock": roleIdentityComponent,
	"Annotation.Orphaned":  roleEnum, // boolean state flag
	"Annotation.Orphan":    roleScalarValue,

	// OrphanContext.
	"OrphanContext.Author":     roleIdentityComponent,
	"OrphanContext.QuotedText": roleScalarValue,

	// TextBlock.
	"TextBlock.BlockID": roleIdentityComponent,
	"TextBlock.LangRef": roleEnum, // registry-issued reference, a closed lookup value
	"TextBlock.Runs":    roleScalarValue,

	// UnresolvableLanguageError.
	"UnresolvableLanguageError.UnitID": roleIdentityComponent,
	"UnresolvableLanguageError.Ref":    roleEnum,

	// Run: the canonical case -- run_id and base_ordinal are BOTH
	// IDENTITY-COMPONENT (never positional arithmetic from a sequence start).
	"Run.RunID":       roleIdentityComponent,
	"Run.BaseOrdinal": roleIdentityComponent,
	"Run.Text":        roleScalarValue,
	"Run.LangRef":     roleEnum,

	// NonNFCError.
	"NonNFCError.RunID": roleIdentityComponent,
	"NonNFCError.Text":  roleScalarValue,

	// Table / CellEntry: row and column ids are content identities (minted
	// once, reorderable like a run_id), never positional indices.
	"Table.ID":          roleIdentityComponent,
	"Table.Rows":        roleIdentityComponent,
	"Table.Columns":     roleIdentityComponent,
	"Table.Cells":       roleScalarValue,
	"CellEntry.Row":     roleIdentityComponent,
	"CellEntry.Col":     roleIdentityComponent,
	"CellEntry.Content": roleIdentityComponent,

	// TableTilingError.
	"TableTilingError.Kind": roleEnum,
	"TableTilingError.Row":  roleIdentityComponent,
	"TableTilingError.Col":  roleIdentityComponent,

	// Note.
	"Note.ID":        roleIdentityComponent,
	"Note.Anchor":    roleIdentityComponent,
	"Note.BodyBlock": roleIdentityComponent,
	"Note.Placement": roleEnum,

	// CrossReference.
	"CrossReference.ID":     roleIdentityComponent,
	"CrossReference.Target": roleIdentityComponent,
	"CrossReference.Kind":   roleEnum,
	"CrossReference.Anchor": roleIdentityComponent,
}

// TestCON_001_AFieldRoleAuditHasNoPersistedCountedPosition is T-0083's named
// test. It enumerates every exported field of every exported struct in the
// content package and asserts (1) each field is classified in the
// A-FIELD-ROLE registry (no untagged field), and (2) no field is classified
// as a persisted document-wide counted position (CON-001). The audit fails
// CI if a new field is added without a registry entry, so the content model
// cannot grow a counted position unnoticed.
func TestCON_001_AFieldRoleAuditHasNoPersistedCountedPosition(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading content package dir: %v", err)
	}

	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", e.Name(), err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if !name.IsExported() {
						continue
					}
					key := ts.Name.Name + "." + name.Name
					seen[key] = true
					role, classified := fieldRoleRegistry[key]
					if !classified {
						t.Errorf("A-FIELD-ROLE audit: %s is untagged; every content struct field must be classified (CON-001, plan.md Section 10)", key)
						continue
					}
					if role == rolePersistedCountedPosition {
						t.Errorf("A-FIELD-ROLE audit: %s is classified as a persisted document-wide counted position, which CON-001 forbids", key)
					}
				}
			}
			return true
		})
	}

	// Every registry entry must correspond to a real field (catch stale
	// entries so the registry cannot drift into fiction).
	for key := range fieldRoleRegistry {
		if !seen[key] {
			t.Errorf("A-FIELD-ROLE registry entry %q names a field that no longer exists", key)
		}
	}
}
