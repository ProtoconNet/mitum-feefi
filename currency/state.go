package currency

import (
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/base/state"
)

/*
var StateKeyCurrencyDesignPrefix = "currencydesign:"

func IsStateCurrencyDesignKey(key string) bool {
	return strings.HasPrefix(key, StateKeyCurrencyDesignPrefix)
}

func StateKeyCurrencyDesign(cid currency.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateKeyCurrencyDesignPrefix, cid)
}

func StateCurrencyDesignValue(st state.State) (extensioncmds.CurrencyDesign, error) {
	v := st.Value()
	if v == nil {
		return extensioncmds.CurrencyDesign{}, util.NotFoundError.Errorf("currency design not found in State")
	}

	s, ok := v.Interface().(extensioncmds.CurrencyDesign)
	if !ok {
		return extensioncmds.CurrencyDesign{}, errors.Errorf("invalid currency design value found, %T", v.Interface())
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
