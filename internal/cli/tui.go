package cli

import (
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/v3/internal/tui"
)

func newTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive Terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := lookOptions(cmd)
			if err != nil {
				return err
			}
			return tui.Run(resolveFilePath(cmd), o)
		},
	}
}
