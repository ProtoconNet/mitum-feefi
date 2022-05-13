package cmds

import (
	extensioncmds "github.com/ProtoconNet/mitum-currency-extension/cmds"
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/util"
)

type CurrencyPolicyUpdaterCommand struct {
	*BaseCommand
	OperationFlags
	Currency                              currencycmds.CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	currencycmds.CurrencyPolicyFlags      `prefix:"policy-" help:"currency policy" required:"true"`
	FeeerString                           string `name:"feeer" help:"feeer type, {nil, feefi, fixed, ratio}" required:"true"`
	CurrencyFeefiFeeerFlags               `prefix:"feeer-feefi-" help:"feefi feeer"`
	extensioncmds.CurrencyFixedFeeerFlags `prefix:"feeer-fixed-" help:"fixed feeer"`
	extensioncmds.CurrencyRatioFeeerFlags `prefix:"feeer-ratio-" help:"ratio feeer"`
	po                                    extensioncurrency.CurrencyPolicy
}

func NewCurrencyPolicyUpdaterCommand() CurrencyPolicyUpdaterCommand {
	return CurrencyPolicyUpdaterCommand{
		BaseCommand: NewBaseCommand("currency-policy-updater-operation"),
	}
}

func (cmd *CurrencyPolicyUpdaterCommand) Run(version util.Version) error { // nolint:dupl
	if err := cmd.Initialize(cmd, version); err != nil {
		return errors.Wrap(err, "failed to initialize command")
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op operation.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "failed to create currency-policy-updater operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid currency-policy-updater operation")
	} else {
		cmd.Log().Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	i, err := operation.NewBaseSeal(
		cmd.OperationFlags.Privatekey,
		[]operation.Operation{op},
		[]byte(cmd.OperationFlags.NetworkID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create operation.Seal")
	}
	cmd.Log().Debug().Interface("seal", i).Msg("seal loaded")

	PrettyPrint(cmd.Out, cmd.Pretty, i)

	return nil
}

func (cmd *CurrencyPolicyUpdaterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyPolicyFlags.IsValid(nil); err != nil {
		return err
	}

	if err := cmd.CurrencyFixedFeeerFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyRatioFeeerFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyFeefiFeeerFlags.IsValid(nil); err != nil {
		return err
	}

	var feeer feefi.Feeer
	switch t := cmd.FeeerString; t {
	case extensioncurrency.FeeerNil, "":
		feeer = extensioncurrency.NewNilFeeer()
	case extensioncurrency.FeeerFixed:
		feeer = cmd.CurrencyFixedFeeerFlags.Feeer()
	case feefi.FeeerFeefi:
		feeer = cmd.CurrencyFeefiFeeerFlags.Feeer()
	case extensioncurrency.FeeerRatio:
		feeer = cmd.CurrencyRatioFeeerFlags.Feeer()
	default:
		return errors.Errorf("unknown feeer type, %q", t)
	}

	if feeer == nil {
		return errors.Errorf("empty feeer flags")
	} else if err := feeer.IsValid(nil); err != nil {
		return err
	}

	cmd.po = extensioncurrency.NewCurrencyPolicy(cmd.CurrencyPolicyFlags.NewAccountMinBalance.Big, feeer)
	if err := cmd.po.IsValid(nil); err != nil {
		return err
	}

	cmd.Log().Debug().Interface("currency-policy", cmd.po).Msg("currency policy loaded")

	return nil
}

func (cmd *CurrencyPolicyUpdaterCommand) createOperation() (extensioncurrency.CurrencyPolicyUpdater, error) {
	fact := extensioncurrency.NewCurrencyPolicyUpdaterFact([]byte(cmd.Token), cmd.Currency.CID, cmd.po)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(
		cmd.OperationFlags.Privatekey,
		fact,
		[]byte(cmd.OperationFlags.NetworkID),
	)
	if err != nil {
		return extensioncurrency.CurrencyPolicyUpdater{}, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.OperationFlags.Privatekey.Publickey(), sig))

	return extensioncurrency.NewCurrencyPolicyUpdater(fact, fs, cmd.OperationFlags.Memo)
}
