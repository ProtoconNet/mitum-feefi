package cmds

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util/isvalid"
)

type GenesisCurrenciesDesign struct {
	AccountKeys *currencycmds.AccountKeysDesign `yaml:"account-keys"`
	Currencies  []*CurrencyDesign               `yaml:"currencies"`
}

func (de *GenesisCurrenciesDesign) IsValid([]byte) error {
	if de.AccountKeys == nil {
		return errors.Errorf("empty account-keys")
	}

	if err := de.AccountKeys.IsValid(nil); err != nil {
		return err
	}

	for i := range de.Currencies {
		if err := de.Currencies[i].IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

type CurrencyDesign struct {
	CurrencyString             *string         `yaml:"currency"`
	BalanceString              *string         `yaml:"balance"`
	NewAccountMinBalanceString *string         `yaml:"new-account-min-balance"`
	Feeer                      *FeeerDesign    `yaml:"feeer"`
	Balance                    currency.Amount `yaml:"-"`
	NewAccountMinBalance       currency.Big    `yaml:"-"`
}

func (de *CurrencyDesign) IsValid([]byte) error {
	var cid currency.CurrencyID
	if de.CurrencyString == nil {
		return errors.Errorf("empty currency")
	}
	cid = currency.CurrencyID(*de.CurrencyString)
	if err := cid.IsValid(nil); err != nil {
		return err
	}

	if de.BalanceString != nil {
		b, err := currency.NewBigFromString(*de.BalanceString)
		if err != nil {
			return isvalid.InvalidError.Wrap(err)
		}
		de.Balance = currency.NewAmount(b, cid)
		if err := de.Balance.IsValid(nil); err != nil {
			return err
		}
	}

	if de.NewAccountMinBalanceString == nil {
		de.NewAccountMinBalance = currency.ZeroBig
	} else {
		b, err := currency.NewBigFromString(*de.NewAccountMinBalanceString)
		if err != nil {
			return isvalid.InvalidError.Wrap(err)
		}
		de.NewAccountMinBalance = b
	}

	if de.Feeer == nil {
		de.Feeer = &FeeerDesign{}
	} else if err := de.Feeer.IsValid(nil); err != nil {
		return err
	}

	return nil
}

// FeeerDesign is used for genesis currencies and naturally it's receiver is genesis account
type FeeerDesign struct {
	Type   string
	Extras map[string]interface{} `yaml:",inline"`
}

func (no *FeeerDesign) IsValid([]byte) error {
	switch t := no.Type; t {
	case feefi.FeeerNil, "":
	case feefi.FeeerFixed:
		if err := no.checkFixed(no.Extras); err != nil {
			return err
		}
	case feefi.FeeerFeefi:
		if err := no.checkFeefi(no.Extras); err != nil {
			return err
		}
	case feefi.FeeerRatio:
		if err := no.checkRatio(no.Extras); err != nil {
			return err
		}
	default:
		return errors.Errorf("unknown type of feeer, %v", t)
	}

	return nil
}

func (no FeeerDesign) checkFixed(c map[string]interface{}) error {
	a, found := c["amount"]
	if !found {
		return errors.Errorf("fixed needs `amount`")
	}
	n, err := currency.NewBigFromInterface(a)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of fixed", a)
	}
	no.Extras["fixed_amount"] = n
	exchangeMinAmount, found := c["exchange-min-amount"]
	if !found {
		return errors.Errorf("fixed needs `exchange-min-amount`")
	}

	e, err := currency.NewBigFromInterface(exchangeMinAmount)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of exchange-min-amount", exchangeMinAmount)
	}
	no.Extras["fixed_exchange_min_amount"] = e
	return nil
}

func (no FeeerDesign) checkFeefi(c map[string]interface{}) error {
	amount, found := c["amount"]
	if !found {
		return errors.Errorf("feefi needs `amount`")
	}
	feefiAmount, err := currency.NewBigFromInterface(amount)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of feefi", amount)
	}
	no.Extras["feefi_amount"] = feefiAmount

	exchangeable, found := c["exchangeable"]
	if !found {
		return errors.Errorf("feefi needs `exchangeable`")
	}
	feefiExchangeable, ok := exchangeable.(bool)
	if !ok {
		return errors.Errorf("invalid bool value, exchangeable")
	}
	no.Extras["feefi_exchangeable"] = feefiExchangeable

	exchangeCID, found := c["exchange-cid"]
	if !found {
		return errors.Errorf("feefi needs `exchange-cid`")
	}
	feefiExchangeCID := currency.CurrencyID(exchangeCID.(string))
	err = feefiExchangeCID.IsValid(nil)
	if err != nil {
		return errors.Errorf("invalid CurrencyID, exchange-cid")
	}
	no.Extras["feefi_exchange_cid"] = feefiExchangeCID
	/*
		feefier, found := c["feefier"]
		if !found {
			return errors.Errorf("feefi needs `feefier`")
		}
		feefiFeefier := currency.NewAddress(feefier.(string))
		err = feefiFeefier.IsValid(nil)
		if err != nil {
			return errors.Errorf("invalid address, feefier")
		}
		no.Extras["feefi_feefier"] = feefiFeefier
	*/
	exchangeMinAmount, found := c["exchange-min-amount"]
	if !found {
		return errors.Errorf("feefi needs `exchange-min-amount`")
	}

	feefiExchangeMinAmount, err := currency.NewBigFromInterface(exchangeMinAmount)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of exchange-min-amount", exchangeMinAmount)
	}
	no.Extras["feefi_exchange_min_amount"] = feefiExchangeMinAmount

	return nil
}

func (no FeeerDesign) checkRatio(c map[string]interface{}) error {
	if a, found := c["ratio"]; !found {
		return errors.Errorf("ratio needs `ratio`")
	} else if f, ok := a.(float64); !ok {
		return errors.Errorf("invalid ratio value type, %T of ratio; should be float64", a)
	} else {
		no.Extras["ratio_ratio"] = f
	}

	if a, found := c["min"]; !found {
		return errors.Errorf("ratio needs `min`")
	} else if n, err := currency.NewBigFromInterface(a); err != nil {
		return errors.Wrapf(err, "invalid min value, %v of ratio", a)
	} else {
		no.Extras["ratio_min"] = n
	}

	if a, found := c["max"]; found {
		n, err := currency.NewBigFromInterface(a)
		if err != nil {
			return errors.Wrapf(err, "invalid max value, %v of ratio", a)
		}
		no.Extras["ratio_max"] = n
	}

	e, found := c["exchange-min-amount"]
	if !found {
		return errors.Errorf("ratio needs `exchange-min-amount`")
	}

	f, err := currency.NewBigFromInterface(e)
	if err != nil {
		return errors.Wrapf(err, "invalid amount value, %v of exchange-min-amount", e)
	}
	no.Extras["ratio_exchange_min_amount"] = f

	return nil
}
