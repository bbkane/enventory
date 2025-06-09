package cli

import (
	"context"
	"fmt"

	"go.bbkane.com/enventory/cli/tableprint"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/wargcore"
)

func VarRefCreateCmd() wargcore.Command {
	return command.New(
		"Create a reference in this env to a variable in another env",
		withEnvService(varRefCreateRun),
		command.NewFlag(
			"--name",
			"Ref name",
			scalar.String(),
			flag.Required(),
		),
		command.NewFlag(
			"--ref-env",
			"Environment we're referencing",
			scalar.String(),
			flag.Required(),
			flag.CompletionCandidates(withEnvServiceCompletions(completeExistingEnvName)),
		),
		command.NewFlag(
			"--ref-var",
			"Variable we're referencing",
			scalar.String(),
			flag.Required(),
			flag.CompletionCandidates(withEnvServiceCompletions(completeExistingRefEnvVarName)),
		),
		command.FlagMap(commonCreateFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeoutFlagMap()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varRefCreateRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	// common create Flags
	commonCreateArgs := mustGetCommonCreateArgs(cmdCtx.Flags)

	name := mustGetNameArg(cmdCtx.Flags)
	refEnvName := cmdCtx.Flags["--ref-env"].(string)
	refVarName := cmdCtx.Flags["--ref-var"].(string)

	envName := mustGetEnvNameArg(cmdCtx.Flags)

	_, err := es.VarRefCreate(
		ctx,
		models.VarRefCreateArgs{
			EnvName:    envName,
			Name:       name,
			Comment:    commonCreateArgs.Comment,
			CreateTime: commonCreateArgs.CreateTime,
			UpdateTime: commonCreateArgs.UpdateTime,
			RefEnvName: refEnvName,
			RefVarName: refVarName,
		},
	)

	if err != nil {
		return fmt.Errorf("couldn't create env ref: %s: %w", name, err)
	}

	fmt.Fprintf(cmdCtx.Stdout, "Created env ref: %s: %s\n", envName, name)
	return nil
}

func VarRefDeleteCmd() wargcore.Command {
	return command.New(
		"Delete a reference to a variablea",
		withConfirm(withEnvService(varRefDeleteRun)),
		command.FlagMap(confirmFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.Flag("--name", varRefFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varRefDeleteRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)

	name := mustGetNameArg(cmdCtx.Flags)

	err := es.VarRefDelete(ctx, envName, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmdCtx.Stdout, "Deleted %s: %s\n", envName, name)
	return nil
}

func VarRefShowCmd() wargcore.Command {
	return command.New(
		"Show details for a reference",
		withEnvService(varRefShowRun),
		command.FlagMap(maskFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(formatFlag()),
		command.FlagMap(widthFlag()),
		command.Flag("--name", varRefFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varRefShowRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	mask := mustGetMaskArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	format := cmdCtx.Flags["--format"].(string)
	width := mustGetWidthArg(cmdCtx.Flags)

	envRef, envVar, err := es.VarRefShow(ctx, envName, name)
	if err != nil {
		return fmt.Errorf("couldn't find env var: %s: %w", name, err)
	}
	c := tableprint.CommonTablePrintArgs{
		Format:          tableprint.Format(format),
		Mask:            mask,
		Tz:              tableprint.Timezone(timezone),
		W:               cmdCtx.Stdout,
		DesiredMaxWidth: width,
	}

	tableprint.VarRefShowPrint(c, *envRef, *envVar)
	return nil
}
