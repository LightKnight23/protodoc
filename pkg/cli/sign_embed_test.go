package cli

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"Protodoc/pkg/container"
	"Protodoc/pkg/integrity"
	"Protodoc/pkg/ledger"
	"Protodoc/pkg/pdlfmt"
)

// attestSegmentBytes frames an ATTEST record body in a ledger ATTEST segment
// header (matching how realSignRun frames its own SIGNATURE segment).
func attestSegmentBytes(body []byte) []byte {
	h := ledger.EncodeSegmentHeader(ledger.SegmentHeader{Type: container.SegmentTypeAttest, FrameCount: 1})
	return append(append([]byte(nil), h...), body...)
}

// writeDocWithEvidence writes a valid document whose ATTEST segments carry a
// credential-chain, a revocation, and a time-attestation ATTESTATION_EVIDENCE
// record (a fully-formed set sign can reference). Returns path + the cred-chain
// and time-attestation ae-ids.
func writeDocWithEvidence(t *testing.T) (path string, credID, timeID pdlfmt.UnitID) {
	t.Helper()
	credID = cliUnit(0xC1)
	revID := cliUnit(0xC2)
	timeID = cliUnit(0xC3)

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
	segs := [][]byte{attestSegmentBytes(credB), attestSegmentBytes(revB), attestSegmentBytes(tmB)}

	h := &container.Header{FormatMajor: 1, DocumentClass: 1, CapabilityWritten: 1, CapabilityRequired: 1, HistoryMode: container.HistoryComplete, UnicodeVersionID: 1, PrefixLayoutID: 1}
	var ring [container.CommitRingSlots]container.CommitRingRecord
	base := uint64(prefixSize)
	off := base
	slots := make([]container.SegmentTableSlot, len(segs))
	body := uint64(0)
	for i, s := range segs {
		var dg [32]byte
		copy(dg[:], hashSeg(s))
		slots[i] = container.SegmentTableSlot{SegmentType: container.SegmentTypeAttest, Offset: off, Length: uint64(len(s)), FrameCount: 1, Digest: dg}
		off += uint64(len(s))
		body += uint64(len(s))
	}
	total := base + body
	for i := range ring {
		ring[i] = container.CommitRingRecord{Sequence: uint64(i + 1), LedgerLength: total, SegmentCount: uint16(len(segs))}
	}
	img := make([]byte, 0, total)
	img = append(img, h.Encode(nil)...)
	img = append(img, container.EncodeCommitRing(&ring, nil)...)
	fmEnc, _ := (&container.Frontmatter{}).Encode(nil)
	img = append(img, fmEnc...)
	stEnc, err := container.EncodeSegmentTable(slots, nil)
	if err != nil {
		t.Fatalf("EncodeSegmentTable: %v", err)
	}
	img = append(img, stEnc...)
	if len(img) != prefixSize {
		t.Fatalf("prefix %d, want %d", len(img), prefixSize)
	}
	for _, s := range segs {
		img = append(img, s...)
	}
	path = filepath.Join(t.TempDir(), "evidence.pdl")
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path, credID, timeID
}

func hashSeg(b []byte) []byte {
	d := sha256.Sum256(b)
	return d[:]
}

// TestTR_012_SignVerbEmbedsRealSignature is T-0392's named integration test
// (TR-012, FR-063/FR-070). It proves BOTH paths against real files:
//   - a document lacking required evidence -> REFUSED (cli.md S11), no --out;
//   - a document carrying credential-chain + time-attestation evidence -> a
//     real SIGNATURE segment is spliced and the re-serialized signed document
//     is written to --out, decodable with one more segment than the input.
func TestTR_012_SignVerbEmbedsRealSignature(t *testing.T) {
	// --- Refusal path: writeValidPrefix has NO evidence. ---
	noEvidence := writeValidPrefix(t)
	outPath := filepath.Join(t.TempDir(), "signed.pdl")
	res := runSign([]string{noEvidence, "--key", "k1", "--coverage", "total", "--intent", "author-approval", "--out", outPath}, nil)
	if res.Status != "REFUSED" {
		t.Fatalf("no-evidence sign: status=%s, want REFUSED (cli.md S11)", res.Status)
	}
	if _, err := os.Stat(outPath); err == nil {
		t.Errorf("refusal must not write --out, but %s exists", outPath)
	}

	// --- Embed path: a document with real evidence. ---
	doc, _, _ := writeDocWithEvidence(t)
	out2 := filepath.Join(t.TempDir(), "signed2.pdl")
	res = runSign([]string{doc, "--key", "k1", "--coverage", "total", "--intent", "author-approval", "--out", out2}, nil)
	if res.Status != "OK" {
		t.Fatalf("evidence sign: status=%s (%+v), want OK", res.Status, res.Findings)
	}
	if res.Extra["signed_document_complete"] != true {
		t.Errorf("signed_document_complete must be true on the embed path")
	}

	// The written file exists, is larger than the input, and decodes with one
	// more ATTEST segment (the spliced SIGNATURE) than the input had.
	signedBytes, err := os.ReadFile(out2)
	if err != nil {
		t.Fatalf("reading signed --out: %v", err)
	}
	origBytes, _ := os.ReadFile(doc)
	if len(signedBytes) <= len(origBytes) {
		t.Errorf("signed document (%d) should be larger than the input (%d)", len(signedBytes), len(origBytes))
	}

	// Decode the signed document's segment table: it must carry a SIGNATURE
	// (0x40) record in a new ATTEST segment beyond the 3 evidence segments.
	table, err := container.DecodeSegmentTable(signedBytes[container.SegmentTableOffset:prefixSize])
	if err != nil {
		t.Fatalf("decode signed segment table: %v", err)
	}
	attestCount, foundSig := 0, false
	for _, slot := range table {
		if slot.SegmentType != container.SegmentTypeAttest {
			continue
		}
		attestCount++
		raw := signedBytes[slot.Offset : slot.Offset+slot.Length]
		disc, derr := readAttestRecordDiscriminant(attestSegmentBody(raw))
		if derr == nil && disc == discSignature {
			foundSig = true
			// The embedded SIGNATURE decodes and references the real evidence.
			sig, serr := integrity.DecodeSignatureRecord(attestSegmentBody(raw))
			if serr != nil {
				t.Errorf("embedded SIGNATURE does not decode: %v", serr)
			} else if sig.CredChainRef == (pdlfmt.UnitID{}) || sig.TimeAttestationRef == (pdlfmt.UnitID{}) {
				t.Errorf("embedded SIGNATURE must reference real cred-chain + time-attestation evidence")
			}
		}
	}
	if attestCount != 4 { // 3 evidence + 1 new signature
		t.Errorf("signed document has %d ATTEST segments, want 4 (3 evidence + 1 signature)", attestCount)
	}
	if !foundSig {
		t.Errorf("signed document must carry a spliced SIGNATURE (0x40) record")
	}

	// Determinism (NFR-006): signing twice yields byte-identical output.
	r1 := realSignRun(doc, "k1", "total", nil)
	r2 := realSignRun(doc, "k1", "total", nil)
	if !bytes.Equal(r1.Output, r2.Output) {
		t.Errorf("signing the same file+key twice must produce byte-identical output (NFR-006)")
	}
}
