# Ledger conditional-write design note (TR-010)

This note distinguishes two mechanisms the frozen plan.md conflates, and
records the contract of the `ConditionalWriter` abstraction added in M02
(tasks T-0044/T-0045) for the CLI surface built on top of it in M18.

## FR-117 vs TR-010 distinction

FR-117 and TR-010 solve two different write-safety problems and must not be
conflated (the frozen plan.md repeatedly describes TR-010 as if it were
FR-117; see the plan.md gap note below).

- **FR-117 (file-internal torn-write detection).** Protodoc's fixed prefix
  carries a 7-slot commit ring (`contracts/container.abnf` S3). A commit
  writes the next ring slot; a reader selects the current state as the slot
  with the strictly highest sequence whose `record-digest` verifies and
  whose `ledger-length` does not exceed the actual file length. A writer
  killed mid-commit therefore leaves the file readable at exactly one
  complete state -- the pre-commit winner or the post-commit winner, never a
  blend. This is entirely **within one file**: it needs no shared storage,
  no external coordination, and no compare-and-swap. It detects a *torn
  write* (a partial write to a single file by a single writer that crashed).

- **TR-010 (external shared-storage last-writer-wins avoidance).** When the
  command-line tool writes a whole document file to storage that may be
  written concurrently by another writer (an object bucket, a shared
  filesystem), the ring inside the file cannot help: two writers who both
  opened the same prior state can each produce a complete, internally
  consistent new file, and a plain overwrite silently discards one of them
  (last-writer-wins). TR-010 requires the tool to write **conditionally on
  the expected prior state** and to **refuse, naming the current holder**,
  when that condition fails. This is a property of the writing *tool* and
  the *external store*, not of the file format: a file format has no shared
  storage, so this cannot be a format requirement (spec.md TR-010 rationale).

In one sentence: FR-117 keeps a single file readable across a crash; TR-010
keeps concurrent writers to a shared store from silently clobbering each
other. Both are needed; neither substitutes for the other.

## ConditionalWriter contract

`ledger.ConditionalWriter` (pkg/ledger/conditionalwriter.go) is the minimal
storage-backend abstraction TR-010 needs. CLI implementers (M18) target this
interface rather than re-deriving the distinction above.

- `Write(expected Token, newContent []byte) (Token, error)` replaces the
  object's whole content with `newContent` **only if** `expected` equals the
  object's current token. On success it returns the new token. On a mismatch
  it makes **no change** and returns a `*ConditionalWriteConflict` (matching
  the `ErrConditionalWriteConflict` sentinel) whose `Current` field and error
  message **name the current holder's token** (TR-010's refusal-naming
  requirement).
- `CurrentToken() (Token, error)` returns the object's current token, or the
  empty `Token` if the object does not exist.
- The empty `Token` is the sentinel for "object does not exist": a create is
  a conditional write whose `expected` is the empty token.
- `LocalFileConditionalWriter` is the reference adapter over a local
  filesystem path. Its token is the SHA-256 of the file's current content,
  and it writes via a temp-file-and-rename so a crash mid-write cannot leave
  a partially written file. A real deployment supplies a bucket-backed
  adapter using the store's native ETag / generation-number token instead.

A CLI commit to shared storage therefore reads the current token, performs
the edit against the opened state, and calls `Write(openedToken, newBytes)`;
a conflict is surfaced to the user naming the current holder, never resolved
by silent overwrite.

## plan.md gap (RESOLVED 2026-09-19)

The frozen `plan.md` described **no** storage-backend abstraction as originally written: TR-010 was
repeatedly folded into FR-117's commit-ring discussion, so the plan had no place for the
conditional-write obligation that TR-010 actually states. M02 (T-0044/T-0045) implemented the missing
piece anyway (`ConditionalWriter` + the local-filesystem reference adapter) so downstream milestones
were not blocked, and M18's CLI surface (T-0341, `--if-match`) was built on top of it before the
amendment landed -- a genuine gap between the frozen plan and the implemented code that this note
originally flagged.

**Resolution:** Eyvar approved `plan.md` Section 9's new Conflict 5, 2026-09-19, formally adopting
`ledger.ConditionalWriter` as the architecture's storage-backend abstraction for TR-010. `plan.md` now
documents the mechanism this note describes; see `plan.md` Section 9 Conflict 5 for the approved text
and `specs/CHANGES.md` for the structured entry. This note is retained as the original design
rationale, not superseded by the amendment.
