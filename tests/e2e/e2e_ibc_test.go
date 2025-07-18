package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ForwardMetadata struct {
	Receiver string `json:"receiver"`
	Port     string `json:"port"`
	Channel  string `json:"channel"`
	// Timeout        time.Duration `json:"timeout"`
	// Retries        *uint8        `json:"retries,omitempty"`
	// Next           *string       `json:"next,omitempty"`
	// RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

//nolint:unparam
func (s *IntegrationTestSuite) sendIBC(c *chain, valIdx int, sender, recipient, token, fees, note string, expErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		kiichaindBinary,
		txCommand,
		"ibc-transfer",
		"transfer",
		transferPort,
		transferChannel,
		recipient,
		token,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, s.chainA.id, sender, s.chainB.id, recipient, note)
	if expErr {
		s.executeKiichainTxCommand(ctx, c, ibcCmd, valIdx, s.expectErrExecValidation(c, valIdx, true))
		s.T().Log("unsuccessfully sent IBC tokens")
	} else {
		s.executeKiichainTxCommand(ctx, c, ibcCmd, valIdx, s.defaultExecValidation(c, valIdx))
		s.T().Log("successfully sent IBC tokens")
	}
}

func (s *IntegrationTestSuite) hermesClearPacket(configPath, chainID, portID, channelID string) (success bool) { //nolint:unparam
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		fmt.Sprintf("--config=%s", configPath),
		"clear",
		"packets",
		fmt.Sprintf("--chain=%s", chainID),
		fmt.Sprintf("--channel=%s", channelID),
		fmt.Sprintf("--port=%s", portID),
	}

	if _, err := s.executeHermesCommand(ctx, hermesCmd); err != nil {
		s.T().Logf("failed to clear packets: %s", err)
		return false
	}

	return true
}

type RelayerPacketsOutput struct {
	Result struct {
		Dst struct {
			UnreceivedPackets []uint64 `json:"unreceived_packets"`
		} `json:"dst"`
		Src struct {
			UnreceivedPackets []uint64 `json:"unreceived_packets"`
		} `json:"src"`
	} `json:"result"`
	Status string `json:"status"`
}

func (s *IntegrationTestSuite) createConnection() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		"create",
		"connection",
		"--a-chain",
		s.chainA.id,
		"--b-chain",
		s.chainB.id,
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to connect chains: %s", err)

	s.T().Logf("connected %s and %s chains via IBC", s.chainA.id, s.chainB.id)
}

func (s *IntegrationTestSuite) createChannel() {
	s.T().Logf("creating IBC transfer channel created between chains %s and %s", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	hermesCmd := []string{
		hermesBinary,
		"--json",
		"create",
		"channel",
		"--a-chain", s.chainA.id,
		"--a-connection", "connection-0",
		"--a-port", "transfer",
		"--b-port", "transfer",
		"--channel-version", "ics20-1",
		"--order", "unordered",
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to create IBC transfer channel between chains: %s", err)

	s.T().Logf("IBC transfer channel created between chains %s and %s", s.chainA.id, s.chainB.id)
}

// This function will complete the channel handshake in cases when ChanOpenInit was initiated
// by some transaction that was previously executed on the chain. For example,
// ICA MsgRegisterInterchainAccount will perform ChanOpenInit during its execution.
func (s *IntegrationTestSuite) completeChannelHandshakeFromTry(
	srcChain, dstChain,
	srcConnection, dstConnection,
	srcPort, dstPort,
	srcChannel, dstChannel string,
) {
	s.T().Logf("completing IBC channel handshake between: (%s, %s, %s, %s) and (%s, %s, %s, %s)",
		srcChain, srcConnection, srcPort, srcChannel,
		dstChain, dstConnection, dstPort, dstChannel)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	hermesCmd := []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-try",
		"--dst-chain", dstChain,
		"--src-chain", srcChain,
		"--dst-connection", dstConnection,
		"--dst-port", dstPort,
		"--src-port", srcPort,
		"--src-channel", srcChannel,
	}

	_, err := s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-try: %s", err)

	hermesCmd = []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-ack",
		"--dst-chain", srcChain,
		"--src-chain", dstChain,
		"--dst-connection", srcConnection,
		"--dst-port", srcPort,
		"--src-port", dstPort,
		"--dst-channel", srcChannel,
		"--src-channel", dstChannel,
	}

	_, err = s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-ack: %s", err)

	hermesCmd = []string{
		hermesBinary,
		"--json",
		"tx",
		"chan-open-confirm",
		"--dst-chain", dstChain,
		"--src-chain", srcChain,
		"--dst-connection", dstConnection,
		"--dst-port", dstPort,
		"--src-port", srcPort,
		"--dst-channel", dstChannel,
		"--src-channel", srcChannel,
	}

	_, err = s.executeHermesCommand(ctx, hermesCmd)
	s.Require().NoError(err, "failed to execute chan-open-confirm: %s", err)

	s.T().Logf("IBC channel handshake completed between: (%s, %s, %s, %s) and (%s, %s, %s, %s)",
		srcChain, srcConnection, srcPort, srcChannel,
		dstChain, dstConnection, dstPort, dstChannel)
}

