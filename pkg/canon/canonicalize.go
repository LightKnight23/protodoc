// Streaming canonicalization (NFR-002; T-0308). Canonicalize streams C(S)
// directly to an io.Writer using the shared canonical traversal (T-0307),
// holding at most ONE content-subtree's serialized bytes in memory at a time —
// it never materializes the complete canonical sequence as an in-memory buffer
// or temp file. Peak allocation therefore scales with the largest single
// subtree, not with total document size.
package canon

import "io"

// Canonicalize writes the canonical octet sequence C(S) of state to w, subtree
// by subtree in canonical (T_C) order. It streams: each subtree's frame is
// written to w and then dropped before the next is considered, so at most one
// subtree's bytes are held at a time. The BOTTOM state writes nothing. The
// output is byte-identical to a reference in-memory serialization of the same
// canonical order.
func Canonicalize(state *Document, w io.Writer) error {
	for _, sub := range Traverse(state) {
		if _, err := w.Write(sub.Frame); err != nil {
			return err
		}
	}
	return nil
}
