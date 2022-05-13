package currency

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var transfersItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransfersItemProcessor)
	},
}

var transfersProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransfersProcessor)
	},
}

type TransfersItemProcessor struct {
	cp *extensioncurrency.CurrencyPool
	h  valuehash.Hash

	item currency.TransfersItem

	rb map[currency.CurrencyID]currency.AmountState
}

func (opp *TransfersItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	if _, err := existsState(currency.StateKeyAccount(opp.item.Receiver()), "receiver", getState); err != nil {
		return err
	}

	rb := map[currency.CurrencyID]currency.AmountState{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		if opp.cp != nil {
			if !opp.cp.Exists(am.Currency()) {
				return errors.Errorf("currency not registered, %q", am.Currency())
			}
		}

		st, _, err := getState(currency.StateKeyBalance(opp.item.Receiver(), am.Currency()))
		if err != nil {
			return err
		}
		rb[am.Currency()] = currency.NewAmountState(st, am.Currency())
	}

	opp.rb = rb

	return nil
}

func (opp *TransfersItemProcessor) Process(
	_ func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) ([]state.State, error) {
	sts := make([]state.State, len(opp.item.Amounts()))
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		sts[i] = opp.rb[am.Currency()].Add(am.Big())
	}

	return sts, nil
}

func (opp *TransfersItemProcessor) Close() error {
	opp.cp = nil
	opp.h = nil
	opp.item = nil
	opp.rb = nil

	transfersItemProcessorPool.Put(opp)

	return nil
}

type TransfersProcessor struct {
	cp *extensioncurrency.CurrencyPool
	currency.Transfers
	sb       map[currency.CurrencyID]currency.AmountState
	rb       []*TransfersItemProcessor
	required map[currency.CurrencyID][2]currency.Big
}

func NewTransfersProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(currency.Transfers)
		if !ok {
			return nil, errors.Errorf("not Transfers, %T", op)
		}

		opp := transfersProcessorPool.Get().(*TransfersProcessor)

		opp.cp = cp
		opp.Transfers = i
		opp.sb = nil
		opp.rb = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *TransfersProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(currency.TransfersFact)

	if err := checkExistsState(currency.StateKeyAccount(fact.Sender()), getState); err != nil {
		return nil, err
	}

	if required, err := opp.calculateItemsFee(getState); err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	} else if sb, err := CheckEnoughBalance(fact.Sender(), required, getState); err != nil {
		return nil, err
	} else {
		opp.required = required
		opp.sb = sb
	}

	rb := make([]*TransfersItemProcessor, len(fact.Items()))
	for i := range fact.Items() {
		c := transfersItemProcessorPool.Get().(*TransfersItemProcessor)
		c.cp = opp.cp
		c.h = opp.Hash()
		c.item = fact.Items()[i]

		if err := c.PreProcess(getState, setState); err != nil {
			return nil, operation.NewBaseReasonErrorFromError(err)
		}

		rb[i] = c
	}

	if err := checkFactSignsByState(fact.Sender(), opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	opp.rb = rb

	return opp, nil
}

func (opp *TransfersProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(currency.TransfersFact)

	var sts []state.State // nolint:prealloc
	for i := range opp.rb {
		s, err := opp.rb[i].Process(getState, setState)
		if err != nil {
			return operation.NewBaseReasonError("failed to process transfer item: %w", err)
		}
		sts = append(sts, s...)
	}

	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.sb[k].Sub(rq[0]).AddFee(rq[1]))
	}

	return setState(fact.Hash(), sts...)
}

func (opp *TransfersProcessor) Close() error {
	for i := range opp.rb {
		_ = opp.rb[i].Close()
	}

	opp.cp = nil
	opp.Transfers = currency.Transfers{}
	opp.sb = nil
	opp.rb = nil
	opp.required = nil

	transfersProcessorPool.Put(opp)

	return nil
}

func (opp *TransfersProcessor) calculateItemsFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(currency.TransfersFact)

	items := make([]currency.AmountsItem, len(fact.Items()))
	for i := range fact.Items() {
		items[i] = fact.Items()[i]
	}

	return CalculateItemsFee(opp.cp, items, getState)
}
