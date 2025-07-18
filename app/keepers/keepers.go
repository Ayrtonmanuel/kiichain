package keepers

import (
	"errors"
	"os"

	"github.com/spf13/cast"

	pfmrouter "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	pfmrouterkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	pfmroutertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v8"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	srvflags "github.com/cosmos/evm/server/flags"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/cosmos/evm/x/ibc/transfer"
	ibctransferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v3/wasmbinding"
	oraclekeeper "github.com/kiichain/kiichain/v3/x/oracle/keeper"
	oracletypes "github.com/kiichain/kiichain/v3/x/oracle/types"
	rewardskeeper "github.com/kiichain/kiichain/v3/x/rewards/keeper"
	rewardstypes "github.com/kiichain/kiichain/v3/x/rewards/types"
	tokenfactorykeeper "github.com/kiichain/kiichain/v3/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/kiichain/kiichain/v3/x/tokenfactory/types"
)

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.Keeper
	CapabilityKeeper   *capabilitykeeper.Keeper
	StakingKeeper      *stakingkeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	DistrKeeper        distrkeeper.Keeper
	GovKeeper          *govkeeper.Keeper
	CrisisKeeper       *crisiskeeper.Keeper
	UpgradeKeeper      *upgradekeeper.Keeper
	ParamsKeeper       paramskeeper.Keeper
	WasmKeeper         wasmkeeper.Keeper
	TokenFactoryKeeper tokenfactorykeeper.Keeper
	RewardsKeeper      rewardskeeper.Keeper
	// IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCKeeper             *ibckeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper

	PFMRouterKeeper *pfmrouterkeeper.Keeper
	RatelimitKeeper ratelimitkeeper.Keeper

	// Modules
	ICAModule       ica.AppModule
	IBCFeeKeeper    ibcfeekeeper.Keeper
	TransferModule  transfer.AppModule
	PFMRouterModule pfmrouter.AppModule
	RateLimitModule ratelimit.AppModule

	// EVM keepers
	FeeMarketKeeper feemarketkeeper.Keeper
	EVMKeeper       *evmkeeper.Keeper
	Erc20Keeper     erc20keeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICSproviderkeeper   capabilitykeeper.ScopedKeeper
	scopedWasmKeeper          capabilitykeeper.ScopedKeeper
}

var tokenFactoryCapabilities = []string{
	tokenfactorytypes.EnableSetMetadata,
	tokenfactorytypes.EnableCommunityPoolFeeFunding,
}

