package extension

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

var (
	WithdrawsItemMultiAmountsType   = hint.Type("mitum-currency-withdraws-item-multi-amounts")
	WithdrawsItemMultiAmountsHint   = hint.NewHint(WithdrawsItemMultiAmountsType, "v0.0.1")
	WithdrawsItemMultiAmountsHinter = WithdrawsItemMultiAmounts{
		BaseWithdrawsItem: BaseWithdrawsItem{BaseHinter: hint.NewBaseHinter(WithdrawsItemMultiAmountsHint)},
	}
)

var maxCurenciesWithdrawsItemMultiAmounts = 10

type WithdrawsItemMultiAmounts struct {
	BaseWithdrawsItem
}

func NewWithdrawsItemMultiAmounts(receiver base.Address, amounts []currency.Amount) WithdrawsItemMultiAmounts {
	return WithdrawsItemMultiAmounts{
		BaseWithdrawsItem: NewBaseWithdrawsItem(WithdrawsItemMultiAmountsHint, receiver, amounts),
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
