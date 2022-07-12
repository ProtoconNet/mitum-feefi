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
	PoolRegisterFactType   = hint.Type("mitum-feefi-pool-register-operation-fact")
	PoolRegisterFactHint   = hint.NewHint(PoolRegisterFactType, "v0.0.1")
	PoolRegisterFactHinter = PoolRegisterFact{BaseHinter: hint.NewBaseHinter(PoolRegisterFactHint)}
	PoolRegisterType       = hint.Type("mitum-feefi-pool-register-operation")
	PoolRegisterHint       = hint.NewHint(PoolRegisterType, "v0.0.1")
	PoolRegisterHinter     = PoolRegister{BaseOperation: operationHinter(PoolRegisterHint)}
)

type PoolRegisterFact struct {
	hint.BaseHinter
	h          valuehash.Hash
	token      []byte
	sender     base.Address
	target     base.Address
	initialFee currency.Amount
	incomeCID  currency.CurrencyID
	outlayCID  currency.CurrencyID
	currency   currency.CurrencyID
}

func NewPoolRegisterFact(token []byte, sender, target base.Address, fee currency.Amount, incomeingcid, outgoingcid, currency currency.CurrencyID) PoolRegisterFact {
	fact := PoolRegisterFact{
		BaseHinter: hint.NewBaseHinter(PoolRegisterFactHint),
		token:      token,
		sender:     sender,
		target:     target,
		initialFee: fee,
		incomeCID:  incomeingcid,
		outlayCID:  outgoingcid,
		currency:   currency,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func MustNewPoolRegisterFact(token []byte, sender, target base.Address, fee currency.Amount, incomeingcid, outgoing, currency currency.CurrencyID) PoolRegisterFact {
	fact := NewPoolRegisterFact(token, sender, target, fee, incomeingcid, outgoing, currency)
	err := fact.IsValid(nil)
	if err != nil {
		panic(err)
	}

	return fact
}

func (fact PoolRegisterFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact PoolRegisterFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact PoolRegisterFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		fact.target.Bytes(),
		fact.initialFee.Bytes(),
		fact.incomeCID.Bytes(),
		fact.outlayCID.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact PoolRegisterFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}
	if fact.initialFee.Currency() != fact.incomeCID {
		return isvalid.InvalidError.Errorf("initialFee currency is different with incomeCID")
	}

	return isvalid.Check(nil, false,
		fact.sender,
		fact.target,
		fact.initialFee,
		fact.incomeCID,
		fact.outlayCID,
		fact.currency,
	)
}

func (fact PoolRegisterFact) Token() []byte {
	return fact.token
}

func (fact PoolRegisterFact) Sender() base.Address {
	return fact.sender
}

func (fact PoolRegisterFact) Target() base.Address {
	return fact.target
}

func (fact PoolRegisterFact) InitialFee() currency.Amount {
	return fact.initialFee
}

func (fact PoolRegisterFact) IncomeCID() currency.CurrencyID {
	return fact.incomeCID
}

func (fact PoolRegisterFact) OutlayCID() currency.CurrencyID {
	return fact.outlayCID
}

func (fact PoolRegisterFact) Currency() currency.CurrencyID {
	return fact.currency
}

func (fact PoolRegisterFact) Addresses() ([]base.Address, error) {
	return []base.Address{fact.sender, fact.target}, nil
}

type PoolRegister struct {
	currency.BaseOperation
}

func NewPoolRegister(fact PoolRegisterFact, fs []base.FactSign, memo string) (PoolRegister, error) {
	bo, err := currency.NewBaseOperationFromFact(PoolRegisterHint, fact, fs, memo)
	if err != nil {
		return PoolRegister{}, err
	}

	return PoolRegister{BaseOperation: bo}, nil
}
