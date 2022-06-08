package feefi

import (
	"math/big"
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

/*
var depositsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(DepositsItemProcessor)
	},
}
*/

var depositsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(DepositsProcessor)
	},
}

func (Deposit) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	// NOTE Process is nil func
	return nil
}

/*
type DepositsItemProcessor struct {
	cp     *extensioncurrency.CurrencyPool
	h      valuehash.Hash
	sender base.Address
	item   DepositsItem
	fs     map[currency.CurrencyID]state.State
	rb     map[currency.CurrencyID]currency.AmountState
}

func (opp *DepositsItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	// check existence of pool account state
	if _, err := existsState(currency.StateKeyAccount(opp.item.Pool()), "receiver", getState); err != nil {
		return err
	}

	// check existence of contract account status state of pool
	_, err := existsState(extensioncurrency.StateKeyContractAccount(opp.item.Pool()), "contract account status", getState)
	if err != nil {
		return err
	}

	rb := map[currency.CurrencyID]currency.AmountState{}
	fs := map[currency.CurrencyID]state.State{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		if opp.cp != nil {
			if !opp.cp.Exists(am.Currency()) {
				return errors.Errorf("currency not registered, %q", am.Currency())
			}
		}
		// check pool allowed to receive currency
		_, err := existsState(StateKeyDesign(opp.item.Pool(), opp.item.PoolCID()), " feefi pool design", getState)
		if err != nil {
			return err
		}

		// prepare pool state
		// keep pool state
		// keep pool value
		st, err := existsState(StateKeyPool(opp.item.Pool(), opp.item.PoolCID()), " feefi pool", getState)
		if err != nil {
			return err
		}
		v, err := StatePoolValue(st)
		if err != nil {
			return err
		}

		// check amount currency
		if v.prevOutlayBalance.Currency() != am.Currency() {
			return errors.Errorf("currency not registered, %q", am.Currency())
		}

		pool, err := UpdatePoolUserDeposit(v, opp.item.Pool(), opp.sender, am, getState)
		if err != nil {
			return err
		}
		nst, err := setStatePoolValue(st, pool)
		if err != nil {
			return err
		}

		fs[am.Currency()] = nst

		st, _, err = getState(stateKeyBalance(opp.item.Pool(), opp.item.PoolCID(), am.Currency()))
		if err != nil {
			return err
		}
		rb[am.Currency()] = currency.NewAmountState(st, am.Currency())
	}

	opp.fs = fs
	opp.rb = rb

	return nil
}

func (opp *DepositsItemProcessor) Process(
	_ func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) ([]state.State, error) {
	sts := make([]state.State, len(opp.item.Amounts())*2)
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		sts[i*2] = opp.rb[am.Currency()].Add(am.Big())
		sts[i*2+1] = opp.fs[am.Currency()]
	}

	return sts, nil
}

func (opp *DepositsItemProcessor) Close() error {
	opp.cp = nil
	opp.h = nil
	opp.sender = nil
	opp.item = nil
	opp.rb = nil

	depositsItemProcessorPool.Put(opp)

	return nil
}
*/

type DepositsProcessor struct {
	cp *extensioncurrency.CurrencyPool
	Deposit
	sb       currency.AmountState
	fs       state.State
	rb       extensioncurrency.AmountState
	required [2]currency.Big
}

func NewDepositsProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(Deposit)
		if !ok {
			return nil, errors.Errorf("not Deposits, %T", op)
		}

		opp := depositsProcessorPool.Get().(*DepositsProcessor)

		opp.cp = cp
		opp.Deposit = i
		opp.sb = currency.AmountState{}
		opp.fs = nil
		opp.rb = extensioncurrency.AmountState{}
		opp.required = [2]currency.Big{}

		return opp, nil
	}
}

