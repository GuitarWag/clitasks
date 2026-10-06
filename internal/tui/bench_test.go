package tui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/GuitarWag/clitasks/v3/internal/board"
	"github.com/GuitarWag/clitasks/v3/internal/model"
)

// bigModel is the demo board plus n generated tasks with dates and links.
func bigModel(b *testing.B, n int) Model {
	t := &testing.T{}
	m := demoBoard(t, "catppuccin-mocha", 200, 60)
	statuses := []model.TaskStatus{model.StatusTodo, model.StatusInProgress, model.StatusBlocked, model.StatusDone}
	prev := ""
	for i := range n {
		start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i%60)
		in := board.AddInput{
			Priority: model.PriorityMedium, Assignee: fmt.Sprintf("dev%d", i%7), Tags: []string{"t" + fmt.Sprint(i%5)},
			Start: start.Format(time.DateOnly), DueDate: start.AddDate(0, 0, 3+i%10).Format(time.DateOnly),
		}
		if i%3 == 0 && prev != "" {
			in.After = []string{prev}
		}
		tk, err := m.board.Add(fmt.Sprintf("Generated task number %d with a long title", i), in)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := m.board.Move(tk.ID, statuses[i%4]); err != nil {
			b.Fatal(err)
		}
		prev = tk.ID
	}
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	return mm.(Model)
}

func BenchmarkRenderBoard(b *testing.B) {
	m := bigModel(b, 500)
	b.ResetTimer()
	for range b.N {
		_ = m.render()
	}
}

func BenchmarkRenderTimeline(b *testing.B) {
	m := send(bigModel(b, 500), "tab")
	b.ResetTimer()
	for range b.N {
		_ = m.render()
	}
}

func BenchmarkKeyDown(b *testing.B) {
	m := bigModel(b, 500)
	b.ResetTimer()
	for range b.N {
		m = send(m, "j")
	}
}
