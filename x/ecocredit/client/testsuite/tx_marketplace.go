package testsuite

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/regen-network/regen-ledger/types/testutil/cli"
	marketplaceclient "github.com/regen-network/regen-ledger/x/ecocredit/client/marketplace"
	"github.com/regen-network/regen-ledger/x/ecocredit/marketplace"
)

func (s *IntegrationTestSuite) TestTxSell() {
	require := s.Require()

	seller := s.addr1.String()

	askPrice := sdk.NewInt64Coin(s.allowedDenoms[0], 10)

	// using json package because array is not a proto message
	bz, err := json.Marshal([]marketplace.MsgSell_Order{
		{
			BatchDenom: s.batchDenom,
			Quantity:   "10",
			AskPrice:   &askPrice,
		},
		{
			BatchDenom: s.batchDenom,
			Quantity:   "10",
			AskPrice:   &askPrice,
		},
	})
	require.NoError(err)

	validJson := testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
	invalidJson := testutil.WriteToNewTempFile(s.T(), `{foo:bar}`).Name()
	duplicateJson := testutil.WriteToNewTempFile(s.T(), `{"foo":"bar","foo":"bar"`).Name()

	testCases := []struct {
		name      string
		args      []string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing args",
			args:      []string{},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 0",
		},
		{
			name:      "too many args",
			args:      []string{"foo", "bar"},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 2",
		},
		{
			name: "missing from flag",
			args: []string{
				validJson,
			},
			expErr:    true,
			expErrMsg: "Error: required flag(s) \"from\" not set",
		},
		{
			name: "invalid json file",
			args: []string{
				"foo.bar",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "no such file or directory",
		},
		{
			name: "invalid json format",
			args: []string{
				invalidJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "failed to parse json: invalid character",
		},
		{
			name: "duplicate json key",
			args: []string{
				duplicateJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "failed to parse json: duplicate key",
		},
		{
			name: "valid",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
		},
		{
			name: "valid from key-name",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Moniker),
			},
		},
		{
			name: "valid with amino-json",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := marketplaceclient.TxSellCmd()
			args := append(tc.args, s.commonTxFlags()...)
			out, err := cli.ExecTestCLICmd(s.val.ClientCtx, cmd, args)
			if tc.expErr {
				require.Error(err)
				require.Contains(out.String(), tc.expErrMsg)
			} else {
				require.NoError(err)

				var res sdk.TxResponse
				require.NoError(s.val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				require.Zero(res.Code, res.RawLog)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateSellOrders() {
	require := s.Require()

	seller := s.addr1.String()

	askPrice := sdk.NewInt64Coin(s.allowedDenoms[0], 10)

	// create new sell orders to not interfere with other tests
	sellOrderIds := s.createSellOrder(s.val.ClientCtx, &marketplace.MsgSell{
		Seller: s.addr1.String(),
		Orders: []*marketplace.MsgSell_Order{
			{
				BatchDenom: s.batchDenom,
				Quantity:   "10",
				AskPrice:   &askPrice,
			},
			{
				BatchDenom: s.batchDenom,
				Quantity:   "10",
				AskPrice:   &askPrice,
			},
		},
	})

	// using json package because array is not a proto message
	bz, err := json.Marshal([]marketplace.MsgUpdateSellOrders_Update{
		{
			SellOrderId: sellOrderIds[0],
			NewQuantity: "20",
			NewAskPrice: &askPrice,
		},
		{
			SellOrderId: sellOrderIds[1],
			NewQuantity: "20",
			NewAskPrice: &askPrice,
		},
	})
	require.NoError(err)

	validJson := testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
	invalidJson := testutil.WriteToNewTempFile(s.T(), `{foo:bar}`).Name()
	duplicateJson := testutil.WriteToNewTempFile(s.T(), `{"foo":"bar","foo":"bar"`).Name()

	testCases := []struct {
		name      string
		args      []string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing args",
			args:      []string{},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 0",
		},
		{
			name:      "too many args",
			args:      []string{"foo", "bar"},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 2",
		},
		{
			name:      "missing from flag",
			args:      []string{validJson},
			expErr:    true,
			expErrMsg: "Error: required flag(s) \"from\" not set",
		},
		{
			name: "invalid json file",
			args: []string{
				"foo.bar",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "no such file or directory",
		},
		{
			name: "invalid json format",
			args: []string{
				invalidJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "failed to parse json: invalid character",
		},
		{
			name: "duplicate json key",
			args: []string{
				duplicateJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
			expErr:    true,
			expErrMsg: "failed to parse json: duplicate key",
		},
		{
			name: "valid",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
		},
		{
			name: "valid from key-name",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Moniker),
			},
		},
		{
			name: "valid with amino-json",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := marketplaceclient.TxUpdateSellOrdersCmd()
			args := append(tc.args, s.commonTxFlags()...)
			out, err := cli.ExecTestCLICmd(s.val.ClientCtx, cmd, args)
			if tc.expErr {
				require.Error(err)
				require.Contains(out.String(), tc.expErrMsg)
			} else {
				require.NoError(err)

				var res sdk.TxResponse
				require.NoError(s.val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				require.Zero(res.Code, res.RawLog)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxBuyDirectCmd() {
	require := s.Require()

	buyer := s.addr2.String()

	sellOrderId := fmt.Sprint(s.sellOrderId)
	bidPrice := sdk.NewInt64Coin(s.allowedDenoms[0], 10).String()

	testCases := []struct {
		name      string
		args      []string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing args",
			args:      []string{"foo", "bar", "baz"},
			expErr:    true,
			expErrMsg: "Error: accepts 4 arg(s), received 3",
		},
		{
			name:      "too many args",
			args:      []string{"foo", "bar", "baz", "bar", "foo"},
			expErr:    true,
			expErrMsg: "Error: accepts 4 arg(s), received 5",
		},
		{
			name: "missing from flag",
			args: []string{
				sellOrderId,
				"10",
				bidPrice,
				"true",
			},
			expErr:    true,
			expErrMsg: "Error: required flag(s) \"from\" not set",
		},
		{
			name: "valid",
			args: []string{
				sellOrderId,
				"10",
				bidPrice,
				"true",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
			},
		},
		{
			name: "valid from key-name",
			args: []string{
				sellOrderId,
				"10",
				bidPrice,
				"true",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "addr2"),
			},
		},
		{
			name: "valid with amino-json",
			args: []string{
				sellOrderId,
				"10",
				bidPrice,
				"true",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := marketplaceclient.TxBuyDirectCmd()
			args := append(tc.args, s.commonTxFlags()...)
			out, err := cli.ExecTestCLICmd(s.val.ClientCtx, cmd, args)
			if tc.expErr {
				require.Error(err)
				require.Contains(out.String(), tc.expErrMsg)
			} else {
				require.NoError(err)

				var res sdk.TxResponse
				require.NoError(s.val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				require.Zero(res.Code, res.RawLog)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxBuyDirectBatchCmd() {
	require := s.Require()

	buyer := s.addr2.String()

	bidPrice := sdk.NewInt64Coin(s.allowedDenoms[0], 10)

	// using json package because array is not a proto message
	bz, err := json.Marshal([]marketplace.MsgBuyDirect_Order{
		{
			SellOrderId:       s.sellOrderId,
			Quantity:          "10",
			BidPrice:          &bidPrice,
			DisableAutoRetire: true,
		},
		{
			SellOrderId:            s.sellOrderId,
			Quantity:               "10",
			BidPrice:               &bidPrice,
			RetirementJurisdiction: "US-WA",
		},
	})
	require.NoError(err)

	validJson := testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
	invalidJson := testutil.WriteToNewTempFile(s.T(), `{foo:bar}`).Name()
	duplicateJson := testutil.WriteToNewTempFile(s.T(), `{"foo":"bar","foo":"bar"`).Name()

	testCases := []struct {
		name      string
		args      []string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing args",
			args:      []string{},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 0",
		},
		{
			name:      "too many args",
			args:      []string{"foo", "bar"},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 2",
		},
		{
			name:      "missing from flag",
			args:      []string{validJson},
			expErr:    true,
			expErrMsg: "Error: required flag(s) \"from\" not set",
		},
		{
			name: "invalid json file",
			args: []string{
				"foo.bar",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
			},
			expErr:    true,
			expErrMsg: "no such file or directory",
		},
		{
			name: "invalid json format",
			args: []string{
				invalidJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
			},
			expErr:    true,
			expErrMsg: "invalid character",
		},
		{
			name: "duplicate json key",
			args: []string{
				duplicateJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
			},
			expErr:    true,
			expErrMsg: "failed to parse json: duplicate key",
		},
		{
			name: "valid",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
			},
		},
		{
			name: "valid from key-name",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "addr2"),
			},
		},
		{
			name: "valid with amino-json",
			args: []string{
				validJson,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, buyer),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := marketplaceclient.TxBuyDirectBulkCmd()
			args := append(tc.args, s.commonTxFlags()...)
			out, err := cli.ExecTestCLICmd(s.val.ClientCtx, cmd, args)
			if tc.expErr {
				require.Error(err)
				require.Contains(out.String(), tc.expErrMsg)
			} else {
				require.NoError(err)

				var res sdk.TxResponse
				require.NoError(s.val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				require.Zero(res.Code, res.RawLog)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxCancelSellOrder() {
	require := s.Require()

	seller := s.addr1.String()

	testCases := []struct {
		name      string
		args      []string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing args",
			args:      []string{},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 0",
		},
		{
			name:      "too many args",
			args:      []string{"foo", "bar"},
			expErr:    true,
			expErrMsg: "Error: accepts 1 arg(s), received 2",
		},
		{
			name: "missing from flag",
			args: []string{
				fmt.Sprintf("%d", s.sellOrderId),
			},
			expErr:    true,
			expErrMsg: "Error: required flag(s) \"from\" not set",
		},
		{
			name: "valid",
			args: []string{
				fmt.Sprintf("%d", s.sellOrderId),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, seller),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := marketplaceclient.TxCancelSellOrderCmd()
			args := append(tc.args, s.commonTxFlags()...)
			out, err := cli.ExecTestCLICmd(s.val.ClientCtx, cmd, args)
			if tc.expErr {
				require.Error(err)
				require.Contains(out.String(), tc.expErrMsg)
			} else {
				require.NoError(err)

				var res sdk.TxResponse
				require.NoError(s.val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				require.Zero(res.Code, res.RawLog)
			}
		})
	}
}
