package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
)

const (
	FormatJSON    = "json"
	FormatCSV     = "csv"
	FormatSummary = "summary"
	FormatGantt   = "gantt"
)

func Render(b model.Board, format string) ([]byte, error) {
	switch format {
	case FormatJSON:
		return ToJSON(b)
	case FormatCSV:
		return ToCSV(b)
	case FormatSummary:
		return ToSummary(b)
	case FormatGantt:
		return ToGantt(b, time.Now()), nil
	default:
		return nil, fmt.Errorf("invalid format %q (want json|csv|summary|gantt)", format)
	}
}

func ToJSON(b model.Board) ([]byte, error) {
	if b.Tasks == nil {
		b.Tasks = []model.Task{}
	}
	return json.MarshalIndent(b, "", "  ")
}

func ToCSV(b model.Board) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"ID", "Title", "Status", "Priority", "Assignee", "Tags", "Due Date", "Created", "Updated"}); err != nil {
		return nil, err
	}
	for _, t := range b.Tasks {
		row := []string{
			t.ID,
			t.Title,
			string(t.Status),
			string(t.Priority),
			t.Assignee,
			strings.Join(t.Tags, ";"),
			t.DueDate,
			t.CreatedAt.UTC().Format(time.RFC3339Nano),
			t.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ToSummary(b model.Board) ([]byte, error) {
	counts := map[model.TaskStatus]int{}
	for _, t := range b.Tasks {
		counts[t.Status]++
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Board: %s\n", b.Name)
	if b.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", b.Description)
	}
	fmt.Fprintf(&sb, "Total Tasks: %d\n", len(b.Tasks))
	sb.WriteString("\n")
	sb.WriteString("Status Breakdown:\n")
	fmt.Fprintf(&sb, "  TODO: %d\n", counts[model.StatusTodo])
	fmt.Fprintf(&sb, "  IN PROGRESS: %d\n", counts[model.StatusInProgress])
	fmt.Fprintf(&sb, "  DONE: %d\n", counts[model.StatusDone])
	fmt.Fprintf(&sb, "  BLOCKED: %d", counts[model.StatusBlocked])
	return []byte(sb.String()), nil
}

var ganttSections = []struct {
	status model.TaskStatus
	label  string
	tag    string
}{
	{model.StatusTodo, "TODO", ""},
	{model.StatusInProgress, "IN PROGRESS", "active, "},
	{model.StatusBlocked, "BLOCKED", "crit, "},
	{model.StatusDone, "DONE", "done, "},
}

// ganttEscaper replaces the characters that end a Mermaid gantt title or
// start a comment with Mermaid entity codes.
var ganttEscaper = strings.NewReplacer(":", "#58;", ";", "#59;", "#", "#35;")

// ToGantt renders a Mermaid gantt chart with one section for each status.
// Bars use explicit dates from the timeline layout; today ends open bars.
func ToGantt(b model.Board, today time.Time) []byte {
	l := timeline.Build(b.Tasks, today)
	var sb strings.Builder
	sb.WriteString("gantt\n")
	fmt.Fprintf(&sb, "    title %s\n", ganttEscaper.Replace(b.Name))
	sb.WriteString("    dateFormat YYYY-MM-DD\n")
	sb.WriteString("    inclusiveEndDates\n")
	for _, sec := range ganttSections {
		header := false
		for _, bar := range l.Bars {
			if bar.Task.Status != sec.status {
				continue
			}
			if !header {
				fmt.Fprintf(&sb, "    section %s\n", sec.label)
				header = true
			}
			fmt.Fprintf(&sb, "    %s :%s%s, %s, %s\n",
				ganttEscaper.Replace(bar.Task.Title), sec.tag, bar.Task.ID,
				bar.Start.Format(time.DateOnly), bar.End.Format(time.DateOnly))
		}
	}
	for _, t := range l.Unscheduled {
		fmt.Fprintf(&sb, "    %%%% unscheduled: %s\n", t.ID)
	}
	return []byte(sb.String())
}
