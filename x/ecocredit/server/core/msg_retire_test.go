package core

import (
	"strconv"
	"testing"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	api "github.com/regen-network/regen-ledger/api/regen/ecocredit/v1"
	"github.com/regen-network/regen-ledger/x/ecocredit/core"
)

type retire struct {
	*baseSuite
	alice            sdk.AccAddress
	creditTypeAbbrev string
	classId          string
	classKey         uint64
	projectId        string
	projectKey       uint64
	batchDenom       string
	batchKey         uint64
	tradableAmount   string
	res              *core.MsgRetireResponse
	err              error
}

func TestRetire(t *testing.T) {
	gocuke.NewRunner(t, &retire{}).Path("./features/msg_retire.feature").Run()
}

func (s *retire) Before(t gocuke.TestingT) {
	s.baseSuite = setupBase(t)
	s.alice = s.addr
	s.creditTypeAbbrev = "C"
	s.classId = "C01"
	s.projectId = "C01-001"
	s.batchDenom = "C01-001-20200101-20210101-001"
	s.tradableAmount = "10"
}

func (s *retire) ACreditTypeWithAbbreviationAndPrecision(a, b string) {
	precision, err := strconv.ParseUint(b, 10, 32)
	require.NoError(s.t, err)

	err = s.stateStore.CreditTypeTable().Insert(s.ctx, &api.CreditType{
		Abbreviation: a,
		Precision:    uint32(precision),
	})
	require.NoError(s.t, err)

	s.creditTypeAbbrev = a
}

func (s *retire) ACreditBatch() {
	s.creditBatchSetup()
}

func (s *retire) ACreditBatchWithDenom(a string) {
	s.projectSetup()

	bKey, err := s.k.stateStore.BatchTable().InsertReturningID(s.ctx, &api.Batch{
		ProjectKey: s.projectKey,
		Denom:      a,
	})
	require.NoError(s.t, err)

	err = s.k.stateStore.BatchSupplyTable().Insert(s.ctx, &api.BatchSupply{
		BatchKey:        bKey,
		TradableAmount:  s.tradableAmount,
		RetiredAmount:   "0",
		CancelledAmount: "0",
	})
	require.NoError(s.t, err)

	s.batchKey = bKey
}

func (s *retire) ACreditBatchFromCreditClassWithCreditType(a string) {
	cKey, err := s.k.stateStore.ClassTable().InsertReturningID(s.ctx, &api.Class{
		Id:               s.classId,
		CreditTypeAbbrev: a,
	})
	require.NoError(s.t, err)

	s.classKey = cKey

	pKey, err := s.k.stateStore.ProjectTable().InsertReturningID(s.ctx, &api.Project{
		Id:       s.projectId,
		ClassKey: cKey,
	})
	require.NoError(s.t, err)

	s.projectKey = pKey

	bKey, err := s.k.stateStore.BatchTable().InsertReturningID(s.ctx, &api.Batch{
		ProjectKey: s.projectKey,
		Denom:      s.batchDenom,
	})
	require.NoError(s.t, err)

	err = s.k.stateStore.BatchSupplyTable().Insert(s.ctx, &api.BatchSupply{
		BatchKey:        bKey,
		TradableAmount:  s.tradableAmount,
		RetiredAmount:   "0",
		CancelledAmount: "0",
	})
	require.NoError(s.t, err)

	s.batchKey = bKey
}

func (s *retire) AliceHasTheBatchBalance(a gocuke.DocString) {
	balance := &api.BatchBalance{}
	err := jsonpb.UnmarshalString(a.Content, balance)
	require.NoError(s.t, err)

	balance.BatchKey = s.batchKey
	balance.Address = s.alice

	// Save because the balance may already exist from setup
	err = s.stateStore.BatchBalanceTable().Save(s.ctx, balance)
	require.NoError(s.t, err)
}

func (s *retire) AliceOwnsTradableCreditAmount(a string) {
	err := s.k.stateStore.BatchBalanceTable().Insert(s.ctx, &api.BatchBalance{
		BatchKey:       s.batchKey,
		Address:        s.alice,
		TradableAmount: a,
	})
	require.NoError(s.t, err)
}