func NewAppKeeper(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	legacyAmino *codec.LegacyAmino,
	maccPerms map[string][]string,
	modAccAddrs map[string]bool,
	blockedAddress map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	logger log.Logger,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
) AppKeepers {
	appKeepers := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()

	/*
		configure state listening capabilities using AppOptions
		we are doing nothing with the returned streamingServices and waitGroup in this case
	*/
	// load state streaming if enabled

	if err := bApp.RegisterStreamingServices(appOpts, appKeepers.keys); err != nil {
		logger.Error("failed to load state streaming", "err", err)
		os.Exit(1)
	}

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		appKeepers.keys[paramstypes.StoreKey],
		appKeepers.tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(appKeepers.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[capabilitytypes.StoreKey],
		appKeepers.memKeys[capabilitytypes.MemStoreKey],
	)

	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedICAControllerKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.scopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	appKeepers.CapabilityKeeper.Seal()

	// Add normal keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[banktypes.StoreKey]),
		appKeepers.AccountKeeper,
		blockedAddress,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[crisistypes.StoreKey]),
		invCheckPeriod,
		appKeepers.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		appKeepers.AccountKeeper.AddressCodec(),
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(appKeepers.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[feegrant.StoreKey]),
		appKeepers.AccountKeeper,
	)

	appKeepers.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[stakingtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[distrtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(appKeepers.keys[slashingtypes.StoreKey]),
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks(),
		),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// UpgradeKeeper must be created before IBCKeeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// gov depends on provider, so needs to be set after
	govConfig := govtypes.DefaultConfig()
	// set the MaxMetadataLen for proposals to the same value as it was pre-sdk v0.47.x
	govConfig.MaxMetadataLen = 10200
	appKeepers.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[govtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		// use the ProviderKeeper as StakingKeeper for gov
		// because governance should be based on the consensus-active validators
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper))

	// Set legacy router for backwards compatibility with gov v1beta1
	appKeepers.GovKeeper.SetLegacyRouter(govRouter)

	appKeepers.GovKeeper = appKeepers.GovKeeper.SetHooks(
		govtypes.NewMultiGovHooks(),
	)

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[evidencetypes.StoreKey]),
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
		appKeepers.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	appKeepers.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec, appKeepers.keys[ibcfeetypes.StoreKey],
		appKeepers.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper,
	)

	// ICA Host keeper
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icahosttypes.StoreKey],
		appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.ScopedICAHostKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// required since ibc-go v7.5.0
	appKeepers.ICAHostKeeper.WithQueryRouter(bApp.GRPCQueryRouter())

	govAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Create RateLimit keeper
	appKeepers.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
		appCodec, // BinaryCodec
		runtime.NewKVStoreService(appKeepers.keys[ratelimittypes.StoreKey]), // StoreKey
		appKeepers.GetSubspace(ratelimittypes.ModuleName),                   // param Subspace
		govAuthority, // authority
		appKeepers.BankKeeper,
		appKeepers.IBCKeeper.ChannelKeeper, // ChannelKeeper
		appKeepers.IBCFeeKeeper,            // ICS4Wrapper
	)

	// ICA Controller keeper
	appKeepers.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icacontrollertypes.StoreKey],
		appKeepers.GetSubspace(icacontrollertypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, // ICS4Wrapper
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedICAControllerKeeper,
		bApp.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// PFMRouterKeeper must be created before TransferKeeper
	appKeepers.PFMRouterKeeper = pfmrouterkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[pfmroutertypes.StoreKey],
		nil, // Will be zero-value here. Reference is set later on with SetTransferKeeper.
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.BankKeeper,
		appKeepers.RatelimitKeeper, // ICS4Wrapper
		govAuthority,
	)

	// Rewards Keeper
	appKeepers.RewardsKeeper = rewardskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[rewardstypes.StoreKey]),
		appKeepers.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authtypes.FeeCollectorName,
	)

	// Cosmos EVM keepers
	appKeepers.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec, authtypes.NewModuleAddress(govtypes.ModuleName),
		appKeepers.keys[feemarkettypes.StoreKey],
		appKeepers.tkeys[feemarkettypes.TransientKey],
		appKeepers.GetSubspace(feemarkettypes.ModuleName),
	)

	// Set up EVM keeper
	tracer := cast.ToString(appOpts.Get(srvflags.EVMTracer))

	appKeepers.EVMKeeper = evmkeeper.NewKeeper(
		// TODO: check why this is not adjusted to use the runtime module methods like SDK native keepers
		appCodec, appKeepers.keys[evmtypes.StoreKey], appKeepers.tkeys[evmtypes.TransientKey],
		authtypes.NewModuleAddress(govtypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.FeeMarketKeeper,
		&appKeepers.Erc20Keeper,
		tracer, appKeepers.GetSubspace(evmtypes.ModuleName),
	)

	appKeepers.Erc20Keeper = erc20keeper.NewKeeper(
		appKeepers.keys[erc20types.StoreKey],
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.EVMKeeper,
		appKeepers.StakingKeeper,
		appKeepers.AuthzKeeper,
		&appKeepers.TransferKeeper,
	)

	appKeepers.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		appKeepers.PFMRouterKeeper, // ISC4 Wrapper: PFM Router middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		appKeepers.Erc20Keeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[oracletypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[tokenfactorytypes.StoreKey],
		maccPerms,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		tokenFactoryCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	wasmOpts = append(
		wasmOpts,
		wasmbinding.RegisterCustomPlugins(
			appKeepers.BankKeeper,
			&appKeepers.TokenFactoryKeeper,
			appKeepers.EVMKeeper,
			appKeepers.OracleKeeper,
		)...,
	)

	// Must be called on PFMRouter AFTER TransferKeeper initialized
	appKeepers.PFMRouterKeeper.SetTransferKeeper(appKeepers.TransferKeeper)

	wasmDir := homePath
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	appKeepers.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[wasmtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCFeeKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.scopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		append(wasmkeeper.BuiltInCapabilities(), "kiichain"),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Middleware Stacks
	appKeepers.ICAModule = ica.NewAppModule(&appKeepers.ICAControllerKeeper, &appKeepers.ICAHostKeeper)
	appKeepers.TransferModule = transfer.NewAppModule(appKeepers.TransferKeeper)
	appKeepers.PFMRouterModule = pfmrouter.NewAppModule(appKeepers.PFMRouterKeeper, appKeepers.GetSubspace(pfmroutertypes.ModuleName))
	appKeepers.RateLimitModule = ratelimit.NewAppModule(appCodec, appKeepers.RatelimitKeeper)

	// Create Transfer Stack (from bottom to top of stack)
	// - core IBC
	// - ibcfee
	// - ratelimit
	// - pfm
	// - provider
	// - transfer
	//
	// This is how transfer stack will work in the end:
	// * RecvPacket -> IBC core -> Fee -> RateLimit -> PFM -> Provider -> Transfer (AddRoute)
	// * SendPacket -> Transfer -> Provider -> PFM -> RateLimit -> Fee -> IBC core (ICS4Wrapper)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(appKeepers.TransferKeeper)
	transferStack = pfmrouter.NewIBCMiddleware(
		transferStack,
		appKeepers.PFMRouterKeeper,
		0, // retries on timeout
		pfmrouterkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	transferStack = ratelimit.NewIBCMiddleware(appKeepers.RatelimitKeeper, transferStack)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, appKeepers.IBCFeeKeeper)

	// Create ICAHost Stack
	var icaHostStack porttypes.IBCModule = icahost.NewIBCModule(appKeepers.ICAHostKeeper)

	// Create Interchain Accounts Controller Stack
	var icaControllerStack porttypes.IBCModule = icacontroller.NewIBCMiddleware(nil, appKeepers.ICAControllerKeeper)

	var wasmStack porttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, appKeepers.IBCFeeKeeper)

	// Create IBC Router & seal
	ibcRouter := porttypes.NewRouter().
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasmtypes.ModuleName, wasmStack)

	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	// Configure EVM precompiles
	corePrecompiles := NewAvailableStaticPrecompiles(
		*appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		appKeepers.Erc20Keeper,
		appKeepers.AuthzKeeper,
		appKeepers.TransferKeeper,
		appKeepers.IBCKeeper.ClientKeeper,
		appKeepers.IBCKeeper.ConnectionKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.EVMKeeper,
		*appKeepers.GovKeeper,
		appKeepers.SlashingKeeper,
		appKeepers.EvidenceKeeper,
		appKeepers.WasmKeeper,
		appKeepers.OracleKeeper,
	)
	appKeepers.EVMKeeper.WithStaticPrecompiles(
		corePrecompiles,
	)

	return appKeepers
}

// GetSubspace returns a param subspace for a given module name.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		panic("couldn't load subspace for module: " + moduleName)
	}
	return subspace
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())         //nolint: staticcheck
	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable())   //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())         //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())       //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())              //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())     //nolint: staticcheck // SA1019
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	paramsKeeper.Subspace(pfmroutertypes.ModuleName)
	paramsKeeper.Subspace(ratelimittypes.ModuleName).WithKeyTable(ratelimittypes.ParamKeyTable())
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(tokenfactorytypes.ModuleName)
	paramsKeeper.Subspace(rewardstypes.ModuleName)
	paramsKeeper.Subspace(oracletypes.ModuleName)

	// Cosmos EVM modules
	paramsKeeper.Subspace(evmtypes.ModuleName)
	paramsKeeper.Subspace(feemarkettypes.ModuleName)
	paramsKeeper.Subspace(erc20types.ModuleName)

	return paramsKeeper
}

type DefaultFeemarketDenomResolver struct{}

func (r *DefaultFeemarketDenomResolver) ConvertToDenom(_ sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.DecCoin{}, errors.New("error resolving denom")
}

func (r *DefaultFeemarketDenomResolver) ExtraDenoms(_ sdk.Context) ([]string, error) {
	return []string{}, nil
}
