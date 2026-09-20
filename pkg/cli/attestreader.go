// ATTEST-segment record reader (T-0392). SIGNATURE (0x40) and
// ATTESTATION_EVIDENCE (0x42) records use PDL-TLV's discriminant as an actual
// tag=0 FIELD inside the record (pdlfmt.DecodeRecord(src, known, -1) decodes it
// directly) — DISTINCT from content-model frames, which carry a raw leading
// discriminant byte with the TLV body starting after it. This reader gets that
// distinction right; it never treats an ATTEST record's first octet as a bare
// discriminant to be skipped. Go stdlib only.
package cli

import (
	"fmt"
	"io"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/ledger"
	"Protodoc/pkg/pdlfmt"
)

// attestRecordDiscriminantTag is field tag 0 across ATTEST record kinds; the
// two discriminants this reader dispatches on.
const (
	attestDiscTag      = 0
	discSignature      = 0x40
	discAttestEvidence = 0x42
)

// attestSegmentBody strips the leading ledger segment header (if present) from
// an ATTEST segment's raw octets, returning the record body. A segment whose
// octets begin with the "PDS1" segment magic carries a SegmentHeaderSize header
// before its record; one that does not is treated as a bare record body (the
// prefix-only test fixtures write bare bodies).
func attestSegmentBody(raw []byte) []byte {
	if len(raw) >= 4 && raw[0] == 'P' && raw[1] == 'D' && raw[2] == 'S' && raw[3] == '1' {
		if len(raw) >= ledger.SegmentHeaderSize {
			return raw[ledger.SegmentHeaderSize:]
		}
	}
	return raw
}

// DiscoveredEvidence indexes the ATTESTATION_EVIDENCE records found in a
// document's ATTEST segments, by kind, to their ae-id (a minted unit-id).
type DiscoveredEvidence struct {
	CredChain     pdlfmt.UnitID
	HasCredChain  bool
	TimeAttest    pdlfmt.UnitID
	HasTimeAttest bool
	Revocation    pdlfmt.UnitID
	HasRevocation bool
}

// readAttestRecord decodes one ATTEST record body, returning its discriminant
// (from tag=0) so the caller can dispatch. reservedFrom = 0 accepts every tag
// as well-formed (the record's own kind-specific decoder validates the full
// schema afterward); this reader only needs the tag=0 discriminant, which in
// ATTEST framing is a real decoded field, not a skipped leading byte.
func readAttestRecordDiscriminant(body []byte) (byte, error) {
	fields, err := pdlfmt.DecodeRecord(body, nil, 0)
	if err != nil {
		return 0, err
	}
	for _, f := range fields {
		if f.Tag == attestDiscTag {
			if len(f.Value) != 1 {
				return 0, fmt.Errorf("cli: ATTEST record tag=0 discriminant is %d octets, want 1", len(f.Value))
			}
			return f.Value[0], nil
		}
	}
	return 0, fmt.Errorf("cli: ATTEST record has no tag=0 discriminant field")
}

// DiscoverEvidence scans every ATTEST segment of the document at r for
// ATTESTATION_EVIDENCE records and indexes their ae-ids by kind. r's prefix
// carries the segment table; each ATTEST segment body is section-read and
// decoded with the ATTEST framing (discriminant = tag=0). A SIGNATURE record
// encountered is skipped (only evidence is indexed here).
func DiscoverEvidence(r io.ReaderAt, table [container.MaxSegments]container.SegmentTableSlot) (DiscoveredEvidence, error) {
	var found DiscoveredEvidence
	for _, slot := range table {
		if slot.SegmentType != container.SegmentTypeAttest {
			continue
		}
		raw := make([]byte, slot.Length)
		if _, err := r.ReadAt(raw, int64(slot.Offset)); err != nil {
			return found, fmt.Errorf("cli: reading ATTEST segment at %d: %w", slot.Offset, err)
		}
		body := attestSegmentBody(raw)
		disc, err := readAttestRecordDiscriminant(body)
		if err != nil {
			return found, err
		}
		if disc != discAttestEvidence {
			continue // SIGNATURE (0x40) or other: not evidence
		}
		ev, err := integrity.DecodeAttestationEvidence(body)
		if err != nil {
			return found, fmt.Errorf("cli: decoding ATTESTATION_EVIDENCE: %w", err)
		}
		switch ev.Kind {
		case integrity.AeCredentialChain:
			found.CredChain, found.HasCredChain = ev.ID, true
		case integrity.AeTimeAttestation:
			found.TimeAttest, found.HasTimeAttest = ev.ID, true
		case integrity.AeRevocationEvidence:
			found.Revocation, found.HasRevocation = ev.ID, true
		}
	}
	return found, nil
}
