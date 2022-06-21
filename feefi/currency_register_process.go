package feefi

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var currencyRegisterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CurrencyRegisterProcessor)
	},
}

type CurrencyRegisterProcessor struct {
	extensioncurrency.CurrencyRegister
	cp        *CurrencyPool
	pubs      []key.Publickey
	threshold base.Threshold
	ga        currency.AmountState
	de        state.State
}

func NewCurrencyRegisterProcessor(
	cp *CurrencyPool, pubs []key.Publickey, threshold base.Threshold,
) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(extensioncurrency.CurrencyRegister)
		if !ok {
			return nil, errors.Errorf("not CurrencyRegister, %T", op)
		}

		opp := currencyRegisterProcessorPool.Get().(*CurrencyRegisterProcessor)

		opp.cp = cp
		opp.CurrencyRegister = i
		opp.pubs = pubs
		opp.threshold = threshold
		opp.ga = currency.AmountState{}
		opp.de = nil

		return opp, nil
	}
}

func (opp *CurrencyRegisterProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	if len(opp.pubs) < 1 {
		return nil, operation.NewBaseReasonError("empty publickeys for operation signs")
	} else if err := checkFactSignsByPubs(opp.pubs, opp.threshold, opp.Signs()); err != nil {
		return nil, err
	}

	item := opp.Fact().(extensioncurrency.CurrencyRegisterFact).Currency()

	if opp.cp != nil {
		if opp.cp.Exists(item.Currency()) {
			return nil, operation.NewBaseReasonError("currency already registered, %q", item.Currency())
		}
	}

	if err := checkExistsState(currency.StateKeyAccount(item.GenesisAccount()), getState); err != nil {
		return nil, errors.Wrap(err, "genesis account not found")
	}

	receiver := item.Policy().Feeer().Receiver()
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

	f, ok := item.Policy().Feeer().(FeefiFeeer)
	if ok {
		if err := checkExistsState(currency.StateKeyAccount(f.Feefier()), getState); err != nil {
			return nil, errors.Wrap(err, "feeer feefier account not found")
		}
		// check whether feeer feefier is contract account
		err := checkExistsState(extensioncurrency.StateKeyContractAccount(f.Feefier()), getState)
		if err != nil {
			return nil, errors.Wrap(err, "feeer feefier account is not contract account")
		}
		// check whether feeer feefier pool is registered
		err = checkExistsState(StateKeyPool(f.Feefier(), extensioncurrency.ContractID(item.Currency())), getState)
		if err != nil {
			return nil, errors.Wrap(err, "feeer feefier pool is not registered")
		}
	}

	switch st, found, err := getState(StateKeyCurrencyDesign(item.Currency())); {
	case err != nil:
		return nil, err
	case found:
		return nil, operation.NewBaseReasonError("currency already registered, %q", item.Currency())
	default:
		opp.de = st
	}

	switch st, found, err := getState(currency.StateKeyBalance(item.GenesisAccount(), item.Currency())); {
	case err != nil:
		return nil, err
	case found:
		return nil, operation.NewBaseReasonError("genesis account has already the currency, %q", item.Currency())
	default:
		opp.ga = currency.NewAmountState(st, item.Currency())
	}

	return opp, nil
}

func (opp *CurrencyRegisterProcessor) Process(
	getState func(string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(extensioncurrency.CurrencyRegisterFact)

	sts := make([]state.State, 4)

	sts[0] = opp.ga.Add(fact.Currency().Big())
	i, err := SetStateCurrencyDesignValue(opp.de, fact.Currency())
	if err != nil {
		return err
	}
	sts[1] = i

	{
		l, err := createZeroAccount(fact.Currency().Currency(), getState)
		if err != nil {
			return err
		}
		sts[2], sts[3] = l[0], l[1]
	}

	return setState(fact.Hash(), sts...)
}

func createZeroAccount(
	cid currency.CurrencyID,
	getState func(string) (state.State, bool, error),
) ([]state.State, error) {
	sts := make([]state.State, 2)

	ac, err := currency.ZeroAccount(cid)
	if err != nil {
		return nil, err
	}
	ast, err := notExistsState(currency.StateKeyAccount(ac.Address()), "keys of zero account", getState)
	if err != nil {
		return nil, err
	}

	ast, err = currency.SetStateAccountValue(ast, ac)
	if err != nil {
		return nil, err
	}
	sts[0] = ast

	bst, _, err := getState(currency.StateKeyBalance(ac.Address(), cid))
	if err != nil {
		return nil, err
	}
	amst := currency.NewAmountState(bst, cid)

	sts[1] = amst

	return sts, nil
}

func (opp *CurrencyRegisterProcessor) Close() error {
	opp.cp = nil
	opp.pubs = nil
	opp.threshold = base.Threshold{}
	opp.ga = currency.AmountState{}
	opp.de = nil

	currencyRegisterProcessorPool.Put(opp)

	return nil
}
