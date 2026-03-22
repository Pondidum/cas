package command

import (
	"cas/localstorage"

	"github.com/hashicorp/cli"
)

func Commands(ui cli.Ui) map[string]cli.CommandFactory {

	storage := localstorage.NewArchiveDecorator(&localstorage.FileStore{})

	return map[string]cli.CommandFactory{
		"version":       NewCommand("version", NewVersionCommand()),
		"fetch":         NewCommand("fetch", NewFetchCommand(storage)),
		"artifact list": NewCommand("artifact list", NewArtifactListCommand(storage)),
		"artifact push": NewCommand("artifact push", NewArtifactPushCommand(storage)),
		"artifact pull": NewCommand("artifact pull", NewArtifactPullCommand(storage)),
		"hash":          NewCommand("hash", NewHashCommand()),
	}
}
