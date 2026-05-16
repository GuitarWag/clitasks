package board

import (
	"errors"
	"math/rand"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
)

var ErrNotFound = errors.New("task not found")

type Board struct {
	store storage.Store
	data  *model.Board
	clock func() time.Time
	rng   *rand.Rand
}

type Option func(*Board)

func WithClock(fn func() time.Time) Option { return func(b *Board) { b.clock = fn } }
func WithRand(r *rand.Rand) Option         { return func(b *Board) { b.rng = r } }

func Open(s storage.Store, opts ...Option) (*Board, error) {
	b := &Board{
		store: s,
		clock: func() time.Time { return time.Now().UTC() },
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, opt := range opts {
		opt(b)
	}
	d, err := s.Read()
	if err != nil {
		return nil, err
	}
	b.data = d
	return b, nil
}

type AddInput struct {
	Description string
	Priority    model.TaskPriority
	Assignee    string
	Tags        []string
	DueDate     string
}

type UpdateInput struct {
	Title       *string
	Description *string
	Priority    *model.TaskPriority
	Assignee    *string
	Tags        *[]string
	DueDate     *string
	Status      *model.TaskStatus
}

type Filter struct {
	Status   *model.TaskStatus
	Priority *model.TaskPriority
	Assignee string
	Tags     []string
}

func (b *Board) Path() string       { return b.store.Path() }
func (b *Board) Info() model.Board  { return *b.data }

func (b *Board) Add(title string, in AddInput) (model.Task, error) {
	now := b.clock().UTC()
	priority := in.Priority
	if priority == "" {
		priority = model.PriorityMedium
	}
	t := model.Task{
		ID:          newID(now, b.rng),
		Title:       title,
		Description: in.Description,
		Status:      model.StatusTodo,
		Priority:    priority,
		Assignee:    in.Assignee,
		Tags:        in.Tags,
		DueDate:     in.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	b.data.Tasks = append(b.data.Tasks, t)
	b.data.UpdatedAt = now
	if err := b.save(); err != nil {
		return model.Task{}, err
	}
	return t, nil
}

func (b *Board) Update(id string, in UpdateInput) (model.Task, error) {
	idx := b.indexOf(id)
	if idx < 0 {
		return model.Task{}, ErrNotFound
	}
	now := b.clock().UTC()
	t := &b.data.Tasks[idx]
	if in.Title != nil {
		t.Title = *in.Title
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.Priority != nil {
		t.Priority = *in.Priority
	}
	if in.Assignee != nil {
		t.Assignee = *in.Assignee
	}
	if in.Tags != nil {
		t.Tags = *in.Tags
	}
	if in.DueDate != nil {
		t.DueDate = *in.DueDate
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	t.UpdatedAt = now
	b.data.UpdatedAt = now
	if err := b.save(); err != nil {
		return model.Task{}, err
	}
	return *t, nil
}

func (b *Board) Move(id string, s model.TaskStatus) (model.Task, error) {
	return b.Update(id, UpdateInput{Status: &s})
}

func (b *Board) Delete(id string) error {
	idx := b.indexOf(id)
	if idx < 0 {
		return ErrNotFound
	}
	b.data.Tasks = append(b.data.Tasks[:idx], b.data.Tasks[idx+1:]...)
	b.data.UpdatedAt = b.clock().UTC()
	return b.save()
}

func (b *Board) UpdateMeta(name, description string) error {
	if name != "" {
		b.data.Name = name
	}
	if description != "" {
		b.data.Description = description
	}
	b.data.UpdatedAt = b.clock().UTC()
	return b.save()
}

func (b *Board) Get(id string) (model.Task, bool) {
	idx := b.indexOf(id)
	if idx < 0 {
		return model.Task{}, false
	}
	return b.data.Tasks[idx], true
}

func (b *Board) List(f Filter) []model.Task {
	out := make([]model.Task, 0, len(b.data.Tasks))
	for _, t := range b.data.Tasks {
		if f.Status != nil && t.Status != *f.Status {
			continue
		}
		if f.Priority != nil && t.Priority != *f.Priority {
			continue
		}
		if f.Assignee != "" && t.Assignee != f.Assignee {
			continue
		}
		if len(f.Tags) > 0 {
			if !hasAnyTag(t.Tags, f.Tags) {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

func (b *Board) ByStatus() map[model.TaskStatus][]model.Task {
	out := map[model.TaskStatus][]model.Task{
		model.StatusTodo:       nil,
		model.StatusInProgress: nil,
		model.StatusDone:       nil,
		model.StatusBlocked:    nil,
	}
	for _, t := range b.data.Tasks {
		out[t.Status] = append(out[t.Status], t)
	}
	return out
}

func (b *Board) indexOf(id string) int {
	for i, t := range b.data.Tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func hasAnyTag(taskTags, want []string) bool {
	for _, w := range want {
		for _, tt := range taskTags {
			if tt == w {
				return true
			}
		}
	}
	return false
}

func (b *Board) save() error { return b.store.Write(b.data) }
