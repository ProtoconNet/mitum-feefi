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
var withdrawsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsItemProcessor)
	},
}
*/

var withdrawsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsProcessor)
	},
}

func (Withdraws) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	// NOTE Process is nil func
	return nil
}

/*
type WithdrawsItemProcessor struct {
	cp       *extensioncurrency.CurrencyPool
	h        valuehash.Hash
	sender   base.Address
	item     WithdrawsItem
	required map[currency.CurrencyID][2]currency.Big
	tb       map[currency.CurrencyID]currency.AmountState // all currency amount state of target account
}

func (opp *WithdrawsItemProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	_ func(valuehash.Hash, ...state.State) error,
) error {
	// check existence of target account state
	if _, err := existsState(currency.StateKeyAccount(opp.item.Target()), "target", getState); err != nil {
		return err
	}

	// check existence of contract account status state
	// check contract account is active
	st, err := existsState(extensioncurrency.StateKeyContractAccount(opp.item.Target()), "contract account status", getState)
	if err != nil {
		return err
	}

	// calculate required amount state of items
	// keep required amount state
	required := make(map[currency.CurrencyID][2]currency.Big)
	tb := map[currency.CurrencyID]currency.AmountState{}
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

		// check existence of pool state
		// check user in pool list
		// check currency in pool
		st, err = existsState(StateKeyPool(opp.item.Target(), opp.item.PoolCID()), "feefi pool", getState)
		if err != nil {
			return err
		}
		pool, err := StatePoolValue(st)
		if err != nil {
			return err
		}

		poolUserBalance, ok := pool.User(opp.sender)
		if !ok {
			return errors.Errorf("receiver account is not in pool user list, %q", opp.sender)
		}
		if poolUserBalance.incomeAmount.Currency() != am.Currency() && poolUserBalance.outlayAmount.Currency() != am.Currency() {
			return errors.Errorf("withdraw currency is not in pool, %q", am.Currency())
		}

		feeer, found := opp.cp.Feeer(am.Currency())
		if !found {
			return errors.Errorf("unknown currency id found, %q", am.Currency())
		}
		// known fee
		switch k, err := feeer.Fee(am.Big()); {
		case err != nil:
			return err
		// if fee is zero, add zero fee
		case !k.OverZero():
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}
		// if fee is not zero, add fee
		default:
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
		}

		var compare currency.Amount
		if poolUserBalance.incomeAmount.Currency() == am.Currency() {
			compare = poolUserBalance.incomeAmount
		} else {
			compare = poolUserBalance.outlayAmount
		}

		if compare.Big().Compare(required[am.Currency()][0].Add(required[am.Currency()][1])) < 0 {
			return operation.NewBaseReasonError(
				"insufficient balance of user, %s; %d !> %d", opp.sender.String(), am.Big(), required[am.Currency()][0].Add(required[am.Currency()][1]))
		}

		st, err := existsState(stateKeyBalance(opp.item.Target(), opp.item.PoolCID(), am.Currency()), "currency of pool", getState)
		if err != nil {
			return err
		}

		am, err = stateBalanceValue(st)
		if err != nil {
			return operation.NewBaseReasonError("insufficient balance of sender: %w", err)
		}

		if am.Big().Compare(required[am.Currency()][0].Add(required[am.Currency()][1])) < 0 {
			return operation.NewBaseReasonError(
				"insufficient balance of sender, %s; %d !> %d", opp.item.Target().String(), am.Big(), rq[0].Add(rq[1]))
		}
		tb[am.Currency()] = currency.NewAmountState(st, am.Currency())

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
	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.tb[k].Sub(rq[0].Sub(rq[1])).AddFee(rq[1]))
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
*/

type WithdrawsProcessor struct {
	cp *extensioncurrency.CurrencyPool
	Withdraws
	wb       map[currency.CurrencyID]currency.AmountState          // amount states of withdrawer account
	pb       map[currency.CurrencyID]extensioncurrency.AmountState // amount states of pool account
	fs       state.State
	required map[currency.CurrencyID][2]currency.Big
}

