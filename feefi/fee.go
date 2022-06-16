package feefi

import (
	"time"

	"github.com/pkg/errors"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	FeeOperationFactType   = hint.Type("mitum-currency-fee-operation-fact")
	FeeOperationFactHint   = hint.NewHint(FeeOperationFactType, "v0.0.1")
	FeeOperationFactHinter = FeeOperationFact{BaseHinter: hint.NewBaseHinter(FeeOperationFactHint)}
	FeeOperationType       = hint.Type("mitum-currency-fee-operation")
	FeeOperationHint       = hint.NewHint(FeeOperationType, "v0.0.1")
	FeeOperationHinter     = FeeOperation{BaseHinter: hint.NewBaseHinter(FeeOperationHint)}
)

type FeeOperationFact struct {
	hint.BaseHinter
	h       valuehash.Hash
	token   []byte
	amounts []currency.Amount
}

func NewFeeOperationFact(height base.Height, ams map[currency.CurrencyID]currency.Big) FeeOperationFact {
	amounts := make([]currency.Amount, len(ams))
	var i int
	for cid := range ams {
		amounts[i] = currency.NewAmount(ams[cid], cid)
		i++
	}

	// TODO replace random bytes with height
	fact := FeeOperationFact{
		BaseHinter: hint.NewBaseHinter(FeeOperationFactHint),
		token:      height.Bytes(), // for unique token
		amounts:    amounts,
	}
	fact.h = valuehash.NewSHA256(fact.Bytes())

	return fact
}

func (fact FeeOperationFact) Hash() valuehash.Hash {
	return fact.h
}

