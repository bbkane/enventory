package cli

import (
	"context"
	"errors"
	"fmt"

	"al.essio.dev/pkg/shellescape"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/wargcore"
)

func ShellZshInitCmd() wargcore.Command {
	return command.New(
		"Prints the zsh initialization script",
		shellZshInitRun,
		command.NewFlag(
			"--print-autoload",
			"Include autoload -Uz add-zsh-hook line (might not be needed if you already autoloaded it)",
			scalar.Bool(
				scalar.Default(true),
			),
			flag.Required(),
		),
		command.NewFlag(
			"--print-export-env",
			"Include export-env/unexport-env to easily use envs from the CLI",
			scalar.Bool(
				scalar.Default(true),
			),
			flag.Required(),
		),
		command.NewFlag(
			"--print-chpwd-hook",
			"Include hook to export/unexport envs when changing directories",
			scalar.Bool(
				scalar.Default(true),
			),
			flag.Required(),
		),
	)
}

func shellZshInitRun(cmdCtx wargcore.Context) error {

	printAutoload := cmdCtx.Flags["--print-autoload"].(bool)
	printChpwdHook := cmdCtx.Flags["--print-chpwd-hook"].(bool)
	printExportEnv := cmdCtx.Flags["--print-export-env"].(bool)

	prelude := `
# https://github.com/bbkane/enventory/
#
# To initialize enventory, add this to your configuration (usually ~/.zshrc):
#
# eval "$(enventory shell zsh init)"
#
`
	fmt.Fprint(cmdCtx.Stdout, prelude)

	autoload := `
autoload -Uz add-zsh-hook
`
	if printAutoload {
		fmt.Fprint(cmdCtx.Stdout, autoload)
	}

	chpwdHook := `
add-zsh-hook -Uz chpwd (){
    eval $(enventory shell zsh unexport --env "$OLDPWD" --no-env-no-problem true)
    eval $(enventory shell zsh export --env "$PWD" --no-env-no-problem true)
}
`
	if printChpwdHook {
		fmt.Fprint(cmdCtx.Stdout, chpwdHook)
	}

	exportEnv := `
export-env() { eval $(enventory shell zsh export --env "$1" --no-env-no-problem true) }
unexport-env() { eval $(enventory shell zsh unexport --env "$1" --no-env-no-problem true) }
`
	if printExportEnv {
		fmt.Fprint(cmdCtx.Stdout, exportEnv)
	}

	return nil
}

func ShellZshExportCmd() wargcore.Command {
	return command.New(
		"Print export script",
		withEnvService(shellZshExportRun),
		command.Flag("--env", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.NewFlag(
			"--no-env-no-problem",
			"Exit without an error if the environment doesn't exit. Useful when runnng envelop on chpwd",
			scalar.Bool(
				scalar.Default(false),
			),
			flag.Required(),
		),
	)
}

func shellZshExportRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "export")
}

func ShellZshUnexportCmd() wargcore.Command {
	return command.New(
		"Print unexport script",
		withEnvService(shellZshUnexportRun),
		command.Flag("--env", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.NewFlag(
			"--no-env-no-problem",
			"Exit without an error if the environment doesn't exit. Useful when runnng envelop on chpwd",
			scalar.Bool(
				scalar.Default(false),
			),
			flag.Required(),
		),
	)
}

func shellZshUnexportRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "unexport")
}

func shellZshExportUnexport(ctx context.Context, cmdCtx wargcore.Context, es models.EnvService, scriptType string) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	noEnvNoProblem := cmdCtx.Flags["--no-env-no-problem"].(bool)

	envVars, err := es.VarList(ctx, envName)
	if err != nil {
		if errors.Is(err, models.ErrEnvNotFound) && noEnvNoProblem {
			return nil
		}
		return fmt.Errorf("could not list env vars: %s: %w", envName, err)
	}

	envRefs, envRefVars, err := es.VarRefList(ctx, envName)
	if err != nil {
		if errors.Is(err, models.ErrEnvNotFound) && noEnvNoProblem {
			return nil
		}
		return fmt.Errorf("could not list env refs: %s: %w", envName, err)
	}

	switch scriptType {
	case "export":
		if len(envVars)+len(envRefs) > 0 {
			fmt.Fprintf(cmdCtx.Stdout, "printf '%s:';\n", cmdCtx.App.Name)
			for _, ev := range envVars {
				fmt.Fprintf(cmdCtx.Stdout, "printf ' +%s';\n", shellescape.Quote(ev.Name))
				fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(ev.Name), shellescape.Quote(ev.Value))
			}

			for i := range len(envRefs) {
				fmt.Fprintf(cmdCtx.Stdout, "printf ' +%s';\n", shellescape.Quote(envRefs[i].Name))
				fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(envRefs[i].Name), shellescape.Quote(envRefVars[i].Value))
			}
			fmt.Fprintf(cmdCtx.Stdout, "echo;\n")
		}

	case "unexport":
		if len(envVars)+len(envRefs) > 0 {
			fmt.Fprintf(cmdCtx.Stdout, "printf '%s:';\n", cmdCtx.App.Name)
			for _, ev := range envVars {
				fmt.Fprintf(cmdCtx.Stdout, "printf ' -%s';\n", shellescape.Quote(ev.Name))
				fmt.Fprintf(cmdCtx.Stdout, "unset %s;\n", shellescape.Quote(ev.Name))
			}

			for _, er := range envRefs {
				fmt.Fprintf(cmdCtx.Stdout, "printf ' -%s';\n", shellescape.Quote(er.Name))
				fmt.Fprintf(cmdCtx.Stdout, "unset %s;\n", shellescape.Quote(er.Name))
			}
			fmt.Fprintf(cmdCtx.Stdout, "echo;\n")
		}
	default:
		return errors.New("unimplemented --script-type: " + scriptType)

	}

	return nil
}

func ShellZshChdirCmd() wargcore.Command {
	return command.New(
		"Change directory and corresponding env vars",
		withEnvService(shellZshChdirRun),
		// TODO: maybe define the flags here to get better help.
		command.Flag("--old", envNameFlag()),
		command.Flag("--new", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
		command.NewFlag(
			"--no-env-no-problem",
			"Exit without an error if the environment doesn't exit. Useful when runnng envelop on chpwd",
			scalar.Bool(
				scalar.Default(false),
			),
			flag.Required(),
		),
	)
}

func shellZshChdirRun(ctx context.Context, es models.EnvService, cmdCtx wargcore.Context) error {
	// TODO: implement
	return errors.New("TODO")
}
