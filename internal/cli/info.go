package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show board information",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard()
			if err != nil {
				return err
			}
			info := b.Info()
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, styleBold.Render(styleCyan.Render("\n"+info.Name)))
			if info.Description != "" {
				fmt.Fprintln(out, styleDim.Render(info.Description))
			}
			fmt.Fprintf(out, "\nFile: %s\n", styleYellow.Render(b.Path()))
			fmt.Fprintf(out, "Total tasks: %s\n", styleCyan.Render(fmt.Sprintf("%d", len(info.Tasks))))
			fmt.Fprintf(out, "Created: %s\n", styleDim.Render(info.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))
			fmt.Fprintf(out, "Updated: %s\n\n", styleDim.Render(info.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))
			return nil
		},
	}
}
