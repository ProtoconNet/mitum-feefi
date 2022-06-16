package feefi

import (
	"fmt"
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/valuehash"
)

var poolWithdrawsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(PoolWithdrawsProcessor)
	},
}

func (PoolWithdraws) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	// NOTE Process is nil func
	return nil
}

type PoolWithdrawsProcessor struct {
	cp *CurrencyPool
	PoolWithdraws
	wb       map[currency.CurrencyID]currency.AmountState          // amount states of withdrawer account
	pb       map[currency.CurrencyID]extensioncurrency.AmountState // amount states of pool account
	fs       state.State
	required map[currency.CurrencyID][2]currency.Big
}

func NewPoolWithdrawsProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(PoolWithdraws)
		if !ok {
			return nil, operation.NewBaseReasonError("not Withdraws, %T", op)
		}

		opp := poolWithdrawsProcessorPool.Get().(*PoolWithdrawsProcessor)

		opp.cp = cp
		opp.PoolWithdraws = i
		opp.wb = nil
		opp.pb = nil
		opp.fs = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *PoolWithdrawsProcessor) PreProcess(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) (state.Processor, error) {
	fact := opp.Fact().(PoolWithdrawsFact)

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

	// check feefi pool registered
	st, err := existsState(StateKeyCurrencyDesign((currency.CurrencyID(fact.poolID))), "currency design", getState)
	if err != nil {
		return nil, err
	}
	currencyDesign, err := StateCurrencyDesignValue(st)
	if err != nil {
		return nil, err
	}
	feeer, ok := currencyDesign.Policy().Feeer().(FeefiFeeer)
	if !ok {
		return nil, operation.NewBaseReasonError("feeer is not feefifeeer")
	}
	if !feeer.feefier.Equal(fact.pool) {
		return nil, operation.NewBaseReasonError("pool is not registered, %q", fact.pool)
	}

	// prepare pool state
	// keep pool state
	st, err = existsState(StateKeyPool(fact.pool, fact.poolID), " feefi pool", getState)
	if err != nil {
		return nil, err
	}

	p, err := StatePoolValue(st)
	if err != nil {
		return nil, err
	}
	if _, found := p.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"]; found {
		fmt.Println("afte StatePoolValue p")
		fmt.Println(p.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"].outlay.Amount().Big())
	}

	_, found := p.users[fact.sender.String()]
	if !found {
		return nil, operation.NewBaseReasonError("sender not in pool users, %q", fact.sender)
	}

	nUsers := make(map[string]PoolUserBalance)
	for k := range p.users {
		nUsers[k] = p.users[k]
	}
	pl := NewPool(p.prevIncomeAmount.Currency(), p.prevOutlayAmount.Currency())
	pl = pl.SetIncomeBalance(p.prevIncomeAmount)
	pl = pl.SetOutlayBalance(p.prevOutlayAmount)
	npl := pl.WithUsers(nUsers)

	pool, err := UpdatePoolUserFromWithdraw(npl, fact.pool, fact.sender, fact.amounts, getState)
	if err != nil {
		return nil, operation.NewBaseReasonError("update pool users balance failed, %q", err)
	}
	if _, found := pool.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"]; found {
		fmt.Println("after UpdatePoolUserFromWithdraw pool")
		fmt.Println(pool.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"].outlay.Amount().Big())
	}

	nst, err := setStatePoolValue(st, pool)
	if err != nil {
		return nil, err
	}
	// feefi pool state after update
	opp.fs = nst

	if required, err := opp.calculateFee(getState); err != nil {
		return nil, operation.NewBaseReasonErrorFromError(err)
	} else if pb, err := CheckEnoughBalanceWithdraw(fact.pool, fact.poolID, required, getState); err != nil {
		return nil, err
	} else {
		// required amount and fee
		opp.required = required
		// pool amount state before update
		opp.pb = pb
	}

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

	if err = checkFactSignsByState(fact.sender, opp.Signs(), getState); err != nil {
		return nil, errors.Wrap(err, "invalid signing")
	}

	return opp, nil
}

func (opp *PoolWithdrawsProcessor) Process( // nolint:dupl
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(PoolWithdrawsFact)
	var sts []state.State // nolint:prealloc
	// sender(withdrawer) account balance state after update
	// pool account balance state after update
	for k := range opp.required {
		rq := opp.required[k]
		amv := opp.pb[k].Sub(rq[0])
		sts = append(sts, opp.wb[k].Add(rq[0].Sub(rq[1])), amv.AddFee(rq[1]), opp.fs)
	}
	fmt.Println("dddddddddddddddddddddddddddd")
	return setState(fact.Hash(), sts...)
}

func (opp *PoolWithdrawsProcessor) Close() error {
	opp.cp = nil
	opp.PoolWithdraws = PoolWithdraws{}
	opp.wb = nil
	opp.pb = nil
	opp.fs = nil
	opp.required = nil

	poolWithdrawsProcessorPool.Put(opp)

	return nil
}

func (opp *PoolWithdrawsProcessor) calculateFee(getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
	fact := opp.Fact().(PoolWithdrawsFact)

	return CalculateWithdrawFee(opp.cp, fact, getState)
}

func CalculateWithdrawFee(cp *CurrencyPool, fact PoolWithdrawsFact, getState func(key string) (state.State, bool, error)) (map[currency.CurrencyID][2]currency.Big, error) {
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
			st, err := existsState(StateKeyPoolDesign(feeer.Feefier(), fact.poolID), "feefi design", getState)
			if err != nil {
				return nil, err
			}

			design, err := StatePoolDesignValue(st)
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
	if _, found := pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"]; found {
		fmt.Println("in function UpdatePoolUserFromWithdraw pl")
		fmt.Println(pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"].outlay.Amount().Big())
	}
	fmt.Println("======================================================")
	pool, nowIncomeBalance, nowOutlayBalance, err := CalculateRewardAndHold(pl, feefier, getState)
	if err != nil {
		return Pool{}, err
	}
	if _, found := pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"]; found {
		fmt.Println("in function UpdatePoolUserFromWithdraw after call CalculateRewardAndHold pool")
		fmt.Println(pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"].outlay.Amount().Big())
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
		fmt.Println("POOL #############################################################")
		fmt.Printf("	%s | Withdraw Amount : %v\n", am.Currency(), am.Big())
	}
	// update pool previous balance
	pool.prevIncomeAmount = nowIncomeBalance.Amount()
	pool.prevOutlayAmount = nowOutlayBalance.Amount()
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | Updated Amount : %v, %s | Updated Amount : %v\n", nowIncomeBalance.Amount().Currency(), pool.prevIncomeAmount.Big(), nowOutlayBalance.Amount().Currency(), pool.prevOutlayAmount.Big())
	fmt.Println("")
	fmt.Println("==================================================================")
	return pool, nil
}

func CheckEnoughBalanceWithdraw(
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
