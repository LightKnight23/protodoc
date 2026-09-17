package ledger

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTR_010_ConditionalWriteRefusalNamesCurrentHolder is T-0045's named
// test. It confirms the refusal path of ConditionalWriter: a conditional
// write with a stale expected token is refused with an error that names the
// current holder's actual token (both as a structured field and in the
// message text), and no octet of the target file is modified.
func TestTR_010_ConditionalWriteRefusalNamesCurrentHolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.pdl")
	w := &LocalFileConditionalWriter{Path: path}

	// Establish a current state written by "holder A".
	holderAContent := []byte("holder A's committed document state")
	tokenA, err := w.Write("", holderAContent)
	if err != nil {
		t.Fatalf("initial write: %v", err)
	}

	// Snapshot the on-disk bytes so we can prove they are untouched by the
	// refused write.
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading committed state: %v", err)
	}

	// "Holder B" attempts a write believing the state is still empty (a
	// stale expected token). This must be refused.
	staleToken := Token("")
	refusalErr := func() error {
		_, e := w.Write(staleToken, []byte("holder B's would-be overwrite"))
		return e
	}()
	if refusalErr == nil {
		t.Fatalf("stale-token write was not refused")
	}

	// The error must be a conditional-write conflict naming the current
	// holder's token.
	var conflict *ConditionalWriteConflict
	if !errors.As(refusalErr, &conflict) {
		t.Fatalf("refusal error is %T, want *ConditionalWriteConflict", refusalErr)
	}
	if conflict.Current != tokenA {
		t.Fatalf("conflict names current token %q, want holder A's token %q", conflict.Current, tokenA)
	}
	if conflict.Expected != staleToken {
		t.Fatalf("conflict names expected token %q, want %q", conflict.Expected, staleToken)
	}
	// The current holder's token must also appear in the human-readable
	// message (TR-010: refuse NAMING the current holder).
	if !strings.Contains(refusalErr.Error(), string(tokenA)) {
		t.Fatalf("refusal message %q does not name the current holder token %q", refusalErr.Error(), tokenA)
	}

	// No octet of the target file may have changed.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading state after refusal: %v", err)
	}
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatalf("target file was modified by a refused conditional write")
	}
	if string(after) != string(holderAContent) {
		t.Fatalf("file content changed after refusal: got %q, want %q", after, holderAContent)
	}

	// errors.Is against the sentinel also matches, for callers that only
	// need to know a conflict occurred.
	if !errors.Is(refusalErr, ErrConditionalWriteConflict) {
		t.Fatalf("refusal does not match ErrConditionalWriteConflict sentinel")
	}
}
