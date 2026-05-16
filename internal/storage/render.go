package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
)

func renderMarkdown(b *model.Board) []byte {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Board: %s\n", b.Name)
	if b.Description != "" {
		fmt.Fprintf(&sb, "> Description: %s\n", b.Description)
	}
	fmt.Fprintf(&sb, "> Created: %s | Updated: %s\n",
		formatTime(b.CreatedAt), formatTime(b.UpdatedAt))
	sb.WriteString("\n")

	writeSection(&sb, "TODO", "[ ]", filterByStatus(b.Tasks, model.StatusTodo))
	writeSection(&sb, "IN PROGRESS", "[>]", filterByStatus(b.Tasks, model.StatusInProgress))
	writeSection(&sb, "DONE", "[x]", filterByStatus(b.Tasks, model.StatusDone))
	writeSection(&sb, "BLOCKED", "[!]", filterByStatus(b.Tasks, model.StatusBlocked))

	return []byte(sb.String())
}

func writeSection(w io.Writer, title, checkbox string, tasks []model.Task) {
	fmt.Fprintf(w, "## %s\n\n", title)
	if len(tasks) == 0 {
		fmt.Fprintln(w, "_No tasks_")
		fmt.Fprintln(w)
		return
	}
	for _, t := range tasks {
		writeTask(w, t, checkbox)
	}
	fmt.Fprintln(w)
}

func writeTask(w io.Writer, t model.Task, checkbox string) {
	var meta []string
	meta = append(meta, fmt.Sprintf("`priority:%s`", t.Priority))
	if t.Assignee != "" {
		meta = append(meta, fmt.Sprintf("`assignee:%s`", t.Assignee))
	}
	if len(t.Tags) > 0 {
		meta = append(meta, fmt.Sprintf("`tags:%s`", strings.Join(t.Tags, ",")))
	}
	if t.DueDate != "" {
		meta = append(meta, fmt.Sprintf("`due:%s`", t.DueDate))
	}
	fmt.Fprintf(w, "- %s [%s] **%s** %s\n", checkbox, t.ID, t.Title, strings.Join(meta, " "))
	if t.Description != "" {
		fmt.Fprintf(w, "  > %s\n", t.Description)
	}
	fmt.Fprintf(w, "  > Created: %s | Updated: %s\n",
		formatTime(t.CreatedAt), formatTime(t.UpdatedAt))
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func filterByStatus(tasks []model.Task, s model.TaskStatus) []model.Task {
	out := make([]model.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Status == s {
			out = append(out, t)
		}
	}
	return out
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tasks-*.md")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}
