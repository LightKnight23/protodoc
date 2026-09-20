package cli

import (
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/pdlfmt"
)

// writeDocWithFramesTCRootAndAttest writes a valid document with real CONTENT
// frames (ring winner carrying the given genuine T_C_root) followed by the
// given ATTEST segment bodies (already framed with a ledger segment header,
// e.g. via attestSegmentBytes), needed to test a document that is BOTH
// content-reconstructable AND carries real ATTEST records.
func writeDocWithFramesTCRootAndAttest(t *testing.T, ids []pdlfmt.UnitID, frames [][]byte, tcRoot [32]byte, attestSegs [][]byte) string {
	t.Helper()
	var segs []fixtureSeg
	for _, fr := range frames {
		segs = append(segs, fixtureSeg{segType: container.SegmentTypeContent, body: fr})
	}
	for _, seg := range attestSegs {
		segs = append(segs, fixtureSeg{segType: container.SegmentTypeAttest, body: seg})
	}
	return writeDoc(t, defaultHeader(), segs, func(r *container.CommitRingRecord) { r.TCRoot = tcRoot })
}

// writeSignableDocWithEvidence writes a valid document carrying real CONTENT
// (with its genuine T_C_root embedded in the ring winner, so verify's content-
// tree recomputation matches) plus the 3 ATTESTATION_EVIDENCE records sign
// requires, ready for realSignRun to produce a real SIGNATURE segment over it.
func writeSignableDocWithEvidence(t *testing.T) string {
	t.Helper()
	id := cliUnit(0x30)
	frame := realFrame(0x01, id)
	recs := []integrity.ContentRecord{{UnitID: id, Frame: frame}}
	tcRoot, err := integrity.TCRoot(recs)
	if err != nil {
		t.Fatalf("TCRoot: %v", err)
	}

	credID := cliUnit(0xD1)
	revID := cliUnit(0xD2)
	timeID := cliUnit(0xD3)
	cred := integrity.AttestationEvidence{ID: credID, Kind: integrity.AeCredentialChain, Format: integrity.AeFormatX509Chain, DEROctets: []byte("x509")}
	rev := integrity.AttestationEvidence{ID: revID, Kind: integrity.AeRevocationEvidence, Format: integrity.AeFormatOCSP, DEROctets: []byte("ocsp")}
	tm := integrity.AttestationEvidence{ID: timeID, Kind: integrity.AeTimeAttestation, Format: integrity.AeFormatTimestamp, DEROctets: []byte("tst"), NestedCredChain: credID, NestedRevocation: revID}
	credB, err := cred.Encode()
	if err != nil {
		t.Fatalf("cred encode: %v", err)
	}
	revB, err := rev.Encode()
	if err != nil {
		t.Fatalf("rev encode: %v", err)
	}
	tmB, err := tm.Encode()
	if err != nil {
		t.Fatalf("time encode: %v", err)
	}

	return writeDocWithFramesTCRootAndAttest(t, []pdlfmt.UnitID{id}, [][]byte{frame}, tcRoot,
		[][]byte{attestSegmentBytes(credB), attestSegmentBytes(revB), attestSegmentBytes(tmB)})
}

// TestTR_012_VerifyDistinguishesSignatureFromEvidence is T-0394's named
// integration test (cli.md S5, real decode bug fix). Before this fix, verify
// never stripped an ATTEST segment's leading ledger segment header before
// decoding, so integrity.DecodeSignatureRecord failed on EVERY real ATTEST
// segment -- reporting "unverified" for a genuine SIGNATURE record as well as
// for the 3 unrelated ATTESTATION_EVIDENCE records sitting alongside it, and
// reporting 4 "signatures" entries for what cli.md S5 defines as exactly one
// signature. This proves verify now reports exactly ONE signatures entry (the
// real SIGNATURE, correctly decoded) and silently skips the 3 evidence
// records, never mis-decoding them as signatures.
func TestTR_012_VerifyDistinguishesSignatureFromEvidence(t *testing.T) {
	doc := writeSignableDocWithEvidence(t)
	signed := realSignRun(doc, "k1", "total", nil)
	if signed.Err != nil || signed.Refused {
		t.Fatalf("signing fixture: err=%v refused=%v reason=%q", signed.Err, signed.Refused, signed.RefusalReason)
	}
	signedPath := filepath.Join(t.TempDir(), "signed.pdl")
	if err := os.WriteFile(signedPath, signed.Output, 0o600); err != nil {
		t.Fatalf("writing signed fixture: %v", err)
	}

	res := runVerify([]string{signedPath}, nil)
	sigs, ok := res.Extra["signatures"].([]map[string]any)
	if !ok {
		t.Fatalf("verify envelope missing signatures payload: %+v", res.Extra)
	}
	// Exactly one signature entry: the 3 ATTESTATION_EVIDENCE segments must be
	// skipped, never reported as (mis-decoded, "unverified") signatures.
	if len(sigs) != 1 {
		t.Fatalf("got %d signature entries, want exactly 1 (3 ATTESTATION_EVIDENCE + 1 SIGNATURE segments exist); before the fix every ATTEST segment was mis-reported: %+v", len(sigs), sigs)
	}
	// The real SIGNATURE record must actually decode (the header-strip fix):
	// a decode failure is reported as "unverified" with no other fields: the
	// real backend reaches a coverage/state verdict instead, proving the body
	// was decoded successfully rather than rejected as malformed.
	if sigs[0]["verdict"] == "unverified" {
		t.Errorf("real SIGNATURE record failed to decode (verdict=unverified); the leading ledger segment header was likely not stripped before DecodeSignatureRecord")
	}
}
