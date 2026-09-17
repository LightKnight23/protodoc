# Extraction benchmark corpus (T-0097)

This is shared test infrastructure for the extraction NFR benchmarks
(NFR-012 octets-read, NFR-013 processor-time, NFR-014 peak-memory: tasks
T-0098..T-0100). It is not itself a functional or non-functional requirement.

## What it generates

`extract.GenerateCorpus(profile)` produces a valid PDL document image with
many CONTENT text segments (varied sizes, a cycling 3-language marker) and
interspersed RESOURCE segments. The benchmark profile
`extract.GiBCorpusProfile(seed)` targets roughly 1 GiB across 10,000 pages;
`extract.SmallCorpusProfile(seed)` is a fast profile for the determinism
test. The generator draws every varying value from a fixed-seed splitmix64
PRNG and writes only via the container encoders, so its output is
byte-identical for a given profile across runs and machines
(`TestFixtureGen_1GiB10000PageCorpusIsDeterministic`).

## NFR-011 reference-measurement-machine dependency (FLAGGED, pending ruling)

NFR-011 requires benchmark results to be reported against a named reference
measurement configuration. The frozen spec/plan/data-model/contracts name no
concrete machine, even though roughly a dozen NFRs depend on one. Until an
Eyvar/themis ruling on NFR-011 lands, the extraction benchmarks use the
documented stand-in already checked into this repository:

    benchconfig.ReferenceConfigID = "PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX"

(pkg/benchconfig). This corpus generator does not itself pick a machine; it
defers to that stand-in. When the NFR-011 ruling lands, the reference
configuration should be updated in `pkg/benchconfig` in one place, and this
note removed.

This is a provisional stand-in recorded for review, NOT a claim that the
NFR-011 reference machine has been decided.
