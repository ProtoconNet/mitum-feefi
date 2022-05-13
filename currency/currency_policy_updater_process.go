package currency

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var currencyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CurrencyPolicyUpdaterProcessor)
	},
}

type CurrencyPolicyUpdaterProcessor struct {
	extensioncurrency.CurrencyPolicyUpdater
	cp        *extensioncurrency.CurrencyPool
	pubs      []key.Publickey
	threshold base.Threshold
	st        state.State
	de        extensioncurrency.CurrencyDesign
}

func NewCurrencyPolicyUpdaterProcessor(
	cp *extensioncurrency.CurrencyPool,
	pubs []key.Publickey,
	threshold base.Threshold,
) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(extensioncurrency.CurrencyPolicyUpdater)
		if !ok {
			return nil, errors.Errorf("not CurrencyPolicyUpdater, %T", op)
		}

		opp := currencyUpdaterProcessorPool.Get().(*CurrencyPolicyUpdaterProcessor)

		opp.cp = cp
		opp.CurrencyPolicyUpdater = i
		opp.pubs = pubs
		opp.threshold = threshold

		return opp, nil
	}
}

func (opp *CurrencyPolicyUpdaterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	if len(opp.pubs) < 1 {
		return nil, operation.NewBaseReasonError("empty publickeys for operation signs")
	} else if err := checkFactSignsByPubs(opp.pubs, opp.threshold, opp.Signs()); err != nil {
		return nil, err
	}

	fact := opp.Fact().(extensioncurrency.CurrencyPolicyUpdaterFact)

	if opp.cp != nil {
		i, found := opp.cp.State(fact.Currency())
		if !found {
			return nil, operation.NewBaseReasonError("unknown currency, %q found", fact.Currency())
		}
		opp.st = i
		opp.de, _ = opp.cp.Get(fact.Currency())
	}

	receiver := fact.Policy().Feeer().Receiver()
	if receiver != nil {
		if err := checkExistsState(currency.StateKeyAccount(receiver), getState); err != nil {
			return nil, errors.Wrap(err, "feeer receiver account not found")
		}
	}

	// check whether fee receiver is contract account
	_, err := notExistsState(extensioncurrency.StateKeyContractAccount(receiver), "contract account status", getState)
	if err != nil {
		return nil, errors.Wrap(err, "feeer receiver account is contract account")
	}

	f, ok := fact.Policy().Feeer().(feefi.FeefiFeeer)
	if ok {
		if err := checkExistsState(currency.StateKeyAccount(f.Feefier()), getState); err != nil {
			return nil, errors.Wrap(err, "feeer feefier account not found")
		}
		// check whether feeer feefier is contract account
		err := checkExistsState(extensioncurrency.StateKeyContractAccount(f.Feefier()), getState)
		if err != nil {
			return nil, errors.Wrap(err, "feeer feefier account is not contract account")
		}
	}

	return opp, nil
}

func (opp *CurrencyPolicyUpdaterProcessor) Process(
	_ func(string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(extensioncurrency.CurrencyPolicyUpdaterFact)

	i, err := extensioncurrency.SetStateCurrencyDesignValue(opp.st, opp.de.SetPolicy(fact.Policy()))
	if err != nil {
		return err
	}
	return setState(fact.Hash(), i)
}

func (opp *CurrencyPolicyUpdaterProcessor) Close() error {
	opp.cp = nil
	opp.CurrencyPolicyUpdater = extensioncurrency.CurrencyPolicyUpdater{}
	opp.pubs = nil
	opp.threshold = base.Threshold{}

	currencyUpdaterProcessorPool.Put(opp)

	return nil
}
