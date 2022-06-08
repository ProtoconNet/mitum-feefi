package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

type BaseDepositsItem struct {
	hint.BaseHinter
	pool    base.Address
	poolID  extensioncurrency.ContractID
	amounts []currency.Amount
}

func NewBaseDepositsItem(ht hint.Hint, receiver base.Address, id extensioncurrency.ContractID, amounts []currency.Amount) BaseDepositsItem {
	return BaseDepositsItem{
		BaseHinter: hint.NewBaseHinter(ht),
		pool:       receiver,
		poolID:     id,
		amounts:    amounts,
	}
}

func (it BaseDepositsItem) Bytes() []byte {
	length := 2
	bs := make([][]byte, len(it.amounts)+length)
	bs[0] = it.pool.Bytes()
	bs[1] = it.poolID.Bytes()

	for i := range it.amounts {
		bs[i+length] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseDepositsItem) IsValid([]byte) error {
	if err := isvalid.Check(nil, false, it.pool, it.poolID); err != nil {
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

func (it BaseDepositsItem) PoolID() extensioncurrency.ContractID {
	return it.poolID
}

func (it BaseDepositsItem) Pool() base.Address {
	return it.pool
}

func (it BaseDepositsItem) Amounts() []currency.Amount {
	return it.amounts
}

func (it BaseDepositsItem) Rebuild() DepositsItem {
	ams := make([]currency.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
