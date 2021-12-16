// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"time"
)

type Author struct {
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Recipe struct {
	ID          int64     `json:"id"`
	Author      string    `json:"author"`
	Ingredients []string  `json:"ingredients"`
	Steps       []string  `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