func (opp *DepositsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(DepositFact)
	// check existence of sender account state
	if err := checkExistsState(currency.StateKeyAccount(fact.sender), getState); err != nil {
		return nil, err
	}

	// check existence of pool account state
	if _, err := existsState(currency.StateKeyAccount(fact.pool), "feefi pool account", getState); err != nil {
		return nil, err
	}

	// check existence of contract account status state of pool
	_, err := existsState(extensioncurrency.StateKeyContractAccount(fact.pool), "contract account status", getState)
	if err != nil {
		return nil, err
	}

	am := fact.amount
	if opp.cp != nil {
		if !opp.cp.Exists(am.Currency()) {
			return nil, errors.Errorf("currency not registered, %q", am.Currency())
		}
	}

	// check pool allowed to receive currency
	_, err = existsState(StateKeyDesign(fact.pool, fact.poolID), " feefi pool design", getState)
	if err != nil {
		return nil, err
	}

	// prepare pool state
	// keep pool state
	st, err := existsState(StateKeyPool(fact.pool, fact.poolID), " feefi pool", getState)
	if err != nil {
		return nil, err
	}
	p, err := StatePoolValue(st)
	if err != nil {
		return nil, err
	}

	pool, err := UpdatePoolUserDeposit(p, fact.pool, fact.sender, am, getState)
	if err != nil {
		return nil, err
	}
	nst, err := setStatePoolValue(st, pool)
	if err != nil {
		return nil, err
	}
	// feefi pool state after update
	opp.fs = nst

	if required, err := opp.calculateFee(getState); err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	} else if sb, err := CheckEnoughBalanceDeposit(fact.sender, am.Currency(), required, getState); err != nil {
		return nil, err
	} else {
		// required amount and fee
		opp.required = required
		// sender amount state before update
		opp.sb = sb
	}

	st, _, err = getState(extensioncurrency.StateKeyBalance(fact.pool, fact.poolID, am.Currency(), StateKeyBalanceSuffix))
	if err != nil {
		return nil, err
	}
	// feefipool account balance state before update
	opp.rb = extensioncurrency.NewAmountState(st, am.Currency(), fact.poolID)

	if err := checkFactSignsByState(fact.sender, opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	return opp, nil
}

func (opp *DepositsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(DepositFact)
	var sts []state.State // nolint:prealloc
	// sender account balance state after update
	st := opp.sb.Sub(opp.required[0]).AddFee(opp.required[1])
	rst := opp.rb.Add(fact.amount.Big())
	sts = append(sts, opp.fs, rst, st)

	return setState(fact.Hash(), sts...)
}

func (opp *DepositsProcessor) Close() error {
	opp.cp = nil
	opp.Deposit = Deposit{}
	opp.sb = currency.AmountState{}
	opp.fs = nil
	opp.rb = extensioncurrency.AmountState{}
	opp.required = [2]currency.Big{}

	depositsProcessorPool.Put(opp)

	return nil
}

func (opp *DepositsProcessor) calculateFee(getState func(key string) (state.State, bool, error)) ([2]currency.Big, error) {
	fact := opp.Fact().(DepositFact)

	return CalculateDepositFee(opp.cp, fact, getState)
}

func CalculateDepositFee(cp *extensioncurrency.CurrencyPool, fact DepositFact, getState func(key string) (state.State, bool, error)) ([2]currency.Big, error) {
	required := [2]currency.Big{}

	am := fact.amount

	rq := [2]currency.Big{currency.ZeroBig, currency.ZeroBig}

	if cp == nil {
		required = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
		return required, nil
	}

	v, found := cp.Feeer(am.Currency())
	if !found {
		return [2]currency.Big{}, errors.Errorf("unknown currency id found, %q", am.Currency())
	}
	feeer, ok := v.(FeefiFeeer)
	var k currency.Big
	if ok {
		st, err := existsState(StateKeyDesign(feeer.Feefier(), fact.poolID), "feefi design", getState)
		if err != nil {
			return [2]currency.Big{}, err
		}

		design, err := StateDesignValue(st)
		if err != nil {
			return [2]currency.Big{}, err
		}
		if design.Fee().Currency() != am.Currency() {
			return [2]currency.Big{}, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Fee().Currency(), am.Currency())
		}
		k = design.Fee().Big()
	} else {
		var err error
		switch k, err = v.Fee(am.Big()); {
		case err != nil:
			return [2]currency.Big{}, err
		}
	}
	if !k.OverZero() {
		required = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
	} else {
		required = [2]currency.Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
	}

	return required, nil
}

