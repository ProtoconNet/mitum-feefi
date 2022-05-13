package feefi

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type BaseWithdrawsItemJSONPacker struct {
	jsonenc.HintedHead
	TG base.Address        `json:"target"`
	PI currency.CurrencyID `json:"poolcid"`
	AM []currency.Amount   `json:"amounts"`
}

func (it BaseWithdrawsItem) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(BaseWithdrawsItemJSONPacker{
		HintedHead: jsonenc.NewHintedHead(it.Hint()),
		TG:         it.target,
		PI:         it.poolCID,
		AM:         it.amounts,
	})
}

type BaseWithdrawsItemJSONUnpacker struct {
	TG base.AddressDecoder `json:"target"`
	PI string              `json:"poolcid"`
	AM json.RawMessage     `json:"amounts"`
}

func (it *BaseWithdrawsItem) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uit BaseWithdrawsItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return err
	}

	return it.unpack(enc, uit.TG, uit.PI, uit.AM)
}
