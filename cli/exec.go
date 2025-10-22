package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/completion"
	"go.bbkane.com/warg/path"
	"go.bbkane.com/warg/value/scalar"
	"go.bbkane.com/warg/value/slice"
	"gopkg.in/yaml.v3"
)

// Var represents one flag entry in the YAML file.
type Var struct {
	Name            string       `yaml:"name"`
	Help            string       `yaml:"help"`
	CompletionsType string       `yaml:"completions_type"`
	Completions     []Completion `yaml:"completions"`
}

// Group represents a group of environment variables.
type Group struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Vars        map[string]string `yaml:"vars"`
}

// Completion can represent either a simple string value or a name/description pair.
type Completion struct {
	Value string `yaml:"value,omitempty"`
	Help  string `yaml:"help,omitempty"`
}

// UnmarshalYAML handles both `values` (list of strings) and
// `values_descriptions` (list of maps) cases.
func (c *Completion) UnmarshalYAML(value *yaml.Node) error {
	// Case 1: plain string (e.g., "true")
	if value.Kind == yaml.ScalarNode {
		c.Value = value.Value
		return nil
	}

	// Case 2: mapping (e.g., { name: "1", description: "Greet once" })
	type completionAlias Completion
	var tmp completionAlias
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	*c = Completion(tmp)
	return nil
}

// Config is the root of the YAML document.
type Config struct {
	Groups []Group `yaml:"groups"`
	Vars   []Var   `yaml:"vars"`
}

type yamlToFlagsResult struct {
	flagMap      warg.FlagMap
	groupFlagMap warg.FlagMap
}

func yamlToFlags(path string) (*yamlToFlagsResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse ENVENTORY_EXEC_CONFIG file: %w", err)
	}

	flagMap := make(warg.FlagMap)
	for _, flag := range config.Vars {
		var completions warg.CompletionsFunc
		switch flag.CompletionsType {
		case "value":
			values := make([]string, 0, len(flag.Completions))
			for _, c := range flag.Completions {
				values = append(values, c.Value)
			}
			completions = warg.CompletionsValues(values)
		case "value_help":
			candidates := make([]completion.Candidate, 0, len(flag.Completions))
			for _, c := range flag.Completions {
				candidates = append(candidates, completion.Candidate{
					Name:        c.Value,
					Description: c.Help,
				})
			}
			completions = warg.CompletionsValuesDescriptions(candidates)
		case "":
			completions = nil
		default:
			return nil, fmt.Errorf("unknown completions_type: %s", flag.CompletionsType)
		}

		flagMap.AddFlag("--"+flag.Name, warg.NewFlag(
			flag.Help,
			scalar.String(),
			warg.FlagCompletions(completions),
		))
	}

	var groupFlagMap warg.FlagMap
	if len(config.Groups) > 0 {
		groups := []string{}
		for _, group := range config.Groups {
			groups = append(groups, group.Name)
		}
		groupFlagMap = warg.FlagMap{
			"--group": warg.NewFlag(
				"Groups defined in $ENVENTORY_EXEC_CONFIG",
				slice.String(
					slice.Choices(groups...),
				),
			),
		}
	}
	return &yamlToFlagsResult{
		flagMap:      flagMap,
		groupFlagMap: groupFlagMap,
	}, nil
}

const helpExec = `Exec a command with environment variable set from an existing env or flags from a YAML file

Example YAML file:

    # ~/my_yaml.yaml
    groups:
    - name: mygroup
      description: Default configuration
      vars:
        DEBUG: "true"
        LOG_LEVEL: "info"

    vars:
    - name: MOTEL_EXPORTER_OTLP_ENDPOINT
      help: 'OpenTelemetry Collector endpoint'
      completions_type: value
      completions:
      - 'localhost:4317'
      - 'otel-collector:4317'
    - name: COUNT
      help: 'Number of times to greet'
      completions_type: value_description
      completions:
      - value: '1'
        help: 'Greet once'
      - value: '5'
        help: 'Greet five times'

Then (assuming there's a pre-existing env named "myenv"):

    export ENVENTORY_EXEC_CONFIG=~/my_yaml.yaml
    enventory exec --env myenv --group mygroup --COUNT 1 -- my_command_to_run arg1 arg2

This does the following:

- updates the process's environment with the variables in "myenv"
- updates the process's environment with the variables in the "mygroup" group from the YAML file
- sets the COUNT environment variable to "1"
- execs "my_command_to_run" with the updated environment

Note that later steps in the process can override earlier steps; for example, if "myenv" defines COUNT=10, the flag --COUNT 1 will override it to COUNT=1.

The goal is to make it easy to run commands with environments while also allowing easy overrides from the YAML file, with tab completions everywhere so you don't have to remember names and values.`

