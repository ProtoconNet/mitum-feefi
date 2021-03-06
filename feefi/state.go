package feefi

import (
	"fmt"
	"strings"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util"
)

var (
	StateKeyPoolDesignSuffix = ":feefipooldesign"
	StateKeyPoolSuffix       = ":feefipool"
	StateKeyBalanceSuffix    = extensioncurrency.StateKeyBalanceSuffix(":feefibalance")
)

func statePoolKeyPrefix(a base.Address, poolID extensioncurrency.ContractID) string {
	return fmt.Sprintf("%s-%s", a.String(), poolID)
}

func StateKeyPool(a base.Address, poolID extensioncurrency.ContractID) string {
	return fmt.Sprintf("%s%s", statePoolKeyPrefix(a, poolID), StateKeyPoolSuffix)
}

func IsStatePoolKey(key string) bool {
	return strings.HasSuffix(key, StateKeyPoolSuffix)
}

func StatePoolValue(st state.State) (Pool, error) {
	v := st.Value()
	if v == nil {
		return Pool{}, util.NotFoundError.Errorf("feefipool not found in State")
	}

	s, ok := v.Interface().(Pool)
	if !ok {
		return Pool{}, errors.Errorf("invalid feefipool value found, %T", v.Interface())
	}
	return s, nil
}

func setStatePoolValue(st state.State, v Pool) (state.State, error) {
	uv, err := state.NewHintedValue(v)
	if err != nil {
		return nil, err
	}
	return st.SetValue(uv)
}

func StateKeyPoolDesign(a base.Address, poolID extensioncurrency.ContractID) string {
	return fmt.Sprintf("%s%s", statePoolKeyPrefix(a, poolID), StateKeyPoolDesignSuffix)
}

func IsStatePoolDesignKey(key string) bool {
	return strings.HasSuffix(key, StateKeyPoolDesignSuffix)
}

func StatePoolDesignValue(st state.State) (PoolDesign, error) {
	v := st.Value()
	if v == nil {
		return PoolDesign{}, util.NotFoundError.Errorf("feefi pool design not found in State")
	}

	s, ok := v.Interface().(PoolDesign)
	if !ok {
		return PoolDesign{}, errors.Errorf("invalid feefi pool design value found, %T", v.Interface())
	}
	return s, nil
}

func SetStatePoolDesignValue(st state.State, v PoolDesign) (state.State, error) {
	uv, err := state.NewHintedValue(v)
	if err != nil {
		return nil, err
	}
	return st.SetValue(uv)
}

var StateKeyCurrencyDesignPrefix = "feeficurrencydesign:"

func IsStateCurrencyDesignKey(key string) bool {
	return strings.HasPrefix(key, StateKeyCurrencyDesignPrefix)
}

func StateKeyCurrencyDesign(cid currency.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateKeyCurrencyDesignPrefix, cid)
}

func StateCurrencyDesignValue(st state.State) (extensioncurrency.CurrencyDesign, error) {
	v := st.Value()
	if v == nil {
		return extensioncurrency.CurrencyDesign{}, util.NotFoundError.Errorf("currency design not found in State")
	}
	s, ok := v.Interface().(extensioncurrency.CurrencyDesign)
	if !ok {
		return extensioncurrency.CurrencyDesign{}, errors.Errorf("invalid currency design value found, %T", v.Interface())
	}
	return s, nil
}

func SetStateCurrencyDesignValue(st state.State, v extensioncurrency.CurrencyDesign) (state.State, error) {
	uv, err := state.NewHintedValue(v)
	if err != nil {
		return nil, err
	}
	return st.SetValue(uv)
}

/*
func stateBalanceKeyPrefix(a base.Address, poolCID currency.CurrencyID, cid currency.CurrencyID) string {
	return fmt.Sprintf("%s-%s-%s", poolCID, a.String(), cid)
}

func stateKeyBalance(a base.Address, poolCID, cid currency.CurrencyID) string {
	return fmt.Sprintf("%s%s", stateBalanceKeyPrefix(a, poolCID, cid), StateKeyBalanceSuffix)
}

func IsStateBalanceKey(key string) bool {
	return strings.HasSuffix(key, StateKeyBalanceSuffix)
}

func stateBalanceValue(st state.State) (currency.Amount, error) {
	v := st.Value()
	if v == nil {
		return currency.Amount{}, util.NotFoundError.Errorf("balance not found in State")
	}

	s, ok := v.Interface().(currency.Amount)
	if !ok {
		return currency.Amount{}, errors.Errorf("invalid balance value found, %T", v.Interface())
	}
	return s, nil
}

func setStateBalanceValue(st state.State, v currency.Amount) (state.State, error) {
	uv, err := state.NewHintedValue(v)
	if err != nil {
		return nil, err
	}
	return st.SetValue(uv)
}
*/
func checkExistsState(
	key string,
	getState func(key string) (state.State, bool, error),
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case !found:
		return operation.NewBaseReasonError("state, %q does not exist", key)
	default:
		return nil
	}
}

func existsState(
	k,
	name string,
	getState func(key string) (state.State, bool, error),
) (state.State, error) {
	switch st, found, err := getState(k); {
	case err != nil:
		return nil, err
	case !found:
		return nil, operation.NewBaseReasonError("%s does not exist", name)
	default:
		return st, nil
	}
}

func notExistsState(
	k,
	name string,
	getState func(key string) (state.State, bool, error),
) (state.State, error) {
	switch st, found, err := getState(k); {
	case err != nil:
		return nil, err
	case found:
		return nil, operation.NewBaseReasonError("%s already exists", name)
	default:
		return st, nil
	}
}
