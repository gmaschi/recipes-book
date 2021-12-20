package authorModel

import "time"

type (
	CreateResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		Email          string    `json:"-"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"-"`
	}

	GetResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		Email          string    `json:"email"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	UpdateResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		Email          string    `json:"email"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	ListResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		Email          string    `json:"email"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}
)
