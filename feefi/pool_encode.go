package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (pl *Pool) unpack(enc encoder.Encoder, ht hint.Hint, bus []byte, bib []byte, bob []byte) error {
	pl.BaseHinter = hint.NewBaseHinter(ht)
	hmap, err := enc.DecodeMap(bus)
	if err != nil {
		return err
	}
	fmap := make(map[string]PoolUserBalance)
	for k := range hmap {
		j, ok := hmap[k].(PoolUserBalance)
		if !ok {
			return util.WrongTypeError.Errorf("expected PoolUserBalance, not %T", hmap[k])
		}
		fmap[k] = j
	}

	pl.users = fmap

	h, err := enc.Decode(bib)
	if err != nil {
		return err
	}
	i, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", h)
	}
	pl.prevIncomeBalance = i

	h, err = enc.Decode(bob)
	if err != nil {
		return err
	}
	j, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", h)
	}
	pl.prevOutlayBalance = j

	return nil
}

func (pl *PoolUserBalance) unpack(enc encoder.Encoder, ht hint.Hint, bia []byte, boa []byte) error {
	pl.BaseHinter = hint.NewBaseHinter(ht)
	h, err := enc.Decode(bia)
	if err != nil {
		return err
	}
	i, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", h)
	}
	pl.incomeAmount = i

	h, err = enc.Decode(boa)
	if err != nil {
		return err
	}
	j, ok := h.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", h)
	}
	pl.outlayAmount = j

	return nil
}
