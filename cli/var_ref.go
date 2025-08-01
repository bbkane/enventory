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
		withSetup(varRefCreateRun),
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

func varRefCreateRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	// common create Flags
	commonCreateArgs := mustGetCommonCreateArgs(cmdCtx.Flags)

	name := mustGetNameArg(cmdCtx.Flags)
	refEnvName := cmdCtx.Flags["--ref-env"].(string)
	refVarName := cmdCtx.Flags["--ref-var"].(string)

	envName := mustGetEnvNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmdCtx.Stdout, "Created env ref: %s: %s\n", envName, name)
	return nil
}

func VarRefDeleteCmd() wargcore.Command {
	return command.New(
		"Delete a reference to a variablea",
		withConfirm(withSetup(varRefDeleteRun)),
		command.FlagMap(confirmFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.Flag("--name", varRefNameFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varRefDeleteRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)

	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
		err := es.VarRefDelete(ctx, envName, name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmdCtx.Stdout, "Deleted %s: %s\n", envName, name)
	return nil
}

func VarRefShowCmd() wargcore.Command {
	return command.New(
		"Show details for a reference",
		withSetup(varRefShowRun),
		command.FlagMap(maskFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(formatFlag()),
		command.FlagMap(widthFlag()),
		command.Flag("--name", varRefNameFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varRefShowRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	mask := mustGetMaskArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	format := cmdCtx.Flags["--format"].(string)
	width := mustGetWidthArg(cmdCtx.Flags)

	var envRef *models.VarRef
	var envVar *models.Var
	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
		var err error
		envRef, envVar, err = es.VarRefShow(ctx, envName, name)
		if err != nil {
			return fmt.Errorf("couldn't find env var: %s: %w", name, err)
		}
		return nil
	})
	if err != nil {
		return err
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

func VarRefUpdateCmd() wargcore.Command {
	return command.New(
		"Update a var ref",
		withConfirm(withSetup(varRefUpdateRun)),
		command.Flag("--env", envNameFlag()),
		command.FlagMap(commonUpdateFlags()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(confirmFlag()),
		command.Flag("--name", varRefNameFlag()),
		command.NewFlag(
			"--new-env",
			"New env name",
			scalar.String(),
		),
		command.NewFlag(
			"--ref-env",
			"New environment we're referencing",
			scalar.String(),
			flag.CompletionCandidates(withEnvServiceCompletions(completeExistingEnvName)),
		),
		command.NewFlag(
			"--ref-var",
			"New variable we're referencing",
			scalar.String(),
			flag.CompletionCandidates(withEnvServiceCompletions(completeExistingRefEnvVarName)),
		),
	)
}

func varRefUpdateRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	// common update flags
	commonUpdateArgs := getCommonUpdateArgs(cmdCtx.Flags)

	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	newEnvName := ptrFromMap[string](cmdCtx.Flags, "--new-env")
	refEnvName := ptrFromMap[string](cmdCtx.Flags, "--ref-env")
	refVarName := ptrFromMap[string](cmdCtx.Flags, "--ref-var")

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
		err := es.VarRefUpdate(ctx, envName, name, models.VarRefUpdateArgs{
			Comment:    commonUpdateArgs.Comment,
			CreateTime: commonUpdateArgs.CreateTime,
			EnvName:    newEnvName,
			Name:       commonUpdateArgs.NewName,
			UpdateTime: commonUpdateArgs.UpdateTime,
			RefEnvName: refEnvName,
			RefVarName: refVarName,
		})
		if err != nil {
			return fmt.Errorf("could not update var ref: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	finalName := name
	if commonUpdateArgs.NewName != nil {
		finalName = *commonUpdateArgs.NewName
	}

	finalEnvName := envName
	if newEnvName != nil {
		finalEnvName = *newEnvName
	}

	fmt.Printf("Updated var ref %s:%s\n", finalEnvName, finalName)
	return nil
}
