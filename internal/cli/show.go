package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard()
			if err != nil {
				return err
			}
			t, ok := b.Get(args[0])
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+args[0]))
				return nil
			}
			renderTask(cmd.OutOrStdout(), t, true)
			return nil
		},
	}
}
