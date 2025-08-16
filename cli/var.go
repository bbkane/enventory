package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"go.bbkane.com/enventory/cli/tableprint"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/value/scalar"
)

func VarCreateCmd() warg.Cmd {
	return warg.NewCmd(
		"Create a variable local to the this env",
		withSetup(varCreateRun),
		warg.CmdFlag(
			"--env",
			envNameFlag(),
		),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(commonCreateFlagMap()),
		warg.NewCmdFlag(
			"--name",
			"New env var name",
			scalar.String(),
			warg.Required(),
		),
		warg.NewCmdFlag(
			"--value",
			"New env var value",
			scalar.String(),
		),
	)
}

func varCreateRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {

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

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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

func VarDeleteCmd() warg.Cmd {
	return warg.NewCmd(
		"Delete a variable local to the this env",
		withConfirm(withSetup(varDeleteRun)),
		warg.CmdFlagMap(confirmFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlag("--name", varNameFlag()),
		warg.CmdFlag(
			"--env",
			envNameFlag(),
		),
	)
}

func varDeleteRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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

func VarShowCmd() warg.Cmd {
	return warg.NewCmd(
		"Show details for a local var",
		withSetup(varShowRun),
		warg.CmdFlagMap(maskFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(timeZoneFlagMap()),
		warg.CmdFlagMap(formatFlag()),
		warg.CmdFlagMap(widthFlag()),
		warg.CmdFlag("--name", varNameFlag()),
		warg.CmdFlag(
			"--env",
			envNameFlag(),
		),
	)
}

func varShowRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {

	mask := mustGetMaskArg(cmdCtx.Flags)
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	format := cmdCtx.Flags["--format"].(string)
	width := mustGetWidthArg(cmdCtx.Flags)

	var envVar *models.Var
	var envRefs []models.VarRef
	err := es.WithTx(ctx, func(ctxt context.Context, es models.Service) error {
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

func VarUpdateCmd() warg.Cmd {
	return warg.NewCmd(
		"Update and env var",
		withConfirm(withSetup(varUpdateRun)),
		warg.CmdFlag("--env", envNameFlag()),
		warg.CmdFlagMap(commonUpdateFlags()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(confirmFlag()),
		warg.CmdFlag("--name", varNameFlag()),
		warg.NewCmdFlag(
			"--new-env",
			"New env name",
			scalar.String(),
			warg.FlagCompletions(withEnvServiceCompletions(
				completeExistingEnvName)),
		),
		warg.NewCmdFlag(
			"--value",
			"New value for this env var",
			scalar.String(),
		),
	)
}

func varUpdateRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
	// common update flags
	commonUpdateArgs := getCommonUpdateArgs(cmdCtx.Flags)

	envName := mustGetEnvNameArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	newEnvName := ptrFromMap[string](cmdCtx.Flags, "--new-env")
	value := ptrFromMap[string](cmdCtx.Flags, "--value")

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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
