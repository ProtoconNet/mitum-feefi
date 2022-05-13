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
	// NilFeeerType     = hint.Type("mitum-feefi-nil-feeer")
	// NilFeeerHint     = hint.NewHint(NilFeeerType, "v0.0.1")
	// NilFeeerHinter   = NilFeeer{BaseHinter: hint.NewBaseHinter(NilFeeerHint)}
	// FixedFeeerType   = hint.Type("mitum-feefi-fixed-feeer")
	// FixedFeeerHint   = hint.NewHint(FixedFeeerType, "v0.0.1")
	// FixedFeeerHinter = FixedFeeer{BaseHinter: hint.NewBaseHinter(FixedFeeerHint)}
	// RatioFeeerType   = hint.Type("mitum-feefi-ratio-feeer")
	// RatioFeeerHint   = hint.NewHint(RatioFeeerType, "v0.0.1")
	// RatioFeeerHinter = RatioFeeer{BaseHinter: hint.NewBaseHinter(RatioFeeerHint)}
	FeefiFeeerType   = hint.Type("mitum-feefi-feeer")
	FeefiFeeerHint   = hint.NewHint(FeefiFeeerType, "v0.0.1")
	FeefiFeeerHinter = FeefiFeeer{BaseHinter: hint.NewBaseHinter(FeefiFeeerHint)}
)

var UnlimitedMaxFeeAmount = currency.NewBig(-1)

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

