package db

import (
	"context"
	"database/sql"
	"github.com/gmaschi/go-recipes-book/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomRecipe(t *testing.T) Recipe {
	author := createRandomAuthor(t)

	createArgs := CreateRecipeParams{
		Author:      author.Username,
		Ingredients: random.StringSlice(6),
		Steps:       random.StringSlice(4),
	}

	recipe, err := testQueries.CreateRecipe(context.Background(), createArgs)
	require.NoError(t, err)
	require.NotEmpty(t, recipe)
	require.NotEmpty(t, recipe.ID)

	require.Equal(t, createArgs.Author, recipe.Author)
	require.Equal(t, createArgs.Ingredients, recipe.Ingredients)
	require.Equal(t, createArgs.Steps, recipe.Steps)
	require.NotEmpty(t, recipe.CreatedAt)
	require.NotEmpty(t, recipe.UpdatedAt)

	return recipe
}

func TestCreateRecipe(t *testing.T) {
	createRandomRecipe(t)
}

func TestGetRecipe(t *testing.T) {
	recipe := createRandomRecipe(t)

	gotRecipe, err := testQueries.GetRecipe(context.Background(), recipe.ID)
	require.NoError(t, err)
	require.NotEmpty(t, gotRecipe)

	require.Equal(t, recipe.ID, gotRecipe.ID)
	require.Equal(t, recipe.Author, gotRecipe.Author)
	require.Equal(t, recipe.Ingredients, gotRecipe.Ingredients)
	require.Equal(t, recipe.Steps, gotRecipe.Steps)
	require.Equal(t, recipe.CreatedAt, gotRecipe.CreatedAt)
	require.Equal(t, recipe.UpdatedAt, gotRecipe.UpdatedAt)
}

func TestUpdateRecipe(t *testing.T) {
	recipe := createRandomRecipe(t)

	updateArgs := UpdateRecipeParams{
		ID: recipe.ID,
		Ingredients: random.StringSlice(4),
		Steps: random.StringSlice(5),
		UpdatedAt: time.Now().UTC(),
	}

	updatedRecipe, err := testQueries.UpdateRecipe(context.Background(),updateArgs)
	require.NoError(t, err)
	require.NotEmpty(t, updatedRecipe)

	require.Equal(t, recipe.ID, updatedRecipe.ID)
	require.Equal(t, updateArgs.Ingredients, updatedRecipe.Ingredients)
	require.Equal(t, updateArgs.Steps, updatedRecipe.Steps)
	require.Equal(t, recipe.CreatedAt, updatedRecipe.CreatedAt)
	require.WithinDuration(t, updateArgs.UpdatedAt, updatedRecipe.UpdatedAt, time.Second)
}

func TestDeleteRecipe(t *testing.T) {
	recipe := createRandomRecipe(t)

	err := testQueries.DeleteRecipe(context.Background(), recipe.ID)
	require.NoError(t, err)

	deletedRecipe, err := testQueries.GetRecipe(context.Background(), recipe.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, deletedRecipe)
}

func TestListRecipes(t *testing.T) {
	n := 10
	for i := 0; i < n; i++ {
		createRandomRecipe(t)
	}

	limit := 5
	offset := 3

	listArgs := ListRecipesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	recipesList, err := testQueries.ListRecipes(context.Background(), listArgs)
	require.NoError(t, err)
	require.NotEmpty(t, recipesList)

	require.Len(t, recipesList, limit)

	for _, recipe := range recipesList {
		require.NotEmpty(t, recipe)
	}
}