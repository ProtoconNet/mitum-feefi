package feefi

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fa FeefiFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(fa.Hint()),
		bson.M{
			"receiver":          fa.receiver,
			"amount":            fa.amount,
			"exchangeable":      fa.exchangeable,
			"exchangecid":       fa.exchangeCID,
			"feefier":           fa.feefier,
			"exchangeminamount": fa.exchangeMinAmount,
		}),
	)
}

type FeefiFeeerBSONUnpacker struct {
	HT hint.Hint           `bson:"_hint"`
	RC base.AddressDecoder `bson:"receiver"`
	AM currency.Big        `bson:"amount"`
	EX bool                `bson:"exchangeable"`
	EC string              `bson:"exchangecid"`
	FR base.AddressDecoder `bson:"feefier"`
	EM currency.Big        `bson:"exchangeminamount"`
}

func (fa *FeefiFeeer) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ufa FeefiFeeerBSONUnpacker
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return err
	}

	return fa.unpack(enc, ufa.HT, ufa.RC, ufa.AM, ufa.EX, ufa.EC, ufa.FR, ufa.EM)
}
