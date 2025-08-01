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
		// TODO: actuall use this flag!
		command.NewFlag(
			"--chpwd-strategy",
			"Temporary flag to revert back to the old chpwd strategy if bugs are found with the new one",
			scalar.String(
				scalar.Choices("v0.0.19", "v0.0.20"),
				scalar.Default("v0.0.20"),
			),
			flag.Required(),
		),
	)
}

func shellZshInitRun(cmdCtx wargcore.Context) error {

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

func ShellZshExportCmd() wargcore.Command {
	return command.New(
		"Print export script",
		withSetup(shellZshExportRun),
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

func shellZshExportRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "export")
}

func ShellZshUnexportCmd() wargcore.Command {
	return command.New(
		"Print unexport script",
		withSetup(shellZshUnexportRun),
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

func shellZshUnexportRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	return shellZshExportUnexport(ctx, cmdCtx, es, "unexport")
}

func shellZshExportUnexport(ctx context.Context, cmdCtx wargcore.Context, es models.Service, scriptType string) error {
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
		withSetup(shellZshChdirRun),
		// TODO: maybe define the flags here to get better descriptions.
		command.Flag("--old", envNameFlag()),
		command.Flag("--new", envNameFlag()),
		command.FlagMap(timeoutFlagMap()),
		command.FlagMap(sqliteDSNFlagMap()),
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

func shellZshChdirRun(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error {
	oldEnvName := cmdCtx.Flags["--old"].(string)
	newEnvName := cmdCtx.Flags["--new"].(string)

	lookupEnv := os.LookupEnv
	if custom := cmdCtx.Context.Value(CustomLookupEnvFuncKey{}); custom != nil {
		lookupEnv = custom.(LookupEnvFunc)
	}

	// TODO: once I update vw_env_var_env_ref_unique_name to have the value, I can just query from there instead of two separate queries.
	// TODO: if the environment isn't found, we shouldn't query for refs in the same env
	oldVars, err := es.VarList(ctx, oldEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list old env vars: %s: %w", oldEnvName, err)
	}
	newVars, err := es.VarList(ctx, newEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list new env vars: %s: %w", newEnvName, err)
	}

	oldRefs, oldRefVars, err := es.VarRefList(ctx, oldEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list old env refs: %s: %w", oldEnvName, err)
	}
	newRefs, newRefVars, err := es.VarRefList(ctx, newEnvName)
	if err != nil && !errors.Is(err, models.ErrEnvNotFound) {
		return fmt.Errorf("could not list new env refs: %s: %w", newEnvName, err)
	}

	// turn them into maps for ease of checking...
	newKVs := make(map[string]string, len(newVars)+len(newRefs))
	for _, v := range newVars {
		newKVs[v.Name] = v.Value
	}
	for i := range newRefs {
		newKVs[newRefs[i].Name] = newRefVars[i].Value
	}
	oldKVs := make(map[string]string, len(oldVars)+len(oldRefs))
	for _, v := range oldVars {
		// if it exists in the new env, we don't need to process in the old env
		if _, exists := newKVs[v.Name]; exists {
			continue
		}
		oldKVs[v.Name] = v.Value
	}
	for i := range oldRefs {
		// if it exists in the new env, we don't need to process in the old env
		if _, exists := newKVs[oldRefs[i].Name]; exists {
			continue
		}
		oldKVs[oldRefs[i].Name] = oldRefVars[i].Value
	}

	todo := computeExportChanges(oldKVs, newKVs, lookupEnv)

	if len(todo.ToAdd)+len(todo.ToChange)+len(todo.ToRemove)+len(todo.Unchanged) == 0 {
		return nil
	}

	fmt.Fprintf(cmdCtx.Stdout, "printf '%s:';\n", cmdCtx.App.Name)

	// print the change script
	for _, kv := range todo.ToRemove {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' -%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "unset %s;\n", shellescape.Quote(kv.Name))
	}
	for _, kv := range todo.Unchanged {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' =%s';\n", shellescape.Quote(kv.Name))
	}
	for _, kv := range todo.ToChange {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' ~%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(kv.Name), shellescape.Quote(kv.Value))
	}
	for _, kv := range todo.ToAdd {
		fmt.Fprintf(cmdCtx.Stdout, "printf ' +%s';\n", shellescape.Quote(kv.Name))
		fmt.Fprintf(cmdCtx.Stdout, "export %s=%s;\n", shellescape.Quote(kv.Name), shellescape.Quote(kv.Value))
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
