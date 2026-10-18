package analyzer

import (
	"cmp"
	"errors"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/heilagg/Text-Analyzer/internal/task"
)

const DefaultTopN = 3

var ErrNoWords = errors.New("text contains no words")

type TextAnalyzer struct {
	TopN int
}

func New(topN int) TextAnalyzer {
	if topN <= 0 {
		topN = DefaultTopN
	}
	return TextAnalyzer{TopN: topN}
}

func (a TextAnalyzer) Process(text string) (task.Result, error) {
	words := splitWords(text)
	if len(words) == 0 {
		return task.Result{}, ErrNoWords
	}

	return task.Result{
		Words:       len(words),
		Characters:  utf8.RuneCountInString(text),
		Bytes:       len(text),
		Sentences:   countSentences(text),
		LongestWord: longestWord(words),
		TopWords:    topWords(words, a.TopN),
	}, nil
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func splitWords(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !isWordRune(r)
	})
}

func countSentences(text string) int {
	count := 0
	inSentence := false
	for _, r := range text {
		switch {
		case r == '.' || r == '!' || r == '?':
			if inSentence {
				count++
				inSentence = false
			}
		case isWordRune(r):
			inSentence = true
		}
	}
	if inSentence {
		count++
	}
	return count
}

func longestWord(words []string) string {
	longest := words[0]
	maxLen := utf8.RuneCountInString(longest)
	for _, w := range words[1:] {
		if n := utf8.RuneCountInString(w); n > maxLen {
			longest, maxLen = w, n
		}
	}
	return longest
}

func topWords(words []string, n int) []task.WordFrequency {
	counts := make(map[string]int)
	for _, w := range words {
		counts[strings.ToLower(w)]++
	}

	freqs := make([]task.WordFrequency, 0, len(counts))
	for w, c := range counts {
		freqs = append(freqs, task.WordFrequency{Word: w, Count: c})
	}

	slices.SortFunc(freqs, func(a, b task.WordFrequency) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Word, b.Word)
	})

	if len(freqs) > n {
		freqs = freqs[:n]
	}
	return freqs
}
