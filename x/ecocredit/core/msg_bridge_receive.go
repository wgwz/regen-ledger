package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	"github.com/regen-network/regen-ledger/types/eth"
	"github.com/regen-network/regen-ledger/types/math"
	"github.com/regen-network/regen-ledger/x/ecocredit"
)

var _ legacytx.LegacyMsg = &MsgBridgeReceive{}

// Route implements the LegacyMsg interface.
func (m MsgBridgeReceive) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the LegacyMsg interface.
func (m MsgBridgeReceive) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgBridgeReceive) GetSignBytes() []byte {
	return sdk.MustSortJSON(ecocredit.ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgBridgeReceive) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Issuer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("issuer: %s", err)
	}

	if err := ValidateClassId(m.ClassId); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// project validation

	if m.Project == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("project cannot be empty")
	}

	if m.Project.ReferenceId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("project reference id cannot be empty")
	}

	if len(m.Project.ReferenceId) > MaxReferenceIdLength {
		return ecocredit.ErrMaxLimit.Wrapf("project reference id: max length %d", MaxReferenceIdLength)
	}

	if err := ValidateJurisdiction(m.Project.Jurisdiction); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if m.Project.Metadata == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("project metadata cannot be empty")
	}

	if len(m.Project.Metadata) > MaxMetadataLength {
		return ecocredit.ErrMaxLimit.Wrapf("project metadata: max length %d", MaxMetadataLength)
	}

	// batch validation

	if m.Batch == nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("batch cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Batch.Recipient); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("batch recipient: %s", err)
	}

	if m.Batch.Amount == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("batch amount cannot be empty")
	}

	if _, err := math.NewPositiveDecFromString(m.Batch.Amount); err != nil {
		return sdkerrors.Wrap(err, "batch amount")
	}

	if m.Batch.StartDate == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("batch start date cannot be empty")
	}

	if m.Batch.EndDate == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("batch end date cannot be empty")
	}

	if m.Batch.StartDate.After(*m.Batch.EndDate) {
		return sdkerrors.ErrInvalidRequest.Wrap("batch start date cannot be after batch end date")
	}

	if m.Batch.Metadata == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("batch metadata cannot be empty")
	}

	if len(m.Batch.Metadata) > MaxMetadataLength {
		return ecocredit.ErrMaxLimit.Wrapf("batch metadata: max length %d", MaxMetadataLength)
	}

	// origin tx validation

	if m.OriginTx == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("origin tx cannot be empty")
	}

	// specific to MsgBridgeReceive
	if !eth.IsValidTxHash(m.OriginTx.Id) {
		return sdkerrors.ErrInvalidRequest.Wrap("origin tx id must be a valid ethereum transaction hash")
	}

	// specific to MsgBridgeReceive
	if m.OriginTx.Source != BridgePolygon {
		return sdkerrors.ErrInvalidRequest.Wrap("origin tx source must be polygon")
	}

	// specific to MsgBridgeReceive
	if m.OriginTx.Contract == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("origin tx contract cannot be empty")
	}

	// basic origin tx validation (includes valid ethereum contract address if contract is not empty)
	if err := m.OriginTx.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the expected signers for MsgCancel.
func (m *MsgBridgeReceive) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Issuer)
	return []sdk.AccAddress{addr}
}
