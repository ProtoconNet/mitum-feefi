package feefi

import (
	"fmt"
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

type DepositsProcessor struct {
	cp *CurrencyPool
	Deposit
	sb       currency.AmountState
	fs       state.State
	rb       extensioncurrency.AmountState
	required [2]currency.Big
}

func NewDepositsProcessor(cp *CurrencyPool) currency.GetNewProcessor {
	return func(op state.Processor) (state.Processor, error) {
		i, ok := op.(Deposit)
		if !ok {
			return nil, operation.NewBaseReasonError("not Deposits, %T", op)
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
	am := fact.amount
	if opp.cp != nil {
		if !opp.cp.Exists(am.Currency()) {
			return nil, operation.NewBaseReasonError("currency not registered, %q", am.Currency())
		}
	}

	// check pool allowed to receive currency
	_, err = existsState(StateKeyPoolDesign(fact.pool, fact.poolID), " feefi pool design", getState)
	if err != nil {
		return nil, err
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

	nUsers := make(map[string]PoolUserBalance)
	for k := range p.users {
		nUsers[k] = p.users[k]
	}
	pl := NewPool(p.prevIncomeAmount.Currency(), p.prevOutlayAmount.Currency())
	pl = pl.SetIncomeBalance(p.prevIncomeAmount)
	pl = pl.SetOutlayBalance(p.prevOutlayAmount)
	npl := pl.WithUsers(nUsers)
	pool, err := UpdatePoolUserFromDeposit(npl, fact.pool, fact.sender, am, getState)
	if err != nil {
		return nil, operation.NewBaseReasonError("Update pool user balance failed, %q", err)
	}
	nst, err := setStatePoolValue(st, pool)
	if err != nil {
		return nil, err
	}
	// feefi pool state after update
	opp.fs = nst

	if required, err := opp.calculateFee(getState); err != nil {
		return nil, operation.NewBaseReasonError("failed to calculate fee: %w", err)
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
	// sender account balance state after update
	var sts []state.State // nolint:prealloc
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

func CalculateDepositFee(cp *CurrencyPool, fact DepositFact, getState func(key string) (state.State, bool, error)) ([2]currency.Big, error) {
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
		st, err := existsState(StateKeyPoolDesign(feeer.Feefier(), fact.poolID), "feefi pool design", getState)
		if err != nil {
			return [2]currency.Big{}, err
		}

		design, err := StatePoolDesignValue(st)
		if err != nil {
			return [2]currency.Big{}, err
		}
		if design.Policy().Fee().Currency() != am.Currency() {
			return [2]currency.Big{}, errors.Errorf("feefi design fee currency id, %q not matched with %q", design.Policy().Fee().Currency(), am.Currency())
		}
		k = design.Policy().Fee().Big()
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

func UpdatePoolUserFromDeposit(pl Pool, feefier base.Address, depositer base.Address, am currency.Amount, getState func(key string) (state.State, bool, error)) (Pool, error) {
	fmt.Println("==================================================================")
	fmt.Println("")
	var (
		pool             Pool
		nowIncomeBalance extensioncurrency.AmountValue
		nowOutlayBalance extensioncurrency.AmountValue
		err              error
	)
	pool, nowIncomeBalance, nowOutlayBalance, err = CalculateRewardAndHold(pl, feefier, getState)
	if err != nil {
		return Pool{}, err
	}
	var id extensioncurrency.ContractID
	id = extensioncurrency.ContractID(pool.prevIncomeAmount.Currency().String())
	// add deposit amount
	if userBalance, found := pool.users[depositer.String()]; !found {
		pool.users[depositer.String()] = NewPoolUserBalance(
			extensioncurrency.NewAmountValue(currency.ZeroBig,
				pool.prevIncomeAmount.Currency(),
				id),
			extensioncurrency.NewAmountValuefromAmount(am, id),
		)
	} else {
		userBalance.AddOutlay(am.Big())
		pool.users[depositer.String()] = userBalance
	}
	nowOutlayBalance, err = nowOutlayBalance.Add(am.Big())
	if err != nil {
		return Pool{}, err
	}

	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | Deposit Amount : %v\n", am.Currency(), am.Big())
	// update pool previous balance
	pool.prevIncomeAmount = nowIncomeBalance.Amount()
	pool.prevOutlayAmount = nowOutlayBalance.Amount()
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | Updated Amount : %v, %s | Updated Amount : %v\n", nowIncomeBalance.Amount().Currency(), pool.prevIncomeAmount.Big(), nowOutlayBalance.Amount().Currency(), pool.prevOutlayAmount.Big())
	fmt.Println("")
	fmt.Println("==================================================================")
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

func CalculateRewardAndHold(pl Pool, feefier base.Address, getState func(key string) (state.State, bool, error)) (Pool, extensioncurrency.AmountValue, extensioncurrency.AmountValue, error) {
	pool := Pool{}
	if _, found := pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"]; found {
		fmt.Println("in function CalculateRewardAndHold pl")
		fmt.Println(pl.users["8iRVFAPiHKaeznfN3CmNjtFtjYSPMPKLuL6qkaJz8RLumca"].outlay.Amount().Big())
	}
	pool = pl.WithUsers(pl.users)
	id := extensioncurrency.ContractID(pool.prevIncomeAmount.Currency().String())
	st, err := existsState(extensioncurrency.StateKeyBalance(feefier, id, pool.prevIncomeAmount.Currency(), StateKeyBalanceSuffix), " feefi pool balance", getState)
	if err != nil {
		return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
	}
	nowIncomeBalance, err := extensioncurrency.StateBalanceValue(st)
	if err != nil {
		return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
	}
	st, err = existsState(extensioncurrency.StateKeyBalance(feefier, id, pool.prevOutlayAmount.Currency(), StateKeyBalanceSuffix), " feefi pool balance", getState)
	if err != nil {
		return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
	}
	nowOutlayBalance, err := extensioncurrency.StateBalanceValue(st)
	if err != nil {
		return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
	}
	prvIncomeAmount := pool.prevIncomeAmount
	prvOutlayAmount := pool.prevOutlayAmount
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | previous amount : %v, now amount : %v\n", nowIncomeBalance.Amount().Currency(), prvIncomeAmount.Big(), nowIncomeBalance.Amount().Big())
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | previous amount : %v, now amount : %v\n", nowOutlayBalance.Amount().Currency(), prvOutlayAmount.Big(), nowOutlayBalance.Amount().Big())
	// difference between previous balance and current balance
	diffIncomeAmount := nowIncomeBalance.Amount().Big().Sub(prvIncomeAmount.Big())
	diffOutlayAmount := prvOutlayAmount.Big().Sub(nowOutlayBalance.Amount().Big())
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | diffAmount : %v, %s | diffAmount : %v\n", nowIncomeBalance.Amount().Currency(), diffIncomeAmount, nowOutlayBalance.Amount().Currency(), diffOutlayAmount)
	// sum of locking tokent
	depositTotal := prvOutlayAmount
	usersIncomeSum := currency.ZeroBig
	usersRewardSum := currency.ZeroBig
	bigUserRatioFloat := new(big.Float).SetUint64(0)
	bigUser := ""
	// calculate income and outlay of users
	if depositTotal.Big().OverZero() {
		for address := range pool.users {
			userBalance := pool.users[address]
			outlayFloat := new(big.Float).SetInt(userBalance.outlay.Amount().Big().Int)
			depositTotalFloat := new(big.Float).SetInt(depositTotal.Big().Int)
			userRatioFloat := new(big.Float).Quo(outlayFloat, depositTotalFloat)
			fmt.Println("USER #############################################################")
			fmt.Printf("	user : %s, %s ratio : %v\n", address[:7], nowOutlayBalance.Amount().Currency(), userRatioFloat)
			fmt.Println("USER #############################################################")
			fmt.Printf("	%s | previous amount : %v, %s | previous amount : %v\n", nowIncomeBalance.Amount().Currency(), userBalance.income.Amount().Big(), nowOutlayBalance.Amount().Currency(), userBalance.outlay.Amount().Big())
			if userRatioFloat.Cmp(bigUserRatioFloat) > 0 {
				bigUserRatioFloat = userRatioFloat
				bigUser = address
			}
			var shareIncomeBig, shareOutlayBig currency.Big

			if !diffIncomeAmount.IsZero() {
				diffIncomeFloat := new(big.Float).SetInt(diffIncomeAmount.Int)
				shareFloat := new(big.Float).Mul(userRatioFloat, diffIncomeFloat)
				shareInt, _ := shareFloat.Int(nil)
				shareIncomeBig = currency.NewBigFromBigInt(shareInt)
				err := userBalance.AddIncome(shareIncomeBig)
				if err != nil {
					return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
				}
			}
			if !diffOutlayAmount.IsZero() {
				// delta := big.NewFloat(-0.5)
				diffOutlayFloat := new(big.Float).SetInt(diffOutlayAmount.Int)
				shareFloat := new(big.Float).Mul(userRatioFloat, diffOutlayFloat)
				// nfcal := new(big.Float).Add(fcal, delta)
				// shareInt, _ := nfcal.Int(nil)
				shareInt, _ := shareFloat.Int(nil)
				shareOutlayBig = currency.NewBigFromBigInt(shareInt)
				err := userBalance.SubOutlay(shareOutlayBig)
				if err != nil {
					return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
				}
			}
			fmt.Println("USER #############################################################")
			fmt.Printf("	%s | change amount : %v, %s | change amount : %v\n", nowIncomeBalance.Amount().Currency(), shareIncomeBig, nowOutlayBalance.Amount().Currency(), shareOutlayBig)
			pool.users[address] = userBalance
			fmt.Println("USER #############################################################")
			fmt.Printf("	%s | updated amount : %v,  %s | updated amount : %v\n", nowIncomeBalance.Amount().Currency(), userBalance.income.Amount().Big(), nowOutlayBalance.Amount().Currency(), userBalance.outlay.Amount().Big())
			usersRewardSum = usersRewardSum.Add(userBalance.income.Amount().Big())
			usersIncomeSum = usersIncomeSum.Add(userBalance.outlay.Amount().Big())
		}
	}

	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | now balance : %v, now user sum : %v\n", nowIncomeBalance.Amount().Currency(), nowIncomeBalance.Amount().Big(), usersRewardSum)
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | now balance : %v, now user sum : %v\n", nowOutlayBalance.Amount().Currency(), nowOutlayBalance.Amount().Big(), usersIncomeSum)

	incomeRemainder := nowIncomeBalance.Amount().Big().Sub(usersRewardSum)
	outlayRemainder := usersIncomeSum.Sub(nowOutlayBalance.Amount().Big())
	fmt.Println("POOL #############################################################")
	fmt.Printf("	%s | user sum diff : %v, %s | user sum diff : %v\n", nowIncomeBalance.Amount().Currency(), incomeRemainder, nowOutlayBalance.Amount().Currency(), outlayRemainder)
	// fmt.Println("#############################################################")
	// fmt.Printf("big user : %v, big user sharehold ratio : %v\n", bigUser, bigUserRatioFloat)
	// fmt.Println("#############################################################")
	// fmt.Printf("big user income : %v, outlay : %v\n", pool.users[bigUser].income.Amount().Big(), pool.users[bigUser].outlay.Amount().Big())
	if len(bigUser) > 0 {
		if outlayRemainder.OverZero() {
			max := pool.users[bigUser]
			err = max.SubOutlay(outlayRemainder)
			if err != nil {
				return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
			}
			pool.users[bigUser] = max
		}
		if incomeRemainder.OverZero() {
			max := pool.users[bigUser]
			err = max.AddIncome(incomeRemainder)
			if err != nil {
				fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
				return Pool{}, extensioncurrency.AmountValue{}, extensioncurrency.AmountValue{}, err
			}
			pool.users[bigUser] = max
		}
	}
	// fmt.Println("#############################################################")
	// fmt.Printf("big user final income : %v, final outlay : %v\n", pool.users[bigUser].income.Amount().Big(), pool.users[bigUser].outlay.Amount().Big())

	return pool, nowIncomeBalance, nowOutlayBalance, nil
}
