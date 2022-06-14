package digest // nolint: dupl, revive

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type FeefiPoolValueJSONPacker struct {
	jsonenc.HintedHead
	IB currency.Amount   `json:"previous_income_balance"`
	OB currency.Amount   `json:"previous_outlay_balance"`
	UC int               `json:"user_count"`
	BL []currency.Amount `json:"balance,omitempty"`
	DE feefi.PoolDesign  `json:"design,omitempty"`
	HT base.Height       `json:"height"`
	PT base.Height       `json:"previous_height"`
}

func (va FeefiPoolValue) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(FeefiPoolValueJSONPacker{
		HintedHead: jsonenc.NewHintedHead(va.Hint()),
		IB:         va.prevIncomeBalance,
		OB:         va.prevOutlayBalance,
		UC:         va.usersCount,
		BL:         va.balance,
		DE:         va.design,
		HT:         va.height,
		PT:         va.previousHeight,
	})
}

type FeefiPoolValueJSONUnpacker struct {
	IB json.RawMessage `json:"previous_income_balance"`
	OB json.RawMessage `json:"previous_outlay_balance"`
	UC int             `json:"user_count"`
	BL json.RawMessage `json:"balance"`
	DE json.RawMessage `json:"design"`
	HT base.Height     `json:"height"`
	PT base.Height     `json:"previous_height"`
}

func (va *FeefiPoolValue) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uva FeefiPoolValueJSONUnpacker
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	return va.unpack(enc, uva.IB, uva.OB, uva.UC, uva.BL, uva.DE, uva.HT, uva.PT)
}
