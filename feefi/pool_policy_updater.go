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
	PoolPolicyUpdaterFactType   = hint.Type("mitum-feefi-pool-policy-updater-operation-fact")
	PoolPolicyUpdaterFactHint   = hint.NewHint(PoolPolicyUpdaterFactType, "v0.0.1")
	PoolPolicyUpdaterFactHinter = PoolPolicyUpdaterFact{BaseHinter: hint.NewBaseHinter(PoolPolicyUpdaterFactHint)}
	PoolPolicyUpdaterType       = hint.Type("mitum-feefi-pool-policy-updater-operation")
	PoolPolicyUpdaterHint       = hint.NewHint(PoolPolicyUpdaterType, "v0.0.1")
	PoolPolicyUpdaterHinter     = PoolPolicyUpdater{BaseOperation: operationHinter(PoolPolicyUpdaterHint)}
)

type PoolPolicyUpdaterFact struct {
	hint.BaseHinter
	h        valuehash.Hash
	token    []byte
	sender   base.Address
	target   base.Address
	fee      currency.Amount
	poolID   extensioncurrency.ContractID
	currency currency.CurrencyID
}

func NewPoolPolicyUpdaterFact(token []byte, sender, target base.Address, fee currency.Amount, poolid extensioncurrency.ContractID, currency currency.CurrencyID) PoolPolicyUpdaterFact {
	fact := PoolPolicyUpdaterFact{
		BaseHinter: hint.NewBaseHinter(PoolPolicyUpdaterFactHint),
		token:      token,
		sender:     sender,
		target:     target,
		fee:        fee,
		poolID:     poolid,
		currency:   currency,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func MustNewPoolPolicyUpdaterFact(token []byte, sender, target base.Address, fee currency.Amount, poolid extensioncurrency.ContractID, currency currency.CurrencyID) PoolPolicyUpdaterFact {
	fact := NewPoolPolicyUpdaterFact(token, sender, target, fee, poolid, currency)
	err := fact.IsValid(nil)
	if err != nil {
		panic(err)
	}

	return fact
}

func (fact PoolPolicyUpdaterFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact PoolPolicyUpdaterFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact PoolPolicyUpdaterFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		fact.target.Bytes(),
		fact.fee.Bytes(),
		fact.poolID.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact PoolPolicyUpdaterFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	return isvalid.Check(nil, false,
		fact.sender,
		fact.target,
		fact.fee,
		fact.poolID,
		fact.currency,
	)
}

func (fact PoolPolicyUpdaterFact) Token() []byte {
	return fact.token
}

func (fact PoolPolicyUpdaterFact) Sender() base.Address {
	return fact.sender
}

func (fact PoolPolicyUpdaterFact) Target() base.Address {
	return fact.target
}

func (fact PoolPolicyUpdaterFact) Fee() currency.Amount {
	return fact.fee
}

func (fact PoolPolicyUpdaterFact) PoolID() extensioncurrency.ContractID {
	return fact.poolID
}

func (fact PoolPolicyUpdaterFact) Currency() currency.CurrencyID {
	return fact.currency
}

func (fact PoolPolicyUpdaterFact) Addresses() ([]base.Address, error) {
	return []base.Address{fact.sender, fact.target}, nil
}

type PoolPolicyUpdater struct {
	currency.BaseOperation
}

func NewPoolPolicyUpdater(fact PoolPolicyUpdaterFact, fs []base.FactSign, memo string) (PoolPolicyUpdater, error) {
	bo, err := currency.NewBaseOperationFromFact(PoolPolicyUpdaterHint, fact, fs, memo)
	if err != nil {
		return PoolPolicyUpdater{}, err
	}

	return PoolPolicyUpdater{BaseOperation: bo}, nil
}
