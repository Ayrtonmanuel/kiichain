package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/kiichain/kiichain/v3/ante"
	"github.com/kiichain/kiichain/v3/app/helpers"
)

func TestGovExpeditedProposalsDecorator(t *testing.T) {
	kiiApp := helpers.Setup(t)

	testCases := []struct {
		name      string
		ctx       sdk.Context
		msgs      []sdk.Msg
		expectErr bool
	}{
		// these cases should pass
		{
			name: "expedited - govv1.MsgSubmitProposal - MsgSoftwareUpgrade",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&upgradetypes.MsgSoftwareUpgrade{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					Plan: upgradetypes.Plan{
						Name:   "upgrade plan-plan",
						Info:   "some text here",
						Height: 123456789,
					},
				}}, true),
			},
			expectErr: false,
		},
		{
			name: "expedited - govv1.MsgSubmitProposal - MsgCancelUpgrade",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&upgradetypes.MsgCancelUpgrade{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
				}}, true),
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - TextProposal",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newLegacyTextProp(false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgCommunityPoolSpend",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&distrtypes.MsgCommunityPoolSpend{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					Recipient: sdk.AccAddress{}.String(),
					Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgTransfer",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgSend{
					FromAddress: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					ToAddress:   "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					Amount:      sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgUpdateParams",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgUpdateParams{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
				}}, false),
			},
			expectErr: false,
		},
		// legacy proposals - antehandler should not affect them
		// submitted using "kiichaind tx gov submit-legacy-proposal"
		{
			name:      "normal - govv1beta.MsgSubmitProposal - LegacySoftwareUpgrade",
			ctx:       sdk.Context{},
			msgs:      []sdk.Msg{newGovV1BETA1LegacyUpgradeProp()},
			expectErr: false,
		},
		{
			name:      "normal - govv1beta.MsgSubmitProposal - LegacyCancelSoftwareUpgrade",
			ctx:       sdk.Context{},
			msgs:      []sdk.Msg{newGovV1BETA1LegacyCancelUpgradeProp()},
			expectErr: false,
		},
		// these cases should fail
		// these are normal proposals, not whitelisted for expedited voting
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - Empty",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - TextProposal",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newLegacyTextProp(true), // expedite
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgCommunityPoolSpend",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&distrtypes.MsgCommunityPoolSpend{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					Recipient: sdk.AccAddress{}.String(),
					Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgTransfer",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgSend{
					FromAddress: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					ToAddress:   "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
					Amount:      sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgUpdateParams",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgUpdateParams{
					Authority: "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
				}}, true),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			txCfg := kiiApp.GetTxConfig()
			decorator := ante.NewGovExpeditedProposalsDecorator(kiiApp.AppCodec())

			txBuilder := txCfg.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msgs...))

			_, err := decorator.AnteHandle(tc.ctx, txBuilder.GetTx(), false,
				func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil })
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newLegacyTextProp(expedite bool) *govv1.MsgSubmitProposal {
	testProposal := govv1beta1.NewTextProposal("Proposal", "Test as normal proposal")
	msgContent, err := govv1.NewLegacyContent(testProposal, "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv")
	if err != nil {
		return nil
	}
	return newGovProp([]sdk.Msg{msgContent}, expedite)
}

func newGovV1BETA1LegacyUpgradeProp() *govv1beta1.MsgSubmitProposal {
	legacyContent := upgradetypes.NewSoftwareUpgradeProposal("test legacy upgrade", "test legacy upgrade", upgradetypes.Plan{
		Name:   "upgrade plan-plan",
		Info:   "some text here",
		Height: 123456789,
	})

	msg, _ := govv1beta1.NewMsgSubmitProposal(legacyContent, sdk.NewCoins(), sdk.AccAddress{})
	return msg
}

func newGovV1BETA1LegacyCancelUpgradeProp() *govv1beta1.MsgSubmitProposal {
	legacyContent := upgradetypes.NewCancelSoftwareUpgradeProposal("test legacy upgrade", "test legacy upgrade")

	msg, _ := govv1beta1.NewMsgSubmitProposal(legacyContent, sdk.NewCoins(), sdk.AccAddress{})
	return msg
}

func newGovProp(msgs []sdk.Msg, expedite bool) *govv1.MsgSubmitProposal {
	msg, _ := govv1.NewMsgSubmitProposal(msgs, sdk.NewCoins(), sdk.AccAddress{}.String(), "", "expedite", "expedite", expedite)
	// fmt.Println("### msg ###", msg, "err", err)
	return msg
}