func (s *IntegrationTestSuite) testIBCTokenTransfer() {
	s.Run("send_akii_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances      sdk.Coins
			err           error
			beforeBalance int64
			ibcStakeDenom string
		)

		address, _ := s.chainA.validators[0].keyInfo.GetAddress()
		sender := address.String()

		address, _ = s.chainB.validators[0].keyInfo.GetAddress()
		recipient := address.String()

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				beforeBalance = c.Amount.Int64()
				break
			}
		}

		tokenAmt := mustNewIntFromString("3300000000000000000000") // 3,300 Kii
		s.sendIBC(s.chainA, 0, sender, recipient, tokenAmt.String()+akiiDenom, standardFees.String(), "", false)

		pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((tokenAmt.Add(math.NewInt(beforeBalance))), c.Amount)
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

/*
TestMultihopIBCTokenTransfer tests that sending an IBC transfer using the IBC Packet Forward Middleware accepts a port, channel and account address

Steps:
1. Check balance of Account 1 on Chain 1
2. Check balance of Account 2 on Chain 1
3. Account 1 on Chain 1 sends x tokens to Account 2 on Chain 1 via Account 1 on Chain 2
4. Check Balance of Account 1 on Chain 1, confirm it is original minus x tokens
5. Check Balance of Account 2 on Chain 1, confirm it is original plus x tokens

*/
// TODO: Add back only if packet forward middleware has a working version compatible with IBC v3.0.x
func (s *IntegrationTestSuite) testMultihopIBCTokenTransfer() {
	time.Sleep(30 * time.Second)

	s.Run("send_successful_multihop_akii_to_chainA_from_chainA", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			err error
		)

		address, _ := s.chainA.validators[0].keyInfo.GetAddress()
		sender := address.String()

		address, _ = s.chainB.validators[0].keyInfo.GetAddress()
		middlehop := address.String()

		address, _ = s.chainA.validators[1].keyInfo.GetAddress()
		recipient := address.String()

		forwardPort := "transfer"
		forwardChannel := "channel-0"

		tokenAmt := mustNewIntFromString("3300000000000000000000") // 3,300 Kii

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderAKiiBalance    sdk.Coin
			beforeRecipientAKiiBalance sdk.Coin
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderAKiiBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, akiiDenom)
				s.Require().NoError(err)

				beforeRecipientAKiiBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, akiiDenom)
				s.Require().NoError(err)

				return beforeSenderAKiiBalance.IsValid() && beforeRecipientAKiiBalance.IsValid()
			},
			1*time.Minute,
			5*time.Second,
		)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: recipient,
				Channel:  forwardChannel,
				Port:     forwardPort,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		s.Require().NoError(err)

		s.sendIBC(s.chainA, 0, sender, middlehop, tokenAmt.String()+akiiDenom, standardFees.String(), string(memo), false)

		pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				afterSenderAKiiBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, akiiDenom)
				s.Require().NoError(err)

				afterRecipientAKiiBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, akiiDenom)
				s.Require().NoError(err)

				decremented := beforeSenderAKiiBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterSenderAKiiBalance)
				incremented := beforeRecipientAKiiBalance.Add(tokenAmount).IsEqual(afterRecipientAKiiBalance)

				return decremented && incremented
			},
			1*time.Minute,
			5*time.Second,
		)
	})
}

/*
TestFailedMultihopIBCTokenTransfer tests that sending a failing IBC transfer using the IBC Packet Forward
Middleware will send the tokens back to the original account after failing.
*/
func (s *IntegrationTestSuite) testFailedMultihopIBCTokenTransfer() {
	time.Sleep(30 * time.Second)

	s.Run("send_failed_multihop_akii_to_chainA_from_chainA", func() {
		address, _ := s.chainA.validators[0].keyInfo.GetAddress()
		sender := address.String()

		address, _ = s.chainB.validators[0].keyInfo.GetAddress()
		middlehop := address.String()

		address, _ = s.chainA.validators[1].keyInfo.GetAddress()
		recipient := strings.Replace(address.String(), "kii", "foobar", 1) // this should be an invalid recipient to force the tx to fail

		forwardPort := "transfer"
		forwardChannel := "channel-0"

		tokenAmt := tokenAmount.Amount // 3,300 Kii

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderAKiiBalance sdk.Coin
			err                     error
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderAKiiBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, akiiDenom)
				s.Require().NoError(err)

				return beforeSenderAKiiBalance.IsValid()
			},
			1*time.Minute,
			5*time.Second,
		)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: recipient,
				Channel:  forwardChannel,
				Port:     forwardPort,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		s.Require().NoError(err)

		s.sendIBC(s.chainA, 0, sender, middlehop, tokenAmt.String()+akiiDenom, standardFees.String(), string(memo), false)

		// Sender account should be initially decremented the full amount
		s.Require().Eventually(
			func() bool {
				afterSenderAKiiBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, akiiDenom)
				s.Require().NoError(err)

				returned := beforeSenderAKiiBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterSenderAKiiBalance)

				return returned
			},
			1*time.Minute,
			5*time.Second,
		)

		// since the forward receiving account is invalid, it should be refunded to the original sender (minus the original fee)
		s.Require().Eventually(
			func() bool {
				pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
				s.Require().True(pass)

				afterSenderAKiiBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, akiiDenom)
				s.Require().NoError(err)
				returned := beforeSenderAKiiBalance.Sub(standardFees).IsEqual(afterSenderAKiiBalance)
				return returned
			},
			1*time.Minute,
			10*time.Second,
		)
	})
}
