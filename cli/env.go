package cli

import (
	"context"
	"fmt"
	"time"

	"go.bbkane.com/enventory/cli/tableprint"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg"

	"go.bbkane.com/warg/value/scalar"
)

func EnvCreateCmd() warg.Cmd {
	var createArgs models.EnvCreateArgs
	return warg.NewCmd(
		"Create an environment",
		withSetup(func(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
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
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(commonCreateFlagMapPtrs(
			&createArgs.Comment,
			&createArgs.CreateTime,
			&createArgs.UpdateTime,
		)),
		warg.NewCmdFlag(
			"--name",
			"Environment name",
			scalar.String(
				scalar.Default(cwd),
				scalar.PointerTo(&createArgs.Name),
			),
			warg.Required(),
		),
	)
}

func EnvDeleteCmd() warg.Cmd {
	return warg.NewCmd(
		"Delete an environment and associated vars",
		withConfirm(withSetup(envDelete)),
		warg.CmdFlag("--name", envNameFlag()),
		warg.CmdFlagMap(confirmFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
	)
}

func envDelete(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
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

func EnvListCmd() warg.Cmd {
	return warg.NewCmd(
		"List environments",
		withSetup(envList),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(timeZoneFlagMap()),
		warg.CmdFlagMap(widthFlag()),
		warg.CmdHelpLong(envListCmdHelpLong),
		warg.NewCmdFlag(
			"--expr",
			"Expression to filter environments",
			scalar.String(),
			warg.EnvVars("ENVENTORY_ENV_LIST_EXPR"),
			warg.UnsetSentinel("UNSET"),
		),
	)
}

func envList(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
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

func EnvShowCmd() warg.Cmd {
	return warg.NewCmd(
		"Print environment details",
		withSetup(envShow),
		warg.CmdFlag("--name", envNameFlag()),
		warg.CmdFlagMap(maskFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(timeZoneFlagMap()),
		warg.CmdFlagMap(widthFlag()),
	)
}

func envShow(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
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

func EnvUpdateCmd() warg.Cmd {
	return warg.NewCmd(
		"Update an environment",
		withConfirm(withSetup(envUpdate)),
		warg.CmdFlagMap(commonUpdateFlags()),
		warg.CmdFlag("--name", envNameFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(confirmFlag()),
	)
}

func envUpdate(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
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
