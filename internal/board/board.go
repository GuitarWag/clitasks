package board

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
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
		rng:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()>>32))),
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
	Start       string
	After       []string
}

type UpdateInput struct {
	Title       *string
	Description *string
	Priority    *model.TaskPriority
	Assignee    *string
	Tags        *[]string
	DueDate     *string
	Start       *string
	After       *[]string
	Status      *model.TaskStatus
}

type Filter struct {
	Status   *model.TaskStatus
	Priority *model.TaskPriority
	Assignee string
	Tags     []string
}

func (b *Board) Path() string      { return b.store.Path() }
func (b *Board) Info() model.Board { return *b.data }

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
		Tags:        slices.Clone(in.Tags),
		DueDate:     in.DueDate,
		Start:       in.Start,
		After:       dedupe(in.After),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := b.validate(t, in.DueDate != "", in.Start != "", len(in.After) > 0); err != nil {
		return model.Task{}, err
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
	nt := b.data.Tasks[idx]
	t := &nt
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
		t.Tags = slices.Clone(*in.Tags)
	}
	if in.DueDate != nil {
		t.DueDate = *in.DueDate
	}
	if in.Start != nil {
		t.Start = *in.Start
	}
	if in.After != nil {
		t.After = dedupe(*in.After)
	}
	if in.Status != nil {
		if *in.Status == model.StatusInProgress && t.Start == "" {
			t.Start = now.In(time.Local).Format(time.DateOnly)
		}
		t.Status = *in.Status
	}
	if err := b.validate(*t, in.DueDate != nil && *in.DueDate != "",
		in.Start != nil && *in.Start != "", in.After != nil); err != nil {
		return model.Task{}, err
	}
	t.UpdatedAt = now
	b.data.Tasks[idx] = nt
	b.data.UpdatedAt = now
	if err := b.save(); err != nil {
		return model.Task{}, err
	}
	return *t, nil
}

func (b *Board) Move(id string, s model.TaskStatus) (model.Task, error) {
	return b.Update(id, UpdateInput{Status: &s})
}

func (b *Board) Delete(id string) (model.Task, error) {
	idx := b.indexOf(id)
	if idx < 0 {
		return model.Task{}, ErrNotFound
	}
	removed := b.data.Tasks[idx]
	b.data.Tasks = append(b.data.Tasks[:idx], b.data.Tasks[idx+1:]...)
	for i := range b.data.Tasks {
		b.data.Tasks[i].After = slices.DeleteFunc(slices.Clone(b.data.Tasks[i].After),
			func(s string) bool { return s == id })
		if len(b.data.Tasks[i].After) == 0 {
			b.data.Tasks[i].After = nil
		}
	}
	b.data.UpdatedAt = b.clock().UTC()
	if err := b.save(); err != nil {
		return model.Task{}, err
	}
	return removed, nil
}

type MetaInput struct {
	Name        *string
	Description *string
}

func (b *Board) UpdateMeta(in MetaInput) error {
	if in.Name != nil {
		b.data.Name = *in.Name
	}
	if in.Description != nil {
		b.data.Description = *in.Description
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

// Dependents returns the IDs of the tasks whose After list holds id.
func (b *Board) Dependents(id string) []string {
	var out []string
	for _, t := range b.data.Tasks {
		if slices.Contains(t.After, id) {
			out = append(out, t.ID)
		}
	}
	return out
}

// validate checks the date fields that the caller set, start <= due when both
// are dates, and the After links when the caller set them.
func (b *Board) validate(t model.Task, checkDue, checkStart, checkAfter bool) error {
	var due, start time.Time
	var dueOK, startOK bool
	if t.DueDate != "" {
		d, err := model.ParseDate(t.DueDate)
		if err != nil && checkDue {
			return fmt.Errorf("due: %w", err)
		}
		due, dueOK = d, err == nil
	}
	if t.Start != "" {
		s, err := model.ParseDate(t.Start)
		if err != nil && checkStart {
			return fmt.Errorf("start: %w", err)
		}
		start, startOK = s, err == nil
	}
	if dueOK && startOK && start.After(due) {
		return fmt.Errorf("start %s is after due %s", t.Start, t.DueDate)
	}
	if !checkAfter {
		return nil
	}
	for _, dep := range t.After {
		if dep == t.ID {
			return fmt.Errorf("after: task cannot wait for itself")
		}
		if b.indexOf(dep) < 0 {
			return fmt.Errorf("after: %w: %s", ErrNotFound, dep)
		}
		if b.reaches(dep, t.ID, map[string]bool{}) {
			return fmt.Errorf("after: %s already waits for %s, this makes a cycle", dep, t.ID)
		}
	}
	return nil
}

// reaches reports whether from waits for target through a chain of After links.
func (b *Board) reaches(from, target string, seen map[string]bool) bool {
	if from == target {
		return true
	}
	if seen[from] {
		return false
	}
	seen[from] = true
	idx := b.indexOf(from)
	if idx < 0 {
		return false
	}
	for _, next := range b.data.Tasks[idx].After {
		if b.reaches(next, target, seen) {
			return true
		}
	}
	return false
}

func dedupe(ids []string) []string {
	var out []string
	for _, id := range ids {
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
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
