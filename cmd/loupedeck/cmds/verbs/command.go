package verbs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	loupedeckcmdcommon "github.com/go-go-golems/loupedeck/cmd/loupedeck/cmds/common"
	"github.com/go-go-golems/loupedeck/pkg/scriptmeta"
	"github.com/spf13/cobra"
)

type helpOnlyCommand struct {
	*cmds.CommandDescription
}

func (c *helpOnlyCommand) Run(context.Context, *values.Values) error {
	return nil
}

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "verbs",
		Short: "Inspect jsverbs metadata for loupedeck scene scripts",
	}

	root.AddCommand(newListCommand())
	root.AddCommand(newHelpCommand())
	return root
}

func newListCommand() *cobra.Command {
	var scriptPath string
	cmd := &cobra.Command{
		Use:   "list --script <path>",
		Short: "List explicit jsverbs discovered for a script or scene directory",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(scriptPath) == "" {
				return fmt.Errorf("missing --script")
			}
			target, registry, err := scriptmeta.ScanVerbRegistry(scriptPath)
			if err != nil {
				return err
			}
			verbs := scriptmeta.EntryVerbs(target, registry)
			if len(verbs) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No explicit jsverbs found.")
				return nil
			}
			paths := make([]string, 0, len(verbs))
			for _, verb := range verbs {
				paths = append(paths, verb.FullPath())
			}
			sort.Strings(paths)
			for _, path := range paths {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&scriptPath, "script", "", "Path to the JavaScript file or scene directory")
	return cmd
}

func newHelpCommand() *cobra.Command {
	var scriptPath string
	var verbName string
	cmd := &cobra.Command{
		Use:   "help --script <path> --verb <name>",
		Short: "Render generated help for one jsverbs verb using its Glazed schema",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(scriptPath) == "" {
				return fmt.Errorf("missing --script")
			}
			if strings.TrimSpace(verbName) == "" {
				return fmt.Errorf("missing --verb")
			}
			target, registry, err := scriptmeta.ScanVerbRegistry(scriptPath)
			if err != nil {
				return err
			}
			verb, err := scriptmeta.FindVerb(target, registry, verbName)
			if err != nil {
				return err
			}
			desc, err := registry.CommandDescriptionForVerb(verb)
			if err != nil {
				return err
			}
			helpCmd, err := loupedeckcmdcommon.BuildCobraCommandDualMode(&helpOnlyCommand{CommandDescription: desc})
			if err != nil {
				return err
			}
			helpCmd.SetOut(cmd.OutOrStdout())
			helpCmd.SetErr(cmd.ErrOrStderr())
			helpCmd.SetArgs([]string{"--help"})
			return helpCmd.ExecuteContext(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&scriptPath, "script", "", "Path to the JavaScript file or scene directory")
	cmd.Flags().StringVar(&verbName, "verb", "", "Verb full path or unique short name")
	return cmd
}
