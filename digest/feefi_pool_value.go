package digest

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	FeefiPoolValueType = hint.Type("mitum-feefi-pool-value")
	FeefiPoolValueHint = hint.NewHint(FeefiPoolValueType, "v0.0.1")
)

type FeefiPoolValue struct {
	prevIncome     currency.Amount
	prevOutlay     currency.Amount
	userCount      int
	address        base.Address
	balance        []currency.Amount
	design         feefi.Design
	height         base.Height
	previousHeight base.Height
}

func NewFeefiPoolValue(
	pool feefi.Pool,
	height base.Height,
) FeefiPoolValue {
	return FeefiPoolValue{
		prevIncome: pool.IncomeBalance(),
		prevOutlay: pool.OutlayBalance(),
		userCount:  len(pool.Users()),
		height:     height,
	}
}

func (FeefiPoolValue) Hint() hint.Hint {
	return FeefiPoolValueHint
}

func (fp FeefiPoolValue) IncomeBalance() feefi.Pool {
	return fp.
}

func (fp FeefiPoolValue) Height() base.Height {
	return fp.height
}
