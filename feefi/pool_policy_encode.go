package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg *PoolPolicy) unpack(enc encoder.Encoder, ht hint.Hint, bfe []byte) error {
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

	return nil
}