func (s *retire) AliceOwnsTradableCreditsWithBatchDenom(a string) {
	batch, err := s.k.stateStore.BatchTable().GetByDenom(s.ctx, a)
	require.NoError(s.t, err)

	err = s.k.stateStore.BatchBalanceTable().Insert(s.ctx, &api.BatchBalance{
		BatchKey:       batch.Key,
		Address:        s.alice,
		TradableAmount: s.tradableAmount,
	})
	require.NoError(s.t, err)
}

func (s *retire) TheBatchSupply(a gocuke.DocString) {
	supply := &api.BatchSupply{}
	err := jsonpb.UnmarshalString(a.Content, supply)
	require.NoError(s.t, err)

	supply.BatchKey = s.batchKey

	// Save because the supply may already exist from setup
	err = s.stateStore.BatchSupplyTable().Save(s.ctx, supply)
	require.NoError(s.t, err)
}

func (s *retire) AliceAttemptsToRetireCreditAmount(a string) {
	s.res, s.err = s.k.Retire(s.ctx, &core.MsgRetire{
		Owner: s.alice.String(),
		Credits: []*core.Credits{
			{
				BatchDenom: s.batchDenom,
				Amount:     a,
			},
		},
	})
}

func (s *retire) AliceAttemptsToRetireCreditsWithBatchDenom(a string) {
	s.res, s.err = s.k.Retire(s.ctx, &core.MsgRetire{
		Owner: s.alice.String(),
		Credits: []*core.Credits{
			{
				BatchDenom: a,
				Amount:     s.tradableAmount,
			},
		},
	})
}

func (s *retire) ExpectNoError() {
	require.NoError(s.t, s.err)
}

func (s *retire) ExpectTheError(a string) {
	require.EqualError(s.t, s.err, a)
}

func (s *retire) ExpectAliceBatchBalance(a gocuke.DocString) {
	expected := &api.BatchBalance{}
	err := jsonpb.UnmarshalString(a.Content, expected)
	require.NoError(s.t, err)

	balance, err := s.stateStore.BatchBalanceTable().Get(s.ctx, s.alice, s.batchKey)
	require.NoError(s.t, err)

	require.Equal(s.t, expected.RetiredAmount, balance.RetiredAmount)
	require.Equal(s.t, expected.TradableAmount, balance.TradableAmount)
	require.Equal(s.t, expected.EscrowedAmount, balance.EscrowedAmount)
}

func (s *retire) ExpectBatchSupply(a gocuke.DocString) {
	expected := &api.BatchSupply{}
	err := jsonpb.UnmarshalString(a.Content, expected)
	require.NoError(s.t, err)

	balance, err := s.stateStore.BatchSupplyTable().Get(s.ctx, s.batchKey)
	require.NoError(s.t, err)

	require.Equal(s.t, expected.RetiredAmount, balance.RetiredAmount)
	require.Equal(s.t, expected.TradableAmount, balance.TradableAmount)
}

func (s *retire) projectSetup() {
	err := s.k.stateStore.CreditTypeTable().Insert(s.ctx, &api.CreditType{
		Abbreviation: s.creditTypeAbbrev,
	})
	require.NoError(s.t, err)

	cKey, err := s.k.stateStore.ClassTable().InsertReturningID(s.ctx, &api.Class{
		Id:               s.classId,
		CreditTypeAbbrev: s.creditTypeAbbrev,
	})
	require.NoError(s.t, err)

	s.classKey = cKey

	pKey, err := s.k.stateStore.ProjectTable().InsertReturningID(s.ctx, &api.Project{
		Id:       s.projectId,
		ClassKey: cKey,
	})
	require.NoError(s.t, err)

	s.projectKey = pKey
}

func (s *retire) creditBatchSetup() {
	s.projectSetup()

	bKey, err := s.k.stateStore.BatchTable().InsertReturningID(s.ctx, &api.Batch{
		ProjectKey: s.projectKey,
		Denom:      s.batchDenom,
	})
	require.NoError(s.t, err)

	err = s.k.stateStore.BatchSupplyTable().Insert(s.ctx, &api.BatchSupply{
		BatchKey:        bKey,
		TradableAmount:  s.tradableAmount,
		RetiredAmount:   "0",
		CancelledAmount: "0",
	})
	require.NoError(s.t, err)

	s.batchKey = bKey
}
