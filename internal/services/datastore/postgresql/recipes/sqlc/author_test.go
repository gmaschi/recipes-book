package db

import (
	"context"
	"database/sql"
	"github.com/gmaschi/go-recipes-book/pkg/tools/password"
	"github.com/gmaschi/go-recipes-book/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAuthor(t *testing.T) Author {
	randomPassword := random.String(8)
	hashedPassword, err := password.HashPassword(randomPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	arg := CreateAuthorParams{
		Username:       random.String(10),
		Email:          random.Email(),
		HashedPassword: hashedPassword,
	}

	author, err := testQueries.CreateAuthor(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, author)
	require.Equal(t, arg.Username, author.Username)
	require.Equal(t, arg.Email, author.Email)
	require.Equal(t, arg.HashedPassword, author.HashedPassword)
	return author
}

func TestCreateAuthor(t *testing.T) {
	createRandomAuthor(t)
}

func TestGetAuthor(t *testing.T) {
	author := createRandomAuthor(t)

	authorRes, err := testQueries.GetAuthor(context.Background(), author.Username)
	require.NoError(t, err)
	require.NotEmpty(t, authorRes)

	require.Equal(t, author.Username, authorRes.Username)
	require.Equal(t, author.Email, authorRes.Email)
	require.Equal(t, author.HashedPassword, authorRes.HashedPassword)
	require.Equal(t, author.CreatedAt, authorRes.CreatedAt)
	require.Equal(t, author.UpdatedAt, authorRes.UpdatedAt)
}

func TestUpdateAuthor(t *testing.T) {
	author := createRandomAuthor(t)
	newPassword := random.String(10)
	hashedPassword, err := password.HashPassword(newPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	updateArgs := UpdateAuthorParams{
		Username:       author.Username,
		Email:          random.Email(),
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now().UTC(),
	}

	updatedAuthor, err := testQueries.UpdateAuthor(context.Background(), updateArgs)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAuthor)

	require.Equal(t, author.Username, updatedAuthor.Username)
	require.Equal(t, updateArgs.Email, updatedAuthor.Email)
	require.Equal(t, updateArgs.HashedPassword, updatedAuthor.HashedPassword)
	require.Equal(t, author.CreatedAt, updatedAuthor.CreatedAt)
	require.WithinDuration(t, updateArgs.UpdatedAt, updatedAuthor.UpdatedAt, time.Second)
}

func TestDeleteAuthor(t *testing.T) {
	author := createRandomAuthor(t)

	err := testQueries.DeleteAuthor(context.Background(), author.Username)
	require.NoError(t, err)

	deletedAuthor, err := testQueries.GetAuthor(context.Background(), author.Username)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, deletedAuthor)
}

func TestListAuthors(t *testing.T) {
	n := 10
	for i := 0; i < n; i++ {
		createRandomAuthor(t)
	}

	limit := 5
	offset := 3

	listArgs := ListAuthorsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	authorsList, err := testQueries.ListAuthors(context.Background(), listArgs)
	require.NoError(t, err)
	require.NotEmpty(t, authorsList)

	require.Len(t, authorsList, limit)

	for _, author := range authorsList {
		require.NotEmpty(t, author)
	}
}
