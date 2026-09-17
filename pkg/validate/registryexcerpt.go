// RegistryExcerpt wire struct and validation step-12 completeness check
// (T-0370, FR-011; data-model.md 2.19 + S7 step 12; document.abnf S7.3).
// A RegistryExcerpt is a RESOURCE-typed record (discriminant 0x09) holding
// self-hosted external-registry entries so a durable-profile document's own
// IDs stay meaningful without the central registry (DP-017). For a document
// declaring durable_claim = 1, an entry MUST exist for every external id the
// document actually references (its unicode_version_id, its shaping_profile_id,
// and every extension token appearing in any extension envelope); absence is a
// structural validation failure at step 12, distinct from every earlier
// verdict category, naming the missing id. A durable_claim = 0 document MAY
// carry the record but is not required to.
package validate

import (
	"encoding/binary"
	"errors"
	"fmt"

	"Protodoc/pkg/pdlfmt"
)

// RegistryExcerpt discriminant (document.abnf S7.3) and TLV tags.
const (
	registryExcerptDiscriminant = 0x09
	reTagDiscriminant           = 0 // re-discriminant, field-value = 0x09
	reTagEntries                = 1 // re-entries, field-value = plain-seq-of(registry-entry)
)

// ExternalIDKind is a registry-entry's re-kind (document.abnf S7.3): the class
// of external id an entry excerpts.
type ExternalIDKind uint8

const (
	KindExtensionToken   ExternalIDKind = 0x00
	KindShapingProfileID ExternalIDKind = 0x01
	KindUnicodeVersionID ExternalIDKind = 0x02
)

func (k ExternalIDKind) String() string {
	switch k {
	case KindExtensionToken:
		return "extension-token"
	case KindShapingProfileID:
		return "shaping_profile_id"
	case KindUnicodeVersionID:
		return "unicode_version_id"
	default:
		return fmt.Sprintf("reserved(0x%02x)", uint8(k))
	}
}

// ExternalID identifies one externally-registered id: its kind and value.
type ExternalID struct {
	Kind  ExternalIDKind
	Value uint16
}

// RegistryEntry is one registry-entry (document.abnf S7.3): re-kind re-value
// re-spec-pointer re-snapshot-digest.
type RegistryEntry struct {
	Kind           ExternalIDKind
	Value          uint16
	SpecPointer    string   // re-spec-pointer, nfc-string (authored prose, never dereferenced)
	SnapshotDigest [32]byte // re-snapshot-digest, digest256
}

// RegistryExcerpt is the decoded RESOURCE-typed record (discriminant 0x09).
type RegistryExcerpt struct {
	Entries []RegistryEntry
}

// ErrRegistryExcerptDiscriminant is returned when a decoded record's
// discriminant is not 0x09.
var ErrRegistryExcerptDiscriminant = errors.New("validate: record discriminant is not REGISTRY_EXCERPT (0x09)")

// encodeEntry encodes one registry-entry: re-kind(u8) re-value(u16 BE)
// re-spec-pointer(nfc-string) re-snapshot-digest(digest256), concatenated in
// field order (an element of the plain sequence, not itself a TLV record).
func encodeEntry(e RegistryEntry) ([]byte, error) {
	var buf []byte
	buf = append(buf, byte(e.Kind))
	var v [2]byte
	binary.BigEndian.PutUint16(v[:], e.Value)
	buf = append(buf, v[:]...)
	var err error
	buf, err = pdlfmt.AppendNFCString(buf, e.SpecPointer)
	if err != nil {
		return nil, fmt.Errorf("validate: registry entry spec pointer: %w", err)
	}
	buf = pdlfmt.AppendDigest256(buf, e.SnapshotDigest)
	return buf, nil
}

// decodeEntry decodes one registry-entry from the start of src, returning the
// entry and the octets consumed.
func decodeEntry(src []byte) (RegistryEntry, int, error) {
	var e RegistryEntry
	pos := 0
	if len(src) < 1+2 {
		return e, 0, fmt.Errorf("validate: registry entry truncated at kind/value")
	}
	e.Kind = ExternalIDKind(src[pos])
	pos++
	e.Value = binary.BigEndian.Uint16(src[pos : pos+2])
	pos += 2
	s, n, err := pdlfmt.DecodeNFCString(src[pos:])
	if err != nil {
		return e, 0, fmt.Errorf("validate: registry entry spec pointer: %w", err)
	}
	e.SpecPointer = s
	pos += n
	d, n, err := pdlfmt.DecodeDigest256(src[pos:])
	if err != nil {
		return e, 0, fmt.Errorf("validate: registry entry snapshot digest: %w", err)
	}
	e.SnapshotDigest = d
	pos += n
	return e, pos, nil
}

