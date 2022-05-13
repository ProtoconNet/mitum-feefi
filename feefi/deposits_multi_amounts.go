package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

var maxDepositsItemMultiAmounts = 10

var (
	DepositsItemMultiAmountsType   = hint.Type("mitum-feefi-deposits-item-multi-amounts")
	DepositsItemMultiAmountsHint   = hint.NewHint(DepositsItemMultiAmountsType, "v0.0.1")
	DepositsItemMultiAmountsHinter = DepositsItemMultiAmounts{
		BaseDepositsItem: BaseDepositsItem{BaseHinter: hint.NewBaseHinter(DepositsItemMultiAmountsHint)},
	}
)

type DepositsItemMultiAmounts struct {
	BaseDepositsItem
}

func NewDepositsItemMultiAmounts(receiver base.Address, cid currency.CurrencyID, amounts []currency.Amount) DepositsItemMultiAmounts {
	return DepositsItemMultiAmounts{
		BaseDepositsItem: NewBaseDepositsItem(DepositsItemMultiAmountsHint, receiver, cid, amounts),
	}
}

func (it DepositsItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseDepositsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxDepositsItemMultiAmounts {
		return isvalid.InvalidError.Errorf("amounts over allowed; %d > %d", n, maxDepositsItemMultiAmounts)
	}

	return nil
}

func (it DepositsItemMultiAmounts) Rebuild() DepositsItem {
	it.BaseDepositsItem = it.BaseDepositsItem.Rebuild().(BaseDepositsItem)

	return it
}
