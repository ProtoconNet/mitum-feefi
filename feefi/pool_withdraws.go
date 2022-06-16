package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	PoolWithdrawsFactType   = hint.Type("mitum-feefi-pool-withdraw-operation-fact")
	PoolWithdrawsFactHint   = hint.NewHint(PoolWithdrawsFactType, "v0.0.1")
	PoolWithdrawsFactHinter = PoolWithdrawsFact{BaseHinter: hint.NewBaseHinter(PoolWithdrawsFactHint)}
	PoolWithdrawsType       = hint.Type("mitum-feefi-pool-withdraw-operation")
	PoolWithdrawsHint       = hint.NewHint(PoolWithdrawsType, "v0.0.1")
	PoolWithdrawsHinter     = PoolWithdraws{BaseOperation: operationHinter(PoolWithdrawsHint)}
)

var (
	MaxWithdrawsAmount uint = 10
)

type PoolWithdrawsFact struct {
	hint.BaseHinter
	h       valuehash.Hash
	token   []byte
	sender  base.Address
	pool    base.Address
	poolID  extensioncurrency.ContractID
	amounts []currency.Amount
}

func NewPoolWithdrawsFact(token []byte, sender base.Address, pool base.Address, id extensioncurrency.ContractID, amounts []currency.Amount) PoolWithdrawsFact {
	fact := PoolWithdrawsFact{
		BaseHinter: hint.NewBaseHinter(PoolWithdrawsFactHint),
		token:      token,
		sender:     sender,
		pool:       pool,
		poolID:     id,
		amounts:    amounts,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func (fact PoolWithdrawsFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact PoolWithdrawsFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact PoolWithdrawsFact) Token() []byte {
	return fact.token
}

func (fact PoolWithdrawsFact) Bytes() []byte {
	length := 4
	bs := make([][]byte, len(fact.amounts)+length)
	bs[0] = fact.token
	bs[1] = fact.sender.Bytes()
	bs[2] = fact.pool.Bytes()
	bs[3] = fact.poolID.Bytes()
	for i := range fact.amounts {
		bs[i+length] = fact.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (fact PoolWithdrawsFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.amounts); n < 1 {
		return isvalid.InvalidError.Errorf("empty amounts")
	} else if n > int(MaxWithdrawsAmount) {
		return isvalid.InvalidError.Errorf("amounts, %d over max, %d", n, MaxWithdrawsAmount)
	}

	if err := isvalid.Check(nil, false, fact.sender, fact.pool, fact.poolID); err != nil {
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

func (fact PoolWithdrawsFact) Sender() base.Address {
	return fact.sender
}

func (fact PoolWithdrawsFact) Pool() base.Address {
	return fact.pool
}

func (fact PoolWithdrawsFact) PoolID() extensioncurrency.ContractID {
	return fact.poolID
}

func (fact PoolWithdrawsFact) Amounts() []currency.Amount {
	return fact.amounts
}

func (fact PoolWithdrawsFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2)
	as[0] = fact.sender
	as[1] = fact.pool

	return as, nil
}

type PoolWithdraws struct {
	currency.BaseOperation
}

func NewPoolWithdraws(
	fact PoolWithdrawsFact,
	fs []base.FactSign,
	memo string,
) (PoolWithdraws, error) {
	bo, err := currency.NewBaseOperationFromFact(PoolWithdrawsHint, fact, fs, memo)
	if err != nil {
		return PoolWithdraws{}, err
	}
	return PoolWithdraws{BaseOperation: bo}, nil
}
