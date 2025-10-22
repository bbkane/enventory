package main

import (
	_ "embed"

	"go.bbkane.com/enventory/cli"
	"go.bbkane.com/warg"
)

var version string

func buildApp() *warg.App {
	app := warg.New(
		"enventory",
		version,
		warg.NewSection(
			"Manage Environmental secrets centrally",
			warg.NewSubSection(
				"completion",
				"Print completion scripts",
				warg.SubCmd("zsh", cli.CompletionZshCmd()),
			),
			warg.NewSubSection(
				"env",
				"Environment commands",
				warg.SubCmd("create", cli.EnvCreateCmd()),
				warg.SubCmd("delete", cli.EnvDeleteCmd()),
				warg.SubCmd("list", cli.EnvListCmd()),
				warg.SubCmd("update", cli.EnvUpdateCmd()),
				warg.SubCmd("show", cli.EnvShowCmd()),
			),
			warg.NewSubSection(
				"shell",
				"Manipulate the current shell",
				warg.NewSubSection(
					"zsh",
					"Zsh-specific commands",
					warg.SubCmd("chdir", cli.ShellZshChdirCmd()),
					warg.SubCmd("init", cli.ShellZshInitCmd()),
					warg.SubCmd("export", cli.ShellZshExportCmd()),
					warg.SubCmd("unexport", cli.ShellZshUnexportCmd()),
				),
			),
			warg.NewSubSection(
				"var",
				"Env vars owned by this environment",
				warg.SubCmd("create", cli.VarCreateCmd()),
				warg.SubCmd("delete", cli.VarDeleteCmd()),
				warg.SubCmd("show", cli.VarShowCmd()),
				warg.SubCmd("update", cli.VarUpdateCmd()),
				warg.NewSubSection(
					"ref",
					"Variable References owned by this environment",
					warg.SubCmd("create", cli.VarRefCreateCmd()),
					warg.SubCmd("delete", cli.VarRefDeleteCmd()),
					warg.SubCmd("show", cli.VarRefShowCmd()),
					warg.SubCmd("update", cli.VarRefUpdateCmd()),
				),
			),
			warg.SubCmd("exec", cli.ExecCmd()),
		),
		warg.HelpFlag(
			warg.DefaultHelpCmdMap(),
			warg.DefaultHelpFlagMap("detailed", warg.DefaultHelpCmdMap().SortedNames()),
		),
		warg.SkipCompletionCmds(),
	)
	return &app
}

func main() {
	app := buildApp()
	app.MustRun()
}
