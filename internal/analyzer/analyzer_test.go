package analyzer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/heilagg/Text-Analyzer/internal/task"
)

func TestProcess(t *testing.T) {
	tests := []struct {
		name string
		text string
		want task.Result
	}{
		{
			name: "simple sentence",
			text: "Hello Go world",
			want: task.Result{
				Words:       3,
				Characters:  14,
				Bytes:       14,
				Sentences:   1,
				LongestWord: "Hello",
				TopWords: []task.WordFrequency{
					{Word: "go", Count: 1},
					{Word: "hello", Count: 1},
					{Word: "world", Count: 1},
				},
			},
		},
		{
			name: "several sentences with repeated words",
			text: "Go is simple. Go is fast! Is Go fun? Yes.",
			want: task.Result{
				Words:       10,
				Characters:  41,
				Bytes:       41,
				Sentences:   4,
				LongestWord: "simple",
				TopWords: []task.WordFrequency{
					{Word: "go", Count: 3},
					{Word: "is", Count: 3},
					{Word: "fast", Count: 1},
				},
			},
		},
		{
			name: "kazakh text counts runes and bytes separately",
			text: "Сәлем, әлем! Бұл Go бағдарламасы.",
			want: task.Result{
				Words:       5,
				Characters:  33,
				Bytes:       57,
				Sentences:   2,
				LongestWord: "бағдарламасы",
				TopWords: []task.WordFrequency{
					{Word: "go", Count: 1},
					{Word: "бағдарламасы", Count: 1},
					{Word: "бұл", Count: 1},
				},
			},
		},
		{
			name: "repeated punctuation ends one sentence",
			text: "Wait... What?! Really",
			want: task.Result{
				Words:       3,
				Characters:  21,
				Bytes:       21,
				Sentences:   3,
				LongestWord: "Really",
				TopWords: []task.WordFrequency{
					{Word: "really", Count: 1},
					{Word: "wait", Count: 1},
					{Word: "what", Count: 1},
				},
			},
		},
		{
			name: "word frequency ignores case",
			text: "Go go GO gopher",
			want: task.Result{
				Words:       4,
				Characters:  15,
				Bytes:       15,
				Sentences:   1,
				LongestWord: "gopher",
				TopWords: []task.WordFrequency{
					{Word: "go", Count: 3},
					{Word: "gopher", Count: 1},
				},
			},
		},
		{
			name: "first longest word wins a tie",
			text: "abc xyz de",
			want: task.Result{
				Words:       3,
				Characters:  10,
				Bytes:       10,
				Sentences:   1,
				LongestWord: "abc",
				TopWords: []task.WordFrequency{
					{Word: "abc", Count: 1},
					{Word: "de", Count: 1},
					{Word: "xyz", Count: 1},
				},
			},
		},
	}

	a := New(DefaultTopN)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.Process(tt.text)
			if err != nil {
				t.Fatalf("Process() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}

func TestProcessNoWords(t *testing.T) {
	tests := []string{"!!! ???", "... ,,, ---"}

	a := New(DefaultTopN)
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			_, err := a.Process(text)
			if !errors.Is(err, ErrNoWords) {
				t.Errorf("Process(%q) error = %v, want %v", text, err, ErrNoWords)
			}
		})
	}
}

func TestNewTopN(t *testing.T) {
	tests := []struct {
		name  string
		topN  int
		wantN int
	}{
		{name: "custom value", topN: 1, wantN: 1},
		{name: "zero falls back to default", topN: 0, wantN: DefaultTopN},
		{name: "negative falls back to default", topN: -5, wantN: DefaultTopN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.topN).Process("one two three four five")
			if err != nil {
				t.Fatalf("Process() unexpected error: %v", err)
			}
			if len(got.TopWords) != tt.wantN {
				t.Errorf("len(TopWords) = %d, want %d", len(got.TopWords), tt.wantN)
			}
		})
	}
}
