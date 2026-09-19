// ConditionalWriter (T-0044, TR-010): a storage-backend abstraction for
// writing a whole Protodoc file to a possibly concurrently-written external
// store (an object bucket, a shared filesystem) CONDITIONALLY on an expected
// prior state, so a stale writer cannot silently overwrite a newer one
// (last-writer-wins). This is deliberately distinct from FR-117's
// file-INTERNAL 7-slot commit ring, which detects a torn write WITHIN one
// file: ConditionalWriter concerns concurrent whole-file replacement across
// writers sharing an external store. See docs (T-0049) for the full
// FR-117-vs-TR-010 distinction.
//
// This is plan.md Section 9 Conflict 5's approved storage-backend
// abstraction for TR-010 (Eyvar García, 2026-09-19); see T-0049's design
// note for the full FR-117-vs-TR-010 rationale.
package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Token is an opaque conditional-write token identifying a stored object's
// current state (an ETag / generation-number analogue). The empty Token is
// the sentinel for "object does not yet exist": a create is expressed as a
// conditional write whose expected token is empty.
type Token string

// ConditionalWriter writes whole-file content to an external store
// conditionally on the object's expected prior state. Write replaces the
// object's content with newContent only if expected matches the object's
// current token; on success it returns the new token. On a mismatch it
// makes no change and returns a *ConditionalWriteConflict naming the
// current holder's token (TR-010: refuse naming the current holder, never
// overwrite last-writer-wins).
type ConditionalWriter interface {
	// Write conditionally replaces the object's content. expected is the
	// token the caller believes is current (empty to create a new object).
	// On success the object holds newContent and the returned Token is its
	// new token. On a conditional failure it returns a
	// *ConditionalWriteConflict and leaves the object unchanged.
	Write(expected Token, newContent []byte) (Token, error)

	// CurrentToken returns the object's current token, or the empty Token
	// if it does not exist.
	CurrentToken() (Token, error)
}

// ConditionalWriteConflict is TR-010's refusal: the expected token did not
// match the object's current token, so the write was refused. It names the
// current holder's token so the caller sees whose state it was racing.
type ConditionalWriteConflict struct {
	Expected Token
	Current  Token
}

func (e *ConditionalWriteConflict) Error() string {
	return fmt.Sprintf("ledger: conditional write refused, expected token %q but current holder's token is %q (TR-010: refusing last-writer-wins overwrite)", e.Expected, e.Current)
}

// ErrConditionalWriteConflict allows errors.Is matching against the
// conflict class independent of the specific tokens.
var ErrConditionalWriteConflict = errors.New("ledger: conditional write conflict")

func (e *ConditionalWriteConflict) Is(target error) bool {
	return target == ErrConditionalWriteConflict
}

// tokenOf derives a stored object's token from its content: the hex
// SHA-256 of the bytes. A content-derived token is a valid ETag analogue
// (it changes iff the content changes) and needs no separate persisted
// counter. The empty content maps to the empty Token so a never-written
// object and an explicitly-empty one are distinguished by existence, not by
// token value.
func tokenOf(content []byte) Token {
	sum := sha256.Sum256(content)
	return Token(hex.EncodeToString(sum[:]))
}

// LocalFileConditionalWriter is the reference ConditionalWriter adapter over
// a single local-filesystem path. Its token is the SHA-256 of the file's
// current content; a nonexistent file has the empty token. It is the
// reference implementation the CLI (M18) and the TR-010 conformance test
// exercise; a real deployment would supply a bucket-backed adapter with the
// store's native ETag/generation token instead.
type LocalFileConditionalWriter struct {
	Path string
}

// CurrentToken returns the token of the file's current content, or the
// empty Token if the file does not exist.
func (w *LocalFileConditionalWriter) CurrentToken() (Token, error) {
	content, err := os.ReadFile(w.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ledger: reading current content of %s: %w", w.Path, err)
	}
	return tokenOf(content), nil
}

// Write performs the conditional whole-file replacement. It reads the
// current token, refuses with a *ConditionalWriteConflict if it differs
// from expected (leaving the file untouched), and otherwise writes
// newContent via a temp-file-and-rename so a crash mid-write cannot leave a
// partially written file. It returns the new content's token on success.
func (w *LocalFileConditionalWriter) Write(expected Token, newContent []byte) (Token, error) {
	current, err := w.CurrentToken()
	if err != nil {
		return "", err
	}
	if current != expected {
		return "", &ConditionalWriteConflict{Expected: expected, Current: current}
	}

	// Write to a temp file in the same directory, then atomically rename
	// over the target so a reader never observes a half-written file.
	dir := filepath.Dir(w.Path)
	tmp, err := os.CreateTemp(dir, ".pdl-condwrite-*")
	if err != nil {
		return "", fmt.Errorf("ledger: creating temp file for conditional write: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if the rename below succeeded

	if _, err := tmp.Write(newContent); err != nil {
		tmp.Close()
		return "", fmt.Errorf("ledger: writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("ledger: closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, w.Path); err != nil {
		return "", fmt.Errorf("ledger: renaming temp file over target: %w", err)
	}
	return tokenOf(newContent), nil
}

// Compile-time assertion that the adapter satisfies the interface.
var _ ConditionalWriter = (*LocalFileConditionalWriter)(nil)
