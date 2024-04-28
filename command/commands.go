package command

import (
	"cas/localstorage"

	"github.com/mitchellh/cli"
)

func Commands(ui cli.Ui) map[string]cli.CommandFactory {

	storage := &localstorage.FileStore{}

	return map[string]cli.CommandFactory{
		"version":  NewCommand("version", NewVersionCommand()),
		"fetch":    NewCommand("fetch", NewFetchCommand(storage)),
		"artifact": NewCommand("artifact", NewArtifactCommand(storage)),

		"hash": func() (cli.Command, error) {
			return NewHashCommand(ui), nil
		},
	}
}
