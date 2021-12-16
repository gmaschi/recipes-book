-- name: CreateAuthor :one
INSERT INTO authors (
    username, hashed_password, email
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: GetAuthor :one
SELECT * FROM authors
WHERE username = $1 LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors
ORDER BY username
LIMIT $1
OFFSET $2;

-- name: UpdateAuthor :one
UPDATE authors SET email = $2
WHERE username = $1
RETURNING *;

-- name: DeleteAuthor :exec
DELETE FROM authors
WHERE username = $1;