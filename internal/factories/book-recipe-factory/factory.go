package book_recipe_factory

import (
	"github.com/gin-gonic/gin"
	authorController "github.com/gmaschi/go-recipes-book/internal/controllers/author"
	recipeController "github.com/gmaschi/go-recipes-book/internal/controllers/recipe"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
)

type (
	Factory struct {
		store              db.Store
		bookRecipesHandler bookRecipesHandler
		Router             *gin.Engine
	}

	bookRecipesHandler struct {
		authorController *authorController.Controller
		recipeController *recipeController.Controller
	}
)

func New(store db.Store) *Factory {
	factory := &Factory{
		store: store,
		bookRecipesHandler: bookRecipesHandler{
			authorController: authorController.New(store),
			recipeController: recipeController.New(store),
		},
	}
	router := gin.Default()

	factory.setupRoutes(router)

	factory.Router = router
	return factory
}

func (f *Factory) setupRoutes(router *gin.Engine) {
	authors := router.Group("/authors")
	{
		authors.POST("", f.bookRecipesHandler.authorController.Create)
		authors.GET("/:username", f.bookRecipesHandler.authorController.Author)
		authors.PATCH("", f.bookRecipesHandler.authorController.Update)
		authors.GET("", f.bookRecipesHandler.authorController.List)
		authors.DELETE("/:username", f.bookRecipesHandler.authorController.Delete)
	}

	recipes := router.Group("/recipes")
	{
		recipes.POST("", f.bookRecipesHandler.recipeController.Create)
		recipes.GET("/:id", f.bookRecipesHandler.recipeController.Recipe)
		recipes.PATCH("", f.bookRecipesHandler.recipeController.Update)
		recipes.DELETE("/:id", f.bookRecipesHandler.recipeController.Delete)
		recipes.GET("", f.bookRecipesHandler.recipeController.List)
	}
}

func (f *Factory) Start(address string) error {
	return f.Router.Run(address)
}
