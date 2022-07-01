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

var withdrawsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsItemProcessor)
	},
}

var withdrawsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsProcessor)
	},
}

type WithdrawsItemProcessor struct {
	cp       *CurrencyPool
	h        valuehash.Hash
	sender   base.Address
	item     extensioncurrency.WithdrawsItem
	required map[currency.CurrencyID][2]currency.Big      // required amount
	tb       map[currency.CurrencyID]currency.AmountState // all currency amount state of target account
}

func (opp *WithdrawsItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	// check existence of target(dA)
	if _, err := existsState(currency.StateKeyAccount(opp.item.Target()), "target", getState); err != nil {
		return err
	}

	// check existence of dA status state
	// check sender matched with dA owner
	st, err := existsState(extensioncurrency.StateKeyContractAccount(opp.item.Target()), "decentralized account status", getState)
	if err != nil {
		return err
	}

	v, err := extensioncurrency.StateContractAccountValue(st)
	if err != nil {
		return err
	}
	if !v.Owner().Equal(opp.sender) {
		return operation.NewBaseReasonError("decentralized account owner is not matched with %q", opp.sender)
	}

	// calculate required amount state of items
	// keep required amount state
	required := make(map[currency.CurrencyID][2]currency.Big)
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		// rq[0] is amount, rq[1] is fee
		rq := [2]currency.Big{currency.ZeroBig, currency.ZeroBig}

		// found required in map of currency id
		if k, found := required[am.Currency()]; found {
			rq = k
		}

		// because currency pool is nil, only add amount and no fee
		if opp.cp == nil {
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
			continue
		}

		feeer, found := opp.cp.Feeer(am.Currency())
		// if feeer not found, unknown currency id
		if !found {
			return operation.NewBaseReasonError("unknown currency id found, %q", am.Currency())
		}
		// known fee
		switch k, err := feeer.Fee(am.Big()); {
		case err != nil:
			return err
		// fee is zero
		case !k.OverZero():
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
		// add fee
		default:
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
		}
	}

	// check amount state of target
	// check target has enough amount
	// keep required amount state
	tb := map[currency.CurrencyID]currency.AmountState{}
	for cid := range required {
		rq := required[cid]

		st, err := existsState(currency.StateKeyBalance(opp.item.Target(), cid), "currency of holder", getState)
		if err != nil {
			return err
		}

		am, err := currency.StateBalanceValue(st)
		if err != nil {
			return operation.NewBaseReasonError("insufficient balance of sender: %w", err)
		}

		if am.Big().Compare(rq[0].Add(rq[1])) < 0 {
			return operation.NewBaseReasonError(
				"insufficient balance of sender, %s; %d !> %d", opp.item.Target().String(), am.Big(), rq[0].Add(rq[1]))
		}
		// NewAmountState return amount state if st is amount state else return new zero amount state
		tb[cid] = currency.NewAmountState(st, cid)
	}
	opp.required = required
	opp.tb = tb

	return nil
}

func (opp *WithdrawsItemProcessor) Process(
	_ func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) ([]state.State, error) {
	var sts []state.State
	// from required calculate target amount state
	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.tb[k].Sub(rq[0]).AddFee(rq[1]))
	}

	return sts, nil
}

func (opp *WithdrawsItemProcessor) Close() error {
	opp.cp = nil
	opp.h = nil
	opp.sender = nil
	opp.item = nil
	opp.required = nil
	opp.tb = nil

	withdrawsItemProcessorPool.Put(opp)

	return nil
}

type WithdrawsProcessor struct {
	cp *CurrencyPool
	extensioncurrency.Withdraws
	rb       map[currency.CurrencyID]currency.AmountState
	tb       []*WithdrawsItemProcessor
	required map[currency.CurrencyID][2]currency.Big // all required amount in items
}

func NewWithdrawsProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(extensioncurrency.Withdraws)
		if !ok {
			return nil, errors.Errorf("not Withdraws, %T", op)
		}

		opp := withdrawsProcessorPool.Get().(*WithdrawsProcessor)

		opp.cp = cp
		opp.Withdraws = i
		opp.rb = nil
		opp.tb = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *WithdrawsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(extensioncurrency.WithdrawsFact)

	// check fact sender(withdrawer) exist
	if err := checkExistsState(currency.StateKeyAccount(fact.Sender()), getState); err != nil {
		return nil, err
	}

	// calculate all required amount in items
	if required, err := opp.calculateItemsFee(getState); err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	} else {
		opp.required = required
	}

	// prepare fact sender(withdrawer) amount state
	// keep fact sender(withdrawer) amount state map
	rb := map[currency.CurrencyID]currency.AmountState{}
	for cid := range opp.required {
		if opp.cp != nil {
			if !opp.cp.Exists(cid) {
				return nil, operation.NewBaseReasonError("currency not registered, %q", cid)
			}
		}

		st, _, err := getState(currency.StateKeyBalance(fact.Sender(), cid))
		if err != nil {
			return nil, err
		}

		rb[cid] = currency.NewAmountState(st, cid)
	}
	opp.rb = rb

	// run preprocess of all withdraw item processor
	// get all preprocessed withdraw item processor
	tb := make([]*WithdrawsItemProcessor, len(fact.Items()))
	for i := range fact.Items() {
		c := withdrawsItemProcessorPool.Get().(*WithdrawsItemProcessor)
		c.cp = opp.cp
		c.h = opp.Hash()
		c.sender = fact.Sender()
		c.item = fact.Items()[i]

		if err := c.PreProcess(getState, setState); err != nil {
			return nil, operation.NewBaseReasonErrorFromError(err)
		}

		tb[i] = c
	}

	if err := checkFactSignsByState(fact.Sender(), opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	opp.tb = tb

	return opp, nil
}

func (opp *WithdrawsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(extensioncurrency.WithdrawsFact)

	var sts []state.State // nolint:prealloc
	// run process of all item processor
	// get target(sender) amount state
	for i := range opp.tb {
		s, err := opp.tb[i].Process(getState, setState)
		if err != nil {
			return operation.NewBaseReasonError("failed to process transfer item: %w", err)
		}
		sts = append(sts, s...)
	}

	// add required amount to sender(withdrawer) account amount
	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.rb[k].Add(rq[0]))
	}

	return setState(fact.Hash(), sts...)
}

func (opp *WithdrawsProcessor) Close() error {
	for i := range opp.tb {
		_ = opp.tb[i].Close()
	}

	opp.cp = nil
	opp.Withdraws = extensioncurrency.Withdraws{}
	opp.rb = nil
	opp.tb = nil
	opp.required = nil

	withdrawsProcessorPool.Put(opp)

	return nil
}

func (opp *WithdrawsProcessor) calculateItemsFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(extensioncurrency.WithdrawsFact)

	items := make([]currency.AmountsItem, len(fact.Items()))
	for i := range fact.Items() {
		items[i] = fact.Items()[i]
	}

	return CalculateItemsFee(opp.cp, items, getState)
}
