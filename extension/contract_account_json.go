package extension // nolint: dupl, revive

import (
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type ContractAccountStatusJSONPacker struct {
	jsonenc.HintedHead
	IA bool         `json:"isactive"`
	OW base.Address `json:"owner"`
}

func (cs ContractAccountStatus) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(ContractAccountStatusJSONPacker{
		HintedHead: jsonenc.NewHintedHead(cs.Hint()),
		IA:         cs.isActive,
		OW:         cs.owner,
	})
}

type ContractAccountStatusJSONUnpacker struct {
	IA bool                `json:"isactive"`
	OW base.AddressDecoder `json:"owner"`
}

func (cs *ContractAccountStatus) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var cuj ContractAccountStatusJSONUnpacker
	if err := enc.Unmarshal(b, &cuj); err != nil {
		return err
	}

	return cs.unpack(enc, cuj.IA, cuj.OW)
}
