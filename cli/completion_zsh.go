package cli

import (
	_ "embed"

	"go.bbkane.com/warg"
)

//go:embed completion_script.zsh
var zshCompletionScript []byte

func CompletionZshCmd() warg.Cmd {
	return warg.NewCmd(
		"Print zsh completion script",
		func(ctx warg.CmdContext) error {
			_, err := ctx.Stdout.Write(zshCompletionScript)
			return err
		},
	)
}
