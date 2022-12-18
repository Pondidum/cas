package command

import (
	"github.com/mitchellh/cli"
)

func Commands(ui cli.Ui) map[string]cli.CommandFactory {

	return map[string]cli.CommandFactory{
		"version": func() (cli.Command, error) {
			cmd := &VersionCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},

		"fetch": func() (cli.Command, error) {
			cmd := &FetchCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},

		"artifact": func() (cli.Command, error) {
			return NewArtifactCommand(ui), nil
		},

		"hash": func() (cli.Command, error) {
			return NewHashCommand(ui), nil
		},
	}
}