func NewWithdrawsProcessor(cp *extensioncurrency.CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(Withdraws)
		if !ok {
			return nil, errors.Errorf("not Withdraws, %T", op)
		}

		opp := withdrawsProcessorPool.Get().(*WithdrawsProcessor)

		opp.cp = cp
		opp.Withdraws = i
		opp.wb = nil
		opp.pb = nil
		opp.fs = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *WithdrawsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(WithdrawsFact)

	// check fact sender(withdrawer) exist
	if err := checkExistsState(currency.StateKeyAccount(fact.sender), getState); err != nil {
		return nil, err
	}

	// check existence of pool account state
	if _, err := existsState(currency.StateKeyAccount(fact.pool), "feefi pool", getState); err != nil {
		return nil, err
	}

	// check existence of contract account status state
	// check contract account is active
	_, err := existsState(extensioncurrency.StateKeyContractAccount(fact.pool), "contract account status", getState)
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
	_, found := p.users[fact.sender.String()]
	if !found {
		return nil, errors.Errorf("sender not in pool users, %q", fact.sender)
	}

	pool, err := UpdatePoolUserWithdraw(p, fact.pool, fact.sender, fact.amounts, getState)
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
	} else if pb, err := CheckEnoughBalance(fact.pool, fact.poolID, required, getState); err != nil {
		return nil, err
	} else {
		// required amount and fee
		opp.required = required
		// pool amount state before update
		opp.pb = pb
	}

	// run preprocess of all withdraw item processor

	// prepare fact sender(withdrawer) amount state
	// keep fact sender(withdrawer) amount state map
	wb := map[currency.CurrencyID]currency.AmountState{}
	for cid := range opp.required {
		if opp.cp != nil {
			if !opp.cp.Exists(cid) {
				return nil, errors.Errorf("currency not registered, %q", cid)
			}
		}

		st, _, err := getState(currency.StateKeyBalance(fact.sender, cid))
		if err != nil {
			return nil, err
		}

		wb[cid] = currency.NewAmountState(st, cid)
	}
	// sender(withdrawer) account balance state before update
	opp.wb = wb

	// prepare fact pool amount state
	// keep pool amount state map
	pb := map[currency.CurrencyID]extensioncurrency.AmountState{}
	for cid := range opp.required {
		if opp.cp != nil {
			if !opp.cp.Exists(cid) {
				return nil, errors.Errorf("currency not registered, %q", cid)
			}
		}

		st, _, err := getState(extensioncurrency.StateKeyBalance(fact.pool, fact.poolID, cid, StateKeyBalanceSuffix))
		if err != nil {
			return nil, err
		}

		pb[cid] = extensioncurrency.NewAmountState(st, cid, fact.poolID)
	}
	// pool account balance state before update
	opp.pb = pb

	if err := checkFactSignsByState(fact.sender, opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	return opp, nil
}

func (opp *WithdrawsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(WithdrawsFact)
	var sts []state.State // nolint:prealloc
	// sender(withdrawer) account balance state after update
	// pool account balance state after update
	for k := range opp.required {
		rq := opp.required[k]
		sts = append(sts, opp.wb[k].Add(rq[0].Sub(rq[1])), opp.pb[k].Sub(rq[0]).AddFee(rq[1]))
	}

	return setState(fact.Hash(), sts...)
}

func (opp *WithdrawsProcessor) Close() error {
	opp.cp = nil
	opp.Withdraws = Withdraws{}
	opp.wb = nil
	opp.pb = nil
	opp.fs = nil
	opp.required = nil

	withdrawsProcessorPool.Put(opp)

	return nil
}

func (opp *WithdrawsProcessor) calculateFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(WithdrawsFact)

	return CalculateWithdrawFee(opp.cp, fact, getState)
}