func ExecCmd() warg.Cmd {
	flagMap := make(warg.FlagMap)
	groupFlagMap := make(warg.FlagMap)

	configPath := os.Getenv("ENVENTORY_EXEC_CONFIG")
	if configPath != "" {
		configPath = path.New(configPath).MustExpand()
		res, err := yamlToFlags(configPath)
		if err != nil {
			panic("could not parse ENVENTORY_EXEC_CONFIG: " + err.Error())
		}
		flagMap = res.flagMap
		groupFlagMap = res.groupFlagMap
	}

	return warg.NewCmd(
		helpExec,
		withSetup(execRun),
		warg.AllowForwardedArgs(),
		warg.CmdFlag(
			"--env",
			warg.NewFlag(
				"Environments to use. Env vars from later environments can override earlier ones",
				slice.String(),
				warg.FlagCompletions(withEnvServiceCompletions(
					completeExistingEnvName)),
			),
		),
		warg.CmdFlag("--inherit-env",
			warg.NewFlag(
				"Inherit the current process's environment variables. Make sure to set up PATH, HOME, etc. if you enable this.",
				scalar.Bool(
					scalar.Default(true),
				),
				warg.Required(),
			),
		),
		warg.CmdFlagMap(timeoutFlagMap()),
		warg.CmdFlagMap(sqliteDSNFlagMap()),
		warg.CmdFlagMap(flagMap),
		warg.CmdFlagMap(groupFlagMap),
	)
}

func execRun(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error {

	if !cmdCtx.Flags["--inherit-env"].(bool) {
		os.Clearenv()
	}

	var command string
	var args []string
	switch len(cmdCtx.ForwardedArgs) {
	case 0:
		return fmt.Errorf("no command provided to exec")
	case 1:
		command = cmdCtx.ForwardedArgs[0]
	default:
		command = cmdCtx.ForwardedArgs[0]
		args = cmdCtx.ForwardedArgs[1:]
	}

	// set env vars from specified envs
	envs := []string{}
	envsIFace, exists := cmdCtx.Flags["--env"]
	if exists {
		envs = envsIFace.([]string)
	}

	for _, envName := range envs {
		envVars, err := es.VarList(ctx, envName)
		if err != nil {

			return fmt.Errorf("could not list env vars: %s: %w", envName, err)
		}
		for _, ev := range envVars {
			os.Setenv(ev.Name, ev.Value)
		}

		envRefs, envRefVars, err := es.VarRefList(ctx, envName)
		if err != nil {
			return fmt.Errorf("could not list env refs: %s: %w", envName, err)
		}
		for i := range envRefs {
			os.Setenv(envRefs[i].Name, envRefVars[i].Value)
		}
	}

	// set groups
	// I don't love reading the YAML twice, but this will do for now
	configPath := os.Getenv("ENVENTORY_EXEC_CONFIG")
	if configPath != "" {
		configPath = path.New(configPath).MustExpand()
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("could not read YAML file: %w", err)
		}
		var config Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("could not parse ENVENTORY_EXEC_CONFIG file: %w", err)
		}

		groups := []string{}
		groupsIFace, exists := cmdCtx.Flags["--group"]
		if exists {
			groups = groupsIFace.([]string)
		}
		for _, groupName := range groups {
			for _, group := range config.Groups {
				if group.Name == groupName {
					for varName, varValue := range group.Vars {
						os.Setenv(varName, varValue)
					}
				}
			}
		}
	}

	// set custom flags as env vars
	cmdFlagNames := warg.NewSet[string]()
	cmdFlagNames.Add("--color")
	cmdFlagNames.Add("--db-path")
	cmdFlagNames.Add("--env")
	cmdFlagNames.Add("--group")
	cmdFlagNames.Add("--help")
	cmdFlagNames.Add("--inherit-env")
	cmdFlagNames.Add("--timeout")

	for flagName, flagValue := range cmdCtx.Flags {
		if cmdFlagNames.Contains(flagName) {
			continue
		}
		os.Setenv(flagName[2:], flagValue.(string))
	}

	cmd := exec.Command(command, args...)
	cmd.Stderr = cmdCtx.Stderr
	cmd.Stdout = cmdCtx.Stdout
	cmd.Stdin = os.Stdin // TODO: update this if I add a cmdCtx.Stdin to warg
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
