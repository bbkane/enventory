package cli

import (
	"context"
	"fmt"
	"time"

	"go.bbkane.com/enventory/cli/tableprint"
	"go.bbkane.com/enventory/models"

	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/wargcore"
)

func EnvCreateCmd() wargcore.Command {
	var createArgs models.EnvCreateArgs
	return command.New(
		"Create an environment",
		withEnvService(func(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
			var env *models.Env
			err := es.WithTx(ctx, func(es models.EnvService) error {
				var err error
				env, err = es.EnvCreate(ctx, createArgs)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("could not create env: %w", err)
			}

			fmt.Fprintf(cmdCtx.Stdout, "Created env: %s\n", env.Name)
			return nil
		}),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(commonCreateFlagMapPtrs(
			&createArgs.Comment,
			&createArgs.CreateTime,
			&createArgs.UpdateTime,
		)),
		command.NewFlag(
			"--name",
			"Environment name",
			scalar.String(
				scalar.Default(cwd),
				scalar.PointerTo(&createArgs.Name),
			),
			flag.Required(),
		),
	)
}

func EnvDeleteCmd() wargcore.Command {
	return command.New(
		"Delete an environment and associated vars",
		withConfirm(withEnvService(envDelete)),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(confirmFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
	)
}

func envDelete(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	name := mustGetNameArg(cmdCtx.Flags)
	err := es.WithTx(ctx, func(es models.EnvService) error {
		err := es.EnvDelete(ctx, name)
		if err != nil {
			return fmt.Errorf("could not delete env: %s: %w", name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(cmdCtx.Stdout, "deleted: %s\n", name)
	return nil
}

func EnvListCmd() wargcore.Command {
	return command.New(
		"List environments",
		withEnvService(envList),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(widthFlag()),
	)
}

func envList(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	var envs []models.Env
	err := es.WithTx(ctx, func(es models.EnvService) error {
		var err error
		envs, err = es.EnvList(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	c := tableprint.CommonTablePrintArgs{
		Format:          tableprint.Format_Table,
		Mask:            false,
		Tz:              tableprint.Timezone(mustGetTimezoneArg(cmdCtx.Flags)),
		W:               cmdCtx.Stdout,
		DesiredMaxWidth: mustGetWidthArg(cmdCtx.Flags),
	}

	tableprint.EnvList(c, envs)
	return nil
}

func EnvShowCmd() wargcore.Command {
	return command.New(
		"Print environment details",
		withEnvService(envShow),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(maskFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(widthFlag()),
	)
}

func envShow(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	mask := mustGetMaskArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	width := mustGetWidthArg(cmdCtx.Flags)

	var env *models.Env
	var localvars []models.Var
	var refs []models.VarRef
	var referencedVars []models.Var

	err := es.WithTx(ctx, func(es models.EnvService) error {
		var err error
		env, err = es.EnvShow(ctx, name)
		if err != nil {
			return fmt.Errorf("could not show env: %s: %w", name, err)
		}

		localvars, err = es.VarList(ctx, name)
		if err != nil {
			return err
		}

		refs, referencedVars, err = es.VarRefList(ctx, name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	c := tableprint.CommonTablePrintArgs{
		Format:          tableprint.Format_Table,
		Mask:            mask,
		Tz:              tableprint.Timezone(timezone),
		W:               cmdCtx.Stdout,
		DesiredMaxWidth: width,
	}
	tableprint.EnvShowRun(c, *env, localvars, refs, referencedVars)
	return nil
}

func EnvUpdateCmd() wargcore.Command {
	return command.New(
		"Update an environment",
		withConfirm(withEnvService(envUpdate)),
		command.FlagMap(commonUpdateFlags()),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(confirmFlag()),
	)
}

func envUpdate(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	// common update flags
	comment := ptrFromMap[string](cmdCtx.Flags, "--comment")
	createTime := ptrFromMap[time.Time](cmdCtx.Flags, "--create-time")
	newName := ptrFromMap[string](cmdCtx.Flags, "--new-name")
	updateTime := ptrFromMap[time.Time](cmdCtx.Flags, "--update-time")

	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(es models.EnvService) error {
		err := es.EnvUpdate(ctx, name, models.EnvUpdateArgs{
			Comment:    comment,
			CreateTime: createTime,
			Name:       newName,
			UpdateTime: updateTime,
		})
		if err != nil {
			return fmt.Errorf("could not update env: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	finalName := name
	if newName != nil {
		finalName = *newName
	}
	fmt.Fprintln(cmdCtx.Stdout, "updated env:", finalName)
	return nil
}
