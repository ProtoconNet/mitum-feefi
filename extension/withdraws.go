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
	WithdrawsFactType   = hint.Type("mitum-currency-contract-account-withdraw-operation-fact")
	WithdrawsFactHint   = hint.NewHint(WithdrawsFactType, "v0.0.1")
	WithdrawsFactHinter = WithdrawsFact{BaseHinter: hint.NewBaseHinter(WithdrawsFactHint)}
	WithdrawsType       = hint.Type("mitum-currency-contract-account-withdraw-operation")
	WithdrawsHint       = hint.NewHint(WithdrawsType, "v0.0.1")
	WithdrawsHinter     = Withdraws{BaseOperation: operationHinter(WithdrawsHint)}
)

var MaxWithdrawsItems uint = 10

type WithdrawsItem interface {
	hint.Hinter
	isvalid.IsValider
	AmountsItem
	Bytes() []byte
	Target() base.Address
	Rebuild() WithdrawsItem
}

type WithdrawsFact struct {
	hint.BaseHinter
	h      valuehash.Hash
	token  []byte
	sender base.Address
	items  []WithdrawsItem
}

func NewWithdrawsFact(token []byte, sender base.Address, items []WithdrawsItem) WithdrawsFact {
	fact := WithdrawsFact{
		BaseHinter: hint.NewBaseHinter(WithdrawsFactHint),
		token:      token,
		sender:     sender,
		items:      items,
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
	its := make([][]byte, len(fact.items))
	for i := range fact.items {
		its[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.token,
		fact.sender.Bytes(),
		util.ConcatBytesSlice(its...),
	)
}

func (fact WithdrawsFact) IsValid(b []byte) error {
	if err := currency.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.items); n < 1 {
		return isvalid.InvalidError.Errorf("empty items")
	} else if n > int(MaxWithdrawsItems) {
		return isvalid.InvalidError.Errorf("items, %d over max, %d", n, MaxWithdrawsItems)
	}

	if err := isvalid.Check(nil, false, fact.sender); err != nil {
		return err
	}

	foundTargets := map[string]struct{}{}
	for i := range fact.items {
		it := fact.items[i]
		if err := isvalid.Check(nil, false, it); err != nil {
			return isvalid.InvalidError.Errorf("invalid item found: %w", err)
		}

		k := it.Target().String()
		switch _, found := foundTargets[k]; {
		case found:
			return isvalid.InvalidError.Errorf("duplicated target found, %s", it.Target())
		case fact.sender.Equal(it.Target()):
			return isvalid.InvalidError.Errorf("receiver is same with sender, %q", fact.sender)
		default:
			foundTargets[k] = struct{}{}
		}
	}

	return nil
}

func (fact WithdrawsFact) Sender() base.Address {
	return fact.sender
}

func (fact WithdrawsFact) Items() []WithdrawsItem {
	return fact.items
}

func (fact WithdrawsFact) Rebuild() WithdrawsFact {
	items := make([]WithdrawsItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.h = fact.GenerateHash()

	return fact
}

func (fact WithdrawsFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)
	for i := range fact.items {
		as[i] = fact.items[i].Target()
	}

	as[len(fact.items)] = fact.Sender()

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
