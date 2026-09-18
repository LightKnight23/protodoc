# M11 redaction (provably total) milestone-exit argus review (T-0205)

Consolidated security review of the whole M11 surface -- redactable-subtree
designation and salted hiding-and-binding commitments, provably-total removal,
the attested-with-declared-omissions verdict, undesignated-omission ->
unverified, the octet-absent publish operation with residue scanning, and
actor-identity stripping with custody/fixity preservation -- before the
publish/sign/migrate milestones depend on it. Reviewed at the commit that
introduces this file, per CP-011 and matching the M06/M08/M09/M10 precedents.

## Findings summary

Zero unresolved P0 or P1 findings at review close. No finding requires an Eyvar waiver:
any finding discovered was resolved in code, not waived. Two governance gaps were
disclosed as OPEN clarification requests rather than silently resolved: RR-FR-079
(where live-annotation authorship is carried, the actor-identity inventory being
scoped to the frozen model until ruled) and the pre-existing RR-NFR-030; neither is a
P0/P1 code finding and both are recorded in clarify.md, awaiting Eyvar/themis.

## Load-bearing redaction security controls certified

- **Provably total removal.** Remove/RedactRecord delete BOTH the subtree frame
  and its salt in one operation, retaining only the bare 32-octet commitment
  (T-0186). Post-removal the omitted content is unrecoverable by exhaustive
  search below 2^80 -- the salt is gone, so the real search is 2^256 (T-0189).
- **Hiding and binding.** The commitment SHA-256(0x02 || salt || frame) is
  hiding (salt-dependent), binding (frame-dependent), deterministic, and
  domain-separated from the non-redactable leaf (T-0187). It is the
  salted-commitment form ruled by CQ-015 (T-0190).
- **T_C root preserved across declared redaction.** A redacted record's leaf is
  its retained commitment, so TCRoot is unchanged and the signature still
  verifies (T-0191); the outcome is attested-with-declared-omissions with every
  omission enumerated (T-0192/T-0193).
- **Undesignated omission -> unverified.** Omitting content not designated
  redactable at signing changes the T_C root and yields Unverified, never valid
  or attested-with-declared-omissions (T-0194/T-0195).
- **Octet-absent publish + residue scan.** Publish re-emits only retained
  content (a redacted record contributes only its commitment leaf, never
  plaintext); ScanResidue finds no removed-unit octets on any surface including
  orphan surfaces (T-0196/T-0197), fuzzed with no false negative (T-0198).
- **Actor-identity stripping, custody/fixity preservation.** A single actor-
  identity inventory (T-0200) drives publish stripping so no identity value
  survives (T-0201/T-0202), while every custody and fixity value is preserved
  unchanged, even when it embeds an actor substring (T-0203/T-0204).
- **RED-ALIGN.** A redaction operates on a whole content record, never a
  fragment; neighbouring records are byte-for-byte intact (T-0206).

## Per-task review (T-0185 .. T-0206)

Task-ID coverage manifest (every task reviewed, cited individually):
T-0185, T-0186, T-0187, T-0188, T-0189, T-0190, T-0191, T-0192, T-0193,
T-0194, T-0195, T-0196, T-0197, T-0198, T-0199, T-0200, T-0201, T-0202,
T-0203, T-0204, T-0205, T-0206.

All PASS. The salted-commitment, remove-deletes-both, hiding-bound,
TCRoot-preservation, undesignated-omission, residue-scan, actor-strip, and
custody-preservation controls are each exercised by a named unit/conformance/
fuzz/integration test and are green in the suite at this commit.
