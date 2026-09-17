package eddsa

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

const conformanceCorpusPath = "testdata/eddsa-protodoc-1/conformance_corpus_v1.txt"

// TestEDDSA_P1_CONFORMANCE_CORPUS_V1 is T-0109's named conformance test
// (vector id EDDSA_P1_CONFORMANCE_CORPUS_V1). It iterates every fixture in
// the golden corpus -- a genuine RFC 8032 signature cross-checked against
// stdlib, non-canonical A, non-canonical R, each of the 8 small-order
// encodings applied to A and separately to R, S==L and S>L -- and asserts
// Verify()'s accept/reject result matches each fixture's recorded verdict,
// with zero mismatches.
func TestEDDSA_P1_CONFORMANCE_CORPUS_V1(t *testing.T) {
	f, err := os.Open(conformanceCorpusPath)
	if err != nil {
		t.Fatalf("opening corpus %s: %v", conformanceCorpusPath, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024)
	count := 0
	sawAccept, sawReject := 0, 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			t.Fatalf("malformed corpus line: %q", line)
		}
		verdict, aHex, mHex, sHex := fields[0], fields[1], fields[2], fields[3]
		note := ""
		if len(fields) >= 5 {
			note = fields[4]
		}

		A := mustHex32(t, aHex)
		msg := mustHex32(t, mHex)
		sigB, err := hex.DecodeString(sHex)
		if err != nil || len(sigB) != SigValueSize {
			t.Fatalf("fixture %q: bad sig hex", note)
		}
		var sig [SigValueSize]byte
		copy(sig[:], sigB)

		got := Verify(A, msg, sig)
		wantAccept := verdict == "accept"
		if got != wantAccept {
			t.Fatalf("fixture %q (%s): Verify=%v, want %v", note, verdict, got, wantAccept)
		}
		switch verdict {
		case "accept":
			sawAccept++
			// The accept fixture must also be a genuine stdlib-valid sig.
			if !ed25519.Verify(ed25519.PublicKey(A[:]), msg[:], sig[:]) {
				t.Fatalf("accept fixture %q is not a genuine stdlib-valid signature", note)
			}
		case "reject":
			sawReject++
		default:
			t.Fatalf("fixture %q: unknown verdict %q", note, verdict)
		}
		count++
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scanning corpus: %v", err)
	}

	// The corpus must contain the full required set: >=1 accept and the 20
	// reject classes (2 non-canonical + 16 small-order + 2 scalar).
	if sawAccept < 1 {
		t.Fatalf("corpus has no accept fixture")
	}
	if sawReject < 20 {
		t.Fatalf("corpus has %d reject fixtures, want at least 20 (2 non-canonical + 16 small-order + 2 scalar)", sawReject)
	}
	if count < 21 {
		t.Fatalf("corpus has %d fixtures, want at least 21", count)
	}
}

func mustHex32(t *testing.T, s string) [32]byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 32 {
		t.Fatalf("bad 32-octet hex %q", s)
	}
	var out [32]byte
	copy(out[:], b)
	return out
}