func (fact FeeOperationFact) Bytes() []byte {
	bs := make([][]byte, len(fact.amounts)+1)
	bs[0] = fact.token

	for i := range fact.amounts {
		bs[i+1] = fact.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (fact FeeOperationFact) IsValid([]byte) error {
	if len(fact.token) < 1 {
		return isvalid.InvalidError.Errorf("empty token for FeeOperationFact")
	}

	if err := isvalid.Check(nil, false, fact.h); err != nil {
		return err
	}

	for i := range fact.amounts {
		if err := fact.amounts[i].IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

func (fact FeeOperationFact) Token() []byte {
	return fact.token
}

func (fact FeeOperationFact) Amounts() []currency.Amount {
	return fact.amounts
}

type FeeOperation struct {
	hint.BaseHinter
	fact FeeOperationFact
	h    valuehash.Hash
}

func NewFeeOperation(fact FeeOperationFact) FeeOperation {
	op := FeeOperation{BaseHinter: hint.NewBaseHinter(FeeOperationHint), fact: fact}
	op.h = op.GenerateHash()

	return op
}

func (op FeeOperation) Fact() base.Fact {
	return op.fact
}

func (op FeeOperation) Hash() valuehash.Hash {
	return op.h
}

func (FeeOperation) Signs() []base.FactSign {
	return nil
}

func (op FeeOperation) IsValid([]byte) error {
	if err := isvalid.Check(nil, false, op.BaseHinter, op.h); err != nil {
		return err
	}

	if l := len(op.fact.Token()); l < 1 {
		return isvalid.InvalidError.Errorf("FeeOperation has empty token")
	} else if l > operation.MaxTokenSize {
		return isvalid.InvalidError.Errorf("FeeOperation token size too large: %d > %d", l, operation.MaxTokenSize)
	}

	if err := op.fact.IsValid(nil); err != nil {
		return err
	}

	if !op.Hash().Equal(op.GenerateHash()) {
		return isvalid.InvalidError.Errorf("wrong FeeOperation hash")
	}

	return nil
}

func (op FeeOperation) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(op.Fact().Hash().Bytes())
}

func (FeeOperation) AddFactSigns(...base.FactSign) (base.FactSignUpdater, error) {
	return nil, nil
}

func (FeeOperation) LastSignedAt() time.Time {
	return time.Time{}
}

func (FeeOperation) Process(
	func(key string) (state.State, bool, error),
	func(valuehash.Hash, ...state.State) error,
) error {
	return nil
}

type FeeOperationProcessor struct {
	FeeOperation
	cp *CurrencyPool
}

func NewFeeOperationProcessor(cp *CurrencyPool, op FeeOperation) state.Processor {
	return &FeeOperationProcessor{
		cp:           cp,
		FeeOperation: op,
	}
}

func (opp *FeeOperationProcessor) Process(
	getState func(key string) (state.State, bool, error),
	setState func(valuehash.Hash, ...state.State) error,
) error {
	fact := opp.Fact().(FeeOperationFact)

	var sts []state.State
	// AmountState map of fixed receiver
	fixedReceiverBalance := make(map[string]currency.AmountState)
	// AmountState map of feefi feefier
	feefiFeefierBalance := make(map[string]extensioncurrency.AmountState)
	// AmountState map of feefi receiver
	feefiReceiverBalance := make(map[string]currency.AmountState)
	for i := range fact.amounts {
		am := fact.amounts[i]
		var feeer extensioncurrency.Feeer
		j, found := opp.cp.Feeer(am.Currency())
		if !found {
			return errors.Errorf("unknown currency id, %q found for FeeOperation", am.Currency())
		}
		feeer = j
		t := feeer.Type()
		switch t {
		case FeeerFixed, FeeerRatio:
			if feeer.Receiver() == nil {
				continue
			}

			if err := checkExistsState(currency.StateKeyAccount(feeer.Receiver()), getState); err != nil {
				return err
			} else if st, _, err := getState(currency.StateKeyBalance(feeer.Receiver(), am.Currency())); err != nil {
				return err
			} else {
				amountst, found := fixedReceiverBalance[am.Currency().String()]
				ra := currency.NewAmountState(st, am.Currency())
				nra := ra.Add(am.Big())
				if !found {
					fixedReceiverBalance[am.Currency().String()] = nra
				} else {
					amount, err := currency.StateBalanceValue(amountst)
					if err != nil {
						return err
					}
					fixedReceiverBalance[am.Currency().String()] = nra.Add(amount.Big())
				}
				// rb := currency.NewAmountState(st, am.Currency())
				// sts[i] = rb.Add(am.Big())
			}
		case FeeerFeefi:
			feefierID := extensioncurrency.ContractID(am.Currency())
			f, ok := feeer.(FeefiFeeer)
			if !ok {
				return errors.Errorf("not FeefiFeeer, %q", feeer)
			}
			var CurrencyReceiverAmountState state.State
			checkErr := true
			// prepare fee receiver amount state
			if err := checkExistsState(currency.StateKeyAccount(feeer.Receiver()), getState); err != nil {
				return err
			} else if CurrencyReceiverAmountState, _, err = getState(currency.StateKeyBalance(feeer.Receiver(), am.Currency())); err != nil {
				return err
			}
			// check exchange currency in currency pool
			exchangeCurrencyFeeer, ok := opp.cp.Feeer(f.exchangeCID)
			if !ok {
				checkErr = false
				// return errors.Errorf("feeer exchange currency not found in currency pool, %q", f.exchangeCID)
			}
			// check exchange currency receiver
			err := checkExistsState(currency.StateKeyAccount(exchangeCurrencyFeeer.Receiver()), getState)
			if err != nil {
				checkErr = false
				// update receiver amount state by adding exchange fee
			}
			exchangeCurrencyReceiverAmountState, _, err := getState(currency.StateKeyBalance(exchangeCurrencyFeeer.Receiver(), f.ExchangeCID()))
			if err != nil {
				checkErr = false
			}
			// check feefier exists
			if err := checkExistsState(currency.StateKeyAccount(f.Feefier()), getState); err != nil {
				checkErr = false
				// check whether feefier is contract account
				// TODO:check whether contract account deactivated
			} else if _, err := existsState(extensioncurrency.StateKeyContractAccount(f.Feefier()), "contract account", getState); err != nil {
				checkErr = false

			}
			feefierIncomeAmountState, _, err := getState(extensioncurrency.StateKeyBalance(f.Feefier(), feefierID, am.Currency(), StateKeyBalanceSuffix))
			if err != nil {
				checkErr = false
			}
			feefiOutlayAmountState, _, err := getState(extensioncurrency.StateKeyBalance(f.Feefier(), feefierID, f.ExchangeCID(), StateKeyBalanceSuffix))
			if err != nil {
				checkErr = false
			}
			amv, err := extensioncurrency.StateBalanceValue(feefiOutlayAmountState)
			if err != nil {
				checkErr = false
			}
			if amv.Amount().Big().Compare(f.ExchangeMin()) < 0 {
				checkErr = false
			}

			if !checkErr {
				if amountst, found := feefiReceiverBalance[f.ExchangeCID().String()]; !found {
					ra := currency.NewAmountState(CurrencyReceiverAmountState, f.ExchangeCID())
					nra := ra.Add(f.ExchangeMin())
					feefiReceiverBalance[f.ExchangeCID().String()] = nra
				} else {
					ra := amountst.Add(f.ExchangeMin())
					feefiReceiverBalance[f.ExchangeCID().String()] = ra
				}
				continue
			}
			sa := extensioncurrency.NewAmountState(feefiOutlayAmountState, f.ExchangeCID(), feefierID)
			nsa := sa.Sub(f.ExchangeMin())
			// update feefier amount state by substracting exchange fee
			if amountst, found := feefiFeefierBalance[am.Currency().String()+"-"+f.ExchangeCID().String()]; found {
				sa := amountst.Sub(f.ExchangeMin())
				feefiFeefierBalance[am.Currency().String()+"-"+f.ExchangeCID().String()] = sa
			} else {
				feefiFeefierBalance[am.Currency().String()+"-"+f.ExchangeCID().String()] = nsa
			}

			// update feefier amount state by adding fee
			if amountst, found := feefiFeefierBalance[am.Currency().String()+"-"+am.Currency().String()]; !found {
				ra := extensioncurrency.NewAmountState(feefierIncomeAmountState, am.Currency(), feefierID)
				nra := ra.Add(am.Big())
				feefiFeefierBalance[am.Currency().String()+"-"+am.Currency().String()] = nra
			} else {
				ra := amountst.Add(am.Big())
				feefiFeefierBalance[am.Currency().String()+"-"+am.Currency().String()] = ra
			}

			if amountst, found := fixedReceiverBalance[f.ExchangeCID().String()]; !found {
				ra := currency.NewAmountState(exchangeCurrencyReceiverAmountState, f.ExchangeCID())
				nra := ra.Add(f.ExchangeMin())
				fixedReceiverBalance[f.ExchangeCID().String()] = nra
			} else {
				ra := amountst.Add(f.ExchangeMin())
				fixedReceiverBalance[f.ExchangeCID().String()] = ra
			}
		default:
			return errors.Errorf("unknown feeer type, %q", t)
		}
	}

	if len(fixedReceiverBalance) > 0 {
		for k := range fixedReceiverBalance {
			am := fixedReceiverBalance[k]
			sts = append(sts, am)
		}
	}

	if len(feefiFeefierBalance) > 0 {
		for k := range feefiFeefierBalance {
			am := feefiFeefierBalance[k]
			sts = append(sts, am)
		}
	}

	if len(feefiReceiverBalance) > 0 {
		for k := range feefiReceiverBalance {
			am := feefiReceiverBalance[k]
			sts = append(sts, am)
		}
	}
	return setState(fact.Hash(), sts...)
}
