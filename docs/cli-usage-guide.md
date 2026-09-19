# Using the `protodoc` CLI

This is a practical usage guide, not the normative spec. For the exact wire contract (every flag,
exit code, and stdout field, with requirement IDs), see
[`contracts/cli.md`](../specs/001-protodoc-format-core/contracts/cli.md) — that file is the source
of truth; this one just shows you how to actually run the thing.

All 11 verbs are wired to real file I/O and independently verified end-to-end as of 2026-09-19
(`DEFECT-2026-09-19` and `DEFECT-2026-09-19b` in `specs/CHANGES.md`, tasks T-0373–T-0390). Two
narrower, honestly-scoped gaps remain — see "Honest limitations" at the bottom before relying on
`merge` or `sign`.

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
| 1 | `INVALID` | The document exists and opened, but is structurally malformed (truncated, bad digest, dangling reference, etc.). |
| 2 | `UNSUPPORTED` | The document declares a newer format version than this build understands. |
| 3 | `UNVERIFIED` | Structurally valid, but a cryptographic check (signature, revocation, time attestation) failed. |
| 4 | `UNAVAILABLE` | A signature covers a state the file can no longer reconstruct. |
| 5 | `OVER_BUDGET` | A resource ceiling (segment count, nesting depth, etc.) was exceeded. |
| 6 | `REFUSED` | A structurally-fine document's write operation was declined by policy (e.g. re-signing an already-fully-signed document, or a non-forward `migrate`). |
| 7 | `USAGE` | The invocation itself is wrong — missing/invalid flag, unreadable or nonexistent path, unknown verb. Never a statement about a document's actual content. |

Every verb reports `USAGE` (not `INVALID`) for a file that doesn't exist or can't be opened at all —
`INVALID` is reserved for a file that opens fine but is malformed inside.

## The 11 verbs

### `validate <file>`

Runs the full structural validation pipeline. No cryptographic checks, no content extraction — just
"is this a well-formed Protodoc file."

```bash
$ protodoc validate my-document.pdl
{"checks":4,"exit_code":0,"findings":[],"status":"OK","verb":"validate"}

$ protodoc validate corrupted-file.pdl
{"checks":1,"exit_code":1,"findings":[{"rule_id":"TR-006","message":"bounded prefix truncated: read 2000 of 1048576 required octets"}],"status":"INVALID","verb":"validate"}

$ protodoc validate does-not-exist.pdl
{"checks":1,"exit_code":7,"findings":[{"rule_id":"TR-012-UNREADABLE","message":"..."}],"status":"USAGE","verb":"validate"}
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
{"added":null,"change_count":0,"changed":null,"changed_constructs":null,"exit_code":0,"findings":[],"identical":true,"removed":null,"status":"OK","verb":"diff"}
```

For two documents that actually differ, `added`/`removed`/`changed` are populated with per-ordinal
descriptions (e.g. `"changed@ordinal-2 (aa->bb)"`) and `identical` is `false`.

`identical`/`added`/`removed`/`changed` are the documented contract fields; `changed_constructs`/
`change_count` are kept alongside for backward compatibility.

### `merge <base> <a> <b> --out <path>`

Reconciles two divergent edits of a common base document.

```bash
$ protodoc merge base.pdl mine.pdl theirs.pdl --out merged.pdl
```

Classifies the merge (clean / `CONFLICT` / `REFUSED` for a real policy violation like a history-mode
mismatch) from real decoded content. See "Honest limitations" below — it does not yet write a fully
reconstructed merged document for the clean case.

### `project <file> --to <path> [--format=text|html]`

Produces a one-directional, non-normative text or HTML projection — for viewing, never for
re-ingestion (`reingestable` is always `false` in the response). `--to` is required.

```bash
$ protodoc project my-document.pdl --to preview.html --format=html
{"exit_code":0,"findings":[],"format":"html","out":"preview.html","reingestable":false,"status":"OK","verb":"project"}
```

### `redact <file> --subtree <unit-id> [--subtree <unit-id>...] --out <path>`

Removes a designated redactable subtree and republishes, leaving the redaction cryptographically
provable (not just visually blacked out). At least one `--subtree` and `--out` are both required.

```bash
$ protodoc redact my-document.pdl --subtree 9f2a...c1 --out redacted.pdl
```

### `publish <file> --out <path> [--partial]`

Full compaction to a fresh, canonicalized file. Refused if any signature is present, unless
`--partial` is given (reclaims only segments outside every signature's coverage). `--out` is
required.

```bash
$ protodoc publish my-document.pdl --out compacted.pdl
```

### `sign <file> --key <ref> --coverage total|subset [--subset-range <start>:<end> ...] --intent <value> --out <path>`

Computes a new EdDSA-Protodoc-1 signature. `--key` names a key held by your platform's key store or
HSM — the tool never accepts a raw private key on the command line. `--intent` must be one of
`author-approval`, `witness-attestation`, `notarization`, `custodial-transfer`. All four flags plus
`--out` are required.

```bash
$ protodoc sign my-document.pdl --key my-signing-key --coverage total --intent author-approval --out signed.pdl
```

See "Honest limitations" below — `--out` does not yet contain the new signature embedded in a
re-serialized document.

### `migrate <file> --to-major <N> --out <path> [--rescind-and-resign --new-key <ref> --new-param-set <id>]`

Migrates a document to a newer major format version. Refusal-first: checks every construct is
representable before writing anything, and refuses (`REFUSED`) if `--to-major` doesn't name a version
strictly greater than the file's current one. `--to-major` and `--out` are both required.

```bash
$ protodoc migrate old-document.pdl --to-major 2 --out migrated.pdl
```

## Honest limitations (as of 2026-09-19)

Two narrower gaps remain, deliberately not papered over — writing fabricated output would be worse
than leaving them unimplemented, per this project's honesty rule (`specs/CHANGES.md`'s
`DEFECT-2026-09-19b` resolution notes):

- **`sign --out` does not yet contain an embedded signature.** The signature itself is genuinely,
  correctly, deterministically computed (`signature_len`, `coverage_total`, etc. in the response are
  real) — but splicing a new `SIGNATURE` segment into a re-serialized document needs real
  segment-table/commit-ring reconstruction that doesn't exist yet. `--out` currently receives an
  unmodified copy of the input, and the response says so explicitly
  (`"signed_document_complete": false` plus a finding). Don't treat `--out` as a signed document yet;
  use the returned signature metadata directly if you need it.
- **`merge`'s clean-merge case does not yet write a fully reconstructed merged document.** Conflict
  detection and refusal conditions (history-mode mismatch, etc.) are real, derived from actual decoded
  content — but assembling genuine merged output needs the full `pkg/merge` R1/R2/R3 classification
  wired in, which is separately scoped future work.

Everything else — `validate`, `inspect`, `extract`, `verify`, `diff`, `project`, `redact`, `publish`,
`migrate` (forward migrations without `--rescind-and-resign`) — reads, decodes, validates, and writes
real files correctly, independently verified end-to-end.
