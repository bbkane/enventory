package cli

import (
	_ "embed"

	"go.bbkane.com/warg/command"
	"go.bbkane.com/warg/wargcore"
)

//go:embed completion_script.zsh
var zshCompletionScript []byte

func CompletionZshCmd() wargcore.Command {
	return command.New(
		"Print zsh completion script",
		func(ctx wargcore.Context) error {
			_, err := ctx.Stdout.Write(zshCompletionScript)
			return err
		},
	)
}
