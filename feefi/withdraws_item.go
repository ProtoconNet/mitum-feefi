package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

type BaseWithdrawsItem struct {
	hint.BaseHinter
	target  base.Address
	poolCID currency.CurrencyID
	amounts []currency.Amount
}

func NewBaseWithdrawsItem(ht hint.Hint, target base.Address, poolCID currency.CurrencyID, amounts []currency.Amount) BaseWithdrawsItem {
	return BaseWithdrawsItem{
		BaseHinter: hint.NewBaseHinter(ht),
		target:     target,
		poolCID:    poolCID,
		amounts:    amounts,
	}
}

func (it BaseWithdrawsItem) Bytes() []byte {
	length := 2
	bs := make([][]byte, len(it.amounts)+length)
	bs[0] = it.target.Bytes()
	bs[1] = it.poolCID.Bytes()

	for i := range it.amounts {
		bs[i+length] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseWithdrawsItem) IsValid([]byte) error {
	err := isvalid.Check(nil, false, it.target, it.poolCID)
	if err != nil {
		return err
	}

	if n := len(it.amounts); n == 0 {
		return isvalid.InvalidError.Errorf("empty amounts")
	}

	founds := map[currency.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return isvalid.InvalidError.Errorf("duplicated currency found, %q", am.Currency())
		}
		founds[am.Currency()] = struct{}{}

		if err := am.IsValid(nil); err != nil {
			return err
		} else if !am.Big().OverZero() {
			return isvalid.InvalidError.Errorf("amount should be over zero")
		}
	}

	return nil
}

func (it BaseWithdrawsItem) Target() base.Address {
	return it.target
}

func (it BaseWithdrawsItem) PoolCID() currency.CurrencyID {
	return it.poolCID
}

func (it BaseWithdrawsItem) Amounts() []currency.Amount {
	return it.amounts
}

func (it BaseWithdrawsItem) Rebuild() WithdrawsItem {
	ams := make([]currency.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
