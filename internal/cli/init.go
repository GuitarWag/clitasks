package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
)

func newInitCmd() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new task board",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			in := board.MetaInput{}
			if cmd.Flags().Changed("name") {
				in.Name = &name
			}
			if cmd.Flags().Changed("description") {
				in.Description = &description
			}
			if err := b.UpdateMeta(in); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Board initialized at "+b.Path()))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "My Board", "Board name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Board description")
	return cmd
}
