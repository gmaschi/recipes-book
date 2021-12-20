package recipeController

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	recipeModel "github.com/gmaschi/go-recipes-book/internal/models/recipe"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"github.com/gmaschi/go-recipes-book/pkg/tools/parseErrors"
	"github.com/lib/pq"
	"net/http"
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

// Create handles the request to create a new recipe
func (c *Controller) Create(ctx *gin.Context) {
	var req recipeModel.CreateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	if len(req.Ingredients) == 0 || len(req.Steps) == 0 {
		ctx.JSON(http.StatusBadRequest, "There should be at least one step and one ingredient per recipe")
		return
	}

	createArgs := db.CreateRecipeParams{
		Author:      req.Author,
		Ingredients: req.Ingredients,
		Steps:       req.Steps,
	}

	recipe, err := c.store.CreateRecipe(ctx, createArgs)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation":
				ctx.JSON(http.StatusForbidden, parseErrors.ErrorResponse(pqErr))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	res := recipeModel.CreateResponse(recipe)
	ctx.JSON(http.StatusOK, res)
}

// Recipe handles the request to get a recipe by ID
func (c *Controller) Recipe(ctx *gin.Context) {
	var req recipeModel.GetRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	recipe, err := c.store.GetRecipe(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	res := recipeModel.GetResponse(recipe)

	ctx.JSON(http.StatusOK, res)
}

// Update handles the request to update a specific recipe by ID
func (c *Controller) Update(ctx *gin.Context) {
	var req recipeModel.UpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	recipe, err := c.store.GetRecipe(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	now := time.Now().UTC()
	updateArgs := db.UpdateRecipeParams{
		ID:          recipe.ID,
		Steps:       recipe.Steps,
		Ingredients: recipe.Ingredients,
	}

	if len(req.Steps) != 0 {
		updateArgs.Steps = req.Steps
		updateArgs.UpdatedAt = now
	}
	if len(req.Ingredients) != 0 {
		updateArgs.Ingredients = req.Ingredients
		updateArgs.UpdatedAt = now
	}

	updatedRecipe, err := c.store.UpdateRecipe(ctx, updateArgs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	res := recipeModel.UpdateResponse(updatedRecipe)

	ctx.JSON(http.StatusOK, res)
}

// Delete handles a request do delete an recipe
func (c *Controller) Delete(ctx *gin.Context) {
	var req recipeModel.DeleteRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	err := c.store.DeleteRecipe(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

// List handles a request to list recipes with pagination
func (c *Controller) List(ctx *gin.Context) {
	var req recipeModel.ListRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrors.ErrorResponse(err))
		return
	}

	listArgs := db.ListRecipesParams{
		Limit:  req.PageSize,
		Offset: req.PageSize * (req.PageID - 1),
	}

	recipes, err := c.store.ListRecipes(ctx, listArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrors.ErrorResponse(err))
		return
	}

	k := len(recipes)
	res := make([]recipeModel.ListResponse, 0, k)
	for _, recipe := range recipes {
		res = append(res, recipeModel.ListResponse(recipe))
	}

	ctx.JSON(http.StatusOK, res)
}
