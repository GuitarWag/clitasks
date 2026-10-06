package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/v3/internal/board"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			dependents := b.Dependents(args[0])
			if _, err := b.Delete(args[0]); err != nil {
				if errors.Is(err, board.ErrNotFound) {
					fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+args[0]))
					return nil
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Task deleted: "+args[0]))
			if len(dependents) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), styleDim.Render("  Removed from after: "+strings.Join(dependents, ", ")))
			}
			return nil
		},
	}
}