// Encode encodes the RegistryExcerpt to its byte-exact PDL-TLV wire form: a
// record of re-discriminant (tag 0, value 0x09) and re-entries (tag 1, value =
// plain-seq-of(registry-entry)).
func (r RegistryExcerpt) Encode() ([]byte, error) {
	elems := make([][]byte, len(r.Entries))
	for i, e := range r.Entries {
		b, err := encodeEntry(e)
		if err != nil {
			return nil, err
		}
		elems[i] = b
	}
	entriesValue := pdlfmt.AppendPlainSeq(nil, elems)
	fields := []pdlfmt.Field{
		{Tag: reTagDiscriminant, Value: []byte{registryExcerptDiscriminant}},
		{Tag: reTagEntries, Value: entriesValue},
	}
	return pdlfmt.EncodeRecord(fields)
}

// DecodeRegistryExcerpt decodes a RegistryExcerpt from its wire form, byte-
// exact inverse of Encode. It rejects a wrong discriminant and any malformed
// entry.
func DecodeRegistryExcerpt(src []byte) (RegistryExcerpt, error) {
	var re RegistryExcerpt
	known := map[byte]bool{reTagDiscriminant: true, reTagEntries: true}
	fields, err := pdlfmt.DecodeRecord(src, known, -1)
	if err != nil {
		return re, err
	}
	var sawDisc, sawEntries bool
	for _, f := range fields {
		switch f.Tag {
		case reTagDiscriminant:
			if len(f.Value) != 1 || f.Value[0] != registryExcerptDiscriminant {
				return re, ErrRegistryExcerptDiscriminant
			}
			sawDisc = true
		case reTagEntries:
			count, n, err := pdlfmt.DecodeSeqCount(f.Value)
			if err != nil {
				return re, fmt.Errorf("validate: registry entries count: %w", err)
			}
			pos := n
			for i := uint64(0); i < count; i++ {
				e, consumed, err := decodeEntry(f.Value[pos:])
				if err != nil {
					return re, err
				}
				re.Entries = append(re.Entries, e)
				pos += consumed
			}
			if pos != len(f.Value) {
				return re, fmt.Errorf("validate: registry entries have %d trailing octets", len(f.Value)-pos)
			}
			sawEntries = true
		}
	}
	if !sawDisc {
		return re, ErrRegistryExcerptDiscriminant
	}
	_ = sawEntries // entries may legitimately be empty (count 0)
	return re, nil
}

// RegistryExcerptCompletenessCheckKey is the cli.md validate stdout check key
// this step-12 check reports under.
const RegistryExcerptCompletenessCheckKey = "registry_excerpt_completeness"

// RegistryCompletenessResult is the outcome of the step-12 completeness check.
type RegistryCompletenessResult struct {
	Passed  bool
	Missing []ExternalID // referenced external ids with no RegistryExcerpt entry
}

// CheckRegistryExcerptCompleteness implements data-model.md S7 step 12 / FR-011.
// For a document with durableClaim true, it confirms a RegistryExcerpt entry
// exists for every referenced external id; a missing entry fails the check and
// names the missing id. For durableClaim false the check always passes (the
// excerpt is optional). referenced is the set of external ids the document
// actually references (its unicode_version_id, shaping_profile_id, and every
// extension token in any extension envelope).
func CheckRegistryExcerptCompleteness(durableClaim bool, excerpt RegistryExcerpt, referenced []ExternalID) RegistryCompletenessResult {
	if !durableClaim {
		return RegistryCompletenessResult{Passed: true}
	}
	have := map[ExternalID]bool{}
	for _, e := range excerpt.Entries {
		have[ExternalID{Kind: e.Kind, Value: e.Value}] = true
	}
	var missing []ExternalID
	for _, id := range referenced {
		if !have[id] {
			missing = append(missing, id)
		}
	}
	return RegistryCompletenessResult{Passed: len(missing) == 0, Missing: missing}
}

// RegistryCompletenessFinding maps a failing completeness check to a step-12
// pipeline Finding, reporting under the registry_excerpt_completeness check
// key and naming the first missing id. Returns nil when the check passed.
func RegistryCompletenessFinding(res RegistryCompletenessResult) *Finding {
	if res.Passed {
		return nil
	}
	msg := RegistryExcerptCompletenessCheckKey + ": durable-profile document references external id(s) with no RegistryExcerpt entry"
	if len(res.Missing) > 0 {
		msg += fmt.Sprintf(" (first missing: %s value %d)", res.Missing[0].Kind, res.Missing[0].Value)
	}
	return &Finding{
		Step:    StepRegistryExcerpt,
		RuleID:  RegistryExcerptCompletenessCheckKey,
		Message: msg,
	}
}
