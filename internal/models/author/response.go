package author

import "time"

type CreateResponse struct {
	Username          string    `json:"username"`
	CreatedAt         time.Time `json:"created_at"`
}