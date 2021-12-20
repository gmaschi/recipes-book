package recipeModel

import "time"

type (
	CreateResponse struct {
		ID          int64     `json:"-"`
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"-"`
	}

	GetResponse struct {
		ID          int64     `json:"-"`
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	UpdateResponse struct {
		ID          int64     `json:"-"`
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	ListResponse struct {
		ID          int64     `json:"-"`
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
)
