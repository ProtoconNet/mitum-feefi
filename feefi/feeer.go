package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
)

const (
	FeeerNil   = "nil"
	FeeerFixed = "fixed"
	FeeerRatio = "ratio"
	FeeerFeefi = "feefi"
)

var (
	FeefiFeeerType   = hint.Type("mitum-feefi-feeer")
	FeefiFeeerHint   = hint.NewHint(FeefiFeeerType, "v0.0.1")
	FeefiFeeerHinter = FeefiFeeer{BaseHinter: hint.NewBaseHinter(FeefiFeeerHint)}
)

var UnlimitedMaxFeeAmount = currency.NewBig(-1)

/*
type Feeer interface {
	isvalid.IsValider
	hint.Hinter
	Type() string
	Bytes() []byte
	Receiver() base.Address
	Min() currency.Big
	ExchangeMin() currency.Big
	Fee(currency.Big) (currency.Big, error)
}
*/

// TODO:remove receiver, amount and exchangeable
type FeefiFeeer struct {
	hint.BaseHinter
	receiver          base.Address
	amount            currency.Big
	exchangeable      bool
	exchangeCID       currency.CurrencyID
	feefier           base.Address
	exchangeMinAmount currency.Big
}

func NewFeefiFeeer(
	receiver base.Address,
	amount currency.Big,
	exchangeable bool,
	currency currency.CurrencyID,
	feefier base.Address,
	exchangeAmount currency.Big) FeefiFeeer {
	return FeefiFeeer{
		BaseHinter:        hint.NewBaseHinter(FeefiFeeerHint),
		receiver:          receiver,
		amount:            amount,
		exchangeable:      exchangeable,
		exchangeCID:       currency,
		feefier:           feefier,
		exchangeMinAmount: exchangeAmount,
	}
}

func (FeefiFeeer) Type() string {
	return FeeerFeefi
}

func (fa FeefiFeeer) Bytes() []byte {
	var v int8
	if fa.exchangeable {
		v = 1
	}
	return util.ConcatBytesSlice(
		fa.receiver.Bytes(),
		fa.receiver.Bytes(),
		fa.amount.Bytes(),
		[]byte{byte(v)},
		fa.exchangeCID.Bytes(),
		fa.exchangeMinAmount.Bytes(),
		fa.feefier.Bytes(),
	)
}

func (fa FeefiFeeer) Receiver() base.Address {
	return fa.receiver
}

func (fa FeefiFeeer) Min() currency.Big {
	return fa.amount
}

func (fa FeefiFeeer) Fee(currency.Big) (currency.Big, error) {
	if fa.isZero() {
		return currency.ZeroBig, nil
	}

	return fa.amount, nil
}

func (fa FeefiFeeer) ExchangeCID() currency.CurrencyID {
	return fa.exchangeCID
}

func (fa *FeefiFeeer) SetExchangeCID(cid currency.CurrencyID) error {
	err := isvalid.Check(nil, false, cid)
	if err != nil {
		return err
	}
	fa.exchangeCID = cid
	return nil
}

func (fa FeefiFeeer) Feefier() base.Address {
	return fa.feefier
}

func (fa *FeefiFeeer) SetFeefier(acc base.Address) error {
	fa.feefier = acc
	return nil
}

func (fa FeefiFeeer) ExchangeMin() currency.Big {
	return fa.exchangeMinAmount
}

func (fa *FeefiFeeer) SetExchangeMinAmount(am currency.Big) error {
	err := isvalid.Check(nil, false, am)
	if err != nil {
		return err
	}
	fa.exchangeMinAmount = am
	return nil
}

func (fa FeefiFeeer) IsValid([]byte) error {
	if err := fa.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := isvalid.Check(nil, false, fa.receiver); err != nil {
		return isvalid.InvalidError.Errorf("invalid receiver for feefi feeer: %w", err)
	}

	if !fa.amount.OverNil() {
		return isvalid.InvalidError.Errorf("feefi feeer amount under zero")
	}

	if fa.exchangeable {
		if !fa.exchangeMinAmount.OverNil() {
			return isvalid.InvalidError.Errorf("exchange minimum amount under zero")
		}
		if len(fa.exchangeCID) < 1 {
			return isvalid.InvalidError.Errorf("feefi feeer exchage currency id empty")
		}
		if fa.feefier.Equal(nil) {
			return isvalid.InvalidError.Errorf("feefi feeer feefier address nil")
		}

		if err := isvalid.Check(nil, false, fa.feefier); err != nil {
			return isvalid.InvalidError.Errorf("invalid feefier for feefi feeer: %w", err)
		}

	}

	return nil
}

func (fa FeefiFeeer) isZero() bool {
	return fa.amount.IsZero()
}
