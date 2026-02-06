package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg"

	"go.bbkane.com/warg/value/scalar"
)

func ShellZshInitCmd() warg.Cmd {
	return warg.NewCmd(
		"Prints the zsh initialization script",
		shellZshInitRun,
		warg.NewCmdFlag(
			"--print-autoload",
			"Include autoload -Uz add-zsh-hook line (might not be needed if you already autoloaded it)",
			scalar.Bool(
				scalar.Default(true),
			),
			warg.Required(),
		),
		// TODO: actuall use this flag!
		warg.NewCmdFlag(
			"--chpwd-strategy",
			"Temporary flag to revert back to the old chpwd strategy if bugs are found with the new one",
			scalar.String(
				scalar.Choices("v0.0.19", "v0.0.20"),
				scalar.Default("v0.0.20"),
			),
			warg.Required(),
		),
	)
}

func shellZshInitRun(cmdCtx warg.CmdContext) error {

	printAutoload := cmdCtx.Flags["--print-autoload"].(bool)
	chpwdStrategy := cmdCtx.Flags["--chpwd-strategy"].(string)

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

	var chpwdHook string
	switch chpwdStrategy {
	case "v0.0.19":
		chpwdHook = `
add-zsh-hook -Uz chpwd (){
    eval $(enventory shell zsh unexport --env "$OLDPWD" --no-env-no-problem true)
    eval $(enventory shell zsh export --env "$PWD" --no-env-no-problem true)
}
`
	case "v0.0.20":
		chpwdHook = `
add-zsh-hook -Uz chpwd (){
    eval $(enventory shell zsh chdir --old "$OLDPWD" --new "$PWD")
}
`
	}

	fmt.Fprint(cmdCtx.Stdout, chpwdHook)

	exportEnv := `
export-env() { eval $(enventory shell zsh export --env "$1" --no-env-no-problem true) }
unexport-env() { eval $(enventory shell zsh unexport --env "$1" --no-env-no-problem true) }
`
	fmt.Fprint(cmdCtx.Stdout, exportEnv)

	return nil
}

func ShellZshExportCmd() warg.Cmd {
	return warg.NewCmd(
		"Print export script",
		withSetup(shellZshExportRun),
		warg.CmdFlag("--env", envNameFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.NewCmdFlag(
			"--no-env-no-problem",
			"Exit without an error if the environment doesn't exit. Useful when runnng envelop on chpwd",
			scalar.Bool(
				scalar.Default(false),
			),
			warg.Required(),
		),
	)
}

func shellZshExportRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "export")
}

func ShellZshUnexportCmd() warg.Cmd {
	return warg.NewCmd(
		"Print unexport script",
		withSetup(shellZshUnexportRun),
		warg.CmdFlag("--env", envNameFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.NewCmdFlag(
			"--no-env-no-problem",
			"Exit without an error if the environment doesn't exit. Useful when runnng envelop on chpwd",
			scalar.Bool(
				scalar.Default(false),
			),
			warg.Required(),
		),
	)
}

func shellZshUnexportRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "unexport")
}

func shellZshExportUnexport(ctx context.Context, cmdCtx warg.CmdContext, es models.Service, scriptType string) error {
	envName := mustGetEnvNameArg(cmdCtx.Flags)
	noEnvNoProblem := cmdCtx.Flags["--no-env-no-problem"].(bool)

	exportables, err := es.EnvExportableList(ctx, envName)
	if err != nil {
		if errors.Is(err, models.ErrEnvNotFound) && noEnvNoProblem {
			return nil
		}
		return fmt.Errorf("could not list exportable env vars: %s: %w", envName, err)
	}

	if len(exportables) == 0 {
		return nil
	}

	kvs := make([]kv, 0, len(exportables))
	for _, e := range exportables {
		if e.Enabled {
			kvs = append(kvs, kv{
				Name:  e.Name,
				Value: e.Value,
			})
		}
	}
	if len(kvs) == 0 {
		return nil
	}
	fmt.Fprintf(cmdCtx.Stdout, "printf '%s:';\n", cmdCtx.App.Name)

	for _, e := range kvs {
		switch scriptType {
		case "export":
			fmt.Fprintf(cmdCtx.Stdout, "printf ' +%s';\n", shellescape.Quote(e.Name))
			fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(e.Name), shellescape.Quote(e.Value))
		case "unexport":
			fmt.Fprintf(cmdCtx.Stdout, "printf ' -%s';\n", shellescape.Quote(e.Name))
			fmt.Fprintf(cmdCtx.Stdout, "unset %s;\n", shellescape.Quote(e.Name))
		default:
			return errors.New("unimplemented --script-type: " + scriptType)
		}
	}
	fmt.Fprintf(cmdCtx.Stdout, "echo;\n")
	return nil
}

