package bookRecipeFactory

import (
	"fmt"
	"github.com/gin-gonic/gin"
	authorController "github.com/gmaschi/go-recipes-book/internal/controllers/author"
	authMiddleware "github.com/gmaschi/go-recipes-book/internal/controllers/middlewares/auth"
	recipeController "github.com/gmaschi/go-recipes-book/internal/controllers/recipe"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
	"github.com/gmaschi/go-recipes-book/pkg/auth/tokenAuth"
	pasetoToken "github.com/gmaschi/go-recipes-book/pkg/auth/tokenAuth/paseto"
	"github.com/gmaschi/go-recipes-book/pkg/config/env"
)


type (
	Factory struct {
		store              db.Store
		bookRecipesHandler bookRecipesHandler
		TokenAuth          tokenAuth.Maker
		Config             env.Config
		Router             *gin.Engine
	}

	bookRecipesHandler struct {
		authorController *authorController.Controller
		recipeController *recipeController.Controller
	}
)

func New(config env.Config, store db.Store) (*Factory, error) {
	tokenMaker, err := pasetoToken.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	factory := &Factory{
		store: store,
		bookRecipesHandler: bookRecipesHandler{
			authorController: authorController.New(store),
			recipeController: recipeController.New(store),
		},
		TokenAuth: tokenMaker,
		Config:    config,
	}
	router := gin.Default()

	factory.setupRoutes(router)

	factory.Router = router
	return factory, nil
}

func (f *Factory) setupRoutes(router *gin.Engine) {

	authors := router.Group("/authors")
	{
		authors.POST("/login", f.bookRecipesHandler.authorController.Login)
		authors.POST("", f.bookRecipesHandler.authorController.Create)
		authors.GET("/:username", f.bookRecipesHandler.authorController.Author)
		authors.GET("", f.bookRecipesHandler.authorController.List)

		authAuthorsRoutes := authors.Group("").Use(authMiddleware.AuthMiddleware(f.TokenAuth))

		authAuthorsRoutes.PATCH("", f.bookRecipesHandler.authorController.Update)
		authAuthorsRoutes.DELETE("/:username", f.bookRecipesHandler.authorController.Delete)
	}

	recipes := router.Group("/recipes").Use(authMiddleware.AuthMiddleware(f.TokenAuth))
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
