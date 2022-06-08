package digest // nolint: dupl, revive

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (va *FeefiPoolValue) unpack(
	enc encoder.Encoder, bib, bob []byte, uc int, bl, de []byte, height, previousHeight base.Height,
) error {
	h, err := enc.Decode(bib)
	if err != nil {
		return err
	}
	v, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected currency.Amount, not %T", h)
	}
	va.prevIncomeBalance = v

	h, err = enc.Decode(bob)
	if err != nil {
		return err
	}
	v, ok = h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected currency.Amount, not %T", h)
	}
	va.prevOutlayBalance = v

	va.usersCount = uc

	ham, err := enc.DecodeSlice(bl)
	if err != nil {
		return err
	}
	balance := make([]currency.Amount, len(ham))
	for i := range ham {
		v, ok := ham[i].(currency.Amount)
		if !ok {
			return util.WrongTypeError.Errorf("expected currency.Amount, not %T", ham[i])
		}
		balance[i] = v
	}

	va.balance = balance

	h, err = enc.Decode(de)
	if err != nil {
		return err
	}
	k, ok := h.(feefi.Design)
	if !ok {
		return util.WrongTypeError.Errorf("expected feefi.Design, not %T", h)
	}

	va.design = k

	va.height = height
	va.previousHeight = previousHeight

	return nil
}
