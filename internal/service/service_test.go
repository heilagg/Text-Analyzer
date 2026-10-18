package service

import (
	"errors"
	"testing"

	"github.com/heilagg/Text-Analyzer/internal/analyzer"
	"github.com/heilagg/Text-Analyzer/internal/storage"
	"github.com/heilagg/Text-Analyzer/internal/task"
)

type failingProcessor struct {
	err error
}

func (p failingProcessor) Process(text string) (task.Result, error) {
	return task.Result{}, p.err
}

func newTestService() *Service {
	return New(storage.NewMemoryStorage(), analyzer.New(analyzer.DefaultTopN))
}

func TestCreate(t *testing.T) {
	svc := newTestService()

	tk, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if tk.ID != 1 {
		t.Errorf("ID = %d, want 1", tk.ID)
	}
	if tk.Text != "Hello Go world" {
		t.Errorf("Text = %q, want %q", tk.Text, "Hello Go world")
	}
	if tk.Status != task.StatusPending {
		t.Errorf("Status = %s, want %s", tk.Status, task.StatusPending)
	}
	if tk.Result != nil {
		t.Errorf("Result = %v, want nil", tk.Result)
	}
}

func TestCreateInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{name: "empty", text: "", wantErr: task.ErrEmptyText},
		{name: "whitespace", text: " \n\t ", wantErr: task.ErrEmptyText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()

			tk, err := svc.Create(tt.text)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tk != nil {
				t.Errorf("Create() task = %v, want nil", tk)
			}
			if got := len(svc.List()); got != 0 {
				t.Errorf("stored tasks = %d, want 0", got)
			}
		})
	}
}

func TestInvalidInputDoesNotConsumeID(t *testing.T) {
	svc := newTestService()

	if _, err := svc.Create("   "); err == nil {
		t.Fatal("Create() expected error for invalid input")
	}
	tk, err := svc.Create("Hello")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if tk.ID != 1 {
		t.Errorf("ID = %d, want 1", tk.ID)
	}
}

func TestGet(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.ID != created.ID || got.Text != created.Text {
		t.Errorf("Get() = %+v, want %+v", got, created)
	}
}

func TestGetNonexistent(t *testing.T) {
	svc := newTestService()

	tk, err := svc.Get(42)
	if !errors.Is(err, storage.ErrTaskNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, storage.ErrTaskNotFound)
	}
	if tk != nil {
		t.Errorf("Get() task = %v, want nil", tk)
	}
}

func TestProcess(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	processed, err := svc.Process(created.ID)
	if err != nil {
		t.Fatalf("Process() unexpected error: %v", err)
	}
	if processed.Status != task.StatusCompleted {
		t.Errorf("Status = %s, want %s", processed.Status, task.StatusCompleted)
	}
	if processed.Result == nil {
		t.Fatal("Result = nil, want result")
	}
	if processed.Result.Words != 3 || processed.Result.Characters != 14 {
		t.Errorf("Result = %+v, want Words = 3, Characters = 14", *processed.Result)
	}

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if stored.Status != task.StatusCompleted || stored.Result == nil {
		t.Errorf("stored task = %+v, want completed with result", stored)
	}
}

func TestProcessTwice(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if _, err := svc.Process(created.ID); err != nil {
		t.Fatalf("first Process() unexpected error: %v", err)
	}

	_, err = svc.Process(created.ID)
	if !errors.Is(err, task.ErrInvalidTransition) {
		t.Fatalf("second Process() error = %v, want %v", err, task.ErrInvalidTransition)
	}

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if stored.Status != task.StatusCompleted {
		t.Errorf("Status = %s, want %s", stored.Status, task.StatusCompleted)
	}
}

func TestProcessNonexistent(t *testing.T) {
	svc := newTestService()

	_, err := svc.Process(42)
	if !errors.Is(err, storage.ErrTaskNotFound) {
		t.Fatalf("Process() error = %v, want %v", err, storage.ErrTaskNotFound)
	}
}

func TestProcessFailure(t *testing.T) {
	processErr := errors.New("processing failed")
	svc := New(storage.NewMemoryStorage(), failingProcessor{err: processErr})

	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	processed, err := svc.Process(created.ID)
	if !errors.Is(err, processErr) {
		t.Fatalf("Process() error = %v, want %v", err, processErr)
	}
	if processed.Status != task.StatusFailed {
		t.Errorf("Status = %s, want %s", processed.Status, task.StatusFailed)
	}
	if processed.Error != processErr.Error() {
		t.Errorf("Error = %q, want %q", processed.Error, processErr.Error())
	}
	if processed.Result != nil {
		t.Errorf("Result = %v, want nil", processed.Result)
	}
}

func TestProcessTextWithoutWords(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("!!! ???")
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	processed, err := svc.Process(created.ID)
	if !errors.Is(err, analyzer.ErrNoWords) {
		t.Fatalf("Process() error = %v, want %v", err, analyzer.ErrNoWords)
	}
	if processed.Status != task.StatusFailed {
		t.Errorf("Status = %s, want %s", processed.Status, task.StatusFailed)
	}
}

func TestMultipleIndependentTasks(t *testing.T) {
	svc := newTestService()

	inputs := []struct {
		text      string
		wantWords int
	}{
		{text: "one", wantWords: 1},
		{text: "one two", wantWords: 2},
		{text: "one two three", wantWords: 3},
	}

	ids := make([]int, 0, len(inputs))
	for _, in := range inputs {
		tk, err := svc.Create(in.text)
		if err != nil {
			t.Fatalf("Create(%q) unexpected error: %v", in.text, err)
		}
		ids = append(ids, tk.ID)
	}

	if ids[0] == ids[1] || ids[1] == ids[2] || ids[0] == ids[2] {
		t.Fatalf("IDs are not unique: %v", ids)
	}

	if _, err := svc.Process(ids[1]); err != nil {
		t.Fatalf("Process(%d) unexpected error: %v", ids[1], err)
	}

	for i, id := range ids {
		tk, err := svc.Get(id)
		if err != nil {
			t.Fatalf("Get(%d) unexpected error: %v", id, err)
		}
		if tk.Text != inputs[i].text {
			t.Errorf("task %d Text = %q, want %q", id, tk.Text, inputs[i].text)
		}
		if i == 1 {
			if tk.Status != task.StatusCompleted || tk.Result == nil {
				t.Fatalf("task %d = %+v, want completed with result", id, tk)
			}
			if tk.Result.Words != inputs[i].wantWords {
				t.Errorf("task %d Words = %d, want %d", id, tk.Result.Words, inputs[i].wantWords)
			}
			continue
		}
		if tk.Status != task.StatusPending || tk.Result != nil {
			t.Errorf("task %d = %+v, want pending without result", id, tk)
		}
	}

	for i, id := range ids {
		if i == 1 {
			continue
		}
		tk, err := svc.Process(id)
		if err != nil {
			t.Fatalf("Process(%d) unexpected error: %v", id, err)
		}
		if tk.Result.Words != inputs[i].wantWords {
			t.Errorf("task %d Words = %d, want %d", id, tk.Result.Words, inputs[i].wantWords)
		}
	}

	if got := len(svc.List()); got != len(inputs) {
		t.Errorf("List() returned %d tasks, want %d", got, len(inputs))
	}
}
