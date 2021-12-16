package factories

import (
	"github.com/gin-gonic/gin"
	authorController "github.com/gmaschi/go-recipes-book/internal/controllers/author"
	recipeController "github.com/gmaschi/go-recipes-book/internal/controllers/recipe"
	db "github.com/gmaschi/go-recipes-book/internal/services/datastore/postgresql/recipes/sqlc"
)

type Factory struct {
	store            db.Store
	authorController *authorController.Controller
	recipeController *recipeController.Controller
	router           *gin.Engine
}

func New(store db.Store) *Factory {
	factory := &Factory{store: store,
		authorController: authorController.New(store),
		recipeController: recipeController.New(store),
	}
	router := gin.Default()

	router.POST("/authors", factory.authorController.Create)
	router.GET("/authors/:username", factory.authorController.Author)
	router.PATCH("/authors", factory.authorController.Update)
	router.GET("/authors", factory.authorController.List)
	router.DELETE("/authors/:username", factory.authorController.Delete)

	router.POST("/recipes", factory.recipeController.Create)
	router.GET("/recipes/:id", factory.recipeController.Recipe)
	router.PATCH("/recipes", factory.recipeController.Update)
	router.DELETE("/recipes/:id", factory.recipeController.Delete)
	router.GET("/recipes", factory.recipeController.List)

	factory.router = router
	return factory
}

func (factory *Factory) Start(address string) error {
	return factory.router.Run(address)
}
