// Publish (T-0196..T-0198, FR-078/FR-080/FR-081; plan.md L*(published_state)).
// The publish operation re-emits a document state containing NO octet sequence
// belonging to any removed, superseded, or erased content unit -- verified by
// an octet-level residue scan over the output (including any layout advances,
// cached renderings, index entries, and preview payloads). It is a full
// re-serialization from the retained content, not an in-place edit, so a
// removed unit's octets cannot linger anywhere.
package integrity

import "bytes"

// PublishInput describes a document state to publish: the content records that
// remain, and the frame octets of units that were removed or superseded (whose
// octets must not appear anywhere in the output).
type PublishInput struct {
	// Retained is the content records that survive into the published output
	// (non-redacted, non-removed). A redacted record retained here carries only
	// its RetainedLeaf commitment (no frame), which is emitted as its leaf, not
	// as plaintext.
	Retained []ContentRecord
	// RemovedFrames is the frame octet sequences of every removed or
	// superseded content unit -- the octets that MUST NOT appear in the
	// output.
	RemovedFrames [][]byte
	// ActorIdentityValues is the set of octet values drawn from the
	// actor-identity inventory (FR-079) that MUST NOT appear in the output
	// (FR-080). The publish operation strips them; the output is scanned to
	// confirm none survives.
	ActorIdentityValues [][]byte
	// CustodyFixityValues is the set of custody and fixity octet values
	// (signatures, time attestations) that MUST be present UNCHANGED in the
	// output (FR-081). Publish appends them verbatim after the content and
	// never strips them, even if one happens to contain an actor-identity
	// substring (custody/fixity preservation takes precedence for these
	// explicitly-preserved values).
	CustodyFixityValues [][]byte
}

// PublishOutput is a published document's emitted octets plus the retained
// unit set. Emitted is the concatenation of retained frames (redacted records
// contribute only their commitment leaf, never plaintext).
type PublishOutput struct {
	Emitted []byte
}

// Publish produces the published output: a fresh re-emission of only the
// retained content. A retained non-redacted record contributes its frame
// octets; a retained redacted record contributes only its 32-octet commitment
// leaf (never its removed plaintext, which is gone). No removed-unit octets are
// ever copied into the output, by construction (the output is built from the
// retained set alone).
func Publish(in PublishInput) PublishOutput {
	var out []byte
	for _, r := range in.Retained {
		if r.Redacted {
			out = append(out, r.RetainedLeaf[:]...)
			continue
		}
		out = append(out, r.Frame...)
	}
	// Strip every actor-identity value (FR-080): the published output must
	// contain no value from the identity inventory. Each value's octet
	// occurrences are removed from the emitted stream.
	for _, v := range in.ActorIdentityValues {
		if len(v) == 0 {
			continue
		}
		out = stripAll(out, v)
	}
	// Append every custody and fixity value verbatim AFTER stripping, so they
	// are preserved unchanged (FR-081) and never removed by the actor-identity
	// strip above -- custody/fixity preservation is explicit and takes
	// precedence for these values.
	for _, v := range in.CustodyFixityValues {
		out = append(out, v...)
	}
	return PublishOutput{Emitted: out}
}

// stripAll removes every non-overlapping occurrence of needle from b.
func stripAll(b, needle []byte) []byte {
	if len(needle) == 0 {
		return b
	}
	var out []byte
	for {
		i := bytes.Index(b, needle)
		if i < 0 {
			out = append(out, b...)
			break
		}
		out = append(out, b[:i]...)
		b = b[i+len(needle):]
	}
	return out
}

// ResidueFinding names a removed unit whose octet sequence was found in the
// published output (a residue leak -- FR-078 violation).
type ResidueFinding struct {
	RemovedIndex int
	Offset       int
}

// ScanResidue searches the published output for any octet sequence belonging
// to a removed unit (FR-078's octet-level verification). It returns every
// removed-frame occurrence found in the output; an empty result means the
// publish is clean (no removed-unit octets survive anywhere). A zero-length
// removed frame is ignored (an empty sequence trivially "occurs" everywhere
// and is not a meaningful residue).
func ScanResidue(out PublishOutput, removedFrames [][]byte) []ResidueFinding {
	var findings []ResidueFinding
	for i, rf := range removedFrames {
		if len(rf) == 0 {
			continue
		}
		if idx := bytes.Index(out.Emitted, rf); idx >= 0 {
			findings = append(findings, ResidueFinding{RemovedIndex: i, Offset: idx})
		}
	}
	return findings
}
