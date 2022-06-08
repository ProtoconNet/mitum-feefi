package feefi

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/valuehash"
)

type DepositsFactJSONPacker struct {
	jsonenc.HintedHead
	H  valuehash.Hash               `json:"hash"`
	TK []byte                       `json:"token"`
	SD base.Address                 `json:"sender"`
	PL base.Address                 `json:"pool"`
	CI extensioncurrency.ContractID `json:"poolid"`
	AM currency.Amount              `json:"amount"`
}

func (fact DepositFact) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(DepositsFactJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fact.Hint()),
		H:          fact.h,
		TK:         fact.token,
		SD:         fact.sender,
		PL:         fact.pool,
		CI:         fact.poolID,
		AM:         fact.amount,
	})
}

func (fact *DepositFact) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufact struct {
		H  valuehash.Bytes     `json:"hash"`
		TK []byte              `json:"token"`
		SD base.AddressDecoder `json:"sender"`
		PL base.AddressDecoder `json:"pool"`
		CI string              `json:"poolid"`
		AM json.RawMessage     `json:"amount"`
	}
	if err := jsonenc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.PL, ufact.CI, ufact.AM)
}

func (op *Deposit) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
