package feefi

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type DesignJSONPacker struct {
	jsonenc.HintedHead
	FE currency.Amount `json:"fee"`
	AD base.Address    `json:"address"`
}

func (dg Design) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(DesignJSONPacker{
		HintedHead: jsonenc.NewHintedHead(dg.Hint()),
		FE:         dg.fee,
		AD:         dg.address,
	})
}

type DesignJSONUnpacker struct {
	HT hint.Hint           `json:"_hint"`
	FE json.RawMessage     `json:"fee"`
	AD base.AddressDecoder `json:"address"`
}

func (dg *Design) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uds DesignJSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.FE, uds.AD)
}
