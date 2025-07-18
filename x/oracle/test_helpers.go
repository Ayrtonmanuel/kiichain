package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/kiichain/kiichain/v3/x/oracle/keeper"
	"github.com/kiichain/kiichain/v3/x/oracle/types"
)

var (
	stakingAmount       = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	randomAExchangeRate = math.LegacyNewDec(1700)
)

// SetUp returns the message server
func SetUp(t *testing.T) (keeper.TestInput, types.MsgServer) {
	t.Helper()
	input := keeper.CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx

	// Update params to test easier and faster
	params, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VotePeriod = 1
	params.SlashWindow = 100
	err = oracleKeeper.Params.Set(ctx, params)
	require.NoError(t, err)

	stakingParams, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	stakingParams.MinCommissionRate = math.LegacyNewDecWithPrec(0, 2) // 0.00
	err = stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	// Create handlers
	oracleMsgServer := keeper.NewMsgServer(oracleKeeper)
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	val0 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], stakingAmount)
	val1 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[1], keeper.ValPubKeys[1], stakingAmount)
	val2 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[2], keeper.ValPubKeys[2], stakingAmount)

	// Register validators
	_, err = stakingMsgServer.CreateValidator(ctx, val0)
	require.NoError(t, err)
	_, err = stakingMsgServer.CreateValidator(ctx, val1)
	require.NoError(t, err)
	_, err = stakingMsgServer.CreateValidator(ctx, val2)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	return input, oracleMsgServer
}

// TestTx is a mock transaction type for testing purposes
type TestTx struct {
	msgs []sdk.Msg
}

// NewTestTx creates a new TestTx with the provided messages
func NewTestTx(msgs []sdk.Msg) TestTx {
	return TestTx{msgs: msgs}
}

// GetMsgs returns the messages contained in the TestTx
func (t TestTx) GetMsgs() []sdk.Msg {
	return t.msgs
}

// ValidateBasic performs basic validation on the TestTx
func (t TestTx) ValidateBasic() error {
	return nil
}

func (t TestTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}
