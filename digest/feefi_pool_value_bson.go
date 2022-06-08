package digest // nolint: dupl, revive

import (
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (va FeefiPoolValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(va.Hint()),
		bson.M{
			"previous_income_balance": va.prevIncomeBalance,
			"previous_outlay_balance": va.prevOutlayBalance,
			"user_count":              va.usersCount,
			"balance":                 va.balance,
			"design":                  va.design,
			"height":                  va.height,
			"previous_height":         va.previousHeight,
		},
	))
}

type FeefiPoolValueBSONUnpacker struct {
	IB bson.Raw    `bson:"previous_income_balance"`
	OB bson.Raw    `bson:"previous_outlay_balance"`
	UC int         `bson:"user_count"`
	BL bson.Raw    `bson:"balance"`
	DE bson.Raw    `bson:"design"`
	HT base.Height `bson:"height"`
	PT base.Height `bson:"previous_height"`
}

func (va *FeefiPoolValue) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uva FeefiPoolValueBSONUnpacker
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	return va.unpack(enc, uva.IB, uva.OB, uva.UC, uva.BL, uva.DE, uva.HT, uva.PT)
}
