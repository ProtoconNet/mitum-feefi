package feefi // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact DepositFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"hash":    fact.h,
				"token":   fact.token,
				"sender":  fact.sender,
				"pool":    fact.pool,
				"poolcid": fact.poolCID,
				"amount":  fact.amount,
			}))
}

type DepositsFactBSONUnpacker struct {
	H  valuehash.Bytes     `bson:"hash"`
	TK []byte              `bson:"token"`
	SD base.AddressDecoder `bson:"sender"`
	PL base.AddressDecoder `bson:"pool"`
	CI string              `bson:"poolcid"`
	AM bson.Raw            `bson:"amount"`
}

func (fact *DepositFact) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ufact DepositsFactBSONUnpacker
	if err := enc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.PL, ufact.CI, ufact.AM)
}

func (op *Deposit) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
