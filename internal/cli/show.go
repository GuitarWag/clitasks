package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/v3/internal/ui/card"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			t, ok := b.Get(args[0])
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+args[0]))
				return nil
			}
			if o := outputOf(cmd); o.rich {
				var rel card.Related
				for _, id := range t.After {
					if a, ok := b.Get(id); ok {
						rel.After = append(rel.After, a)
					}
				}
				for _, id := range b.Dependents(t.ID) {
					if d, ok := b.Get(id); ok {
						rel.Dependents = append(rel.Dependents, d)
					}
				}
				for _, l := range card.Detail(t, rel, o.ctx(), min(o.width, 72)) {
					fmt.Fprintln(cmd.OutOrStdout(), l)
				}
				return nil
			}
			renderTask(cmd.OutOrStdout(), t, true)
			return nil
		},
	}
}
