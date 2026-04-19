package doc

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/loupedeck/pkg/scriptmeta"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var scriptPath string
	var format string
	cmd := &cobra.Command{
		Use:   "doc --script <path>",
		Short: "Extract jsdoc/jsdocex documentation from scene scripts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(scriptPath) == "" {
				return fmt.Errorf("missing --script")
			}
			_, store, err := scriptmeta.BuildDocStore(cmd.Context(), scriptPath)
			if err != nil {
				return err
			}
			output, err := scriptmeta.ExportDocStore(cmd.Context(), store, format)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(output)
			return err
		},
	}
	cmd.Flags().StringVar(&scriptPath, "script", "", "Path to the JavaScript file or scene directory")
	cmd.Flags().StringVar(&format, "format", "json", "Output format: json or markdown")
	return cmd
}
