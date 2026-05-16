package cli

import (
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/model"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <id>",
		Short: "Start working on a task (move to in-progress)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, args[0], model.StatusInProgress, "✓ Started task: ",
				func(s string) string { return styleSuccess.Render(s) })
		},
	}
}

func newCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark task as complete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, args[0], model.StatusDone, "✓ Task completed: ",
				func(s string) string { return styleSuccess.Render(s) })
		},
	}
}

func newBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "block <id>",
		Short: "Mark task as blocked",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(cmd, args[0], model.StatusBlocked, "! Task blocked: ",
				func(s string) string { return styleWarn.Render(s) })
		},
	}
}
