package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"go.bbkane.com/enventory/cli/tableprint"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/wargcore"
)

func VarCreateCmd() wargcore.Command {
	return command.New(
		"Create a variable local to the this env",
		withEnvService(varCreateRun),
		command.Flag(
			"--env",
			envNameFlag(),
		),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(commonCreateFlagMap()),
		command.NewFlag(
			"--name",
			"New env var name",
			scalar.String(),
			flag.Required(),
		),
		command.NewFlag(
			"--value",
			"New env var value",
			scalar.String(),
		),
	)
}

func varCreateRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {

	// common create Flags
	commonCreateArgs := mustGetCommonCreateArgs(cmdCtx.Flags)

	envName := mustGetEnvNameArg(cmdCtx.Flags)
	value, exists := cmdCtx.Flags["--value"].(string)
	if !exists {
		fmt.Print("Enter value: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Err() != nil {
			return fmt.Errorf("couldn't read --value: %w", scanner.Err())
		}
		value = scanner.Text()
	}

	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(es models.EnvService) error {
		_, err := es.VarCreate(
			ctx,
			models.VarCreateArgs{
				EnvName:    envName,
				Name:       name,
				Comment:    commonCreateArgs.Comment,
				CreateTime: commonCreateArgs.CreateTime,
				UpdateTime: commonCreateArgs.UpdateTime,
				Value:      value,
			},
		)
		if err != nil {
			return fmt.Errorf("couldn't create env var: %s: %w", name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmdCtx.Stdout, "Created env var: %s: %s\n", envName, name)
	return nil
}

func VarDeleteCmd() wargcore.Command {
	return command.New(
		"Delete a variable local to the this env",
		withConfirm(withEnvService(varDeleteRun)),
		command.FlagMap(confirmFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.Flag("--name", varNameFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varDeleteRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(es models.EnvService) error {
		err := es.VarDelete(ctx, envName, name)
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

func VarShowCmd() wargcore.Command {
	return command.New(
		"Show details for a local var",
		withEnvService(varShowRun),
		command.FlagMap(maskFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(formatFlag()),
		command.FlagMap(widthFlag()),
		command.Flag("--name", varNameFlag()),
		command.Flag(
			"--env",
			envNameFlag(),
		),
	)
}

func varShowRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {

	mask := mustGetMaskArg(cmdCtx.Flags)
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	format := cmdCtx.Flags["--format"].(string)
	width := mustGetWidthArg(cmdCtx.Flags)

	var envVar *models.Var
	var envRefs []models.VarRef
	err := es.WithTx(ctx, func(es models.EnvService) error {
		var err error
		envVar, envRefs, err = es.VarShow(ctx, envName, name)
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

	tableprint.VarShowPrint(c, *envVar, envRefs)
	return nil
}

func VarUpdateCmd() wargcore.Command {
	return command.New(
		"Update and env var",
		withConfirm(withEnvService(varUpdateRun)),
		command.Flag("--env", envNameFlag()),
		command.FlagMap(commonUpdateFlags()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(confirmFlag()),
		command.Flag("--name", varNameFlag()),
		command.NewFlag(
			"--new-env",
			"New env name",
			scalar.String(),
			flag.CompletionCandidates(withEnvServiceCompletions(
				completeExistingEnvName)),
		),
		command.NewFlag(
			"--value",
			"New value for this env var",
			scalar.String(),
		),
	)
}

func varUpdateRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	// common update flags
	commonUpdateArgs := getCommonUpdateArgs(cmdCtx.Flags)

	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	newEnvName := ptrFromMap[string](cmdCtx.Flags, "--new-env")
	value := ptrFromMap[string](cmdCtx.Flags, "--value")

	err := es.WithTx(ctx, func(es models.EnvService) error {
		err := es.VarUpdate(ctx, envName, name, models.VarUpdateArgs{
			Comment:    commonUpdateArgs.Comment,
			CreateTime: commonUpdateArgs.CreateTime,
			EnvName:    newEnvName,
			Name:       commonUpdateArgs.NewName,
			UpdateTime: commonUpdateArgs.UpdateTime,
			Value:      value,
		})
		if err != nil {
			return fmt.Errorf("could not update env var: %w", err)
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
	fmt.Fprintf(cmdCtx.Stdout, "updated env var:  %s: %s\n", finalEnvName, finalName)
	return nil
}
