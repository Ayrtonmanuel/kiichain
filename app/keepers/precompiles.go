package keepers

import (
	"fmt"
	"maps"

	"github.com/ethereum/go-ethereum/common"

	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	connectionkeeper "github.com/cosmos/ibc-go/v8/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	bankprecompile "github.com/cosmos/evm/precompiles/bank"
	"github.com/cosmos/evm/precompiles/bech32"
	distprecompile "github.com/cosmos/evm/precompiles/distribution"
	evidenceprecompile "github.com/cosmos/evm/precompiles/evidence"
	govprecompile "github.com/cosmos/evm/precompiles/gov"
	ics20precompile "github.com/cosmos/evm/precompiles/ics20"
	"github.com/cosmos/evm/precompiles/p256"
	slashingprecompile "github.com/cosmos/evm/precompiles/slashing"
	stakingprecompile "github.com/cosmos/evm/precompiles/staking"
	erc20Keeper "github.com/cosmos/evm/x/erc20/keeper"
	transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	"github.com/cosmos/evm/x/vm/core/vm"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"

	"github.com/kiichain/kiichain/v3/precompiles/ibc"
	"github.com/kiichain/kiichain/v3/precompiles/oracle"
	"github.com/kiichain/kiichain/v3/precompiles/wasmd"
	oraclekeeper "github.com/kiichain/kiichain/v3/x/oracle/keeper"
)

const bech32PrecompileBaseGas = 6_000

// NewAvailableStaticPrecompiles returns the list of all available static precompiled contracts from EVM.
//
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	evidenceKeeper evidencekeeper.Keeper,
	wasmdKeeper wasmkeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	// Prepare the bech32 precompile
	bech32Precompile, err := bech32.NewPrecompile(bech32PrecompileBaseGas)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	// Prepare the staking precompile
	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	// Prepare the distribution precompile
	distributionPrecompile, err := distprecompile.NewPrecompile(
		distributionKeeper,
		stakingKeeper,
		authzKeeper,
		evmKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	// Prepare the ibc precompile
	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(
		stakingKeeper,
		transferKeeper,
		channelKeeper,
		authzKeeper,
		evmKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	// Prepare the bank precompile
	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	// Prepare the gov precompile
	govPrecompile, err := govprecompile.NewPrecompile(govKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate gov precompile: %w", err))
	}

	// Prepare the slashing precompile
	slashingPrecompile, err := slashingprecompile.NewPrecompile(slashingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate slashing precompile: %w", err))
	}

	// Prepare the evidence precompile
	evidencePrecompile, err := evidenceprecompile.NewPrecompile(evidenceKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate evidence precompile: %w", err))
	}

	// Prepare the wasmd precompile
	wasmdPrecompile, err := wasmd.NewPrecompile(wasmdKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate wasmd precompile: %w", err))
	}

	// Prepare the ibc precompile
	ibcPrecompile, err := ibc.NewPrecompile(transferKeeper, clientKeeper, connectionKeeper, channelKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ibc precompile: %w", err))
	}

	// Prepare the oracle precompile
	oraclePrecompile, err := oracle.NewPrecompile(oracleKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate oracle precompile: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile
	precompiles[govPrecompile.Address()] = govPrecompile
	precompiles[slashingPrecompile.Address()] = slashingPrecompile
	precompiles[evidencePrecompile.Address()] = evidencePrecompile
	precompiles[wasmdPrecompile.Address()] = wasmdPrecompile
	precompiles[ibcPrecompile.Address()] = ibcPrecompile
	precompiles[oraclePrecompile.Address()] = oraclePrecompile

	// Return the precompiles
	return precompiles
}
