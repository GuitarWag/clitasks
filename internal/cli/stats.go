package cli

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/model"
	uitheme "github.com/GuitarWag/clitasks/internal/ui/theme"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show board statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			info := b.Info()
			by := b.ByStatus()
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, styleBold.Render(styleCyan.Render("\n"+info.Name+" - Statistics")))
			fmt.Fprintln(out)

			fmt.Fprintln(out, styleBold.Render("Status Breakdown:"))
			o := outputOf(cmd)
			total := len(info.Tasks)
			row := func(label string, n int, st model.TaskStatus) {
				c := uitheme.Fg(o.look.Theme.Status(st))
				fmt.Fprintf(out, "  %-12s %s", label+":", c.Render(fmt.Sprintf("%-3d", n)))
				if o.rich {
					fmt.Fprint(out, " "+bar(n, total, 30, c))
				}
				fmt.Fprintln(out)
			}
			row("TODO", len(by[model.StatusTodo]), model.StatusTodo)
			row("IN PROGRESS", len(by[model.StatusInProgress]), model.StatusInProgress)
			row("DONE", len(by[model.StatusDone]), model.StatusDone)
			row("BLOCKED", len(by[model.StatusBlocked]), model.StatusBlocked)
			fmt.Fprintln(out, "  "+styleDim.Render("──────────────"))
			fmt.Fprintf(out, "  Total:       %s\n\n", styleCyan.Render(fmt.Sprintf("%d", len(info.Tasks))))

			counts := map[model.TaskPriority]int{}
			for _, t := range info.Tasks {
				counts[t.Priority]++
			}
			fmt.Fprintln(out, styleBold.Render("Priority Breakdown:"))
			fmt.Fprintf(out, "  Critical: %s\n", styleRed.Render(fmt.Sprintf("%d", counts[model.PriorityCritical])))
			fmt.Fprintf(out, "  High:     %s\n", styleYellow.Render(fmt.Sprintf("%d", counts[model.PriorityHigh])))
			fmt.Fprintf(out, "  Medium:   %s\n", styleBlue.Render(fmt.Sprintf("%d", counts[model.PriorityMedium])))
			fmt.Fprintf(out, "  Low:      %s\n\n", styleGray.Render(fmt.Sprintf("%d", counts[model.PriorityLow])))

			assignees := map[string]int{}
			for _, t := range info.Tasks {
				if t.Assignee != "" {
					assignees[t.Assignee]++
				}
			}
			if len(assignees) > 0 {
				type kv struct {
					name  string
					count int
				}
				rows := make([]kv, 0, len(assignees))
				for n, c := range assignees {
					rows = append(rows, kv{n, c})
				}
				slices.SortFunc(rows, func(a, b kv) int { return cmp.Compare(b.count, a.count) })

				fmt.Fprintln(out, styleBold.Render("Assignee Breakdown:"))
				for _, r := range rows {
					suffix := ""
					if r.count > 1 {
						suffix = "s"
					}
					fmt.Fprintf(out, "  %s: %d task%s\n", styleGreen.Render(r.name), r.count, suffix)
				}
				fmt.Fprintln(out)
			}

			rate := 0.0
			if len(info.Tasks) > 0 {
				rate = float64(len(by[model.StatusDone])) / float64(len(info.Tasks)) * 100
			}
			fmt.Fprintln(out, styleBold.Render(fmt.Sprintf("Completion Rate: %s",
				styleCyan.Render(fmt.Sprintf("%.1f%%", rate)))))
			fmt.Fprintln(out)
			return nil
		},
	}
}
