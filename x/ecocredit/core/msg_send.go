package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	"github.com/regen-network/regen-ledger/types/math"
	"github.com/regen-network/regen-ledger/x/ecocredit"
)

var _ legacytx.LegacyMsg = &MsgSend{}

// Route implements the LegacyMsg interface.
func (m MsgSend) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the LegacyMsg interface.
func (m MsgSend) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ecocredit.ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgSend) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("sender: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.Recipient); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("recipient: %s", err)
	}

	if len(m.Credits) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("credits cannot be empty")
	}

	for _, credits := range m.Credits {
		if err := ValidateBatchDenom(credits.BatchDenom); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}

		if credits.TradableAmount == "" && credits.RetiredAmount == "" {
			return sdkerrors.ErrInvalidRequest.Wrap("tradable amount or retired amount required")
		}

		if _, err := math.NewNonNegativeDecFromString(credits.TradableAmount); err != nil {
			return err
		}

		retiredAmount, err := math.NewNonNegativeDecFromString(credits.RetiredAmount)
		if err != nil {
			return err
		}

		if !retiredAmount.IsZero() {
			if err = ValidateJurisdiction(credits.RetirementJurisdiction); err != nil {
				return sdkerrors.ErrInvalidRequest.Wrapf("retirement jurisdiction: %s", err.Error())
			}
		}
	}

	return nil
}

// GetSigners returns the expected signers for MsgSend.
func (m *MsgSend) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
