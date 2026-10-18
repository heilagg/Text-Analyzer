package task

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const MaxTextLength = 10000

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

var (
	ErrEmptyText         = errors.New("text is empty")
	ErrTextTooLong       = errors.New("text is too long")
	ErrInvalidTransition = errors.New("invalid state transition")
)

var allowedTransitions = map[Status][]Status{
	StatusPending:    {StatusProcessing},
	StatusProcessing: {StatusCompleted, StatusFailed},
}

func (s Status) CanTransitionTo(next Status) bool {
	for _, allowed := range allowedTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

type WordFrequency struct {
	Word  string
	Count int
}

type Result struct {
	Words       int
	Characters  int
	Bytes       int
	Sentences   int
	LongestWord string
	TopWords    []WordFrequency
}

type Task struct {
	ID     int
	Text   string
	Status Status
	Result *Result
	Error  string
}

func ValidateText(text string) error {
	if strings.TrimSpace(text) == "" {
		return ErrEmptyText
	}
	if n := utf8.RuneCountInString(text); n > MaxTextLength {
		return fmt.Errorf("%w: %d characters, max %d", ErrTextTooLong, n, MaxTextLength)
	}
	return nil
}

func New(id int, text string) (*Task, error) {
	if err := ValidateText(text); err != nil {
		return nil, err
	}
	return &Task{ID: id, Text: text, Status: StatusPending}, nil
}

func (t *Task) transitionTo(next Status) error {
	if !t.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, t.Status, next)
	}
	t.Status = next
	return nil
}

func (t *Task) Start() error {
	return t.transitionTo(StatusProcessing)
}

func (t *Task) Complete(result Result) error {
	if err := t.transitionTo(StatusCompleted); err != nil {
		return err
	}
	t.Result = &result
	return nil
}

func (t *Task) Fail(reason error) error {
	if err := t.transitionTo(StatusFailed); err != nil {
		return err
	}
	t.Error = reason.Error()
	return nil
}
