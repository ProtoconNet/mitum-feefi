package extension // nolint: dupl, revive

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
)

func (cs *ContractAccountStatus) unpack(
	enc encoder.Encoder,
	ia bool,
	ow base.AddressDecoder,
) error {
	a, err := ow.Encode(enc)
	if err != nil {
		return err
	}
	cs.owner = a
	cs.isActive = ia

	return nil
}
