package cli

import (
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/tui"
)

func newTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive Terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(resolveFilePath(cmd))
		},
	}
}
