package common

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

func BuildRuntimeCobraCommand(command cmds.Command) (*cobra.Command, error) {
	return cli.BuildCobraCommandFromCommand(command,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsSection: true,
		}),
	)
}
