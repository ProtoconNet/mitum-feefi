package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	WithdrawsFactType   = hint.Type("mitum-feefi-withdraw-operation-fact")
	WithdrawsFactHint   = hint.NewHint(WithdrawsFactType, "v0.0.1")
	WithdrawsFactHinter = WithdrawsFact{BaseHinter: hint.NewBaseHinter(WithdrawsFactHint)}
	WithdrawsType       = hint.Type("mitum-feefi-withdraw-operation")
	WithdrawsHint       = hint.NewHint(WithdrawsType, "v0.0.1")
	WithdrawsHinter     = Withdraws{BaseOperation: operationHinter(WithdrawsHint)}
)

var (
	MaxWithdrawsItems  uint = 10
	MaxWithdrawsAmount uint = 10
)

type WithdrawsItem interface {
	hint.Hinter
	isvalid.IsValider
	currency.AmountsItem
	Bytes() []byte
	Target() base.Address
	PoolCID() currency.CurrencyID
	Rebuild() WithdrawsItem
}

type WithdrawsFact struct {
	hint.BaseHinter
	h       valuehash.Hash
	token   []byte
	sender  base.Address
	pool    base.Address
	poolCID currency.CurrencyID
	amounts []currency.Amount
}

func NewWithdrawsFact(token []byte, sender base.Address, pool base.Address, poolCID currency.CurrencyID, amounts []currency.Amount) WithdrawsFact {
	fact := WithdrawsFact{
		BaseHinter: hint.NewBaseHinter(WithdrawsFactHint),
		token:      token,
		sender:     sender,
		pool:       pool,
		poolCID:    poolCID,
		amounts:    amounts,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func (fact WithdrawsFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact WithdrawsFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact WithdrawsFact) Token() []byte {
	return fact.token
}

func (fact WithdrawsFact) Bytes() []byte {
	length := 4
	bs := make([][]byte, len(fact.amounts)+length)
	bs[0] = fact.token
	bs[1] = fact.sender.Bytes()
	bs[2] = fact.pool.Bytes()
	bs[3] = fact.poolCID.Bytes()
	for i := range fact.amounts {
		bs[i+length] = fact.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (fact WithdrawsFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.amounts); n < 1 {
		return isvalid.InvalidError.Errorf("empty amounts")
	} else if n > int(MaxWithdrawsAmount) {
		return isvalid.InvalidError.Errorf("amounts, %d over max, %d", n, MaxWithdrawsAmount)
	}

	if err := isvalid.Check(nil, false, fact.sender, fact.pool, fact.poolCID); err != nil {
		return err
	}

	founds := map[currency.CurrencyID]struct{}{}
	for i := range fact.amounts {
		am := fact.amounts[i]
		_, found := founds[am.Currency()]
		if found {
			return isvalid.InvalidError.Errorf("duplicated currency found, %q", am.Currency())
		}
		founds[am.Currency()] = struct{}{}

		if err := am.IsValid(nil); err != nil {
			return err
		} else if !am.Big().OverZero() {
			return isvalid.InvalidError.Errorf("amount should be over zero")
		}
	}

	return nil
}

func (fact WithdrawsFact) Sender() base.Address {
	return fact.sender
}

func (fact WithdrawsFact) Pool() base.Address {
	return fact.pool
}

func (fact WithdrawsFact) PoolCID() currency.CurrencyID {
	return fact.poolCID
}

func (fact WithdrawsFact) Amounts() []currency.Amount {
	return fact.amounts
}

func (fact WithdrawsFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2)
	as[0] = fact.sender
	as[1] = fact.pool

	return as, nil
}

type Withdraws struct {
	currency.BaseOperation
}

func NewWithdraws(
	fact WithdrawsFact,
	fs []base.FactSign,
	memo string,
) (Withdraws, error) {
	bo, err := currency.NewBaseOperationFromFact(WithdrawsHint, fact, fs, memo)
	if err != nil {
		return Withdraws{}, err
	}
	return Withdraws{BaseOperation: bo}, nil
}
