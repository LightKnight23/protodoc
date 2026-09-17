// Deterministic benchmark corpus generator (T-0097): produces a valid,
// byte-identical-per-seed PDL document with many CONTENT text segments
// (varied sizes, mixed languages) plus interspersed RESOURCE segments, for
// the NFR-012/013/014 extraction benchmarks (T-0098..T-0100). It is shared
// test infrastructure, not a functional requirement.
//
// NFR-011 REFERENCE-CONFIG DEPENDENCY (flagged): the frozen spec/plan/
// data-model/contracts name no concrete reference measurement machine for
// the ~dozen NFRs that depend on one. Until an Eyvar/themis ruling on
// NFR-011 lands, the benchmarks use the documented stand-in already checked
// into this repo, benchconfig.ReferenceConfigID
// ("PDL-REFCFG-2026-09-DARWIN-ARM64-M2MAX"); see docs for the corpus. This
// generator does not itself pick a machine; it defers to that stand-in and
// records the dependency here so the ruling can replace it in one place.
package extract

import (
	"Protodoc/pkg/container"
)

// CorpusProfile parameterises a generated corpus. The 1 GiB / 10,000-page
// benchmark profile is CorpusProfile{Pages: 10000, TargetBytes: 1<<30}; the
// generator scales deterministically to any profile so the determinism test
// can run at a small size while the benchmarks run at the full one.
type CorpusProfile struct {
	Seed          int64
	Pages         int // number of CONTENT text segments ("pages")
	AvgSegBytes   int // average CONTENT segment size; sizes vary deterministically around it
	ResourceEvery int // insert a RESOURCE segment every N content segments (0 = none)
}

// SmallCorpusProfile is a fast profile for the determinism unit test.
func SmallCorpusProfile(seed int64) CorpusProfile {
	return CorpusProfile{Seed: seed, Pages: 64, AvgSegBytes: 512, ResourceEvery: 8}
}

// GiBCorpusProfile is the NFR-012/013/014 benchmark profile: ~1 GiB across
// 10,000 pages. It is documented here; the benchmarks that consume it live
// in T-0098..T-0100.
func GiBCorpusProfile(seed int64) CorpusProfile {
	return CorpusProfile{Seed: seed, Pages: 10000, AvgSegBytes: 100 * 1024, ResourceEvery: 20}
}

// GenerateCorpus produces a valid PDL document image for the profile,
// byte-identical for a given profile (same seed and parameters). It lays out
// a fixed prefix followed by CONTENT and RESOURCE segments; CONTENT payloads
// carry deterministic pseudo-random text bytes and a mixed language marker
// so extraction has realistic, varied work. It uses only the container
// encoders (no ambient input), so two runs on the same profile yield
// identical octets.
func GenerateCorpus(p CorpusProfile) ([]byte, error) {
	rng := newCorpusRand(uint64(p.Seed) ^ 0xD1B54A32D192ED03)

	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord

	// First pass: decide segment sizes/types deterministically and lay out
	// slots at increasing offsets.
	type segPlan struct {
		typ    byte
		length uint64
	}
	var plans []segPlan
	contentCount := 0
	for page := 0; page < p.Pages; page++ {
		// Vary size deterministically in [0.5x, 1.5x] of the average.
		delta := int(rng.next()%uint64(p.AvgSegBytes)) - p.AvgSegBytes/2
		length := uint64(p.AvgSegBytes + delta)
		if length < 16 {
			length = 16
		}
		plans = append(plans, segPlan{typ: container.SegmentTypeContent, length: length})
		contentCount++
		if p.ResourceEvery > 0 && contentCount%p.ResourceEvery == 0 {
			plans = append(plans, segPlan{typ: container.SegmentTypeResource, length: uint64(256 + int(rng.next()%256))})
		}
	}

	base := uint64(container.SegmentTableOffset + container.SegmentTableRegionSize)
	offset := base
	slots := make([]container.SegmentTableSlot, len(plans))
	for i, pl := range plans {
		slots[i] = container.SegmentTableSlot{SegmentType: pl.typ, Offset: offset, Length: pl.length, FrameCount: 1}
		offset += pl.length
	}
	totalLen := offset

	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: totalLen, SegmentCount: uint16(min(len(slots), 0xFFFF))}
	}

	img := make([]byte, 0, totalLen)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, err := (&container.Frontmatter{Title: "Benchmark corpus", PageCount: uint32(p.Pages), Language: "en-US"}).Encode(nil)
	if err != nil {
		return nil, err
	}
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		return nil, err
	}
	img = append(img, stEnc...)

	// Segment payloads: deterministic bytes, with a language marker in the
	// first octet of each CONTENT segment (cycling 3 languages).
	langMarkers := []byte{1, 2, 3}
	li := 0
	for _, pl := range plans {
		payload := make([]byte, pl.length)
		if pl.typ == container.SegmentTypeContent {
			payload[0] = langMarkers[li%len(langMarkers)]
			li++
		}
		for j := 1; j < len(payload); j++ {
			payload[j] = byte(rng.next())
		}
		img = append(img, payload...)
	}
	return img, nil
}

// corpusRand is a deterministic splitmix64 PRNG (no ambient state) so the
// generated corpus is byte-identical for a given seed across runs/machines.
type corpusRand struct{ state uint64 }

func newCorpusRand(seed uint64) *corpusRand { return &corpusRand{state: seed} }

func (r *corpusRand) next() uint64 {
	r.state += 0x9E3779B97F4A7C15
	z := r.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}
