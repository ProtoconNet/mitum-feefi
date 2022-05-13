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

var poolRegisterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(PoolRegisterProcessor)
	},
}

func (PoolRegister) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	return nil
}

type PoolRegisterProcessor struct {
	cp *extensioncurrency.CurrencyPool
	PoolRegister
	cs  state.State          // contract account status state
	ds  state.State          // feefi design state
	ps  state.State          // feefi pool state
	sb  currency.AmountState // sender amount state
	ib  currency.AmountState // target incoming amount state
	ob  currency.AmountState // target outgoing amount state
	fee currency.Big
	as  extensioncurrency.ContractAccount // contract account status value
	pl  Pool                              // feefi pool value
	dg  Design                            // feefi design value
}

func NewPoolRegisterProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(PoolRegister)
		if !ok {
			return nil, errors.Errorf("not ConfigContractAccount, %T", op)
		}

		opp := poolRegisterProcessorPool.Get().(*PoolRegisterProcessor)

		opp.cp = cp
		opp.PoolRegister = i
		opp.cs = nil
		opp.ds = nil
		opp.ps = nil
		opp.sb = currency.AmountState{}
		opp.ib = currency.AmountState{}
		opp.ob = currency.AmountState{}
		opp.fee = currency.ZeroBig
		opp.as = extensioncurrency.ContractAccount{}
		opp.pl = Pool{}
		opp.dg = Design{}

		return opp, nil
	}
}

func (opp *PoolRegisterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(PoolRegisterFact)

	// check existence of target account state
	// keep target account state
	st, err := existsState(currency.StateKeyAccount(fact.target), "target account", getState)
	if err != nil {
		return nil, err
	}

	// check existence of contract account status state
	// check sender matched with contract account owner
	// keep contract account status state
	// keep contract account status value
	st, err = existsState(extensioncurrency.StateKeyContractAccount(fact.target), "contract account status", getState)
	if err != nil {
		return nil, err
	}
	v, err := extensioncurrency.StateContractAccountValue(st)
	if err != nil {
		return nil, err
	}
	if !v.Owner().Equal(fact.sender) {
		return nil, errors.Errorf("contract account owner, %q is not matched with %q", v.Owner(), fact.sender)
	}

	opp.cs = st
	opp.as = v

	// check sender has amount of currency
	// keep amount state of sender
	st, err = existsState(currency.StateKeyBalance(fact.sender, fact.currency), "balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.sb = currency.NewAmountState(st, fact.currency)

	// check target don't have pool state
	// keep pool state of target
	st, err = notExistsState(StateKeyPool(fact.target, fact.IncomeCID()), "pool of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ps = st

	// prepare new pool
	// keep pool value
	opp.pl = NewPool(fact.IncomeCID(), fact.OutlayCID())

	// check target don't have design state
	// keep design state of target
	st, err = notExistsState(StateKeyDesign(fact.target, fact.IncomeCID()), "design of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ds = st
	opp.dg = NewDesign(fact.InitialFee(), fact.target)

	// check target don't have incoming amount state
	// keep incoming amount state of target
	st, err = notExistsState(stateKeyBalance(fact.target, fact.IncomeCID(), fact.IncomeCID()), "incoming balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ib = currency.NewAmountState(st, fact.IncomeCID())

	// check target don't have outgoing amount state
	// keep outgoing amount state of target
	st, err = notExistsState(stateKeyBalance(fact.target, fact.IncomeCID(), fact.OutlayCID()), "outgoing balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ob = currency.NewAmountState(st, fact.OutlayCID())

	// check fact sign
	if err = checkFactSignsByState(fact.sender, opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	// check feeer
	// TODO : feefi fee check
	feeer, found := opp.cp.Feeer(fact.currency)
	if !found {
		return nil, operation.NewBaseReasonError("currency, %q not found of PoolRegister", fact.currency)
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

func (opp *PoolRegisterProcessor) Process(
	_ func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(PoolRegisterFact)

	opp.sb = opp.sb.Sub(opp.fee).AddFee(opp.fee)
	pst, err := setStatePoolValue(opp.ps, opp.pl)
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	dst, err := setStateDesignValue(opp.ds, opp.dg)
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	ibst, err := setStateBalanceValue(opp.ib, currency.NewZeroAmount(fact.IncomeCID()))
	obst, err := setStateBalanceValue(opp.ob, currency.NewZeroAmount(fact.OutlayCID()))
	return setState(fact.Hash(), pst, dst, ibst, obst, opp.sb)
}

func (opp *PoolRegisterProcessor) Close() error {
	opp.cp = nil
	opp.PoolRegister = PoolRegister{}
	opp.cs = nil
	opp.ds = nil
	opp.ps = nil
	opp.sb = currency.AmountState{}
	opp.ib = currency.AmountState{}
	opp.ob = currency.AmountState{}
	opp.fee = currency.ZeroBig
	opp.as = extensioncurrency.ContractAccount{}
	opp.pl = Pool{}
	opp.dg = Design{}

	poolRegisterProcessorPool.Put(opp)

	return nil
}
