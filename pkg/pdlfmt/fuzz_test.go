// CP-012 continuous fuzzing (T-0359): PDL-VARINT and PDL-TLV are two of the
// six untrusted-byte decode entry points the constitution's continuous-
// fuzzing cross-cutting concern names M01 as owning. Every fuzz target
// here asserts exactly one property regardless of input: the decoder
// returns an ordinary error on malformed input, it never panics — decode
// is a trust boundary (CP-006), and a panic there is itself the defect
// CP-012 exists to catch before it reaches a caller.
package pdlfmt

import (
	"math"
	"testing"
)

// FuzzCP_012_DecodeUntrustedBytes_Varint seeds from
// TestNFR_001_VarintMinimalEncodingRoundTrip's boundary values (encoded,
// so the corpus starts from well-formed input) and from
// TestPD_VARINT_001_RejectsNonMinimalEncoding / RejectsTruncatedInput's
// named rejection vectors (already malformed, so the corpus also starts
// from the two documented rejection classes).
func FuzzCP_012_DecodeUntrustedBytes_Varint(f *testing.F) {
	for _, v := range []uint64{
		0, 1,
		varintU8Max, varintU8Max + 1,
		varintU16Max, varintU16Max + 1,
		varintU32Max, varintU32Max + 1,
		math.MaxUint64,
	} {
		f.Add(EncodeVarint(v))
	}
	for _, src := range [][]byte{
		{0xFD, 0x00, 0x05},             // non-minimal u16 form
		{0xFE, 0x00, 0x00, 0x00, 0x05}, // non-minimal u32 form
		{0xFF, 0, 0, 0, 0, 0, 0, 0, 5}, // non-minimal u64 form
		{},                             // empty
		{0xFD},                         // truncated u16 form, no octets
		{0xFD, 0x01},                   // truncated u16 form, one octet short
		{0xFE, 0x00, 0x01, 0x00},       // truncated u32 form
		{0xFF, 0, 0, 0, 1, 0, 0, 0},    // truncated u64 form
	} {
		f.Add(src)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		v, consumed, err := DecodeVarint(data)
		if err == nil {
			// A successful decode must never claim to have consumed more
			// octets than it was given, and must consume at least one.
			if consumed < 1 || consumed > len(data) {
				t.Fatalf("DecodeVarint(%x) = (%d, %d, nil): consumed out of [1,%d] bounds", data, v, consumed, len(data))
			}
		}
	})
}

// FuzzCP_012_DecodeUntrustedBytes_TLV seeds from
// TestNFR_001_TLVClosedGrammarRoundTrip's encoded multi-field record and
// from the named PD-TLV-001 / FR-106 / reserved-tail rejection vectors in
// tlv_test.go. known and reservedFrom are themselves fuzzed inputs
// (testing.F cannot fuzz a map directly): knownMask's low 32 bits gate
// which of tags 0..31 the schema recognises, and reservedFromRaw==255 is
// the sentinel for "no reserved tail" (reservedFrom=-1), matching every
// existing call site's own -1/non-negative convention.
func FuzzCP_012_DecodeUntrustedBytes_TLV(f *testing.F) {
	var digest Digest256
	for i := range digest {
		digest[i] = byte(i)
	}
	nfcVal, err := AppendNFCString(nil, "héllo")
	if err != nil {
		f.Fatalf("AppendNFCString: %v", err)
	}
	plainSeqVal := AppendPlainSeq(nil, [][]byte{{0x01}, {0x03}, {0x02}})
	sortedVecVal, err := AppendSortedVec(nil, [][]byte{{0x01}, {0x02}, {0x03}})
	if err != nil {
		f.Fatalf("AppendSortedVec: %v", err)
	}
	validRecord, err := EncodeRecord([]Field{
		{Tag: 1, Value: EncodeVarint(300)},
		{Tag: 2, Value: AppendDigest256(nil, digest)},
		{Tag: 3, Value: nfcVal},
		{Tag: 4, Value: plainSeqVal},
		{Tag: 5, Value: sortedVecVal},
	})
	if err != nil {
		f.Fatalf("EncodeRecord: %v", err)
	}
	const maskTags12345 = uint32(0b111110)
	const maskTag1 = uint32(0b10)
	const noReservedTail = uint8(255)

	f.Add(validRecord, maskTags12345, noReservedTail)
	f.Add([]byte{2, 1, 0x00, 1, 1, 0x00}, maskTags12345, noReservedTail)                             // PD-TLV-001: descending
	f.Add([]byte{1, 1, 0x00, 1, 1, 0x00}, maskTag1, noReservedTail)                                  // PD-TLV-001: repeated
	f.Add([]byte{9, 1, 0x00}, maskTag1, uint8(8))                                                    // reserved tail covers tag 9
	f.Add([]byte{9, 1, 0x00}, maskTag1, noReservedTail)                                              // unknown tag, no reserved tail
	f.Add([]byte{1, 5, 0x00, 0x00}, maskTag1, noReservedTail)                                        // FR-106: declared length exceeds input
	f.Add([]byte{1}, maskTag1, noReservedTail)                                                       // no length octet at all
	f.Add([]byte{1, 0xFD, 0x00}, maskTag1, noReservedTail)                                           // truncated varint length
	f.Add([]byte{1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, maskTag1, noReservedTail) // huge declared length
	f.Add([]byte{}, uint32(0), noReservedTail)

	f.Fuzz(func(t *testing.T, data []byte, knownMask uint32, reservedFromRaw uint8) {
		known := make(map[byte]bool, 32)
		for tag := 0; tag < 32; tag++ {
			if knownMask&(1<<uint(tag)) != 0 {
				known[byte(tag)] = true
			}
		}
		reservedFrom := int(reservedFromRaw)
		if reservedFromRaw == noReservedTail {
			reservedFrom = -1
		}
		_, _ = DecodeRecord(data, known, reservedFrom)
	})
}
