package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
)

var flagFile string

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "tasks",
		Short:         "CLI task management with Markdown storage",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.PersistentFlags().StringVarP(&flagFile, "file", "f", "",
		"Path to the markdown file (default: $TASK_BOARD_FILE or tasks.md)")

	root.AddCommand(
		newInitCmd(), newAddCmd(), newListCmd(), newBoardCmd(), newShowCmd(),
		newUpdateCmd(), newMoveCmd(), newStartCmd(), newCompleteCmd(),
		newBlockCmd(), newDeleteCmd(), newInfoCmd(), newStatsCmd(),
		newExportCmd(), newTuiCmd(), newClaudeCmd(), newCodexCmd(),
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

func resolveFilePath() string {
	if flagFile != "" {
		return flagFile
	}
	if v := os.Getenv("TASK_BOARD_FILE"); v != "" {
		return v
	}
	return storage.DefaultFile
}

func openBoard() (*board.Board, error) {
	return board.Open(storage.NewMarkdown(resolveFilePath()))
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
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleDim     = lipgloss.NewStyle().Faint(true)
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleCyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	styleMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	styleRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleGray    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func priorityStyle(p model.TaskPriority) lipgloss.Style {
	switch p {
	case model.PriorityCritical:
		return styleRed
	case model.PriorityHigh:
		return styleYellow
	case model.PriorityMedium:
		return styleBlue
	default:
		return styleGray
	}
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