func CalculateWithdrawFee(cp *extensioncurrency.CurrencyPool, fact WithdrawsFact, getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	required := map[currency.CurrencyID][2]currency.Big{}

	for i := range fact.amounts {
		am := fact.amounts[i]

		rq := [2]currency.Big{currency.ZeroBig, currency.ZeroBig}
		if cp == nil {
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1]}

			continue
		}
		v, found := cp.Feeer(am.Currency())
		if !found {
			return nil, errors.Errorf("unknown currency id found, %q", am.Currency())
		}
		feeer, ok := v.(FeefiFeeer)
		var k currency.Big
		if ok {
			st, err := existsState(StateKeyDesign(feeer.Feefier(), fact.poolID), "feefi design", getState)
			if err != nil {
				return nil, err
			}

			design, err := StateDesignValue(st)
			if err != nil {
				return nil, err
			}
			if design.Fee().Currency() != am.Currency() {
				return nil, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Fee().Currency(), am.Currency())
			}
			k = design.Fee().Big()
		} else {
			var err error
			switch k, err = v.Fee(am.Big()); {
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

	return required, nil
}

func UpdatePoolUserWithdraw(pool Pool, feefier base.Address, withdrawer base.Address, amounts []currency.Amount, getState func(key string) (state.State, bool, error)) (Pool, error) {
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
	diffIncomeAmount := currentIncomeBalance.Amount().Big().Sub(prvIncomeBalance.Big())
	diffOutlayAmount := currentOutlayBalance.Amount().Big().Sub(prvOutlayBalance.Big())
	// sum of locking tokent
	total := prvOutlayBalance
	// calculate income and outlay of users
	if total.Big().OverZero() {
		for k := range pool.users {
			c := pool.users[k]
			foutlay := new(big.Float).SetInt(c.outlay.Amount().Big().Int)
			ftotal := new(big.Float).SetInt(total.Big().Int)
			fratio := new(big.Float).Quo(foutlay, ftotal)
			if !diffIncomeAmount.IsZero() {
				fdiff := new(big.Float).SetInt(diffIncomeAmount.Int)
				fcal := new(big.Float).Mul(fratio, fdiff)
				ical, _ := fcal.Int(nil)
				am := currency.NewAmount(currency.NewBigFromBigInt(ical), pool.prevIncomeAmount.Currency())
				c.AddIncome(am)
			}
			if !diffOutlayAmount.IsZero() {
				delta := big.NewFloat(-0.5)
				fdiff := new(big.Float).SetInt(diffOutlayAmount.Int)
				fcal := new(big.Float).Mul(fratio, fdiff)
				nfcal := new(big.Float).Add(fcal, delta)
				ical, _ := nfcal.Int(nil)
				am := currency.NewAmount(currency.NewBigFromBigInt(ical), pool.prevOutlayAmount.Currency())
				c.AddOutlay(am)
			}
			pool.users[k] = c
		}
	}
	// sub withdraw amount
	for i := range amounts {
		am := amounts[i]
		wb, _ := pool.users[withdrawer.String()]
		switch am.Currency() {
		case pool.IncomeBalance().Currency():
			wb.SubIncome(am)
			pool.users[withdrawer.String()] = wb
			currentIncomeBalance, err = currentIncomeBalance.Sub(am)
			if err != nil {
				return Pool{}, err
			}
		case pool.OutlayBalance().Currency():
			wb.SubOutlay(am)
			pool.users[withdrawer.String()] = wb
			currentOutlayBalance, err = currentOutlayBalance.Sub(am)
			if err != nil {
				return Pool{}, err
			}
		}
	}

	// update pool previous balance
	pool.prevIncomeAmount = currentIncomeBalance.Amount()
	pool.prevOutlayAmount = currentOutlayBalance.Amount()
	return pool, nil
}

func CheckEnoughBalance(
	holder base.Address,
	id extensioncurrency.ContractID,
	required map[currency.CurrencyID][2]currency.Big,
	getState func(key string) (state.State, bool, error),
) (map[currency.CurrencyID]extensioncurrency.AmountState, error) {
	sb := map[currency.CurrencyID]extensioncurrency.AmountState{}

	for cid := range required {
		rq := required[cid]

		st, err := existsState(extensioncurrency.StateKeyBalance(holder, id, cid, StateKeyBalanceSuffix), "currency of holder", getState)
		if err != nil {
			return nil, err
		}

		am, err := extensioncurrency.StateBalanceValue(st)
		if err != nil {
			return nil, operation.NewBaseReasonError("insufficient balance of sender: %w", err)
		}

		if am.Amount().Big().Compare(rq[0]) < 0 {
			return nil, operation.NewBaseReasonError(
				"insufficient balance of sender, %s; %d !> %d", holder.String(), am.Amount().Big(), rq[0])
		}
		sb[cid] = extensioncurrency.NewAmountState(st, cid, id)
	}

	return sb, nil
}
