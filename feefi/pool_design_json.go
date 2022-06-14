package feefi

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type PoolDesignJSONPacker struct {
	jsonenc.HintedHead
	PO PoolPolicy   `json:"policy"`
	AD base.Address `json:"address"`
}

func (dg PoolDesign) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolDesignJSONPacker{
		HintedHead: jsonenc.NewHintedHead(dg.Hint()),
		PO:         dg.policy,
		AD:         dg.address,
	})
}

type PoolDesignJSONUnpacker struct {
	HT hint.Hint           `json:"_hint"`
	PO json.RawMessage     `json:"policy"`
	AD base.AddressDecoder `json:"address"`
}

func (dg *PoolDesign) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uds PoolDesignJSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.PO, uds.AD)
}
