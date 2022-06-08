package currency

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var keyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(KeyUpdaterProcessor)
	},
}

type KeyUpdaterProcessor struct {
	cp *extensioncurrency.CurrencyPool
	currency.KeyUpdater
	sa  state.State
	sb  currency.AmountState
	fee currency.Big
}

func NewKeyUpdaterProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(currency.KeyUpdater)
		if !ok {
			return nil, errors.Errorf("not KeyUpdater, %T", op)
		}

		opp := keyUpdaterProcessorPool.Get().(*KeyUpdaterProcessor)

		opp.cp = cp
		opp.KeyUpdater = i
		opp.sa = nil
		opp.sb = currency.AmountState{}
		opp.fee = currency.ZeroBig

		return opp, nil
	}
}

func (opp *KeyUpdaterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(currency.KeyUpdaterFact)

	st, err := existsState(currency.StateKeyAccount(fact.Target()), "target keys", getState)
	if err != nil {
		return nil, err
	}
	opp.sa = st

	if ks, e := currency.StateKeysValue(opp.sa); err != nil {
		return nil, operation.NewBaseReasonErrorFromError(e)
	} else if ks.Equal(fact.Keys()) {
		return nil, operation.NewBaseReasonError("same Keys with the existing")
	}

	st, err = existsState(currency.StateKeyBalance(fact.Target(), fact.Currency()), "balance of target", getState)
	if err != nil {
		return nil, err
	}
	opp.sb = currency.NewAmountState(st, fact.Currency())

	if err = checkFactSignsByState(fact.Target(), opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	f, found := opp.cp.Feeer(fact.Currency())
	if !found {
		return nil, operation.NewBaseReasonError("currency, %q not found of KeyUpdater", fact.Currency())
	}
	// 수정
	feeer, ok := f.(feefi.FeefiFeeer)
	cid := fact.Currency()
	id := extensioncurrency.ContractID(cid.String())
	var fee currency.Big
	if ok {
		st, err := existsState(feefi.StateKeyDesign(feeer.Feefier(), id), "feefi design", getState)
		if err != nil {
			return nil, operation.NewBaseReasonErrorFromError(err)
		}
		design, err := feefi.StateDesignValue(st)
		if err != nil {
			return nil, operation.NewBaseReasonErrorFromError(err)
		}
		if design.Fee().Currency() != cid {
			return nil, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Fee().Currency(), cid)
		}
		fee = design.Fee().Big()
	} else {
		var err error
		switch fee, err = f.Fee(currency.ZeroBig); {
		case err != nil:
			return nil, operation.NewBaseReasonErrorFromError(err)
		}
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

func (opp *KeyUpdaterProcessor) Process(
	_ func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(currency.KeyUpdaterFact)

	opp.sb = opp.sb.Sub(opp.fee).AddFee(opp.fee)
	st, err := currency.SetStateKeysValue(opp.sa, fact.Keys())
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}
	return setState(fact.Hash(), st, opp.sb)
}

func (opp *KeyUpdaterProcessor) Close() error {
	opp.cp = nil
	opp.KeyUpdater = currency.KeyUpdater{}
	opp.sa = nil
	opp.sb = currency.AmountState{}
	opp.fee = currency.ZeroBig

	keyUpdaterProcessorPool.Put(opp)

	return nil
}
