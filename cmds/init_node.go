package cmds

import (
	"context"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	mitumcmds "github.com/spikeekips/mitum/launch/cmds"
	"github.com/spikeekips/mitum/launch/config"
	"github.com/spikeekips/mitum/launch/pm"
	"github.com/spikeekips/mitum/launch/process"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"gopkg.in/yaml.v3"
)

var (
	GenesisAccountKey = "genesis_account"
	GenesisBalanceKey = "genesis_balance"
)

var InitCommandHooks = func(cmd *InitCommand) []pm.Hook {
	genesisOperationHandlers := map[string]process.HookHandlerGenesisOperations{
		"genesis-currencies": GenesisOperationsHandlerGenesisCurrencies,
	}

	for k, v := range process.DefaultHookHandlersGenesisOperations {
		genesisOperationHandlers[k] = v
	}

	return []pm.Hook{
		pm.NewHook(pm.HookPrefixPre, process.ProcessNameProposalProcessor,
			"initialize_proposal_processor", cmd.hookInitializeProposalProcessor).SetOverride(true),
		pm.NewHook(pm.HookPrefixPost, process.ProcessNameConfig,
			process.HookNameConfigGenesisOperations, process.HookGenesisOperationFunc(genesisOperationHandlers)).
			SetOverride(true),
	}
}

type InitCommand struct {
	*BaseNodeCommand
	*mitumcmds.InitCommand
}

func NewInitCommand(dryrun bool) (InitCommand, error) {
	co := mitumcmds.NewInitCommand(dryrun)
	cmd := InitCommand{
		InitCommand:     &co,
		BaseNodeCommand: NewBaseNodeCommand(co.Logging),
	}

	ps, err := cmd.BaseProcesses(co.Processes())
	if err != nil {
		return cmd, err
	}

	hooks := InitCommandHooks(&cmd)
	for i := range hooks {
		if err := hooks[i].Add(ps); err != nil {
			return cmd, err
		}
	}

	_ = cmd.SetProcesses(ps)

	return cmd, nil
}

func (*InitCommand) hookInitializeProposalProcessor(ctx context.Context) (context.Context, error) {
	var oprs *hint.Hintmap
	if err := process.LoadOperationProcessorsContextValue(ctx, &oprs); err != nil {
		if !errors.Is(err, util.ContextValueNotFoundError) {
			return ctx, err
		}
	}

	if oprs == nil {
		oprs = hint.NewHintmap()

		ctx = context.WithValue(ctx, process.ContextValueOperationProcessors, oprs)
	}

	return ctx, nil
}

func GenesisOperationsHandlerGenesisCurrencies(
	ctx context.Context,
	m map[string]interface{},
) (operation.Operation, error) {
	var conf config.LocalNode
	if err := config.LoadConfigContextValue(ctx, &conf); err != nil {
		return nil, err
	}

	var de *GenesisCurrenciesDesign
	if b, err := yaml.Marshal(m); err != nil {
		return nil, err
	} else if err := yaml.Unmarshal(b, &de); err != nil {
		return nil, err
	}

	if err := de.IsValid(nil); err != nil {
		return nil, err
	}

	cds := make([]extensioncurrency.CurrencyDesign, len(de.Currencies))
	for i := range de.Currencies {
		c := de.Currencies[i]

		j, err := loadCurrencyDesign(*c, de.AccountKeys.Address)
		if err != nil {
			return nil, err
		}
		cds[i] = j
	}

	if op, err := feefi.NewGenesisCurrencies(
		conf.Privatekey(),
		de.AccountKeys.Keys,
		cds,
		conf.NetworkID(),
	); err != nil {
		return nil, err
	} else if err := op.IsValid(conf.NetworkID()); err != nil {
		return nil, err
	} else {
		return op, nil
	}
}

func loadCurrencyDesign(de CurrencyDesign, ga base.Address) (extensioncurrency.CurrencyDesign, error) {
	j, err := loadGenesisCurrenciesFeeer(*de.Feeer, ga)
	if err != nil {
		return extensioncurrency.CurrencyDesign{}, err
	}
	po := extensioncurrency.NewCurrencyPolicy(de.NewAccountMinBalance, j)

	cd := extensioncurrency.NewCurrencyDesign(de.Balance, nil, po)
	if err := cd.IsValid(nil); err != nil {
		return extensioncurrency.CurrencyDesign{}, err
	}

	return cd, nil
}

func loadGenesisCurrenciesFeeer(de FeeerDesign, ga base.Address) (extensioncurrency.Feeer, error) {
	var feeer extensioncurrency.Feeer
	switch de.Type {
	case extensioncurrency.FeeerNil, "":
		return extensioncurrency.NewNilFeeer(), nil
	case extensioncurrency.FeeerFixed:
		amount, found := de.Extras["fixed_amount"]
		if !found {
			return extensioncurrency.NewNilFeeer(), nil
		}
		exchangeMinAmount, found := de.Extras["fixed_exchange_min_amount"]
		if !found {
			return extensioncurrency.NewNilFeeer(), nil
		}
		if !de.Extras["fixed_exchange_min_amount"].(currency.Big).OverNil() {
			return extensioncurrency.NewNilFeeer(), nil
		}
		feeer = extensioncurrency.NewFixedFeeer(ga, amount.(currency.Big), exchangeMinAmount.(currency.Big))
	case feefi.FeeerFeefi:
		return nil, errors.Errorf("not used type of feeer in genesis currency, %q", feefi.FeeerFeefi)
		/*
			amount, found := de.Extras["feefi_amount"]
			if !found {
				return extensioncurrency.NewNilFeeer(), nil
			}
			exchangeable, found := de.Extras["feefi_exchangeable"]
			if !found {
				return extensioncurrency.NewNilFeeer(), nil
			}

			feefier, found := de.Extras["feefi_feefier"]
			if !found {
				return extensioncurrency.NewNilFeeer(), nil
			}
			exchangeCID, found := de.Extras["feefi_exchange_cid"]
			if !found {
				return extensioncurrency.NewNilFeeer(), nil
			}
			if len(de.Extras["feefi_exchange_cid"].(currency.CurrencyID)) < 1 {
				return extensioncurrency.NewNilFeeer(), nil
			}
			exchangeMinAmount, found := de.Extras["feefi_exchange_min_amount"]
			if !found {
				return extensioncurrency.NewNilFeeer(), nil
			}
			if !de.Extras["feefi_exchange_min_amount"].(currency.Big).OverNil() {
				return extensioncurrency.NewNilFeeer(), nil
			}

			feeer = feefi.NewFeefiFeeer(ga, amount.(currency.Big), exchangeable.(bool), exchangeCID.(currency.CurrencyID), feefier.(base.Address), exchangeMinAmount.(currency.Big))
		*/

	case extensioncurrency.FeeerRatio:
		var max currency.Big
		if i, found := de.Extras["ratio_max"]; !found {
			max = currency.UnlimitedMaxFeeAmount
		} else {
			max = i.(currency.Big)
		}

		feeer = extensioncurrency.NewRatioFeeer(
			ga,
			de.Extras["ratio_ratio"].(float64),
			de.Extras["ratio_min"].(currency.Big),
			max,
			de.Extras["fixed_exchange_min_amount"].(currency.Big),
		)
	default:
		return nil, errors.Errorf("unknown type of feeer, %q", de.Type)
	}

	if err := feeer.IsValid(nil); err != nil {
		return nil, err
	}

	return feeer, nil
}
