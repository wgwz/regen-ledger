package tests

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/regen-network/regen-ledger/types/module/server"
	ecocredittypes "github.com/regen-network/regen-ledger/x/ecocredit"
	"github.com/regen-network/regen-ledger/x/ecocredit/basket"
	ecocredit "github.com/regen-network/regen-ledger/x/ecocredit/module"
)

func NewEcocreditModule(ff *server.FixtureFactory) *ecocredit.Module {
	baseApp := ff.BaseApp()
	cdc := ff.Codec()
	amino := codec.NewLegacyAmino()

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	params.RegisterInterfaces(cdc.InterfaceRegistry())

	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	distKey := sdk.NewKVStoreKey(disttypes.StoreKey)
	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	baseApp.MountStore(authKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(bankKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(distKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(paramsKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(tkey, sdk.StoreTypeTransient)

	authSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, authtypes.ModuleName)
	bankSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, banktypes.ModuleName)
	ecocreditSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, ecocredittypes.ModuleName)

	maccPerms := map[string][]string{
		minttypes.ModuleName:       {authtypes.Minter},
		ecocredittypes.ModuleName:  {authtypes.Burner},
		basket.BasketSubModuleName: {authtypes.Burner, authtypes.Minter},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc, authKey, authSubspace, authtypes.ProtoBaseAccount, maccPerms,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc, bankKey, accountKeeper, bankSubspace, nil,
	)

	return ecocredit.NewModule(ecocreditSubspace, accountKeeper, bankKeeper)
}
