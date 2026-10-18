package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/heilagg/Text-Analyzer/internal/analyzer"
	"github.com/heilagg/Text-Analyzer/internal/service"
	"github.com/heilagg/Text-Analyzer/internal/storage"
	"github.com/heilagg/Text-Analyzer/internal/task"
)

func main() {
	svc := service.New(storage.NewMemoryStorage(), analyzer.New(analyzer.DefaultTopN))

	inputs := []string{
		"Hello Go world",
		"Go is simple. Go is fast! Is Go fun? Yes.",
		"Сәлем, әлем! Бұл Go бағдарламасы.",
		"!!! ???",
		"   ",
	}

	fmt.Println("== Create ==")
	for _, text := range inputs {
		t, err := svc.Create(text)
		if err != nil {
			fmt.Printf("create %q: %v\n", text, err)
			continue
		}
		fmt.Printf("created task %d [%s]\n", t.ID, t.Status)
	}

	fmt.Println("\n== Process ==")
	for _, t := range svc.List() {
		if _, err := svc.Process(t.ID); err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("task %d processed\n", t.ID)
	}

	fmt.Println("\n== Results ==")
	for _, t := range svc.List() {
		printTask(t)
	}

	fmt.Println("\n== Invalid operations ==")
	if _, err := svc.Process(1); errors.Is(err, task.ErrInvalidTransition) {
		fmt.Println("process task 1 again:", err)
	}
	if _, err := svc.Get(42); errors.Is(err, storage.ErrTaskNotFound) {
		fmt.Println("get task 42:", err)
	}
}

func printTask(t *task.Task) {
	fmt.Printf("task %d [%s] %q\n", t.ID, t.Status, t.Text)
	switch t.Status {
	case task.StatusCompleted:
		r := t.Result
		fmt.Printf("  words: %d, characters: %d, bytes: %d, sentences: %d\n",
			r.Words, r.Characters, r.Bytes, r.Sentences)
		fmt.Printf("  longest word: %s\n", r.LongestWord)
		fmt.Printf("  top words: %s\n", formatTopWords(r.TopWords))
	case task.StatusFailed:
		fmt.Printf("  error: %s\n", t.Error)
	}
}

func formatTopWords(words []task.WordFrequency) string {
	parts := make([]string, 0, len(words))
	for _, w := range words {
		parts = append(parts, fmt.Sprintf("%s (%d)", w.Word, w.Count))
	}
	return strings.Join(parts, ", ")
}
