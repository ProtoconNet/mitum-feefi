package digest

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/state"
	mongodbstorage "github.com/spikeekips/mitum/storage/mongodb"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

type FeefiPoolDoc struct {
	mongodbstorage.BaseDoc
	st state.State
}

func NewFeefiPoolDoc(st state.State, enc encoder.Encoder) (FeefiPoolDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return FeefiPoolDoc{}, err
	}

	return FeefiPoolDoc{
		BaseDoc: b,
		st:      st,
	}, nil
}

func (doc FeefiPoolDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyPoolSuffix)]
	m["address"] = address
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type FeefiDesignDoc struct {
	mongodbstorage.BaseDoc
	st state.State
}

func NewFeefiDesignDoc(st state.State, enc encoder.Encoder) (FeefiDesignDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return FeefiDesignDoc{}, err
	}

	return FeefiDesignDoc{
		BaseDoc: b,
		st:      st,
	}, nil
}

func (doc FeefiDesignDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyDesignSuffix)]
	m["address"] = address
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type FeefiBalanceDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	am currency.Amount
}

// NewBalanceDoc gets the State of Amount
func NewFeefiBalanceDoc(st state.State, enc encoder.Encoder) (FeefiBalanceDoc, error) {
	am, err := currency.StateBalanceValue(st)
	if err != nil {
		return FeefiBalanceDoc{}, errors.Wrap(err, "feefibalanceDoc needs Amount state")
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return FeefiBalanceDoc{}, err
	}

	return FeefiBalanceDoc{
		BaseDoc: b,
		st:      st,
		am:      am,
	}, nil
}

func (doc FeefiBalanceDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyBalanceSuffix)-len(doc.am.Currency())-len(doc.am.Currency())-2]
	m["address"] = address
	m["currency"] = doc.am.Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
