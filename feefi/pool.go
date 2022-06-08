package feefi // nolint: dupl, revive

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
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
	income extensioncurrency.AmountValue
	outlay extensioncurrency.AmountValue
}

func NewPoolUserBalance(income, outlay extensioncurrency.AmountValue) PoolUserBalance {
	fp := PoolUserBalance{
		BaseHinter: hint.NewBaseHinter(PoolUserBalanceHint),
		income:     income,
		outlay:     outlay,
	}
	return fp
}

func (pl PoolUserBalance) Income() extensioncurrency.AmountValue {
	return pl.income
}

func (pl PoolUserBalance) Outlay() extensioncurrency.AmountValue {
	return pl.outlay
}

func (pl *PoolUserBalance) AddIncome(v currency.Amount) error {
	if v.Currency() != pl.income.Amount().Currency() {
		return errors.Errorf("currency, %q not matched with pool income currency", v.Currency())
	}
	k := pl.income.Amount().Big().Add(v.Big())
	id := extensioncurrency.ContractID(pl.income.ID())
	pl.income = extensioncurrency.NewAmountValue(k, v.Currency(), id)

	return nil
}

func (pl *PoolUserBalance) SubIncome(v currency.Amount) error {
	if v.Currency() != pl.income.Amount().Currency() {
		return errors.Errorf("currency, %q not matched with pool income currency", v.Currency())
	}
	k := pl.income.Amount().Big().Sub(v.Big())
	if !k.OverNil() {
		return errors.Errorf("under zero value, %q after SubIncome", k)
	}
	id := extensioncurrency.ContractID(pl.income.ID())
	pl.income = extensioncurrency.NewAmountValue(k, v.Currency(), id)

	return nil
}

func (pl *PoolUserBalance) AddOutlay(v currency.Amount) error {
	if v.Currency() != pl.outlay.Amount().Currency() {
		return errors.Errorf("currency, %q not matched with pool outlay currency", v.Currency())
	}
	k := pl.outlay.Amount().Big().Add(v.Big())
	id := extensioncurrency.ContractID(pl.income.ID())
	pl.outlay = extensioncurrency.NewAmountValue(k, v.Currency(), id)

	return nil
}

func (pl *PoolUserBalance) SubOutlay(v currency.Amount) error {
	if v.Currency() != pl.outlay.Amount().Currency() {
		return errors.Errorf("currency, %q not matched with pool outlay currency", v.Currency())
	}
	k := pl.outlay.Amount().Big().Sub(v.Big())
	if !k.OverNil() {
		return errors.Errorf("under zero value, %q after SubOutlay", k)
	}
	id := extensioncurrency.ContractID(pl.income.ID())
	pl.outlay = extensioncurrency.NewAmountValue(k, v.Currency(), id)

	return nil
}

var (
	PoolType   = hint.Type("mitum-feefi-pool")
	PoolHint   = hint.NewHint(PoolType, "v0.0.1")
	PoolHinter = Pool{BaseHinter: hint.NewBaseHinter(PoolHint)}
)

type Pool struct {
	hint.BaseHinter
	users            map[string]PoolUserBalance
	prevIncomeAmount currency.Amount
	prevOutlayAmount currency.Amount
}

func NewPool(incomeCID, outlayCID currency.CurrencyID) Pool {
	users := make(map[string]PoolUserBalance, MaxFeeFiPoolLength)
	prvIncomeAmount := currency.NewAmount(currency.ZeroBig, incomeCID)
	prvOutlayAmount := currency.NewAmount(currency.ZeroBig, outlayCID)
	fp := Pool{
		BaseHinter:       hint.NewBaseHinter(PoolHint),
		users:            users,
		prevIncomeAmount: prvIncomeAmount,
		prevOutlayAmount: prvOutlayAmount,
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
	bs[1] = fp.prevIncomeAmount.Bytes()
	bs[2] = fp.prevOutlayAmount.Bytes()
	return util.ConcatBytesSlice(bs...)
}

func (fp Pool) Hash() valuehash.Hash {
	return fp.GenerateHash()
}

func (fp Pool) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fp.Bytes())
}

func (fp Pool) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, fp.prevIncomeAmount, fp.prevOutlayAmount)
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
	return fp.prevIncomeAmount
}

func (fp Pool) OutlayBalance() currency.Amount {
	return fp.prevOutlayAmount
}

func (fp Pool) Equal(b Pool) bool {
	ausers, _ := json.Marshal(fp.users)
	busers, _ := json.Marshal(b.users)
	for i := range ausers {
		if ausers[i] != busers[i] {
			return false
		}
	}

	if fp.prevIncomeAmount.Equal(b.prevIncomeAmount) {
		return false
	}

	if fp.prevOutlayAmount.Equal(b.prevOutlayAmount) {
		return false
	}

	return true
}

type PoolUsers struct {
	hint.BaseHinter
	users            map[string]PoolUserBalance
	prevIncomeAmount currency.Amount
	prevOutlayAmount currency.Amount
}

func NewPoolUsers(incomeCID, outlayCID currency.CurrencyID) PoolUsers {
	users := make(map[string]PoolUserBalance, MaxFeeFiPoolLength)
	prvIncomeAmount := currency.NewAmount(currency.ZeroBig, incomeCID)
	prvOutlayAmount := currency.NewAmount(currency.ZeroBig, outlayCID)
	fp := PoolUsers{
		BaseHinter:       hint.NewBaseHinter(PoolHint),
		users:            users,
		prevIncomeAmount: prvIncomeAmount,
		prevOutlayAmount: prvOutlayAmount,
	}
	return fp
}

func MustNewPoolUsers(incomeCID, outlayCID currency.CurrencyID) (PoolUsers, error) {
	fp := NewPoolUsers(incomeCID, outlayCID)
	err := fp.IsValid(nil)
	if err != nil {
		return PoolUsers{}, err
	}
	return fp, nil
}

func (fp PoolUsers) Bytes() []byte {
	length := 3
	bs := make([][]byte, length)

	if fp.users != nil {
		users, _ := json.Marshal(fp.users)
		bs[0] = valuehash.NewSHA256(users).Bytes()
	}
	bs[1] = fp.prevIncomeAmount.Bytes()
	bs[2] = fp.prevOutlayAmount.Bytes()
	return util.ConcatBytesSlice(bs...)
}

func (fp PoolUsers) Hash() valuehash.Hash {
	return fp.GenerateHash()
}

func (fp PoolUsers) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fp.Bytes())
}

func (fp PoolUsers) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, fp.prevIncomeAmount, fp.prevOutlayAmount)
	if err != nil {
		return err
	}
	return nil
}

func (fp PoolUsers) User(a base.Address) (PoolUserBalance, bool) {
	v, ok := fp.users[a.String()]
	if !ok {
		return PoolUserBalance{}, false
	}
	return v, true
}

func (fp PoolUsers) Users() map[string]PoolUserBalance {
	return fp.users
}

func (fp PoolUsers) IncomeBalance() currency.Amount {
	return fp.prevIncomeAmount
}

func (fp PoolUsers) OutlayBalance() currency.Amount {
	return fp.prevOutlayAmount
}

func (fp PoolUsers) Equal(b PoolUsers) bool {
	ausers, _ := json.Marshal(fp.users)
	busers, _ := json.Marshal(b.users)
	for i := range ausers {
		if ausers[i] != busers[i] {
			return false
		}
	}

	if fp.prevIncomeAmount.Equal(b.prevIncomeAmount) {
		return false
	}

	if fp.prevOutlayAmount.Equal(b.prevOutlayAmount) {
		return false
	}

	return true
}
