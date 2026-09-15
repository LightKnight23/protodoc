# NFR-011 Reference Measurement Configuration

Status: Draft, pending Eyvar's review and approval. Non-normative operational
document: it fixes no requirement text, it fills the operational parameter
NFR-011 already presumes exists. Until approved, no NFR-012..019 "satisfied"
verdict may be treated as final on the strength of this document alone.

## Purpose

NFR-011 requires:

> The Protodoc specification SHALL define a reference measurement
> configuration, naming the processor model, core count, memory, storage
> class and operating system, against which every stated time and memory
> bound is measured.

and its Verify clause (Audit A-REFPLAT) requires every time/memory bound to
cite this configuration by identifier, and CI benchmarks to fail if the
identifier they ran on does not match. This document is that configuration.
It does not restate or alter NFR-012..019's own numbers; it names the
denominator they are measured against.

## Reference configuration

| Field | Value |
|---|---|
| Identifier | `PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX` |
| Processor model | Apple M2 Max |
| Core count | 12 (8 performance + 4 efficiency) |
| Memory | 68719476736 octets (64 GiB) unified memory |
| Storage class | Internal NVMe SSD, APFS |
| Operating system | macOS, Darwin kernel 27.0.0, arm64 |
| Go toolchain | go1.25.1 darwin/arm64 |

This is the host this repository's own benchmark suite (`go test -bench`) is
run on as of this writing. It was captured directly off the machine running
CI for this repo at the time this document was authored (`go version`,
`uname -a`, `sysctl hw.physicalcpu`/`hw.logicalcpu`/`hw.memsize`), not
estimated or copied from a vendor spec sheet.

## Scope and honesty note

This document names a real, currently-in-use machine configuration. It does
not assert that this is the configuration Protodoc's v1 release will
ultimately ship its official reference numbers against, and it does not
assert Eyvar has ruled on it as final — that decision belongs to Eyvar per
the SDD phase-gate rule (a frozen spec's operational parameters still need a
human sign-off before other verdicts may cite them as authoritative). Every
benchmark job citing `PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX` (see
`Protodoc/pkg/benchconfig`) is citing exactly this identifier, so a future
change to the identifier or the fields above (a different CI host, a
revision after Eyvar's review) is a single point of edit that
`TestNFR_011_ReferenceConfigPublished` keeps in sync with the constant, and
`TestNFR_011_BenchmarksCiteReferenceConfig` keeps every benchmark job
citing.

## Benchmarks citing this configuration

| Requirement(s) | Benchmark | Status |
|---|---|---|
| NFR-015, NFR-016 | `BenchmarkNFR_015_ColdCachePreviewLatency` (`pkg/container/coldcache_bench_test.go`) | cites `PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX` via `benchconfig.Stamp` |
| NFR-012, NFR-013, NFR-014 | extraction benchmark (milestone M05) | not yet implemented; no extraction package exists in this repo yet |
| NFR-017, NFR-018 | page-render benchmark (milestone M14) | not yet implemented; no render package exists in this repo yet |
| NFR-019 | rasterizer determinism check (milestone M14) | not yet implemented; NFR-019 is a determinism requirement, not itself a time/memory bound, so whether it needs a reference-config citation under A-REFPLAT is for M14's implementer to confirm against the reference-config's own scope when that package is built |

`TestNFR_011_BenchmarksCiteReferenceConfig` (`pkg/benchconfig`) enforces this
table's claim mechanically: it fails the build if any `func Benchmark...`
anywhere in this module does not call `benchconfig.Stamp`, so a future
NFR-012/013/014/017/018 benchmark job cannot be added without also citing
this configuration.