func ShellZshChdirCmd() warg.Cmd {
	return warg.NewCmd(
		"Change directory and corresponding env vars",
		withSetup(shellZshChdirRun),
		// TODO: maybe define the flags here to get better descriptions.
		warg.CmdFlag("--old", envNameFlag()),
		warg.CmdFlag("--new", envNameFlag()),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
	)
}

type LookupEnvFunc = func(key string) (string, bool)
type CustomLookupEnvFuncKey struct{}

// LookupMap loooks up keys from a provided map. Useful to mock os.LookupEnv when parsing
func LookupMap(m map[string]string) LookupEnvFunc {
	return func(key string) (string, bool) {
		val, exists := m[key]
		return val, exists
	}
}

func shellZshChdirRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {
	oldEnvName := cmdCtx.Flags["--old"].(string)
	newEnvName := cmdCtx.Flags["--new"].(string)

	lookupEnv := os.LookupEnv
	if custom, exists := cmdCtx.ParseMetadata.Get(CustomLookupEnvFuncKey{}); exists {
		lookupEnv = custom.(LookupEnvFunc)
	}

	newExportables, err := es.EnvExportableList(ctx, newEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list new env exportables: %s: %w", newEnvName, err)
	}
	oldExportables, err := es.EnvExportableList(ctx, oldEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list old env exportables: %s: %w", oldEnvName, err)
	}

	// TODO: figure out how envs and exportables being disabled should be handled here
	newKVs := make(map[string]string, len(newExportables))
	oldKVs := make(map[string]string, len(oldExportables))
	for _, ev := range newExportables {
		if ev.Enabled {
			newKVs[ev.Name] = ev.Value
		}
	}
	for _, ev := range oldExportables {
		// if it exists in the new env, we don't need to process in the old env
		// Let's also not consider enabled here, as these are slated to be removed anyway. So we want to unset them even if they are disabled in the old env, as long as they don't exist in the new env.
		if _, exists := newKVs[ev.Name]; exists {
			continue
		}
		oldKVs[ev.Name] = ev.Value
	}

	todo := computeExportChanges(oldKVs, newKVs, lookupEnv)

	if len(todo.ToAdd)+len(todo.ToChange)+len(todo.ToRemove)+len(todo.Unchanged) == 0 {
		return nil
	}

	fmt.Fprintf(cmdCtx.Stdout, "printf '%s:';\n", cmdCtx.App.Name)

	// print the change script
	for _, kv := range todo.ToAdd {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' +%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(kv.Name), shellescape.Quote(kv.Value))
	}
	for _, kv := range todo.ToChange {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' ~%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(kv.Name), shellescape.Quote(kv.Value))
	}
	for _, kv := range todo.ToRemove {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' -%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "unset %s;\n", shellescape.Quote(kv.Name))
	}
	for _, kv := range todo.Unchanged {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' =%s';\n", shellescape.Quote(kv.Name))
	}

	fmt.Fprint(cmdCtx.Stdout, "echo;\n")
	return nil
}

type kv struct {
	Name  string
	Value string
}

type computeExportChangesResult struct {
	ToAdd     []kv
	ToChange  []kv
	ToRemove  []kv
	Unchanged []kv
}

func computeExportChanges(oldKVs, newKVs map[string]string, lookupFunc func(string) (string, bool)) computeExportChangesResult {
	res := computeExportChangesResult{
		ToAdd:     nil,
		ToChange:  nil,
		ToRemove:  nil,
		Unchanged: nil,
	}

	for key, val := range oldKVs {
		_, exists := lookupFunc(key)
		if exists {
			res.ToRemove = append(res.ToRemove, kv{Name: key, Value: val})
		}
	}

	for key, val := range newKVs {
		envVal, exists := lookupFunc(key)
		if exists {
			if envVal == val {
				res.Unchanged = append(res.Unchanged, kv{Name: key, Value: val})
			} else {
				res.ToChange = append(res.ToChange, kv{Name: key, Value: val})
			}
		} else {
			res.ToAdd = append(res.ToAdd, kv{Name: key, Value: val})
		}
	}
	cmp := func(a, b kv) int {
		return strings.Compare(a.Name, b.Name)
	}
	slices.SortFunc(res.ToAdd, cmp)
	slices.SortFunc(res.ToChange, cmp)
	slices.SortFunc(res.ToRemove, cmp)
	slices.SortFunc(res.Unchanged, cmp)
	return res
}
