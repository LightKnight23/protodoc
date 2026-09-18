package integrity

import (
	"errors"
	"testing"
)

// TestFR_070_SigIntentRejectsOutOfRange is T-0177's named unit test (FR-070).
// sig-intent is a value from a CLOSED set (S10.3): 0x00 author-approval, 0x01
// witness-attestation, 0x02 notarization, 0x03 custodial-transfer. Every
// reserved value 0x04-0xFF is rejected, never defaulted.
func TestFR_070_SigIntentRejectsOutOfRange(t *testing.T) {
	for _, ok := range []SigIntent{IntentAuthorApproval, IntentWitnessAttestation, IntentNotarization, IntentCustodialTransfer} {
		if err := ValidateSigIntent(ok); err != nil {
			t.Errorf("valid sig-intent 0x%02x rejected: %v", uint8(ok), err)
		}
	}
	for v := 4; v <= 0xFF; v++ {
		if err := ValidateSigIntent(SigIntent(v)); !errors.Is(err, ErrSigIntentOutOfRange) {
			t.Fatalf("reserved sig-intent 0x%02x: err = %v, want ErrSigIntentOutOfRange", v, err)
		}
	}
}
