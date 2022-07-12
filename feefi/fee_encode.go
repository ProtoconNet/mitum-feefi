package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *FeeOperationFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bam []byte,
) error {
	fact.h = h
	fact.token = token

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

	fact.amounts = amounts

	return nil
}

func (op *FeeOperation) unpack(enc encoder.Encoder, h valuehash.Hash, bfact []byte) error {
	if err := encoder.Decode(bfact, enc, &op.fact); err != nil {
		return err
	}

	op.h = h

	return nil
}
