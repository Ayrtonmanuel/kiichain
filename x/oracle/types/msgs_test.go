package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgAggregateExchangeRateVote(t *testing.T) {
	type test struct {
		voter         sdk.AccAddress
		exchangeRates string
		expectPass    bool
	}

	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1___________")),
	}

	invalidExchangeRates := "a,b"
	exchangeRates := "12.00atom,1234.12eth"
	abstainExchangeRates := "0.0atom,123.12eth"
	overFlowExchangeRates := "1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0atom,123.13eth"

	tests := []test{
		{addrs[0], exchangeRates, true},
		{addrs[0], invalidExchangeRates, false},
		{addrs[0], abstainExchangeRates, true},
		{addrs[0], overFlowExchangeRates, false},
		{sdk.AccAddress{}, exchangeRates, false},
	}

	// validation
	for i, test := range tests {
		msg := NewMsgAggregateExchangeRateVote(test.exchangeRates, test.voter, sdk.ValAddress(test.voter))
		if test.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)

			continue
		}

		require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
	}
}

func TestMsgDelegateFeedConsent(t *testing.T) {
	type test struct {
		validatorOwner sdk.AccAddress
		delegated      sdk.AccAddress
		expectPass     bool
	}

	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1___________")),
		sdk.AccAddress([]byte("addr2___________")),
	}

	tests := []test{
		{addrs[0], addrs[1], true},
		{sdk.AccAddress{}, addrs[1], false},
		{addrs[0], sdk.AccAddress{}, false},
		{addrs[0], addrs[0], true},
	}

	// validation
	for i, test := range tests {
		msg := NewMsgDelegateFeedConsent(test.validatorOwner, test.delegated)
		if test.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
			continue
		}

		require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
	}
}
