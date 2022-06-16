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

var suffrageInflationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(SuffrageInflationProcessor)
	},
}

type SuffrageInflationProcessor struct {
	extensioncurrency.SuffrageInflation
	cp        *CurrencyPool
	pubs      []key.Publickey
	threshold base.Threshold
	ast       map[string]currency.AmountState
	dst       map[currency.CurrencyID]state.State
	dc        map[currency.CurrencyID]extensioncurrency.CurrencyDesign
}

func NewSuffrageInflationProcessor(cp *CurrencyPool, pubs []key.Publickey, threshold base.Threshold) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(extensioncurrency.SuffrageInflation)
		if !ok {
			return nil, errors.Errorf("not SuffrageInflation, %T", op)
		}

		opp := suffrageInflationProcessorPool.Get().(*SuffrageInflationProcessor)

		opp.cp = cp
		opp.SuffrageInflation = i
		opp.pubs = pubs
		opp.threshold = threshold

		return opp, nil
	}
}

func (opp *SuffrageInflationProcessor) PreProcess(
	getState func(string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	if len(opp.pubs) < 1 {
		return nil, operation.NewBaseReasonError("empty publickeys for operation signs")
	} else if err := checkFactSignsByPubs(opp.pubs, opp.threshold, opp.Signs()); err != nil {
		return nil, err
	}

	items := opp.Fact().(extensioncurrency.SuffrageInflationFact).Items()

	ast := map[string]currency.AmountState{}
	dst := map[currency.CurrencyID]state.State{}
	dc := map[currency.CurrencyID]extensioncurrency.CurrencyDesign{}
	for i := range items {
		item := items[i]
		cid := item.Amount().Currency()
		st, found := opp.cp.State(cid)
		if !found {
			return nil, operation.NewBaseReasonError("unknown currency, %q for SuffrageInflation", cid)
		}
		dst[cid] = st

		if err := checkExistsState(currency.StateKeyAccount(item.Receiver()), getState); err != nil {
			return nil, errors.Wrap(err, "unknown receiver of SuffrageInflation")
		}

		aid := currency.StateKeyBalance(item.Receiver(), item.Amount().Currency())
		if _, found := ast[aid]; !found {
			bst, _, err := getState(currency.StateKeyBalance(item.Receiver(), cid))
			if err != nil {
				return nil, err
			}

			ast[aid] = currency.NewAmountState(bst, cid)
		}

		dc[cid], _ = opp.cp.Get(cid)
	}

	opp.ast = ast
	opp.dst = dst
	opp.dc = dc

	return opp, nil
}

func (opp *SuffrageInflationProcessor) Process(
	_ func(string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	items := opp.Fact().(extensioncurrency.SuffrageInflationFact).Items()

	sts := make([]state.State, len(opp.ast)+len(opp.dst))

	inc := map[currency.CurrencyID]currency.Big{}
	for i := range items {
		item := items[i]
		aid := currency.StateKeyBalance(item.Receiver(), item.Amount().Currency())
		opp.ast[aid] = opp.ast[aid].Add(item.Amount().Big())
		inc[item.Amount().Currency()] = item.Amount().Big()
	}

	var i int
	for k := range opp.ast {
		sts[i] = opp.ast[k]
		i++
	}

	for cid := range inc {
		dc, err := opp.dc[cid].AddAggregate(inc[cid])
		if err != nil {
			return operation.NewBaseReasonErrorFromError(err)
		}

		j, err := SetStateCurrencyDesignValue(opp.dst[cid], dc)
		if err != nil {
			return err
		}

		sts[i] = j
		i++
	}

	return setState(opp.Fact().Hash(), sts...)
}

func (opp *SuffrageInflationProcessor) Close() error {
	opp.cp = nil
	opp.SuffrageInflation = extensioncurrency.SuffrageInflation{}
	opp.pubs = nil
	opp.threshold = base.Threshold{}

	suffrageInflationProcessorPool.Put(opp)

	return nil
}
