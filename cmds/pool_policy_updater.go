package cmds

import (
	extensioncmds "github.com/ProtoconNet/mitum-currency-extension/cmds"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/util"
)

type PoolPolicyUpdaterCommand struct {
	*BaseCommand
	OperationFlags
	Sender   AddressFlag                     `arg:"" name:"sender" help:"sender address" required:"true"`
	Pool     currencycmds.AddressFlag        `arg:"" name:"pool" help:"pool address" required:"true"`
	PoolID   extensioncmds.ContractIDFlag    `arg:"" name:"feefi-pool-id" help:"feefi pool id" required:"true"`
	Fee      currencycmds.CurrencyAmountFlag `arg:"" name:"fee" help:"fee amount" required:"true"`
	Currency currencycmds.CurrencyIDFlag     `arg:"" name:"currency-id" help:"currency id" required:"true"`
	sender   base.Address
	pool     base.Address
}

func NewPoolPolicyUpdaterCommand() PoolPolicyUpdaterCommand {
	return PoolPolicyUpdaterCommand{
		BaseCommand: NewBaseCommand("pool-policy-updater-operation"),
	}
}

func (cmd *PoolPolicyUpdaterCommand) Run(version util.Version) error { // nolint:dupl
	if err := cmd.Initialize(cmd, version); err != nil {
		return errors.Wrap(err, "failed to initialize command")
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op operation.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "failed to create pool-policy-updater operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid pool-policy-updater operation")
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

func (cmd *PoolPolicyUpdaterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}
	if sender, err := cmd.Sender.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else if target, err := cmd.Pool.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid pool address format, %q", cmd.Pool.String())
	} else {
		cmd.sender = sender
		cmd.pool = target
	}

	return nil
}

func (cmd *PoolPolicyUpdaterCommand) createOperation() (feefi.PoolPolicyUpdater, error) {
	am := currency.NewAmount(cmd.Fee.Big, cmd.Fee.CID)
	fact := feefi.NewPoolPolicyUpdaterFact([]byte(cmd.Token), cmd.sender, cmd.pool, am, cmd.PoolID.ID, cmd.Currency.CID)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(
		cmd.OperationFlags.Privatekey,
		fact,
		[]byte(cmd.OperationFlags.NetworkID),
	)
	if err != nil {
		return feefi.PoolPolicyUpdater{}, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.OperationFlags.Privatekey.Publickey(), sig))

	return feefi.NewPoolPolicyUpdater(fact, fs, cmd.OperationFlags.Memo)
}