/*
type NilFeeer struct {
	hint.BaseHinter
}

func NewNilFeeer() NilFeeer {
	return NilFeeer{BaseHinter: hint.NewBaseHinter(NilFeeerHint)}
}

func (NilFeeer) Type() string {
	return FeeerNil
}

func (NilFeeer) Bytes() []byte {
	return nil
}

func (NilFeeer) Receiver() base.Address {
	return nil
}

func (NilFeeer) Min() currency.Big {
	return currency.ZeroBig
}

func (NilFeeer) ExchangeMin() currency.Big {
	return currency.ZeroBig
}

func (NilFeeer) Fee(currency.Big) (currency.Big, error) {
	return currency.ZeroBig, nil
}

func (fa NilFeeer) IsValid([]byte) error {
	return fa.BaseHinter.IsValid(nil)
}

type FixedFeeer struct {
	hint.BaseHinter
	receiver          base.Address
	amount            currency.Big
	exchangeMinAmount currency.Big
}

func NewFixedFeeer(receiver base.Address, amount currency.Big, exchagneMinAmount currency.Big) FixedFeeer {
	return FixedFeeer{
		BaseHinter:        hint.NewBaseHinter(currency.FixedFeeerHint),
		receiver:          receiver,
		amount:            amount,
		exchangeMinAmount: exchagneMinAmount,
	}
}

func (FixedFeeer) Type() string {
	return FeeerFixed
}

func (fa FixedFeeer) Bytes() []byte {
	return util.ConcatBytesSlice(fa.receiver.Bytes(), fa.amount.Bytes())
}

func (fa FixedFeeer) Receiver() base.Address {
	return fa.receiver
}

func (fa FixedFeeer) Min() currency.Big {
	return fa.amount
}

func (fa FixedFeeer) Fee(currency.Big) (currency.Big, error) {
	if fa.isZero() {
		return currency.ZeroBig, nil
	}

	return fa.amount, nil
}

func (fa FixedFeeer) ExchangeMin() currency.Big {
	return fa.exchangeMinAmount
}

func (fa FixedFeeer) IsValid([]byte) error {
	if err := fa.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := isvalid.Check(nil, false, fa.receiver); err != nil {
		return isvalid.InvalidError.Errorf("invalid receiver for fixed feeer: %w", err)
	}

	if !fa.amount.OverNil() {
		return isvalid.InvalidError.Errorf("fixed feeer amount under zero")
	}

	if !fa.exchangeMinAmount.OverNil() {
		return isvalid.InvalidError.Errorf("fixed feeer exchange minimum amount under zero")
	}

	return nil
}

func (fa FixedFeeer) isZero() bool {
	return fa.amount.IsZero()
}

type RatioFeeer struct {
	hint.BaseHinter
	receiver    base.Address
	ratio       float64 // 0 >=, or <= 1.0
	min         currency.Big
	max         currency.Big
	exchangeMin currency.Big
}

func NewRatioFeeer(receiver base.Address, ratio float64, min, max, exchangeMin currency.Big) RatioFeeer {
	return RatioFeeer{
		BaseHinter:  hint.NewBaseHinter(RatioFeeerHint),
		receiver:    receiver,
		ratio:       ratio,
		min:         min,
		max:         max,
		exchangeMin: exchangeMin,
	}
}

func (RatioFeeer) Type() string {
	return FeeerRatio
}

func (fa RatioFeeer) Bytes() []byte {
	var rb bytes.Buffer
	_ = binary.Write(&rb, binary.BigEndian, fa.ratio)

	return util.ConcatBytesSlice(fa.receiver.Bytes(), rb.Bytes(), fa.min.Bytes(), fa.max.Bytes())
}

func (fa RatioFeeer) Receiver() base.Address {
	return fa.receiver
}

func (fa RatioFeeer) Min() currency.Big {
	return fa.min
}

func (fa RatioFeeer) ExchangeMin() currency.Big {
	return fa.exchangeMin
}

func (fa RatioFeeer) Fee(a currency.Big) (currency.Big, error) {
	if fa.isZero() {
		return currency.ZeroBig, nil
	} else if a.IsZero() {
		return fa.min, nil
	}

	if fa.isOne() {
		return a, nil
	} else if f := a.MulFloat64(fa.ratio); f.Compare(fa.min) < 0 {
		return fa.min, nil
	} else {
		if !fa.isUnlimited() && f.Compare(fa.max) > 0 {
			return fa.max, nil
		}
		return f, nil
	}
}

func (fa RatioFeeer) IsValid([]byte) error {
	if err := fa.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := isvalid.Check(nil, false, fa.receiver); err != nil {
		return isvalid.InvalidError.Errorf("invalid receiver for ratio feeer: %w", err)
	}

	if fa.ratio < 0 || fa.ratio > 1 {
		return isvalid.InvalidError.Errorf("invalid ratio, %v; it should be 0 >=, <= 1", fa.ratio)
	}

	if !fa.min.OverNil() {
		return isvalid.InvalidError.Errorf("ratio feeer min amount under zero")
	} else if !fa.max.Equal(UnlimitedMaxFeeAmount) {
		if !fa.max.OverNil() {
			return isvalid.InvalidError.Errorf("ratio feeer max amount under zero")
		} else if fa.min.Compare(fa.max) > 0 {
			return isvalid.InvalidError.Errorf("ratio feeer min should over max")
		}
	}

	return nil
}

func (fa RatioFeeer) isUnlimited() bool {
	return fa.max.Equal(UnlimitedMaxFeeAmount)
}

func (fa RatioFeeer) isZero() bool {
	return fa.ratio == 0
}

func (fa RatioFeeer) isOne() bool {
	return fa.ratio == 1
}

func NewFeeToken(feeer Feeer, height base.Height) []byte {
	return util.ConcatBytesSlice(feeer.Bytes(), height.Bytes())
}

*/

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
		return isvalid.InvalidError.Errorf("invalid receiver for fixed feeer: %w", err)
	}

	if !fa.amount.OverNil() {
		return isvalid.InvalidError.Errorf("fixed feeer amount under zero")
	}

	if fa.exchangeable {
		if !fa.exchangeMinAmount.OverNil() {
			return isvalid.InvalidError.Errorf("exchangeable fixed feeer min amount under zero")
		}
		if len(fa.exchangeCID) < 1 {
			return isvalid.InvalidError.Errorf("exchangeable fixed feeer exchagecid empty")
		}
		if fa.feefier.Equal(nil) {
			return isvalid.InvalidError.Errorf("exchangeable fixed feeer feefier nil")
		}

	}

	return nil
}

func (fa FeefiFeeer) isZero() bool {
	return fa.amount.IsZero()
}
