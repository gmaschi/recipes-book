package authorController

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	authorModel "github.com/gmaschi/go-recipes-book/internal/models/author"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"github.com/gmaschi/go-recipes-book/pkg/tools/parseErrors"
	"github.com/gmaschi/go-recipes-book/pkg/tools/password"
	"github.com/gmaschi/go-recipes-book/pkg/tools/validators"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

type Controller struct {
	store db.Store
}

// New creates a pointer to a Controller
func New(store db.Store) *Controller {
	return &Controller{
		store: store,
	}
}

// Create handles the request to create a new author
func (c *Controller) Create(ctx *gin.Context) {
	var req authorModel.CreateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	createArgs := db.CreateAuthorParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		Email:          req.Email,
	}

	author, err := c.store.CreateAuthor(ctx, createArgs)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, parseErrors.ErrorResponse(pqError))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	authorResponse := authorModel.CreateResponse(author)
	ctx.JSON(http.StatusOK, authorResponse)
}

// Author handles the request to get an author based on the username
func (c *Controller) Author(ctx *gin.Context) {
	var req authorModel.GetRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	author, err := c.store.GetAuthor(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	res := authorModel.GetResponse{
		Username:  author.Username,
		Email:     author.Email,
		CreatedAt: author.CreatedAt,
		UpdatedAt: author.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, res)
}

// Update handles the request to update an author email and/or password
func (c *Controller) Update(ctx *gin.Context) {
	var req authorModel.UpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	author, err := c.store.GetAuthor(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	updateArgs := db.UpdateAuthorParams{
		Username:       author.Username,
		Email:          author.Email,
		HashedPassword: author.HashedPassword,
		UpdatedAt:      author.UpdatedAt,
	}

	now := time.Now().UTC()
	trimmedEmail := strings.Trim(req.Email, " ")
	trimmedPassword := strings.Trim(req.Password, " ")

	if trimmedEmail != "" {
		if !validators.Email(trimmedEmail) {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{"error:": "Invalid email"})
			return
		}
		updateArgs.Email = trimmedEmail
		updateArgs.UpdatedAt = now
	}
	if trimmedPassword != "" {
		if !validators.Password(trimmedPassword) {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{"error:": "Invalid password"})
			return
		}
		hashedPassword, err := password.HashPassword(trimmedPassword)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
			return
		}
		updateArgs.HashedPassword = hashedPassword
		updateArgs.UpdatedAt = now
	}

	updatedAuthor, err := c.store.UpdateAuthor(ctx, updateArgs)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, parseErrors.ErrorResponse(pqError))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	res := authorModel.UpdateResponse(updatedAuthor)

	ctx.JSON(http.StatusOK, res)
}

// Delete handles the request to delete an author
func (c *Controller) Delete(ctx *gin.Context) {
	var req authorModel.DeleteRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	err := c.store.DeleteAuthor(ctx, req.Username)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "foreign_key_violation":
				ctx.JSON(http.StatusForbidden, parseErrors.ErrorResponse(pqError))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

// List handles the request to list the authors with pagination
func (c *Controller) List(ctx *gin.Context) {
	var req authorModel.ListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	listArgs := db.ListAuthorsParams{
		Limit:  req.PageSize,
		Offset: req.PageSize * (req.PageID - 1),
	}

	authors, err := c.store.ListAuthors(ctx, listArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	k := len(authors)
	res := make([]authorModel.ListResponse, 0, k)
	for _, author := range authors {
		res = append(res, authorModel.ListResponse(author))
	}

	ctx.JSON(http.StatusOK, res)
}
