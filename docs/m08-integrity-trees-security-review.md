# M08 integrity-trees milestone-exit argus review (T-0362)

Consolidated security review of the whole M08 surface (T_S/T_C integrity
trees, ATTEST exclusion, CoverageDescriptor, structure_digest) before any
downstream milestone (M09 signature & coverage, M11 redaction, publish/sign/
migrate) depends on it. Reviewed at the commit that introduces this file.

## Findings summary

Zero unresolved P0 or P1 findings. Any P0/P1 discovered during review would
be filed as a new task, never fixed silently inside this artifact; none were.

## Per-task review (T-0129 .. T-0144)

- **T-0129 domain-tag registry + ABSENT_CHILD_DIGEST.** PASS. Single shared
  registry; ABSENT_CHILD_DIGEST = SHA-256(0x00) pinned; every real preimage
  starts with a distinct nonzero tag, so the filler is collision-free with
  any genuine digest (domain separation, the "T_C leaf-tag reuse" bug class
  closed).
- **T-0130 T_S leaf/internal + arity-16/depth-4 builder.** PASS. Fixed
  513-octet internal preimage; full 65536-leaf base; ABSENT_CHILD_DIGEST at
  every level; over-capacity is a named error.
- **T-0131 TSRoot never-cached.** PASS. No cached-root parameter;
  CompareLedgerRoot uses a stored root only for comparison and reports
  divergence as a distinct verdict (no-trust-stored-digest control).
- **T-0132/T-0133 T_C leaves (0x02/0x07).** PASS. Stored frame bytes used
  verbatim (no re-encode); redactable leaf salted; domain separation between
  the two proven under a salt sweep.
- **T-0134 T_C traversal order.** PASS with FLAG carried forward: the
  interim unit-id byte-lexicographic order is deterministic and satisfies
  the negative constraints; the aspirational reading-order requirement is a
  self-disclosed, documented gap (future ROOT_SEQUENCE work), not silently
  "fixed."
- **T-0135/T-0136 T_C internal (0x08) + TCRoot.** PASS. Fixed 513-octet
  internal preimage; minimal depth <= 5; ABSENT fill at every level; TCRoot
  depends on no SegmentTableSlot or T_S value (least-coupling of the signed
  root).
- **T-0137 commit-ring state-id wiring.** PASS. state-id-field and t-c-root
  set from the freshly recomputed T_C_root, never a cached value.
- **T-0138 ATTEST exclusion.** PASS. One shared IsAttestTyped predicate;
  the covered set / no-op comparison / CoverageDescriptor all route through
  it (single-definition discipline).
- **T-0139 CoverageDescriptor + PD-COVER-001..004.** PASS. The single
  canonical wire implementation; PD-COVER-004 closes the FR-002
  self-coverage circularity (no ATTEST ordinal nameable).
- **T-0140 structure_digest 8-item preimage.** PASS. Exact fixed order;
  items 1/4/6 always; conditional items driven by the bitmask; segment
  digests and T_S root recomputed fresh by the caller (no-trust-stored);
  reserved bit rejected before assembly.
- **T-0141 tamper conformance.** PASS. A tampered slot is detected by
  comparison of freshly recomputed structure_digest, never by trusting the
  slot's own digest.
- **T-0142 determinism conformance.** PASS. Both roots pinned and
  reproduced exactly from raw inputs.
- **T-0143 sensitivity conformance.** PASS. Every covered-value mutation
  changes the root; an identical-copy control does not.
- **T-0144 capacity conformance.** PASS. At-limit computes; one past the
  builder capacity is a named error.

## Controls checked (OWASP ASVS / ISO 27034 touchpoints)

- **Domain separation** (V6/crypto): every preimage class is domain-tagged;
  ABSENT_CHILD_DIGEST is uncollidable with any real digest. PASS.
- **No-trust-stored-digest**: T_S root, T_C root, segment summaries and the
  ledger-root comparison all recompute fresh; stored values are compared,
  never trusted. PASS.
- **Least-exposure of redactable salts**: salts live inside the redactable
  subtree they commit and are destroyed with it on redaction; they are not
  logged or surfaced by the tree code. PASS (the destruction step itself is
  M11's redaction task; T_C only commits to them).

## Downstream gate

M09/M11 and the publish/sign/migrate paths MAY depend on pkg/integrity as of
this review. Zero P0/P1 open.
