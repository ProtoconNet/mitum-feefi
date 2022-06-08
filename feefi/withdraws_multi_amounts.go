package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

var (
	WithdrawsItemMultiAmountsType   = hint.Type("mitum-feefi-withdraws-item-multi-amounts")
	WithdrawsItemMultiAmountsHint   = hint.NewHint(WithdrawsItemMultiAmountsType, "v0.0.1")
	WithdrawsItemMultiAmountsHinter = WithdrawsItemMultiAmounts{
		BaseWithdrawsItem: BaseWithdrawsItem{BaseHinter: hint.NewBaseHinter(WithdrawsItemMultiAmountsHint)},
	}
)

var maxCurenciesWithdrawsItemMultiAmounts = 10

type WithdrawsItemMultiAmounts struct {
	BaseWithdrawsItem
}

func NewWithdrawsItemMultiAmounts(receiver base.Address, id extensioncurrency.ContractID, amounts []currency.Amount) WithdrawsItemMultiAmounts {
	return WithdrawsItemMultiAmounts{
		BaseWithdrawsItem: NewBaseWithdrawsItem(WithdrawsItemMultiAmountsHint, receiver, id, amounts),
	}
}

func (it WithdrawsItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseWithdrawsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesWithdrawsItemMultiAmounts {
		return isvalid.InvalidError.Errorf("amounts over allowed; %d > %d", n, maxCurenciesWithdrawsItemMultiAmounts)
	}

	return nil
}

func (it WithdrawsItemMultiAmounts) Rebuild() WithdrawsItem {
	it.BaseWithdrawsItem = it.BaseWithdrawsItem.Rebuild().(BaseWithdrawsItem)

	return it
}
