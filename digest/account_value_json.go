package digest // nolint: dupl, revive

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type AccountValueJSONPacker struct {
	jsonenc.HintedHead
	currency.AccountPackerJSON
	BL []currency.Amount `json:"balance,omitempty"`
	OW base.Address      `json:"owner,omitempty"`
	IA bool              `json:"isactivecontractaccount"`
	HT base.Height       `json:"height"`
	PT base.Height       `json:"previous_height"`
}

func (va AccountValue) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(AccountValueJSONPacker{
		HintedHead:        jsonenc.NewHintedHead(va.Hint()),
		AccountPackerJSON: va.ac.PackerJSON(),
		BL:                va.balance,
		OW:                va.owner,
		IA:                va.isActiveContractAccount,
		HT:                va.height,
		PT:                va.previousHeight,
	})
}

type AccountValueJSONUnpacker struct {
	BL json.RawMessage     `json:"balance"`
	OW base.AddressDecoder `json:"owner"`
	IA bool                `json:"isactivecontractaccount"`
	HT base.Height         `json:"height"`
	PT base.Height         `json:"previous_height"`
}

func (va *AccountValue) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uva AccountValueJSONUnpacker
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	ac := new(currency.Account)
	if err := va.unpack(enc, nil, uva.BL, uva.OW, uva.IA, uva.HT, uva.PT); err != nil {
		return err
	} else if err := ac.UnpackJSON(b, enc); err != nil {
		return err
	} else {
		va.ac = *ac

		return nil
	}
}
