package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (it *BaseWithdrawsItem) unpack(
	enc encoder.Encoder,
	bTarget base.AddressDecoder,
	spi string,
	bam []byte,
) error {
	a, err := bTarget.Encode(enc)
	if err != nil {
		return err
	}
	it.target = a

	it.poolID = extensioncurrency.ContractID(spi)

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
	bPool base.AddressDecoder,
	spi string,
	bam []byte,
) error {
	sender, err := bSender.Encode(enc)
	if err != nil {
		return err
	}

	pool, err := bPool.Encode(enc)
	if err != nil {
		return err
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return err
	}

	amounts := make([]currency.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(currency.Amount)
		if !ok {
			return util.WrongTypeError.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	fact.h = h
	fact.token = token
	fact.sender = sender
	fact.pool = pool
	fact.poolID = extensioncurrency.ContractID(spi)
	fact.amounts = amounts

	return nil
}
