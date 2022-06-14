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
	DepositFactType   = hint.Type("mitum-feefi-deposits-operation-fact")
	DepositFactHint   = hint.NewHint(DepositFactType, "v0.0.1")
	DepositFactHinter = DepositFact{BaseHinter: hint.NewBaseHinter(DepositFactHint)}
	DepositType       = hint.Type("mitum-feefi-deposits-operation")
	DepositHint       = hint.NewHint(DepositType, "v0.0.1")
	DepositHinter     = Deposit{BaseOperation: operationHinter(DepositHint)}
)

type DepositFact struct {
	hint.BaseHinter
	h      valuehash.Hash
	token  []byte
	sender base.Address
	pool   base.Address
	poolID extensioncurrency.ContractID
	amount currency.Amount
}

func NewDepositFact(token []byte, sender base.Address, pool base.Address, id extensioncurrency.ContractID, amount currency.Amount) DepositFact {
	fact := DepositFact{
		BaseHinter: hint.NewBaseHinter(DepositFactHint),
		token:      token,
		sender:     sender,
		pool:       pool,
		poolID:     id,
		amount:     amount,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func (fact DepositFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact DepositFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact DepositFact) Token() []byte {
	return fact.token
}

func (fact DepositFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		fact.pool.Bytes(),
		fact.poolID.Bytes(),
		fact.amount.Bytes(),
	)
}

func (fact DepositFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := isvalid.Check(nil, false, fact.sender, fact.pool, fact.poolID); err != nil {
		return err
	}

	am := fact.amount
	if err := am.IsValid(nil); err != nil {
		return err
	} else if !am.Big().OverZero() {
		return isvalid.InvalidError.Errorf("amount should be over zero")
	}

	return nil
}

func (fact DepositFact) Sender() base.Address {
	return fact.sender
}

func (fact DepositFact) PoolID() extensioncurrency.ContractID {
	return fact.poolID
}

func (fact DepositFact) Pool() base.Address {
	return fact.pool
}

func (fact DepositFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 2)
	as[0] = fact.sender
	as[1] = fact.pool

	return as, nil
}

type Deposit struct {
	currency.BaseOperation
}

func NewDeposit(
	fact DepositFact,
	fs []base.FactSign,
	memo string,
) (Deposit, error) {
	bo, err := currency.NewBaseOperationFromFact(DepositHint, fact, fs, memo)
	if err != nil {
		return Deposit{}, err
	}
	return Deposit{BaseOperation: bo}, nil
}
