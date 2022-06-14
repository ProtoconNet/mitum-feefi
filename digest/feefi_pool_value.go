package digest

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	FeefiPoolValueType = hint.Type("mitum-feefi-pool-value")
	FeefiPoolValueHint = hint.NewHint(FeefiPoolValueType, "v0.0.1")
)

type FeefiPoolValue struct {
	prevIncomeBalance currency.Amount
	prevOutlayBalance currency.Amount
	usersCount        int
	balance           []currency.Amount
	design            feefi.PoolDesign
	height            base.Height
	previousHeight    base.Height
}

func NewFeefiPoolValue(st state.State) (FeefiPoolValue, error) {
	var fp feefi.Pool
	switch a, ok, err := IsFeefiPoolState(st); {
	case err != nil:
		return FeefiPoolValue{}, err
	case !ok:
		return FeefiPoolValue{}, errors.Errorf("not state for feefi.Pool, %T", st.Value().Interface())
	default:
		fp = a
	}
	address := st.Key()[:len(st.Key())-len(feefi.StateKeyPoolSuffix)-len(fp.IncomeBalance().Currency())-1]
	return FeefiPoolValue{
		prevIncomeBalance: fp.IncomeBalance(),
		prevOutlayBalance: fp.OutlayBalance(),
		usersCount:        len(fp.Users()),
		design:            feefi.NewPoolDesign(currency.NewZeroAmount(fp.IncomeBalance().Currency()), currency.NewAddress(address)),
		height:            st.Height(),
		previousHeight:    st.PreviousHeight(),
	}, nil
}

func (FeefiPoolValue) Hint() hint.Hint {
	return FeefiPoolValueHint
}

func (fp FeefiPoolValue) PrevIncomeBalance() currency.Amount {
	return fp.prevIncomeBalance
}

func (fp FeefiPoolValue) PrevOutlayBalance() currency.Amount {
	return fp.prevOutlayBalance
}

func (fp FeefiPoolValue) Design() feefi.PoolDesign {
	return fp.design
}

func (fp FeefiPoolValue) Address() string {
	return fp.design.Address().String()
}

func (fp FeefiPoolValue) Balance() []currency.Amount {
	return fp.balance
}

func (fp FeefiPoolValue) Height() base.Height {
	return fp.height
}

func (fp FeefiPoolValue) SetHeight(height base.Height) FeefiPoolValue {
	if int64(height) > int64(fp.height) {
		fp.height = height
	}

	return fp
}

func (fp FeefiPoolValue) SetPreviousHeight(height base.Height) FeefiPoolValue {
	if int64(height) > int64(fp.previousHeight) {
		fp.previousHeight = height
	}

	return fp
}

func (fp FeefiPoolValue) SetBalance(balance []currency.Amount) FeefiPoolValue {
	fp.balance = balance

	return fp
}

func (fp FeefiPoolValue) SetFeefiDesign(design feefi.PoolDesign) FeefiPoolValue {
	fp.design = design

	return fp
}
