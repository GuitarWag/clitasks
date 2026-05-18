package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
)

const (
	FormatJSON    = "json"
	FormatCSV     = "csv"
	FormatSummary = "summary"
)

func Render(b model.Board, format string) ([]byte, error) {
	switch format {
	case FormatJSON:
		return ToJSON(b)
	case FormatCSV:
		return ToCSV(b)
	case FormatSummary:
		return ToSummary(b)
	default:
		return nil, fmt.Errorf("invalid format %q (want json|csv|summary)", format)
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
