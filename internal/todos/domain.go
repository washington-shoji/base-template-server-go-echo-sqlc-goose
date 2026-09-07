package todos

import (
	"time"

	"github.com/google/uuid"
)

// Todo is the domain model returned by the application service.
type Todo struct {
	TodoID    uuid.UUID `json:"todo_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Label     string    `json:"label"`
	Completed bool      `json:"completed"`
}

type Params struct {
	Label     string `json:"label"`
	Completed bool   `json:"completed"`
}

// Summary holds aggregate todo counts (global; todos are not user-owned).
type Summary struct {
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Outstanding int `json:"outstanding"`
}

const (
	EventTodoCreated = "todos.created"
)
