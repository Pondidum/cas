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

		"write": func() (cli.Command, error) {
			cmd := &WriteCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},

		"read": func() (cli.Command, error) {
			cmd := &ReadCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},

		"store": func() (cli.Command, error) {
			cmd := &StoreCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},

		"fetch": func() (cli.Command, error) {
			cmd := &FetchCommand{}
			cmd.Meta = NewMeta(ui, cmd)

			return cmd, nil
		},
	}
}
