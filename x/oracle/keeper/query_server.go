package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/oracle/types"
)

// QueryServer struct that handlers the rpc request
type QueryServer struct {
	Keeper Keeper
}

// Ensure the struct queryServer implement the QueryServer interface
var _ types.QueryServer = QueryServer{}

// NewQueryServer returns a new instance of the QueryServer
func NewQueryServer(keeper Keeper) QueryServer {
	return QueryServer{
		Keeper: keeper,
	}
}

// Params returns the oracle's params
func (qs QueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// Get the module's params from the keeper
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.Keeper.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &params}, nil
}

// ExchangeRate returns the exchange rate specific by denom
func (qs QueryServer) ExchangeRate(ctx context.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	// Validate 544
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.Denom) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty denom")
	}

	// Get exchange rate by denom
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	exchangeRate, err := qs.Keeper.ExchangeRate.Get(sdkCtx, req.Denom)
	if err != nil {
		return nil, err
	}

	// Prepare response
	response := &types.QueryExchangeRateResponse{
		OracleExchangeRate: &exchangeRate,
	}

	return response, nil
}

// ExchangeRates returns all exchange rates
func (qs QueryServer) ExchangeRates(ctx context.Context, req *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	exchangeRates := []types.DenomOracleExchangeRate{}
	err := qs.Keeper.ExchangeRate.Walk(sdkCtx, nil, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		exchangeRates = append(exchangeRates, types.DenomOracleExchangeRate{Denom: denom, OracleExchangeRate: &exchangeRate})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryExchangeRatesResponse{DenomOracleExchangeRate: exchangeRates}, nil
}

// Actives queries all denoms for which exchange rates exist
func (qs QueryServer) Actives(ctx context.Context, req *types.QueryActivesRequest) (*types.QueryActivesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	denomsActive := []string{}
	err := qs.Keeper.ExchangeRate.Walk(sdkCtx, nil, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		denomsActive = append(denomsActive, denom)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryActivesResponse{Actives: denomsActive}, nil
}

// VoteTargets queries the voting target list on current vote period
func (qs QueryServer) VoteTargets(ctx context.Context, req *types.QueryVoteTargetsRequest) (*types.QueryVoteTargetsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Get the vote targets
	voteTargets, err := qs.Keeper.GetVoteTargets(sdkCtx)

	// Return the response and the error
	return &types.QueryVoteTargetsResponse{VoteTargets: voteTargets}, err
}

// PriceSnapshotHistory queries all snapshots
func (qs QueryServer) PriceSnapshotHistory(ctx context.Context, req *types.QueryPriceSnapshotHistoryRequest) (*types.QueryPriceSnapshotHistoryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get the snapshots available on the KVStore
	priceSnapshots := []types.PriceSnapshot{}
	err := qs.Keeper.PriceSnapshot.Walk(sdkCtx, nil, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
		priceSnapshots = append(priceSnapshots, snapshot)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryPriceSnapshotHistoryResponse{PriceSnapshot: priceSnapshots}, nil
}

// Twaps queries the Time-weighted average price (TWAPs) whitin an specific period of time
func (qs QueryServer) Twaps(ctx context.Context, req *types.QueryTwapsRequest) (*types.QueryTwapsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	twaps, err := qs.Keeper.CalculateTwaps(sdkCtx, req.LookbackSeconds)
	if err != nil {
		return nil, err
	}

	return &types.QueryTwapsResponse{OracleTwap: twaps}, err
}

// FeederDelegation queries the account data address assigned as a delegator by a validator
func (qs QueryServer) FeederDelegation(ctx context.Context, req *types.QueryFeederDelegationRequest) (*types.QueryFeederDelegationResponse, error) {
	// Validate request information
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the delegator by the Validator address
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	feederDelegation, err := qs.Keeper.GetFeederDelegationOrDefault(sdkCtx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeederDelegationResponse{FeedAddr: feederDelegation.String()}, nil
}

// VotePenaltyCounter queries the validator penalty's counter information
func (qs QueryServer) VotePenaltyCounter(ctx context.Context, req *types.QueryVotePenaltyCounterRequest) (*types.QueryVotePenaltyCounterResponse, error) {
	// Validate request information
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the penalty counters by the validator address
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	voteCounter, err := qs.Keeper.VotePenaltyCounter.Get(sdkCtx, valAddr)
	if err != nil {
		return nil, err
	}
	return &types.QueryVotePenaltyCounterResponse{VotePenaltyCounter: &voteCounter}, nil
}

// SlashWindow queries the slash window progress
func (qs QueryServer) SlashWindow(ctx context.Context, req *types.QuerySlashWindowRequest) (*types.QuerySlashWindowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.Keeper.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	// The window progress is the number of vote periods that have been completed in the current slashing window.
	// With a vote period of 1, this will be equivalent to the number of blocks that have progressed in the slash window.
	windowProgress := (uint64(sdkCtx.BlockHeight()) % params.SlashWindow) / params.VotePeriod

	return &types.QuerySlashWindowResponse{WindowProgress: windowProgress}, nil
}
