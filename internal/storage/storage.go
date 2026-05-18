package storage

import (
	"os"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
)

const DefaultFile = "tasks.md"

type Store interface {
	Read() (*model.Board, error)
	Write(*model.Board) error
	Path() string
}

type MarkdownStore struct {
	path  string
	clock func() time.Time
}

func NewMarkdown(path string) *MarkdownStore {
	if path == "" {
		path = DefaultFile
	}
	return &MarkdownStore{path: path, clock: time.Now}
}

func (s *MarkdownStore) WithClock(fn func() time.Time) *MarkdownStore {
	s.clock = fn
	return s
}

func (s *MarkdownStore) Path() string { return s.path }

func (s *MarkdownStore) Read() (*model.Board, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.defaultBoard(), nil
		}
		return nil, err
	}
	return parseMarkdown(data, s.clock), nil
}

func (s *MarkdownStore) Write(b *model.Board) error {
	out := renderMarkdown(b)
	return atomicWrite(s.path, out)
}

func (s *MarkdownStore) defaultBoard() *model.Board {
	now := s.clock().UTC()
	return &model.Board{
		Name:        "My Board",
		Description: "Task management board",
		Tasks:       []model.Task{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
