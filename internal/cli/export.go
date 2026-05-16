package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/export"
)

func newExportCmd() *cobra.Command {
	var format, output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export board data",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard()
			if err != nil {
				return err
			}
			data, err := export.Render(b.Info(), format)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ "+err.Error()))
				return nil
			}
			if output != "" {
				if err := os.WriteFile(output, data, 0o644); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Exported to "+output))
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Export format (json|csv|summary)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (defaults to stdout)")
	return cmd
}
