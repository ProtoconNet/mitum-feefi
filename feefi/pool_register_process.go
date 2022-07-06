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
	cp *CurrencyPool
	PoolRegister
	cs  state.State                   // contract account status state
	ds  state.State                   // feefi design state
	ps  state.State                   // feefi pool state
	sb  currency.AmountState          // sender amount state
	ib  extensioncurrency.AmountState // target incoming amount state
	ob  extensioncurrency.AmountState // target outgoing amount state
	fee currency.Big
	as  extensioncurrency.ContractAccount // contract account status value
	pl  Pool                              // feefi pool value
	dg  PoolDesign                        // feefi design value
}

func NewPoolRegisterProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		pr, ok := op.(PoolRegister)
		if !ok {
			return nil, errors.Errorf("not PoolRegister, %T", op)
		}

		opp := poolRegisterProcessorPool.Get().(*PoolRegisterProcessor)

		opp.cp = cp
		opp.PoolRegister = pr
		opp.cs = nil
		opp.ds = nil
		opp.ps = nil
		opp.sb = currency.AmountState{}
		opp.ib = extensioncurrency.AmountState{}
		opp.ob = extensioncurrency.AmountState{}
		opp.fee = currency.ZeroBig
		opp.as = extensioncurrency.ContractAccount{}
		opp.pl = Pool{}
		opp.dg = PoolDesign{}

		return opp, nil
	}
}

func (opp *PoolRegisterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(PoolRegisterFact)
	if opp.cp != nil {
		_, found := opp.cp.Policy(fact.incomeCID)
		if !found {
			return nil, operation.NewBaseReasonError("income currency not registered, %q", fact.incomeCID)
		}
		_, found = opp.cp.Policy(fact.outlayCID)
		if !found {
			return nil, operation.NewBaseReasonError("outlay currency not registered, %q", fact.outlayCID)
		}
	}
	// check existence of target account state
	// keep target account state
	_, err := existsState(currency.StateKeyAccount(fact.target), "target account", getState)
	if err != nil {
		return nil, err
	}

	// check existence of contract account status state
	// check sender matched with contract account owner
	// keep contract account status state
	// keep contract account status value
	st, err := existsState(extensioncurrency.StateKeyContractAccount(fact.target), "contract account status", getState)
	if err != nil {
		return nil, err
	}
	ca, err := extensioncurrency.StateContractAccountValue(st)
	if err != nil {
		return nil, err
	}
	if !ca.Owner().Equal(fact.sender) {
		return nil, operation.NewBaseReasonError(
			"contract account owner, %q is not matched with %q",
			ca.Owner(),
			fact.sender,
		)
	}

	opp.cs = st
	opp.as = ca

	// check sender has amount of currency
	// keep amount state of sender
	st, err = existsState(currency.StateKeyBalance(fact.sender, fact.currency), "balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.sb = currency.NewAmountState(st, fact.currency)

	// check target don't have pool state
	// keep pool state of target
	id := extensioncurrency.ContractID(fact.IncomeCID())
	st, err = notExistsState(StateKeyPool(fact.target, id), "pool of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ps = st

	// prepare new pool
	// keep pool value
	opp.pl = NewPool(fact.IncomeCID(), fact.OutlayCID())

	// check target don't have design state
	// keep design state of target
	st, err = notExistsState(StateKeyPoolDesign(fact.target, id), "design of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ds = st
	opp.dg = NewPoolDesign(fact.InitialFee(), fact.target)

	// check target don't have incoming amount state
	// keep incoming amount state of target
	st, err = notExistsState(extensioncurrency.StateKeyBalance(
		fact.target,
		id,
		fact.IncomeCID(),
		StateKeyBalanceSuffix,
	), "incoming balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.ib = extensioncurrency.NewAmountState(st, fact.IncomeCID(), id)

	// check target don't have outgoing amount state
	// keep outgoing amount state of target
	st, err = notExistsState(extensioncurrency.StateKeyBalance(
		fact.target,
		id,
		fact.OutlayCID(),
		StateKeyBalanceSuffix),
		"outgoing balance of target",
		getState,
	)
	if err != nil {
		return nil, err
	}
	opp.ob = extensioncurrency.NewAmountState(st, fact.OutlayCID(), id)

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
	dst, err := SetStatePoolDesignValue(opp.ds, opp.dg)
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	id := extensioncurrency.ContractID(fact.incomeCID.String())

	ibst, err := extensioncurrency.SetStateBalanceValue(opp.ib, extensioncurrency.NewAmountValuefromAmount(
		currency.NewZeroAmount(fact.incomeCID),
		id,
	))
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	obst, err := extensioncurrency.SetStateBalanceValue(opp.ob, extensioncurrency.NewAmountValuefromAmount(
		currency.NewZeroAmount(fact.outlayCID),
		id,
	))
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	return setState(fact.Hash(), pst, dst, ibst, obst, opp.sb)
}

func (opp *PoolRegisterProcessor) Close() error {
	opp.cp = nil
	opp.PoolRegister = PoolRegister{}
	opp.cs = nil
	opp.ds = nil
	opp.ps = nil
	opp.sb = currency.AmountState{}
	opp.ib = extensioncurrency.AmountState{}
	opp.ob = extensioncurrency.AmountState{}
	opp.fee = currency.ZeroBig
	opp.as = extensioncurrency.ContractAccount{}
	opp.pl = Pool{}
	opp.dg = PoolDesign{}

	poolRegisterProcessorPool.Put(opp)

	return nil
}
