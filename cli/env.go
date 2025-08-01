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
		withSetup(func(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
			var env *models.Env
			err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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
		withConfirm(withSetup(envDelete)),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(confirmFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
	)
}

func envDelete(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	name := mustGetNameArg(cmdCtx.Flags)
	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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

const envListCmdHelpLong = `List environments with optional filtering expression. The expression must return a list of environments.
See https://expr-lang.org/docs/language-definition for details on the expression syntax.
I've also added the following functions to the expression language:

- duration: Takes a string like "52w" or "5d" and returns a duration.
  Example: duration("52w").
  See https://github.com/xhit/go-str2duration for more details
- pathExists: Takes a string and returns true if the path exists in the environment. Example: pathExists(.Name)

Examples:

# list envs that are not paths (i.e., likely to be re-used) or have been updated in the last 90 days
enventory env list --expr 'filter(Envs, not pathExists(.Name) or .UpdateTime > now() - duration("90d"))'

# list envs that start with "test"
enventory env list --expr 'filter(Envs, hasPrefix(.Name, "test"))'

# sort envs by comment
enventory env list --expr 'sortBy(Envs, .Comment, "asc")'`

func EnvListCmd() wargcore.Command {
	return command.New(
		"List environments",
		withSetup(envList),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(widthFlag()),
		command.HelpLong(envListCmdHelpLong),
		command.NewFlag(
			"--expr",
			"Expression to filter environments",
			scalar.String(),
			flag.EnvVars("ENVENTORY_ENV_LIST_EXPR"),
		),
	)
}

func envList(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	var envs []models.Env
	expr := ptrFromMap[string](cmdCtx.Flags, "--expr")

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
		var err error
		// TODO: Pass the expr argument - change nil to actual args later
		envs, err = es.EnvList(ctx, models.EnvListArgs{
			Expr: expr,
		})
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
		withSetup(envShow),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(maskFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(timeZoneFlagMap()),
		command.FlagMap(widthFlag()),
	)
}

func envShow(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	mask := mustGetMaskArg(cmdCtx.Flags)
	name := mustGetNameArg(cmdCtx.Flags)
	timezone := mustGetTimezoneArg(cmdCtx.Flags)
	width := mustGetWidthArg(cmdCtx.Flags)

	var env *models.Env
	var localvars []models.Var
	var refs []models.VarRef
	var referencedVars []models.Var

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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
		withConfirm(withSetup(envUpdate)),
		command.FlagMap(commonUpdateFlags()),
		command.Flag("--name", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.FlagMap(confirmFlag()),
	)
}

func envUpdate(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	// common update flags
	comment := ptrFromMap[string](cmdCtx.Flags, "--comment")
	createTime := ptrFromMap[time.Time](cmdCtx.Flags, "--create-time")
	newName := ptrFromMap[string](cmdCtx.Flags, "--new-name")
	updateTime := ptrFromMap[time.Time](cmdCtx.Flags, "--update-time")

	name := mustGetNameArg(cmdCtx.Flags)

	err := es.WithTx(ctx, func(ctx context.Context, es models.Service) error {
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
