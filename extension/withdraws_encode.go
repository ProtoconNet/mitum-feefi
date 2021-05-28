package extension

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (it *BaseWithdrawsItem) unpack(
	enc encoder.Encoder,
	bTarget base.AddressDecoder,
	bam []byte,
) error {
	a, err := bTarget.Encode(enc)
	if err != nil {
		return err
	}
	it.target = a

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return err
	}

	am := make([]currency.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(currency.Amount)
		if !ok {
			return util.WrongTypeError.Errorf("expected Amount, not %T", ham[i])
		}

		am[i] = j
	}

	it.amounts = am

	return nil
}

func (fact *WithdrawsFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bSender base.AddressDecoder,
	bits []byte,
) error {
	sender, err := bSender.Encode(enc)
	if err != nil {
		return err
	}

	hits, err := enc.DecodeSlice(bits)
	if err != nil {
		return err
	}

	items := make([]WithdrawsItem, len(hits))
	for i := range hits {
		j, ok := hits[i].(WithdrawsItem)
		if !ok {
			return util.WrongTypeError.Errorf("expected TransfersItem, not %T", hits[i])
		}

		items[i] = j
	}

	fact.h = h
	fact.token = token
	fact.sender = sender
	fact.items = items

	return nil
}
