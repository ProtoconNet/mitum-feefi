package feefi // nolint: dupl, revive

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var MaxFeeFiPoolLength = 5000

var (
	PoolUserBalanceType   = hint.Type("mitum-feefi-pool-user")
	PoolUserBalanceHint   = hint.NewHint(PoolUserBalanceType, "v0.0.1")
	PoolUserBalanceHinter = PoolUserBalance{BaseHinter: hint.NewBaseHinter(PoolUserBalanceHint)}
)

type PoolUserBalance struct {
	hint.BaseHinter
	incomeAmount currency.Amount
	outlayAmount currency.Amount
}

func NewPoolUserBalance(income, outlay currency.Amount) PoolUserBalance {
	fp := PoolUserBalance{
		BaseHinter:   hint.NewBaseHinter(PoolUserBalanceHint),
		incomeAmount: income,
		outlayAmount: outlay,
	}
	return fp
}

func (pl *PoolUserBalance) AddIncome(v currency.Amount) error {
	if v.Currency() != pl.incomeAmount.Currency() {
		return errors.Errorf("currency, %q not matched with pool income currency", v.Currency())
	}
	k := pl.incomeAmount.Big().Add(v.Big())
	pl.incomeAmount = currency.NewAmount(k, v.Currency())

	return nil
}

func (pl *PoolUserBalance) SubIncome(v currency.Amount) error {
	if v.Currency() != pl.incomeAmount.Currency() {
		return errors.Errorf("currency, %q not matched with pool income currency", v.Currency())
	}
	k := pl.incomeAmount.Big().Sub(v.Big())
	if !k.OverNil() {
		return errors.Errorf("under zero value, %q after SubIncome", k)
	}
	pl.incomeAmount = currency.NewAmount(k, v.Currency())

	return nil
}

func (pl *PoolUserBalance) AddOutlay(v currency.Amount) error {
	if v.Currency() != pl.outlayAmount.Currency() {
		return errors.Errorf("currency, %q not matched with pool outlay currency", v.Currency())
	}
	k := pl.outlayAmount.Big().Add(v.Big())
	pl.outlayAmount = currency.NewAmount(k, v.Currency())

	return nil
}

func (pl *PoolUserBalance) SubOutlay(v currency.Amount) error {
	if v.Currency() != pl.outlayAmount.Currency() {
		return errors.Errorf("currency, %q not matched with pool outlay currency", v.Currency())
	}
	k := pl.outlayAmount.Big().Sub(v.Big())
	if !k.OverNil() {
		return errors.Errorf("under zero value, %q after SubOutlay", k)
	}
	pl.outlayAmount = currency.NewAmount(k, v.Currency())

	return nil
}

var (
	PoolType   = hint.Type("mitum-feefi-pool")
	PoolHint   = hint.NewHint(PoolType, "v0.0.1")
	PoolHinter = Pool{BaseHinter: hint.NewBaseHinter(PoolHint)}
)

type Pool struct {
	hint.BaseHinter
	users             map[string]PoolUserBalance
	prevIncomeBalance currency.Amount
	prevOutlayBalance currency.Amount
}

func NewPool(incomeCID, outlayCID currency.CurrencyID) Pool {
	users := make(map[string]PoolUserBalance, MaxFeeFiPoolLength)
	prvIncomeBalance := currency.NewAmount(currency.ZeroBig, incomeCID)
	prvOutlayBalance := currency.NewAmount(currency.ZeroBig, outlayCID)
	fp := Pool{
		BaseHinter:        hint.NewBaseHinter(PoolHint),
		users:             users,
		prevIncomeBalance: prvIncomeBalance,
		prevOutlayBalance: prvOutlayBalance,
	}
	return fp
}

func MustNewPool(incomeCID, outlayCID currency.CurrencyID) (Pool, error) {
	fp := NewPool(incomeCID, outlayCID)
	err := fp.IsValid(nil)
	if err != nil {
		return Pool{}, err
	}
	return fp, nil
}

func (fp Pool) Bytes() []byte {
	length := 3
	bs := make([][]byte, length)

	if fp.users != nil {
		users, _ := json.Marshal(fp.users)
		bs[0] = valuehash.NewSHA256(users).Bytes()
	}
	bs[1] = fp.prevIncomeBalance.Bytes()
	bs[2] = fp.prevOutlayBalance.Bytes()
	return util.ConcatBytesSlice(bs...)
}

func (fp Pool) Hash() valuehash.Hash {
	return fp.GenerateHash()
}

func (fp Pool) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fp.Bytes())
}

func (fp Pool) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, fp.prevIncomeBalance, fp.prevOutlayBalance)
	if err != nil {
		return err
	}
	return nil
}

func (fp Pool) User(a base.Address) (PoolUserBalance, bool) {
	v, ok := fp.users[a.String()]
	if !ok {
		return PoolUserBalance{}, false
	}
	return v, true
}

func (fp Pool) Users() map[string]PoolUserBalance {
	return fp.users
}

func (fp Pool) IncomeBalance() currency.Amount {
	return fp.prevIncomeBalance
}

func (fp Pool) OutlayBalance() currency.Amount {
	return fp.prevOutlayBalance
}

func (fp Pool) Equal(b Pool) bool {
	ausers, _ := json.Marshal(fp.users)
	busers, _ := json.Marshal(b.users)
	for i := range ausers {
		if ausers[i] != busers[i] {
			return false
		}
	}

	if fp.prevIncomeBalance.Equal(b.prevIncomeBalance) {
		return false
	}

	if fp.prevOutlayBalance.Equal(b.prevOutlayBalance) {
		return false
	}

	return true
}
