package cmds

import (
	"github.com/pkg/errors"
	"github.com/protoconNet/mitum-account-extension/extension"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/util"

	currencycmds "github.com/spikeekips/mitum-currency/cmds"
)

type DeactivateCommand struct {
	*BaseCommand
	OperationFlags
	Sender   AddressFlag                 `arg:"" name:"sender" help:"sender address" required:"true"`
	Target   AddressFlag                 `arg:"" name:"target" help:"target address" required:"true"`
	Currency currencycmds.CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	sender   base.Address
	target   base.Address
}

func NewDeactivateCommand() DeactivateCommand {
	return DeactivateCommand{
		BaseCommand: NewBaseCommand("deactivate-operation"),
	}
}

func (cmd *DeactivateCommand) Run(version util.Version) error { // nolint:dupl
	if err := cmd.Initialize(cmd, version); err != nil {
		return errors.Wrap(err, "failed to initialize command")
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	bs, err := operation.NewBaseSeal(
		cmd.Privatekey,
		[]operation.Operation{op},
		cmd.NetworkID.NetworkID(),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create operation.Seal")
	}
	PrettyPrint(cmd.Out, cmd.Pretty, bs)

	return nil
}

func (cmd *DeactivateCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(jenc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	}
	cmd.sender = a

	a, err = cmd.Target.Encode(jenc)
	if err != nil {
		return errors.Wrapf(err, "invalid target format, %q", cmd.Target.String())
	}
	cmd.target = a

	return nil
}

func (cmd *DeactivateCommand) createOperation() (operation.Operation, error) {
	fact := extension.NewDeactivateFact(
		[]byte(cmd.Token),
		cmd.sender,
		cmd.target,
		cmd.Currency.CID,
	)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(cmd.Privatekey, fact, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.Privatekey.Publickey(), sig))

	op, err := extension.NewDeactivate(fact, fs, cmd.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key-updater operation")
	}
	return op, nil
}
