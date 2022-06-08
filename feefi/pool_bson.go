package feefi

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (pl Pool) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(pl.Hint()),
		bson.M{
			"users":                pl.users,
			"previncomefeebalance": pl.prevIncomeAmount,
			"prevoutlayfeebalance": pl.prevOutlayAmount,
		}),
	)
}

type PoolBSONUnpacker struct {
	HT hint.Hint `bson:"_hint"`
	US bson.Raw  `bson:"users"`
	IB bson.Raw  `bson:"previncomefeebalance"`
	OB bson.Raw  `bson:"prevoutlayfeebalance"`
}

func (pl *Pool) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var upl PoolBSONUnpacker
	if err := enc.Unmarshal(b, &upl); err != nil {
		return err
	}

	return pl.unpack(enc, upl.HT, upl.US, upl.IB, upl.OB)
}

func (pl PoolUserBalance) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(pl.Hint()),
		bson.M{
			"incomeamount": pl.income,
			"outlayamount": pl.outlay,
		}),
	)
}

type PoolUserBalanceBSONUnpacker struct {
	HT hint.Hint `bson:"_hint"`
	IA bson.Raw  `bson:"incomeamount"`
	OA bson.Raw  `bson:"outlayamount"`
}

func (pl *PoolUserBalance) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var upl PoolUserBalanceBSONUnpacker
	if err := enc.Unmarshal(b, &upl); err != nil {
		return err
	}

	return pl.unpack(enc, upl.HT, upl.IA, upl.OA)
}
