package feefi

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var createContractAccountsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountsItemProcessor)
	},
}

var createContractAccountsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountsProcessor)
	},
}

type CreateContractAccountsItemProcessor struct {
	cp     *CurrencyPool
	h      valuehash.Hash
	sender base.Address
	item   extensioncurrency.CreateContractAccountsItem
	ns     state.State                                  // target account(contract account) state
	oas    state.State                                  // contract account status state
	oac    currency.Account                             // owner account value
	nb     map[currency.CurrencyID]currency.AmountState // target account amount state
}

func (opp *CreateContractAccountsItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		// check currency registered
		var policy extensioncurrency.CurrencyPolicy
		if opp.cp != nil {
			i, found := opp.cp.Policy(am.Currency())
			if !found {
				return operation.NewBaseReasonError("currency not registered, %q", am.Currency())
			}
			policy = i
		}

		// check minimum balance
		if am.Big().Compare(policy.NewAccountMinBalance()) < 0 {
			return operation.NewBaseReasonError(
				"amount should be over minimum balance, %v < %v", am.Big(), policy.NewAccountMinBalance())
		}
	}

	target, err := opp.item.Address()
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}

	// check not existence of target account state
	// keep target account state
	st, err := notExistsState(currency.StateKeyAccount(target), "keys of target", getState)
	if err != nil {
		return err
	}
	opp.ns = st

	// check existence of owner account state
	// keep owner account value
	st, err = existsState(currency.StateKeyAccount(opp.sender), "account of owner", getState)
	if err != nil {
		return err
	}
	oac, err := currency.LoadStateAccountValue(st)
	if err != nil {
		return err
	}
	opp.oac = oac

	// check not existence of contract account status state
	// keep contract account status state
	st, err = notExistsState(extensioncurrency.StateKeyContractAccount(target), "contract account status", getState)
	if err != nil {
		return err
	}
	opp.oas = st

	// prepare and keep target account balance value
	nb := map[currency.CurrencyID]currency.AmountState{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		b, _, err := getState(currency.StateKeyBalance(target, am.Currency()))
		if err != nil {
			return err
		}
		nb[am.Currency()] = currency.NewAmountState(b, am.Currency())
	}

	opp.nb = nb

	return nil
}

func (opp *CreateContractAccountsItemProcessor) Process(
	_ func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) ([]state.State, error) {
	// new target account from keys and remove keys
	nac, err := currency.NewAccountFromKeys(opp.item.Keys())
	ks := NewContractAccountKeys()
	ncac, err := nac.SetKeys(ks)
	if err != nil {
		return nil, err
	}
	// count of states except amount states
	statesCount := 2
	// set target account value to state
	sts := make([]state.State, len(opp.item.Amounts())+statesCount)
	nst, err := currency.SetStateAccountValue(opp.ns, ncac)
	if err != nil {
		return nil, err
	}
	cas := extensioncurrency.NewContractAccount(opp.oac.Address(), true)
	// set contract account status value to state
	ost, err := extensioncurrency.SetStateContractAccountValue(opp.oas, cas)
	if err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	}

	// add target account state and contract account owner state to states
	sts[0] = nst
	sts[1] = ost
	// calculate target account balance value
	// add target balance state to states
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		sts[i+statesCount] = opp.nb[am.Currency()].Add(am.Big())
	}

	return sts, nil
}

func (opp *CreateContractAccountsItemProcessor) Close() error {
	opp.cp = nil
	opp.h = nil
	opp.sender = nil
	opp.item = nil
	opp.ns = nil
	opp.oas = nil
	opp.oac = currency.Account{}
	opp.nb = nil

	createContractAccountsItemProcessorPool.Put(opp)

	return nil
}

type CreateContractAccountsProcessor struct {
	cp *CurrencyPool
	extensioncurrency.CreateContractAccounts
	sb       map[currency.CurrencyID]currency.AmountState
	ns       []*CreateContractAccountsItemProcessor
	required map[currency.CurrencyID][2]currency.Big
}

func NewCreateContractAccountsProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(extensioncurrency.CreateContractAccounts)
		if !ok {
			return nil, errors.Errorf("not CreateContractAccounts, %T", op)
		}

		opp := createContractAccountsProcessorPool.Get().(*CreateContractAccountsProcessor)

		opp.cp = cp
		opp.CreateContractAccounts = i
		opp.sb = nil
		opp.ns = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *CreateContractAccountsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(extensioncurrency.CreateContractAccountsFact)

	if err := checkExistsState(currency.StateKeyAccount(fact.Sender()), getState); err != nil {
		return nil, err
	}

	if required, err := opp.calculateItemsFee(getState); err != nil {
		return nil, operation.NewBaseReasonError("failed to calculate fee: %w", err)
	} else if sb, err := CheckEnoughBalance(fact.Sender(), required, getState); err != nil {
		return nil, err
	} else {
		opp.required = required
		opp.sb = sb
	}

	ns := make([]*CreateContractAccountsItemProcessor, len(fact.Items()))
	for i := range fact.Items() {
		c := createContractAccountsItemProcessorPool.Get().(*CreateContractAccountsItemProcessor)
		c.cp = opp.cp
		c.h = opp.Hash()
		c.sender = fact.Sender()
		c.item = fact.Items()[i]

		if err := c.PreProcess(getState, setState); err != nil {
			return nil, err
		}

		ns[i] = c
	}

	if err := checkFactSignsByState(fact.Sender(), opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	opp.ns = ns

	return opp, nil
}

func (opp *CreateContractAccountsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(extensioncurrency.CreateContractAccountsFact)

	var sts []state.State // nolint:prealloc
	for i := range opp.ns {
		s, err := opp.ns[i].Process(getState, setState)
		if err != nil {
			return operation.NewBaseReasonError("failed to process create account item: %w", err)
		}
		sts = append(sts, s...)
	}

	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.sb[k].Sub(rq[0]).AddFee(rq[1]))
	}

	return setState(fact.Hash(), sts...)
}

func (opp *CreateContractAccountsProcessor) Close() error {
	for i := range opp.ns {
		_ = opp.ns[i].Close()
	}

	opp.cp = nil
	opp.CreateContractAccounts = extensioncurrency.CreateContractAccounts{}

	createContractAccountsProcessorPool.Put(opp)

	return nil
}

func (opp *CreateContractAccountsProcessor) calculateItemsFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(extensioncurrency.CreateContractAccountsFact)

	items := make([]currency.AmountsItem, len(fact.Items()))
	for i := range fact.Items() {
		items[i] = fact.Items()[i]
	}

	return CalculateItemsFee(opp.cp, items, getState)
}

/*
func CalculateItemsFee(cp *extensioncurrency.CurrencyPool, items []AmountsItem) (map[currency.CurrencyID][2]currency.Big, error) {
	required := map[currency.CurrencyID][2]currency.Big{}

	for i := range items {
		it := items[i]

		for j := range it.Amounts() {
			am := it.Amounts()[j]

			rq := [2]currency.Big{currency.ZeroBig, currency.ZeroBig}
			if k, found := required[am.Currency()]; found {
				rq = k
			}

			if cp == nil {
				required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}

				continue
			}

			feeer, found := cp.Feeer(am.Currency())
			if !found {
				return nil, errors.Errorf("unknown currency id found, %q", am.Currency())
			}
			switch k, err := feeer.Fee(am.Big()); {
			case err != nil:
				return nil, err
			case !k.OverZero():
				required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
			default:
				required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
			}
		}
	}

	return required, nil
}

func CheckEnoughBalance(
	holder base.Address,
	required map[currency.CurrencyID][2]currency.Big,
	getState func(key string) (state.State, bool, error),
) (map[currency.CurrencyID]AmountState, error) {
	sb := map[currency.CurrencyID]AmountState{}

	for cid := range required {
		rq := required[cid]

		st, err := existsState(currency.StateKeyBalance(holder, cid), "currency of holder", getState)
		if err != nil {
			return nil, err
		}

		am, err := currency.StateBalanceValue(st)
		if err != nil {
			return nil, operation.NewBaseReasonError("insufficient balance of sender: %w", err)
		}

		if am.Big().Compare(rq[0]) < 0 {
			return nil, operation.NewBaseReasonError(
				"insufficient balance of sender, %s; %d !> %d", holder.String(), am.Big(), rq[0])
		}
		sb[cid] = currency.NewAmountState(st, cid)
	}

	return sb, nil
}
*/
