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

## plan.md gap (still open as of 2026-09-19 -- flagged for an Eyvar-approved amendment)

The frozen `plan.md` describes **no** storage-backend abstraction: TR-010 is
repeatedly folded into FR-117's commit-ring discussion, so the plan as
written has no place for the conditional-write obligation that TR-010
actually states. M02 (T-0044/T-0045) implements the minimal missing piece
(`ConditionalWriter` + the local-filesystem reference adapter) so downstream
milestones are not blocked, but this is a genuine gap between the frozen plan
and the implemented code.

**Status update:** this note originally recommended the plan.md amendment
land before M18's CLI surface was built on top of `ConditionalWriter`. M18
is now complete (T-0341 wires `--if-match` conditional writes into the CLI),
and the amendment was never recorded -- `plan.md`'s architecture sections
still describe no storage-backend abstraction. This is a real, still-open
governance gap, surfaced honestly rather than silently closed: the code and
its tests (`TestTR_010_ConditionalWriterInterfaceContract`,
`TestTR_010_ConditionalWriteRefusalNamesCurrentHolder`, and the others cited
in `analysis.md`'s TR-010 row) are correct and pass, but `plan.md` itself has
not been updated to document the abstraction it depends on. It should still
be closed by an **Eyvar-approved plan.md amendment** adding the
storage-backend abstraction to the architecture. This note records the gap;
it does not itself amend the frozen plan.
