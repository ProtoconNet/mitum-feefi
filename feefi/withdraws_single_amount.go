package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

var (
	WithdrawsItemSingleAmountType   = hint.Type("mitum-feefi-withdraws-item-single-amount")
	WithdrawsItemSingleAmountHint   = hint.NewHint(WithdrawsItemSingleAmountType, "v0.0.1")
	WithdrawsItemSingleAmountHinter = WithdrawsItemSingleAmount{
		BaseWithdrawsItem: BaseWithdrawsItem{BaseHinter: hint.NewBaseHinter(WithdrawsItemSingleAmountHint)},
	}
)

type WithdrawsItemSingleAmount struct {
	BaseWithdrawsItem
}

func NewWithdrawsItemSingleAmount(receiver base.Address, id extensioncurrency.ContractID, amount currency.Amount) WithdrawsItemSingleAmount {
	return WithdrawsItemSingleAmount{
		BaseWithdrawsItem: NewBaseWithdrawsItem(WithdrawsItemSingleAmountHint, receiver, id, []currency.Amount{amount}),
	}
}

func (it WithdrawsItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseWithdrawsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return isvalid.InvalidError.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it WithdrawsItemSingleAmount) Rebuild() WithdrawsItem {
	it.BaseWithdrawsItem = it.BaseWithdrawsItem.Rebuild().(BaseWithdrawsItem)

	return it
}
