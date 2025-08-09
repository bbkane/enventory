package main

import (
	_ "embed"

	"go.bbkane.com/enventory/cli"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/help"
	"go.bbkane.com/warg/section"
	wargcli "go.bbkane.com/warg/wargcore"
)

var version string

func buildApp() *wargcli.App {
	app := warg.New(
		"enventory",
		version,
		section.New(
			"Manage Environmental secrets centrally",
			section.NewSection(
				"completion",
				"Print completion scripts",
				section.Command("zsh", cli.CompletionZshCmd()),
			),
			section.NewSection(
				"env",
				"Environment commands",
				section.Command("create", cli.EnvCreateCmd()),
				section.Command("delete", cli.EnvDeleteCmd()),
				section.Command("list", cli.EnvListCmd()),
				section.Command("update", cli.EnvUpdateCmd()),
				section.Command("show", cli.EnvShowCmd()),
			),
			section.NewSection(
				"shell",
				"Manipulate the current shell",
				section.NewSection(
					"zsh",
					"Zsh-specific commands",
					section.Command("chdir", cli.ShellZshChdirCmd()),
					section.Command("init", cli.ShellZshInitCmd()),
					section.Command("export", cli.ShellZshExportCmd()),
					section.Command("unexport", cli.ShellZshUnexportCmd()),
				),
			),
			section.NewSection(
				"var",
				"Env vars owned by this environment",
				section.Command("create", cli.VarCreateCmd()),
				section.Command("delete", cli.VarDeleteCmd()),
				section.Command("show", cli.VarShowCmd()),
				section.Command("update", cli.VarUpdateCmd()),
				section.NewSection(
					"ref",
					"Variable References owned by this environment",
					section.Command("create", cli.VarRefCreateCmd()),
					section.Command("delete", cli.VarRefDeleteCmd()),
					section.Command("show", cli.VarRefShowCmd()),
					section.Command("update", cli.VarRefUpdateCmd()),
				),
			),
		),
		// use "detailed" as the default choice
		warg.HelpFlag(
			help.DefaultHelpCommandMap(),
			help.DefaultHelpFlagMap("detailed", help.DefaultHelpCommandMap().SortedNames()),
		),
		warg.SkipCompletionCommands(),
	)
	return &app
}

func main() {
	app := buildApp()
	app.MustRun()
}
