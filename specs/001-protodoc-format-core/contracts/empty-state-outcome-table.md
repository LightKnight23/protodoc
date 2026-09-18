# Empty-state (BOTTOM) outcome table (FR-124)

FR-124 requires exactly ONE defined outcome, per reader role, for every degenerate/empty-content case.
This table is the normative acceptance contract for the canonical empty-state fixture (the T-0315 BOTTOM
state: a valid, minimal file — fixed prefix plus zero content/resource/history segments).

M17 owns canonicalization and validation; it does not depend on M05 (extract) or M14 (render). This table
is the acceptance contract those milestones' own conformance vectors execute against the BOTTOM fixture
when they are built; publishing the table is M17's obligation, executing extract/render vectors is M05's
and M14's.

## Outcome table — one row per reader role

| Reader role | Required outcome against the BOTTOM fixture |
|---|---|
| extraction | Succeeds, emitting ZERO text units (empty extraction, not an error). |
| preview | Succeeds, producing an EMPTY preview raster (zero preview octets), not an error. |
| reflow | Succeeds, producing ZERO reflowed lines (empty flow), never an infeasible-layout error. |
| rasterisation | Succeeds, producing a blank page raster of the declared geometry (zero drawn glyphs), not an error. |
| validation | Verdict is a well-defined success: the minimal file is structurally valid (fixed prefix, zero segments); never Unverified/UnavailableState/Unattested on account of emptiness. |

## Schema

Each of the FIVE reader roles above MUST have exactly one row naming its exact expected output or verdict.
CI (`TestFR_124_EmptyStateOutcomeTableSchemaComplete`) fails if any role is missing an entry.
