package currency

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var createAccountsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateAccountsItemProcessor)
	},
}

var createAccountsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateAccountsProcessor)
	},
}

type CreateAccountsItemProcessor struct {
	cp   *extensioncurrency.CurrencyPool
	h    valuehash.Hash
	item currency.CreateAccountsItem
	ns   state.State
	nb   map[currency.CurrencyID]currency.AmountState
}

func (opp *CreateAccountsItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		var policy extensioncurrency.CurrencyPolicy
		if opp.cp != nil {
			i, found := opp.cp.Policy(am.Currency())
			if !found {
				return operation.NewBaseReasonError("currency not registered, %q", am.Currency())
			}
			policy = i
		}

		if am.Big().Compare(policy.NewAccountMinBalance()) < 0 {
			return operation.NewBaseReasonError(
				"amount should be over minimum balance, %v < %v", am.Big(), policy.NewAccountMinBalance())
		}
	}

	target, err := opp.item.Address()
	if err != nil {
		return operation.NewBaseReasonErrorFromError(err)
	}

	st, err := notExistsState(currency.StateKeyAccount(target), "keys of target", getState)
	if err != nil {
		return err
	}
	opp.ns = st

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

func (opp *CreateAccountsItemProcessor) Process(
	_ func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) ([]state.State, error) {
	nac, err := currency.NewAccountFromKeys(opp.item.Keys())
	if err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	}

	sts := make([]state.State, len(opp.item.Amounts())+1)
	st, err := currency.SetStateAccountValue(opp.ns, nac)
	if err != nil {
		return nil, err
	}
	sts[0] = st

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		sts[i+1] = opp.nb[am.Currency()].Add(am.Big())
	}

	return sts, nil
}

func (opp *CreateAccountsItemProcessor) Close() error {
	opp.cp = nil
	opp.h = nil
	opp.item = nil
	opp.ns = nil
	opp.nb = nil

	createAccountsItemProcessorPool.Put(opp)

	return nil
}

type CreateAccountsProcessor struct {
	cp *extensioncurrency.CurrencyPool
	currency.CreateAccounts
	sb       map[currency.CurrencyID]currency.AmountState
	ns       []*CreateAccountsItemProcessor
	required map[currency.CurrencyID][2]currency.Big
}

func NewCreateAccountsProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(currency.CreateAccounts)
		if !ok {
			return nil, errors.Errorf("not CreateAccounts, %T", op)
		}

		opp := createAccountsProcessorPool.Get().(*CreateAccountsProcessor)

		opp.cp = cp
		opp.CreateAccounts = i
		opp.sb = nil
		opp.ns = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *CreateAccountsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(currency.CreateAccountsFact)

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

	ns := make([]*CreateAccountsItemProcessor, len(fact.Items()))
	for i := range fact.Items() {
		c := createAccountsItemProcessorPool.Get().(*CreateAccountsItemProcessor)
		c.cp = opp.cp
		c.h = opp.Hash()
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

func (opp *CreateAccountsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(currency.CreateAccountsFact)

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

func (opp *CreateAccountsProcessor) Close() error {
	for i := range opp.ns {
		_ = opp.ns[i].Close()
	}

	opp.cp = nil
	opp.CreateAccounts = currency.CreateAccounts{}

	createAccountsProcessorPool.Put(opp)

	return nil
}

func (opp *CreateAccountsProcessor) calculateItemsFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(currency.CreateAccountsFact)

	items := make([]currency.AmountsItem, len(fact.Items()))
	for i := range fact.Items() {
		items[i] = fact.Items()[i]
	}

	return CalculateItemsFee(opp.cp, items, getState)
}

func CalculateItemsFee(cp *extensioncurrency.CurrencyPool, items []currency.AmountsItem, getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
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

			f, found := cp.Feeer(am.Currency())
			if !found {
				return nil, errors.Errorf("unknown currency id found, %q", am.Currency())
			}
			// 수정
			feeer, ok := f.(feefi.FeefiFeeer)
			id := extensioncurrency.ContractID(am.Currency().String())
			var k currency.Big
			if ok {
				st, err := existsState(feefi.StateKeyDesign(feeer.Feefier(), id), "feefi design", getState)
				if err != nil {
					return nil, err
				}
				design, err := feefi.StateDesignValue(st)
				if err != nil {
					return nil, err
				}
				if design.Policy().Fee().Currency() != am.Currency() {
					return nil, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Policy().Fee().Currency(), am.Currency())
				}
				k = design.Policy().Fee().Big()
			} else {
				var err error
				switch k, err = f.Fee(am.Big()); {
				case err != nil:
					return nil, err
				}
			}
			if !k.OverZero() {
				required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
			} else {
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
) (map[currency.CurrencyID]currency.AmountState, error) {
	sb := map[currency.CurrencyID]currency.AmountState{}

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
