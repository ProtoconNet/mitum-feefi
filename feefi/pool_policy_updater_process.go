package feefi

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var poolPolicyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(PoolPolicyUpdaterProcessor)
	},
}

func (PoolPolicyUpdater) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	return nil
}

type PoolPolicyUpdaterProcessor struct {
	cp *CurrencyPool
	PoolPolicyUpdater
	cs  state.State          // contract account status state
	ds  state.State          // feefi design state
	ps  state.State          // feefi pool state
	sb  currency.AmountState // sender amount state
	fee currency.Big
	dg  PoolDesign // feefi design value
}

func NewPoolPolicyUpdaterProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		oppu, ok := op.(PoolPolicyUpdater)
		if !ok {
			return nil, operation.NewBaseReasonError("not ConfigContractAccount, %T", op)
		}
		opp := poolPolicyUpdaterProcessorPool.Get().(*PoolPolicyUpdaterProcessor)
		opp.cp = cp
		opp.PoolPolicyUpdater = oppu
		opp.cs = nil
		opp.ds = nil
		opp.ps = nil
		opp.sb = currency.AmountState{}
		opp.fee = currency.ZeroBig
		opp.dg = PoolDesign{}

		return opp, nil
	}
}

func (opp *PoolPolicyUpdaterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(PoolPolicyUpdaterFact)

	// check existence of target account state
	// keep target account state
	st, err := existsState(currency.StateKeyAccount(fact.target), "target account", getState)
	if err != nil {
		return nil, err
	}
	// check existence of contract account status state
	// check sender matched with contract account owner
	// keep contract account state
	// keep contract account value
	st, err = existsState(extensioncurrency.StateKeyContractAccount(fact.target), "contract account", getState)
	if err != nil {
		return nil, err
	}
	ca, err := extensioncurrency.StateContractAccountValue(st)
	if err != nil {
		return nil, err
	}
	if !ca.Owner().Equal(fact.sender) {
		return nil, operation.NewBaseReasonError("contract account owner, %q is not matched with %q", ca.Owner(), fact.sender)
	}

	opp.cs = st

	// check sender has amount of currency
	// keep amount state of sender
	st, err = existsState(currency.StateKeyBalance(fact.sender, fact.currency), "balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.sb = currency.NewAmountState(st, fact.currency)

	// check target pool state exists
	// keep pool state of target
	id := fact.poolID
	st, err = existsState(StateKeyPool(fact.target, id), "pool of target", getState)
	if err != nil {
		return nil, err
	}

	// check target design state exists
	// keep design state of target
	st, err = existsState(StateKeyPoolDesign(fact.target, id), "design of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ds = st
	dg, err := StatePoolDesignValue(st)
	if err != nil {
		return nil, err
	}
	opp.dg = NewPoolDesign(fact.fee, dg.Address())

	// check fact sign
	if err = checkFactSignsByState(fact.sender, opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	// check feeer
	// TODO : feefi fee check
	feeer, found := opp.cp.Feeer(fact.currency)
	if !found {
		return nil, operation.NewBaseReasonError("currency, %q not found of PoolPolicyUpdater", fact.currency)
	}

	// get fee value
	// keep fee value
	fee, err := feeer.Fee(currency.ZeroBig)
	if err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	}
	switch b, err := currency.StateBalanceValue(opp.sb); {
	case err != nil:
		return nil, operation.NewBaseReasonErrorFromError(err)
	case b.Big().Compare(fee) < 0:
		return nil, operation.NewBaseReasonError("insufficient balance with fee")
	default:
		opp.fee = fee
	}

	return opp, nil
}

func (opp *PoolPolicyUpdaterProcessor) Process(
	_ func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(PoolPolicyUpdaterFact)

	opp.sb = opp.sb.Sub(opp.fee).AddFee(opp.fee)
	var (
		dst state.State
		err error
	)
	dst, err = SetStatePoolDesignValue(opp.ds, opp.dg)
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	return setState(fact.Hash(), dst, opp.sb)
}

func (opp *PoolPolicyUpdaterProcessor) Close() error {
	opp.cp = nil
	opp.PoolPolicyUpdater = PoolPolicyUpdater{}
	opp.cs = nil
	opp.ds = nil
	opp.ps = nil
	opp.sb = currency.AmountState{}
	opp.fee = currency.ZeroBig
	opp.dg = PoolDesign{}

	poolPolicyUpdaterProcessorPool.Put(opp)

	return nil
}
