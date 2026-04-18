package main

import (
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	helpcmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	loupedeckcmdcommon "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/common"
	doccmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/doc"
	runcmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/run"
	verbscmd "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/verbs"
	doc "github.com/go-go-golems/loupedeck/docs/help"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	bootstrap, err := verbscmd.DiscoverBootstrap(os.Args[1:])
	cobra.CheckErr(err)

	rootCmd := &cobra.Command{
		Use:     "loupedeck",
		Short:   "Run Loupedeck Live JavaScript scenes and hardware workflows",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	cobra.CheckErr(logging.AddLoggingSectionToRootCommand(rootCmd, "loupedeck"))

	helpSystem := help.NewHelpSystem()
	cobra.CheckErr(doc.AddDocToHelpSystem(helpSystem))
	helpcmd.SetupCobraRootCommand(helpSystem, rootCmd)

	runCommand, err := runcmd.NewCommand()
	cobra.CheckErr(err)
	runCobraCmd, err := loupedeckcmdcommon.BuildRuntimeCobraCommand(runCommand)
	cobra.CheckErr(err)
	rootCmd.PersistentFlags().StringArray(verbscmd.VerbRepositoryFlag, nil, "Additional JavaScript verb repositories to scan before command registration (repeatable)")
	rootCmd.AddCommand(runCobraCmd)
	verbsCommand, err := verbscmd.NewCommand(bootstrap)
	cobra.CheckErr(err)
	rootCmd.AddCommand(verbsCommand)
	rootCmd.AddCommand(doccmd.NewCommand())

	cobra.CheckErr(rootCmd.Execute())
}
