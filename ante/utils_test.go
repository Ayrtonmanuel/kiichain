package ante_test

import (
	"fmt"
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/evm/ante"
	"github.com/cosmos/evm/crypto/ethsecp256k1"
)

// generatePubKeysAndSignatures generates random pub keys and signatures
func generatePubKeysAndSignatures(n int, msg []byte, _ bool) (pubkeys []cryptotypes.PubKey, signatures [][]byte) {
	pubkeys = make([]cryptotypes.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		privkey, _ := ethsecp256k1.GenerateKey()
		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

// expectedGasCostByKeys check the expected gas costs per bytes
func expectedGasCostByKeys(pubkeys []cryptotypes.PubKey) uint64 {
	cost := uint64(0)
	for _, pubkey := range pubkeys {
		pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
		switch {
		case strings.Contains(pubkeyType, "ed25519"):
			cost += authtypes.DefaultSigVerifyCostED25519
		case strings.Contains(pubkeyType, "ethsecp256k1"):
			cost += ante.Secp256k1VerifyCost
		default:
			panic("unexpected key type")
		}
	}
	return cost
}
