package feefi

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg *PoolDesign) unpack(enc encoder.Encoder, ht hint.Hint, bpo []byte, ad base.AddressDecoder) error {
	dg.BaseHinter = hint.NewBaseHinter(ht)

	h, err := enc.Decode(bpo)
	if err != nil {
		return err
	}
	k, ok := h.(PoolPolicy)
	if !ok {
		return util.WrongTypeError.Errorf("expected PoolPolicy, not %T", k)
	}
	dg.policy = k

	a, err := ad.Encode(enc)
	if err != nil {
		return err
	}
	dg.address = a

	return nil
}