func UpdatePoolUserDeposit(pool Pool, feefier base.Address, depositer base.Address, am currency.Amount, getState func(key string) (state.State, bool, error)) (Pool, error) {
	id := extensioncurrency.ContractID(pool.prevIncomeAmount.Currency().String())
	st, err := existsState(extensioncurrency.StateKeyBalance(feefier, id, pool.prevIncomeAmount.Currency(), StateKeyBalanceSuffix), " feefi pool balance", getState)
	if err != nil {
		return Pool{}, err
	}
	currentIncomeBalance, err := extensioncurrency.StateBalanceValue(st)
	if err != nil {
		return Pool{}, err
	}
	st, err = existsState(extensioncurrency.StateKeyBalance(feefier, id, pool.prevOutlayAmount.Currency(), StateKeyBalanceSuffix), " feefi pool balance", getState)
	if err != nil {
		return Pool{}, err
	}
	currentOutlayBalance, err := extensioncurrency.StateBalanceValue(st)
	if err != nil {
		return Pool{}, err
	}
	prvIncomeBalance := pool.prevIncomeAmount
	prvOutlayBalance := pool.prevOutlayAmount
	// difference between previous balance and current balance
	diffIncome := currentIncomeBalance.Amount().Big().Sub(prvIncomeBalance.Big())
	diffOutlay := currentOutlayBalance.Amount().Big().Sub(prvOutlayBalance.Big())
	// sum of locking tokent
	total := prvOutlayBalance
	// calculate income and outlay of users
	if total.Big().OverZero() {
		for k := range pool.users {
			c := pool.users[k]
			foutlay := new(big.Float).SetInt(c.outlay.Amount().Big().Int)
			ftotal := new(big.Float).SetInt(total.Big().Int)
			fratio := new(big.Float).Quo(foutlay, ftotal)
			if !diffIncome.IsZero() {
				fdiff := new(big.Float).SetInt(diffIncome.Int)
				fcal := new(big.Float).Mul(fratio, fdiff)
				ical, _ := fcal.Int(nil)
				am := currency.NewAmount(currency.NewBigFromBigInt(ical), pool.prevIncomeAmount.Currency())
				c.AddIncome(am)
			}
			if !diffOutlay.IsZero() {
				delta := big.NewFloat(-0.5)
				fdiff := new(big.Float).SetInt(diffOutlay.Int)
				fcal := new(big.Float).Mul(fratio, fdiff)
				nfcal := new(big.Float).Add(fcal, delta)
				ical, _ := nfcal.Int(nil)
				am := currency.NewAmount(currency.NewBigFromBigInt(ical), pool.prevOutlayAmount.Currency())
				c.AddOutlay(am)
			}
			pool.users[k] = c
		}
	}
	// add deposit amount
	if db, found := pool.users[depositer.String()]; !found {
		pool.users[depositer.String()] = NewPoolUserBalance(
			extensioncurrency.NewAmountValue(currency.ZeroBig,
				pool.prevIncomeAmount.Currency(),
				id),
			extensioncurrency.NewAmountValuefromAmount(am, id),
		)
	} else {
		db.AddOutlay(am)
		pool.users[depositer.String()] = db
	}
	currentOutlayBalance, err = currentOutlayBalance.Add(am)
	if err != nil {
		return Pool{}, err
	}
	// update pool previous balance
	pool.prevIncomeAmount = currentIncomeBalance.Amount()
	pool.prevOutlayAmount = currentOutlayBalance.Amount()
	return pool, nil
}

func CheckEnoughBalanceDeposit(
	holder base.Address,
	cid currency.CurrencyID,
	required [2]currency.Big,
	getState func(key string) (state.State, bool, error),
) (currency.AmountState, error) {
	st, err := existsState(currency.StateKeyBalance(holder, cid), "balance of holder", getState)
	if err != nil {
		return currency.AmountState{}, err
	}

	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return currency.AmountState{}, operation.NewBaseReasonError("insufficient balance of sender: %w", err)
	}

	if am.Big().Compare(required[0].Add(required[1])) < 0 {
		return currency.AmountState{}, operation.NewBaseReasonError(
			"insufficient balance of sender, %s; %d !> %d", holder.String(), am.Big(), required[0].Add(required[1]))
	}

	return currency.NewAmountState(st, cid), nil
}
