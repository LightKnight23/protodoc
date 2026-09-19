# Using the `protodoc` CLI

This is a practical usage guide, not the normative spec. For the exact wire contract (every flag,
exit code, and stdout field, with requirement IDs), see
[`contracts/cli.md`](../specs/001-protodoc-format-core/contracts/cli.md) — that file is the source
of truth; this one just shows you how to actually run the thing.

**Read the "Known issues" section before you rely on any write-producing verb.** As of 2026-09-19,
the read side of the CLI (`validate`, `inspect`, `extract`, `verify`, `diff`) is solid and tested
against real files. The write side (`project`, `merge`, `redact`, `publish`, `sign`, `migrate`) has a
known, tracked defect: none of them actually write their output file yet. See below before you build
anything on top of them.

## Building it

```bash
git clone https://github.com/LightKnight23/protodoc
cd protodoc
go build -o protodoc ./cmd/protodoc
./protodoc --help
```

Requires Go 1.25+. No external dependencies (`go.mod` has no `require` block).

## Invocation shape

```
protodoc <verb> <file> [<file2>] [flags...]
```

Every verb emits exactly one JSON object to stdout, and nothing else — no banners, no progress lines.
`--format=text` (accepted by every verb) additionally prints a human-readable rendering to *stderr*;
the JSON on stdout is always present regardless, so scripts have one stable shape to parse.

Every result carries at least:

```json
{"verb": "validate", "status": "OK", "exit_code": 0, "findings": []}
```

## Exit codes

| Code | Name | Meaning |
|---|---|---|
| 0 | `OK` | Operation completed successfully. |
| 1 | `INVALID` | The document is structurally malformed (truncated, bad digest, dangling reference, etc.) — but note the current caveat below on nonexistent files. |
| 2 | `UNSUPPORTED` | The document declares a newer format version than this build understands. |
| 3 | `UNVERIFIED` | Structurally valid, but a cryptographic check (signature, revocation, time attestation) failed. |
| 4 | `UNAVAILABLE` | A signature covers a state the file can no longer reconstruct. |
| 5 | `OVER_BUDGET` | A resource ceiling (segment count, nesting depth, etc.) was exceeded. |
| 6 | `REFUSED` | A structurally-fine document's write operation was declined by policy (e.g. re-signing an already-fully-signed document). |
| 7 | `USAGE` | The invocation itself is wrong — missing flag, unreadable path, unknown verb. Never a statement about document content. |

## The 11 verbs

### `validate <file>`

Runs the full structural validation pipeline. No cryptographic checks, no content extraction — just
"is this a well-formed Protodoc file."

```bash
$ protodoc validate my-document.pdl
{"checks":4,"exit_code":0,"findings":[],"status":"OK","verb":"validate"}

$ protodoc validate corrupted-file.pdl
{"checks":1,"exit_code":1,"findings":[{"rule_id":"TR-006","message":"bounded prefix truncated: read 2000 of 1048576 required octets"}],"status":"INVALID","verb":"validate"}
```

### `inspect <file>`

Reads only the leading 1 MiB fixed prefix — never touches content past it. Reports the header,
winning commit-ring slot, frontmatter summary, and the segment inventory.

```bash
$ protodoc inspect my-document.pdl
{"exit_code":0,"findings":[],"prefix":{"frontmatter":{"title":"My Document", ...}, "header":{...}, "ring_winner":{...}, "segments":[]}, "status":"OK","verb":"inspect"}
```

### `extract <file> [--to <path>] [--locators]`

Reproduces the writer's own input text exactly, in reading order.

```bash
$ protodoc extract my-document.pdl
{"exit_code":0,"findings":[],"status":"OK","verb":"extract", ...}
```

### `verify <file> [--signature <id>] [--offline]`

Runs signature verification (EdDSA-Protodoc-1) plus long-term-validation checks for every signature
in the document, or just the one named by `--signature`. `--offline` is the only mode this ever
runs in (no network access exists in any reader path) — the flag exists purely to make that explicit
in scripts.

```bash
$ protodoc verify my-document.pdl
{"exit_code":0,"findings":[],"signatures":[],"status":"OK","verb":"verify"}
```

