-- name: CreateRecipe :one
INSERT INTO recipes (
    author, ingredients, steps
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: GetRecipe :one
SELECT * FROM recipes
WHERE id = $1 LIMIT 1;

-- name: ListRecipes :many
SELECT * FROM recipes
ORDER BY id
LIMIT $1
    OFFSET $2;

-- name: UpdateRecipe :one
UPDATE recipes SET (author, ingredients, steps) = ($2, $3, $4)
WHERE id = $1
RETURNING *;

-- name: DeleteRecipe :exec
DELETE FROM recipes
WHERE id = $1;