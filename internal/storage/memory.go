package storage

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/heilagg/Text-Analyzer/internal/task"
)

var ErrTaskNotFound = errors.New("task not found")

type MemoryStorage struct {
	tasks  map[int]*task.Task
	nextID int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks:  make(map[int]*task.Task),
		nextID: 1,
	}
}

func (s *MemoryStorage) NextID() int {
	id := s.nextID
	s.nextID++
	return id
}

func (s *MemoryStorage) Save(t *task.Task) {
	s.tasks[t.ID] = t
}

func (s *MemoryStorage) Get(id int) (*task.Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("%w: id %d", ErrTaskNotFound, id)
	}
	return t, nil
}

func (s *MemoryStorage) List() []*task.Task {
	ids := slices.Sorted(maps.Keys(s.tasks))
	result := make([]*task.Task, 0, len(ids))
	for _, id := range ids {
		result = append(result, s.tasks[id])
	}
	return result
}
