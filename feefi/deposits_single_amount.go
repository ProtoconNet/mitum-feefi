package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

var (
	DepositsItemSingleAmountType   = hint.Type("mitum-feefi-deposits-item-single-amount")
	DepositsItemSingleAmountHint   = hint.NewHint(DepositsItemSingleAmountType, "v0.0.1")
	DepositsItemSingleAmountHinter = DepositsItemSingleAmount{
		BaseDepositsItem: BaseDepositsItem{BaseHinter: hint.NewBaseHinter(DepositsItemSingleAmountHint)},
	}
)

type DepositsItemSingleAmount struct {
	BaseDepositsItem
}

func NewDepositsItemSingleAmount(receiver base.Address, cid currency.CurrencyID, amount currency.Amount) DepositsItemSingleAmount {
	return DepositsItemSingleAmount{
		BaseDepositsItem: NewBaseDepositsItem(DepositsItemSingleAmountHint, receiver, cid, []currency.Amount{amount}),
	}
}

func (it DepositsItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseDepositsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return isvalid.InvalidError.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it DepositsItemSingleAmount) Rebuild() DepositsItem {
	it.BaseDepositsItem = it.BaseDepositsItem.Rebuild().(BaseDepositsItem)

	return it
}
