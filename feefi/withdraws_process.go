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
			return nil, operation.NewBaseReasonError("not Withdraws, %T", op)
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
		return nil, operation.NewBaseReasonError("sender not in pool users, %q", fact.sender)
	}

	pool, err := UpdatePoolUserFromWithdraw(p, fact.pool, fact.sender, fact.amounts, getState)
	if err != nil {
		return nil, operation.NewBaseReasonError("update pool users balance failed, %q", err)
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
				return nil, operation.NewBaseReasonError("currency not registered, %q", cid)
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
				return nil, operation.NewBaseReasonError("currency not registered, %q", cid)
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
		sts = append(sts, opp.wb[k].Add(rq[0].Sub(rq[1])), opp.pb[k].Sub(rq[0]).AddFee(rq[1]), opp.fs)
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
			if design.Policy().Fee().Currency() != am.Currency() {
				return nil, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Policy().Fee().Currency(), am.Currency())
			}
			k = design.Policy().Fee().Big()
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
			required[am.Currency()] = [2]currency.Big{rq[0].Add(am.Big()), rq[1].Add(k)}
		}
	}

	return required, nil
}

func UpdatePoolUserFromWithdraw(pl Pool, feefier base.Address, withdrawer base.Address, amounts []currency.Amount, getState func(key string) (state.State, bool, error)) (Pool, error) {
	pool, nowIncomeBalance, nowOutlayBalance, err := CalculateRewardAndHold(pl, feefier, getState)
	if err != nil {
		return Pool{}, err
	}
	// sub withdraw amount
	for i := range amounts {
		am := amounts[i]
		userBalance, _ := pool.users[withdrawer.String()]
		switch am.Currency() {
		case pool.IncomeBalance().Currency():
			err := userBalance.SubIncome(am.Big())
			if err != nil {
				return Pool{}, err
			}
			pool.users[withdrawer.String()] = userBalance
			nowIncomeBalance, err = nowIncomeBalance.Sub(am.Big())
			if err != nil {
				return Pool{}, err
			}
		case pool.OutlayBalance().Currency():
			err := userBalance.SubOutlay(am.Big())
			if err != nil {
				return Pool{}, err
			}
			pool.users[withdrawer.String()] = userBalance
			nowOutlayBalance, err = nowOutlayBalance.Sub(am.Big())
			if err != nil {
				return Pool{}, err
			}
		}
	}

	// update pool previous balance
	pool.prevIncomeAmount = nowIncomeBalance.Amount()
	pool.prevOutlayAmount = nowOutlayBalance.Amount()
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
