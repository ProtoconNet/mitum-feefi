package feefi

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg PoolPolicy) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(dg.Hint()),
		bson.M{
			"fee": dg.fee,
		}),
	)
}

type PoolPolicyBSONUnpacker struct {
	HT hint.Hint `bson:"_hint"`
	FE bson.Raw  `bson:"fee"`
}

func (dg *PoolPolicy) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uds PoolPolicyBSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.FE)
}
