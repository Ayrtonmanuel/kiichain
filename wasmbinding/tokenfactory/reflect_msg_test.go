package tokenfactory_test

import (
	"encoding/json"
	"fmt"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	app "github.com/kiichain/kiichain/v3/app"
	"github.com/kiichain/kiichain/v3/app/apptesting"
	"github.com/kiichain/kiichain/v3/wasmbinding"
	"github.com/kiichain/kiichain/v3/wasmbinding/helpers"
	bindingtypes "github.com/kiichain/kiichain/v3/wasmbinding/tokenfactory/types"
	"github.com/kiichain/kiichain/v3/x/tokenfactory/types"
)

// TestQueryFullDenom tests the query for full denom with a reflect contract
func TestCreateDenomMsg(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	lucky := apptesting.RandomAccountAddress()
	reflect := helpers.InstantiateReflectContract(t, ctx, app, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, reflect, reflectAmount)

	msg := bindingtypes.Msg{CreateDenom: &bindingtypes.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// query the denom and see if it matches
	query := bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	require.Equal(t, resp.Denom, fmt.Sprintf("factory/%s/SUN", reflect.String()))
}

// TestMintMsg tests the minting of tokens with a reflect contract
func TestMintMsg(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	lucky := apptesting.RandomAccountAddress()
	reflect := helpers.InstantiateReflectContract(t, ctx, app, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, reflect, reflectAmount)

	// lucky was broke
	balances := app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := bindingtypes.Msg{CreateDenom: &bindingtypes.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010808")
	require.True(t, ok)
	msg = bindingtypes.Msg{MintTokens: &bindingtypes.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin := balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query := bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// mint the same denom again
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin = balances[0]
	require.Equal(t, amount.MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// now mint another amount / denom
	// create it first
	msg = bindingtypes.Msg{CreateDenom: &bindingtypes.CreateDenom{
		Subdenom: "MOON",
	}}
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	moonDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount = amount.SubRaw(1)
	msg = bindingtypes.Msg{MintTokens: &bindingtypes.MintTokens{
		Denom:         moonDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 2)
	coin = balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "MOON",
		},
	}
	resp = bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// and check the first denom is unchanged
	coin = balances[1]
	require.Equal(t, amount.AddRaw(1).MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)
}

// TestForceTransfer tests the force transfer of tokens with a reflect contract
func TestForceTransfer(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	lucky := apptesting.RandomAccountAddress()
	rcpt := apptesting.RandomAccountAddress()
	reflect := helpers.InstantiateReflectContract(t, ctx, app, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, reflect, reflectAmount)

	// lucky was broke
	balances := app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := bindingtypes.Msg{CreateDenom: &bindingtypes.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010808")
	require.True(t, ok)

	// Mint new tokens to lucky
	msg = bindingtypes.Msg{MintTokens: &bindingtypes.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// Checks if force transfer is enabled
	capabilities := app.TokenFactoryKeeper.GetEnabledCapabilities()
	forceTransferEnabled := types.IsCapabilityEnabled(capabilities, types.EnableForceTransfer)

	if forceTransferEnabled {
		// Force move 100 tokens from lucky to rcpt
		msg = bindingtypes.Msg{ForceTransfer: &bindingtypes.ForceTransfer{
			Denom:       sunDenom,
			Amount:      sdkmath.NewInt(100),
			FromAddress: lucky.String(),
			ToAddress:   rcpt.String(),
		}}
		err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
		require.NoError(t, err)

		// check the balance of rcpt
		balances = app.BankKeeper.GetAllBalances(ctx, rcpt)
		require.Len(t, balances, 1)
		coin := balances[0]
		require.Equal(t, sdkmath.NewInt(100), coin.Amount)
	}
}

// TestBurnMsg tests the burn of tokens with a reflect contract
func TestBurnMsg(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	lucky := apptesting.RandomAccountAddress()
	reflect := helpers.InstantiateReflectContract(t, ctx, app, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, reflect, reflectAmount)

	// lucky was broke
	balances := app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := bindingtypes.Msg{CreateDenom: &bindingtypes.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdkmath.NewIntFromString("808010809")
	require.True(t, ok)

	msg = bindingtypes.Msg{MintTokens: &bindingtypes.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// Checks if burnFrom is enabled
	capabilities := app.TokenFactoryKeeper.GetEnabledCapabilities()
	burnFromEnabled := types.IsCapabilityEnabled(capabilities, types.EnableBurnFrom)

	if burnFromEnabled {
		// can burn from different address with burnFrom
		amt, ok := sdkmath.NewIntFromString("1")
		require.True(t, ok)
		msg = bindingtypes.Msg{BurnTokens: &bindingtypes.BurnTokens{
			Denom:           sunDenom,
			Amount:          amt,
			BurnFromAddress: lucky.String(),
		}}
		err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
		require.NoError(t, err)

		// lucky needs to send balance to reflect contract to burn it
		luckyBalance := app.BankKeeper.GetAllBalances(ctx, lucky)
		err = app.BankKeeper.SendCoins(ctx, lucky, reflect, luckyBalance)
		require.NoError(t, err)

		msg = bindingtypes.Msg{BurnTokens: &bindingtypes.BurnTokens{
			Denom:           sunDenom,
			Amount:          amount.Abs().Sub(sdkmath.NewInt(1)),
			BurnFromAddress: reflect.String(),
		}}
		err = executeCustom(t, ctx, app, reflect, lucky, msg, sdk.Coin{})
		require.NoError(t, err)
	}
}

// ReflectExec is a wrapper for the reflect message
type ReflectExec struct {
	ReflectMsg    *ReflectMsgs    `json:"reflect_msg,omitempty"`
	ReflectSubMsg *ReflectSubMsgs `json:"reflect_sub_msg,omitempty"`
}

// ReflectMsgs is a wrapper for the reflect message
type ReflectMsgs struct {
	Msgs []wasmvmtypes.CosmosMsg `json:"msgs"`
}

// ReflectSubMsgs is a wrapper for the reflect sub message
type ReflectSubMsgs struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}

// executeCustom executes a custom message on the reflect contract
func executeCustom(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contract sdk.AccAddress, sender sdk.AccAddress, msg bindingtypes.Msg, funds sdk.Coin) error { //nolint:unparam // funds is always nil but could change in the future.
	t.Helper()

	// Make the request a kiichain msg
	kiichainMsg := wasmbinding.KiichainMsg{
		TokenFactory: &msg,
	}

	customBz, err := json.Marshal(kiichainMsg)
	require.NoError(t, err)

	reflectMsg := ReflectExec{
		ReflectMsg: &ReflectMsgs{
			Msgs: []wasmvmtypes.CosmosMsg{{
				Custom: customBz,
			}},
		},
	}
	reflectBz, err := json.Marshal(reflectMsg)
	require.NoError(t, err)

	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	_, err = contractKeeper.Execute(ctx, contract, sender, reflectBz, coins)
	return err
}
