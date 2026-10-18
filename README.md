# Text Analyzer

SIS project for the Golang Application Development course (KBTU, Fall 2026).
Checkpoint 1: Core Go.

A small synchronous in-memory Go application that processes text analysis tasks.
Each task takes a piece of text, validates it, analyzes it and stores the result.

## What a task does

Input: a text.

Result:
- number of words
- number of characters (runes) and bytes
- number of sentences
- the longest word
- top 3 most frequent words (case-insensitive)

Example for "Hello Go world": 3 words, 14 characters, 1 sentence.

## Requirements

Go 1.27

## How to run

```bash
go run ./cmd/analyzer
```

The demo creates several tasks (including invalid ones), processes them,
prints the results and shows how invalid operations are handled.

## How to test

```bash
gofmt -l .
go vet ./...
go test ./...
```

## Project structure

```
cmd/analyzer/        entry point with a demo
internal/task/       Task model, statuses, lifecycle rules, input validation
internal/analyzer/   text analysis logic
internal/storage/    in-memory task storage (map)
internal/service/    creates, gets and processes tasks
```

## Task lifecycle

```
pending -> processing -> completed
                      -> failed
```

| From       | Allowed transitions    |
|------------|------------------------|
| pending    | processing             |
| processing | completed, failed      |
| completed  | none (final state)     |
| failed     | none (final state)     |

- A new task starts as `pending`.
- Processing moves it to `processing`.
- If the analysis succeeds, the task becomes `completed` and gets a result.
- If the analysis fails (for example, the text has no words, like "!!! ???"), the task becomes `failed` and stores the error.
- Any other transition returns `ErrInvalidTransition`. For example, processing a `completed` task again is rejected.

## Validation rules

- Empty or whitespace-only text is rejected.
- Text longer than 10000 characters is rejected.

Invalid input is rejected with an error and the task is not created.

## Error handling

| Situation                     | Error                  |
|-------------------------------|------------------------|
| empty or whitespace-only text | `ErrEmptyText`         |
| text is too long              | `ErrTextTooLong`       |
| task with this ID not found   | `ErrTaskNotFound`      |
| invalid state transition      | `ErrInvalidTransition` |
| text contains no words        | `ErrNoWords`           |

All errors are returned as values and can be checked with `errors.Is`. `panic` is not used.

## Design notes

- `service` depends on two interfaces, `Processor` and `Repository`, not on concrete types. This makes it easy to test (a fake processor is used to test the `failed` state) and to replace the storage later.
- Tasks are stored as `map[int]*task.Task`, so every part of the app works with the same task and status updates are not lost on copies.
- Methods that change a task (`Start`, `Complete`, `Fail`) use pointer receivers. `TextAnalyzer.Process` uses a value receiver because it does not change the analyzer.