package task

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{name: "valid text", text: "Hello Go world", wantErr: nil},
		{name: "empty text", text: "", wantErr: ErrEmptyText},
		{name: "only whitespace", text: "   \t\n ", wantErr: ErrEmptyText},
		{name: "exactly max length", text: strings.Repeat("a", MaxTextLength), wantErr: nil},
		{name: "longer than max", text: strings.Repeat("a", MaxTextLength+1), wantErr: ErrTextTooLong},
		{name: "max length counts runes, not bytes", text: strings.Repeat("ә", MaxTextLength), wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateText(tt.text)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateText() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tk, err := New(7, "Hello Go")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if tk.ID != 7 {
		t.Errorf("ID = %d, want 7", tk.ID)
	}
	if tk.Text != "Hello Go" {
		t.Errorf("Text = %q, want %q", tk.Text, "Hello Go")
	}
	if tk.Status != StatusPending {
		t.Errorf("Status = %s, want %s", tk.Status, StatusPending)
	}
	if tk.Result != nil {
		t.Errorf("Result = %v, want nil", tk.Result)
	}
}

func TestNewRejectsInvalidText(t *testing.T) {
	tk, err := New(1, "   ")
	if !errors.Is(err, ErrEmptyText) {
		t.Fatalf("New() error = %v, want %v", err, ErrEmptyText)
	}
	if tk != nil {
		t.Errorf("New() task = %v, want nil", tk)
	}
}

func TestCompleteLifecycle(t *testing.T) {
	tk, err := New(1, "Hello Go")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if err := tk.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
	if tk.Status != StatusProcessing {
		t.Fatalf("Status after Start = %s, want %s", tk.Status, StatusProcessing)
	}

	if err := tk.Complete(Result{Words: 2}); err != nil {
		t.Fatalf("Complete() unexpected error: %v", err)
	}
	if tk.Status != StatusCompleted {
		t.Errorf("Status after Complete = %s, want %s", tk.Status, StatusCompleted)
	}
	if tk.Result == nil || tk.Result.Words != 2 {
		t.Errorf("Result = %v, want Words = 2", tk.Result)
	}
}

func TestFailLifecycle(t *testing.T) {
	tk, err := New(1, "Hello Go")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if err := tk.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
	if err := tk.Fail(errors.New("boom")); err != nil {
		t.Fatalf("Fail() unexpected error: %v", err)
	}
	if tk.Status != StatusFailed {
		t.Errorf("Status = %s, want %s", tk.Status, StatusFailed)
	}
	if tk.Error != "boom" {
		t.Errorf("Error = %q, want %q", tk.Error, "boom")
	}
	if tk.Result != nil {
		t.Errorf("Result = %v, want nil", tk.Result)
	}
}

func TestInvalidTransitions(t *testing.T) {
	tests := []struct {
		name   string
		from   Status
		action func(tk *Task) error
	}{
		{
			name:   "complete without start",
			from:   StatusPending,
			action: func(tk *Task) error { return tk.Complete(Result{}) },
		},
		{
			name:   "fail without start",
			from:   StatusPending,
			action: func(tk *Task) error { return tk.Fail(errors.New("boom")) },
		},
		{
			name:   "start twice",
			from:   StatusProcessing,
			action: func(tk *Task) error { return tk.Start() },
		},
		{
			name:   "restart completed task",
			from:   StatusCompleted,
			action: func(tk *Task) error { return tk.Start() },
		},
		{
			name:   "restart failed task",
			from:   StatusFailed,
			action: func(tk *Task) error { return tk.Start() },
		},
		{
			name:   "complete twice",
			from:   StatusCompleted,
			action: func(tk *Task) error { return tk.Complete(Result{}) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &Task{ID: 1, Text: "Hello", Status: tt.from}

			err := tt.action(tk)
			if !errors.Is(err, ErrInvalidTransition) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidTransition)
			}
			if tk.Status != tt.from {
				t.Errorf("Status changed to %s, want it to stay %s", tk.Status, tt.from)
			}
			if tk.Result != nil {
				t.Errorf("Result = %v, want nil", tk.Result)
			}
		})
	}
}