An unsigned document reports `OK` with an empty `signatures` array — that's expected, not an error.

### `diff <fileA> <fileB> [--format=json|text]`

Enumerates every content unit that differs between two documents.

```bash
$ protodoc diff v1.pdl v2.pdl
```

**Known issue:** the current stdout shape doesn't match the documented contract — see below.

### `merge <base> <a> <b> --out <path>`

Reconciles two divergent edits of a common base document.

```bash
$ protodoc merge base.pdl mine.pdl theirs.pdl --out merged.pdl
```

**Known issue:** doesn't currently write `merged.pdl` — see below.

### `project <file> --to <path> [--format=text|html]`

Produces a one-directional, non-normative text or HTML projection — for viewing, never for
re-ingestion (`reingestable` is always `false` in the response).

```bash
$ protodoc project my-document.pdl --to preview.html --format=html
```

**Known issue:** `--to` isn't currently enforced as required, and the file isn't currently written —
see below.

### `redact <file> --subtree <unit-id> [--subtree <unit-id>...] --out <path>`

Removes a designated redactable subtree and republishes, leaving the redaction cryptographically
provable (not just visually blacked out).

```bash
$ protodoc redact my-document.pdl --subtree 9f2a...c1 --out redacted.pdl
```

**Known issue:** required flags aren't currently enforced and the file isn't currently written — see
below.

### `publish <file> --out <path> [--partial]`

Full compaction to a fresh, canonicalized file. Refused if any signature is present, unless
`--partial` is given (reclaims only segments outside every signature's coverage).

```bash
$ protodoc publish my-document.pdl --out compacted.pdl
```

**Known issue:** doesn't currently write `compacted.pdl` — see below.

### `sign <file> --key <ref> --coverage total|subset [--subset-range <start>:<end> ...] --intent <value> --out <path>`

Creates a new signature. `--key` names a key held by your platform's key store or HSM — the tool
never accepts a raw private key on the command line.

```bash
$ protodoc sign my-document.pdl --key my-signing-key --coverage total --intent author --out signed.pdl
```

**Known issue:** required flags aren't currently enforced and the file isn't currently written — see
below.

### `migrate <file> --to-major <N> --out <path> [--rescind-and-resign --new-key <ref> --new-param-set <id>]`

Migrates a document to a newer major format version. Refusal-first: checks every construct is
representable before writing anything.

```bash
$ protodoc migrate old-document.pdl --to-major 2 --out migrated.pdl
```

**Known issue:** `--to-major` isn't currently enforced as required and the file isn't currently
written — see below.

## Known issues (as of 2026-09-19)

Found via an extensive black-box test (42 cases across all 11 verbs) after the read-path defect fix
landed. Full detail: `specs/CHANGES.md`'s `DEFECT-2026-09-19b` entry; tracked as tasks T-0383–T-0390.

- **No write verb writes its file yet.** `project`, `merge`, `redact`, `publish`, `sign`, `migrate`
  all report `"status":"OK"` without creating the `--out`/`--to` file on disk. Don't build automation
  on top of these until T-0383–T-0388 close.
- **Required flags aren't enforced** on `project`, `redact`, `publish`, `sign`, `migrate` — omitting
  one silently "succeeds" instead of returning `USAGE`. Always pass every flag the syntax above shows,
  even though the tool won't currently stop you if you forget.
- **`merge` reports `REFUSED` (exit 6) for a missing input file**, which the contract reserves for a
  different case entirely. Treat any non-zero exit from `merge` as a real failure regardless of which
  code it uses, until T-0384 closes.
- **`diff`'s JSON doesn't have the documented `identical` field yet** — don't script against it until
  T-0389 closes.
- **A nonexistent file reports `INVALID` on most verbs, `USAGE` on `inspect`.** `inspect`'s answer is
  the contract-correct one; treat exit code 1 vs. 7 as unreliable for "file doesn't exist" until
  T-0390 unifies it. A file that exists but is malformed correctly reports `INVALID` either way.

**What's solid today:** `validate`, `inspect`, `extract`, and `verify` genuinely read and decode real
files, correctly reject garbage/truncated/nonexistent input (module the exit-code nuance above), and
correctly accept real documents. This is the part of the CLI worth relying on right now.
