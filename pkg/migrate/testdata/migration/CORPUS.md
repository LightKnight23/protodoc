# Migration golden corpus (FR-119 / CP-003 / NFR-028)

This directory is the migrate-specific golden fixture set: `(source document, target major version)`
pairs paired with their **expected canonical migrated octets**, for consumption by a second,
independently-authored implementation.

## Format

`manifest.tsv` — one row per pair, tab-separated:

```
<pair-name>    <target-major>    <expected-canonical-octets-hex>
```

The expected octets are the canonical output of `migrate.Transform(source, target)` under the
traversal-ordering contract (segment ordinal, then intra-segment offset; no map iteration, no
wall-clock, no goroutine order — FR-119). The per-pair source construct lists are defined in
`corpus_test.go` (the same file that regenerates and checks these bytes), so an independent
implementation can reconstruct each source input from that definition.

## Reuse by an independent implementation

An independently-authored Protodoc implementation MAY reuse this corpus by:

1. constructing each pair's source document from the construct list in `corpus_test.go`,
2. running its own `Transform` to the stated target major version,
3. asserting its output equals the hex octets in `manifest.tsv`.

## Scope and honest status

This corpus **feeds** M19's broader two-implementation gate; it does **not** itself run that gate.

- The golden octets here are the **author's single (reference) implementation's** output. They are
  correct-by-construction against this implementation, and are frozen so drift is caught.
- **No second, independently-authored implementation exists yet, and no cross-implementation trial
  has been run.** The two-implementation byte-identity gate (CP-003 / NFR-028) is M19 scope and is
  **OPEN**. This file does not claim, and must not be read as claiming, that an independent
  implementation has reproduced these bytes. Publishing the corpus is the deliverable here; running
  the trial is a later milestone requiring a genuinely separate implementer.
