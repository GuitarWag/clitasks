package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
	"github.com/GuitarWag/clitasks/internal/theme"
	"github.com/GuitarWag/clitasks/internal/ui"
	uitheme "github.com/GuitarWag/clitasks/internal/ui/theme"
)

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "tasks",
		Short:         "CLI task management with Markdown storage",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		// Lip Gloss v2 always emits full color. The writer downsamples it to
		// what the output supports and strips it for pipes and NO_COLOR.
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			r := cmd.Root()
			r.SetOut(colorprofile.NewWriter(r.OutOrStdout(), os.Environ()))
		},
	}
	root.PersistentFlags().StringP("file", "f", "",
		"Path to the markdown file (default: $TASK_BOARD_FILE or tasks.md)")
	root.PersistentFlags().String("theme", "",
		"Color theme: "+strings.Join(uitheme.Names(), "|")+" (default: $TASKS_THEME or auto)")
	root.PersistentFlags().String("icons", "",
		"Icon set: auto|nerd|unicode|ascii (default: $TASKS_ICONS or auto)")

	root.AddCommand(
		newInitCmd(), newAddCmd(), newListCmd(), newBoardCmd(), newShowCmd(),
		newUpdateCmd(), newMoveCmd(), newStartCmd(), newCompleteCmd(),
		newBlockCmd(), newDeleteCmd(), newInfoCmd(), newStatsCmd(),
		newExportCmd(), newTimelineCmd(), newTuiCmd(), newClaudeCmd(), newCodexCmd(),
	)
	return root
}

func Execute(version string) int {
	root := newRootCmd(version)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func resolveFilePath(cmd *cobra.Command) string {
	if cmd != nil {
		if v, _ := cmd.Flags().GetString("file"); v != "" {
			return v
		}
	}
	if v := os.Getenv("TASK_BOARD_FILE"); v != "" {
		return v
	}
	return storage.DefaultFile
}

// lookOptions merges the --theme and --icons flags over the env vars.
func lookOptions(cmd *cobra.Command) (ui.Options, error) {
	o := ui.OptionsFromEnv(os.Getenv)
	if v, _ := cmd.Flags().GetString("theme"); v != "" {
		o.Theme = v
	}
	if v, _ := cmd.Flags().GetString("icons"); v != "" {
		o.Icons = v
	}
	return o, o.Validate()
}

func openBoard(cmd *cobra.Command) (*board.Board, error) {
	return board.Open(storage.NewMarkdown(resolveFilePath(cmd)))
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

var (
	styleSuccess = theme.Success
	styleError   = theme.Error
	styleWarn    = theme.Warn
	styleDim     = theme.Dim
	styleBold    = theme.Bold
	styleCyan    = theme.Cyan
	styleGreen   = theme.Green
	styleYellow  = theme.Yellow
	styleBlue    = theme.Blue
	styleMagenta = theme.Magenta
	styleRed     = theme.Red
	styleGray    = theme.Gray
)

func priorityStyle(p model.TaskPriority) lipgloss.Style {
	return theme.PriorityStyle(p)
}

func statusGlyph(s model.TaskStatus) string {
	switch s {
	case model.StatusTodo:
		return "☐"
	case model.StatusInProgress:
		return "▶"
	case model.StatusDone:
		return "✓"
	case model.StatusBlocked:
		return "!"
	}
	return "?"
}

func renderTask(w io.Writer, t model.Task, detailed bool) {
	fmt.Fprintf(w, "\n%s %s\n", styleCyan.Render(t.ID), styleBold.Render(t.Title))
	fmt.Fprintf(w, "  Status: %s %s | Priority: %s\n",
		statusGlyph(t.Status), t.Status, priorityStyle(t.Priority).Render(string(t.Priority)))
	if t.Assignee != "" {
		fmt.Fprintf(w, "  Assignee: %s\n", styleGreen.Render(t.Assignee))
	}
	if len(t.Tags) > 0 {
		var tagged []string
		for _, tag := range t.Tags {
			tagged = append(tagged, styleMagenta.Render("#"+tag))
		}
		fmt.Fprintf(w, "  Tags: %s\n", strings.Join(tagged, " "))
	}
	if t.DueDate != "" {
		fmt.Fprintf(w, "  Due: %s\n", styleYellow.Render(t.DueDate))
	}
	if t.Start != "" {
		fmt.Fprintf(w, "  Start: %s\n", styleYellow.Render(t.Start))
	}
	if len(t.After) > 0 {
		fmt.Fprintf(w, "  After: %s\n", styleCyan.Render(strings.Join(t.After, ", ")))
	}
	if detailed && t.Description != "" {
		fmt.Fprintf(w, "  %s\n", styleDim.Render(t.Description))
	}
	if detailed {
		fmt.Fprintf(w, "  %s\n",
			styleDim.Render(fmt.Sprintf("Created: %s | Updated: %s",
				t.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
				t.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"))))
	}
}
