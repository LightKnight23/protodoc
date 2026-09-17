package ledger

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestTR_010_ConditionalWriterInterfaceContract is T-0044's named test. It
// exercises the ConditionalWriter contract through the local-filesystem
// reference adapter: a create, a successful conditional update using the
// returned token, and a conflicting concurrent write attempt using the now
// stale token, which must be refused.
func TestTR_010_ConditionalWriterInterfaceContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.pdl")
	var w ConditionalWriter = &LocalFileConditionalWriter{Path: path}

	// The object does not exist yet: current token is empty.
	cur, err := w.CurrentToken()
	if err != nil {
		t.Fatalf("CurrentToken (before create): %v", err)
	}
	if cur != "" {
		t.Fatalf("nonexistent object token = %q, want empty", cur)
	}

	// Create: conditional write with the empty expected token succeeds.
	v1 := []byte("version one content")
	tok1, err := w.Write("", v1)
	if err != nil {
		t.Fatalf("create write: %v", err)
	}
	if tok1 == "" {
		t.Fatalf("create returned empty token")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(v1) {
		t.Fatalf("after create, file content = %q (err %v), want %q", got, err, v1)
	}

	// Successful update: conditional write with the current token succeeds
	// and yields a new token.
	v2 := []byte("version two content, longer than one")
	tok2, err := w.Write(tok1, v2)
	if err != nil {
		t.Fatalf("update write with correct token: %v", err)
	}
	if tok2 == tok1 {
		t.Fatalf("update did not change the token")
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != string(v2) {
		t.Fatalf("after update, file content = %q (err %v), want %q", got, err, v2)
	}

	// Conflicting concurrent write: a second writer still holding the stale
	// tok1 attempts a write. It must be refused, the file left as v2.
	v3 := []byte("version three, from a stale concurrent writer")
	_, err = w.Write(tok1, v3)
	if err == nil {
		t.Fatalf("conditional write with a stale token was not refused")
	}
	if !errors.Is(err, ErrConditionalWriteConflict) {
		t.Fatalf("stale-token write returned %v, want a conditional-write conflict", err)
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != string(v2) {
		t.Fatalf("after refused write, file content = %q (err %v), want unchanged %q", got, err, v2)
	}

	// The current token still matches v2, so a correctly-conditioned write
	// still succeeds after the refusal.
	if _, err := w.Write(tok2, []byte("version four")); err != nil {
		t.Fatalf("write with the current token after a refusal: %v", err)
	}
}
