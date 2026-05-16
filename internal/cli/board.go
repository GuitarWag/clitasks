package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/model"
)

func newBoardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "board",
		Short: "Display kanban board view",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard()
			if err != nil {
				return err
			}
			info := b.Info()
			by := b.ByStatus()
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, styleBold.Render(styleCyan.Render("\n# "+info.Name)))
			if info.Description != "" {
				fmt.Fprintln(out, styleDim.Render(info.Description))
			}
			fmt.Fprintln(out, styleDim.Render("Last updated: "+info.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")))

			order := []struct {
				status model.TaskStatus
				label  string
			}{
				{model.StatusTodo, "TODO"},
				{model.StatusInProgress, "IN PROGRESS"},
				{model.StatusBlocked, "BLOCKED"},
				{model.StatusDone, "DONE"},
			}
			for _, col := range order {
				renderColumn(out, col.label, by[col.status])
			}
			fmt.Fprintln(out)
			return nil
		},
	}
}

func renderColumn(w io.Writer, label string, tasks []model.Task) {
	fmt.Fprintln(w, styleBold.Render(fmt.Sprintf("\n## %s (%d)", label, len(tasks))))
	if len(tasks) == 0 {
		fmt.Fprintln(w, styleDim.Render("  No tasks"))
		return
	}
	for _, t := range tasks {
		prio := priorityStyle(t.Priority).Render(string(t.Priority))
		assignee := ""
		if t.Assignee != "" {
			assignee = styleGreen.Render("@" + t.Assignee)
		}
		var tagged []string
		for _, tg := range t.Tags {
			tagged = append(tagged, styleMagenta.Render("#"+tg))
		}
		tagsStr := strings.Join(tagged, " ")
		fmt.Fprintf(w, "  %s %s %s %s %s\n",
			styleCyan.Render(t.ID), t.Title, prio, assignee, tagsStr)
		if t.Description != "" {
			fmt.Fprintln(w, styleDim.Render("    "+t.Description))
		}
	}
}
