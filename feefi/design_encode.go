package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg *Design) unpack(enc encoder.Encoder, ht hint.Hint, bfe []byte, ad base.AddressDecoder) error {
	dg.BaseHinter = hint.NewBaseHinter(ht)

	h, err := enc.Decode(bfe)
	if err != nil {
		return err
	}
	k, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", k)
	}
	dg.fee = k

	a, err := ad.Encode(enc)
	if err != nil {
		return err
	}
	dg.address = a

	return nil
}
