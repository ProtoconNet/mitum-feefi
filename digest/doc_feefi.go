package digest

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	mongodbstorage "github.com/spikeekips/mitum/storage/mongodb"
	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

type FeefiPoolValueDoc struct {
	mongodbstorage.BaseDoc
	va          FeefiPoolValue
	feefipoolid string
	height      base.Height
}

func NewFeefiPoolValueDoc(st state.State, enc encoder.Encoder) (FeefiPoolValueDoc, FeefiPoolDoc, error) {
	if rs, err := NewFeefiPoolValue(st); err != nil {
		return FeefiPoolValueDoc{}, FeefiPoolDoc{}, err
	} else if b, err := mongodbstorage.NewBaseDoc(nil, rs, enc); err != nil {
		return FeefiPoolValueDoc{}, FeefiPoolDoc{}, err
	} else if pl, err := feefi.StatePoolValue(st); err != nil {
		return FeefiPoolValueDoc{}, FeefiPoolDoc{}, errors.Wrap(err, "FeefiPoolDoc needs feefi pool state")
	} else if b2, err := mongodbstorage.NewBaseDoc(nil, st, enc); err != nil {
		return FeefiPoolValueDoc{}, FeefiPoolDoc{}, err
	} else {
		return FeefiPoolValueDoc{
				BaseDoc:     b,
				va:          rs,
				feefipoolid: rs.PrevIncomeBalance().Currency().String(),
				height:      rs.height,
			},
			FeefiPoolDoc{
				BaseDoc: b2,
				st:      st,
				pl:      pl,
			},
			nil
	}
}

func (doc FeefiPoolValueDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["feefipoolid"] = doc.feefipoolid
	m["height"] = doc.height

	return bsonenc.Marshal(m)
}

type FeefiPoolDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	pl feefi.Pool
}

func (doc FeefiPoolDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyPoolSuffix)-len(doc.pl.IncomeBalance().Currency())-1]
	m["feefipoolid"] = doc.pl.IncomeBalance().Currency().String()
	m["address"] = address
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type FeefiDesignDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	de feefi.PoolDesign
}

func NewFeefiDesignDoc(st state.State, enc encoder.Encoder) (FeefiDesignDoc, error) {
	de, err := feefi.StateDesignValue(st)
	if err != nil {
		return FeefiDesignDoc{}, errors.Wrap(err, "feefiDesignDoc needs feefi design state")
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return FeefiDesignDoc{}, err
	}

	return FeefiDesignDoc{
		BaseDoc: b,
		st:      st,
		de:      de,
	}, nil
}

func (doc FeefiDesignDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyDesignSuffix)-len(doc.de.Policy().Fee().Currency())-1]
	m["feefipoolid"] = doc.de.Policy().Fee().Currency().String()
	m["address"] = address
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}

type FeefiBalanceDoc struct {
	mongodbstorage.BaseDoc
	st state.State
	am extensioncurrency.AmountValue
}

// NewFeefiBalanceDoc gets the State of feefi pool balance amount
func NewFeefiBalanceDoc(st state.State, enc encoder.Encoder) (FeefiBalanceDoc, error) {
	am, err := extensioncurrency.StateBalanceValue(st)
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

	address := doc.st.Key()[:len(doc.st.Key())-len(feefi.StateKeyBalanceSuffix)-len(doc.am.Amount().Currency())-len(doc.am.ID())-2]

	m["address"] = address
	m["feefipoolid"] = doc.am.ID().String()
	m["currency"] = doc.am.Amount().Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
