package extension

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	DeactivateFactType   = hint.Type("mitum-currency-contract-account-deactivate-operation-fact")
	DeactivateFactHint   = hint.NewHint(DeactivateFactType, "v0.0.1")
	DeactivateFactHinter = DeactivateFact{BaseHinter: hint.NewBaseHinter(DeactivateFactHint)}
	DeactivateType       = hint.Type("mitum-currency-contract-account-deactivate-operation")
	DeactivateHint       = hint.NewHint(DeactivateType, "v0.0.1")
	DeactivateHinter     = Deactivate{BaseOperation: operationHinter(DeactivateHint)}
)

type DeactivateFact struct {
	hint.BaseHinter
	h        valuehash.Hash
	token    []byte
	sender   base.Address
	target   base.Address
	currency currency.CurrencyID
}

func NewDeactivateFact(token []byte, sender, target base.Address, currency currency.CurrencyID) DeactivateFact {
	fact := DeactivateFact{
		BaseHinter: hint.NewBaseHinter(DeactivateFactHint),
		token:      token,
		sender:     sender,
		target:     target,
		currency:   currency,
	}
	fact.h = fact.GenerateHash()

	return fact
}

func (fact DeactivateFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact DeactivateFact) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact DeactivateFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		fact.target.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact DeactivateFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	return isvalid.Check(nil, false,
		fact.sender,
		fact.target,
		fact.currency,
	)
}

func (fact DeactivateFact) Token() []byte {
	return fact.token
}

func (fact DeactivateFact) Sender() base.Address {
	return fact.sender
}

func (fact DeactivateFact) Target() base.Address {
	return fact.target
}

func (fact DeactivateFact) Currency() currency.CurrencyID {
	return fact.currency
}

func (fact DeactivateFact) Addresses() ([]base.Address, error) {
	return []base.Address{fact.sender, fact.target}, nil
}

type Deactivate struct {
	currency.BaseOperation
}

func NewDeactivate(fact DeactivateFact, fs []base.FactSign, memo string) (Deactivate, error) {
	bo, err := currency.NewBaseOperationFromFact(DeactivateHint, fact, fs, memo)
	if err != nil {
		return Deactivate{}, err
	}

	return Deactivate{BaseOperation: bo}, nil
}
