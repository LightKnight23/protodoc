package content

import (
	"encoding/binary"
	"sync/atomic"

	"Protodoc/pkg/pdlfmt"
)

// testMintCounter drives testMintID. Tests only need distinct identities,
// not CSPRNG entropy (the CSPRNG minting primitive is tested in its own
// package, pkg/content/mint); using a deterministic counter here keeps the
// content package's own tests free of any crypto dependency, consistent with
// package content being crypto-free.
var testMintCounter uint64

// testMintID returns a distinct 16-octet identity for tests. It is not a
// CSPRNG mint; it exists solely to hand tests unique unit ids without
// importing the crypto-backed minter.
func testMintID() (pdlfmt.UnitID, error) {
	n := atomic.AddUint64(&testMintCounter, 1)
	var id pdlfmt.UnitID
	binary.BigEndian.PutUint64(id[0:8], 0xC0FFEE0000000000|n)
	binary.BigEndian.PutUint64(id[8:16], n*0x9E3779B97F4A7C15)
	return id, nil
}
