package service

import (
	"fmt"

	"github.com/heilagg/Text-Analyzer/internal/task"
)

type Processor interface {
	Process(text string) (task.Result, error)
}

type Repository interface {
	NextID() int
	Save(t *task.Task)
	Get(id int) (*task.Task, error)
	List() []*task.Task
}

type Service struct {
	repo      Repository
	processor Processor
}

func New(repo Repository, processor Processor) *Service {
	return &Service{repo: repo, processor: processor}
}

func (s *Service) Create(text string) (*task.Task, error) {
	if err := task.ValidateText(text); err != nil {
		return nil, err
	}

	t, err := task.New(s.repo.NextID(), text)
	if err != nil {
		return nil, err
	}

	s.repo.Save(t)
	return t, nil
}

func (s *Service) Get(id int) (*task.Task, error) {
	return s.repo.Get(id)
}

func (s *Service) List() []*task.Task {
	return s.repo.List()
}

func (s *Service) Process(id int) (*task.Task, error) {
	t, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}

	if err := t.Start(); err != nil {
		return nil, err
	}
	s.repo.Save(t)

	result, err := s.processor.Process(t.Text)
	if err != nil {
		if failErr := t.Fail(err); failErr != nil {
			return nil, failErr
		}
		s.repo.Save(t)
		return t, fmt.Errorf("process task %d: %w", id, err)
	}

	if err := t.Complete(result); err != nil {
		return nil, err
	}
	s.repo.Save(t)
	return t, nil
}
