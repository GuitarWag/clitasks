package board

import (
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBoard(t *testing.T) *Board {
	t.Helper()
	p := filepath.Join(t.TempDir(), "b.md")
	s := storage.NewMarkdown(p)
	clk := func() time.Time { return time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC) }
	b, err := Open(s, WithClock(clk), WithRand(rand.New(rand.NewPCG(1, 2))))
	require.NoError(t, err)
	return b
}

func ptr[T any](v T) *T { return &v }

func TestAdd_defaults(t *testing.T) {
	b := newTestBoard(t)
	tk, err := b.Add("hello", AddInput{})
	require.NoError(t, err)
	assert.Equal(t, "hello", tk.Title)
	assert.Equal(t, model.StatusTodo, tk.Status)
	assert.Equal(t, model.PriorityMedium, tk.Priority)
	assert.Contains(t, tk.ID, "T-")
}

func TestAdd_persists(t *testing.T) {
	b := newTestBoard(t)
	_, err := b.Add("hello", AddInput{Priority: model.PriorityHigh, Tags: []string{"x"}})
	require.NoError(t, err)

	// Reopen from disk
	b2, err := Open(storage.NewMarkdown(b.Path()))
	require.NoError(t, err)
	tasks := b2.List(Filter{})
	require.Len(t, tasks, 1)
	assert.Equal(t, "hello", tasks[0].Title)
	assert.Equal(t, model.PriorityHigh, tasks[0].Priority)
	assert.Equal(t, []string{"x"}, tasks[0].Tags)
}

func TestAdd_uniqueIDs(t *testing.T) {
	b := newTestBoard(t)
	ids := map[string]bool{}
	for i := 0; i < 50; i++ {
		tk, err := b.Add("t", AddInput{})
		require.NoError(t, err)
		assert.False(t, ids[tk.ID], "duplicate id %s", tk.ID)
		ids[tk.ID] = true
	}
}

func TestUpdate_partial(t *testing.T) {
	b := newTestBoard(t)
	tk, _ := b.Add("orig", AddInput{Priority: model.PriorityMedium})

	upd, err := b.Update(tk.ID, UpdateInput{
		Title:    ptr("renamed"),
		Priority: ptr(model.PriorityCritical),
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed", upd.Title)
	assert.Equal(t, model.PriorityCritical, upd.Priority)
	assert.Equal(t, model.StatusTodo, upd.Status, "untouched fields preserved")
}

func TestUpdate_notFound(t *testing.T) {
	b := newTestBoard(t)
	_, err := b.Update("nope", UpdateInput{Title: ptr("x")})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMove(t *testing.T) {
	b := newTestBoard(t)
	tk, _ := b.Add("x", AddInput{})
	moved, err := b.Move(tk.ID, model.StatusDone)
	require.NoError(t, err)
	assert.Equal(t, model.StatusDone, moved.Status)
}

func TestDelete(t *testing.T) {
	b := newTestBoard(t)
	tk, _ := b.Add("x", AddInput{})
	removed, err := b.Delete(tk.ID)
	require.NoError(t, err)
	assert.Equal(t, tk.ID, removed.ID, "Delete returns the removed task")
	_, ok := b.Get(tk.ID)
	assert.False(t, ok)
}

func TestDelete_notFound(t *testing.T) {
	b := newTestBoard(t)
	_, err := b.Delete("nope")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList_filters(t *testing.T) {
	b := newTestBoard(t)
	a, _ := b.Add("a", AddInput{Priority: model.PriorityHigh, Assignee: "alice", Tags: []string{"x", "y"}})
	bb, _ := b.Add("b", AddInput{Priority: model.PriorityLow, Assignee: "bob", Tags: []string{"y"}})
	_, _ = b.Move(bb.ID, model.StatusDone)

	statusTodo := model.StatusTodo
	got := b.List(Filter{Status: &statusTodo})
	require.Len(t, got, 1)
	assert.Equal(t, a.ID, got[0].ID)

	prio := model.PriorityLow
	got = b.List(Filter{Priority: &prio})
	require.Len(t, got, 1)
	assert.Equal(t, bb.ID, got[0].ID)

	got = b.List(Filter{Assignee: "alice"})
	require.Len(t, got, 1)
	assert.Equal(t, a.ID, got[0].ID)

	got = b.List(Filter{Tags: []string{"y"}})
	assert.Len(t, got, 2, "both tasks have tag y")

	got = b.List(Filter{Tags: []string{"x"}})
	require.Len(t, got, 1)
	assert.Equal(t, a.ID, got[0].ID)
}

func TestByStatus(t *testing.T) {
	b := newTestBoard(t)
	_, _ = b.Add("a", AddInput{})
	t2, _ := b.Add("b", AddInput{})
	_, _ = b.Move(t2.ID, model.StatusDone)

	g := b.ByStatus()
	assert.Len(t, g[model.StatusTodo], 1)
	assert.Len(t, g[model.StatusDone], 1)
	assert.Empty(t, g[model.StatusInProgress])
	assert.Empty(t, g[model.StatusBlocked])
}

func TestUpdateMeta(t *testing.T) {
	b := newTestBoard(t)
	require.NoError(t, b.UpdateMeta(MetaInput{Name: ptr("New Name"), Description: ptr("New Desc")}))
	assert.Equal(t, "New Name", b.Info().Name)
	assert.Equal(t, "New Desc", b.Info().Description)
}

func TestUpdateMeta_clearsWithEmptyPointer(t *testing.T) {
	b := newTestBoard(t)
	require.NoError(t, b.UpdateMeta(MetaInput{Name: ptr("Initial"), Description: ptr("Initial desc")}))
	require.NoError(t, b.UpdateMeta(MetaInput{Description: ptr("")}))
	assert.Equal(t, "Initial", b.Info().Name, "name preserved when nil")
	assert.Empty(t, b.Info().Description, "empty pointer clears description")
}

func TestNewID_format(t *testing.T) {
	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	r := rand.New(rand.NewPCG(42, 0))
	id := newID(now, r)
	assert.Regexp(t, `^T-[0-9A-Z]+-[0-9A-Z]{3}$`, id)
}
