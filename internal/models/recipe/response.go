package recipeModel

import "time"

type (
	CreateResponse struct {
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
	}

	GetResponse struct {
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	UpdateResponse struct {
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	ListResponse struct {
		Author      string    `json:"author"`
		Ingredients []string  `json:"ingredients"`
		Steps       []string  `json:"steps"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
)
